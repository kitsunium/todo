// Dates as a person reads them. Pure: every function takes "now".
import {
  addDays,
  differenceInCalendarDays,
  endOfWeek,
  format,
  isSameDay,
  isSameYear,
  startOfDay,
} from "date-fns";
import type { Priority, Task } from "../api/types";

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

/** Whether the person's locale reads the clock in 12 hours ("5 PM") or 24 ("17:00"). */
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

/** "5:00 PM" — or "17:00" where the clock has 24 hours. One format everywhere; the words stay English. */
export function formatTime(d: Date, twelve = H12): string {
  return (twelve ? FMT12 : FMT24).format(d);
}

/** "Today", "Tomorrow 5 PM", "Sat", "Oct 3", "Yesterday" — a due as a chip says it. */
export function formatDue(iso: string, now: Date): { label: string; overdue: boolean; soon: boolean } {
  const d = new Date(iso);
  const days = differenceInCalendarDays(d, now);
  const time = isDateOnly(d) ? "" : ` ${formatTime(d)}`;
  const overdue = d.getTime() < now.getTime();
  let label: string;
  if (days === 0) label = `Today${time}`;
  else if (days === 1) label = `Tomorrow${time}`;
  else if (days === -1) label = "Yesterday";
  else if (days > 1 && days < 7) label = `${format(d, "EEE")}${time}`;
  else if (days < -1 && days > -7) label = format(d, "EEE");
  else label = isSameYear(d, now) ? format(d, "MMM d") : format(d, "MMM d, yyyy");
  return { label, overdue, soon: !overdue && days === 0 };
}

/** A due written out in full: "Thursday, September 24 at 5 PM". */
export function formatDueLong(iso: string): string {
  const d = new Date(iso);
  const day = format(d, "EEEE, MMMM d");
  return isDateOnly(d) ? day : `${day} at ${formatTime(d)}`;
}

/** "just now", "5m ago", "3h ago", "Yesterday", "Mon", "Sep 21". */
export function timeAgo(iso: string, now: Date): string {
  const d = new Date(iso);
  const s = Math.round((now.getTime() - d.getTime()) / 1000);
  if (s < 45) return "just now";
  if (s < 3600) return `${Math.max(1, Math.round(s / 60))}m ago`;
  if (isSameDay(d, now)) return `${Math.round(s / 3600)}h ago`;
  const days = differenceInCalendarDays(now, d);
  if (days === 1) return "Yesterday";
  if (days < 7) return format(d, "EEE");
  return isSameYear(d, now) ? format(d, "MMM d") : format(d, "MMM d, yyyy");
}

/** A day heading: "Today", "Yesterday", "Monday, September 21". */
export function dayHeading(d: Date, now: Date): string {
  const days = differenceInCalendarDays(now, d);
  if (days === 0) return "Today";
  if (days === 1) return "Yesterday";
  return isSameYear(d, now) ? format(d, "EEEE, MMMM d") : format(d, "EEEE, MMMM d, yyyy");
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

const SECTION_LABELS: Record<string, string> = {
  overdue: "Overdue",
  today: "Today",
  tomorrow: "Tomorrow",
  week: "This week",
  later: "Later",
  none: "No date",
  completed: "Completed",
  earlier: "Earlier",
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
export function groupByDue(list: readonly Task[], now: Date, pins: Readonly<Record<string, SectionKey>> = {}): Section[] {
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
    const s: Section = { key, label: SECTION_LABELS[key]!, tasks };
    if (key === "overdue") s.tone = "danger";
    out.push(s);
  }
  return out;
}

/** Completed tasks by the day they were completed: Today, Yesterday, Monday…, Earlier. */
export function groupByCompletion(list: readonly Task[], now: Date, pinnedAt: Readonly<Record<string, string>> = {}): Section[] {
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
      label = days === 0 ? "Today" : days === 1 ? "Yesterday" : format(d, "EEEE");
    } else {
      key = "earlier";
      label = SECTION_LABELS.earlier!;
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
