import { CircleAlert, Eye, EyeOff } from "lucide-react";
import { useId, useLayoutEffect, useRef, useState, type ComponentProps, type ReactNode } from "react";
import { cn } from "../../lib/cn";

const FIELD =
  "w-full min-w-0 rounded-md bg-surface text-fg shadow-card outline-none transition-shadow duration-150 " +
  "placeholder:text-fg-4 disabled:opacity-60 " +
  "focus:shadow-[0_0_0_1px_var(--accent),0_0_0_4px_var(--accent-soft)] " +
  "aria-[invalid=true]:shadow-[0_0_0_1px_var(--danger)] aria-[invalid=true]:focus:shadow-[0_0_0_1px_var(--danger),0_0_0_4px_var(--danger-soft)]";

export type InputProps = ComponentProps<"input"> & { inputSize?: "md" | "lg" };

export function Input({ className, inputSize = "md", ...rest }: InputProps) {
  return (
    <input
      className={cn(FIELD, inputSize === "lg" ? "h-10 px-3 text-base" : "h-8 px-2.5 text-sm", className)}
      {...rest}
    />
  );
}

/** A textarea that grows with its content. */
export function AutoTextarea({
  className,
  value,
  minRows = 1,
  ...rest
}: ComponentProps<"textarea"> & { minRows?: number }) {
  const ref = useRef<HTMLTextAreaElement | null>(null);
  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.style.height = "0px";
    el.style.height = `${el.scrollHeight}px`;
  }, [value]);
  return <textarea ref={ref} rows={minRows} value={value} className={cn("resize-none overflow-hidden", className)} {...rest} />;
}

export function Label({ className, ...rest }: ComponentProps<"label">) {
  return <label className={cn("text-sm font-medium text-fg", className)} {...rest} />;
}

export function FieldError({ id, children }: { id?: string; children: ReactNode }) {
  return (
    <p id={id} className="flex items-start gap-1.5 text-xs text-danger-ink animate-fade-in" role="alert">
      <CircleAlert className="mt-px size-3.5 shrink-0" aria-hidden="true" />
      <span>{children}</span>
    </p>
  );
}

type FieldRender = { id: string; "aria-invalid": boolean; "aria-describedby"?: string };

/** A label, a control, and its hint or error — wired for screen readers. */
export function Field({
  label,
  hint,
  error,
  aside,
  children,
  className,
}: {
  label: ReactNode;
  hint?: ReactNode;
  error?: string | undefined;
  aside?: ReactNode;
  children: (p: FieldRender) => ReactNode;
  className?: string;
}) {
  const id = useId();
  const described = error ? `${id}-error` : hint ? `${id}-hint` : undefined;
  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <div className="flex items-baseline justify-between gap-2">
        <Label htmlFor={id}>{label}</Label>
        {aside}
      </div>
      {children({ id, "aria-invalid": !!error, ...(described ? { "aria-describedby": described } : {}) })}
      {error ? (
        <FieldError id={`${id}-error`}>{error}</FieldError>
      ) : hint ? (
        <div id={`${id}-hint`} className="text-xs text-fg-3">
          {hint}
        </div>
      ) : null}
    </div>
  );
}

export function PasswordInput(props: InputProps) {
  const [shown, setShown] = useState(false);
  return (
    <div className="relative">
      <Input {...props} type={shown ? "text" : "password"} className={cn("pr-10", props.className)} />
      <button
        type="button"
        onClick={() => setShown((s) => !s)}
        aria-label={shown ? "Hide password" : "Show password"}
        aria-pressed={shown}
        className="absolute top-1/2 right-1.5 flex size-7 -translate-y-1/2 items-center justify-center rounded-md text-fg-4 transition-colors hover:bg-hover hover:text-fg-2 focus-visible:outline-2 focus-visible:outline-ring"
      >
        {shown ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
      </button>
    </div>
  );
}
