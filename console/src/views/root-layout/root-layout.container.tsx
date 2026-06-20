import { type ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { i18next } from '@/utils/i18n-setup.ts';
import { TooltipProvider } from '@/design-system/components/ui/tooltip.tsx';
import { Toaster } from 'sonner';
import { CircleCheckIcon, InfoIcon, Loader2Icon, OctagonXIcon, TriangleAlertIcon } from 'lucide-react';
import { I18nextProvider, useTranslation } from 'react-i18next';
import { useLocation } from '@tanstack/react-router';
import { SidebarProvider } from '@/design-system/components/ui/sidebar.tsx';
import { ErrorDetails } from '@/components/error-boundary.tsx';
import { SplashOverlay } from './splash-overlay.tsx';
import { SidebarContainer } from '@/views/root-layout/sidebar.container.tsx';
import { useCurrentUser } from './hooks/use-current-user.ts';

const queryClient = new QueryClient();

const toaster = (
    <Toaster
        closeButton
        icons={{
            success: <CircleCheckIcon className="size-4 text-primary" />,
            info: <InfoIcon className="size-4 text-primary" />,
            warning: <TriangleAlertIcon className="size-4 text-warning" />,
            error: <OctagonXIcon className="size-4 text-destructive" />,
            loading: <Loader2Icon className="size-4 text-primary animate-spin" />,
        }}
        toastOptions={{
            classNames: {
                toast: 'bg-popover! border-border!',
                title: 'text-popover-foreground!',
                description: 'text-popover-foreground!',
            },
        }}
        visibleToasts={10}
    />
);

const ProtectedLayoutContainer = ({ children }: { children: ReactNode }) => {
    const { t } = useTranslation();
    const { isPending, isError, error, firstName, lastName } = useCurrentUser();

    if (isError) {
        return <ErrorDetails summary={t('error.generic')} details={error?.message ?? ''} />;
    }

    return (
        <>
            <SplashOverlay visible={isPending} />
            <SidebarProvider>
                <SidebarContainer firstName={firstName} lastName={lastName} />
                <main className="p-20 w-screen h-screen">{children}</main>
            </SidebarProvider>
        </>
    );
};

export const RootLayoutContainer = ({ children }: { children: ReactNode }) => {
    const location = useLocation();
    const isProtected = !location.href.startsWith('/auth');

    return (
        <I18nextProvider i18n={i18next}>
            <QueryClientProvider client={queryClient}>
                <TooltipProvider>
                    {toaster}
                    {isProtected ? <ProtectedLayoutContainer>{children}</ProtectedLayoutContainer> : children}
                </TooltipProvider>
            </QueryClientProvider>
        </I18nextProvider>
    );
};
