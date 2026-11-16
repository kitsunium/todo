import type { ErrorBody, Violation } from "./types";

/** An error answered by the API — or the network failing to reach it (status 0). */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly violations: Violation[];

  constructor(status: number, code: string, message: string, violations: Violation[] = []) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.violations = violations;
  }

  /** The message of the violation at path, if the request failed that rule. */
  violation(path: string): string | undefined {
    return this.violations.find((v) => v.path === path)?.message;
  }
}

/** The codes the client itself produces. */
export const NETWORK_ERROR = "network_error";
export const BAD_RESPONSE = "bad_response";

function isErrorBody(v: unknown): v is ErrorBody {
  if (typeof v !== "object" || v === null || !("error" in v)) return false;
  const e = (v as { error: unknown }).error;
  return typeof e === "object" && e !== null && typeof (e as { code?: unknown }).code === "string";
}

/** Builds the ApiError of a failed response from its status and parsed body. */
export function toApiError(status: number, body: unknown): ApiError {
  if (isErrorBody(body)) {
    const { code, message, violations } = body.error;
    return new ApiError(status, code, typeof message === "string" ? message : "", Array.isArray(violations) ? violations : []);
  }
  return new ApiError(status, status >= 500 ? "internal" : BAD_RESPONSE, "");
}

type Listener = () => void;
const unauthorizedListeners = new Set<Listener>();

/** Subscribes to "the session is gone": any 401 answered to a signed-in call. */
export function onUnauthorized(listener: Listener): () => void {
  unauthorizedListeners.add(listener);
  return () => unauthorizedListeners.delete(listener);
}

/** Calls whose 401 means "wrong credentials", not "your session ended". */
const CREDENTIAL_PATHS = ["/api/auth/login", "/api/auth/password"];

export type RequestOptions = { signal?: AbortSignal };

/**
 * Sends a JSON request to the API on the same origin, with the session
 * cookie. Resolves with the parsed body (undefined for 204) or rejects with
 * an ApiError.
 */
export async function request<T>(method: string, path: string, body?: unknown, opts: RequestOptions = {}): Promise<T> {
  const init: RequestInit = { method, credentials: "same-origin", headers: { Accept: "application/json" } };
  if (opts.signal) init.signal = opts.signal;
  if (body !== undefined) {
    init.headers = { Accept: "application/json", "Content-Type": "application/json" };
    init.body = JSON.stringify(body);
  }
  let res: Response;
  try {
    res = await fetch(path, init);
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") throw err;
    throw new ApiError(0, NETWORK_ERROR, "");
  }
  const text = res.status === 204 ? "" : await res.text().catch(() => "");
  let data: unknown = undefined;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      if (res.ok) throw new ApiError(res.status, BAD_RESPONSE, "");
    }
  }
  if (!res.ok) {
    const error = toApiError(res.status, data);
    if (res.status === 401 && !CREDENTIAL_PATHS.includes(path) && error.code !== "invalid_credentials") {
      for (const l of unauthorizedListeners) l();
    }
    throw error;
  }
  return data as T;
}

export const api = {
  get: <T>(path: string, opts?: RequestOptions) => request<T>("GET", path, undefined, opts),
  post: <T>(path: string, body?: unknown) => request<T>("POST", path, body),
  patch: <T>(path: string, body: unknown) => request<T>("PATCH", path, body),
  del: <T = void>(path: string) => request<T>("DELETE", path),
};

/** Encodes one path segment. */
export const seg = (s: string) => encodeURIComponent(s);
