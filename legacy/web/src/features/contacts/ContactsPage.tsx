import { useQueryClient } from "@tanstack/react-query";
import { BookUser, Ellipsis, Mail, UserMinus, UserRoundPlus } from "lucide-react";
import { useState, type ReactNode } from "react";
import * as ep from "../../api/endpoints";
import { errorMessage } from "../../api/errors";
import { qk, useContacts } from "../../api/queries";
import type { Contact, ContactRequest, ContactsPayload } from "../../api/types";
import { ContactsIllo } from "../../components/brand/illustrations";
import { Avatar } from "../../components/ui/avatar";
import { Button, IconButton } from "../../components/ui/button";
import { Chip } from "../../components/ui/chip";
import { ConfirmDialog } from "../../components/ui/dialog";
import { EmptyState, ErrorState } from "../../components/ui/empty";
import { Menu, MenuContent, MenuItem, MenuTrigger } from "../../components/ui/menu";
import { Skeleton } from "../../components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../components/ui/tabs";
import { toast } from "../../components/ui/toast";
import { tr, useT, type Key } from "../../i18n";
import { distance, formatDate } from "../../lib/dates";
import { Page } from "../shell/Page";
import { AddContactDialog } from "./AddContactDialog";

export function ContactsPage() {
  const t = useT();
  const q = useContacts();
  const [adding, setAdding] = useState(false);
  const [tab, setTab] = useState<string | null>(null);
  const d = q.data;
  const current = tab ?? (d && d.contacts.length === 0 && d.incoming.length ? "requests" : "contacts");
  return (
    <Page
      title={t("contacts.title")}
      icon={<BookUser className="text-[#3e63dd]" strokeWidth={2} />}
      subtitle={t("contacts.subtitle")}
      actions={
        <Button size="sm" variant="primary" icon={<UserRoundPlus className="size-4" />} onClick={() => setAdding(true)}>
          {t("contacts.add")}
        </Button>
      }
    >
      {q.isPending ? (
        <ListSkeleton />
      ) : q.isError ? (
        <ErrorState message={errorMessage(q.error)} onRetry={() => void q.refetch()} />
      ) : (
        <Tabs value={current} onValueChange={setTab}>
          <TabsList className="mb-5">
            <TabsTrigger value="contacts" count={d!.contacts.length}>
              {t("contacts.tabContacts")}
            </TabsTrigger>
            <TabsTrigger value="requests" count={d!.incoming.length + d!.outgoing.length || undefined}>
              {t("contacts.tabRequests")}
              {d!.incoming.length ? <span className="size-1.5 rounded-full bg-accent" aria-label={t("common.new")} /> : null}
            </TabsTrigger>
            <TabsTrigger value="invites" count={d!.invites.length || undefined}>
              {t("contacts.tabInvites")}
            </TabsTrigger>
          </TabsList>
          <TabsContent value="contacts" className="outline-none">
            <ContactList data={d!} onAdd={() => setAdding(true)} />
          </TabsContent>
          <TabsContent value="requests" className="outline-none">
            <Requests data={d!} />
          </TabsContent>
          <TabsContent value="invites" className="outline-none">
            <Invites data={d!} onAdd={() => setAdding(true)} />
          </TabsContent>
        </Tabs>
      )}
      <AddContactDialog open={adding} onOpenChange={setAdding} />
    </Page>
  );
}

function Card({ children }: { children: ReactNode }) {
  return <ul className="overflow-hidden rounded-xl bg-surface shadow-card [&>li+li]:border-t [&>li+li]:border-line-soft">{children}</ul>;
}

function ListSkeleton() {
  return (
    <div className="mt-14 overflow-hidden rounded-xl bg-surface shadow-card">
      {[0, 1, 2, 3].map((i) => (
        <div key={i} className="flex items-center gap-3 border-t border-line-soft px-4 py-3 first:border-t-0">
          <Skeleton className="size-8 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-3 w-32" />
            <Skeleton className="h-2.5 w-44" />
          </div>
        </div>
      ))}
    </div>
  );
}

function ContactList({ data, onAdd }: { data: ContactsPayload; onAdd: () => void }) {
  const t = useT();
  if (!data.contacts.length) {
    return (
      <EmptyState
        art={<ContactsIllo />}
        title={t("contacts.emptyTitle")}
        action={
          <Button variant="primary" onClick={onAdd} icon={<UserRoundPlus className="size-4" />}>
            {t("contacts.addFirst")}
          </Button>
        }
      >
        {t("contacts.emptyBody")}
      </EmptyState>
    );
  }
  return (
    <Card>
      {data.contacts.map((c) => (
        <ContactRow key={c.id} c={c} />
      ))}
    </Card>
  );
}

function ContactRow({ c }: { c: Contact }) {
  const t = useT();
  const qc = useQueryClient();
  const [confirm, setConfirm] = useState(false);
  return (
    <li className="group/c flex items-center gap-3 px-4 py-3">
      <Avatar user={c.user} size="lg" />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-fg">{c.user.name}</p>
        <p className="truncate text-xs text-fg-3">{c.user.email}</p>
      </div>
      <span className="text-xs whitespace-nowrap text-fg-4 max-sm:hidden">
        {t("contacts.since", { date: formatDate(new Date(c.since), "date.pattern.shortMonthYear", t.locale) })}
      </span>
      <Menu>
        <MenuTrigger asChild>
          <IconButton label={t("contacts.actionsFor", { name: c.user.name })} size="sm">
            <Ellipsis className="size-4" />
          </IconButton>
        </MenuTrigger>
        <MenuContent align="end">
          <MenuItem
            icon={<Mail className="size-4" />}
            onSelect={() => void navigator.clipboard?.writeText(c.user.email).then(() => toast.success(tr()("contacts.emailCopied")))}
          >
            {t("contacts.copyEmail")}
          </MenuItem>
          <MenuItem danger icon={<UserMinus className="size-4" />} onSelect={() => setConfirm(true)}>
            {t("contacts.remove")}
          </MenuItem>
        </MenuContent>
      </Menu>
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={t("contacts.removeTitle", { name: c.user.name })}
        description={t("contacts.removeBody")}
        confirm={t("common.remove")}
        onConfirm={async () => {
          try {
            await ep.contacts.remove(c.id);
            await qc.invalidateQueries({ queryKey: qk.contacts });
            toast(tr()("contacts.removed", { name: c.user.name }));
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
    </li>
  );
}

function Requests({ data }: { data: ContactsPayload }) {
  const t = useT();
  if (!data.incoming.length && !data.outgoing.length) {
    return (
      <EmptyState art={<ContactsIllo />} title={t("contacts.noRequestsTitle")}>
        {t("contacts.noRequestsBody")}
      </EmptyState>
    );
  }
  return (
    <div className="flex flex-col gap-6">
      {data.incoming.length ? (
        <section>
          <h2 className="mb-2 px-1 text-[13px] font-semibold text-fg">{t("contacts.waiting")}</h2>
          <Card>
            {data.incoming.map((r) => (
              <RequestRow key={r.id} r={r} incoming />
            ))}
          </Card>
        </section>
      ) : null}
      {data.outgoing.length ? (
        <section>
          <h2 className="mb-2 px-1 text-[13px] font-semibold text-fg">{t("contacts.sent")}</h2>
          <Card>
            {data.outgoing.map((r) => (
              <RequestRow key={r.id} r={r} />
            ))}
          </Card>
        </section>
      ) : null}
    </div>
  );
}

function RequestRow({ r, incoming }: { r: ContactRequest; incoming?: boolean }) {
  const t = useT();
  const qc = useQueryClient();
  const [busy, setBusy] = useState<string | null>(null);
  async function run(kind: "accept" | "decline" | "cancel") {
    setBusy(kind);
    const say = tr();
    try {
      if (kind === "accept") {
        await ep.contacts.accept(r.id);
        toast.success(say("contacts.accepted", { name: r.user.name }));
      } else if (kind === "decline") {
        await ep.contacts.decline(r.id);
        toast(say("contacts.declined", { name: r.user.name }));
      } else {
        await ep.contacts.cancel(r.id);
        toast(say("contacts.cancelled"));
      }
      await Promise.all([qc.invalidateQueries({ queryKey: qk.contacts }), qc.invalidateQueries({ queryKey: qk.counts })]);
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setBusy(null);
    }
  }
  return (
    <li className="flex items-center gap-3 px-4 py-3 max-sm:flex-wrap">
      <Avatar user={r.user} size="lg" />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-fg">{r.user.name}</p>
        <p className="truncate text-xs text-fg-3">
          {r.user.email} · {distance(r.createdAt, t.locale)}
        </p>
      </div>
      {incoming ? (
        <div className="flex items-center gap-1.5 max-sm:w-full max-sm:justify-end">
          <Button size="sm" variant="ghost" loading={busy === "decline"} disabled={!!busy} onClick={() => run("decline")}>
            {t("common.decline")}
          </Button>
          <Button size="sm" variant="primary" loading={busy === "accept"} disabled={!!busy} onClick={() => run("accept")}>
            {t("common.accept")}
          </Button>
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <Chip tone="warn">{t("contacts.pending")}</Chip>
          <Button size="sm" variant="ghost" loading={busy === "cancel"} onClick={() => run("cancel")}>
            {t("common.cancel")}
          </Button>
        </div>
      )}
    </li>
  );
}

function WithdrawButton({ id, email }: { id: string; email: string }) {
  const t = useT();
  const qc = useQueryClient();
  const [busy, setBusy] = useState(false);
  return (
    <Button
      size="sm"
      variant="ghost"
      loading={busy}
      onClick={async () => {
        setBusy(true);
        try {
          await ep.contacts.cancel(id);
          await qc.invalidateQueries({ queryKey: qk.contacts });
          toast(tr()("contacts.withdrawn", { email }));
        } catch (err) {
          toast.error(errorMessage(err));
        } finally {
          setBusy(false);
        }
      }}
    >
      {t("contacts.withdraw")}
    </Button>
  );
}

const INVITE_STATUS: Record<string, Key> = {
  pending: "contacts.invite.pending",
  accepted: "contacts.invite.accepted",
  cancelled: "contacts.invite.cancelled",
  canceled: "contacts.invite.cancelled",
  revoked: "contacts.invite.cancelled",
  expired: "contacts.invite.expired",
  declined: "contacts.invite.declined",
};

function Invites({ data, onAdd }: { data: ContactsPayload; onAdd: () => void }) {
  const t = useT();
  if (!data.invites.length) {
    return (
      <EmptyState
        art={<ContactsIllo />}
        title={t("contacts.noInvitesTitle")}
        action={
          <Button variant="secondary" onClick={onAdd}>
            {t("contacts.inviteSomeone")}
          </Button>
        }
      >
        {t("contacts.noInvitesBody")}
      </EmptyState>
    );
  }
  return (
    <Card>
      {data.invites.map((i) => (
        <li key={i.id} className="flex items-center gap-3 px-4 py-3">
          <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-inset text-fg-3">
            <Mail className="size-4" />
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-fg">{i.email}</p>
            <p className="truncate text-xs text-fg-3">{t("contacts.sentAgo", { ago: distance(i.createdAt, t.locale) })}</p>
          </div>
          <Chip tone={i.status === "pending" ? "warn" : i.status === "accepted" ? "success" : "neutral"}>
            {INVITE_STATUS[i.status] ? t(INVITE_STATUS[i.status]!) : i.status}
          </Chip>
          {i.status === "pending" ? <WithdrawButton id={i.id} email={i.email} /> : null}
        </li>
      ))}
    </Card>
  );
}
