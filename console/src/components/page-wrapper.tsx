import { Fragment, type ReactNode } from 'react';
import { cn } from '@/design-system/lib/utils.ts';
import { Link, useLocation } from '@tanstack/react-router';
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbSeparator,
} from '@/design-system/components/ui/breadcrumb.tsx';
import { capitalize, compact } from 'lodash';

type PageWrapperProps = {
    children: ReactNode;
    className?: string;
    title: string;
    description: string;
};

const Breadcrumbs = () => {
    const { pathname } = useLocation();
    const paths = compact(pathname.split('/'));

    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <BreadcrumbLink render={<Link to="/" />}>Home</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                {paths.map((path, i) => (
                    <Fragment key={path}>
                        <BreadcrumbItem>
                            {/* @ts-ignore: error can be ignored as only the valida path can land here */}
                            <BreadcrumbLink render={i !== paths.length - 1 ? <Link to={`/${path}`} /> : undefined}>
                                {path.split('-').map(capitalize).join(' ')}
                            </BreadcrumbLink>
                        </BreadcrumbItem>
                        {i !== paths.length - 1 && <BreadcrumbSeparator />}
                    </Fragment>
                ))}
            </BreadcrumbList>
        </Breadcrumb>
    );
};

export const PageWrapper = ({ children, className, title, description }: PageWrapperProps) => (
    <div className={cn('pl-5 pt-5 pr-5 flex flex-col gap-4', className)}>
        <Breadcrumbs />
        <div className="flex flex-col gap-1">
            <h1 className="text-xl font-semibold">{title}</h1>
            <p className="max-w-3xl text-sm text-muted-foreground">{description}</p>
        </div>
        {children}
    </div>
);
