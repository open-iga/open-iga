import { AppSidebar } from '@/components/sidebar/app-sidebar.tsx';
import { useLogout } from './hooks/use-logout.ts';
import { SidebarTrigger, useSidebar } from '@/design-system/components/ui/sidebar.tsx';
import { Favicon } from '@/design-system/components/icons/favicon.tsx';

interface SidebarContainerProps {
    firstName: string;
    lastName: string;
}

export const SidebarContainer = ({ firstName, lastName }: SidebarContainerProps) => {
    const { logout, isPending } = useLogout();
    const { open } = useSidebar();

    return open ? (
        <AppSidebar firstName={firstName} lastName={lastName} onLogout={logout} isLogoutPending={isPending} />
    ) : (
        <div className="fixed flex m-3 p-2 items-center rounded-xl bg-muted">
            <Favicon size={32} />
            <div className="ml-2">
                <SidebarTrigger />
            </div>
        </div>
    );
};
