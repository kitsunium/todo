import { useQueryClient } from "@tanstack/react-query";
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
import { tr, useT, type Key } from "../../i18n";
import { formatDate } from "../../lib/dates";
import { normalize } from "../../lib/quickadd";
import { Page } from "../shell/Page";
import { QuickAdd } from "../tasks/QuickAdd";
import { TaskList } from "../tasks/TaskList";
import { CreateGroupDialog } from "./CreateGroupDialog";
import { InvitationCard } from "./InvitationBanner";

const ROLE: Record<Role, { label: Key; icon: ReactNode }> = {
  owner: { label: "role.owner", icon: <Crown className="size-3" /> },
  admin: { label: "role.admin", icon: <Shield className="size-3" /> },
  member: { label: "role.member", icon: null },
};

const YOU_ARE: Record<Role, Key> = { owner: "group.youAreOwner", admin: "group.youAreAdmin", member: "group.youAreMember" };

export function GroupPage() {
  const t = useT();
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
          subtitle={t("group.invited")}
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
      <Page title={t("group.fallbackTitle")}>
        {missing ? (
          <EmptyState
            art={<LostIllo />}
            title={t("group.goneTitle")}
            action={
              <Button asChild variant="secondary">
                <Link to="/app/today">{t("group.goToday")}</Link>
              </Button>
            }
          >
            {t("group.goneBody")}
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
            {t("group.members", { count: group.members.length })} · {t(YOU_ARE[group.role])}
          </span>
        </span>
      }
      actions={<GroupActions group={group} onMembers={() => setTab("members")} />}
    >
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="mb-5">
          <TabsTrigger value="tasks" count={group.openTasks || undefined}>
            {t("group.tabTasks")}
          </TabsTrigger>
          <TabsTrigger value="members" count={group.members.length}>
            {t("group.tabMembers")}
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
  const t = useT();
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
            <EmptyState art={<GroupIllo />} title={t("group.emptyTitle", { name: group.name })}>
              {t("group.emptyBody")}
            </EmptyState>
          }
        />
      )}
    </>
  );
}

function GroupActions({ group, onMembers }: { group: Group; onMembers: () => void }) {
  const t = useT();
  const me = useCurrentUser();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const [editing, setEditing] = useState(false);
  const [confirm, setConfirm] = useState<"delete" | "leave" | null>(null);
  const manage = group.role === "owner" || group.role === "admin";

  return (
    <>
      <Menu>
        <Tooltip content={t("group.settings")}>
          <MenuTrigger asChild>
            <IconButton label={t("group.settings")} size="sm">
              <Ellipsis className="size-4" />
            </IconButton>
          </MenuTrigger>
        </Tooltip>
        <MenuContent align="end" className="w-[210px]">
          {manage ? (
            <MenuItem icon={<Pencil className="size-4" />} onSelect={() => setEditing(true)}>
              {t("group.edit")}
            </MenuItem>
          ) : null}
          <MenuItem icon={<UsersRound className="size-4" />} onSelect={onMembers}>
            {t("group.tabMembers")}
          </MenuItem>
          <MenuSeparator />
          {group.role !== "owner" ? (
            <MenuItem danger icon={<DoorOpen className="size-4" />} onSelect={() => setConfirm("leave")}>
              {t("group.leave")}
            </MenuItem>
          ) : (
            <MenuItem danger icon={<Trash2 className="size-4" />} onSelect={() => setConfirm("delete")}>
              {t("group.delete")}
            </MenuItem>
          )}
        </MenuContent>
      </Menu>
      <CreateGroupDialog open={editing} onOpenChange={setEditing} group={group} />
      <ConfirmDialog
        open={confirm === "delete"}
        onOpenChange={(o) => !o && setConfirm(null)}
        title={t("group.deleteTitle", { name: group.name })}
        description={t("group.deleteBody")}
        confirm={t("group.deleteConfirm")}
        onConfirm={async () => {
          try {
            await ep.groups.remove(group.id);
            toast(tr()("group.deleted", { name: group.name }));
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
        title={t("group.leaveTitle", { name: group.name })}
        description={t("group.leaveBody")}
        confirm={t("group.leaveConfirm")}
        onConfirm={async () => {
          try {
            await ep.groups.removeMember(group.id, me.id);
            toast(tr()("group.left", { name: group.name }));
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
  const t = useT();
  const me = useCurrentUser();
  const manage = group.role === "owner" || group.role === "admin";
  const order: Record<Role, number> = { owner: 0, admin: 1, member: 2 };
  const members = [...group.members].sort((a, b) => order[a.role] - order[b.role] || a.user.name.localeCompare(b.user.name));
  return (
    <div>
      <div className="mb-3 flex items-center justify-between gap-3">
        <p className="text-sm text-fg-3">
          {manage ? t("group.inviteHint") : t("group.onlyManagersInvite")}
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
  const t = useT();
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
          {isMe ? <span className="font-normal text-fg-3"> ({t("common.youInline")})</span> : null}
        </p>
        <p className="truncate text-xs text-fg-3">
          {t("group.joined", { email: m.user.email, date: formatDate(new Date(m.joinedAt), "date.pattern.dayMonthYear", t.locale) })}
        </p>
      </div>
      {canRole ? (
        <Menu>
          <MenuTrigger asChild>
            <button type="button" className="inline-flex h-7 items-center gap-1 rounded-md px-2 text-xs font-medium text-fg-2 shadow-[inset_0_0_0_1px_var(--line)] hover:bg-hover data-[state=open]:bg-hover">
              {ROLE[m.role].icon}
              {t(ROLE[m.role].label)}
            </button>
          </MenuTrigger>
          <MenuContent align="end" className="w-[240px]">
            <MenuRadioGroup
              value={m.role}
              onValueChange={async (role) => {
                try {
                  const g = await ep.groups.setRole(group.id, m.user.id, role as "admin" | "member");
                  await refresh(g);
                  toast.success(tr()(role === "admin" ? "role.nowAdmin" : "role.nowMember", { name: m.user.name.split(" ")[0] ?? m.user.name }));
                } catch (err) {
                  toast.error(errorMessage(err));
                }
              }}
            >
              <MenuRadioItem value="admin" icon={<Shield className="size-4" />}>
                <span className="flex flex-col py-1">
                  <span>{t("role.admin")}</span>
                  <span className="text-xs text-fg-3">{t("role.admin.hint")}</span>
                </span>
              </MenuRadioItem>
              <MenuRadioItem value="member" icon={<UsersRound className="size-4" />}>
                <span className="flex flex-col py-1">
                  <span>{t("role.member")}</span>
                  <span className="text-xs text-fg-3">{t("role.member.hint")}</span>
                </span>
              </MenuRadioItem>
            </MenuRadioGroup>
          </MenuContent>
        </Menu>
      ) : (
        <Chip tone={m.role === "owner" ? "accent" : "neutral"}>
          {ROLE[m.role].icon}
          {t(ROLE[m.role].label)}
        </Chip>
      )}
      {canRemove ? (
        <Tooltip content={t("group.removeFromGroup")}>
          <IconButton label={t("group.removeName", { name: m.user.name })} size="sm" onClick={() => setConfirm(true)}>
            <UserMinus className="size-4" />
          </IconButton>
        </Tooltip>
      ) : (
        <span className="w-7" aria-hidden="true" />
      )}
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={t("group.removeTitle", { name: m.user.name })}
        description={t("group.removeBody", { group: group.name })}
        confirm={t("common.remove")}
        onConfirm={async () => {
          try {
            await ep.groups.removeMember(group.id, m.user.id);
            await refresh();
            toast(tr()("group.removed", { name: m.user.name }));
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
    </li>
  );
}

function InviteButton({ group }: { group: Group }) {
  const t = useT();
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
      toast.success(tr()("group.inviteSentTo", { name }));
    } catch (err) {
      toast.error(errorMessage(err));
    }
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="secondary" size="sm" icon={<UserRoundPlus className="size-4" />}>
          {t("common.invite")}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[300px] p-0">
        <div className="flex items-center gap-2 border-b border-line-soft px-3">
          <Search className="size-4 text-fg-4" aria-hidden="true" />
          <input autoFocus value={q} onChange={(e) => setQ(e.target.value)} placeholder={t("group.invitePlaceholder")} aria-label={t("common.searchContacts")} className="h-10 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-fg-4" />
        </div>
        <div className="max-h-[280px] overflow-y-auto p-1">
          {shown.length === 0 ? (
            <p className="px-2 py-3 text-sm text-fg-3">
              {candidates.length ? t("common.noOneMatches") : t("group.everyoneHere")}
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
                  <span className="text-xs font-medium whitespace-nowrap text-success-ink">{t("group.inviteSent")}</span>
                ) : (
                  <Button size="xs" variant="secondary" onClick={() => invite(u.id, u.name)}>
                    {t("common.invite")}
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
