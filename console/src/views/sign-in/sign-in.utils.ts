import type { SupportedOauthProvider } from '@/views/types';

export const supportedProviders: SupportedOauthProvider[] = ['google'];

export const isSupportedOauthProvider = (provider: string): provider is SupportedOauthProvider =>
    supportedProviders.includes(provider as SupportedOauthProvider);
