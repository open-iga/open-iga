import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from '@/design-system/components/ui/card.tsx';
import { TextBanner } from '@/design-system/components/ui/text-banner.tsx';
import { supportedProviders } from '@/views/sign-in/sign-in.utils.ts';
import { Button } from '@/design-system/components/ui/button.tsx';
import { cn } from '@/design-system/lib/utils.ts';
import { GoogleIcon } from '@/design-system/components/icons/google.tsx';
import Markdown, { type Components } from 'react-markdown';
import { useTranslation } from 'react-i18next';
import type { SupportedOauthProvider } from '@/views/types';
import { useCallback } from 'react';
import { Spinner } from '@/design-system/components/ui/spinner.tsx';

type SignInProps =
    | {
          disableProviderSelection: false;
          onProviderSelection: (providerSelection: SupportedOauthProvider) => void;
      }
    | { disableProviderSelection: true; providerToHighlight?: SupportedOauthProvider };

const Link: Components['a'] = ({ children, href }) => (
    <a href={href} target="_blank" rel="noopener noreferrer" className="underline text-primary">
        {children}
    </a>
);

export const SignIn = (args: SignInProps) => {
    const { t } = useTranslation();
    const { disableProviderSelection } = args;

    const handleProviderSelection = useCallback((providerSelection: SupportedOauthProvider) => {
        if ('onProviderSelection' in args) {
            args.onProviderSelection(providerSelection);
        }
    }, []);

    return (
        <div className="h-screen w-full flex items-center justify-center px-4">
            <Card className="w-full max-w-md">
                <CardHeader className="text-center pb-5">
                    <CardTitle>
                        <TextBanner className="pb-10" />
                        <h1 className="text-2xl tracking-tighter text-balance">{t('auth.signIn.title')}</h1>
                    </CardTitle>
                    <CardDescription>{t('auth.signIn.description')}</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col items-center gap-4 text-center">
                    {supportedProviders.map((provider) => (
                        <Button
                            variant="outline"
                            disabled={disableProviderSelection}
                            key={provider}
                            onClick={() => handleProviderSelection(provider)}
                            className={cn('w-full justify-center gap-2')}
                        >
                            <GoogleIcon /> {t(`auth.provider.${provider}`)}{' '}
                            {'providerToHighlight' in args && args.providerToHighlight === provider && <Spinner />}
                        </Button>
                    ))}
                    <span className="text-xs text-muted-foreground">{t('auth.provider.moreProviders')}</span>
                </CardContent>
                <CardFooter className="text-xs text-muted-foreground">
                    <Markdown components={{ a: Link }}>{t('termsOfUse.description')}</Markdown>
                </CardFooter>
            </Card>
        </div>
    );
};
