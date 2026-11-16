// The interface's language: French first, English second. No library — a
// dictionary per language (fr.ts is the source of truth), {placeholders},
// plurals through Intl.PluralRules, and dates through date-fns locales.
//
// Which language: signed in, the account's (user.locale, applied by
// useLocaleSync); signed out, the last choice made in this browser; else
// French. <html lang> follows.
import type { Locale as DateLocale } from "date-fns";
import { enUS } from "date-fns/locale/en-US";
import { fr as dateFr } from "date-fns/locale/fr";
import {
  createElement,
  Fragment,
  useSyncExternalStore,
  type ReactNode,
} from "react";
import { en } from "./en";
import { fr } from "./fr";
import type { Args, Key, Locale, Messages, Placeholder, Plural } from "./types";

export type { Key, Locale } from "./types";

export const LOCALES: readonly Locale[] = ["fr", "en"];
export const DEFAULT_LOCALE: Locale = "fr";

const DICTS: Record<Locale, Messages> = { fr, en };
const DATE_LOCALES: Record<Locale, DateLocale> = { fr: dateFr, en: enUS };

export function isLocale(v: unknown): v is Locale {
  return v === "fr" || v === "en";
}

// ── The current language ──────────────────────────────────────────────

const KEY = "todo.locale";
const browser = typeof window !== "undefined";

function stored(): Locale | null {
  if (!browser) return null;
  try {
    const v = window.localStorage.getItem(KEY);
    return isLocale(v) ? v : null;
  } catch {
    return null;
  }
}

let current: Locale = stored() ?? DEFAULT_LOCALE;
const listeners = new Set<() => void>();

export function getLocale(): Locale {
  return current;
}

/** <html lang> says the page's language — for screen readers, hyphenation, spell checking. */
export function applyLocale(): void {
  if (typeof document !== "undefined") document.documentElement.lang = current;
}

/** Switches the interface; remembers the choice for the next signed-out visit. */
export function setLocale(next: Locale): void {
  if (browser) {
    try {
      window.localStorage.setItem(KEY, next);
    } catch {
      /* private mode: the choice lasts for this page */
    }
  }
  if (next === current) return;
  current = next;
  applyLocale();
  for (const l of listeners) l();
}

function subscribe(l: () => void) {
  listeners.add(l);
  return () => listeners.delete(l);
}

export function useLocale(): Locale {
  return useSyncExternalStore(subscribe, getLocale, getLocale);
}

// ── Formatting ────────────────────────────────────────────────────────

const plurals = new Map<Locale, Intl.PluralRules>();
const numbers = new Map<Locale, Intl.NumberFormat>();

function pluralRules(loc: Locale): Intl.PluralRules {
  let r = plurals.get(loc);
  if (!r) plurals.set(loc, (r = new Intl.PluralRules(loc)));
  return r;
}

export function formatNumber(n: number, loc: Locale = current): string {
  let f = numbers.get(loc);
  if (!f) numbers.set(loc, (f = new Intl.NumberFormat(loc)));
  return f.format(n);
}

/** The date-fns locale of a language: month and day names, week start. */
export function dateLocale(loc: Locale = current): DateLocale {
  return DATE_LOCALES[loc];
}

/** "Jeudi 24 septembre": the first letter up, as a heading or a label starts. */
export function capitalize(s: string, loc: Locale = current): string {
  return s ? s.charAt(0).toLocaleUpperCase(loc) + s.slice(1) : s;
}

type Values = Record<string, unknown> | undefined;

/** The form of a message for these values: a plural picks by "count". */
function form(loc: Locale, key: Key, values: Values): string {
  const msg: string | Plural = DICTS[loc][key] ?? fr[key];
  if (typeof msg === "string") return msg;
  const count = Number(values?.count ?? 0);
  // French says "0 tâche", "1 tâche", "2 tâches" (and "many" for millions,
  // which reads fine as "other").
  return pluralRules(loc).select(count) === "one" ? msg.one : msg.other;
}

const HOLE = /\{(\w+)\}/g;

export function translate(loc: Locale, key: Key, values?: Values): string {
  const s = form(loc, key, values);
  if (!values) return s;
  return s.replace(HOLE, (whole, name: string) => {
    const v = values[name];
    if (v === undefined || v === null) return whole;
    return typeof v === "number" ? formatNumber(v, loc) : String(v);
  });
}

/** The message with React nodes in its holes — a link, a bold name — never HTML. */
export function translateRich(
  loc: Locale,
  key: Key,
  values: Record<string, ReactNode>,
): ReactNode {
  const s = form(loc, key, values);
  const parts: ReactNode[] = [];
  let at = 0;
  for (const m of s.matchAll(HOLE)) {
    if (m.index > at) parts.push(s.slice(at, m.index));
    const v = values[m[1]!];
    parts.push(
      typeof v === "number" ? formatNumber(v, loc) : v === undefined ? m[0] : v,
    );
    at = m.index + m[0].length;
  }
  if (at < s.length) parts.push(s.slice(at));
  // Spread as children: static positions, no keys needed.
  return createElement(Fragment, null, ...parts);
}

export type RichValues<K extends Key> = { [P in Placeholder<K>]: ReactNode };

/** A translator bound to one language: t("view.today"), t("group.members", { count: 3 }). */
export interface T {
  <K extends Key>(key: K, ...args: Args<K>): string;
  rich<K extends Key>(key: K, values: RichValues<K>): ReactNode;
  readonly locale: Locale;
}

function translator(loc: Locale): T {
  const t = ((key: Key, values?: Values) => translate(loc, key, values)) as T;
  t.rich = (key, values) =>
    translateRich(loc, key, values as Record<string, ReactNode>);
  (t as { locale: Locale }).locale = loc;
  return t;
}

const TRANSLATORS: Record<Locale, T> = {
  fr: translator("fr"),
  en: translator("en"),
};

/** The translator of a language — or of the current one, outside React. */
export function tr(loc: Locale = current): T {
  return TRANSLATORS[loc];
}

/** The translator of the current language; the component re-renders when it changes. */
export function useT(): T {
  return TRANSLATORS[useLocale()];
}
