import { useEffect, useState } from 'react';
import { cn } from '@/design-system/lib/utils.ts';
import { Logo } from '@/design-system/components/icons/logo.tsx';

export const SplashOverlay = ({ visible }: { visible: boolean }) => {
    const [mounted, setMounted] = useState(true);
    const [exiting, setExiting] = useState(false);

    useEffect(() => {
        if (!visible) {
            const t1 = setTimeout(() => setExiting(true), 800);
            const t2 = setTimeout(() => setMounted(false), 1200);
            return () => {
                clearTimeout(t1);
                clearTimeout(t2);
            };
        }
    }, [visible]);

    if (!mounted) return null;

    return (
        <div
            className={cn(
                'fixed inset-0 z-50 flex items-center justify-center bg-background transition-opacity duration-600 ease-in-out',
                exiting ? 'opacity-0' : 'opacity-100',
            )}
        >
            <Logo width={200} height={80} className="animate-in fade-in zoom-in-95 duration-500" />
        </div>
    );
};
