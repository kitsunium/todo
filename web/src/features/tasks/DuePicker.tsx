import {
  addDays,
  addMonths,
  eachDayOfInterval,
  endOfMonth,
  endOfWeek,
  format,
  isSameDay,
  isSameMonth,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";
import { CalendarArrowUp, CalendarX, ChevronLeft, ChevronRight, Clock, Star, Sunrise, X } from "lucide-react";
import { useRef, useState, type KeyboardEvent, type ReactNode } from "react";
import { capitalize, dateLocale, useT } from "../../i18n";
import { cn } from "../../lib/cn";
import { dueAt, formatDate, formatTime, isDateOnly, nextWeek } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { parseTime } from "../../lib/quickadd";

function Shortcut({ icon, label, hint, onClick }: { icon: ReactNode; label: string; hint?: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex h-8 w-full items-center gap-2.5 rounded-md px-2 text-sm text-fg outline-none hover:bg-hover focus-visible:bg-hover"
    >
      <span className="flex size-4 items-center justify-center text-fg-3">{icon}</span>
      <span className="flex-1 text-left">{label}</span>
      {hint ? <span className="text-xs text-fg-4">{hint}</span> : null}
    </button>
  );
}

const TIMES = [
  { h: 9, m: 0 },
  { h: 12, m: 0 },
  { h: 17, m: 0 },
  { h: 20, m: 0 },
];

/**
 * Shortcuts, a month calendar and a time. onChange receives an RFC 3339
 * instant, or "" for "no date".
 */
export function DuePicker({
  value,
  onChange,
  onPicked,
}: {
  value?: string | undefined;
  onChange: (iso: string) => void;
  /** A day or a shortcut was chosen: the caller usually closes its popover. */
  onPicked?: () => void;
}) {
  const now = useNow();
  const t = useT();
  const loc = t.locale;
  // The week starts on Monday in French, on Sunday in (US) English.
  const weekStartsOn = dateLocale(loc).options?.weekStartsOn ?? 1;
  const current = value ? new Date(value) : null;
  const [month, setMonth] = useState(() => startOfMonth(current ?? now));
  const [time, setTime] = useState(() => (current && !isDateOnly(current) ? { h: current.getHours(), m: current.getMinutes() } : null));
  const [timeText, setTimeText] = useState(() => (current && !isDateOnly(current) ? formatTime(current, loc) : ""));
  const [timeError, setTimeError] = useState(false);
  const grid = useRef<HTMLDivElement>(null);

  const set = (day: Date, t = time) => onChange(dueAt(day, t ?? undefined).toISOString());
  const pick = (day: Date) => {
    set(day);
    onPicked?.();
  };

  const days = eachDayOfInterval({
    start: startOfWeek(startOfMonth(month), { weekStartsOn }),
    end: endOfWeek(endOfMonth(month), { weekStartsOn }),
  });

  const focusDay = current && isSameMonth(current, month) ? current : isSameMonth(now, month) ? now : month;

  function commitTime(text: string) {
    if (!text.trim()) {
      setTime(null);
      setTimeError(false);
      if (current) set(current, null);
      return;
    }
    const t = parseTime(text.includes(":") || /am|pm/i.test(text) ? text : `at ${text}`);
    if (!t) {
      setTimeError(true);
      return;
    }
    setTimeError(false);
    setTime(t);
    setTimeText(formatTime(dueAt(now, t), loc));
    set(current ?? (dueAt(now, t) > now ? now : addDays(now, 1)), t);
  }

  function onGridKey(e: KeyboardEvent<HTMLDivElement>) {
    const delta = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -7, ArrowDown: 7 }[e.key];
    if (delta === undefined) return;
    e.preventDefault();
    const focused = document.activeElement as HTMLElement | null;
    const iso = focused?.dataset.day;
    if (!iso) return;
    const next = addDays(new Date(iso), delta);
    if (!isSameMonth(next, month)) setMonth(startOfMonth(next));
    requestAnimationFrame(() => grid.current?.querySelector<HTMLButtonElement>(`[data-day="${format(next, "yyyy-MM-dd")}T00:00"]`)?.focus());
  }

  return (
    <div className="w-[280px] p-1.5">
      <Shortcut icon={<Star className="size-4" />} label={t("due.today")} hint={formatDate(now, "date.pattern.weekday", loc)} onClick={() => pick(now)} />
      <Shortcut
        icon={<Sunrise className="size-4" />}
        label={t("due.tomorrow")}
        hint={formatDate(addDays(now, 1), "date.pattern.weekday", loc)}
        onClick={() => pick(addDays(now, 1))}
      />
      <Shortcut
        icon={<CalendarArrowUp className="size-4" />}
        label={t("due.nextWeek")}
        hint={formatDate(nextWeek(now), "date.pattern.weekdayDayMonth", loc)}
        onClick={() => pick(nextWeek(now))}
      />
      {value ? (
        <Shortcut
          icon={<CalendarX className="size-4" />}
          label={t("due.noDate")}
          onClick={() => {
            onChange("");
            onPicked?.();
          }}
        />
      ) : null}

      <div className="-mx-1.5 my-1.5 h-px bg-line-soft" />

      <div className="flex items-center justify-between px-1.5 pt-1 pb-2">
        <span className="text-sm font-semibold text-fg">{capitalize(formatDate(month, "date.pattern.monthYear", loc), loc)}</span>
        <div className="flex items-center">
          <button type="button" aria-label={t("due.previousMonth")} onClick={() => setMonth((m) => addMonths(m, -1))} className="flex size-7 items-center justify-center rounded-md text-fg-3 hover:bg-hover hover:text-fg">
            <ChevronLeft className="size-4" />
          </button>
          <button type="button" aria-label={t("due.nextMonth")} onClick={() => setMonth((m) => addMonths(m, 1))} className="flex size-7 items-center justify-center rounded-md text-fg-3 hover:bg-hover hover:text-fg">
            <ChevronRight className="size-4" />
          </button>
        </div>
      </div>
      <div className="grid grid-cols-7 px-0.5 text-center text-2xs font-medium text-fg-4" aria-hidden="true">
        {days.slice(0, 7).map((d) => (
          <span key={d.toISOString()} className="py-1">
            {capitalize(format(d, "EEEEEE", { locale: dateLocale(loc) }), loc)}
          </span>
        ))}
      </div>
      <div ref={grid} role="grid" aria-label={capitalize(formatDate(month, "date.pattern.monthYear", loc), loc)} className="grid grid-cols-7 gap-y-0.5 px-0.5" onKeyDown={onGridKey}>
        {days.map((d) => {
          const selected = current && isSameDay(d, current);
          const today = isSameDay(d, now);
          const past = d < startOfDay(now);
          const focusable = isSameDay(d, focusDay);
          return (
            <button
              key={d.toISOString()}
              type="button"
              data-day={`${format(d, "yyyy-MM-dd")}T00:00`}
              tabIndex={focusable ? 0 : -1}
              aria-pressed={!!selected}
              aria-label={capitalize(formatDate(d, "date.pattern.long", loc), loc)}
              onClick={() => pick(d)}
              className={cn(
                "relative mx-auto flex size-8 items-center justify-center rounded-lg text-sm tabular outline-none transition-colors",
                "focus-visible:ring-2 focus-visible:ring-ring",
                selected
                  ? "bg-accent font-semibold text-white"
                  : today
                    ? "font-semibold text-accent-ink hover:bg-hover"
                    : isSameMonth(d, month)
                      ? cn("hover:bg-hover", past ? "text-fg-4" : "text-fg")
                      : "text-fg-4/60 hover:bg-hover",
              )}
            >
              {d.getDate()}
              {today && !selected ? <span className="absolute bottom-1 size-1 rounded-full bg-accent" aria-hidden="true" /> : null}
            </button>
          );
        })}
      </div>

      <div className="-mx-1.5 mt-2 mb-1.5 h-px bg-line-soft" />
      <div className="flex items-center gap-2 px-1.5 pb-1">
        <Clock className="size-4 shrink-0 text-fg-3" aria-hidden="true" />
        <input
          value={timeText}
          onChange={(e) => {
            setTimeText(e.target.value);
            setTimeError(false);
          }}
          onBlur={(e) => commitTime(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              commitTime(e.currentTarget.value);
            }
          }}
          placeholder={t("due.timePlaceholder")}
          aria-label={t("due.time")}
          aria-invalid={timeError}
          className={cn(
            "h-8 min-w-0 flex-1 rounded-md bg-transparent px-1.5 text-sm text-fg outline-none placeholder:text-fg-4 focus:bg-inset",
            timeError && "text-danger-ink",
          )}
        />
        {time ? (
          <button
            type="button"
            aria-label={t("due.removeTime")}
            onClick={() => {
              setTimeText("");
              commitTime("");
            }}
            className="flex size-6 items-center justify-center rounded-md text-fg-4 hover:bg-hover hover:text-fg"
          >
            <X className="size-3.5" />
          </button>
        ) : null}
      </div>
      <div className="flex gap-1 px-1.5 pb-1">
        {TIMES.map((t) => (
          <button
            key={t.h}
            type="button"
            onClick={() => {
              setTime(t);
              setTimeText(formatTime(dueAt(now, t), loc));
              setTimeError(false);
              set(current ?? (dueAt(now, t) > now ? now : addDays(now, 1)), t);
            }}
            className={cn(
              "h-6 flex-1 rounded-md text-xs text-fg-2 shadow-[inset_0_0_0_1px_var(--line)] hover:bg-hover",
              time?.h === t.h && time.m === t.m && "bg-accent-soft text-accent-ink shadow-none",
            )}
          >
            {formatTime(dueAt(now, t), loc)}
          </button>
        ))}
      </div>
    </div>
  );
}
