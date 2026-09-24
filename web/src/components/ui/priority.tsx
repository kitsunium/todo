import type { Priority } from "../../api/types";
import { cn } from "../../lib/cn";

export const PRIORITY_LABEL: Record<Priority, string> = {
  0: "No priority",
  1: "Urgent",
  2: "High",
  3: "Medium",
  4: "Low",
};

export const PRIORITY_SHORT: Record<Priority, string> = { 0: "None", 1: "Urgent", 2: "High", 3: "Medium", 4: "Low" };

/** The CSS color of a priority, for rings and icons. */
export const PRIORITY_COLOR: Record<Priority, string> = {
  0: "var(--fg-4)",
  1: "var(--urgent)",
  2: "var(--high)",
  3: "var(--medium)",
  4: "var(--low)",
};

/** Urgent is a filled square with "!"; high/medium/low are three bars, filled from the left. */
export function PriorityIcon({ priority, className }: { priority: Priority; className?: string }) {
  if (priority === 1) {
    return (
      <svg viewBox="0 0 16 16" className={cn("size-4", className)} aria-hidden="true">
        <rect x="1.5" y="1.5" width="13" height="13" rx="3.5" fill="var(--urgent)" />
        <path d="M8 4.6v4.2" stroke="#fff" strokeWidth="1.7" strokeLinecap="round" />
        <circle cx="8" cy="11.3" r="1" fill="#fff" />
      </svg>
    );
  }
  const filled = priority === 2 ? 3 : priority === 3 ? 2 : priority === 4 ? 1 : 0;
  const color = PRIORITY_COLOR[priority];
  const bars = [
    { x: 2.5, y: 9.5, h: 4 },
    { x: 6.75, y: 6.5, h: 7 },
    { x: 11, y: 3, h: 10.5 },
  ];
  return (
    <svg viewBox="0 0 16 16" className={cn("size-4", className)} aria-hidden="true">
      {bars.map((b, i) => (
        <rect
          key={b.x}
          x={b.x}
          y={b.y}
          width="2.75"
          height={b.h}
          rx="1"
          fill={i < filled ? color : "var(--fg-4)"}
          opacity={i < filled ? 1 : priority === 0 ? 0.55 : 0.35}
        />
      ))}
    </svg>
  );
}
