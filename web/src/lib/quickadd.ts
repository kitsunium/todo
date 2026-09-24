// The quick-add language: "Call Sam tomorrow at 5pm !high #launch @sam".
// Pure and unit-tested: parseQuickAdd(input, now, options).
//
//   priority  !urgent !high !medium !low, or !1 … !4 (1 = urgent)
//   due       today · tomorrow · mon…sun · next week · in 3 days · in 2 weeks,
//             optionally with a time: at 5pm · 5:30pm · 17:30 · at 9
//             (a time alone means its next occurrence)
//   group     #name  (matched against the groups, ignoring case, spaces and dashes)
//   person    @name  (first name, full name or email local part)
//
// A #group or @person that matches nothing stays in the title.
import { addDays, startOfDay } from "date-fns";
import type { Priority } from "../api/types";
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

const PRIORITY_RE = new RegExp(`${START}!(urgent|high|medium|med|low|p?[1-4])${END}`, "giu");
const GROUP_RE = new RegExp(`${START}#([\\p{L}\\p{N}][\\p{L}\\p{N}_-]*)${END}`, "gu");
const CONTACT_RE = new RegExp(`${START}@([\\p{L}\\p{N}][\\p{L}\\p{N}._-]*)${END}`, "gu");

const WEEKDAYS: Record<string, number> = {
  sunday: 0, sun: 0,
  monday: 1, mon: 1,
  tuesday: 2, tues: 2, tue: 2,
  wednesday: 3, wed: 3,
  thursday: 4, thurs: 4, thur: 4, thu: 4,
  friday: 5, fri: 5,
  saturday: 6, sat: 6,
};
const WEEKDAY_ALT = Object.keys(WEEKDAYS)
  .sort((a, b) => b.length - a.length)
  .join("|");
const DAY = String.raw`today|tomorrow|tmrw|tmr|next\s+week|in\s+\d{1,3}\s+(?:days?|weeks?)|` + WEEKDAY_ALT;
const TIME = String.raw`(?:at\s+)?\d{1,2}(?::\d{2})?\s?(?:am|pm)|(?:at\s+)?\d{1,2}:\d{2}|at\s+\d{1,2}`;
const DUE_RE = new RegExp(`${START}(?:(${DAY})(?:\\s+(${TIME}))?|(${TIME})(?:\\s+(${DAY}))?)${END}`, "giu");

const PRIORITY_WORDS: Record<string, Priority> = {
  urgent: 1, "1": 1, p1: 1,
  high: 2, "2": 2, p2: 2,
  medium: 3, med: 3, "3": 3, p3: 3,
  low: 4, "4": 4, p4: 4,
};

export const PRIORITY_TOKENS: Record<Exclude<Priority, 0>, string> = { 1: "!urgent", 2: "!high", 3: "!medium", 4: "!low" };

/** Parses "5pm", "at 17:30", "at 9", "5:30 pm". */
export function parseTime(s: string): { h: number; m: number } | null {
  const m = /^(?:at\s+)?(\d{1,2})(?::(\d{2}))?\s?(am|pm)?$/i.exec(s.trim());
  if (!m) return null;
  let h = Number(m[1]);
  const min = m[2] ? Number(m[2]) : 0;
  const half = m[3]?.toLowerCase();
  if (min > 59) return null;
  if (half) {
    if (h < 1 || h > 12) return null;
    h = (h % 12) + (half === "pm" ? 12 : 0);
  } else if (h > 23) return null;
  return { h, m: min };
}

/** Parses a day word into the start of that day. */
export function parseDay(s: string, now: Date): Date | null {
  const w = s.trim().toLowerCase().replace(/\s+/g, " ");
  const today = startOfDay(now);
  if (w === "today") return today;
  if (w === "tomorrow" || w === "tmrw" || w === "tmr") return addDays(today, 1);
  if (w === "next week") return nextWeek(now);
  const inN = /^in (\d{1,3}) (days?|weeks?)$/.exec(w);
  if (inN) return addDays(today, Number(inN[1]) * (inN[2]!.startsWith("week") ? 7 : 1));
  const wd = WEEKDAYS[w];
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
