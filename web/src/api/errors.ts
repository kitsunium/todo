// What to tell a person about a failed call. Pure: unit-tested.
import { ApiError, BAD_RESPONSE, NETWORK_ERROR } from "./client";

const BY_CODE: Record<string, string> = {
  invalid_credentials: "That email and password don’t match.",
  invalid_token: "This link is invalid or has expired. Ask for a new one.",
  email_unverified: "Confirm your email address first — check your inbox for the link.",
  account_locked: "Too many attempts. Your account is locked for 15 minutes.",
  unauthenticated: "Your session has ended. Sign in again.",
  permission_denied: "You don’t have permission to do that.",
  not_found: "It isn’t here anymore — it may have been deleted.",
  rate_limited: "Slow down a little — try again in a few seconds.",
  too_large: "That’s too much text for one request.",
  unavailable: "The service is busy. Try again in a moment.",
  deadline_exceeded: "That took too long. Try again.",
  internal: "Something went wrong on our side. Try again.",
  [NETWORK_ERROR]: "Can’t reach the server. Check your connection.",
  [BAD_RESPONSE]: "The server sent an unexpected answer.",
};

/** A sentence for a person, for any thrown value. */
export function errorMessage(err: unknown, fallback = "Something went wrong. Try again."): string {
  if (err instanceof ApiError) {
    // Conflicts and invalid requests carry the backend's own wording
    // ("you are already contacts"), which is more precise than ours.
    if ((err.code === "conflict" || err.code === "invalid_argument") && err.message) {
      if (err.violations.length === 1) return capitalize(err.violations[0]!.message);
      return capitalize(err.message);
    }
    const known = BY_CODE[err.code];
    if (known) return known;
    if (err.message) return capitalize(err.message);
    if (err.status >= 500) return BY_CODE.internal!;
  }
  return fallback;
}

/** Field errors keyed by the request member they concern. */
export function fieldErrors(err: unknown): Record<string, string> {
  const out: Record<string, string> = {};
  if (!(err instanceof ApiError)) return out;
  for (const v of err.violations) {
    const key = v.path.split(/[.[]/)[0] ?? v.path;
    if (key && !out[key]) out[key] = capitalize(v.message);
  }
  return out;
}

/** Whether the error is the API saying "sign in again". */
export function isUnauthenticated(err: unknown): boolean {
  return err instanceof ApiError && err.status === 401 && err.code !== "invalid_credentials";
}

function capitalize(s: string): string {
  return s ? s.charAt(0).toUpperCase() + s.slice(1) : s;
}
