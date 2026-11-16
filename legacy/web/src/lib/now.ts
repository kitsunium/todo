// One clock for the whole UI, ticking every 30 seconds: relative dates
// ("5m ago", "Today") stay right without each row running its own timer.
import { useSyncExternalStore } from "react";

let now = new Date();
const listeners = new Set<() => void>();
let timer: ReturnType<typeof setInterval> | undefined;

function subscribe(l: () => void) {
  listeners.add(l);
  if (!timer) {
    timer = setInterval(() => {
      now = new Date();
      for (const x of listeners) x();
    }, 30_000);
  }
  return () => {
    listeners.delete(l);
    if (listeners.size === 0 && timer) {
      clearInterval(timer);
      timer = undefined;
    }
  };
}

export function useNow(): Date {
  return useSyncExternalStore(subscribe, () => now, () => now);
}
