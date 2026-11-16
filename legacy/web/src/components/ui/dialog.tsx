import { AlertDialog as A, Dialog as D } from "radix-ui";
import { X } from "lucide-react";
import { useState, type ComponentProps, type ReactNode } from "react";
import { useT } from "../../i18n";
import { cn } from "../../lib/cn";
import { Button } from "./button";

export const Dialog = D.Root;
export const DialogTrigger = D.Trigger;
export const DialogClose = D.Close;

const OVERLAY =
  "fixed inset-0 z-[60] bg-scrim backdrop-blur-[2px] data-[state=open]:animate-fade-in data-[state=closed]:animate-fade-out";

export function DialogContent({
  className,
  children,
  size = "md",
  hideClose,
  ...rest
}: ComponentProps<typeof D.Content> & { size?: "sm" | "md" | "lg"; hideClose?: boolean }) {
  const t = useT();
  const width = size === "sm" ? "max-w-[400px]" : size === "lg" ? "max-w-[640px]" : "max-w-[480px]";
  return (
    <D.Portal>
      <D.Overlay className={OVERLAY} />
      <D.Content
        className={cn(
          "fixed top-[12vh] left-1/2 z-[61] w-[calc(100vw-24px)] -translate-x-1/2 rounded-2xl bg-surface shadow-dialog outline-none",
          "data-[state=open]:animate-dialog-in data-[state=closed]:animate-dialog-out",
          width,
          className,
        )}
        {...rest}
      >
        {children}
        {hideClose ? null : (
          <D.Close
            aria-label={t("common.close")}
            className="absolute top-3.5 right-3.5 flex size-7 items-center justify-center rounded-md text-fg-4 transition-colors hover:bg-hover hover:text-fg focus-visible:outline-2 focus-visible:outline-ring"
          >
            <X className="size-4" />
          </D.Close>
        )}
      </D.Content>
    </D.Portal>
  );
}

export function DialogHeader({ title, description, icon }: { title: ReactNode; description?: ReactNode; icon?: ReactNode }) {
  return (
    <div className="flex gap-3 px-5 pt-5 pb-1">
      {icon}
      <div className="min-w-0 flex-1 pr-8">
        <D.Title className="text-md font-semibold tracking-[-0.01em] text-fg">{title}</D.Title>
        {description ? <D.Description className="mt-1 text-sm text-fg-3">{description}</D.Description> : null}
      </div>
    </div>
  );
}

export function DialogBody({ className, ...rest }: ComponentProps<"div">) {
  return <div className={cn("px-5 py-4", className)} {...rest} />;
}

export function DialogFooter({ className, ...rest }: ComponentProps<"div">) {
  return (
    <div
      className={cn("flex items-center justify-end gap-2 rounded-b-2xl border-t border-line-soft bg-sheet/60 px-5 py-3", className)}
      {...rest}
    />
  );
}

/** A destructive confirmation. `onConfirm` may return a promise: the button spins until it settles. */
export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirm,
  onConfirm,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: ReactNode;
  description: ReactNode;
  confirm: string;
  onConfirm: () => unknown;
}) {
  const t = useT();
  const [busy, setBusy] = useState(false);
  return (
    <A.Root open={open} onOpenChange={onOpenChange}>
      <A.Portal>
        <A.Overlay className={OVERLAY} />
        <A.Content
          className={cn(
            "fixed top-[18vh] left-1/2 z-[61] w-[calc(100vw-24px)] max-w-[400px] -translate-x-1/2 rounded-2xl bg-surface shadow-dialog outline-none",
            "data-[state=open]:animate-dialog-in data-[state=closed]:animate-dialog-out",
          )}
        >
          <div className="px-5 pt-5 pb-4">
            <A.Title className="text-md font-semibold tracking-[-0.01em] text-fg">{title}</A.Title>
            <A.Description className="mt-1.5 text-sm text-fg-2">{description}</A.Description>
          </div>
          <div className="flex justify-end gap-2 rounded-b-2xl border-t border-line-soft bg-sheet/60 px-5 py-3">
            <A.Cancel asChild>
              <Button variant="ghost">{t("common.cancel")}</Button>
            </A.Cancel>
            <Button
              variant="danger"
              loading={busy}
              onClick={async () => {
                setBusy(true);
                try {
                  await onConfirm();
                  onOpenChange(false);
                } finally {
                  setBusy(false);
                }
              }}
            >
              {confirm}
            </Button>
          </div>
        </A.Content>
      </A.Portal>
    </A.Root>
  );
}
