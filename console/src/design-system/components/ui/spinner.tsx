import { cn } from '@/design-system/lib/utils';
import { Loader2Icon } from 'lucide-react';
import type { ComponentProps } from 'react';

export const Spinner = ({ className, ...props }: ComponentProps<'svg'>) => {
    return <Loader2Icon role="status" className={cn('size-4 animate-spin', className)} {...props} />;
};
