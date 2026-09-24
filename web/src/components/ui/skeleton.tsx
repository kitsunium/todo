import { cn } from "../../lib/cn";

export function Skeleton({ className }: { className?: string }) {
  return <span className={cn("skeleton block", className)} aria-hidden="true" />;
}

/** The shape of a task list while it loads. */
export function TaskListSkeleton({ rows = 6 }: { rows?: number }) {
  const widths = ["w-[62%]", "w-[44%]", "w-[71%]", "w-[38%]", "w-[55%]", "w-[48%]", "w-[66%]", "w-[41%]"];
  return (
    <div className="mt-2" aria-busy="true" aria-label="Loading tasks">
      <Skeleton className="mb-3 ml-1 h-3 w-16" />
      {Array.from({ length: rows }, (_, i) => (
        <div key={i} className="flex items-start gap-3 px-2 py-2.5">
          <Skeleton className="mt-0.5 size-[18px] rounded-full" />
          <div className="flex-1 space-y-2 pt-0.5">
            <Skeleton className={cn("h-3.5", widths[i % widths.length])} />
            {i % 3 !== 2 ? <Skeleton className="h-2.5 w-24" /> : null}
          </div>
        </div>
      ))}
    </div>
  );
}
