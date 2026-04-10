import type { OauthService } from '@/application/auth/oauth-service.ts';

export type RuntimeApplication = {
    oauthService: OauthService;
};

export type RuntimeRemotes = {
    oauthProviders: OauthProvidersAdapter;
};
