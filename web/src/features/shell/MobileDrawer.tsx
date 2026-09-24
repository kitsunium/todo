import { Dialog as D } from "radix-ui";
import { useEffect } from "react";
import { useLocation } from "react-router";
import { Sidebar } from "./Sidebar";
import { setUI, useUI } from "./store";

/** Below 900px the sidebar lives in a drawer. */
export function MobileDrawer() {
  const open = useUI((s) => s.drawer);
  const { pathname } = useLocation();
  useEffect(() => setUI({ drawer: false }), [pathname]);
  useEffect(() => {
    const mq = window.matchMedia("(min-width: 900px)");
    const onChange = () => mq.matches && setUI({ drawer: false });
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);
  return (
    <D.Root open={open} onOpenChange={(o) => setUI({ drawer: o })}>
      <D.Portal>
        <D.Overlay className="fixed inset-0 z-[55] bg-scrim data-[state=open]:animate-fade-in data-[state=closed]:animate-fade-out" />
        <D.Content
          aria-describedby={undefined}
          tabIndex={-1}
          onOpenAutoFocus={(e) => {
            e.preventDefault();
            (e.currentTarget as HTMLElement | null)?.focus({ preventScroll: true });
          }}
          className="fixed inset-y-0 left-0 outline-none z-[56] flex w-[min(300px,86vw)] flex-col bg-sidebar shadow-dialog outline-none data-[state=open]:animate-drawer-in data-[state=closed]:animate-drawer-out"
        >
          <D.Title className="sr-only">Navigation</D.Title>
          <Sidebar onNavigate={() => setUI({ drawer: false })} />
        </D.Content>
      </D.Portal>
    </D.Root>
  );
}
