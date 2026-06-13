import createFetchClient from 'openapi-fetch';
import type { paths } from '../schema';

export const fetchClient = createFetchClient<paths>({
    baseUrl: 'http://localhost',
    // lazy load fetch so it returns the fetch patched by msw
    fetch: (input, init?: RequestInit) => globalThis.fetch(input, init),
});
