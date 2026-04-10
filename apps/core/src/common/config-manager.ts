import { type } from 'arktype';

const appConfig = type({
    'PORT?': 'string',
    NODE_ENV: 'string',
    GOOGLE_OAUTH_CLIENT_ID: 'string',
    GOOGLE_OAUTH_CLIENT_SECRET: 'string',
    BASE_URL: 'string',
}).pipe((config) => ({
    port: config.PORT ?? '3000',
    env: config.NODE_ENV,
    baseUrl: config.BASE_URL,
    oauth: {
        google: {
            clientId: config.GOOGLE_OAUTH_CLIENT_ID,
            clientSecret: config.GOOGLE_OAUTH_CLIENT_SECRET,
        },
    },
}));

export type AppConfig = typeof appConfig.inferOut;

export const getAppConfig = (): AppConfig => {
    const parsedConfig = appConfig(process.env);
    if (parsedConfig instanceof type.errors) {
        throw new TypeError(`Invalid app configuration: ${parsedConfig.summary}`);
    }

    return parsedConfig;
};
