import { Tooltip as T } from "radix-ui";
import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

export const TooltipProvider = T.Provider;

// No hover, no tooltip: on touch screens a tooltip only ever appears by accident.
const touchOnly = typeof window !== "undefined" && window.matchMedia("(hover: none)").matches;

export function Tooltip({
  content,
  keys,
  side = "top",
  align = "center",
  children,
  disabled,
}: {
  content: ReactNode;
  keys?: string[];
  side?: "top" | "right" | "bottom" | "left";
  align?: "start" | "center" | "end";
  children: ReactNode;
  disabled?: boolean;
}) {
  if (disabled || touchOnly) return <>{children}</>;
  return (
    <T.Root>
      <T.Trigger asChild>{children}</T.Trigger>
      <T.Portal>
        <T.Content
          side={side}
          align={align}
          sideOffset={6}
          collisionPadding={8}
          className={cn(
            "z-[80] flex items-center gap-2 rounded-md px-2 py-1 text-xs font-medium select-none",
            "bg-[#1c1917] text-[#fafaf9] shadow-pop",
            "data-[state=delayed-open]:animate-pop-in data-[state=instant-open]:animate-pop-in data-[state=closed]:animate-pop-out",
          )}
        >
          {content}
          {keys ? (
            <span className="flex items-center gap-0.5">
              {keys.map((k) => (
                <kbd
                  key={k}
                  className="inline-flex h-4 min-w-4 items-center justify-center rounded-[4px] bg-white/12 px-1 font-sans text-[10px] leading-none text-white/80"
                >
                  {k}
                </kbd>
              ))}
            </span>
          ) : null}
        </T.Content>
      </T.Portal>
    </T.Root>
  );
}
