// Light, dark, or the system's: stored per browser, applied as a class on
// <html>. CSS already follows the system before this runs, so there is no
// flash of the wrong theme.
import { useSyncExternalStore } from "react";

export type ThemeChoice = "light" | "dark" | "system";
const KEY = "todo.theme";
const listeners = new Set<() => void>();
const media = typeof window !== "undefined" ? window.matchMedia("(prefers-color-scheme: dark)") : null;

function read(): ThemeChoice {
  try {
    const v = localStorage.getItem(KEY);
    return v === "light" || v === "dark" ? v : "system";
  } catch {
    return "system";
  }
}

let choice: ThemeChoice = typeof window !== "undefined" ? read() : "system";

export function resolvedTheme(c: ThemeChoice = choice): "light" | "dark" {
  if (c === "system") return media?.matches ? "dark" : "light";
  return c;
}

export function applyTheme(): void {
  const root = document.documentElement;
  const dark = resolvedTheme() === "dark";
  root.classList.toggle("dark", dark);
  root.classList.toggle("light", !dark);
  for (const meta of document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')) {
    meta.content = dark ? "#0e0e10" : "#f5f5f4";
  }
}

export function setTheme(c: ThemeChoice): void {
  choice = c;
  try {
    if (c === "system") localStorage.removeItem(KEY);
    else localStorage.setItem(KEY, c);
  } catch {
    /* private mode: the choice lasts for this page */
  }
  applyTheme();
  for (const l of listeners) l();
}

export function toggleTheme(): void {
  setTheme(resolvedTheme() === "dark" ? "light" : "dark");
}

media?.addEventListener("change", () => {
  if (choice === "system") {
    applyTheme();
    for (const l of listeners) l();
  }
});

function subscribe(l: () => void) {
  listeners.add(l);
  return () => listeners.delete(l);
}

export function useTheme(): { choice: ThemeChoice; resolved: "light" | "dark" } {
  const c = useSyncExternalStore(subscribe, () => choice, () => "system" as ThemeChoice);
  const resolved = useSyncExternalStore(subscribe, () => resolvedTheme(), () => "light" as const);
  return { choice: c, resolved };
}
