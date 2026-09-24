import { useQueryClient } from "@tanstack/react-query";
import { format } from "date-fns";
import { Crown, DoorOpen, Ellipsis, Pencil, Search, Shield, Trash2, UserMinus, UserRoundPlus, UsersRound } from "lucide-react";
import { useState, type ReactNode } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { ApiError } from "../../api/client";
import * as ep from "../../api/endpoints";
import { errorMessage } from "../../api/errors";
import { qk, useContacts, useCurrentUser, useGroup, useInvitations, useTasks } from "../../api/queries";
import type { Group, Member, Role } from "../../api/types";
import { GroupIllo, LostIllo } from "../../components/brand/illustrations";
import { Avatar, AvatarStack } from "../../components/ui/avatar";
import { Button, IconButton } from "../../components/ui/button";
import { Chip } from "../../components/ui/chip";
import { ConfirmDialog } from "../../components/ui/dialog";
import { EmptyState, ErrorState } from "../../components/ui/empty";
import { GroupBadge } from "../../components/ui/group-color";
import { Menu, MenuContent, MenuItem, MenuRadioGroup, MenuRadioItem, MenuSeparator, MenuTrigger } from "../../components/ui/menu";
import { Popover, PopoverContent, PopoverTrigger } from "../../components/ui/popover";
import { Skeleton, TaskListSkeleton } from "../../components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../components/ui/tabs";
import { toast } from "../../components/ui/toast";
import { Tooltip } from "../../components/ui/tooltip";
import { normalize } from "../../lib/quickadd";
import { Page } from "../shell/Page";
import { QuickAdd } from "../tasks/QuickAdd";
import { TaskList } from "../tasks/TaskList";
import { CreateGroupDialog } from "./CreateGroupDialog";
import { InvitationCard } from "./InvitationBanner";

const ROLE: Record<Role, { label: string; icon: ReactNode }> = {
  owner: { label: "Owner", icon: <Crown className="size-3" /> },
  admin: { label: "Admin", icon: <Shield className="size-3" /> },
  member: { label: "Member", icon: null },
};

export function GroupPage() {
  const { groupId = "" } = useParams();
  const g = useGroup(groupId);
  const invitations = useInvitations();
  const [tab, setTab] = useState("tasks");

  if (g.isError) {
    const missing = g.error instanceof ApiError && (g.error.status === 404 || g.error.status === 403);
    // A group invitation mail links here: the invitee is not a member yet.
    const invitation = invitations.data?.find((i) => i.group.id === groupId);
    if (missing && invitation) {
      return (
        <Page
          title={invitation.group.name}
          icon={<GroupBadge name={invitation.group.name} color={invitation.group.color} size="md" />}
          subtitle="You’re invited to this group."
        >
          <InvitationCard inv={invitation} onAccepted={() => void g.refetch()} />
        </Page>
      );
    }
    if (missing && invitations.isPending) {
      return (
        <Page title={<Skeleton className="h-6 w-40" />}>
          <TaskListSkeleton />
        </Page>
      );
    }
    return (
      <Page title="Group">
        {missing ? (
          <EmptyState art={<LostIllo />} title="This group isn’t here" action={<Button asChild variant="secondary"><Link to="/app/today">Go to Today</Link></Button>}>
            It was deleted, or you’re no longer a member.
          </EmptyState>
        ) : (
          <ErrorState message={errorMessage(g.error)} onRetry={() => void g.refetch()} />
        )}
      </Page>
    );
  }
  if (!g.data) {
    return (
      <Page title={<Skeleton className="h-6 w-40" />}>
        <TaskListSkeleton />
      </Page>
    );
  }
  const group = g.data;
  return (
    <Page
      key={group.id}
      title={group.name}
      icon={<GroupBadge name={group.name} color={group.color} size="md" />}
      subtitle={
        <span className="flex items-center gap-2">
          <AvatarStack users={group.members.map((m) => m.user)} size="sm" max={5} />
          <span>
            {group.members.length} member{group.members.length === 1 ? "" : "s"} · {ROLE[group.role].label === "Member" ? "You’re a member" : `You’re ${group.role === "owner" ? "the owner" : "an admin"}`}
          </span>
        </span>
      }
      actions={<GroupActions group={group} onMembers={() => setTab("members")} />}
    >
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="mb-5">
          <TabsTrigger value="tasks" count={group.openTasks || undefined}>
            Tasks
          </TabsTrigger>
          <TabsTrigger value="members" count={group.members.length}>
            Members
          </TabsTrigger>
        </TabsList>
        <TabsContent value="tasks" className="outline-none">
          <GroupTasks group={group} />
        </TabsContent>
        <TabsContent value="members" className="outline-none">
          <Members group={group} />
        </TabsContent>
      </Tabs>
    </Page>
  );
}

function GroupTasks({ group }: { group: Group }) {
  const q = useTasks("group", group.id);
  return (
    <>
      <QuickAdd view="group" groupId={group.id} defaults={{ groupId: group.id }} />
      {q.isPending ? (
        <TaskListSkeleton />
      ) : q.isError ? (
        <ErrorState message={errorMessage(q.error)} onRetry={() => void q.refetch()} />
      ) : (
        <TaskList
          tasks={q.data}
          view="group"
          showGroup={false}
          empty={
            <EmptyState art={<GroupIllo />} title={`Nothing in ${group.name} yet`}>
              Add the first task above. Everyone in the group sees it and can pick it up.
            </EmptyState>
          }
        />
      )}
    </>
  );
}

function GroupActions({ group, onMembers }: { group: Group; onMembers: () => void }) {
  const me = useCurrentUser();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const [editing, setEditing] = useState(false);
  const [confirm, setConfirm] = useState<"delete" | "leave" | null>(null);
  const manage = group.role === "owner" || group.role === "admin";

  return (
    <>
      <Menu>
        <Tooltip content="Group settings">
          <MenuTrigger asChild>
            <IconButton label="Group settings" size="sm">
              <Ellipsis className="size-4" />
            </IconButton>
          </MenuTrigger>
        </Tooltip>
        <MenuContent align="end" className="w-[210px]">
          {manage ? (
            <MenuItem icon={<Pencil className="size-4" />} onSelect={() => setEditing(true)}>
              Rename or recolor
            </MenuItem>
          ) : null}
          <MenuItem icon={<UsersRound className="size-4" />} onSelect={onMembers}>
            Members
          </MenuItem>
          <MenuSeparator />
          {group.role !== "owner" ? (
            <MenuItem danger icon={<DoorOpen className="size-4" />} onSelect={() => setConfirm("leave")}>
              Leave group…
            </MenuItem>
          ) : (
            <MenuItem danger icon={<Trash2 className="size-4" />} onSelect={() => setConfirm("delete")}>
              Delete group…
            </MenuItem>
          )}
        </MenuContent>
      </Menu>
      <CreateGroupDialog open={editing} onOpenChange={setEditing} group={group} />
      <ConfirmDialog
        open={confirm === "delete"}
        onOpenChange={(o) => !o && setConfirm(null)}
        title={`Delete ${group.name}?`}
        description="The group and its tasks are deleted for every member. This can’t be undone."
        confirm="Delete group"
        onConfirm={async () => {
          try {
            await ep.groups.remove(group.id);
            toast(`${group.name} was deleted`);
            navigate("/app/today", { replace: true });
            qc.removeQueries({ queryKey: qk.group(group.id) });
            await Promise.all([qc.invalidateQueries({ queryKey: qk.groups }), qc.invalidateQueries({ queryKey: qk.tasksRoot }), qc.invalidateQueries({ queryKey: qk.counts })]);
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
      <ConfirmDialog
        open={confirm === "leave"}
        onOpenChange={(o) => !o && setConfirm(null)}
        title={`Leave ${group.name}?`}
        description="You’ll stop seeing its tasks. A member can invite you again later."
        confirm="Leave group"
        onConfirm={async () => {
          try {
            await ep.groups.removeMember(group.id, me.id);
            toast(`You left ${group.name}`);
            navigate("/app/today", { replace: true });
            qc.removeQueries({ queryKey: qk.group(group.id) });
            await Promise.all([qc.invalidateQueries({ queryKey: qk.groups }), qc.invalidateQueries({ queryKey: qk.tasksRoot }), qc.invalidateQueries({ queryKey: qk.counts })]);
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
    </>
  );
}

function Members({ group }: { group: Group }) {
  const me = useCurrentUser();
  const manage = group.role === "owner" || group.role === "admin";
  const order: Record<Role, number> = { owner: 0, admin: 1, member: 2 };
  const members = [...group.members].sort((a, b) => order[a.role] - order[b.role] || a.user.name.localeCompare(b.user.name));
  return (
    <div>
      <div className="mb-3 flex items-center justify-between gap-3">
        <p className="text-sm text-fg-3">
          {manage ? "Invite contacts to share this group’s tasks." : "Only owners and admins can invite people."}
        </p>
        {manage ? <InviteButton group={group} /> : null}
      </div>
      <ul className="overflow-hidden rounded-xl bg-surface shadow-card">
        {members.map((m, i) => (
          <MemberRow key={m.user.id} group={group} m={m} isMe={m.user.id === me.id} first={i === 0} />
        ))}
      </ul>
    </div>
  );
}

function MemberRow({ group, m, isMe, first }: { group: Group; m: Member; isMe: boolean; first: boolean }) {
  const qc = useQueryClient();
  const [confirm, setConfirm] = useState(false);
  const canRemove = !isMe && m.role !== "owner" && (group.role === "owner" || (group.role === "admin" && m.role === "member"));
  const canRole = group.role === "owner" && m.role !== "owner";
  const refresh = (g?: Group) => {
    if (g) qc.setQueryData(qk.group(g.id), g);
    return Promise.all([qc.invalidateQueries({ queryKey: qk.group(group.id) }), qc.invalidateQueries({ queryKey: qk.groups })]);
  };
  return (
    <li className={`flex items-center gap-3 px-4 py-3 ${first ? "" : "border-t border-line-soft"}`}>
      <Avatar user={m.user} size="lg" />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-fg">
          {m.user.name}
          {isMe ? <span className="font-normal text-fg-3"> (you)</span> : null}
        </p>
        <p className="truncate text-xs text-fg-3">
          {m.user.email} · joined {format(new Date(m.joinedAt), "MMM d, yyyy")}
        </p>
      </div>
      {canRole ? (
        <Menu>
          <MenuTrigger asChild>
            <button type="button" className="inline-flex h-7 items-center gap-1 rounded-md px-2 text-xs font-medium text-fg-2 shadow-[inset_0_0_0_1px_var(--line)] hover:bg-hover data-[state=open]:bg-hover">
              {ROLE[m.role].icon}
              {ROLE[m.role].label}
            </button>
          </MenuTrigger>
          <MenuContent align="end" className="w-[240px]">
            <MenuRadioGroup
              value={m.role}
              onValueChange={async (role) => {
                try {
                  const g = await ep.groups.setRole(group.id, m.user.id, role as "admin" | "member");
                  await refresh(g);
                  toast.success(`${m.user.name.split(" ")[0]} is now ${role === "admin" ? "an admin" : "a member"}`);
                } catch (err) {
                  toast.error(errorMessage(err));
                }
              }}
            >
              <MenuRadioItem value="admin" icon={<Shield className="size-4" />}>
                <span className="flex flex-col py-1">
                  <span>Admin</span>
                  <span className="text-xs text-fg-3">Invites and removes members</span>
                </span>
              </MenuRadioItem>
              <MenuRadioItem value="member" icon={<UsersRound className="size-4" />}>
                <span className="flex flex-col py-1">
                  <span>Member</span>
                  <span className="text-xs text-fg-3">Sees and edits the tasks</span>
                </span>
              </MenuRadioItem>
            </MenuRadioGroup>
          </MenuContent>
        </Menu>
      ) : (
        <Chip tone={m.role === "owner" ? "accent" : "neutral"}>
          {ROLE[m.role].icon}
          {ROLE[m.role].label}
        </Chip>
      )}
      {canRemove ? (
        <Tooltip content="Remove from the group">
          <IconButton label={`Remove ${m.user.name}`} size="sm" onClick={() => setConfirm(true)}>
            <UserMinus className="size-4" />
          </IconButton>
        </Tooltip>
      ) : (
        <span className="w-7" aria-hidden="true" />
      )}
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={`Remove ${m.user.name}?`}
        description={`They’ll stop seeing ${group.name}’s tasks. You can invite them again later.`}
        confirm="Remove"
        onConfirm={async () => {
          try {
            await ep.groups.removeMember(group.id, m.user.id);
            await refresh();
            toast(`${m.user.name} was removed`);
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
    </li>
  );
}

function InviteButton({ group }: { group: Group }) {
  const { data } = useContacts();
  const [q, setQ] = useState("");
  const [sent, setSent] = useState<Set<string>>(new Set());
  const members = new Set(group.members.map((m) => m.user.id));
  const candidates = (data?.contacts ?? []).map((c) => c.user).filter((u) => !members.has(u.id));
  const n = normalize(q);
  const shown = n ? candidates.filter((u) => normalize(u.name).includes(n) || normalize(u.email).includes(n)) : candidates;

  async function invite(userId: string, name: string) {
    try {
      await ep.groups.invite(group.id, userId);
      setSent((s) => new Set(s).add(userId));
      toast.success(`Invitation sent to ${name}`);
    } catch (err) {
      toast.error(errorMessage(err));
    }
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="secondary" size="sm" icon={<UserRoundPlus className="size-4" />}>
          Invite
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[300px] p-0">
        <div className="flex items-center gap-2 border-b border-line-soft px-3">
          <Search className="size-4 text-fg-4" aria-hidden="true" />
          <input autoFocus value={q} onChange={(e) => setQ(e.target.value)} placeholder="Invite a contact…" aria-label="Search contacts" className="h-10 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-fg-4" />
        </div>
        <div className="max-h-[280px] overflow-y-auto p-1">
          {shown.length === 0 ? (
            <p className="px-2 py-3 text-sm text-fg-3">
              {candidates.length ? "No one matches." : "Everyone in your contacts is already here. Add contacts to invite more people."}
            </p>
          ) : (
            shown.map((u) => (
              <div key={u.id} className="flex h-11 items-center gap-2.5 rounded-md px-2 hover:bg-hover">
                <Avatar user={u} size="md" />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm text-fg">{u.name}</span>
                  <span className="block truncate text-xs text-fg-3">{u.email}</span>
                </span>
                {sent.has(u.id) ? (
                  <span className="text-xs font-medium text-success-ink">Invited</span>
                ) : (
                  <Button size="xs" variant="secondary" onClick={() => invite(u.id, u.name)}>
                    Invite
                  </Button>
                )}
              </div>
            ))
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
