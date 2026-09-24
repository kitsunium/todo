import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

export const isMac = typeof navigator !== "undefined" && /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
export const MOD = isMac ? "⌘" : "Ctrl";

export function Kbd({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <kbd
      className={cn(
        "inline-flex h-[18px] min-w-[18px] items-center justify-center rounded-[5px] px-1 font-sans text-[11px] leading-none font-medium",
        "bg-surface text-fg-3 shadow-[0_0_0_1px_var(--line),0_1px_0_1px_var(--line)]",
        className,
      )}
    >
      {children}
    </kbd>
  );
}

/** "⌘ K", "Shift X": each key in its own cap. */
export function Keys({ keys, className }: { keys: string[]; className?: string }) {
  return (
    <span className={cn("inline-flex items-center gap-1", className)}>
      {keys.map((k) => (
        <Kbd key={k}>{k}</Kbd>
      ))}
    </span>
  );
}
