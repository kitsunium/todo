import { useEffect, useRef } from "react";

/** Whether a key event comes from somewhere the person is typing. */
export function isTyping(e: KeyboardEvent): boolean {
  const el = e.target as HTMLElement | null;
  if (!el) return false;
  const tag = el.tagName;
  return tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT" || el.isContentEditable;
}

/** Whether a modal layer (dialog, menu, popover) is open above the page. */
export function overlayOpen(): boolean {
  return !!document.querySelector(
    '[role="dialog"][data-state="open"], [role="alertdialog"][data-state="open"], [role="menu"][data-state="open"], [role="listbox"]',
  );
}

type Binding = { key: string; mod?: boolean; shift?: boolean; allowInInputs?: boolean; allowInOverlay?: boolean; run: (e: KeyboardEvent) => void };

/**
 * Global keyboard shortcuts. Single keys are ignored while typing and while
 * an overlay is open; "g then x" sequences are supported with key "g x".
 */
export function useHotkeys(bindings: Binding[], enabled = true) {
  const ref = useRef(bindings);
  ref.current = bindings;
  useEffect(() => {
    if (!enabled) return;
    let pending = "";
    let timer: ReturnType<typeof setTimeout> | undefined;
    const onKey = (e: KeyboardEvent) => {
      if (e.defaultPrevented || e.isComposing) return;
      const mod = e.metaKey || e.ctrlKey;
      const key = e.key.length === 1 ? e.key.toLowerCase() : e.key;
      const typing = isTyping(e);
      const overlay = overlayOpen();
      const seq = pending ? `${pending} ${key}` : key;
      for (const b of ref.current) {
        if (!!b.mod !== mod) continue;
        if (b.shift !== undefined && b.shift !== e.shiftKey) continue;
        if (typing && !b.allowInInputs) continue;
        if (overlay && !b.allowInOverlay) continue;
        if (e.altKey) continue;
        if (b.key === seq || (!pending && b.key === key)) {
          e.preventDefault();
          pending = "";
          b.run(e);
          return;
        }
      }
      if (!mod && !typing && !overlay && key === "g") {
        pending = "g";
        clearTimeout(timer);
        timer = setTimeout(() => (pending = ""), 900);
        return;
      }
      pending = "";
    };
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("keydown", onKey);
      clearTimeout(timer);
    };
  }, [enabled]);
}
