import { cn } from '@/design-system/lib/utils';
import type { ComponentProps } from 'react';

export const Skeleton = ({ className, ...props }: ComponentProps<'div'>) => {
    return <div data-slot="skeleton" className={cn('animate-pulse rounded-xl bg-muted', className)} {...props} />;
};
