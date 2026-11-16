/** A same-site path from ?next=, or the fallback: never an open redirect. */
export function safeNext(next: string | null | undefined, fallback = "/app/today"): string {
  if (!next || !next.startsWith("/") || next.startsWith("//") || next.startsWith("/\\")) return fallback;
  if (next.startsWith("/login") || next.startsWith("/signup")) return fallback;
  return next;
}
