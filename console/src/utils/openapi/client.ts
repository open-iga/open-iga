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

// API errors are { message: string }; return that, falling back to the raw body.
const parseErrorMessage = (body: string): string | undefined => {
    try {
        const parsed = JSON.parse(body) as { message?: string };
        return parsed.message ?? (body || undefined);
    } catch {
        return body || undefined;
    }
};

// throw error so that tanstack query error is invoked
const throwOnErrorMiddleware: Middleware = {
    onResponse: async ({ response }) => {
        if (!response.ok) {
            const body = await response.clone().text();
            const message = parseErrorMessage(body) ?? `HTTP status ${response.status}: ${response.statusText}`;
            throw new Error(message);
        }

        return undefined;
    },
};

export const fetchClient = createFetchClient<paths>({
    baseUrl: '/',
});

fetchClient.use(throwOnErrorMiddleware, authMiddleware);
