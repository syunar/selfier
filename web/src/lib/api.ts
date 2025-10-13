import createClient from "openapi-fetch";
import type { paths } from "./api-types"; // your generated file

export const apiClient = createClient<paths>({
    baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080", // e.g. https://api.example.com
});
