"use client"
import { CameraIcon, CloudUploadIcon, LucideCamera, Smartphone, Trash2, WandSparklesIcon, XIcon } from 'lucide-react'
import React, { useState } from 'react'
import { ImageCarousel } from './image-carousel';
import StarIcon from './feather-sparkles';
import { Button } from './ui/button';
import { Badge } from './badge';
import { Input } from './ui/input';
import { cn } from '~/lib/utils';
import { apiClient } from '~/lib/api';
import { request } from 'http';
import toast from 'react-hot-toast';
import { redirect, useRouter } from 'next/navigation';
import { Spinner } from './ui/spinner';

function ClientNew() {

    const images = [
        {
            src: 'https://images.unsplash.com/photo-1504051771394-dd2e66b2e08f?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NjJ8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            inputSrc: 'https://images.unsplash.com/photo-1504051771394-dd2e66b2e08f?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NjJ8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            alt: 'Professional portrait of a woman',
        },
        {
            src: 'https://images.unsplash.com/photo-1526510747491-58f928ec870f?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NTl8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            inputSrc: 'https://images.unsplash.com/photo-1526510747491-58f928ec870f?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NTl8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            alt: 'Scenic landscape with mountains and a lake',
        },
        {
            src: 'https://plus.unsplash.com/premium_photo-1670282392820-e3590c1c5c54?w=900&auto=format&fit=crop&q=60',
            inputSrc: 'https://plus.unsplash.com/premium_photo-1670282392820-e3590c1c5c54?w=900&auto=format&fit=crop&q=60',
            alt: 'Artistic photo of a girl with flowers',
        },
        {
            src: 'https://images.unsplash.com/photo-1581403341630-a6e0b9d2d257?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NDN8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            inputSrc: 'https://images.unsplash.com/photo-1581403341630-a6e0b9d2d257?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NDN8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            alt: 'A dog wearing sunglasses',
        },
        {
            src: 'https://images.unsplash.com/photo-1524504388940-b1c1722653e1?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Mjh8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            inputSrc: 'https://images.unsplash.com/photo-1524504388940-b1c1722653e1?w=900&auto=format&fit=crop&q=60&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Mjh8fGdpcmx8ZW58MHx8MHx8fDA%3D',
            alt: 'Creative shot of a person from behind',
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
            console.error("No photo selected.");
            toast.error("Please select a photo.");
            return;
        }
        setIsGenerating(true);

        // 1. Create a new FormData object.
        const formData = new FormData();

        // 2. Append the file with the key 'image', as specified in the schema.
        formData.append('image', selectedPhotoFile);

        // 3. Define the options payload. This should likely come from your UI state.
        //    For this example, we'll use a hardcoded value.
        const jobOptions = [
            { "prompt": "a photo of a person" },
            { "prompt": "a photo of a cat" },
            { "prompt": "a photo of a dog" },
            { "prompt": "a photo of a fish" },
        ];

        // 4. The schema requires the 'options' to be a JSON string, so we must stringify it.
        formData.append('options', JSON.stringify(jobOptions));

        try {
            // 5. Pass the FormData object directly as the body.
            //    The client library will automatically set the 'Content-Type' to 'multipart/form-data'.
            const response = await apiClient.POST("/api/v1/jobs", {
                body: formData as any,
            });

            if (response.data) {
                console.log("Job created successfully:", response.data);
                toast.success("Job created successfully!");
                router.push("/history")

            } else if (response.error) {
                // Handle API error response
                console.error("Failed to create job:", response.error);
                toast.error("Failed to create job: " + response.error.detail);
            }

        } catch (error) {
            // Handle network or other unexpected errors
            console.error("An unexpected error occurred:", error);
            toast.error("Create job failed");
        } finally {
            setIsGenerating(false);
        }
    };

    return (
        <div className='flex h-screen w-full flex-col items-center justify-center gap-2 md:gap-8 bg-background rounded-md px-6 py-6 transition-all duration-300 ease-in-out'>
            <div className="flex flex-col items-center justify-center gap-3">
                <Badge icon={<Smartphone size={14} className="rotate-12" />} className='border-t-highlight/65 border-1 border-solid bg-background-light text-foreground text-caption bg-gradient-to-b from-background-light to-background shadow-sm '>Selfier</Badge>
                <span className="text-heading-1 font-heading-1 text-foreground">
                    Create Stunning Photos — With Your Selfie
                </span>
                <span className="text-body text-xl font-heading-2 text-muted-foreground">
                    No strangers. No tripods. No missed moments.
                </span>
            </div>

            <ImageCarousel images={images} />

            <div className="flex w-full max-w-[576px] flex-col items-center justify-center gap-2 px-2 py-2 shadow-sm">
                <div className={cn(
                    "flex w-full flex-col items-center justify-center gap-2 rounded-md border-t-highlight border border-solid px-2 py-2 shadow-sm",
                    !selectedPhotoUrl ? "bg-background-light bg-gradient-to-b from-highlight/20 to-background hover:bg-background-light transition-all duration-300 ease-in-out" : "bg-background bg-gradient-to-b from-highlight/10 to-background",
                )}>
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
                                    className="absolute top-0 right-0 cursor-pointer rounded-full bg-error-100 hover:bg-error-300 border transition-all duration-300 ease-in-out border-error-300"
                                >
                                    <XIcon className="text-error-700" />
                                </Button>
                            </div>
                        </div>
                    ) : <label className='cursor-pointer border rounded-sm border-dashed flex flex-col items-center justify-center h-full w-full px-4 py-4'>
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
                    }
                </div>

                <Button
                    className="w-full flex-none cursor-pointer"
                    onClick={(event: React.MouseEvent<HTMLButtonElement>) => { handleGenerate() }}
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
    )
}

export default ClientNew
