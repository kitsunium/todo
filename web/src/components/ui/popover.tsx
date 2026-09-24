import { Popover as P } from "radix-ui";
import type { ComponentProps } from "react";
import { cn } from "../../lib/cn";

export const Popover = P.Root;
export const PopoverTrigger = P.Trigger;
export const PopoverAnchor = P.Anchor;
export const PopoverClose = P.Close;

export function PopoverContent({ className, sideOffset = 6, align = "start", ...rest }: ComponentProps<typeof P.Content>) {
  return (
    <P.Portal>
      <P.Content
        sideOffset={sideOffset}
        align={align}
        collisionPadding={8}
        className={cn(
          "z-[70] rounded-xl bg-raised text-sm text-fg shadow-pop outline-none",
          "data-[state=open]:animate-pop-in data-[state=closed]:animate-pop-out origin-[var(--radix-popover-content-transform-origin)]",
          className,
        )}
        {...rest}
      />
    </P.Portal>
  );
}
