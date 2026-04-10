import type { AppConfig } from '@/common/config-manager.ts';
import { GoogleOAuthProvider } from '@/remote/oauth-providers/google.ts';
import type { RuntimeRemotes } from '@/types';

export const createRuntimeRemotes = (args: { appConfig: AppConfig }): RuntimeRemotes => ({
    oauthProviders: {
        google: new GoogleOAuthProvider(args.appConfig),
    },
});
