import { useQueryClient } from "@tanstack/react-query";
import {
  Activity as ActivityIcon,
  BookUser,
  Keyboard,
  Languages,
  LogOut,
  Plus,
  Search,
  Settings,
  ChevronsUpDown,
} from "lucide-react";
import { NavLink, useNavigate } from "react-router";
import * as ep from "../../api/endpoints";
import { useContacts, useCounts, useCurrentUser, useGroups, useInvitations } from "../../api/queries";
import { Logo } from "../../components/brand/logo";
import { Avatar } from "../../components/ui/avatar";
import { IconButton } from "../../components/ui/button";
import { Badge } from "../../components/ui/chip";
import { GroupDot } from "../../components/ui/group-color";
import { Kbd, MOD } from "../../components/ui/kbd";
import {
  Menu,
  MenuContent,
  MenuItem,
  MenuSegmented,
  MenuSeparator,
  MenuTrigger,
} from "../../components/ui/menu";
import { Skeleton } from "../../components/ui/skeleton";
import { toast } from "../../components/ui/toast";
import { Tooltip } from "../../components/ui/tooltip";
import { cn } from "../../lib/cn";
import { LOCALES, tr, useLocale, useT, type Locale } from "../../i18n";
import { useSwitchLocale } from "../settings/language";
import { CreateGroupDialog } from "../groups/CreateGroupDialog";
import { VIEWS } from "./nav";
import { newTask, setUI } from "./store";
import { useState } from "react";

const ITEM =
  "group/nav relative flex h-[30px] items-center gap-2.5 rounded-md px-2 text-[13.5px] text-fg-2 outline-none transition-colors duration-100 " +
  "hover:bg-hover hover:text-fg focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring " +
  "aria-[current=page]:bg-surface aria-[current=page]:font-medium aria-[current=page]:text-fg aria-[current=page]:shadow-card";

function Count({ n, tone }: { n: number | undefined; tone?: "danger" }) {
  if (!n) return null;
  return <span className={cn("tabular ml-auto text-xs", tone === "danger" ? "font-medium text-danger-ink" : "text-fg-4")}>{n}</span>;
}

export function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const t = useT();
  const counts = useCounts();
  const groups = useGroups();
  const contacts = useContacts();
  const invitations = useInvitations();
  const c = counts.data;
  const pendingContacts = contacts.data?.incoming.length ?? 0;
  const [creating, setCreating] = useState(false);

  return (
    <nav className="flex h-full w-full flex-col bg-sidebar" aria-label={t("app.mainNav")}>
      <div className="flex h-12 shrink-0 items-center justify-between pr-2 pl-4">
        <Logo />
        <Tooltip content={t("common.search")} keys={[MOD, "K"]} side="bottom">
          <IconButton label={t("common.searchAndCommands")} size="sm" onClick={() => setUI({ palette: true, paletteMode: "all" })}>
            <Search className="size-4" />
          </IconButton>
        </Tooltip>
      </div>

      <div className="px-3 pt-1 pb-3">
        <button
          type="button"
          onClick={() => {
            onNavigate?.();
            newTask({});
          }}
          className="group flex h-8 w-full items-center gap-2 rounded-lg bg-surface px-2.5 text-[13.5px] font-medium text-fg shadow-card transition-shadow hover:shadow-[0_0_0_1px_var(--line-strong),0_1px_2px_rgb(0_0_0/0.06)] focus-visible:outline-2 focus-visible:outline-ring"
        >
          <span className="flex size-[18px] items-center justify-center rounded-full bg-accent text-white shadow-[inset_0_1px_0_rgb(255_255_255/0.25)]">
            <Plus className="size-3" strokeWidth={3} />
          </span>
          {t("sidebar.newTask")}
          <Kbd className="ml-auto">C</Kbd>
        </button>
      </div>

      <div className="no-scrollbar flex min-h-0 flex-1 flex-col overflow-y-auto px-3 pb-3">
        <ul className="flex flex-col gap-px">
          {VIEWS.map((v) => {
            const Icon = v.icon;
            const overdue = v.view === "today" && (c?.overdue ?? 0) > 0;
            return (
              <li key={v.view}>
                <NavLink to={v.path} className={ITEM} onClick={onNavigate}>
                  {({ isActive }) => (
                    <>
                      <Icon
                        className={cn("size-4 shrink-0", isActive ? "" : "text-fg-3 group-hover/nav:text-fg-2")}
                        style={isActive ? { color: v.color } : undefined}
                        strokeWidth={isActive ? 2.1 : 1.8}
                        aria-hidden="true"
                      />
                      <span className="truncate">{t(v.label)}</span>
                      {c ? <Count n={v.count(c)} {...(overdue ? { tone: "danger" as const } : {})} /> : null}
                    </>
                  )}
                </NavLink>
              </li>
            );
          })}
        </ul>

        <div className="mt-5 flex h-7 items-center justify-between pr-1 pl-2">
          <span className="text-2xs font-semibold tracking-[0.06em] text-fg-4 uppercase">{t("sidebar.groups")}</span>
          <Tooltip content={t("sidebar.newGroup")} side="right">
            <IconButton label={t("sidebar.newGroup")} size="xs" onClick={() => setCreating(true)}>
              <Plus className="size-3.5" />
            </IconButton>
          </Tooltip>
        </div>
        <ul className="flex flex-col gap-px">
          {groups.isPending ? (
            <>
              <li className="flex h-[30px] items-center gap-2.5 px-2"><Skeleton className="size-2 rounded-full" /><Skeleton className="h-2.5 w-20" /></li>
              <li className="flex h-[30px] items-center gap-2.5 px-2"><Skeleton className="size-2 rounded-full" /><Skeleton className="h-2.5 w-14" /></li>
            </>
          ) : groups.data?.length ? (
            groups.data.map((g) => (
              <li key={g.id}>
                <NavLink to={`/app/groups/${g.id}`} className={ITEM} onClick={onNavigate}>
                  <span className="flex size-4 items-center justify-center">
                    <GroupDot color={g.color} className="size-[9px]" />
                  </span>
                  <span className="truncate">{g.name}</span>
                  <Count n={c?.groups[g.id] ?? g.openTasks} />
                </NavLink>
              </li>
            ))
          ) : (
            <li>
              <button
                type="button"
                onClick={() => setCreating(true)}
                className="flex h-[30px] w-full items-center gap-2.5 rounded-md px-2 text-[13.5px] text-fg-3 hover:bg-hover hover:text-fg"
              >
                <Plus className="size-4" aria-hidden="true" /> {t("sidebar.createGroup")}
              </button>
            </li>
          )}
        </ul>

        <ul className="mt-5 flex flex-col gap-px">
          <li>
            <NavLink to="/app/contacts" className={ITEM} onClick={onNavigate}>
              {({ isActive }) => (
                <>
                  <BookUser className={cn("size-4 shrink-0", isActive ? "text-accent" : "text-fg-3")} strokeWidth={isActive ? 2.1 : 1.8} aria-hidden="true" />
                  <span className="truncate">{t("sidebar.contacts")}</span>
                  {pendingContacts ? (
                    <Badge className="ml-auto" aria-label={t("sidebar.pending", { count: pendingContacts })}>
                      {pendingContacts}
                    </Badge>
                  ) : null}
                </>
              )}
            </NavLink>
          </li>
          <li>
            <NavLink to="/app/activity" className={ITEM} onClick={onNavigate}>
              {({ isActive }) => {
                const unread = (c?.unread ?? 0) + (invitations.data?.length ?? 0);
                return (
                  <>
                    <ActivityIcon className={cn("size-4 shrink-0", isActive ? "text-accent" : "text-fg-3")} strokeWidth={isActive ? 2.1 : 1.8} aria-hidden="true" />
                    <span className="truncate">{t("sidebar.activity")}</span>
                    {unread && !isActive ? (
                      <Badge className="ml-auto" aria-label={t("sidebar.unread", { count: unread })}>
                        {unread}
                      </Badge>
                    ) : null}
                  </>
                );
              }}
            </NavLink>
          </li>
        </ul>
      </div>

      <div className="shrink-0 border-t border-line-soft p-2">
        <UserMenu onNavigate={onNavigate} />
      </div>
      <CreateGroupDialog open={creating} onOpenChange={setCreating} />
    </nav>
  );
}

function UserMenu({ onNavigate }: { onNavigate: (() => void) | undefined }) {
  const me = useCurrentUser();
  const t = useT();
  const locale = useLocale();
  const switchLocale = useSwitchLocale();
  const navigate = useNavigate();
  const qc = useQueryClient();

  async function signOut() {
    try {
      await ep.auth.logout();
    } catch {
      /* the session may already be gone: sign out locally anyway */
    }
    qc.clear();
    navigate("/login", { replace: true });
    toast(tr()("app.signedOut"));
  }

  return (
    <Menu>
      <MenuTrigger asChild>
        <button
          type="button"
          className="flex h-11 w-full items-center gap-2.5 rounded-lg px-2 text-left outline-none transition-colors hover:bg-hover data-[state=open]:bg-hover focus-visible:outline-2 focus-visible:outline-ring"
        >
          <Avatar user={me} size="lg" />
          <span className="min-w-0 flex-1">
            <span className="block truncate text-[13px] leading-4 font-medium text-fg">{me.name}</span>
            <span className="block truncate text-xs text-fg-3">{me.email}</span>
          </span>
          <ChevronsUpDown className="size-3.5 shrink-0 text-fg-4" aria-hidden="true" />
        </button>
      </MenuTrigger>
      <MenuContent side="top" align="start" sideOffset={8} className="w-[232px]">
        <MenuItem
          icon={<Settings className="size-4" />}
          shortcut={`${MOD},`}
          onSelect={() => {
            onNavigate?.();
            navigate("/app/settings");
          }}
        >
          {t("user.settings")}
        </MenuItem>
        <MenuSegmented<Locale>
          icon={<Languages className="size-4" />}
          label={t("lang.label")}
          value={locale}
          options={LOCALES.map((l) => ({ value: l, short: tr(l)(`lang.${l}.short`), label: tr(l)(`lang.${l}`), lang: l }))}
          onChange={switchLocale}
        />
        <MenuItem icon={<Keyboard className="size-4" />} shortcut="?" onSelect={() => setUI({ shortcuts: true })}>
          {t("user.shortcuts")}
        </MenuItem>
        <MenuSeparator />
        <MenuItem icon={<LogOut className="size-4" />} onSelect={signOut}>
          {t("user.signOut")}
        </MenuItem>
      </MenuContent>
    </Menu>
  );
}
