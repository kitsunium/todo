// Mock mode: answers /api/* (and the Studio probe /_kit/api/graph, with a
// 404) from the in-memory backend, after a short realistic delay.
// Imported only when import.meta.env.MODE === "mock".
import { clear, DEMO_EMAIL, seed, type DB } from "./db";
import { handle, mockDB, resetDB } from "./server";

declare global {
  interface Window {
    __mock?: {
      reset: () => void;
      signOut: () => void;
      signIn: (email?: string) => void;
      db: () => DB;
      setLatency: (ms: number) => void;
    };
  }
}

const LATENCY_KEY = "todo.mock.latency";
function latency(): number {
  const v = Number(localStorage.getItem(LATENCY_KEY));
  return Number.isFinite(v) && localStorage.getItem(LATENCY_KEY) !== null ? v : 220;
}

export function installMock(): void {
  const realFetch = window.fetch.bind(window);
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const href = typeof input === "string" ? input : input instanceof URL ? input.href : input.url;
    const url = new URL(href, location.origin);
    const ours = url.origin === location.origin && (url.pathname.startsWith("/api/") || url.pathname.startsWith("/_kit/api/"));
    if (!ours) return realFetch(input, init);
    const method = (init?.method ?? (input instanceof Request ? input.method : "GET")).toUpperCase();
    const body = typeof init?.body === "string" ? init.body : undefined;
    const signal = init?.signal;
    const wait = latency() * (0.7 + Math.random() * 0.6);
    await new Promise<void>((resolve, reject) => {
      const t = setTimeout(resolve, wait);
      signal?.addEventListener("abort", () => {
        clearTimeout(t);
        reject(new DOMException("The operation was aborted.", "AbortError"));
      });
    });
    const res = url.pathname.startsWith("/_kit/")
      ? { status: 404, body: { error: { code: "not_found", message: "no Studio in mock mode" } } }
      : handle(method, url, body);
    return new Response(res.status === 204 ? null : JSON.stringify(res.body), {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  };

  window.__mock = {
    reset: () => {
      clear();
      resetDB(seed());
    },
    signOut: () => {
      const db = mockDB();
      delete db.cookie;
      resetDB(db);
    },
    signIn: (email = DEMO_EMAIL) => {
      const db = mockDB();
      const u = db.users.find((x) => x.email === email);
      const s = db.sessions.find((x) => x.userId === u?.id);
      if (s) db.cookie = s.token;
      resetDB(db);
    },
    db: () => mockDB(),
    setLatency: (ms: number) => localStorage.setItem(LATENCY_KEY, String(ms)),
  };
  console.info(`[mock] todo API in memory — sign in as ${DEMO_EMAIL}`);
}
