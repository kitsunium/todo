import { CircleAlert, CircleCheck, Info } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

const TONES = {
  danger: { box: "bg-danger-soft text-danger-ink", Icon: CircleAlert },
  success: { box: "bg-success-soft text-success-ink", Icon: CircleCheck },
  info: { box: "bg-inset text-fg-2", Icon: Info },
  warn: { box: "bg-warn-soft text-warn-ink", Icon: CircleAlert },
} as const;

export function Callout({
  tone = "info",
  title,
  children,
  action,
  className,
}: {
  tone?: keyof typeof TONES;
  title?: ReactNode;
  children?: ReactNode;
  action?: ReactNode;
  className?: string;
}) {
  const { box, Icon } = TONES[tone];
  return (
    <div role={tone === "danger" ? "alert" : "status"} className={cn("flex gap-2.5 rounded-lg px-3 py-2.5 text-sm animate-fade-in", box, className)}>
      <Icon className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
      <div className="min-w-0 flex-1">
        {title ? <p className="font-medium">{title}</p> : null}
        {children ? <div className={cn(title && "mt-0.5", "opacity-90")}>{children}</div> : null}
        {action ? <div className="mt-2">{action}</div> : null}
      </div>
    </div>
  );
}
