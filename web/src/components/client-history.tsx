"use client"
import { ArrowUpRightIcon, CameraIcon, CloudUploadIcon, FolderInputIcon, FolderPlusIcon, HistoryIcon, LucideCamera, PlusIcon, Smartphone, WandSparklesIcon } from 'lucide-react'
import React, { useEffect, useRef, useState } from 'react'
import { ImageCarousel } from './image-carousel';
import StarIcon from './feather-sparkles';
import { Button } from './ui/button';
import { Badge } from './badge';
import { redirect, useRouter } from 'next/navigation';
import type { components } from '~/lib/api-types';
import HistoryJobCard from './history-job';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from './ui/empty';
import { apiClient } from '~/lib/api';

const TERMINAL_STATUSES = ['completed', 'failed'];

type Job = components["schemas"]["Job"];

type Props = {
    initialJobs: Job[]
}

function ClientHistory({ initialJobs }: Props) {
    const [jobs, setJobs] = useState<Job[]>(initialJobs);
    const timeoutRef = useRef<NodeJS.Timeout | null>(null);
    const router = useRouter();

    useEffect(() => {
        const fetchAndUpdateJobs = async () => {
            // Check if there are any jobs that are still processing.
            const hasActiveJobs = jobs.some(job =>
                !job.tasks?.every(task => TERMINAL_STATUSES.includes(task.status))
            );

            // 3. Efficiency check: If no jobs are active, stop polling.
            if (!hasActiveJobs) {
                console.log("All jobs have completed. Stopping polling.");
                return;
            }

            // Fetch the latest job statuses from the API.
            console.log("Fetching updated job statuses...");
            const { data, error } = await apiClient.GET("/api/v1/jobs");

            if (data) {
                setJobs(data as Job[]);
            } else if (error) {
                console.error("Failed to fetch jobs:", error);
            }

            // 4. Schedule the next poll in 5 seconds.
            timeoutRef.current = setTimeout(fetchAndUpdateJobs, 5000); // 5000ms = 5s
        };

        // Start the first poll after a short delay.
        timeoutRef.current = setTimeout(fetchAndUpdateJobs, 5000);

        // 5. Cleanup function: This is essential to prevent memory leaks.
        // It runs when the component unmounts.
        return () => {
            if (timeoutRef.current) {
                clearTimeout(timeoutRef.current);
            }
        };
    }, [jobs]); // The effect re-runs whenever the `jobs` state changes.

    const handleDelete = (jobID: string) => {
        setJobs(jobs.filter((job) => job.id !== jobID));
    }

    return (
        <div className="flex h-full w-full flex-col items-center gap-4 rounded-md bg-background-dark px-6 py-12">
            <div className="flex w-full max-w-[1024px] grow shrink-0 basis-0 flex-col items-center gap-4 rounded-md border border-solid bg-background px-6 py-6 shadow-sm border-t-highlight bg-gradient-to-b from-highlight/15 to-background ">
                <div className="flex w-full items-center gap-2">
                    <div className="flex grow shrink-0 basis-0 items-center gap-4">

                        <div className="flex items-center justify-center w-8 h-8 bg-background rounded-md border-t-highlight border border-solid bg-gradient-to-b from-highlight/15 to-background">
                            <HistoryIcon size={16} />
                        </div>

                        <span className="text-heading-3 font-heading-3 text-foreground">
                            History
                        </span>
                    </div>

                    <Button variant="default" size="sm" onClick={(event: React.MouseEvent<HTMLButtonElement>) => { router.push("/") }}>
                        <PlusIcon size={16} />
                        <span>New</span>
                    </Button>
                </div>

                {jobs?.length === 0 && (
                    <Empty className='-mt-10'>
                        <EmptyHeader>
                            <EmptyMedia variant="icon">
                                <FolderPlusIcon />
                            </EmptyMedia>
                            <EmptyTitle>No Tasks Yet</EmptyTitle>
                            <EmptyDescription>
                                You haven&apos;t created any tasks yet. Get started by creating
                                your first tasks.
                            </EmptyDescription>
                        </EmptyHeader>
                        <EmptyContent>
                            <div className="flex gap-2">
                                <Button variant="default" size="sm" onClick={(event: React.MouseEvent<HTMLButtonElement>) => { router.push("/") }}>
                                    <PlusIcon size={16} />
                                    <span>New</span>
                                </Button>
                            </div>
                        </EmptyContent>

                    </Empty>
                )}
                <div className="flex w-full flex-col items-start">
                    {jobs?.map((job) => (
                        <HistoryJobCard key={job.id} job={job} onDelete={handleDelete} />
                    ))}
                </div>


            </div>
        </div>
    )
}

export default ClientHistory
