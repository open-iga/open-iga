import { useTranslation } from 'react-i18next';
import {
    SidebarFooter,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from '@/design-system/components/ui/sidebar.tsx';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/design-system/components/ui/dropdown-menu.tsx';
import { Avatar, AvatarFallback } from '@/design-system/components/ui/avatar.tsx';
import { ChevronsUpDown, LogOut } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';

interface AppSidebarFooterProps {
    firstName: string;
    lastName: string;
}

export const AppSidebarFooter = ({ firstName, lastName }: AppSidebarFooterProps) => {
    const { t } = useTranslation();
    const navigate = useNavigate();
    const logout = () => navigate({ to: '/auth/logout' });

    return (
        <SidebarFooter>
            <SidebarMenu>
                <SidebarMenuItem>
                    <DropdownMenu>
                        <DropdownMenuTrigger
                            render={
                                <SidebarMenuButton size="lg">
                                    <Avatar>
                                        <AvatarFallback className="bg-primary text-foreground">
                                            {`${firstName.at(0)}${lastName.at(0)}`}
                                        </AvatarFallback>
                                    </Avatar>
                                    {firstName} {lastName}
                                    <ChevronsUpDown aria-disabled className="ml-auto" />
                                </SidebarMenuButton>
                            }
                        />
                        <DropdownMenuContent side="top">
                            <DropdownMenuItem className="cursor-pointer" onClick={logout}>
                                <LogOut />
                                {t('auth.logout.label')}
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </SidebarMenuItem>
            </SidebarMenu>
        </SidebarFooter>
    );
};
