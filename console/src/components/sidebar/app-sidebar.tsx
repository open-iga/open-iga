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
                <div className="relative group/favicon">
                    <span className="group-hover/favicon:opacity-0">
                        <Favicon size={32} />
                    </span>
                    <SidebarTrigger
                        icon={<PanelLeftOpen />}
                        className="pointer-events-none absolute inset-0 flex opacity-0 focus-visible:pointer-events-auto focus-visible:opacity-100 group-hover/favicon:pointer-events-auto group-hover/favicon:opacity-100"
                    />
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
