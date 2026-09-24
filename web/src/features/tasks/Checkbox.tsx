import type { CSSProperties } from "react";
import type { Priority } from "../../api/types";
import { PRIORITY_COLOR } from "../../components/ui/priority";
import { cn } from "../../lib/cn";

/**
 * The round checkbox of a task, ringed with its priority color. Checking it
 * fills the circle and draws the check.
 */
export function TaskCheckbox({
  priority,
  checked,
  title,
  onToggle,
  disabled,
  size = "md",
}: {
  priority: Priority;
  checked: boolean;
  title: string;
  onToggle: () => void;
  disabled?: boolean;
  size?: "md" | "lg";
}) {
  const style = { "--pc": PRIORITY_COLOR[priority] } as CSSProperties;
  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={checked}
      aria-label={checked ? `Mark “${title}” as not done` : `Complete “${title}”`}
      disabled={disabled}
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        onToggle();
      }}
      style={style}
      className={cn(
        "group/check relative -m-[3px] flex shrink-0 items-center justify-center rounded-full p-[3px] outline-none",
        "focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50",
        "pointer-coarse:before:absolute pointer-coarse:before:-inset-2.5 pointer-coarse:before:content-['']",
        size === "lg" ? "size-[26px]" : "size-6",
      )}
    >
      <span
        className={cn(
          "absolute inset-[3px] rounded-full border-[1.6px] transition-[background-color,border-color,transform] duration-200",
          checked
            ? "scale-100 border-done bg-done"
            : priority === 0
              ? "border-fg-4 group-hover/check:border-fg-3"
              : "border-[var(--pc)] bg-[color-mix(in_srgb,var(--pc)_10%,transparent)]",
          "group-active/check:scale-90",
        )}
      />
      <svg viewBox="0 0 12 12" className="relative size-[11px]" aria-hidden="true">
        <path
          d="M2.6 6.3 4.9 8.5 9.4 3.8"
          fill="none"
          strokeWidth="1.9"
          strokeLinecap="round"
          strokeLinejoin="round"
          pathLength={1}
          className={cn(
            "transition-[stroke-dashoffset,opacity] duration-300 ease-out [stroke-dasharray:1]",
            checked
              ? "stroke-white opacity-100 [stroke-dashoffset:0] delay-75"
              : "stroke-[var(--pc)] opacity-0 [stroke-dashoffset:1] group-hover/check:opacity-70 group-hover/check:[stroke-dashoffset:0]",
            !checked && priority === 0 && "stroke-fg-3",
          )}
        />
      </svg>
    </button>
  );
}
