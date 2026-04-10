import Elysia from 'elysia';
import { type } from 'arktype';

export const createHealthRouter = () =>
    new Elysia({ prefix: '/health' }).get('/', () => ({ status: 'ok' }), { response: type({ status: 'string' }) });
