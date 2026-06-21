import { AppSidebar } from '@/components/sidebar/app-sidebar.tsx';
import { SidebarTrigger } from '@/design-system/components/ui/sidebar.tsx';
import { Favicon } from '@/design-system/components/icons/favicon.tsx';

interface SidebarContainerProps {
    firstName: string;
    lastName: string;
}

export const SidebarContainer = ({ firstName, lastName }: SidebarContainerProps) => {
    return (
        <>
            <AppSidebar firstName={firstName} lastName={lastName} />
            <div className="fixed flex m-3 p-2 items-center rounded-xl bg-muted">
                <Favicon size={32} />
                <SidebarTrigger className="ml-2" />
            </div>
        </>
    );
};
