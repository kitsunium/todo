import { Slot } from "radix-ui";
import type { ComponentProps, ReactNode } from "react";
import { cn } from "../../lib/cn";
import { Spinner } from "./spinner";

export type ButtonVariant = "primary" | "accent" | "secondary" | "ghost" | "danger" | "quiet";
export type ButtonSize = "xs" | "sm" | "md" | "lg";

const VARIANTS: Record<ButtonVariant, string> = {
  primary:
    "bg-primary text-primary-fg hover:bg-primary-hover shadow-[inset_0_1px_0_rgb(255_255_255/0.1),0_1px_2px_rgb(0_0_0/0.14)]",
  accent:
    "bg-accent text-accent-fg hover:bg-accent-hover shadow-[inset_0_1px_0_rgb(255_255_255/0.2),0_1px_2px_rgb(194_65_12/0.3)]",
  secondary: "bg-surface text-fg shadow-card hover:bg-inset",
  ghost: "text-fg-2 hover:bg-hover hover:text-fg active:bg-active",
  quiet: "text-fg-3 hover:text-fg",
  danger:
    "bg-danger text-white hover:brightness-[0.94] shadow-[inset_0_1px_0_rgb(255_255_255/0.16),0_1px_2px_rgb(206_44_49/0.3)]",
};

const SIZES: Record<ButtonSize, string> = {
  xs: "h-6 gap-1 rounded-sm px-2 text-xs",
  sm: "h-7 gap-1.5 rounded-md px-2.5 text-sm",
  md: "h-8 gap-1.5 rounded-md px-3 text-sm",
  lg: "h-10 gap-2 rounded-lg px-4 text-base",
};

export type ButtonProps = ComponentProps<"button"> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  asChild?: boolean;
  icon?: ReactNode;
};

export function Button({
  variant = "secondary",
  size = "md",
  loading = false,
  asChild = false,
  icon,
  className,
  children,
  disabled,
  type,
  ...rest
}: ButtonProps) {
  const cls = cn(
    "relative inline-flex shrink-0 items-center justify-center font-medium whitespace-nowrap select-none",
    "transition-[background-color,color,box-shadow,filter] duration-150",
    "disabled:pointer-events-none disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
    VARIANTS[variant],
    SIZES[size],
    className,
  );
  if (asChild) {
    return (
      <Slot.Root className={cls} {...rest}>
        {children}
      </Slot.Root>
    );
  }
  return (
    <button type={type ?? "button"} className={cls} disabled={disabled || loading} aria-busy={loading || undefined} {...rest}>
      {loading ? (
        <span className="absolute inset-0 flex items-center justify-center">
          <Spinner className="size-4" />
        </span>
      ) : null}
      <span className={cn("inline-flex items-center gap-[inherit]", loading && "invisible")}>
        {icon}
        {children}
      </span>
    </button>
  );
}

export type IconButtonProps = ComponentProps<"button"> & {
  label: string;
  size?: "xs" | "sm" | "md";
  active?: boolean;
};

const ICON_SIZES = { xs: "size-6 rounded-sm", sm: "size-7 rounded-md", md: "size-8 rounded-md" };

/** A square ghost button. The label is its accessible name. */
export function IconButton({ label, size = "sm", active, className, children, type, ...rest }: IconButtonProps) {
  return (
    <button
      type={type ?? "button"}
      aria-label={label}
      className={cn(
        "inline-flex shrink-0 items-center justify-center text-fg-3 transition-colors duration-150",
        "hover:bg-hover hover:text-fg active:bg-active aria-expanded:bg-active aria-expanded:text-fg",
        "disabled:pointer-events-none disabled:opacity-40 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring",
        active && "bg-active text-fg",
        ICON_SIZES[size],
        className,
      )}
      {...rest}
    >
      {children}
    </button>
  );
}
