import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

export function EmptyState({
  art,
  title,
  children,
  action,
  className,
}: {
  art?: ReactNode;
  title: ReactNode;
  children?: ReactNode;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col items-center px-6 py-14 text-center animate-rise", className)}>
      {art}
      <h3 className="mt-4 text-md font-semibold tracking-[-0.01em] text-fg">{title}</h3>
      {children ? <p className="mt-1.5 max-w-[340px] text-sm text-fg-3">{children}</p> : null}
      {action ? <div className="mt-5 flex items-center gap-2">{action}</div> : null}
    </div>
  );
}

/** A failed load, with a way to try again. */
export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="mx-auto mt-10 flex max-w-[380px] flex-col items-center rounded-xl bg-surface px-6 py-8 text-center shadow-card">
      <p className="text-sm font-medium text-fg">Couldn’t load this</p>
      <p className="mt-1 text-sm text-fg-3">{message}</p>
      {onRetry ? (
        <button
          type="button"
          onClick={onRetry}
          className="mt-4 rounded-md px-3 py-1.5 text-sm font-medium text-accent-ink hover:bg-accent-soft focus-visible:outline-2 focus-visible:outline-ring"
        >
          Try again
        </button>
      ) : null}
    </div>
  );
}
