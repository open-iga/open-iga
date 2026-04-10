import type { RuntimeApplication } from '@/types';
import type { AppConfig } from '@/common/config-manager.ts';
import Elysia from 'elysia';
import { createHealthRouter } from '@/api/health/health.routes.ts';
import { createLoginRouter } from '@/api/auth/login.router.ts';
import openapi from '@elysiajs/openapi';

export const createRuntimeApi = (args: { runtimeApplication: RuntimeApplication; appConfig: AppConfig }) => {
    const mainRouter = new Elysia();

    mainRouter.use(
        openapi({
            documentation: {
                info: {
                    title: 'Open IGA API',
                    version: '1.0.0',
                },
            },
        }),
    );
    mainRouter.use(createHealthRouter());
    mainRouter.use(createLoginRouter(args));

    return mainRouter;
};
