import { addDays } from "date-fns";
import { BookUser } from "lucide-react";
import { Link } from "react-router";
import { useCounts, useCurrentUser, useTasks } from "../../api/queries";
import type { NewTask } from "../../api/types";
import { errorMessage } from "../../api/errors";
import {
  AssignedIllo,
  CompletedIllo,
  InboxIllo,
  SharedIllo,
  TodayIllo,
  UpcomingIllo,
} from "../../components/brand/illustrations";
import { Button } from "../../components/ui/button";
import { EmptyState, ErrorState } from "../../components/ui/empty";
import { Kbd } from "../../components/ui/kbd";
import { TaskListSkeleton } from "../../components/ui/skeleton";
import { useT } from "../../i18n";
import { dueAt, formatDayLong } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { InvitationBanner } from "../groups/InvitationBanner";
import { VIEWS, type ViewDef } from "../shell/nav";
import { Page } from "../shell/Page";
import { QuickAdd } from "./QuickAdd";
import { TaskList } from "./TaskList";

function Empty({ view }: { view: ViewDef["view"] }) {
  const t = useT();
  const me = useCurrentUser();
  switch (view) {
    case "inbox":
      return (
        <EmptyState art={<InboxIllo />} title={t("empty.inbox.title")}>
          {t.rich("empty.inbox.body", { key: <Kbd>C</Kbd> })}
        </EmptyState>
      );
    case "today":
      return (
        <EmptyState art={<TodayIllo />} title={t("empty.today.title")}>
          {t("empty.today.body")}
        </EmptyState>
      );
    case "upcoming":
      return (
        <EmptyState art={<UpcomingIllo />} title={t("empty.upcoming.title")}>
          {t("empty.upcoming.body")}
        </EmptyState>
      );
    case "shared":
      return (
        <EmptyState
          art={<SharedIllo />}
          title={t("empty.shared.title")}
          action={
            <Button asChild variant="secondary" icon={<BookUser className="size-4" />}>
              <Link to="/app/contacts">{t("empty.shared.action")}</Link>
            </Button>
          }
        >
          {t("empty.shared.body")}
        </EmptyState>
      );
    case "assigned":
      return (
        <EmptyState art={<AssignedIllo />} title={t("empty.assigned.title")}>
          {t.rich("empty.assigned.body", {
            mention: <span className="font-medium text-fg-2">@{mentionName(me.name, me.email)}</span>,
          })}
        </EmptyState>
      );
    case "completed":
      return (
        <EmptyState art={<CompletedIllo />} title={t("empty.completed.title")}>
          {t("empty.completed.body")}
        </EmptyState>
      );
  }
}

/** The @mention that names me in the quick add: my first name. */
function mentionName(name: string, email: string): string {
  return (name.trim().split(/\s+/)[0] || email.split("@")[0] || "").toLowerCase();
}

export function TaskViewPage({ view }: { view: ViewDef["view"] }) {
  const def = VIEWS.find((v) => v.view === view)!;
  const t = useT();
  const me = useCurrentUser();
  const now = useNow();
  const q = useTasks(view);
  const counts = useCounts();
  const Icon = def.icon;

  const defaults: Partial<NewTask> =
    view === "today"
      ? { due: dueAt(now).toISOString() }
      : view === "upcoming"
        ? { due: dueAt(addDays(now, 1)).toISOString() }
        : view === "assigned"
          ? { assigneeId: me.id }
          : {};

  const done = counts.data?.completedThisWeek;
  const week = done !== undefined ? t("view.completed.week", { count: done }) : "";
  const subtitle =
    view === "today"
      ? formatDayLong(now, now, t.locale)
      : view === "completed"
        ? done !== undefined
          ? done >= 5
            ? t("view.completed.nice", { text: week })
            : week
          : " "
        : def.blurb
          ? t(def.blurb)
          : undefined;

  return (
    <Page
      title={t(def.label)}
      icon={<Icon style={{ color: def.color }} fill={view === "today" ? def.color : "none"} strokeWidth={2} />}
      subtitle={subtitle}
    >
      {view === "inbox" || view === "today" ? <InvitationBanner /> : null}
      {view !== "completed" && view !== "shared" ? <QuickAdd view={view} defaults={defaults} /> : null}
      {q.isPending ? (
        <TaskListSkeleton />
      ) : q.isError ? (
        <ErrorState message={errorMessage(q.error)} onRetry={() => void q.refetch()} />
      ) : (
        <TaskList
          tasks={q.data}
          view={view}
          mode={view === "completed" ? "completion" : "due"}
          empty={<Empty view={view} />}
        />
      )}
    </Page>
  );
}
