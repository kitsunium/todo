import { Tabs as T } from "radix-ui";
import type { ComponentProps, ReactNode } from "react";
import { cn } from "../../lib/cn";

export const Tabs = T.Root;
export const TabsContent = T.Content;

export function TabsList({ className, ...rest }: ComponentProps<typeof T.List>) {
  return <T.List className={cn("flex items-center gap-5 border-b border-line", className)} {...rest} />;
}

export function TabsTrigger({ count, children, className, ...rest }: ComponentProps<typeof T.Trigger> & { count?: ReactNode }) {
  return (
    <T.Trigger
      className={cn(
        "group relative -mb-px flex h-9 items-center gap-1.5 border-b-2 border-transparent text-sm font-medium text-fg-3 outline-none transition-colors",
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
