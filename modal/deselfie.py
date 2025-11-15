from io import BytesIO
from pathlib import Path
import math
import uuid

import modal
from pydantic import BaseModel, Field

# Pinning a specific commit is a great practice for reproducibility.
diffusers_commit_sha = "23ebbb4bc81a17ebea17cb7cb94f301199e49a7f"

image = (
    modal.Image.from_registry(
        "nvidia/cuda:12.8.1-devel-ubuntu22.04",  # Match PyTorch's CUDA version
        add_python="3.12",  # Use a widely supported Python version
    )
    .entrypoint([])  # remove verbose logging by base image on entry
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
        "boto3",  # For S3 interaction
        "requests",
        "peft",
        "fastapi[standard]",
        "bitsandbytes",
        extra_options="--index-strategy unsafe-best-match",
        extra_index_url="https://download.pytorch.org/whl/cu128",
    )
)

# --- Constants and Configuration ---
# MODEL_NAME = "Qwen/Qwen-Image-Edit-2509"
MODEL_NAME = "ovedrive/Qwen-Image-Edit-2509-4bit"
# 💡 IMPORTANT: Make sure this S3 bucket exists and you have configured secrets for it.
BUCKET_NAME = "selfier-modal"
CACHE_DIR = Path("/cache")

# --- Modal Resources ---
cache_volume = modal.Volume.from_name("hf-hub-cache", create_if_missing=True)
volumes = {CACHE_DIR: cache_volume}

# Reference existing secrets. `from_name` does not take a `required_keys` argument.
secrets = [
    modal.Secret.from_name("huggingface-secret"),
    modal.Secret.from_name("s3-secret"),
]

# Apply environment variables to the image
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


# --- Pydantic Models for API ---
class DeselfieRequest(BaseModel):
    image_url: str
    prompt: str

class DeselfieResponse(BaseModel):
    image_url: str


@app.cls(
    image=image,
    gpu="L4",  # A10G is a cost-effective choice for this kind of inference.
    # gpu="H100",  # A10G is a cost-effective choice for this kind of inference.
    volumes=volumes,
    secrets=secrets,
    # The container will stay active for 4 minutes of inactivity before shutting down.
    scaledown_window=10,
    # enable_memory_snapshot=True,
    # experimental_options={"enable_gpu_snapshot": True},
    max_containers=1,
)
class DeselfieModal:

    # @modal.enter(snap=True)
    @modal.enter()
    def load(self):
        """
        This method runs once when the container starts.
        It downloads the model and initializes the pipeline and S3 client.
        """
        # Initialize the S3 client once.
        # It will automatically use the credentials from the Modal secret.
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
        print("✅ Model and S3 client initialized.")

        print(f"Downloading {MODEL_NAME} if necessary...")

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
        """
        The core background task:
        1. Runs inference to generate an image.
        2. Uploads the image bytes directly to S3.
        3. Generates and returns a presigned GET URL for the uploaded image.
        """
        print(f"Running inference for prompt: {request.prompt}")
        try:
            # 1. Run inference
            init_image = load_image(str(request.image_url)).convert("RGB")
            pil_image = self.pipe(
                prompt=request.prompt,
                image=[init_image],
                num_inference_steps=4,
                true_cfg_scale=1.0,
                output_type="pil",
            ).images[0]

            # Convert PIL Image to bytes in memory
            with BytesIO() as buffer:
                pil_image.save(buffer, format="JPEG")
                image_bytes = buffer.getvalue()

            # 2. Upload image bytes directly to S3
            object_name = f"generated/{uuid.uuid4()}.jpg"
            print(f"Uploading image to S3 object: s3://{BUCKET_NAME}/{object_name}")

            self.s3_client.put_object(
                Bucket=BUCKET_NAME,
                Key=object_name,
                Body=image_bytes,
                ContentType="image/jpeg",
            )
            print("✅ Image uploaded successfully.")

            # 3. Generate a presigned URL to *view* (GET) the object
            presigned_get_url = self.s3_client.generate_presigned_url(
                "get_object",
                Params={"Bucket": BUCKET_NAME, "Key": object_name},
                ExpiresIn=3600,  # The URL will be valid for 1 hour
            )
            print(f"✅ Generated presigned GET URL. {presigned_get_url=}")

            return presigned_get_url

        except (NoCredentialsError, ClientError) as e:
            print(f"❌ S3 Error: {e}")
            raise  # Re-raise the exception to signal a failure to the caller
        except Exception as e:
            print(f"❌ An error occurred during generation or upload: {e}")
            raise

    @modal.fastapi_endpoint(method="POST", requires_proxy_auth=True)
    def generate(self, request: DeselfieRequest) -> DeselfieResponse:
        """
        This is the public-facing API endpoint.
        It runs the image generation and S3 upload synchronously and returns
        the presigned URL of the final image.
        """
        print(f"Received request for prompt: '{request.prompt}'. Running synchronously.")

        # The .local() call runs the method and waits for it to complete.
        # The entire API request will block until the image is generated, uploaded,
        # and the URL is returned.
        output_image_url = self.generate_and_upload.local(request)

        print(f"✅ Generated presigned GET URL.")
        return DeselfieResponse(image_url=output_image_url)


@app.local_entrypoint()
def main():
    import requests
    import os

    # This local entrypoint is for testing your Modal app from your local machine.
    # It requires your Modal client to be authenticated (modal token set).
    # You need to manually set the required secrets as environment variables for local testing.
    # For example:
    # export AWS_ACCESS_KEY_ID="your_key"
    # export AWS_SECRET_ACCESS_KEY="your_secret"
    # modal run your_script.py

    # IMPORTANT: The class needs to be instantiated to access its web endpoint URL.
    # This does not run the @modal.enter() hook.
    instance = DeselfieModal()
    endpoint_url = instance.generate.get_web_url()
    print(f"Testing endpoint: {endpoint_url}")

    request_data = DeselfieRequest(
        # Replace with a valid, publicly accessible image URL for testing
        image_url="https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/sign/selfier/output-7e0067d6-dfbd-4d70-b704-2a7f38a26b4c.jpg_1760380898?token=eyJraWQiOiJzdG9yYWdlLXVybC1zaWduaW5nLWtleV8wYjU0ZGYwYy0zODQxLTRiNWUtYTYwZC0wMzcwZjVhOTQzYjUiLCJhbGciOiJIUzI1NiJ9.eyJ1cmwiOiJzZWxmaWVyL291dHB1dC03ZTAwNjdkNi1kZmJkLTRkNzAtYjcwNC0yYTdmMzhhMjZiNGMuanBnXzE3NjAzODA4OTgiLCJpYXQiOjE3NjA4MDk1MzYsImV4cCI6MTc2MTQxNDMzNn0.r2un3lZhwCohev1hWvB_Ms3jCDS7-fcWDlGJQYQrl3Y",
        prompt="Convert image to lineart.",
    )

    payload = request_data.model_dump()

    print("Sending POST request with payload:", payload)
    response = requests.post(endpoint_url, json=payload)

    if response.ok:
        print("✅ Request successful!")
        print(f"{response.json()['image_url']=}")
    else:
        print(f"❌ Request failed with status code: {response.status_code}")
        print("Response text:", response.text)
