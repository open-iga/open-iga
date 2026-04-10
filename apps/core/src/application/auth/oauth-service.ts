import type { OauthProvidersAdapter } from '../adapters/oauth-provider.adapter.ts';

export const SUPPORTED_OAUTH_PROVIDERS = ['google'] as const;

export class OauthService {
    constructor(private readonly oauthProviders: OauthProvidersAdapter) {}

    isSupportedProvider = (provider: string): provider is (typeof SUPPORTED_OAUTH_PROVIDERS)[number] =>
        SUPPORTED_OAUTH_PROVIDERS.includes(provider as (typeof SUPPORTED_OAUTH_PROVIDERS)[number]);

    getAuthorizationUrl = async (provider: (typeof SUPPORTED_OAUTH_PROVIDERS)[number], args: { redirectUri: string }) =>
        this.oauthProviders[provider].getAuthorizationUrl(args);

    createSession = async ({
        code,
        codeVerifier,
        redirectUri,
        provider,
    }: {
        code: string;
        codeVerifier: string;
        redirectUri: string;
        provider: (typeof SUPPORTED_OAUTH_PROVIDERS)[number];
    }) => {
        const oauthProvider = this.oauthProviders[provider];
        const accessToken = await oauthProvider.performTokenExchange({ code, codeVerifier, redirectUri });

        if (!accessToken) {
            return null;
        }

        return await oauthProvider.getOauthUser({ accessToken });
    };
}
