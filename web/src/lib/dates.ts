// Dates as a person reads them, in their language. Pure: every function
// takes "now", and the language (the current one by default).
import {
  addDays,
  differenceInCalendarDays,
  endOfWeek,
  format,
  formatDistanceToNowStrict,
  isSameDay,
  isSameYear,
  startOfDay,
} from "date-fns";
import type { Priority, Task } from "../api/types";
import { capitalize, dateLocale, getLocale, tr, type Key, type Locale } from "../i18n";

/** A due without a time of day is stored as 23:59 local time. */
export const DATE_ONLY_HOUR = 23;
export const DATE_ONLY_MINUTE = 59;

export function isDateOnly(d: Date): boolean {
  return d.getHours() === DATE_ONLY_HOUR && d.getMinutes() === DATE_ONLY_MINUTE;
}

/** The instant a due means: that day at time, or at 23:59 when there is none. */
export function dueAt(day: Date, time?: { h: number; m: number }): Date {
  const d = new Date(day.getFullYear(), day.getMonth(), day.getDate());
  if (time) d.setHours(time.h, time.m, 0, 0);
  else d.setHours(DATE_ONLY_HOUR, DATE_ONLY_MINUTE, 0, 0);
  return d;
}

export type DatePattern = Extract<Key, `date.pattern.${string}`>;

/** A date-fns pattern of the dictionary ("d MMM" in French, "MMM d" in English), applied in that language. */
export function formatDate(d: Date, pattern: DatePattern, loc: Locale = getLocale()): string {
  return format(d, tr(loc)(pattern), { locale: dateLocale(loc) });
}

/** Whether the browser's clock reads 12 hours ("5 PM") or 24 ("17:00"). */
function uses12h(): boolean {
  try {
    const o = new Intl.DateTimeFormat(undefined, { hour: "numeric" }).resolvedOptions();
    return o.hourCycle ? o.hourCycle === "h12" || o.hourCycle === "h11" : !!o.hour12;
  } catch {
    return true;
  }
}
const H12 = uses12h();
const FMT12 = new Intl.DateTimeFormat("en-US", { hour: "numeric", minute: "2-digit" });
const FMT24 = new Intl.DateTimeFormat("en-GB", { hour: "2-digit", minute: "2-digit", hourCycle: "h23" });

/**
 * "17:00" in French, always. In English the person's own clock: "5:00 PM",
 * or "17:00" where the browser counts 24 hours.
 */
export function formatTime(d: Date, loc: Locale = getLocale(), twelve = H12): string {
  return loc === "en" && twelve ? FMT12.format(d) : FMT24.format(d);
}

/** "Today", "Tomorrow 5 PM", "Sat", "Oct 3", "Yesterday" — "Aujourd’hui", "Demain 17:00", "sam.", "3 oct.", "Hier". */
export function formatDue(
  iso: string,
  now: Date,
  loc: Locale = getLocale(),
): { label: string; overdue: boolean; soon: boolean } {
  const t = tr(loc);
  const d = new Date(iso);
  const days = differenceInCalendarDays(d, now);
  const withTime = (day: string) => (isDateOnly(d) ? day : t("date.withTime", { day, time: formatTime(d, loc) }));
  const overdue = d.getTime() < now.getTime();
  let label: string;
  if (days === 0) label = withTime(t("date.today"));
  else if (days === 1) label = withTime(t("date.tomorrow"));
  else if (days === -1) label = t("date.yesterday");
  else if (days > 1 && days < 7) label = withTime(formatDate(d, "date.pattern.weekday", loc));
  else if (days < -1 && days > -7) label = formatDate(d, "date.pattern.weekday", loc);
  else label = formatDate(d, isSameYear(d, now) ? "date.pattern.dayMonth" : "date.pattern.dayMonthYear", loc);
  return { label, overdue, soon: !overdue && days === 0 };
}

/** A day written out in full, as a heading starts: "Thursday, September 24", "Jeudi 24 septembre". */
export function formatDayLong(d: Date, now: Date, loc: Locale = getLocale()): string {
  return capitalize(formatDate(d, isSameYear(d, now) ? "date.pattern.long" : "date.pattern.longYear", loc), loc);
}

/** A due written out in full: "Thursday, September 24 at 5 PM", "Jeudi 24 septembre à 17:00". */
export function formatDueLong(iso: string, loc: Locale = getLocale(), now = new Date()): string {
  const d = new Date(iso);
  const day = formatDayLong(d, now, loc);
  return isDateOnly(d) ? day : tr(loc)("date.at", { day, time: formatTime(d, loc) });
}

const RELATIVE = new Map<Locale, Intl.RelativeTimeFormat>();
function relative(loc: Locale): Intl.RelativeTimeFormat {
  let r = RELATIVE.get(loc);
  // English says "5m ago", French "il y a 5 min": each language's own short form.
  if (!r) RELATIVE.set(loc, (r = new Intl.RelativeTimeFormat(loc, { style: loc === "en" ? "narrow" : "short", numeric: "always" })));
  return r;
}

/**
 * How long ago, short: "just now", "5m ago", "3h ago", then the day —
 * "à l’instant", "il y a 5 min", "il y a 3 h". As a label the day is
 * "Yesterday", "Mon", "Sep 21"; inside a sentence ("Modifiée {ago}") it is
 * "hier", "lundi", "le 21 sept.".
 */
export function timeAgo(iso: string, now: Date, loc: Locale = getLocale(), inline = false): string {
  const t = tr(loc);
  const d = new Date(iso);
  const s = Math.round((now.getTime() - d.getTime()) / 1000);
  if (s < 45) return t("date.justNow");
  if (s < 3600) return relative(loc).format(-Math.max(1, Math.round(s / 60)), "minute");
  if (isSameDay(d, now)) return relative(loc).format(-Math.round(s / 3600), "hour");
  const days = differenceInCalendarDays(now, d);
  if (days === 1) return t(inline ? "date.yesterdayInline" : "date.yesterday");
  if (days < 7) return inline && loc === "fr" ? format(d, "EEEE", { locale: dateLocale(loc) }) : formatDate(d, "date.pattern.weekday", loc);
  const day = formatDate(d, isSameYear(d, now) ? "date.pattern.dayMonth" : "date.pattern.dayMonthYear", loc);
  return inline ? t("date.on", { day }) : day;
}

const DAYS = new Map<Locale, Intl.RelativeTimeFormat>();

/**
 * How far a day is, in whole days: "dans 2 jours", "il y a 3 jours", "in 2
 * days". Nothing for yesterday, today and tomorrow: those have their own word.
 */
export function relativeDays(iso: string, now: Date, loc: Locale = getLocale()): string {
  const days = differenceInCalendarDays(new Date(iso), now);
  if (Math.abs(days) <= 1) return "";
  let r = DAYS.get(loc);
  if (!r) DAYS.set(loc, (r = new Intl.RelativeTimeFormat(loc, { numeric: "always" })));
  return r.format(days, "day");
}

/** "3 days ago", "il y a 3 jours", "in 2 days", "dans 2 jours". */
export function distance(iso: string, loc: Locale = getLocale()): string {
  return formatDistanceToNowStrict(new Date(iso), { addSuffix: true, locale: dateLocale(loc) });
}

/** A day heading: "Today", "Yesterday", "Monday, September 21" — "Aujourd’hui", "Hier", "Lundi 21 septembre". */
export function dayHeading(d: Date, now: Date, loc: Locale = getLocale()): string {
  const t = tr(loc);
  const days = differenceInCalendarDays(now, d);
  if (days === 0) return t("date.today");
  if (days === 1) return t("date.yesterday");
  return formatDayLong(d, now, loc);
}

// ── Sections ──────────────────────────────────────────────────────────

export type SectionKey =
  | "overdue"
  | "today"
  | "tomorrow"
  | "week"
  | "later"
  | "none"
  | "completed"
  | `day:${string}`
  | "earlier";

export type Section = { key: SectionKey; label: string; tone?: "danger"; tasks: Task[] };

const SECTION_LABELS: Record<Exclude<SectionKey, `day:${string}`>, Key> = {
  overdue: "section.overdue",
  today: "section.today",
  tomorrow: "section.tomorrow",
  week: "section.week",
  later: "section.later",
  none: "section.none",
  completed: "section.completed",
  earlier: "section.earlier",
};

/** Urgent first, "none" last. */
export function priorityRank(p: Priority): number {
  return p === 0 ? 5 : p;
}

function byDue(a: Task, b: Task): number {
  const da = a.due ? Date.parse(a.due) : 0;
  const db = b.due ? Date.parse(b.due) : 0;
  return da - db || priorityRank(a.priority) - priorityRank(b.priority) || a.createdAt.localeCompare(b.createdAt);
}

function byPriorityNewest(a: Task, b: Task): number {
  return priorityRank(a.priority) - priorityRank(b.priority) || b.createdAt.localeCompare(a.createdAt);
}

function byCompletion(a: Task, b: Task): number {
  return (b.completedAt ?? b.updatedAt).localeCompare(a.completedAt ?? a.updatedAt);
}

export function isOverdue(t: Task, now: Date): boolean {
  if (t.status === "done" || t.status === "archived") return false;
  if (t.status === "overdue") return true;
  return !!t.due && Date.parse(t.due) < now.getTime();
}

/**
 * Splits a list into Overdue / Today / Tomorrow / This week / Later /
 * No date, then Completed for done tasks. The week ends on Sunday.
 *
 * pins keeps a task in the section it was in — a task just completed stays
 * in place, checked, for the moment it takes to see it happen.
 */
export function groupByDue(
  list: readonly Task[],
  now: Date,
  pins: Readonly<Record<string, SectionKey>> = {},
  loc: Locale = getLocale(),
): Section[] {
  const buckets = new Map<SectionKey, Task[]>();
  const put = (k: SectionKey, t: Task) => {
    const b = buckets.get(k);
    if (b) b.push(t);
    else buckets.set(k, [t]);
  };
  const weekEnd = endOfWeek(now, { weekStartsOn: 1 });
  for (const t of list) {
    const pin = pins[t.id];
    if (pin) put(pin, t);
    else if (t.status === "archived") continue;
    else if (t.status === "done") put("completed", t);
    else if (isOverdue(t, now)) put("overdue", t);
    else if (!t.due) put("none", t);
    else {
      const d = new Date(t.due);
      const days = differenceInCalendarDays(d, now);
      if (days <= 0) put("today", t);
      else if (days === 1) put("tomorrow", t);
      else if (d.getTime() <= weekEnd.getTime()) put("week", t);
      else put("later", t);
    }
  }
  const order: SectionKey[] = ["overdue", "today", "tomorrow", "week", "later", "none", "completed"];
  const out: Section[] = [];
  for (const key of order) {
    const tasks = buckets.get(key);
    if (!tasks?.length) continue;
    tasks.sort(key === "none" ? byPriorityNewest : key === "completed" ? byCompletion : byDue);
    const s: Section = { key, label: tr(loc)(SECTION_LABELS[key as keyof typeof SECTION_LABELS]), tasks };
    if (key === "overdue") s.tone = "danger";
    out.push(s);
  }
  return out;
}

/** Completed tasks by the day they were completed: Today, Yesterday, Monday…, Earlier. */
export function groupByCompletion(
  list: readonly Task[],
  now: Date,
  pinnedAt: Readonly<Record<string, string>> = {},
  loc: Locale = getLocale(),
): Section[] {
  const say = tr(loc);
  const out: Section[] = [];
  const index = new Map<string, Section>();
  const when = (t: Task) => pinnedAt[t.id] ?? t.completedAt ?? t.updatedAt;
  const sorted = [...list].sort((a, b) => when(b).localeCompare(when(a)));
  for (const t of sorted) {
    const d = new Date(when(t));
    const days = differenceInCalendarDays(startOfDay(now), startOfDay(d));
    let key: SectionKey;
    let label: string;
    if (days <= 6) {
      key = `day:${format(d, "yyyy-MM-dd")}`;
      label =
        days === 0
          ? say("date.today")
          : days === 1
            ? say("date.yesterday")
            : capitalize(format(d, "EEEE", { locale: dateLocale(loc) }), loc);
    } else {
      key = "earlier";
      label = say(SECTION_LABELS.earlier);
    }
    let s = index.get(key);
    if (!s) {
      s = { key, label, tasks: [] };
      index.set(key, s);
      out.push(s);
    }
    s.tasks.push(t);
  }
  return out;
}

/** The next occurrence of a weekday (0 = Sunday), strictly after today. */
export function nextWeekday(now: Date, weekday: number): Date {
  const delta = (weekday - now.getDay() + 7) % 7 || 7;
  return addDays(startOfDay(now), delta);
}

/** Next Monday. */
export function nextWeek(now: Date): Date {
  return nextWeekday(now, 1);
}
