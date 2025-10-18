from io import BytesIO
from pathlib import Path
import math
import uuid

import modal
from pydantic import BaseModel, Field, HttpUrl
from typing import Optional


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
        "boto3",
        "requests",
        "peft",
        "fastapi[standard]",
        extra_options="--index-strategy unsafe-best-match",
        extra_index_url="https://download.pytorch.org/whl/cu128",
    )
)

# --- Constants and Configuration ---
MODEL_NAME = "Qwen/Qwen-Image-Edit-2509"
BUCKET_NAME = "selfier-modal" # Make sure this bucket exists and you have access
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
    import requests
    from diffusers import DiffusionPipeline, FlowMatchEulerDiscreteScheduler
    from diffusers.utils import load_image
    from botocore.exceptions import NoCredentialsError, ClientError
    from botocore.client import Config
    import os


# --- Pydantic Models for API ---
class DeselfieRequest(BaseModel):
    image_url: str
    prompt: str
    # callback_url: Optional[str] = Field(None, description="URL to send results back to when processing is done.")

class DeselfieResponse(BaseModel):
    image_url: str


@app.cls(
    image=image,
    gpu="L40S",  # A10G is a cost-effective choice for this kind of inference. B200 is very new and expensive.
    volumes=volumes,
    secrets=secrets,
    scaledown_window=240, # Renamed from scaledown_window
)
class DeselfieModal:
    @modal.enter()
    def enter(self):
        """
        This method runs once when the container starts.
        It downloads the model and initializes the pipeline and S3 client.
        """
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
        self.device = "cuda"

        self.pipe = DiffusionPipeline.from_pretrained(
            MODEL_NAME,
            scheduler=scheduler,
            torch_dtype=dtype,
            cache_dir=CACHE_DIR,
        )
        self.pipe.enable_model_cpu_offload()

        self.pipe.load_lora_weights(
            "lightx2v/Qwen-Image-Lightning",
            subfolder="Qwen-Image-Edit-2509",
            weight_name="Qwen-Image-Edit-2509-Lightning-4steps-V1.0-bf16.safetensors",
            adapter_name="lightning",
            cache_dir=CACHE_DIR,
        )

        # Initialize the S3 client once
        self.s3_client = boto3.client(
            "s3",
            # endpoint_url=os.environ["AWS_ENDPOINT_URL"],
            # aws_access_key_id=os.environ["AWS_ACCESS_KEY_ID"], # Use "supabase" as the access key ID
            # aws_secret_access_key=os.environ["AWS_SECRET_ACCESS_KEY"],
            # config=Config(signature_version="s3v4"),
            # region_name=os.environ["AWS_REGION"] # Can be any valid AWS region, as it's not used by Supabase
        )

        print("✅ Model and S3 client initialized.")

    def _generate_presigned_url(self, bucket_name, object_name, expiration=3600):
        """Helper method to generate a presigned URL for uploading a file."""
        try:
            response = self.s3_client.generate_presigned_url(
                "put_object",
                Params={"Bucket": bucket_name, "Key": object_name, "ContentType": "image/jpeg"},
                ExpiresIn=expiration,
            )
            return response
        except (NoCredentialsError, ClientError) as e:
            print(f"❌ Error generating presigned URL: {e}")
            return None

    @modal.method()
    def generate_and_upload(self, request: DeselfieRequest):
        """
        The core background task:
        1. Runs inference to generate an image.
        2. Generates a presigned URL for S3 upload.
        3. Uploads the image to S3.
        4. Sends a callback with the final public image URL.
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

            # 2. Prepare for S3 upload
            object_name = f"generated/{uuid.uuid4()}.jpg"
            presigned_url = self._generate_presigned_url(BUCKET_NAME, object_name)

            if not presigned_url:
                raise Exception("Failed to get a presigned URL.")

            # 3. Upload image bytes to S3 using the presigned URL
            print(f"Uploading image to S3 object: {object_name}")
            response = requests.put(
                presigned_url,
                data=image_bytes,
                headers={"Content-Type": "image/jpeg"},
            )
            response.raise_for_status()  # Will raise an exception for 4xx/5xx responses
            print(f"✅ Image uploaded to presigned_url: {presigned_url}")

            # # 4. Send callback if a URL was provided
            # if request.callback_url:
            #     print(f"Sending success callback to {request.callback_url}")
            #     requests.post(
            #         str(request.callback_url),
            #         json={"status": "success", "image_url": presigned_url},
            #     )

        except Exception as e:
            print(f"❌ An error occurred: {e}")
            # Optionally send a failure callback
            # if request.callback_url:
            #     print(f"Sending failure callback to {request.callback_url}")
            #     requests.post(
            #         str(request.callback_url),
            #         json={"status": "error", "message": str(e)},
            #     )
            presigned_url = None

        return presigned_url

    @modal.fastapi_endpoint(method="POST", docs=True)
    def generate(self, request: DeselfieRequest) -> DeselfieResponse:
        """
        This is the public-facing API endpoint.
        It immediately spawns the generation task to run in the background
        and returns a confirmation to the client.
        """
        print(
            f"Received request to generate image. Spawning background task."
        )

        # Use .spawn() to run the method in the background (fire-and-forget)
        output_image_url = self.generate_and_upload.local(request)

        # return DeselfieResponse(
        #     status="processing",
        #     message="Image generation has started. You will receive a notification at your callback_url.",
        # )
        return DeselfieResponse(image_url=output_image_url)

@app.local_entrypoint()
def main():

    import requests

    server = DeselfieModal()
    endpoint_url = server.generate.get_web_url()

    request = DeselfieRequest(
        image_url="https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/sign/selfier/output-7e0067d6-dfbd-4d70-b704-2a7f38a26b4c.jpg_1760380898?token=eyJraWQiOiJzdG9yYWdlLXVybC1zaWduaW5nLWtleV8wYjU0ZGYwYy0zODQxLTRiNWUtYTYwZC0wMzcwZjVhOTQzYjUiLCJhbGciOiJIUzI1NiJ9.eyJ1cmwiOiJzZWxmaWVyL291dHB1dC03ZTAwNjdkNi1kZmJkLTRkNzAtYjcwNC0yYTdmMzhhMjZiNGMuanBnXzE3NjAzODA4OTgiLCJpYXQiOjE3NjA4MDk1MzYsImV4cCI6MTc2MTQxNDMzNn0.r2un3lZhwCohev1hWvB_Ms3jCDS7-fcWDlGJQYQrl3Y",
        prompt="Convert image to Full-body shot with fashion pose.",
        # callback_url=HttpUrl("https://example.com/callback"),
    )

    payload = request.model_dump()

    response = requests.post(
        endpoint_url,
        json=payload,
        # headers=headers
        )
    response.raise_for_status()

    print(response.json())
