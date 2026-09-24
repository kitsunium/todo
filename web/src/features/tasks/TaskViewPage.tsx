import { addDays, format } from "date-fns";
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
import { dueAt } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { InvitationBanner } from "../groups/InvitationBanner";
import { VIEWS, type ViewDef } from "../shell/nav";
import { Page } from "../shell/Page";
import { QuickAdd } from "./QuickAdd";
import { TaskList } from "./TaskList";

function Empty({ view }: { view: ViewDef["view"] }) {
  switch (view) {
    case "inbox":
      return (
        <EmptyState art={<InboxIllo />} title="Inbox zero">
          Nothing is waiting for a decision. Capture a task above — or press <Kbd>C</Kbd> anywhere.
        </EmptyState>
      );
    case "today":
      return (
        <EmptyState art={<TodayIllo />} title="Nothing due today">
          Enjoy the calm. Anything due today — or overdue — shows up here.
        </EmptyState>
      );
    case "upcoming":
      return (
        <EmptyState art={<UpcomingIllo />} title="Nothing planned">
          Give a task a date — type “fri” or “next week” — and it lands here.
        </EmptyState>
      );
    case "shared":
      return (
        <EmptyState
          art={<SharedIllo />}
          title="Nothing shared with you yet"
          action={
            <Button asChild variant="secondary" icon={<BookUser className="size-4" />}>
              <Link to="/app/contacts">Find your people</Link>
            </Button>
          }
        >
          When a contact shares a task with you, it shows up here.
        </EmptyState>
      );
    case "assigned":
      return (
        <EmptyState art={<AssignedIllo />} title="Nothing assigned to you">
          Tasks someone hands to you appear here — so do the ones you take on with “@you”.
        </EmptyState>
      );
    case "completed":
      return (
        <EmptyState art={<CompletedIllo />} title="Nothing completed yet">
          Check off a task and it moves here. Done tasks are archived a day later.
        </EmptyState>
      );
  }
}

export function TaskViewPage({ view }: { view: ViewDef["view"] }) {
  const def = VIEWS.find((v) => v.view === view)!;
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
  const subtitle =
    view === "today"
      ? format(now, "EEEE, MMMM d")
      : view === "completed"
        ? done !== undefined
          ? `${done} completed this week${done >= 5 ? " — nice work" : ""}`
          : " "
        : def.blurb;

  return (
    <Page
      title={def.label}
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
