import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createRootRoute, Outlet } from '@tanstack/react-router';
import { I18nextProvider } from 'react-i18next';
import { i18next } from '@/utils/i18n-setup';
import { Toaster } from 'sonner';
import { CircleCheckIcon, InfoIcon, Loader2Icon, OctagonXIcon, TriangleAlertIcon } from 'lucide-react';

const queryClient = new QueryClient();

const RootLayout = () => (
    <I18nextProvider i18n={i18next}>
        <QueryClientProvider client={queryClient}>
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
                        toast: '!bg-popover !border-border',
                        title: '!text-popover-foreground',
                        description: '!text-popover-foreground',
                    },
                }}
                visibleToasts={10}
            />
            <Outlet />
        </QueryClientProvider>
    </I18nextProvider>
);

export const Route = createRootRoute({ component: RootLayout });
