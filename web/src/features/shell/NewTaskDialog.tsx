import { Dialog as D } from "radix-ui";
import { useLocation } from "react-router";
import { QuickAdd } from "../tasks/QuickAdd";
import { setUI, useUI } from "./store";

/** Quick add, anywhere: C, the sidebar button, or the palette. */
export function NewTaskDialog() {
  const open = useUI((s) => s.newTask);
  const defaults = useUI((s) => s.newTaskDefaults);
  const { pathname } = useLocation();
  const groupId = /^\/app\/groups\/([^/]+)/.exec(pathname)?.[1];
  const { title, ...rest } = defaults;
  return (
    <D.Root open={open} onOpenChange={(o) => setUI({ newTask: o })}>
      <D.Portal>
        <D.Overlay className="fixed inset-0 z-[60] bg-scrim backdrop-blur-[2px] data-[state=open]:animate-fade-in data-[state=closed]:animate-fade-out" />
        <D.Content
          aria-describedby={undefined}
          className="fixed top-[16vh] left-1/2 z-[61] w-[calc(100vw-24px)] max-w-[600px] -translate-x-1/2 rounded-2xl bg-surface p-1.5 shadow-dialog outline-none data-[state=open]:animate-dialog-in data-[state=closed]:animate-dialog-out"
        >
          <D.Title className="sr-only">New task</D.Title>
          {open ? (
            <QuickAdd
              variant="dialog"
              autoFocus
              {...(title ? { initialText: title } : {})}
              defaults={{ ...(groupId && !rest.groupId ? { groupId } : {}), ...rest }}
              onDone={() => setUI({ newTask: false })}
            />
          ) : null}
        </D.Content>
      </D.Portal>
    </D.Root>
  );
}
