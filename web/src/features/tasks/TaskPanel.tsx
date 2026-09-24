import { Dialog as D } from "radix-ui";
import { format } from "date-fns";
import {
  Archive,
  ArchiveRestore,
  CalendarDays,
  ChevronDown,
  ChevronUp,
  CircleCheck,
  CircleDashed,
  Ellipsis,
  Hash,
  Inbox,
  Link2,
  Plus,
  RotateCcw,
  Signal,
  Trash2,
  UserRound,
  UsersRound,
  X,
  CircleAlert,
} from "lucide-react";
import { useEffect, useRef, useState, type ComponentProps, type ReactNode } from "react";
import { ApiError } from "../../api/client";
import { errorMessage } from "../../api/errors";
import { useCurrentUser, useDeleteTask, useShareTask, useTask, useTaskAction, useUpdateTask } from "../../api/queries";
import type { Status, Task } from "../../api/types";
import { LostIllo } from "../../components/brand/illustrations";
import { Avatar } from "../../components/ui/avatar";
import { Button, IconButton } from "../../components/ui/button";
import { ConfirmDialog } from "../../components/ui/dialog";
import { GroupDot } from "../../components/ui/group-color";
import { AutoTextarea } from "../../components/ui/input";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuTrigger } from "../../components/ui/menu";
import { Popover, PopoverContent, PopoverTrigger } from "../../components/ui/popover";
import { PRIORITY_LABEL, PriorityIcon } from "../../components/ui/priority";
import { Skeleton } from "../../components/ui/skeleton";
import { toast } from "../../components/ui/toast";
import { Tooltip } from "../../components/ui/tooltip";
import { cn } from "../../lib/cn";
import { formatDue, formatDueLong, formatTime, timeAgo } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { useHotkeys } from "../shell/hotkeys";
import { useUI } from "../shell/store";
import { TaskCheckbox } from "./Checkbox";
import { DuePicker } from "./DuePicker";
import { AssigneeItems, GroupItems, PriorityItems, ShareList } from "./menus";

export function TaskPanel({ id, onClose, onNavigate }: { id: string; onClose: () => void; onNavigate: (id: string) => void }) {
  const [open, setOpen] = useState(true);
  const q = useTask(id);
  const order = useUI((s) => s.order);
  const i = order.indexOf(id);
  const prev = i > 0 ? order[i - 1] : undefined;
  const next = i >= 0 && i < order.length - 1 ? order[i + 1] : undefined;

  const closed = useRef(false);
  const finish = () => {
    if (closed.current) return;
    closed.current = true;
    onClose();
  };
  const close = () => {
    setOpen(false);
    // The exit animation ends the panel; the timer is for a browser without it.
    setTimeout(finish, 320);
  };
  useHotkeys(
    [
      { key: "j", allowInOverlay: true, run: () => next && onNavigate(next) },
      { key: "k", allowInOverlay: true, run: () => prev && onNavigate(prev) },
    ],
    open,
  );

  return (
    <D.Root open={open} onOpenChange={(o) => !o && close()}>
      <D.Portal>
        <D.Overlay className="fixed inset-0 z-[50] bg-[rgb(28_25_23/0.1)] data-[state=open]:animate-fade-in data-[state=closed]:animate-fade-out dark:bg-[rgb(0_0_0/0.35)]" />
        <D.Content
          aria-describedby={undefined}
          tabIndex={-1}
          onOpenAutoFocus={(e) => {
            e.preventDefault();
            (e.currentTarget as HTMLElement | null)?.focus({ preventScroll: true });
          }}
          onAnimationEnd={(e) => {
            if (!open && e.target === e.currentTarget) finish();
          }}
          className={cn(
            "fixed top-2 right-2 bottom-2 z-[51] flex w-[min(540px,calc(100vw-16px))] flex-col overflow-hidden rounded-2xl bg-surface shadow-panel outline-none",
            "data-[state=open]:animate-panel-in data-[state=closed]:animate-panel-out",
            "max-[640px]:top-0 max-[640px]:right-0 max-[640px]:bottom-0 max-[640px]:w-full max-[640px]:rounded-none",
          )}
        >
          <header className="flex h-12 shrink-0 items-center gap-1 border-b border-line-soft pr-2 pl-4">
            <Crumb task={q.data} />
            <div className="ml-auto flex items-center gap-0.5">
              <Tooltip content="Previous task" keys={["K"]}>
                <IconButton label="Previous task" size="sm" disabled={!prev} onClick={() => prev && onNavigate(prev)}>
                  <ChevronUp className="size-4" />
                </IconButton>
              </Tooltip>
              <Tooltip content="Next task" keys={["J"]}>
                <IconButton label="Next task" size="sm" disabled={!next} onClick={() => next && onNavigate(next)}>
                  <ChevronDown className="size-4" />
                </IconButton>
              </Tooltip>
              <span className="mx-1 h-4 w-px bg-line" aria-hidden="true" />
              {q.data ? <PanelMenu task={q.data} onDeleted={close} /> : null}
              <Tooltip content="Close" keys={["Esc"]}>
                <D.Close asChild>
                  <IconButton label="Close" size="sm">
                    <X className="size-4" />
                  </IconButton>
                </D.Close>
              </Tooltip>
            </div>
          </header>
          {q.data ? (
            <Body key={q.data.id} task={q.data} />
          ) : q.isError ? (
            <Gone error={q.error} onClose={close} />
          ) : (
            <PanelSkeleton />
          )}
        </D.Content>
      </D.Portal>
    </D.Root>
  );
}

function Crumb({ task }: { task: Task | undefined }) {
  const me = useCurrentUser();
  if (!task) return <D.Title className="text-sm text-fg-3">Task</D.Title>;
  return (
    <D.Title className="flex min-w-0 items-center gap-1.5 text-sm text-fg-3">
      {task.group ? (
        <>
          <GroupDot color={task.group.color} />
          <span className="truncate">{task.group.name}</span>
        </>
      ) : (
        <>
          {task.owner.id === me.id ? <Inbox className="size-3.5" aria-hidden="true" /> : <UsersRound className="size-3.5" aria-hidden="true" />}
          <span className="shrink-0">{task.owner.id === me.id ? "Inbox" : `${task.owner.name.split(" ")[0]}’s`}</span>
        </>
      )}
      <span className="text-fg-4" aria-hidden="true">/</span>
      <span className="truncate text-fg-2">{task.title}</span>
    </D.Title>
  );
}

function PanelMenu({ task, onDeleted }: { task: Task; onDeleted: () => void }) {
  const action = useTaskAction();
  const remove = useDeleteTask();
  const [confirm, setConfirm] = useState(false);
  return (
    <>
      <Menu>
        <Tooltip content="More">
          <MenuTrigger asChild>
            <IconButton label="More actions" size="sm">
              <Ellipsis className="size-4" />
            </IconButton>
          </MenuTrigger>
        </Tooltip>
        <MenuContent align="end" className="w-[210px]">
          <MenuItem
            icon={<Link2 className="size-4" />}
            onSelect={() =>
              void navigator.clipboard?.writeText(`${location.origin}/app/tasks/${task.id}`).then(
                () => toast.success("Link copied"),
                () => toast.error("Couldn’t copy the link"),
              )
            }
          >
            Copy link
          </MenuItem>
          {task.status === "done" ? (
            <MenuItem icon={<Archive className="size-4" />} onSelect={() => action.mutate({ id: task.id, action: "archive" })}>
              Archive
            </MenuItem>
          ) : null}
          {task.status === "archived" ? (
            <MenuItem icon={<ArchiveRestore className="size-4" />} onSelect={() => action.mutate({ id: task.id, action: "restore" })}>
              Restore
            </MenuItem>
          ) : null}
          {task.can.delete ? (
            <>
              <MenuSeparator />
              <MenuItem danger icon={<Trash2 className="size-4" />} onSelect={() => setConfirm(true)}>
                Delete…
              </MenuItem>
            </>
          ) : null}
        </MenuContent>
      </Menu>
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title="Delete this task?"
        description={`“${task.title}” will be deleted for everyone who can see it. This can’t be undone.`}
        confirm="Delete task"
        onConfirm={async () => {
          await remove.mutateAsync(task.id);
          toast("Task deleted");
          onDeleted();
        }}
      />
    </>
  );
}

const STATUS: Record<Status, { label: string; icon: ReactNode; cls: string }> = {
  open: { label: "Open", icon: <CircleDashed className="size-3.5" />, cls: "bg-inset text-fg-2" },
  overdue: { label: "Overdue", icon: <CircleAlert className="size-3.5" />, cls: "bg-danger-soft text-danger-ink" },
  done: { label: "Done", icon: <CircleCheck className="size-3.5" />, cls: "bg-success-soft text-success-ink" },
  archived: { label: "Archived", icon: <Archive className="size-3.5" />, cls: "bg-inset text-fg-3" },
};

function Body({ task }: { task: Task }) {
  const me = useCurrentUser();
  const now = useNow();
  const update = useUpdateTask();
  const action = useTaskAction();
  const share = useShareTask();
  const [title, setTitle] = useState(task.title);
  const [notes, setNotes] = useState(task.notes);
  const [savedAt, setSavedAt] = useState<number | null>(null);
  const editingTitle = useRef(false);
  const editingNotes = useRef(false);
  const notesTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const ro = !task.can.edit;
  const done = task.status === "done" || task.status === "archived";

  useEffect(() => {
    if (!editingTitle.current) setTitle(task.title);
  }, [task.title]);
  useEffect(() => {
    if (!editingNotes.current) setNotes(task.notes);
  }, [task.notes]);
  useEffect(() => () => clearTimeout(notesTimer.current), []);

  const saveTitle = () => {
    editingTitle.current = false;
    const t = title.replace(/\s+/g, " ").trim();
    if (!t) {
      setTitle(task.title);
      return;
    }
    if (t !== task.title) update.mutate({ id: task.id, patch: { title: t } });
  };
  const saveNotes = (value: string) => {
    clearTimeout(notesTimer.current);
    if (value !== task.notes) {
      update.mutate({ id: task.id, patch: { notes: value } }, { onSuccess: () => setSavedAt(Date.now()) });
    }
  };

  const due = task.due ? formatDue(task.due, now) : null;
  const [dueOpen, setDueOpen] = useState(false);

  return (
    <>
      <div className="min-h-0 flex-1 overflow-y-auto">
        <div className="grid grid-cols-[20px_minmax(0,1fr)] gap-x-3 px-5 pt-5 sm:px-6">
          <div className="pt-[4px]">
              <TaskCheckbox
                size="lg"
                priority={task.priority}
                checked={done}
                title={task.title}
                onToggle={() =>
                  action.mutate({ id: task.id, action: task.status === "archived" ? "restore" : done ? "reopen" : "complete" })
                }
              />
          </div>
            <AutoTextarea
              value={title}
              readOnly={ro}
              aria-label="Title"
              maxLength={200}
              onFocus={() => (editingTitle.current = true)}
              onChange={(e) => setTitle(e.target.value.replace(/\n/g, ""))}
              onBlur={saveTitle}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  e.currentTarget.blur();
                } else if (e.key === "Escape") {
                  e.stopPropagation();
                  setTitle(task.title);
                  editingTitle.current = false;
                  requestAnimationFrame(() => (e.target as HTMLTextAreaElement).blur());
                }
              }}
              className={cn(
                "-mx-1.5 block w-[calc(100%+12px)] min-w-0 rounded-md bg-transparent px-1.5 py-0.5 text-[20px] leading-7 font-semibold tracking-[-0.015em] outline-none",
                "hover:bg-hover focus:bg-inset",
                done ? "text-fg-3 line-through decoration-fg-4 decoration-1" : "text-fg",
              )}
            />
          <div className="relative col-start-2 mt-1">
            <AutoTextarea
              value={notes}
              readOnly={ro}
              minRows={2}
              maxLength={10000}
              aria-label="Notes"
              placeholder={ro ? "No notes" : "Add notes…"}
              onFocus={() => (editingNotes.current = true)}
              onChange={(e) => {
                const v = e.target.value;
                setNotes(v);
                clearTimeout(notesTimer.current);
                notesTimer.current = setTimeout(() => saveNotes(v), 800);
              }}
              onBlur={(e) => {
                editingNotes.current = false;
                saveNotes(e.target.value);
              }}
              className="-mx-1.5 block w-[calc(100%+12px)] rounded-md bg-transparent px-1.5 py-1 text-[14px] leading-[22px] text-fg-2 outline-none hover:bg-hover focus:bg-inset"
            />
            {savedAt ? (
              <span key={savedAt} className="absolute right-0 -bottom-5 text-2xs text-fg-4 animate-fade-in">
                Saved
              </span>
            ) : null}
          </div>
        </div>

        <div className="mx-5 mt-6 mb-3 h-px bg-line-soft sm:mx-6" />

        <div className="flex flex-col gap-0.5 px-3 pb-2 sm:px-4">
          <Row icon={<CircleDashed className="size-4" />} label="Status">
            <span className={cn("inline-flex h-6 items-center gap-1.5 rounded-md px-2 text-xs font-medium", STATUS[task.status].cls)}>
              {STATUS[task.status].icon}
              {STATUS[task.status].label}
            </span>
            {task.status === "done" ? (
              <button type="button" className="ml-2 text-xs text-fg-3 hover:text-fg" onClick={() => action.mutate({ id: task.id, action: "archive" })}>
                Archive
              </button>
            ) : null}
            {task.status === "archived" ? (
              <button type="button" className="ml-2 text-xs text-fg-3 hover:text-fg" onClick={() => action.mutate({ id: task.id, action: "restore" })}>
                Restore
              </button>
            ) : null}
          </Row>

          <Row icon={<Signal className="size-4" />} label="Priority">
            <Menu>
              <MenuTrigger asChild disabled={ro}>
                <Value>
                  <PriorityIcon priority={task.priority} />
                  <span className={task.priority ? "text-fg" : "text-fg-3"}>{PRIORITY_LABEL[task.priority]}</span>
                </Value>
              </MenuTrigger>
              <MenuContent>
                <PriorityItems value={task.priority} onPick={(priority) => update.mutate({ id: task.id, patch: { priority } })} />
              </MenuContent>
            </Menu>
          </Row>

          <Row icon={<CalendarDays className="size-4" />} label="Due date">
            <Popover open={dueOpen} onOpenChange={setDueOpen}>
              <PopoverTrigger asChild disabled={ro}>
                <Value>
                  {due ? (
                    <span className={cn(!done && due.overdue ? "font-medium text-danger-ink" : due.soon ? "text-accent-ink" : "text-fg")}>
                      {formatDueLong(task.due!)}
                    </span>
                  ) : (
                    <span className="text-fg-3">No date</span>
                  )}
                </Value>
              </PopoverTrigger>
              <PopoverContent className="p-0">
                <DuePicker value={task.due} onChange={(d) => update.mutate({ id: task.id, patch: { due: d } })} onPicked={() => setDueOpen(false)} />
              </PopoverContent>
            </Popover>
          </Row>

          <Row icon={<Hash className="size-4" />} label="Group">
            <Menu>
              <MenuTrigger asChild disabled={ro || task.owner.id !== me.id}>
                <Value title={task.owner.id !== me.id ? "Only the owner can move this task" : undefined}>
                  {task.group ? (
                    <>
                      <GroupDot color={task.group.color} />
                      <span className="text-fg">{task.group.name}</span>
                    </>
                  ) : (
                    <span className="text-fg-3">No group</span>
                  )}
                </Value>
              </MenuTrigger>
              <MenuContent>
                <GroupItems value={task.group?.id} onPick={(groupId) => update.mutate({ id: task.id, patch: { groupId } })} />
              </MenuContent>
            </Menu>
          </Row>

          <Row icon={<UserRound className="size-4" />} label="Assignee">
            <Menu>
              <MenuTrigger asChild disabled={ro}>
                <Value>
                  {task.assignee ? (
                    <>
                      <Avatar user={task.assignee} size="sm" />
                      <span className="text-fg">{task.assignee.id === me.id ? `${task.assignee.name} (you)` : task.assignee.name}</span>
                    </>
                  ) : (
                    <span className="text-fg-3">Unassigned</span>
                  )}
                </Value>
              </MenuTrigger>
              <MenuContent className="w-[260px]">
                <AssigneeItems task={task} />
              </MenuContent>
            </Menu>
          </Row>

          <Row icon={<UsersRound className="size-4" />} label="Shared with" top>
            <div className="flex min-w-0 flex-wrap items-center gap-1.5 py-1">
              {task.sharedWith.map((u) => (
                <span key={u.id} className="group/pill inline-flex h-7 items-center gap-1.5 rounded-full bg-inset py-0.5 pr-1 pl-0.5 text-sm text-fg">
                  <Avatar user={u} size="md" />
                  <span className="max-w-[140px] truncate">{u.id === me.id ? "You" : u.name}</span>
                  {task.can.share || u.id === me.id ? (
                    <button
                      type="button"
                      aria-label={u.id === me.id ? "Leave this task" : `Stop sharing with ${u.name}`}
                      onClick={() => share.mutate({ id: task.id, userId: u.id, remove: true })}
                      className="flex size-5 items-center justify-center rounded-full text-fg-4 hover:bg-active hover:text-fg"
                    >
                      <X className="size-3" />
                    </button>
                  ) : (
                    <span className="w-1" />
                  )}
                </span>
              ))}
              {task.can.share ? (
                <Popover>
                  <PopoverTrigger asChild>
                    <button
                      type="button"
                      className="inline-flex h-7 items-center gap-1 rounded-full px-2.5 text-sm text-fg-3 shadow-[inset_0_0_0_1px_var(--line)] hover:bg-hover hover:text-fg data-[state=open]:bg-hover"
                    >
                      <Plus className="size-3.5" /> {task.sharedWith.length ? "Add" : "Share"}
                    </button>
                  </PopoverTrigger>
                  <PopoverContent className="p-0">
                    <ShareList task={task} />
                  </PopoverContent>
                </Popover>
              ) : task.sharedWith.length === 0 ? (
                <span className="text-sm text-fg-3">Only people in its group</span>
              ) : null}
            </div>
          </Row>
        </div>

        <div className="mx-5 mt-2 mb-4 h-px bg-line-soft sm:mx-6" />
        <div className="flex flex-col gap-1.5 px-5 pb-6 text-xs text-fg-3 sm:px-6">
          <p className="flex items-center gap-2">
            <Avatar user={task.owner} size="xs" />
            <span>
              Created by <span className="text-fg-2">{task.owner.id === me.id ? "you" : task.owner.name}</span> ·{" "}
              {format(new Date(task.createdAt), "MMM d, yyyy")} at {formatTime(new Date(task.createdAt))}
            </span>
          </p>
          {task.completedAt && task.completedBy ? (
            <p className="flex items-center gap-2">
              <CircleCheck className="size-[18px] text-done" />
              <span>
                Completed by <span className="text-fg-2">{task.completedBy.id === me.id ? "you" : task.completedBy.name}</span> ·{" "}
                {timeAgo(task.completedAt, now)}
              </span>
            </p>
          ) : null}
          <p className="pl-[26px] text-fg-4">Updated {timeAgo(task.updatedAt, now)}</p>
        </div>
      </div>
      <footer className="flex shrink-0 items-center gap-2 border-t border-line-soft bg-sheet/60 px-4 py-3">
        <span className="text-xs text-fg-4 max-sm:hidden">
          {ro ? "You can view this task." : "Changes save as you go."}
        </span>
        <div className="ml-auto flex items-center gap-2">
          {task.status === "archived" ? (
            <Button variant="secondary" icon={<ArchiveRestore className="size-4" />} onClick={() => action.mutate({ id: task.id, action: "restore" })}>
              Restore
            </Button>
          ) : task.status === "done" ? (
            <Button variant="secondary" icon={<RotateCcw className="size-4" />} onClick={() => action.mutate({ id: task.id, action: "reopen" })}>
              Mark as not done
            </Button>
          ) : (
            <Button variant="primary" icon={<CircleCheck className="size-4" />} onClick={() => action.mutate({ id: task.id, action: "complete" })}>
              Complete
            </Button>
          )}
        </div>
      </footer>
    </>
  );
}

function Row({ icon, label, children, top }: { icon: ReactNode; label: string; children: ReactNode; top?: boolean }) {
  return (
    <div className={cn("grid min-h-9 grid-cols-[124px_minmax(0,1fr)] gap-2", top ? "items-start" : "items-center")}>
      <span className={cn("flex items-center gap-2 px-2 text-sm text-fg-3", top && "pt-2")}>
        <span className="text-fg-4">{icon}</span>
        {label}
      </span>
      <div className="flex min-w-0 items-center">{children}</div>
    </div>
  );
}

function Value({ children, ...rest }: { children: ReactNode } & ComponentProps<"button">) {
  return (
    <button
      type="button"
      {...rest}
      className="-ml-0.5 inline-flex h-8 max-w-full min-w-0 items-center gap-2 truncate rounded-md px-2 text-sm outline-none transition-colors hover:bg-hover focus-visible:ring-2 focus-visible:ring-ring disabled:hover:bg-transparent data-[state=open]:bg-hover"
    >
      {children}
    </button>
  );
}

function PanelSkeleton() {
  return (
    <div className="flex-1 px-6 pt-6" aria-busy="true" aria-label="Loading the task">
      <div className="flex items-center gap-3">
        <Skeleton className="size-5 rounded-full" />
        <Skeleton className="h-5 w-2/3" />
      </div>
      <Skeleton className="mt-4 ml-8 h-3 w-1/2" />
      <Skeleton className="mt-2 ml-8 h-3 w-1/3" />
      <div className="mt-8 flex flex-col gap-4">
        {["w-40", "w-28", "w-36", "w-32", "w-44"].map((w) => (
          <div key={w} className="flex items-center gap-8">
            <Skeleton className="h-3 w-20" />
            <Skeleton className={cn("h-3", w)} />
          </div>
        ))}
      </div>
    </div>
  );
}

function Gone({ error, onClose }: { error: unknown; onClose: () => void }) {
  const missing = error instanceof ApiError && (error.status === 404 || error.status === 403);
  return (
    <div className="flex flex-1 flex-col items-center justify-center px-8 text-center">
      <LostIllo />
      <h3 className="mt-3 text-md font-semibold text-fg">{missing ? "This task isn’t here" : "Couldn’t open this task"}</h3>
      <p className="mt-1 max-w-[300px] text-sm text-fg-3">
        {missing ? "It was deleted, or it’s no longer shared with you." : errorMessage(error)}
      </p>
      <Button className="mt-5" variant="secondary" onClick={onClose}>
        Close
      </Button>
    </div>
  );
}
