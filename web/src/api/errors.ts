// What to tell a person about a failed call, in their language. Pure but for
// the current language, which every function also takes explicitly: unit-tested.
//
// The API's codes and violation rules are mapped to our own sentences; the
// server's English text is only a last resort, for a code we don't know.
import { capitalize, getLocale, tr, type Key, type Locale } from "../i18n";
import { ApiError, BAD_RESPONSE, NETWORK_ERROR } from "./client";
import type { Violation } from "./types";

const BY_CODE: Record<string, Key> = {
  invalid_credentials: "error.invalid_credentials",
  invalid_token: "error.invalid_token",
  email_unverified: "error.email_unverified",
  account_locked: "error.account_locked",
  unauthenticated: "error.unauthenticated",
  permission_denied: "error.permission_denied",
  not_found: "error.not_found",
  rate_limited: "error.rate_limited",
  too_large: "error.too_large",
  unavailable: "error.unavailable",
  deadline_exceeded: "error.deadline_exceeded",
  internal: "error.internal",
  conflict: "error.conflict",
  invalid_argument: "error.invalid_argument",
  [NETWORK_ERROR]: "error.network_error",
  [BAD_RESPONSE]: "error.bad_response",
};

// A conflict says what it collided with only in the server's words: the ones
// we recognize get a precise sentence, any other the generic one.
const CONFLICTS: [RegExp, Key][] = [
  [/already contacts/i, "error.conflict.alreadyContacts"],
  [/request between you/i, "error.conflict.requestPending"],
  [/already invited this address/i, "error.conflict.addressInvited"],
  [/already a member/i, "error.conflict.alreadyMember"],
  [/already invited/i, "error.conflict.alreadyInvited"],
  [/not waiting for an answer/i, "error.conflict.answered"],
  [/owner cannot leave/i, "error.conflict.ownerLeaves"],
  [/task cannot/i, "error.conflict.task"],
];

/** The members of a request a violation may name; each is labelled by its
 *  "field.<name>" message, which TypeScript checks exists. */
const FIELD_NAMES = [
  "title",
  "notes",
  "name",
  "email",
  "password",
  "current",
  "due",
  "priority",
  "color",
  "groupId",
  "assigneeId",
  "userId",
  "locale",
  "token",
  "role",
] as const;

const FIELDS: Record<string, Key> = Object.fromEntries(
  FIELD_NAMES.map((f) => [f, `field.${f}` satisfies Key] as const),
);

/** The member of the request a violation concerns: "items[2].zip" → "items". */
function fieldOf(v: Violation): string {
  return v.path.split(/[.[]/)[0] ?? v.path;
}

/**
 * A violation as a fragment — "obligatoire", "200 caractères maximum" — from
 * its rule. A limit is only in the server's message ("must be at most 200
 * characters long"): its number is read from there.
 */
export function violationText(v: Violation, loc: Locale = getLocale()): string {
  const t = tr(loc);
  const n = /\d+/.exec(v.message)?.[0];
  switch (v.rule) {
    case "required":
      return t("violation.required");
    case "maxlen":
      return n
        ? t("violation.maxlen", { max: Number(n) })
        : t("violation.maxlenPlain");
    case "minlen":
      return n
        ? t("violation.minlen", { min: Number(n) })
        : t("violation.minlenPlain");
    case "min":
      return n
        ? t("violation.min", { min: Number(n) })
        : t("violation.invalid");
    case "max":
      return n
        ? t("violation.max", { max: Number(n) })
        : t("violation.invalid");
    case "email":
      return t("violation.email");
    case "oneof":
    case "one_of":
      return t("violation.oneof");
    case "self":
      return fieldOf(v) === "email"
        ? t("violation.selfEmail")
        : t("violation.self");
    case "match":
      return t("violation.match");
    case "member":
      return t("violation.member");
    case "audience":
    case "assignable":
      return t("violation.audience");
    case "datetime":
    case "time":
      return t("violation.datetime");
    case "timezone":
      return t("violation.timezone");
    case "line":
      return t("violation.line");
    default:
      return v.message || t("violation.invalid");
  }
}

const sentence = (s: string, loc: Locale) => {
  const c = capitalize(s, loc);
  return /[.!?…]$/.test(c) ? c : `${c}.`;
};

export type ErrorOptions = {
  /** What to say when the error is not the API's. */
  fallback?: string;
  locale?: Locale;
};

/** A sentence for a person, for any thrown value. */
export function errorMessage(err: unknown, opts: ErrorOptions = {}): string {
  const loc = opts.locale ?? getLocale();
  const t = tr(loc);
  if (err instanceof ApiError) {
    if (err.code === "invalid_argument" && err.violations.length === 1) {
      const v = err.violations[0]!;
      const field = FIELDS[fieldOf(v)];
      const text = violationText(v, loc);
      return sentence(
        field
          ? t("violation.withField", { field: t(field), message: text })
          : text,
        loc,
      );
    }
    if (err.code === "conflict") {
      const hit = CONFLICTS.find(([re]) => re.test(err.message));
      return t(hit ? hit[1] : "error.conflict");
    }
    const known = BY_CODE[err.code];
    if (known) return t(known);
    if (err.status >= 500) return t("error.internal");
    // A code we don't know: the server's own words beat saying nothing.
    if (err.message) return sentence(err.message, loc);
  }
  return opts.fallback ?? t("error.generic");
}

/** Field errors keyed by the request member they concern, each a sentence. */
export function fieldErrors(
  err: unknown,
  loc: Locale = getLocale(),
): Record<string, string> {
  const out: Record<string, string> = {};
  if (!(err instanceof ApiError)) return out;
  for (const v of err.violations) {
    const key = fieldOf(v);
    if (key && !out[key]) out[key] = sentence(violationText(v, loc), loc);
  }
  return out;
}

/** Whether the error is the API saying "sign in again". */
export function isUnauthenticated(err: unknown): boolean {
  return (
    err instanceof ApiError &&
    err.status === 401 &&
    err.code !== "invalid_credentials"
  );
}
