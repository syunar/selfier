import React from 'react'
import ClientHistory from '~/components/client-history'
import { apiClient } from '~/lib/api';
import type { components } from '~/lib/api-types';

const History = async () => {
    const { data } = await apiClient.GET("/api/v1/jobs");

    const jobs = (data ?? null) as components["schemas"]["Job"][] | null;

    return (
        <ClientHistory initialJobs={jobs ?? []} />
    )
}

export default History
