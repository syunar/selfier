from io import BytesIO
from pathlib import Path
import math
import uuid

import modal
from pydantic import BaseModel, Field

diffusers_commit_sha = "23ebbb4bc81a17ebea17cb7cb94f301199e49a7f"

image = (
    modal.Image.from_registry(
        "nvidia/cuda:12.8.1-devel-ubuntu22.04",
        add_python="3.12",
    )
    .entrypoint([])
    .apt_install("git")
    .uv_pip_install(
        "accelerate~=1.8.1",
        f"git+https://github.com/huggingface/diffusers.git@{diffusers_commit_sha}",
        "huggingface-hub[hf-transfer]",
        "Pillow",
        "safetensors",
        "transformers",
        "sentencepiece",
        "torch==2.7.1",
        "torchvision==0.22.1",
        "optimum-quanto==0.2.7",
        "boto3",
        "requests",
        "peft",
        "fastapi[standard]",
        "bitsandbytes",
        extra_options="--index-strategy unsafe-best-match",
        extra_index_url="https://download.pytorch.org/whl/cu128",
    )
)

MODEL_NAME = "ovedrive/Qwen-Image-Edit-2509-4bit"
BUCKET_NAME = "selfier-modal"
CACHE_DIR = Path("/cache")

cache_volume = modal.Volume.from_name("hf-hub-cache", create_if_missing=True)
volumes = {CACHE_DIR: cache_volume}

secrets = [
    modal.Secret.from_name("huggingface-secret"),
    modal.Secret.from_name("s3-secret"),
]

image = image.env(
    {
        "HF_HUB_ENABLE_HF_TRANSFER": "1",
        "HF_HOME": str(CACHE_DIR),
    }
)

app = modal.App("selfier-deselfie")

with image.imports():
    import torch
    import boto3
    from diffusers import DiffusionPipeline, FlowMatchEulerDiscreteScheduler
    from diffusers.utils import load_image
    from botocore.exceptions import NoCredentialsError, ClientError
    from botocore.client import Config
    import os


class DeselfieRequest(BaseModel):
    image_url: str
    prompt: str

class DeselfieResponse(BaseModel):
    image_url: str


@app.cls(
    image=image,
    gpu="L4",
    volumes=volumes,
    secrets=secrets,
    scaledown_window=10,
    # enable_memory_snapshot=True,
    # experimental_options={"enable_gpu_snapshot": True},
    max_containers=1,
)
class DeselfieModal:

    # @modal.enter(snap=True)
    @modal.enter()
    def load(self):

        self.s3_client = boto3.client(
            "s3",
            endpoint_url=os.environ["AWS_ENDPOINT_URL"],
            aws_access_key_id=os.environ["AWS_ACCESS_KEY_ID"],
            aws_secret_access_key=os.environ["AWS_SECRET_ACCESS_KEY"],
            region_name=os.environ["AWS_REGION"],
            config=Config(
                signature_version="s3v4",
                # signature_version="v4",
                s3={'addressing_style': 'path'},
                ),
        )

        scheduler_config = {
            "base_image_seq_len": 256,
            "base_shift": math.log(3),
            "invert_sigmas": False,
            "max_image_seq_len": 8192,
            "max_shift": math.log(3),
            "num_train_timesteps": 1000,
            "shift": 1.0,
            "shift_terminal": None,
            "stochastic_sampling": False,
            "time_shift_type": "exponential",
            "use_beta_sigmas": False,
            "use_dynamic_shifting": True,
            "use_exponential_sigmas": False,
            "use_karras_sigmas": False,
        }
        scheduler = FlowMatchEulerDiscreteScheduler.from_config(scheduler_config)

        dtype = torch.bfloat16
        device = "cuda"

        self.pipe = DiffusionPipeline.from_pretrained(
            MODEL_NAME,
            scheduler=scheduler,
            torch_dtype=dtype,
            cache_dir=CACHE_DIR,
        ).to(device)
        # self.pipe.enable_model_cpu_offload()

        self.pipe.load_lora_weights(
            "lightx2v/Qwen-Image-Lightning",
            subfolder="Qwen-Image-Edit-2509",
            weight_name="Qwen-Image-Edit-2509-Lightning-4steps-V1.0-bf16.safetensors",
            adapter_name="lightning",
            cache_dir=CACHE_DIR,
        )

    # @modal.enter(snap=False)
    # def setup(self):
    #     # Move to GPU after restore
    #     self.pipe.to("cuda")
    #     # self.pipe.enable_model_cpu_offload()

    @modal.method()
    def generate_and_upload(self, request: DeselfieRequest) -> str:
        print(f"Running inference for prompt: {request.prompt}")
        try:
            init_image = load_image(str(request.image_url)).convert("RGB")
            pil_image = self.pipe(
                prompt=request.prompt,
                image=[init_image],
                num_inference_steps=4,
                true_cfg_scale=1.0,
                output_type="pil",
            ).images[0]

            with BytesIO() as buffer:
                pil_image.save(buffer, format="JPEG")
                image_bytes = buffer.getvalue()

            object_name = f"generated/{uuid.uuid4()}.jpg"
            print(f"Uploading image to S3 object: s3://{BUCKET_NAME}/{object_name}")

            self.s3_client.put_object(
                Bucket=BUCKET_NAME,
                Key=object_name,
                Body=image_bytes,
                ContentType="image/jpeg",
            )

            presigned_get_url = self.s3_client.generate_presigned_url(
                "get_object",
                Params={"Bucket": BUCKET_NAME, "Key": object_name},
                ExpiresIn=3600,
            )

            return presigned_get_url

        except (NoCredentialsError, ClientError) as e:
            print(f"S3 Error: {e}")
            raise
        except Exception as e:
            print(f"An error occurred during generation or upload: {e}")
            raise

    @modal.fastapi_endpoint(method="POST", requires_proxy_auth=True)
    def generate(self, request: DeselfieRequest) -> DeselfieResponse:
        output_image_url = self.generate_and_upload.local(request)

        return DeselfieResponse(image_url=output_image_url)


@app.local_entrypoint()
def main():
    instance = DeselfieModal()
    endpoint_url = instance.generate.get_web_url()
    print(f"Testing endpoint: {endpoint_url}")
