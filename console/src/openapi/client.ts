import createFetchClient from 'openapi-fetch';
import type { paths } from './schema';

export const fetchClient = createFetchClient<paths>({
    baseUrl: '/',
});
