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
import { Link } from '@tanstack/react-router';
import { useCurrentUser } from '@/hooks/use-current-user.ts';

export const AppSidebarFooter = () => {
    const { t } = useTranslation();
    const { firstName, lastName } = useCurrentUser();

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
                            <DropdownMenuItem render={<Link to="/auth/logout" />} className="cursor-pointer">
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
