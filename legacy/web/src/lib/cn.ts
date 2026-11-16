/** Joins class names, skipping anything that is not a non-empty string. */
export function cn(...parts: unknown[]): string {
  return parts.filter((p): p is string => typeof p === "string" && p.length > 0).join(" ");
}
