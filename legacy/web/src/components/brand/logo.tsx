import { cn } from "../../lib/cn";

/** The fox: two ears, a pointed chin, and a check where its face is. */
export function FoxMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 32 32" className={cn("size-6", className)} aria-hidden="true">
      <path
        d="M5.5 4.5 12.5 10h7l7-5.5 1 11L16 27.5 4.5 15.5Z"
        fill="var(--accent)"
        stroke="var(--accent)"
        strokeWidth="2.4"
        strokeLinejoin="round"
      />
      <path
        d="m10.8 16.2 3.8 3.6 6.8-7"
        fill="none"
        stroke="#fff"
        strokeWidth="2.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function Logo({ className, markClassName }: { className?: string; markClassName?: string }) {
  return (
    <span className={cn("inline-flex items-center gap-2 select-none", className)}>
      <FoxMark className={cn("size-[22px]", markClassName)} />
      <span className="text-[17px] leading-none font-semibold tracking-[-0.03em] text-fg">todo</span>
    </span>
  );
}
