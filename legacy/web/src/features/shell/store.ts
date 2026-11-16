// UI state shared across the shell: which overlay is open, the order of the
// visible task list (for next/previous in the panel), and who owns "C".
import { useSyncExternalStore } from "react";
import type { NewTask } from "../../api/types";

type UI = {
  palette: boolean;
  paletteMode: "all" | "tasks";
  newTask: boolean;
  newTaskDefaults: Partial<NewTask>;
  shortcuts: boolean;
  drawer: boolean;
  /** Task IDs of the list on screen, in visual order. */
  order: string[];
};

let state: UI = {
  palette: false,
  paletteMode: "all",
  newTask: false,
  newTaskDefaults: {},
  shortcuts: false,
  drawer: false,
  order: [],
};
const listeners = new Set<() => void>();

export function setUI(patch: Partial<UI>) {
  state = { ...state, ...patch };
  for (const l of listeners) l();
}

export function getUI(): UI {
  return state;
}

function subscribe(l: () => void) {
  listeners.add(l);
  return () => listeners.delete(l);
}

export function useUI<T>(select: (s: UI) => T): T {
  return useSyncExternalStore(subscribe, () => select(state), () => select(state));
}

// The inline quick add of the current page, if any: "C" focuses it rather
// than opening the dialog.
let quickAddFocus: (() => void) | null = null;
export function registerQuickAdd(focus: () => void) {
  quickAddFocus = focus;
  return () => {
    if (quickAddFocus === focus) quickAddFocus = null;
  };
}

export function newTask(defaults: Partial<NewTask> = {}) {
  if (quickAddFocus && Object.keys(defaults).length === 0) quickAddFocus();
  else setUI({ newTask: true, newTaskDefaults: defaults });
}
