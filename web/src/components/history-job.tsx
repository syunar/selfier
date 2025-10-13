"use client";

import React, { useEffect, useRef, useState } from 'react'
import type { components } from '~/lib/api-types'
import { Badge } from './badge'
import { Avatar } from './ui/avatar'
import { AlertCircleIcon, CheckCheckIcon, CircleArrowDownIcon, ClockFadingIcon, HistoryIcon, MoreHorizontalIcon, RotateCwIcon } from 'lucide-react'
import { AvatarImage } from '@radix-ui/react-avatar'
import { apiClient } from '~/lib/api';
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuShortcut, DropdownMenuTrigger } from './ui/dropdown-menu'
import { Button } from './ui/button'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast';


const CACHE_TTL = 0.5 * 60 * 60 * 1000; // 30 mins

const setWithExpiry = (key: string, value: string, ttl: number) => {
    const now = Date.now();
    const item = { value, expiry: now + ttl };
    localStorage.setItem(key, JSON.stringify(item));
};

const getWithExpiry = (key: string): string | null => {
    const itemStr = localStorage.getItem(key);
    if (!itemStr) return null;

    try {
        const item = JSON.parse(itemStr);
        if (Date.now() > item.expiry) {
            localStorage.removeItem(key);
            return null;
        }
        return item.value;
    } catch {
        localStorage.removeItem(key);
        return null;
    }
};

export const useCachedImage = (job: components["schemas"]["Job"]) => {
    const [inputImage, setInputImage] = useState<string | null>(null);
    const [outputImages, setOutputImages] = useState<string[]>([]);

    useEffect(() => {
        if (!job?.tasks?.length) return;

        const controller = new AbortController();
        const signal = controller.signal;

        const firstTask = job.tasks[0];
        const inputImageId = firstTask?.input_image?.id;
        const inputCacheKey = inputImageId ? `image_${inputImageId}` : null;

        const outputImageIds =
            job.tasks
                ?.map((task) => task.output_image?.id)
                .filter(Boolean) || [];

        const fetchImage = async (
            jobId: string,
            taskId: string,
            imageId: string,
            setImage: (url: string) => void,
            cacheKey: string
        ) => {
            try {
                // Check cache first
                const cachedUrl = getWithExpiry(cacheKey);
                if (cachedUrl) {
                    setImage(cachedUrl);
                    return;
                }

                const { data } = await apiClient.GET(
                    "/api/v1/jobs/{job_id}/tasks/{task_id}/images/{image_id}",
                    {
                        params: {
                            path: { job_id: jobId, task_id: taskId, image_id: imageId },
                        },
                        signal,
                    }
                );

                const imageUrl = data?.url;
                if (imageUrl) {
                    setImage(imageUrl);
                    setWithExpiry(cacheKey, imageUrl, CACHE_TTL);
                }
            } catch (error) {
                if (signal.aborted) return;
                console.error("Error fetching image:", error);
            }
        };

        // Fetch input image
        if (inputImageId && firstTask?.id) {
            fetchImage(
                job.id,
                firstTask.id,
                inputImageId,
                setInputImage,
                inputCacheKey!
            );
        }

        // Fetch all output images
        const outputPromises = outputImageIds.map(async (imageId, index) => {
            const task = job.tasks?.find((t) => t.output_image?.id === imageId);
            if (!task) return;

            const cacheKey = `image_${imageId}`;
            await fetchImage(job.id, task.id, imageId, (url) => {
                setOutputImages((prev) => {
                    const newImages = [...prev];
                    newImages[index] = url;
                    return newImages;
                });
            }, cacheKey);
        });

        Promise.all(outputPromises);

        return () => controller.abort();
    }, [job]);

    return { inputImage, outputImages };
};


type Props = {
    job: components["schemas"]["Job"]
    onDelete: (jobId: string) => void
}

function HistoryJobCard({ job, onDelete }: Props) {
    const { inputImage, outputImages } = useCachedImage(job);

    const date = new Date(job.created_at); // JS converts to local timezone

    const formatted = date.toLocaleDateString(undefined, { // undefined = use client locale
        weekday: "short", // Sun
        day: "2-digit",  // 05
        month: "short",  // Oct
        year: "numeric"  // 2025
    }).replace(/,/g, ''); // remove comma

    const formattedTime = date.toLocaleTimeString(undefined, {
        hour: "2-digit",
        minute: "2-digit",
        hour12: false, // change to true for 12-hour format
    });

    const router = useRouter();

    const handleDelete = async () => {

        const { error } = await apiClient.DELETE("/api/v1/jobs/{id}", {
            params: {
                path: { id: job.id },
            },
        })

        if (error) {
            // alert(`Failed to delete job: ${error.detail}`);
            toast.error("Failed to delete job: " + error.detail);

        } else {
            // alert("Job deleted successfully!");
            toast.success("Job deleted successfully!");
            onDelete(job.id);

        }
    };

    return (
        <div className="flex w-full items-center gap-4 border-b border-solid px-4 py-4 hover:bg-highlight/10 transition-all duration-300 ease-in-out">
            <Avatar className='rounded-md' >
                <AvatarImage src={inputImage!} />
            </Avatar>
            <div className="flex grow shrink-0 basis-0 flex-col items-start">
                <span className="line-clamp-1 w-full text-body-bold font-body-bold text-foreground">
                    {formatted}
                </span>
                <span className="line-clamp-1 w-full text-caption font-caption text-muted-foreground">
                    {formattedTime}
                </span>
            </div>
            <div className="flex grow shrink-0 basis-0 flex-col items-start justify-center gap-2">
                {(() => {
                    const status = job.tasks?.[0]?.status ?? 'failed';
                    const statusMap: any = {
                        completed: { variant: 'success', icon: <CheckCheckIcon size={14} /> },
                        running: { variant: 'warning', icon: <RotateCwIcon size={14} /> },
                        pending: { variant: 'neutral', icon: <ClockFadingIcon size={14} /> },
                        failed: { variant: 'error', icon: <AlertCircleIcon size={14} /> },
                    };

                    const { variant, icon } = statusMap[status] || statusMap.failed;
                    const label = status.charAt(0).toUpperCase() + status.slice(1).toLowerCase();

                    return (
                        <Badge className="text-caption" variant={variant} icon={icon}>
                            {label}
                        </Badge>
                    );
                })()}
            </div>
            <div className="flex grow shrink-0 basis-0 items-center gap-1">
                {outputImages.map((image, index) => (
                    <Avatar key={index} className='rounded-md'>
                        <AvatarImage src={image} />
                    </Avatar>
                ))}

            </div>

            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="ghost">
                        <MoreHorizontalIcon />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-16" align="start">
                    <DropdownMenuGroup>
                        <DropdownMenuItem onClick={handleDelete}>
                            Delete
                        </DropdownMenuItem>
                    </DropdownMenuGroup>
                </DropdownMenuContent>
            </DropdownMenu>
        </div >
    )
}

export default HistoryJobCard
