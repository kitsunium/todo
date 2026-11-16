import type { ComponentProps } from "react";
import { cn } from "../../lib/cn";

type Tone = "neutral" | "danger" | "accent" | "success" | "warn" | "outline";

const TONES: Record<Tone, string> = {
  neutral: "bg-inset text-fg-2",
  danger: "bg-danger-soft text-danger-ink",
  accent: "bg-accent-soft text-accent-ink",
  success: "bg-success-soft text-success-ink",
  warn: "bg-warn-soft text-warn-ink",
  outline: "text-fg-2 shadow-[inset_0_0_0_1px_var(--line)]",
};

export function Chip({ tone = "neutral", className, ...rest }: ComponentProps<"span"> & { tone?: Tone }) {
  return (
    <span
      className={cn(
        "inline-flex h-5 max-w-full shrink-0 items-center gap-1 rounded-[5px] px-1.5 text-xs font-medium whitespace-nowrap",
        TONES[tone],
        className,
      )}
      {...rest}
    />
  );
}

/** A small count badge for navigation. */
export function Badge({ tone = "accent", className, ...rest }: ComponentProps<"span"> & { tone?: "accent" | "danger" | "neutral" }) {
  return (
    <span
      className={cn(
        "tabular inline-flex h-[18px] min-w-[18px] items-center justify-center rounded-full px-1.5 text-[11px] leading-none font-semibold",
        tone === "accent" && "bg-accent text-white",
        tone === "danger" && "bg-danger text-white",
        tone === "neutral" && "bg-active text-fg-2",
        className,
      )}
      {...rest}
    />
  );
}
