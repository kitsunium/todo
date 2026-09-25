import { Tabs as T } from "radix-ui";
import type { ComponentProps, ReactNode } from "react";
import { cn } from "../../lib/cn";

export const Tabs = T.Root;
export const TabsContent = T.Content;

/** A row of tabs; on a narrow screen it scrolls sideways rather than wrap a label. */
export function TabsList({ className, ...rest }: ComponentProps<typeof T.List>) {
  return (
    <div className={cn("no-scrollbar overflow-x-auto", className)}>
      <T.List className="flex w-max min-w-full items-center gap-5 border-b border-line" {...rest} />
    </div>
  );
}

export function TabsTrigger({ count, children, className, ...rest }: ComponentProps<typeof T.Trigger> & { count?: ReactNode }) {
  return (
    <T.Trigger
      className={cn(
        "group relative -mb-px flex h-9 shrink-0 items-center gap-1.5 border-b-2 border-transparent text-sm font-medium whitespace-nowrap text-fg-3 outline-none transition-colors",
        "hover:text-fg data-[state=active]:border-fg data-[state=active]:text-fg focus-visible:text-fg",
        className,
      )}
      {...rest}
    >
      {children}
      {count !== undefined && count !== null ? (
        <span className="tabular rounded-full bg-inset px-1.5 text-[11px] leading-[18px] text-fg-3 group-data-[state=active]:bg-active group-data-[state=active]:text-fg-2">
          {count}
        </span>
      ) : null}
    </T.Trigger>
  );
}
