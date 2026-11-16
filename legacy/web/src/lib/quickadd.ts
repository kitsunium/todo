// The quick-add language, French first and English too, whatever the
// interface's language: "Appeler Marco demain à 10h !haute #lancement @sam",
// "Call Sam tomorrow at 5pm !high #launch @sam". Pure and unit-tested:
// parseQuickAdd(input, now, options).
//
//   priority  !urgente !haute !moyenne !basse — !urgent !high !medium !low —
//             or !1 … !4 (1 = urgent)
//   due       aujourd’hui · demain · après-demain · lundi…dimanche (prochain) ·
//             (la) semaine prochaine · dans 3 jours · dans 2 semaines —
//             today · tomorrow · mon…sun · next week · in 3 days · in 2 weeks —
//             optionally with a time: à 17h · 17h30 · à 17:30 · à midi —
//             at 5pm · 5:30pm · 17:30 · at 9 (a time alone means its next occurrence)
//   group     #name  (matched against the groups, ignoring case, accents, spaces and dashes)
//   person    @name  (first name, full name or email local part)
//
// A #group or @person that matches nothing stays in the title.
import { addDays, startOfDay } from "date-fns";
import type { Priority } from "../api/types";
import type { Locale } from "../i18n/types";
import { dueAt, nextWeek, nextWeekday } from "./dates";

export type TokenKind = "priority" | "due" | "group" | "contact";

export type Token = {
  kind: TokenKind;
  start: number;
  end: number;
  raw: string;
  /** For #group and @person: whether it names someone or something known. */
  resolved: boolean;
};

export type Candidate = { id: string; name: string; email?: string };

export type QuickAdd = {
  title: string;
  priority?: Priority;
  due?: { date: Date; hasTime: boolean };
  group?: Candidate;
  groupQuery?: string;
  contact?: Candidate;
  contactQuery?: string;
  /** The tokens that were understood, in input order. */
  tokens: Token[];
};

export type QuickAddOptions = {
  groups?: readonly Candidate[];
  contacts?: readonly Candidate[];
  /** Kinds to leave as plain text (the person dismissed their chip). */
  ignore?: ReadonlySet<TokenKind>;
};

const START = String.raw`(?<=^|\s)`;
const END = String.raw`(?=$|[\s,.;:!?)])`;

const PRIORITY_WORDS: Record<string, Priority> = {
  urgente: 1, urgent: 1, "1": 1, p1: 1,
  haute: 2, haut: 2, high: 2, "2": 2, p2: 2,
  moyenne: 3, moyen: 3, medium: 3, med: 3, "3": 3, p3: 3,
  basse: 4, bas: 4, low: 4, "4": 4, p4: 4,
};
const alternatives = (words: string[]) => [...words].sort((a, b) => b.length - a.length).join("|");

const PRIORITY_RE = new RegExp(`${START}!(${alternatives(Object.keys(PRIORITY_WORDS))})${END}`, "giu");
const GROUP_RE = new RegExp(`${START}#([\\p{L}\\p{N}][\\p{L}\\p{N}_-]*)${END}`, "gu");
const CONTACT_RE = new RegExp(`${START}@([\\p{L}\\p{N}][\\p{L}\\p{N}._-]*)${END}`, "gu");

const WEEKDAYS_EN: Record<string, number> = {
  sunday: 0, sun: 0,
  monday: 1, mon: 1,
  tuesday: 2, tues: 2, tue: 2,
  wednesday: 3, wed: 3,
  thursday: 4, thurs: 4, thur: 4, thu: 4,
  friday: 5, fri: 5,
  saturday: 6, sat: 6,
};
// Full names only: "mer", "jeu" and "sam" are also a sea, a game and a name.
const WEEKDAYS_FR: Record<string, number> = {
  dimanche: 0, lundi: 1, mardi: 2, mercredi: 3, jeudi: 4, vendredi: 5, samedi: 6,
};
const WEEKDAYS: Record<string, number> = { ...WEEKDAYS_EN, ...WEEKDAYS_FR };

const DAY =
  String.raw`aujourd['’]?hui|apr[eè]s[\s-]demain|demain|(?:la\s+)?semaine\s+prochaine|dans\s+(?:\d{1,3}|une?)\s+(?:jours?|semaines?)|` +
  `(?:${alternatives(Object.keys(WEEKDAYS_FR))})(?:\\s+prochain)?|` +
  String.raw`today|tomorrow|tmrw|tmr|next\s+week|in\s+\d{1,3}\s+(?:days?|weeks?)|` +
  `(?:next\\s+)?(?:${alternatives(Object.keys(WEEKDAYS_EN))})`;
const TIME =
  String.raw`(?:at\s+|[àa]\s+)?\d{1,2}(?::\d{2})?\s?(?:am|pm)|(?:at\s+|[àa]\s+)?\d{1,2}:\d{2}|` +
  String.raw`(?:[àa]\s+)?\d{1,2}\s?h(?:\d{2})?|(?:at\s+|[àa]\s+)?(?:midi|noon)|at\s+\d{1,2}`;
const DUE_RE = new RegExp(`${START}(?:(${DAY})(?:\\s+(${TIME}))?|(${TIME})(?:\\s+(${DAY}))?)${END}`, "giu");

// "dans 2h", "en 24h", "in 2h", "for 1h": a duration, not a time of day.
const DURATION_BEFORE = /(?:^|\s)(?:dans|en|pendant|sous|in|for|within)\s+$/iu;

/** The token that sets a priority, in a language: "!haute", "!high". */
export function priorityToken(p: Exclude<Priority, 0>, loc: Locale): string {
  const words: Record<Locale, Record<Exclude<Priority, 0>, string>> = {
    fr: { 1: "!urgente", 2: "!haute", 3: "!moyenne", 4: "!basse" },
    en: { 1: "!urgent", 2: "!high", 3: "!medium", 4: "!low" },
  };
  return words[loc][p];
}

/** Parses "17h", "17h30", "à 17:30", "à midi", "5pm", "at 17:30", "at 9", "5:30 pm". */
export function parseTime(s: string): { h: number; m: number } | null {
  const w = s.trim().toLowerCase().replace(/^(?:at|à|a)\s+/u, "");
  if (w === "midi" || w === "noon") return { h: 12, m: 0 };
  const fr = /^(\d{1,2})\s?h(?:\s?(\d{2}))?$/.exec(w);
  if (fr) {
    const h = Number(fr[1]);
    const min = fr[2] ? Number(fr[2]) : 0;
    return h > 23 || min > 59 ? null : { h, m: min };
  }
  const m = /^(\d{1,2})(?::(\d{2}))?\s?(am|pm)?$/.exec(w);
  if (!m) return null;
  let h = Number(m[1]);
  const min = m[2] ? Number(m[2]) : 0;
  const half = m[3];
  if (min > 59) return null;
  if (half) {
    if (h < 1 || h > 12) return null;
    h = (h % 12) + (half === "pm" ? 12 : 0);
  } else if (h > 23) return null;
  return { h, m: min };
}

/** Parses a day word, in French or in English, into the start of that day. */
export function parseDay(s: string, now: Date): Date | null {
  const w = s.trim().toLowerCase().replace(/\s+/g, " ").replace(/’/g, "'");
  const today = startOfDay(now);
  if (w === "today" || /^aujourd'?hui$/.test(w)) return today;
  if (w === "tomorrow" || w === "tmrw" || w === "tmr" || w === "demain") return addDays(today, 1);
  if (/^apr[eè]s[ -]demain$/.test(w)) return addDays(today, 2);
  if (w === "next week" || /^(?:la )?semaine prochaine$/.test(w)) return nextWeek(now);
  const inN = /^(?:in|dans) (\d{1,3}|une?) (days?|weeks?|jours?|semaines?)$/.exec(w);
  if (inN) {
    const n = /^\d/.test(inN[1]!) ? Number(inN[1]) : 1;
    return addDays(today, n * (/^(?:week|semaine)/.test(inN[2]!) ? 7 : 1));
  }
  const wd = WEEKDAYS[w.replace(/^next /, "").replace(/ prochain$/, "")];
  if (wd !== undefined) return nextWeekday(now, wd);
  return null;
}

export function normalize(s: string): string {
  return s
    .normalize("NFD")
    .replace(/\p{M}/gu, "")
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]/gu, "");
}

function keysOf(c: Candidate, withEmail: boolean): string[] {
  const keys = [normalize(c.name)];
  const words = c.name.trim().split(/\s+/);
  if (words.length > 1) keys.push(...words.map(normalize));
  if (withEmail && c.email) keys.push(normalize(c.email.split("@")[0] ?? ""));
  return keys.filter(Boolean);
}

/** The one candidate a query names: exact match first, else a unique prefix. */
export function resolve(query: string, candidates: readonly Candidate[], withEmail = false): Candidate | undefined {
  const q = normalize(query);
  if (!q) return undefined;
  const exact = candidates.filter((c) => keysOf(c, withEmail).includes(q));
  if (exact.length === 1) return exact[0];
  if (exact.length > 1) return undefined;
  const prefix = candidates.filter((c) => keysOf(c, withEmail).some((k) => k.startsWith(q)));
  return prefix.length === 1 ? prefix[0] : undefined;
}

/** Candidates matching a partial query, best first: for autocomplete. */
export function suggest(query: string, candidates: readonly Candidate[], withEmail = false, limit = 6): Candidate[] {
  const q = normalize(query);
  const scored: { c: Candidate; score: number }[] = [];
  for (const c of candidates) {
    const keys = keysOf(c, withEmail);
    let score = -1;
    if (!q) score = 1;
    else if (keys.some((k) => k === q)) score = 3;
    else if (keys.some((k) => k.startsWith(q))) score = 2;
    else if (keys.some((k) => k.includes(q))) score = 1;
    if (score >= 0) scored.push({ c, score });
  }
  scored.sort((a, b) => b.score - a.score || a.c.name.localeCompare(b.c.name));
  return scored.slice(0, limit).map((s) => s.c);
}

/** The #group or @person being typed at the caret, for autocomplete. */
export function mentionAt(input: string, caret: number): { kind: "group" | "contact"; query: string; start: number } | null {
  let i = caret;
  while (i > 0 && !/\s/.test(input[i - 1]!)) i--;
  const word = input.slice(i, caret);
  if (word.length < 1) return null;
  const sigil = word[0];
  if (sigil !== "#" && sigil !== "@") return null;
  const query = word.slice(1);
  if (!/^[\p{L}\p{N}._-]*$/u.test(query)) return null;
  return { kind: sigil === "#" ? "group" : "contact", query, start: i };
}

/** The text to insert for a completed mention: "#book-club ". */
export function mentionText(kind: "group" | "contact", c: Candidate): string {
  const base = kind === "contact" ? (c.name.trim().split(/\s+/)[0] ?? c.name) : c.name.trim();
  return `${kind === "group" ? "#" : "@"}${base.replace(/\s+/g, "-")} `;
}

function lastMatch(re: RegExp, input: string): RegExpExecArray | null {
  re.lastIndex = 0;
  let last: RegExpExecArray | null = null;
  for (let m = re.exec(input); m; m = re.exec(input)) last = m;
  return last;
}

export function parseQuickAdd(input: string, now: Date, opts: QuickAddOptions = {}): QuickAdd {
  const ignore = opts.ignore ?? new Set<TokenKind>();
  const out: QuickAdd = { title: "", tokens: [] };
  const taken: Token[] = [];
  const overlaps = (s: number, e: number) => taken.some((t) => s < t.end && e > t.start);

  if (!ignore.has("priority")) {
    const m = lastMatch(PRIORITY_RE, input);
    if (m) {
      out.priority = PRIORITY_WORDS[m[1]!.toLowerCase()];
      taken.push({ kind: "priority", start: m.index, end: m.index + m[0].length, raw: m[0], resolved: true });
    }
  }

  if (!ignore.has("due")) {
    DUE_RE.lastIndex = 0;
    let found: { start: number; end: number; day: Date | null; time: { h: number; m: number } | null } | null = null;
    for (let m = DUE_RE.exec(input); m; m = DUE_RE.exec(input)) {
      const [whole, dayA, timeA, timeB, dayB] = m;
      const dayText = dayA ?? dayB;
      const timeText = timeA ?? timeB;
      let start = m.index;
      let end = m.index + whole.length;
      const day = dayText ? parseDay(dayText, now) : null;
      let time = timeText ? parseTime(timeText) : null;
      if (timeText && !time) {
        // An impossible time ("at 27"): keep the day alone, if any.
        if (!dayText || !day) continue;
        const at = whole.toLowerCase().indexOf(dayText.toLowerCase());
        start = m.index + at;
        end = start + dayText.length;
        time = null;
      }
      if (dayText && !day) continue;
      if (!dayText && DURATION_BEFORE.test(input.slice(0, start))) continue;
      if (overlaps(start, end)) continue;
      found = { start, end, day, time };
    }
    if (found) {
      let date: Date;
      if (found.day) date = dueAt(found.day, found.time ?? undefined);
      else {
        date = dueAt(now, found.time!);
        if (date.getTime() <= now.getTime()) date = dueAt(addDays(now, 1), found.time!);
      }
      out.due = { date, hasTime: !!found.time };
      taken.push({ kind: "due", start: found.start, end: found.end, raw: input.slice(found.start, found.end), resolved: true });
    }
  }

  if (!ignore.has("group")) {
    const m = lastMatch(GROUP_RE, input);
    if (m && !overlaps(m.index, m.index + m[0].length)) {
      const query = m[1]!;
      const g = resolve(query, opts.groups ?? []);
      if (g) out.group = g;
      else out.groupQuery = query;
      taken.push({ kind: "group", start: m.index, end: m.index + m[0].length, raw: m[0], resolved: !!g });
    }
  }

  if (!ignore.has("contact")) {
    const m = lastMatch(CONTACT_RE, input);
    if (m && !overlaps(m.index, m.index + m[0].length)) {
      const query = m[1]!.replace(/\.+$/, "");
      const c = resolve(query, opts.contacts ?? [], true);
      if (c) out.contact = c;
      else out.contactQuery = query;
      const end = m.index + 1 + query.length;
      taken.push({ kind: "contact", start: m.index, end, raw: input.slice(m.index, end), resolved: !!c });
    }
  }

  taken.sort((a, b) => a.start - b.start);
  out.tokens = taken;
  let title = "";
  let at = 0;
  for (const t of taken) {
    if (!t.resolved) continue;
    title += input.slice(at, t.start);
    at = t.end;
  }
  title += input.slice(at);
  out.title = title
    .replace(/\s+/g, " ")
    .trim()
    .replace(/[\s,;:–—-]+$/u, "")
    .replace(/^[\s,;:–—-]+/u, "");
  return out;
}
