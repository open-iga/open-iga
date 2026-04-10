import type { AppConfig } from '@/common/config-manager.ts';
import { generateCodeChallenge, generateRandomValues } from './oauth-provider.utils.ts';
import { logger } from '@/common/logger.ts';
import { match, P } from 'ts-pattern';
import type { OauthUser } from '@/domain/user/oauth-user.ts';
import type { OauthProviderAdapter } from '@/application/adapters/oauth-provider.adapter.ts';
import { safeFetch } from '@/common/fetch.ts';
import { type } from 'arktype';

const GOOGLE_AUTHORIZATION_ENDPOINT = 'https://accounts.google.com/o/oauth2/v2/auth';
const GOOGLE_TOKEN_ENDPOINT = 'https://oauth2.googleapis.com/token';
const GOOGLE_USERINFO_ENDPOINT = 'https://www.googleapis.com/oauth2/v3/userinfo';
const SCOPE = ['openid', 'email', 'profile'].join(' ');

const tokenExchangeResponse = type({
    error: 'string',
    error_description: 'string',
}).or({ access_token: 'string' });

const userInfoResponse = type({
    family_name: 'string',
    given_name: 'string',
    email: 'string',
});

export class GoogleOAuthProvider implements OauthProviderAdapter {
    private readonly oauthConfig: AppConfig['oauth']['google'];

    constructor(private readonly appConfig: AppConfig) {
        this.oauthConfig = this.appConfig.oauth.google;
    }

    getAuthorizationUrl = async (args: {
        redirectUri: string;
    }): Promise<{ redirectUrl: URL; state: string; codeVerifier: string }> => {
        // Prevent CSRF
        const state = generateRandomValues(16);
        // Prevent Code Theft
        const codeVerifier = generateRandomValues(32);
        const codeChallenge = await generateCodeChallenge(codeVerifier);

        const redirectUrl = new URL(GOOGLE_AUTHORIZATION_ENDPOINT);
        redirectUrl.searchParams.set('client_id', this.oauthConfig.clientId);
        redirectUrl.searchParams.set('response_type', 'code');
        redirectUrl.searchParams.set('redirect_uri', args.redirectUri);
        redirectUrl.searchParams.set('scope', SCOPE);
        redirectUrl.searchParams.set('state', state);
        redirectUrl.searchParams.set('code_challenge', codeChallenge);
        redirectUrl.searchParams.set('code_challenge_method', 'S256');

        return { redirectUrl, codeVerifier, state };
    };

    performTokenExchange = async (args: {
        code: string;
        codeVerifier: string;
        redirectUri: string;
    }): Promise<string | null> => {
        const { code, codeVerifier, redirectUri } = args;
        const parsedResponse = await safeFetch({
            endpoint: GOOGLE_TOKEN_ENDPOINT,
            init: {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: new URLSearchParams({
                    grant_type: 'authorization_code',
                    client_id: this.oauthConfig.clientId,
                    client_secret: this.oauthConfig.clientSecret,
                    redirect_uri: redirectUri,
                    code,
                    code_verifier: codeVerifier,
                }),
            },
            typeChecker: tokenExchangeResponse,
        });

        return match(parsedResponse)
            .with({ errorType: 'validationError' }, ({ error }) => {
                logger.error(`Failed to parsed token from ${GOOGLE_TOKEN_ENDPOINT}. Reason: ${error}`);

                return null;
            })
            .with({ errorType: 'fetchError' }, ({ error }) => {
                logger.error(`Fetch error with token exchange with ${GOOGLE_TOKEN_ENDPOINT}. Reason: ${error}`);

                return null;
            })
            .with({ response: { error: P.string } }, ({ response }) => {
                logger.error(
                    `Failed to perform token exchange with ${GOOGLE_TOKEN_ENDPOINT}. Reason: ${response.error}. ${response.error_description}`,
                );

                return null;
            })
            .with({ response: { access_token: P.string } }, ({ response }) => response.access_token)
            .exhaustive();
    };

    getOauthUser = async ({ accessToken }: { accessToken: string }): Promise<OauthUser | null> => {
        const parsedResponse = await safeFetch({
            endpoint: GOOGLE_USERINFO_ENDPOINT,
            init: {
                headers: {
                    authorization: `Bearer ${accessToken}`,
                },
            },
            typeChecker: userInfoResponse,
        });

        return match(parsedResponse)
            .with({ errorType: 'validationError' }, ({ error }) => {
                logger.error(`Failed to parse user info from ${GOOGLE_USERINFO_ENDPOINT}. Reason: ${error}`);

                return null;
            })
            .with({ errorType: 'fetchError' }, ({ error }) => {
                logger.error(`Failed to fetch user infor from ${GOOGLE_USERINFO_ENDPOINT}. Reason: ${error}`);

                return null;
            })
            .with(
                { response: { family_name: P.string } },
                ({ response }) =>
                    ({
                        firstName: response.given_name,
                        lastName: response.family_name,
                        email: response.email,
                    }) satisfies OauthUser,
            )
            .exhaustive();
    };
}
