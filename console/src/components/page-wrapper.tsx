import type { ReactNode } from 'react';

type PageWrapperProps = {
    children: ReactNode;
};

export const PageWrapper = ({ children }: PageWrapperProps) => {
    return <div className="pl-5 pt-5">{children}</div>;
};
