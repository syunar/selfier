"use client";
import {
  CameraIcon,
  CloudUploadIcon,
  LucideCamera,
  Smartphone,
  Trash2,
  WandSparklesIcon,
  XIcon,
} from "lucide-react";
import React, { useState } from "react";
import { ImageCarousel } from "./image-carousel";
import StarIcon from "./feather-sparkles";
import { Button } from "./ui/button";
import { Badge } from "./badge";
import { Input } from "./ui/input";
import { cn } from "~/lib/utils";
import { apiClient } from "~/lib/api";
import { request } from "http";
import toast from "react-hot-toast";
import { redirect, useRouter } from "next/navigation";
import { Spinner } from "./ui/spinner";

function ClientNew() {
  const images = [
    {
      src: "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/output_md_lighning_005.png",
      inputSrc:
        "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/005.png",
    },
    {
      src: "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/output_md_lighning_004.png",
      inputSrc:
        "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/004.png",
    },
    {
      src: "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/output_md_lighning_003.png",
      inputSrc:
        "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/003.png",
    },
    {
      src: "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/output_md_lighning_002.png",
      inputSrc:
        "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/002.png",
    },
    {
      src: "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/output_md_lighning_001.png",
      inputSrc:
        "https://ltxijvwnxpxbbvpkcnhh.supabase.co/storage/v1/object/public/marketing/001.png",
    },
  ];

  const [selectedPhotoUrl, setSelectedPhotoUrl] = useState<string | null>(null);
  const [selectedPhotoFile, setSelectedPhotoFile] = useState<File | null>(null);
  const [isGenerating, setIsGenerating] = useState<boolean>(false);

  const router = useRouter();

  const handlePhotoFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedPhotoUrl(URL.createObjectURL(file));
      setSelectedPhotoFile(file);
    }
  };

  const handleGenerate = async () => {
    if (!selectedPhotoFile) {
      toast.error("Please select a photo.");
      return;
    }
    setIsGenerating(true);

    // 1. Create a new FormData object.
    const formData = new FormData();

    // 2. Append the file with the key 'image', as specified in the schema.
    formData.append("image", selectedPhotoFile);

    // 3. Define the options payload. This should likely come from your UI state.
    //    For this example, we'll use a hardcoded value.
    const jobOptions = [
      { prompt: "Convert image to Full-body shot with fashion pose." },
      {
        prompt:
          "Convert image to Ultra-Wide shot showing the full person with fashion pose.",
      },
      {
        prompt:
          "Convert image to Medium-full shot from the knees up with fashion pose.",
      },
      { prompt: "Convert image to Drone shot with fashion pose." },
    ];

    // 4. The schema requires the 'options' to be a JSON string, so we must stringify it.
    formData.append("options", JSON.stringify(jobOptions));

    try {
      // 5. Pass the FormData object directly as the body.
      //    The client library will automatically set the 'Content-Type' to 'multipart/form-data'.
      const response = await apiClient.POST("/api/v1/jobs", {
        body: formData as any,
      });

      if (response.data) {
        console.log("Job created successfully:", response.data);
        toast.success("Job created successfully!");
        router.push("/history");
      } else if (response.error) {
        // Handle API error response
        console.error("Failed to create job:", response.error);
        toast.error("Failed to create job: " + response.error.detail);
      }
    } catch (error) {
      // Handle network or other unexpected errors
      console.error("An unexpected error occurred:", error);
      toast.error("Create job failed");
    }
  };

  return (
    <div className="bg-background flex h-screen w-full flex-col items-center justify-center gap-2 rounded-md px-6 py-6 transition-all duration-300 ease-in-out md:gap-8">
      <div className="flex flex-col items-center justify-center gap-3">
        <Badge
          icon={<Smartphone size={14} className="rotate-12" />}
          className="border-t-highlight/60 bg-background-light text-foreground text-caption border-1 border-solid shadow-sm"
        >
          Selfier
        </Badge>
        <span className="text-heading-1 font-heading-1 text-foreground">
          Create Stunning Photos — With Your Selfie
        </span>
        <span className="text-body font-heading-2 text-muted-foreground text-xl">
          No strangers. No tripods. No missed moments.
        </span>
      </div>

      <ImageCarousel images={images} />

      <div className="flex w-full max-w-[576px] flex-col items-center justify-center gap-2 px-2 py-2 shadow-sm">
        <div
          className={cn(
            "border-t-highlight flex w-full flex-col items-center justify-center gap-2 rounded-md border border-solid px-2 py-2 shadow-sm",
            !selectedPhotoUrl
              ? "bg-background-light from-highlight/20 to-background hover:bg-background-light bg-gradient-to-b transition-all duration-300 ease-in-out"
              : "bg-background-light",
          )}
        >
          {selectedPhotoUrl ? (
            <div className="flex flex-col items-center justify-between gap-4">
              <div className="relative p-2">
                <img
                  src={selectedPhotoUrl}
                  crossOrigin="anonymous"
                  className="max-h-[240px] max-w-full rounded-xl border object-contain md:max-h-[240px]"
                />
                <Button
                  variant="default"
                  size={"icon-sm"}
                  onClick={() => {
                    setSelectedPhotoFile(null);
                    setSelectedPhotoUrl(null);
                  }}
                  className="bg-error-100 hover:bg-error-300 border-error-300 absolute top-0 right-0 cursor-pointer rounded-full border transition-all duration-300 ease-in-out"
                >
                  <XIcon className="text-error-700" />
                </Button>
              </div>
            </div>
          ) : (
            <label className="flex h-full w-full cursor-pointer flex-col items-center justify-center rounded-sm border border-dashed px-4 py-4">
              <CloudUploadIcon className="text-heading-1 font-heading-1" />
              <div className="flex flex-col items-center justify-center gap-1">
                <span className="text-body font-body text-foreground text-center">
                  Click to select image or drag to upload
                </span>
                <span className="text-caption font-caption text-muted-foreground text-center">
                  Single image, max file size 5MB
                </span>
              </div>
              <Input
                onChange={handlePhotoFileChange}
                type="file"
                accept="image/*"
                className="hidden"
              />
            </label>
          )}
        </div>

        <Button
          className="w-full flex-none cursor-pointer"
          onClick={(event: React.MouseEvent<HTMLButtonElement>) => {
            handleGenerate();
          }}
        >
          {isGenerating ? (
            <Spinner /> // Show spinner when loading
          ) : (
            <StarIcon /> // Show star icon when not loading
          )}
          Generate
        </Button>
      </div>
    </div>
  );
}

export default ClientNew;
