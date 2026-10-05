import { Sidebar, SidebarHeader, SidebarTrigger, useSidebar } from '@/design-system/components/ui/sidebar.tsx';
import { Logo } from '@/design-system/components/icons/logo.tsx';
import { AppSidebarFooter } from './app-sidebar-footer.tsx';
import { AppSidebarContent } from '@/components/sidebar/app-sidebar-content.tsx';
import { Favicon } from '@/design-system/components/icons/favicon.tsx';
import { Link } from '@tanstack/react-router';
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';

const AppSidebarHeader = () => {
    const { open } = useSidebar();

    return (
        <SidebarHeader className="flex flex-row items-center">
            {open ? (
                <>
                    <Link to="/">
                        <Logo width={150} height={50} />
                    </Link>
                    <SidebarTrigger icon={<PanelLeftClose />} />
                </>
            ) : (
                <div className="group/favicon">
                    <span className="group-hover/favicon:hidden">
                        <Favicon size={32} />
                    </span>
                    <SidebarTrigger icon={<PanelLeftOpen />} className="hidden group-hover/favicon:flex" />
                </div>
            )}
        </SidebarHeader>
    );
};

export const AppSidebar = () => {
    return (
        <Sidebar collapsible="icon">
            <AppSidebarHeader />
            <AppSidebarContent />
            <AppSidebarFooter />
        </Sidebar>
    );
};
