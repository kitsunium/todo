import {
  Archive,

  CalendarDays,
  CircleCheck,
  Ellipsis,
  FolderInput,
  Link2,
  PanelRightOpen,
  RotateCcw,
  Signal,
  Trash2,
  UserRoundCheck,
  UserRoundPlus,
  AlignLeft,
} from "lucide-react";
import { memo, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { useCurrentUser, useDeleteTask, useTaskAction, useUpdateTask } from "../../api/queries";
import type { Task } from "../../api/types";
import { AvatarStack } from "../../components/ui/avatar";
import { IconButton } from "../../components/ui/button";
import { ConfirmDialog } from "../../components/ui/dialog";
import { GroupDot } from "../../components/ui/group-color";
import { Menu, MenuContent, MenuItem, MenuSeparator, MenuSub, MenuSubContent, MenuSubTrigger, MenuTrigger } from "../../components/ui/menu";
import { Popover, PopoverContent, PopoverTrigger } from "../../components/ui/popover";
import { PriorityIcon } from "../../components/ui/priority";
import { toast } from "../../components/ui/toast";
import { Tooltip } from "../../components/ui/tooltip";
import { tr, useT } from "../../i18n";
import { cn } from "../../lib/cn";
import { formatDue } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { AssigneeItems, GroupItems, PriorityItems, ShareList } from "./menus";
import { TaskCheckbox } from "./Checkbox";
import { DuePicker } from "./DuePicker";

export type RowProps = {
  task: Task;
  meId: string;
  selected: boolean;
  showGroup: boolean;
  onToggle: (t: Task) => void;
  onSelect: (id: string) => void;
};

/** Everyone on a task but me: assignee first, then the owner, then the people it is shared with. */
export function peopleOn(t: Task, meId: string) {
  const out = new Map<string, Task["owner"]>();
  if (t.assignee) out.set(t.assignee.id, t.assignee);
  if (t.owner.id !== meId) out.set(t.owner.id, t.owner);
  for (const u of t.sharedWith) out.set(u.id, u);
  out.delete(meId);
  return [...out.values()];
}

export const TaskRow = memo(function TaskRow({ task, meId, selected, showGroup, onToggle, onSelect }: RowProps) {
  const t = useT();
  const now = useNow();
  const location = useLocation();
  const done = task.status === "done" || task.status === "archived";
  const due = task.due ? formatDue(task.due, now, t.locale) : null;
  const overdue = !done && !!due?.overdue;
  const people = peopleOn(task, meId);
  const assignedToMe = task.assignee?.id === meId && task.owner.id !== meId;
  const meta = !!due || (showGroup && !!task.group) || assignedToMe || task.status === "archived" || people.length > 0;

  return (
    <div
      data-task-id={task.id}
      data-selected={selected || undefined}
      className={cn(
        "group/row relative flex scroll-mt-12 items-start gap-3 rounded-[10px] px-2.5 py-2 transition-colors duration-100",
        "hover:bg-hover data-[selected]:bg-selected",
        "before:absolute before:top-2.5 before:bottom-2.5 before:left-0 before:w-[2px] before:rounded-full before:bg-accent before:opacity-0 data-[selected]:before:opacity-100",
      )}
    >
      <div className="pt-px">
        <TaskCheckbox priority={task.priority} checked={done} title={task.title} onToggle={() => onToggle(task)} disabled={task.id.startsWith("tmp_")} />
      </div>
      <Link
        to={`/app/tasks/${task.id}`}
        state={{ background: location }}
        onFocus={() => onSelect(task.id)}
        className="min-w-0 flex-1 rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        draggable={false}
      >
        <span className="flex min-w-0 items-center gap-1.5">
          <span className="relative inline-block max-w-full min-w-0 align-top">
            <span className={cn("block truncate text-[14px] leading-5 transition-colors duration-300", done ? "text-fg-3" : "text-fg")}>
              {task.title}
            </span>
            <span
              aria-hidden="true"
              className={cn(
                "pointer-events-none absolute top-[10.5px] left-0 h-px bg-fg-3 transition-[width] duration-300 ease-out",
                done ? "w-full" : "w-0",
              )}
            />
          </span>
          {task.notes ? (
            <span className="inline-flex shrink-0 text-fg-4" title={t("row.hasNotes")}>
              <AlignLeft className="size-3.5" aria-hidden="true" />
              <span className="sr-only">{t("row.hasNotes")}</span>
            </span>
          ) : null}
        </span>
        {meta ? (
          <span className="mt-[3px] flex min-w-0 items-center gap-3 text-xs leading-4 text-fg-3">
            {due ? (
              <span
                className={cn(
                  "inline-flex shrink-0 items-center gap-1",
                  done ? "" : overdue ? "font-medium text-danger-ink" : due.soon ? "text-accent-ink" : "",
                )}
              >
                <CalendarDays className="size-3" aria-hidden="true" />
                {due.label}
              </span>
            ) : null}
            {showGroup && task.group ? (
              <span className="inline-flex min-w-0 items-center gap-1.5">
                <GroupDot color={task.group.color} className="size-[7px]" />
                <span className="truncate">{task.group.name}</span>
              </span>
            ) : null}
            {assignedToMe ? (
              <span className="inline-flex shrink-0 items-center gap-1">
                <UserRoundCheck className="size-3" aria-hidden="true" />
                {t("row.forYou")}
              </span>
            ) : null}
            {task.status === "archived" ? (
              <span className="inline-flex shrink-0 items-center gap-1">
                <Archive className="size-3" aria-hidden="true" />
                {t("status.archived")}
              </span>
            ) : null}
            {people.length ? (
              <span className="inline-flex min-w-0 items-center gap-1.5">
                <AvatarStack users={people} size="xs" max={4} />
                {people.length === 1 ? <span className="truncate">{people[0]!.name.split(" ")[0]}</span> : null}
              </span>
            ) : null}
          </span>
        ) : null}
      </Link>
      <div className="relative flex w-0 shrink-0 items-center self-center">
        <QuickActions task={task} />
      </div>
    </div>
  );
});

function QuickActions({ task }: { task: Task }) {
  const t = useT();
  const me = useCurrentUser();
  const [dueOpen, setDueOpen] = useState(false);
  const navigate = useNavigate();
  const here = useLocation();
  const update = useUpdateTask();
  const action = useTaskAction();
  const remove = useDeleteTask();
  const [confirm, setConfirm] = useState(false);
  const done = task.status === "done";
  const hidden =
    "opacity-0 group-hover/row:opacity-100 group-focus-within/row:opacity-100 group-has-[[aria-expanded=true]]/row:opacity-100 max-[899px]:hidden";
  return (
    <div
      className={cn(
        "absolute top-1/2 right-0 flex -translate-y-1/2 items-center gap-0.5 rounded-r-md py-0.5 pl-8 transition-opacity duration-100",
        "bg-[linear-gradient(to_right,transparent,var(--row-hover-solid)_28px)]",
        hidden,
      )}
    >
      <Popover open={dueOpen} onOpenChange={setDueOpen}>
        <Tooltip content={t("row.dueDate")}>
          <PopoverTrigger asChild>
            <IconButton label={t("row.setDue")} size="xs">
              <CalendarDays className="size-3.5" />
            </IconButton>
          </PopoverTrigger>
        </Tooltip>
        <PopoverContent align="end" className="p-0">
          <DuePicker value={task.due} onChange={(due) => update.mutate({ id: task.id, patch: { due } })} onPicked={() => setDueOpen(false)} />
        </PopoverContent>
      </Popover>
      <Menu>
        <Tooltip content={t("panel.priority")}>
          <MenuTrigger asChild>
            <IconButton label={t("row.setPriority")} size="xs">
              <PriorityIcon priority={task.priority} className="size-3.5" />
            </IconButton>
          </MenuTrigger>
        </Tooltip>
        <MenuContent align="end">
          <PriorityItems value={task.priority} onPick={(priority) => update.mutate({ id: task.id, patch: { priority } })} />
        </MenuContent>
      </Menu>
      {task.can.share ? (
        <Popover>
          <Tooltip content={t("row.share")}>
            <PopoverTrigger asChild>
              <IconButton label={t("row.share")} size="xs">
                <UserRoundPlus className="size-3.5" />
              </IconButton>
            </PopoverTrigger>
          </Tooltip>
          <PopoverContent align="end" className="p-0">
            <ShareList task={task} />
          </PopoverContent>
        </Popover>
      ) : null}
      <Menu>
        <Tooltip content={t("common.more")}>
          <MenuTrigger asChild>
            <IconButton label={t("common.moreActions")} size="xs">
              <Ellipsis className="size-4" />
            </IconButton>
          </MenuTrigger>
        </Tooltip>
        <MenuContent align="end" className="w-[220px]">
          <MenuItem
            icon={<PanelRightOpen className="size-4" />}
            shortcut="E"
            onSelect={() => navigate(`/app/tasks/${task.id}`, { state: { background: here } })}
          >
            {t("task.open")}
          </MenuItem>
          <MenuItem
            icon={done ? <RotateCcw className="size-4" /> : <CircleCheck className="size-4" />}
            shortcut="X"
            onSelect={() =>
              action.mutate({ id: task.id, action: task.status === "archived" ? "restore" : done ? "reopen" : "complete" })
            }
          >
            {task.status === "archived" ? t("task.restore") : done ? t("task.reopen") : t("task.complete")}
          </MenuItem>
          <MenuSeparator />
          <MenuSub>
            <MenuSubTrigger icon={<Signal className="size-4" />}>{t("panel.priority")}</MenuSubTrigger>
            <MenuSubContent>
              <PriorityItems value={task.priority} onPick={(priority) => update.mutate({ id: task.id, patch: { priority } })} />
            </MenuSubContent>
          </MenuSub>
          {task.owner.id === me.id ? (
            <MenuSub>
              <MenuSubTrigger icon={<FolderInput className="size-4" />}>{t("task.moveToGroup")}</MenuSubTrigger>
              <MenuSubContent>
                <GroupItems value={task.group?.id} onPick={(groupId) => update.mutate({ id: task.id, patch: { groupId } })} />
              </MenuSubContent>
            </MenuSub>
          ) : null}
          <MenuSub>
            <MenuSubTrigger icon={<UserRoundCheck className="size-4" />}>{t("task.assign")}</MenuSubTrigger>
            <MenuSubContent className="w-[240px]">
              <AssigneeItems task={task} />
            </MenuSubContent>
          </MenuSub>
          <MenuSeparator />
          <MenuItem
            icon={<Link2 className="size-4" />}
            onSelect={() => {
              void navigator.clipboard?.writeText(`${location.origin}/app/tasks/${task.id}`).then(
                () => toast.success(tr()("common.linkCopied")),
                () => toast.error(tr()("common.linkCopyFailed")),
              );
            }}
          >
            {t("common.copyLink")}
          </MenuItem>
          {task.status === "done" ? (
            <MenuItem icon={<Archive className="size-4" />} onSelect={() => action.mutate({ id: task.id, action: "archive" })}>
              {t("task.archive")}
            </MenuItem>
          ) : null}
          {task.can.delete ? (
            <MenuItem danger icon={<Trash2 className="size-4" />} onSelect={() => setConfirm(true)}>
              {t("task.delete")}
            </MenuItem>
          ) : null}
        </MenuContent>
      </Menu>
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={t("task.deleteTitle")}
        description={t("task.deleteBody", { title: task.title })}
        confirm={t("task.deleteConfirm")}
        onConfirm={async () => {
          await remove.mutateAsync(task.id);
          toast(tr()("task.deleted"));
        }}
      />
    </div>
  );
}
