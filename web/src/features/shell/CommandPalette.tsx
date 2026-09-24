import { Command } from "cmdk";
import { Dialog as D } from "radix-ui";
import {
  Activity as ActivityIcon,
  BookUser,
  CornerDownLeft,
  Keyboard,
  Moon,
  Plus,
  Search,
  Settings,
  Sun,
  UserRoundPlus,
} from "lucide-react";
import { useState, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { useCounts, useGroups, useTasks } from "../../api/queries";
import type { Task } from "../../api/types";
import { GroupDot } from "../../components/ui/group-color";
import { Kbd } from "../../components/ui/kbd";
import { cn } from "../../lib/cn";
import { formatDue } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { resolvedTheme, toggleTheme, useTheme } from "../../lib/theme";
import { TaskCheckbox } from "../tasks/Checkbox";
import { VIEWS } from "./nav";
import { newTask, setUI, useUI } from "./store";

const ITEM =
  "flex h-10 cursor-default items-center gap-3 rounded-lg px-3 text-sm text-fg-2 outline-none select-none " +
  "data-[selected=true]:bg-hover data-[selected=true]:text-fg aria-disabled:opacity-40";

function Item({ icon, children, hint, onSelect, value, keywords }: { icon: ReactNode; children: ReactNode; hint?: ReactNode; onSelect: () => void; value: string; keywords?: string[] }) {
  return (
    <Command.Item value={value} {...(keywords ? { keywords } : {})} onSelect={onSelect} className={ITEM}>
      <span className="flex size-4 shrink-0 items-center justify-center text-fg-3">{icon}</span>
      <span className="min-w-0 flex-1 truncate">{children}</span>
      {hint ? <span className="flex shrink-0 items-center gap-1 text-xs text-fg-4">{hint}</span> : null}
    </Command.Item>
  );
}

function TaskItem({ task, onSelect }: { task: Task; onSelect: () => void }) {
  const now = useNow();
  const due = task.due ? formatDue(task.due, now) : null;
  return (
    <Command.Item value={`task ${task.title} ${task.id}`} keywords={[task.group?.name ?? "", task.notes.slice(0, 200)]} onSelect={onSelect} className={ITEM}>
      <span className="pointer-events-none flex shrink-0 items-center" aria-hidden="true">
        <TaskCheckbox priority={task.priority} checked={task.status === "done"} title={task.title} onToggle={() => undefined} />
      </span>
      <span className="min-w-0 flex-1 truncate text-fg">{task.title}</span>
      <span className="flex shrink-0 items-center gap-2 text-xs text-fg-4">
        {task.group ? (
          <span className="inline-flex items-center gap-1.5">
            <GroupDot color={task.group.color} className="size-[7px]" />
            {task.group.name}
          </span>
        ) : null}
        {due ? <span className={cn(due.overdue && task.status !== "done" && "text-danger-ink")}>{due.label}</span> : null}
      </span>
    </Command.Item>
  );
}

export function CommandPalette() {
  const open = useUI((s) => s.palette);
  const mode = useUI((s) => s.paletteMode);
  const [search, setSearch] = useState("");
  const navigate = useNavigate();
  const tasks = useTasks(undefined, undefined, open);
  const groups = useGroups();
  const counts = useCounts();
  const { resolved } = useTheme();

  const close = () => {
    setUI({ palette: false });
    setSearch("");
  };
  const go = (path: string) => {
    close();
    navigate(path);
  };

  const q = search.trim();
  const matching = (tasks.data ?? []).filter((t) => t.status !== "archived");

  return (
    <D.Root open={open} onOpenChange={(o) => (o ? setUI({ palette: true }) : close())}>
      <D.Portal>
        <D.Overlay className="fixed inset-0 z-[60] bg-scrim backdrop-blur-[2px] data-[state=open]:animate-fade-in data-[state=closed]:animate-fade-out" />
        <D.Content
          aria-describedby={undefined}
          className="fixed top-[14vh] left-1/2 z-[61] w-[calc(100vw-24px)] max-w-[620px] -translate-x-1/2 overflow-hidden rounded-2xl bg-raised shadow-dialog outline-none data-[state=open]:animate-dialog-in data-[state=closed]:animate-dialog-out"
        >
          <D.Title className="sr-only">Search and commands</D.Title>
          <Command label="Search and commands" loop className="flex max-h-[min(560px,70vh)] flex-col">
            <div className="flex items-center gap-3 border-b border-line-soft px-4">
              <Search className="size-[18px] shrink-0 text-fg-4" aria-hidden="true" />
              <Command.Input
                value={search}
                onValueChange={setSearch}
                placeholder={mode === "tasks" ? "Search tasks…" : "Search tasks, jump to a list, or run a command…"}
                className="h-14 min-w-0 flex-1 bg-transparent text-[15px] text-fg outline-none placeholder:text-fg-4"
              />
              <Kbd>Esc</Kbd>
            </div>
            <Command.List className="min-h-0 flex-1 overflow-y-auto p-2 [&_[cmdk-group-heading]]:px-3 [&_[cmdk-group-heading]]:pt-2 [&_[cmdk-group-heading]]:pb-1.5 [&_[cmdk-group-heading]]:text-2xs [&_[cmdk-group-heading]]:font-semibold [&_[cmdk-group-heading]]:tracking-[0.06em] [&_[cmdk-group-heading]]:text-fg-4 [&_[cmdk-group-heading]]:uppercase">
              <Command.Empty className="px-3 py-10 text-center text-sm text-fg-3">
                {tasks.isPending ? "Searching…" : `Nothing matches “${q}”.`}
              </Command.Empty>

              {q || mode === "tasks" ? (
                <Command.Group heading="Tasks">
                  {matching.slice(0, q ? 50 : 8).map((t) => (
                    <TaskItem key={t.id} task={t} onSelect={() => go(`/app/tasks/${t.id}`)} />
                  ))}
                </Command.Group>
              ) : null}

              {q && mode !== "tasks" ? (
                <Command.Group heading="Create">
                  <Item
                    value={`create task ${q}`}
                    icon={<Plus className="size-4" />}
                    onSelect={() => {
                      close();
                      newTask({ title: q });
                    }}
                    hint={<CornerDownLeft className="size-3.5" />}
                  >
                    New task “<span className="text-fg">{q}</span>”
                  </Item>
                </Command.Group>
              ) : null}

              {mode !== "tasks" ? (
                <>
                  <Command.Group heading="Go to">
                    {VIEWS.map((v) => {
                      const Icon = v.icon;
                      const n = counts.data ? v.count(counts.data) : undefined;
                      return (
                        <Item
                          key={v.view}
                          value={`go ${v.label}`}
                          icon={<Icon className="size-4" style={{ color: v.color }} />}
                          onSelect={() => go(v.path)}
                          hint={
                            <>
                              {n ? <span className="tabular mr-2">{n}</span> : null}
                              <Kbd>{v.keys[0]}</Kbd>
                              <Kbd>{v.keys[1]}</Kbd>
                            </>
                          }
                        >
                          {v.label}
                        </Item>
                      );
                    })}
                    {(groups.data ?? []).map((g) => (
                      <Item key={g.id} value={`go group ${g.name}`} icon={<GroupDot color={g.color} />} onSelect={() => go(`/app/groups/${g.id}`)}>
                        {g.name}
                      </Item>
                    ))}
                    <Item value="go contacts" icon={<BookUser className="size-4" />} onSelect={() => go("/app/contacts")}>
                      Contacts
                    </Item>
                    <Item value="go activity" icon={<ActivityIcon className="size-4" />} onSelect={() => go("/app/activity")}>
                      Activity
                    </Item>
                    <Item value="go settings" icon={<Settings className="size-4" />} onSelect={() => go("/app/settings")}>
                      Settings
                    </Item>
                  </Command.Group>
                  <Command.Group heading="Actions">
                    <Item
                      value="new task"
                      icon={<Plus className="size-4" />}
                      onSelect={() => {
                        close();
                        newTask({});
                      }}
                      hint={<Kbd>C</Kbd>}
                    >
                      New task
                    </Item>
                    <Item value="add contact" icon={<UserRoundPlus className="size-4" />} onSelect={() => go("/app/contacts")}>
                      Add a contact
                    </Item>
                    <Item
                      value="toggle theme dark light"
                      icon={resolved === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
                      onSelect={() => {
                        toggleTheme();
                        close();
                      }}
                    >
                      Switch to {resolvedTheme() === "dark" ? "light" : "dark"} theme
                    </Item>
                    <Item
                      value="keyboard shortcuts help"
                      icon={<Keyboard className="size-4" />}
                      onSelect={() => {
                        close();
                        setUI({ shortcuts: true });
                      }}
                      hint={<Kbd>?</Kbd>}
                    >
                      Keyboard shortcuts
                    </Item>
                  </Command.Group>
                </>
              ) : null}
            </Command.List>
            <div className="flex items-center gap-4 border-t border-line-soft px-4 py-2.5 text-xs text-fg-4">
              <span className="flex items-center gap-1.5">
                <Kbd>↑</Kbd>
                <Kbd>↓</Kbd> to move
              </span>
              <span className="flex items-center gap-1.5">
                <Kbd>↵</Kbd> to open
              </span>
            </div>
          </Command>
        </D.Content>
      </D.Portal>
    </D.Root>
  );
}
