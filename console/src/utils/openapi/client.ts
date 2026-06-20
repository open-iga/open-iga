import createFetchClient, { type Middleware } from 'openapi-fetch';
import type { paths } from './schema';

const authMiddleware: Middleware = {
    onResponse: async ({ response }) => {
        if (response.status === 401) {
            globalThis.location.href = '/auth/sign-in';
        }

        return undefined;
    },
};

// throw error so that tanstack query error is invoked
const throwOnErrorMiddleware: Middleware = {
    onResponse: async ({ response }) => {
        if (!response.ok) {
            const body = await response.clone().text();
            throw new Error(body || `HTTP status ${response.status}: ${response.statusText}`);
        }

        return undefined;
    },
};

export const fetchClient = createFetchClient<paths>({
    baseUrl: '/',
});

fetchClient.use(throwOnErrorMiddleware, authMiddleware);
