import { AnimatePresence, motion } from "motion/react";
import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { useLocation, useNavigate } from "react-router";
import { useCurrentUser, useTaskAction, useUpdateTask } from "../../api/queries";
import type { Priority, Task, TaskView } from "../../api/types";
import { toast } from "../../components/ui/toast";
import { cn } from "../../lib/cn";
import { dueAt, groupByCompletion, groupByDue, isDateOnly, type Section, type SectionKey } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { useHotkeys } from "../shell/hotkeys";
import { setUI } from "../shell/store";
import { TaskRow } from "./TaskRow";
import { viewAccepts } from "./view";

const LINGER_MS = 900;

/**
 * A task list in sections, with the completion choreography (check, strike,
 * collapse, Undo) and the list shortcuts: J/K to move, X to complete, E or
 * Enter to open, 1–4 and 0 for the priority, Esc to let go.
 */
export function TaskList({
  tasks,
  view,
  mode = "due",
  showGroup = true,
  empty,
}: {
  tasks: Task[];
  view?: TaskView;
  mode?: "due" | "completion";
  showGroup?: boolean;
  empty: ReactNode;
}) {
  const me = useCurrentUser();
  const now = useNow();
  const navigate = useNavigate();
  const location = useLocation();
  const action = useTaskAction();
  const update = useUpdateTask();
  // A task just checked (or unchecked) keeps its place for a moment, as it
  // now is: the refetch that follows the mutation may already have dropped it.
  const [held, setHeld] = useState<Record<string, { key: SectionKey; task: Task }>>({});
  const [selected, setSelected] = useState<string | null>(null);
  const timers = useRef(new Map<string, ReturnType<typeof setTimeout>>());

  useEffect(() => {
    const all = timers.current;
    return () => {
      for (const t of all.values()) clearTimeout(t);
    };
  }, []);

  const sections: Section[] = useMemo(() => {
    const shown = tasks.filter((t) => viewAccepts(view, t) && !held[t.id]);
    for (const h of Object.values(held)) shown.push(h.task);
    const pins: Record<string, SectionKey> = {};
    const pinnedAt: Record<string, string> = {};
    for (const [id, h] of Object.entries(held)) {
      pins[id] = h.key;
      if (h.task.completedAt) pinnedAt[id] = h.task.completedAt;
    }
    return mode === "completion" ? groupByCompletion(shown, now, pinnedAt) : groupByDue(shown, now, pins);
  }, [tasks, view, mode, now, held]);

  const order = useMemo(() => sections.flatMap((s) => s.tasks.map((t) => t.id)), [sections]);
  useEffect(() => setUI({ order }), [order]);
  useEffect(() => () => setUI({ order: [] }), []);

  const hold = useCallback((task: Task, key: SectionKey) => {
    setHeld((h) => ({ ...h, [task.id]: { key, task } }));
    clearTimeout(timers.current.get(task.id));
    timers.current.set(
      task.id,
      setTimeout(() => {
        setHeld(({ [task.id]: _, ...rest }) => rest);
        timers.current.delete(task.id);
      }, LINGER_MS),
    );
  }, []);

  const toggle = useCallback(
    (t: Task) => {
      const section = sections.find((s) => s.tasks.some((x) => x.id === t.id));
      const now = new Date().toISOString();
      const after: Task =
        t.status === "open" || t.status === "overdue"
          ? { ...t, status: "done", completedAt: now, completedBy: { id: me.id, name: me.name, email: me.email } }
          : { ...t, status: "open" };
      if (section) hold(after, section.key);
      if (t.status === "archived") {
        action.mutate({ id: t.id, action: "restore" });
        toast(`Restored “${t.title}”`);
        return;
      }
      if (t.status === "done") {
        action.mutate({ id: t.id, action: "reopen" });
        return;
      }
      action.mutate({ id: t.id, action: "complete" });
      toast(`Completed “${t.title}”`, {
        id: `done-${t.id}`,
        action: { label: "Undo", onClick: () => action.mutate({ id: t.id, action: "reopen" }) },
      });
    },
    [sections, hold, action, me],
  );

  const move = (delta: number) => {
    if (!order.length) return;
    const i = selected ? order.indexOf(selected) : -1;
    const next = order[Math.max(0, Math.min(order.length - 1, i < 0 ? (delta > 0 ? 0 : order.length - 1) : i + delta))]!;
    setSelected(next);
    document.querySelector(`[data-task-id="${next}"]`)?.scrollIntoView({ block: "nearest" });
  };
  const current = () => tasks.find((t) => t.id === selected);
  const open = () => {
    if (selected) navigate(`/app/tasks/${selected}`, { state: { background: location } });
  };
  const setPriority = (p: Priority) => {
    const t = current();
    if (t && t.priority !== p) update.mutate({ id: t.id, patch: { priority: p } });
  };

  useHotkeys([
    { key: "j", run: () => move(1) },
    { key: "ArrowDown", run: () => move(1) },
    { key: "k", run: () => move(-1) },
    { key: "ArrowUp", run: () => move(-1) },
    { key: "x", run: () => current() && toggle(current()!) },
    { key: "e", run: open },
    { key: "Enter", run: open },
    { key: "Escape", run: () => setSelected(null) },
    { key: "1", run: () => setPriority(1) },
    { key: "2", run: () => setPriority(2) },
    { key: "3", run: () => setPriority(3) },
    { key: "4", run: () => setPriority(4) },
    { key: "0", run: () => setPriority(0) },
  ]);

  const rescheduleOverdue = (list: Task[]) => {
    for (const t of list) {
      const d = t.due ? new Date(t.due) : null;
      const time = d && !isDateOnly(d) ? { h: d.getHours(), m: d.getMinutes() } : undefined;
      let next = dueAt(now, time);
      if (next < now) next = dueAt(now);
      update.mutate({ id: t.id, patch: { due: next.toISOString() } });
    }
    toast.success(`Moved ${list.length} task${list.length === 1 ? "" : "s"} to today`);
  };

  if (sections.length === 0) return <>{empty}</>;

  return (
    <div role="list" aria-label="Tasks" className="pb-6">
      <AnimatePresence initial={false}>
        {sections.map((s) => (
          <motion.section
            key={s.key}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.18 }}
            aria-label={s.label}
            className="mb-3"
          >
            <header className="sticky top-0 z-10 -mx-1 flex h-9 items-center gap-2 bg-sheet/90 px-3.5 backdrop-blur-md supports-[backdrop-filter]:bg-sheet/75">
              <h2 className={cn("text-[13px] font-semibold tracking-[-0.005em]", s.tone === "danger" ? "text-danger-ink" : "text-fg")}>{s.label}</h2>
              <span className="tabular text-xs text-fg-4">{s.tasks.length}</span>
              {s.key === "overdue" && s.tasks.some((t) => t.can.edit) ? (
                <button
                  type="button"
                  onClick={() => rescheduleOverdue(s.tasks.filter((t) => t.can.edit && t.status !== "done"))}
                  className="ml-auto rounded-md px-2 py-1 text-xs font-medium text-fg-3 transition-colors hover:bg-hover hover:text-fg"
                >
                  Move to today
                </button>
              ) : null}
            </header>
            <AnimatePresence initial={false}>
              {s.tasks.map((t) => (
                <motion.div
                  key={t.id}
                  role="listitem"
                  layout="position"
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  exit={{ opacity: 0, height: 0, transition: { duration: 0.22, ease: [0.4, 0, 0.2, 1] } }}
                  transition={{ duration: 0.24, ease: [0.22, 1, 0.36, 1] }}
                  className="overflow-hidden"
                >
                  <TaskRow task={t} meId={me.id} selected={selected === t.id} showGroup={showGroup} onToggle={toggle} onSelect={setSelected} />
                </motion.div>
              ))}
            </AnimatePresence>
          </motion.section>
        ))}
      </AnimatePresence>
    </div>
  );
}
