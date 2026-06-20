import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/design-system/components/ui/dialog.tsx';
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
import { Spinner } from '@/design-system/components/ui/spinner.tsx';

interface AppSidebarFooterProps {
    firstName: string;
    lastName: string;
    onLogout: () => void;
    isLogoutPending: boolean;
}

export const AppSidebarFooter = ({ firstName, lastName, onLogout, isLogoutPending }: AppSidebarFooterProps) => {
    const { t } = useTranslation();

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
                            <DropdownMenuItem className="cursor-pointer" onClick={onLogout}>
                                <LogOut />
                                {t('auth.logout.label')}
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </SidebarMenuItem>
            </SidebarMenu>
            <Dialog open={isLogoutPending}>
                <DialogContent showCloseButton={false}>
                    <DialogHeader>
                        <DialogTitle>{t('auth.logout.pending.title')}</DialogTitle>
                        <DialogDescription className="flex">
                            {t('auth.logout.pending.description')} <Spinner className="ml-2" />
                        </DialogDescription>
                    </DialogHeader>
                </DialogContent>
            </Dialog>
        </SidebarFooter>
    );
};
