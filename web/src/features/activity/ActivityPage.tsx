import { useQueryClient } from "@tanstack/react-query";
import { startOfDay } from "date-fns";
import {
  Activity as ActivityIcon,
  AlarmClock,
  Archive,
  CircleCheck,
  DoorOpen,
  Mail,
  Pencil,
  Plus,
  RotateCcw,
  Trash2,
  UserMinus,
  UserRoundCheck,
  UserRoundPlus,
  UsersRound,
  type LucideIcon,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { Link, useLocation } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage } from "../../api/errors";
import { qk, useActivity, useInvitations } from "../../api/queries";
import type { ActivityEntry } from "../../api/types";
import { ActivityIllo } from "../../components/brand/illustrations";
import { Avatar } from "../../components/ui/avatar";
import { EmptyState, ErrorState } from "../../components/ui/empty";
import { GroupDot } from "../../components/ui/group-color";
import { Skeleton } from "../../components/ui/skeleton";
import { cn } from "../../lib/cn";
import { dayHeading, formatTime } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { InvitationCard } from "../groups/InvitationBanner";
import { Page } from "../shell/Page";

const KINDS: Record<string, { icon: LucideIcon; color: string }> = {
  completed: { icon: CircleCheck, color: "var(--done)" },
  reopened: { icon: RotateCcw, color: "var(--fg-3)" },
  created: { icon: Plus, color: "var(--fg-3)" },
  updated: { icon: Pencil, color: "var(--fg-3)" },
  shared: { icon: UserRoundPlus, color: "#12a594" },
  unshared: { icon: UserMinus, color: "var(--fg-3)" },
  assigned: { icon: UserRoundCheck, color: "#8e4ec6" },
  overdue: { icon: AlarmClock, color: "var(--urgent)" },
  deleted: { icon: Trash2, color: "var(--fg-3)" },
  archived: { icon: Archive, color: "var(--fg-3)" },
  restored: { icon: RotateCcw, color: "var(--fg-3)" },
  requested: { icon: UserRoundPlus, color: "#3e63dd" },
  accepted: { icon: UserRoundCheck, color: "#3e63dd" },
  invited: { icon: Mail, color: "var(--accent)" },
  joined: { icon: UsersRound, color: "var(--accent)" },
  left: { icon: DoorOpen, color: "var(--fg-3)" },
  removed: { icon: UserMinus, color: "var(--fg-3)" },
};

/** Bolds the names the sentence mentions — as React nodes, never HTML. */
function emphasize(text: string, marks: { find: string; render: (s: string) => ReactNode }[]): ReactNode[] {
  const out: ReactNode[] = [];
  let rest = text;
  let key = 0;
  const pending = marks.filter((m) => m.find);
  while (rest) {
    let best: { i: number; m: (typeof pending)[number] } | null = null;
    for (const m of pending) {
      const i = rest.indexOf(m.find);
      if (i >= 0 && (!best || i < best.i)) best = { i, m };
    }
    if (!best) {
      out.push(rest);
      break;
    }
    if (best.i > 0) out.push(rest.slice(0, best.i));
    out.push(<span key={key++}>{best.m.render(best.m.find)}</span>);
    rest = rest.slice(best.i + best.m.find.length);
    pending.splice(pending.indexOf(best.m), 1);
  }
  return out;
}

export function ActivityPage() {
  const q = useActivity();
  const invitations = useInvitations();
  const qc = useQueryClient();
  const now = useNow();
  const [unread, setUnread] = useState<Set<string> | null>(null);
  const marked = useRef(false);

  // What was new when the page opened stays marked as new for this visit…
  useEffect(() => {
    if (q.data && !unread) setUnread(new Set(q.data.filter((e) => !e.read).map((e) => e.id)));
  }, [q.data, unread]);

  // …and the server hears it was seen, a moment after it showed.
  useEffect(() => {
    if (!q.data || marked.current || !q.data.some((e) => !e.read)) return;
    marked.current = true;
    const t = setTimeout(() => {
      void ep.activity.markRead().then(() => {
        qc.setQueryData<ActivityEntry[]>(qk.activity, (l) => l?.map((e) => ({ ...e, read: true })));
        void qc.invalidateQueries({ queryKey: qk.counts });
      });
    }, 1200);
    return () => {
      clearTimeout(t);
      marked.current = false;
    };
  }, [q.data, qc]);

  const days = useMemo(() => {
    const groups: { key: string; label: string; entries: ActivityEntry[] }[] = [];
    for (const e of q.data ?? []) {
      const d = startOfDay(new Date(e.at));
      const key = d.toISOString();
      let g = groups[groups.length - 1];
      if (!g || g.key !== key) {
        g = { key, label: dayHeading(d, now), entries: [] };
        groups.push(g);
      }
      g.entries.push(e);
    }
    return groups;
  }, [q.data, now]);

  const count = unread?.size ?? 0;
  return (
    <Page
      title="Activity"
      icon={<ActivityIcon className="text-accent" strokeWidth={2} />}
      subtitle={count ? `${count} new since your last visit` : "What happened to your tasks, groups and contacts."}
    >
      {invitations.data?.length ? (
        <section className="mb-7">
          <h2 className="mb-2 px-1 text-[13px] font-semibold text-fg">Invitations</h2>
          <div className="flex flex-col gap-2">
            {invitations.data.map((inv) => (
              <InvitationCard key={inv.id} inv={inv} />
            ))}
          </div>
        </section>
      ) : null}
      {q.isPending ? (
        <div className="flex flex-col gap-5 pt-2">
          {[0, 1, 2, 3, 4].map((i) => (
            <div key={i} className="flex items-center gap-3">
              <Skeleton className="size-8 rounded-full" />
              <div className="flex-1 space-y-2">
                <Skeleton className={cn("h-3", i % 2 ? "w-1/2" : "w-2/3")} />
                <Skeleton className="h-2.5 w-16" />
              </div>
            </div>
          ))}
        </div>
      ) : q.isError ? (
        <ErrorState message={errorMessage(q.error)} onRetry={() => void q.refetch()} />
      ) : !q.data.length ? (
        <EmptyState art={<ActivityIllo />} title="All quiet">
          When someone shares, assigns or completes a task with you, you’ll see it here.
        </EmptyState>
      ) : (
        days.map((d) => (
          <section key={d.key} className="mb-4">
            <h2 className="sticky top-0 z-10 -mx-1 bg-sheet/90 px-1 py-2 text-[13px] font-semibold text-fg backdrop-blur-md">{d.label}</h2>
            <ol className="relative">
              <span className="absolute top-3 bottom-3 left-[15.5px] w-px bg-line" aria-hidden="true" />
              {d.entries.map((e) => (
                <Entry key={e.id} e={e} fresh={!!unread?.has(e.id)} />
              ))}
            </ol>
          </section>
        ))
      )}
    </Page>
  );
}

function Entry({ e, fresh }: { e: ActivityEntry; fresh: boolean }) {
  const location = useLocation();
  // Kinds are namespaced ("task.completed", "group.invited"); a group deleted
  // reads like a task deleted.
  const kind = KINDS[e.kind.split(".").pop() ?? e.kind] ?? { icon: ActivityIcon, color: "var(--fg-3)" };
  const Icon = kind.icon;
  const marks: { find: string; render: (s: string) => ReactNode }[] = [];
  if (e.actor) marks.push({ find: e.actor.name, render: (s) => <span className="font-medium text-fg">{s}</span> });
  if (e.task) {
    marks.push({
      find: e.task.title,
      render: (s) => (
        <Link
          to={`/app/tasks/${e.task!.id}`}
          state={{ background: location }}
          className="font-medium text-fg underline decoration-line-strong underline-offset-[3px] hover:decoration-fg-3"
        >
          {s}
        </Link>
      ),
    });
  }
  if (e.group) marks.push({ find: e.group.name, render: (s) => <span className="font-medium text-fg">{s}</span> });
  const text = emphasize(e.text, marks);
  const taskMentioned = e.task && e.text.includes(e.task.title);
  return (
    <li className={cn("relative flex gap-3 rounded-lg py-2.5 pr-2", fresh && "animate-rise")}>
      <span className="relative z-[1] flex size-8 shrink-0 items-center justify-center rounded-full bg-sheet">
        {e.actor ? (
          <Avatar user={e.actor} size="lg" />
        ) : (
          <span className="flex size-8 items-center justify-center rounded-full bg-surface shadow-card">
            <Icon className="size-4" style={{ color: kind.color }} />
          </span>
        )}
        {e.actor ? (
          <span className="absolute -right-1 -bottom-1 flex size-[18px] items-center justify-center rounded-full bg-surface shadow-card">
            <Icon className="size-[11px]" style={{ color: kind.color }} strokeWidth={2.4} />
          </span>
        ) : null}
      </span>
      <div className="min-w-0 flex-1 pt-[5px]">
        <p className="text-sm leading-5 text-fg-2">
          {text}
          {e.task && !taskMentioned ? (
            <>
              {" · "}
              <Link to={`/app/tasks/${e.task.id}`} state={{ background: location }} className="font-medium text-fg underline decoration-line-strong underline-offset-[3px]">
                {e.task.title}
              </Link>
            </>
          ) : null}
        </p>
        <p className="mt-0.5 flex items-center gap-2 text-xs text-fg-4">
          <time dateTime={e.at}>{formatTime(new Date(e.at))}</time>
          {e.group ? (
            <Link to={`/app/groups/${e.group.id}`} className="inline-flex items-center gap-1.5 hover:text-fg-2">
              <GroupDot color={e.group.color} className="size-[7px]" />
              {e.group.name}
            </Link>
          ) : null}
        </p>
      </div>
      {fresh ? <span className="mt-3 size-2 shrink-0 rounded-full bg-accent" aria-label="New" /> : null}
    </li>
  );
}
