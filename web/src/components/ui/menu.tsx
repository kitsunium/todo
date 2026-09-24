import { DropdownMenu as M } from "radix-ui";
import { Check, ChevronRight } from "lucide-react";
import type { ComponentProps, ReactNode } from "react";
import { cn } from "../../lib/cn";

export const Menu = M.Root;
export const MenuTrigger = M.Trigger;
export const MenuGroup = M.Group;
export const MenuSub = M.Sub;
export const MenuRadioGroup = M.RadioGroup;

export const POP_SURFACE =
  "z-[70] overflow-hidden rounded-lg bg-raised p-1 text-sm text-fg shadow-pop outline-none " +
  "data-[state=open]:animate-pop-in data-[state=closed]:animate-pop-out origin-[var(--radix-dropdown-menu-content-transform-origin)]";

export function MenuContent({ className, sideOffset = 6, align = "start", ...rest }: ComponentProps<typeof M.Content>) {
  return (
    <M.Portal>
      <M.Content
        sideOffset={sideOffset}
        align={align}
        collisionPadding={8}
        className={cn(POP_SURFACE, "min-w-[200px]", className)}
        {...rest}
      />
    </M.Portal>
  );
}

const ITEM =
  "relative flex h-8 items-center gap-2 rounded-md px-2 outline-none select-none " +
  "data-[highlighted]:bg-hover data-[disabled]:opacity-40 data-[disabled]:pointer-events-none";

export function MenuItem({
  icon,
  shortcut,
  danger,
  className,
  children,
  ...rest
}: ComponentProps<typeof M.Item> & { icon?: ReactNode; shortcut?: ReactNode; danger?: boolean }) {
  return (
    <M.Item className={cn(ITEM, danger && "text-danger-ink data-[highlighted]:bg-danger-soft", className)} {...rest}>
      {icon ? <span className={cn("flex size-4 items-center justify-center", danger ? "" : "text-fg-3")}>{icon}</span> : null}
      <span className="min-w-0 flex-1 truncate">{children}</span>
      {shortcut ? <span className="ml-4 text-xs text-fg-4">{shortcut}</span> : null}
    </M.Item>
  );
}

export function MenuCheckItem({
  icon,
  checked,
  className,
  children,
  ...rest
}: ComponentProps<typeof M.Item> & { icon?: ReactNode; checked?: boolean }) {
  return (
    <M.Item className={cn(ITEM, className)} {...rest}>
      {icon ? <span className="flex size-4 items-center justify-center text-fg-3">{icon}</span> : null}
      <span className="min-w-0 flex-1 truncate">{children}</span>
      <Check className={cn("size-4 text-accent", checked ? "opacity-100" : "opacity-0")} aria-hidden="true" />
    </M.Item>
  );
}

export function MenuRadioItem({
  icon,
  className,
  children,
  ...rest
}: ComponentProps<typeof M.RadioItem> & { icon?: ReactNode }) {
  return (
    <M.RadioItem className={cn(ITEM, className)} {...rest}>
      {icon ? <span className="flex size-4 items-center justify-center text-fg-3">{icon}</span> : null}
      <span className="min-w-0 flex-1 truncate">{children}</span>
      <M.ItemIndicator>
        <Check className="size-4 text-accent" aria-hidden="true" />
      </M.ItemIndicator>
    </M.RadioItem>
  );
}

export function MenuSubTrigger({ icon, children, className, ...rest }: ComponentProps<typeof M.SubTrigger> & { icon?: ReactNode }) {
  return (
    <M.SubTrigger className={cn(ITEM, "data-[state=open]:bg-hover", className)} {...rest}>
      {icon ? <span className="flex size-4 items-center justify-center text-fg-3">{icon}</span> : null}
      <span className="flex-1">{children}</span>
      <ChevronRight className="size-3.5 text-fg-4" aria-hidden="true" />
    </M.SubTrigger>
  );
}

export function MenuSubContent({ className, ...rest }: ComponentProps<typeof M.SubContent>) {
  return (
    <M.Portal>
      <M.SubContent sideOffset={6} alignOffset={-4} collisionPadding={8} className={cn(POP_SURFACE, "min-w-[180px]", className)} {...rest} />
    </M.Portal>
  );
}

export function MenuSeparator({ className }: { className?: string }) {
  return <M.Separator className={cn("-mx-1 my-1 h-px bg-line-soft", className)} />;
}

export function MenuLabel({ className, ...rest }: ComponentProps<typeof M.Label>) {
  return <M.Label className={cn("px-2 pt-1.5 pb-1 text-2xs font-medium tracking-wide text-fg-4 uppercase", className)} {...rest} />;
}
