// The property menus of a task — shared by the row's quick actions and the panel.
import { Check, Plus, Search, UserRoundX } from "lucide-react";
import { useMemo, useState } from "react";
import { useContacts, useCurrentUser, useGroups, useShareTask, useUpdateTask } from "../../api/queries";
import { PRIORITIES, type Priority, type Task, type UserRef } from "../../api/types";
import { Avatar } from "../../components/ui/avatar";
import { GroupDot } from "../../components/ui/group-color";
import { MenuItem, MenuLabel, MenuSeparator } from "../../components/ui/menu";
import { PRIORITY_LABEL, PriorityIcon } from "../../components/ui/priority";
import { toast } from "../../components/ui/toast";
import { cn } from "../../lib/cn";
import { normalize } from "../../lib/quickadd";

export function PriorityItems({ value, onPick }: { value: Priority; onPick: (p: Priority) => void }) {
  return (
    <>
      {PRIORITIES.map((p) => (
        <MenuItem
          key={p}
          icon={<PriorityIcon priority={p} />}
          shortcut={value === p ? <Check className="size-4 text-accent" /> : String(p)}
          onSelect={() => onPick(p)}
        >
          {PRIORITY_LABEL[p]}
        </MenuItem>
      ))}
    </>
  );
}

export function GroupItems({ value, onPick, onCreate }: { value?: string | undefined; onPick: (groupId: string) => void; onCreate?: () => void }) {
  const { data: groups = [] } = useGroups();
  return (
    <>
      <MenuItem icon={<span className="size-2 rounded-full border border-dashed border-fg-4" />} shortcut={!value ? <Check className="size-4 text-accent" /> : null} onSelect={() => onPick("")}>
        No group
      </MenuItem>
      {groups.length ? <MenuSeparator /> : null}
      {groups.map((g) => (
        <MenuItem key={g.id} icon={<GroupDot color={g.color} />} shortcut={value === g.id ? <Check className="size-4 text-accent" /> : null} onSelect={() => onPick(g.id)}>
          {g.name}
        </MenuItem>
      ))}
      {onCreate ? (
        <>
          <MenuSeparator />
          <MenuItem icon={<Plus className="size-4" />} onSelect={onCreate}>
            New group…
          </MenuItem>
        </>
      ) : null}
    </>
  );
}

/** Who may be assigned: the owner, the people it is shared with, the members of its group. */
export function useAssignable(task: Task): { direct: UserRef[]; viaShare: UserRef[] } {
  const me = useCurrentUser();
  const { data: groups = [] } = useGroups();
  const { data: contacts } = useContacts();
  return useMemo(() => {
    const seen = new Map<string, UserRef>();
    const add = (u: UserRef) => !seen.has(u.id) && seen.set(u.id, u);
    add({ id: me.id, name: me.name, email: me.email });
    add(task.owner);
    for (const u of task.sharedWith) add(u);
    const g = groups.find((x) => x.id === task.group?.id);
    for (const m of g?.members ?? []) add(m.user);
    const viaShare = task.can.share ? (contacts?.contacts ?? []).map((c) => c.user).filter((u) => !seen.has(u.id)) : [];
    return { direct: [...seen.values()], viaShare };
  }, [me, task, groups, contacts]);
}

export function AssigneeItems({ task, onDone }: { task: Task; onDone?: () => void }) {
  const me = useCurrentUser();
  const { direct, viaShare } = useAssignable(task);
  const update = useUpdateTask();
  const share = useShareTask();
  const assign = (u: UserRef | null) => {
    update.mutate({ id: task.id, patch: { assigneeId: u ? u.id : "" } });
    onDone?.();
  };
  // Chained with mutateAsync: the menu (and this component) is gone before
  // the share answers, and mutate's own callbacks would not run.
  const shareAndAssign = async (u: UserRef) => {
    onDone?.();
    try {
      await share.mutateAsync({ id: task.id, userId: u.id });
      await update.mutateAsync({ id: task.id, patch: { assigneeId: u.id } });
      toast.success(`Shared with and assigned to ${u.name}`);
    } catch {
      /* the mutation cache already said what went wrong */
    }
  };
  return (
    <>
      <MenuItem icon={<UserRoundX className="size-4" />} shortcut={!task.assignee ? <Check className="size-4 text-accent" /> : null} onSelect={() => assign(null)}>
        Unassigned
      </MenuItem>
      <MenuSeparator />
      {direct.map((u) => (
        <MenuItem key={u.id} icon={<Avatar user={u} size="xs" />} shortcut={task.assignee?.id === u.id ? <Check className="size-4 text-accent" /> : null} onSelect={() => assign(u)}>
          {u.id === me.id ? `${u.name} (you)` : u.name}
        </MenuItem>
      ))}
      {viaShare.length ? (
        <>
          <MenuSeparator />
          <MenuLabel>Share and assign</MenuLabel>
          {viaShare.map((u) => (
            <MenuItem
              key={u.id}
              icon={<Avatar user={u} size="xs" />}
              onSelect={() => void shareAndAssign(u)}
            >
              {u.name}
            </MenuItem>
          ))}
        </>
      ) : null}
    </>
  );
}

/** A searchable list of contacts with a check on those the task is shared with. */
export function ShareList({ task, autoFocus = true }: { task: Task; autoFocus?: boolean }) {
  const { data, isPending } = useContacts();
  const share = useShareTask();
  const [q, setQ] = useState("");
  const people = (data?.contacts ?? []).map((c) => c.user);
  const n = normalize(q);
  const shown = n ? people.filter((u) => normalize(u.name).includes(n) || normalize(u.email).includes(n)) : people;
  const sharedIds = new Set(task.sharedWith.map((u) => u.id));
  const toggle = (u: UserRef) => {
    const remove = sharedIds.has(u.id);
    share.mutate(
      { id: task.id, userId: u.id, remove },
      { onSuccess: () => toast.success(remove ? `Stopped sharing with ${u.name}` : `Shared with ${u.name}`) },
    );
  };
  return (
    <div className="w-[280px]">
      <div className="flex items-center gap-2 border-b border-line-soft px-3">
        <Search className="size-4 text-fg-4" aria-hidden="true" />
        <input
          autoFocus={autoFocus}
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Share with…"
          aria-label="Search contacts"
          className="h-10 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-fg-4"
        />
      </div>
      <div className="max-h-[260px] overflow-y-auto p-1">
        {isPending ? (
          <p className="px-2 py-3 text-sm text-fg-3">Loading contacts…</p>
        ) : shown.length === 0 ? (
          <p className="px-2 py-3 text-sm text-fg-3">{people.length ? "No one matches." : "Add contacts to share tasks with them."}</p>
        ) : (
          shown.map((u) => {
            const on = sharedIds.has(u.id);
            return (
              <button
                key={u.id}
                type="button"
                role="menuitemcheckbox"
                aria-checked={on}
                onClick={() => toggle(u)}
                className="flex h-10 w-full items-center gap-2.5 rounded-md px-2 text-left outline-none hover:bg-hover focus-visible:bg-hover"
              >
                <Avatar user={u} size="md" />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm text-fg">{u.name}</span>
                  <span className="block truncate text-xs text-fg-3">{u.email}</span>
                </span>
                <span
                  className={cn(
                    "flex size-4 items-center justify-center rounded-[5px] transition-colors",
                    on ? "bg-accent text-white" : "shadow-[inset_0_0_0_1.5px_var(--line-strong)]",
                  )}
                >
                  {on ? <Check className="size-3" strokeWidth={3} /> : null}
                </span>
              </button>
            );
          })
        )}
      </div>
    </div>
  );
}
