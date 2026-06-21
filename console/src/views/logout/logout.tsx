import { Card, CardDescription, CardHeader, CardTitle } from '@/design-system/components/ui/card.tsx';
import { Logo } from '@/design-system/components/icons/logo.tsx';
import { useTranslation } from 'react-i18next';
import { Spinner } from '@/design-system/components/ui/spinner.tsx';
import { useLogout } from './hooks/use-logout.ts';
import { useEffect } from 'react';

export const Logout = () => {
    const { t } = useTranslation();
    const { logout } = useLogout();

    useEffect(() => {
        logout();
    }, []);

    return (
        <div className="h-screen w-full flex items-center justify-center px-4">
            <Card className="w-full max-w-md">
                <CardHeader className="text-center pb-5">
                    <CardTitle className="justify-self-center">
                        <Logo width={200} height={80} />
                    </CardTitle>
                    <CardDescription>
                        <h1 className="text-2xl tracking-tighter text-foreground text-balance mb-4">
                            {t('auth.logout.pending.title')}
                        </h1>
                        <div className="flex flex-row justify-center items-center">
                            {t('auth.logout.pending.description')}
                            <Spinner className="ml-1" />
                        </div>
                    </CardDescription>
                </CardHeader>
            </Card>
        </div>
    );
};
