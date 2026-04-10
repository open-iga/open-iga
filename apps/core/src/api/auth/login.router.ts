import Elysia from 'elysia';
import { type } from 'arktype';
import type { AppConfig } from '@/common/config-manager.ts';
import type { RuntimeApplication } from '@/types';

const AUTH_STATE_COOKIE_NAME = 'authState';

const paramsType = type({
    provider: "'google'",
});

const authStateCookieType = type({ codeVerifier: 'string', state: 'string' });

export const createLoginRouter = ({
    appConfig,
    runtimeApplication: { oauthService },
}: {
    appConfig: AppConfig;
    runtimeApplication: RuntimeApplication;
}) => {
    const app = new Elysia({ prefix: '/login' });
    app.get(
        '/:provider',
        async ({ params: { provider }, cookie, redirect }) => {
            const redirectUri = `${appConfig.baseUrl}/login/${provider}/callback`;
            const { redirectUrl, codeVerifier, state } = await oauthService.getAuthorizationUrl(provider, {
                redirectUri,
            });

            cookie.authState?.set({
                value: { codeVerifier, state },
                httpOnly: true,
                maxAge: 300,
                secure: appConfig.env === 'production',
                sameSite: 'lax',
            });

            return redirect(redirectUrl.toString());
        },
        {
            params: paramsType,
            cookie: type({
                '[AUTH_STATE_COOKIE_NAME]?': authStateCookieType.pipe((v) => JSON.stringify(v)),
            }),
            response: {
                302: type({}),
            },
        },
    ).get(
        '/:provider/callback',
        async ({ cookie: { authState }, query: { code, state }, params: { provider }, status }) => {
            if (state !== authState.value.state) {
                return status(422, { message: 'Invalid request. State mismatch' });
            }

            const redirectUri = `${appConfig.baseUrl}/login/${provider}/callback`;
            const session = await oauthService.createSession({
                code,
                codeVerifier: authState.value.codeVerifier,
                redirectUri,
                provider,
            });

            if (!session) {
                return status(422, { message: 'Unable to create a session' });
            }

            return session;
        },
        {
            params: paramsType,
            query: type({
                code: 'string',
                state: 'string',
            }),
            cookie: type({
                [AUTH_STATE_COOKIE_NAME]: authStateCookieType,
            }),
            response: {
                200: type({
                    firstName: 'string',
                    lastName: 'string',
                    email: 'string',
                }),
                422: type({
                    message: 'string',
                }),
            },
        },
    );

    return app;
};
