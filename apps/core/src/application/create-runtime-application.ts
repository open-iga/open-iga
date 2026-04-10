import { OauthService } from '@/application/auth/oauth-service.ts';
import type { RuntimeApplication, RuntimeRemotes } from '@/types';

export const createRuntimeApplication = (args: { runtimeRemotes: RuntimeRemotes }): RuntimeApplication => {
    const { runtimeRemotes } = args;

    return { oauthService: new OauthService(runtimeRemotes.oauthProviders) };
};
