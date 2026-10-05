import { useTranslation } from 'react-i18next';
import type { FileRouteTypes } from '@/routeTree.gen.ts';
import type { ForwardRefExoticComponent, RefAttributes } from 'react';
import { Cable, type LucideProps } from 'lucide-react';
import {
    SidebarContent,
    SidebarGroup,
    SidebarGroupContent,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from '@/design-system/components/ui/sidebar.tsx';
import { Link, useLocation } from '@tanstack/react-router';
import { cn } from '@/design-system/lib/utils.ts';

type SidebarContent = {
    link: FileRouteTypes['to'];
    label: string;
    icon: ForwardRefExoticComponent<Omit<LucideProps, 'ref'> & RefAttributes<SVGSVGElement>>;
    isActive: boolean;
};

const useSidebarContent = (): SidebarContent[] => {
    const { t } = useTranslation();
    const { pathname } = useLocation();

    return [
        {
            link: '/managed-systems',
            label: t('managedSystems.name'),
            icon: Cable,
            isActive: pathname.startsWith('/managed-systems'),
        },
    ];
};

export const AppSidebarContent = () => {
    const sidebarContent = useSidebarContent();

    return (
        <SidebarContent>
            <SidebarGroup>
                <SidebarGroupContent>
                    <SidebarMenu>
                        {sidebarContent.map(({ label, link, icon: Icon, isActive }) => (
                            <SidebarMenuItem key={label}>
                                <SidebarMenuButton
                                    size="lg"
                                    className={cn(
                                        'group-data-[collapsible=icon]:justify-center',
                                        isActive ? 'bg-sidebar-accent' : '',
                                    )}
                                    render={<Link to={link} />}
                                >
                                    <Icon />
                                    <span className="group-data-[collapsible=icon]:hidden">{label}</span>
                                </SidebarMenuButton>
                            </SidebarMenuItem>
                        ))}
                    </SidebarMenu>
                </SidebarGroupContent>
            </SidebarGroup>
        </SidebarContent>
    );
};
