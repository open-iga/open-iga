import { setupServer } from 'msw/node';
import { createOpenApiHttp } from 'openapi-msw';
import type { paths } from '@/utils/openapi/schema';

export const mockHttpHandlers = createOpenApiHttp<paths>({ baseUrl: 'http://localhost' });
export const mockServer = setupServer();
