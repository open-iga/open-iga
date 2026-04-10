import type { OauthUser } from '@/domain/user/oauth-user.ts';
import { SUPPORTED_OAUTH_PROVIDERS } from '@/application/auth/oauth-service.ts';

type OauthProviders = (typeof SUPPORTED_OAUTH_PROVIDERS)[number];

export type OauthProvidersAdapter = Record<OauthProviders, OauthProviderAdapter>;

export interface OauthProviderAdapter {
    getAuthorizationUrl(args: {
        /**
         * URI to which the OAUTH provider should redirect the user for login
         * */
        redirectUri: string;
    }): Promise<{ redirectUrl: URL; state: string; codeVerifier: string }>;

    performTokenExchange(args: { codeVerifier: string; code: string; redirectUri: string }): Promise<string | null>;

    getOauthUser(args: { accessToken: string }): Promise<OauthUser | null>;
}
