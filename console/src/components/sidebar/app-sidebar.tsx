import {
    Sidebar,
    SidebarContent,
    SidebarHeader,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarTrigger,
} from '@/design-system/components/ui/sidebar.tsx';
import { Logo } from '@/design-system/components/icons/logo.tsx';
import { AppSidebarFooter } from './app-sidebar-footer.tsx';
import { Construction } from 'lucide-react';

interface AppSidebarProps {
    firstName: string;
    lastName: string;
    onLogout: () => void;
    isLogoutPending: boolean;
}

export const AppSidebar = ({ firstName, lastName, onLogout, isLogoutPending }: AppSidebarProps) => {
    return (
        <Sidebar>
            <SidebarHeader className="flex flex-row items-center">
                <Logo width={150} height={50} />
                <SidebarTrigger />
            </SidebarHeader>
            <SidebarContent>
                <SidebarMenuItem>
                    <SidebarMenuButton size="lg">
                        <Construction />
                        In Progress
                    </SidebarMenuButton>
                </SidebarMenuItem>
            </SidebarContent>
            <AppSidebarFooter
                firstName={firstName}
                lastName={lastName}
                onLogout={onLogout}
                isLogoutPending={isLogoutPending}
            />
        </Sidebar>
    );
};
