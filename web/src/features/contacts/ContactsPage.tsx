import { useQueryClient } from "@tanstack/react-query";
import { format, formatDistanceToNowStrict } from "date-fns";
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
import { Page } from "../shell/Page";
import { AddContactDialog } from "./AddContactDialog";

export function ContactsPage() {
  const q = useContacts();
  const [adding, setAdding] = useState(false);
  const [tab, setTab] = useState<string | null>(null);
  const d = q.data;
  const current = tab ?? (d && d.contacts.length === 0 && d.incoming.length ? "requests" : "contacts");
  return (
    <Page
      title="Contacts"
      icon={<BookUser className="text-[#3e63dd]" strokeWidth={2} />}
      subtitle="The people you share tasks and groups with."
      actions={
        <Button size="sm" variant="primary" icon={<UserRoundPlus className="size-4" />} onClick={() => setAdding(true)}>
          Add contact
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
              Contacts
            </TabsTrigger>
            <TabsTrigger value="requests" count={d!.incoming.length + d!.outgoing.length || undefined}>
              Requests
              {d!.incoming.length ? <span className="size-1.5 rounded-full bg-accent" aria-label="new" /> : null}
            </TabsTrigger>
            <TabsTrigger value="invites" count={d!.invites.length || undefined}>
              Invitations sent
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
  if (!data.contacts.length) {
    return (
      <EmptyState art={<ContactsIllo />} title="No contacts yet" action={<Button variant="primary" onClick={onAdd} icon={<UserRoundPlus className="size-4" />}>Add your first contact</Button>}>
        Add people by email to share tasks with them and work in groups together.
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
  const qc = useQueryClient();
  const [confirm, setConfirm] = useState(false);
  return (
    <li className="group/c flex items-center gap-3 px-4 py-3">
      <Avatar user={c.user} size="lg" />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-fg">{c.user.name}</p>
        <p className="truncate text-xs text-fg-3">{c.user.email}</p>
      </div>
      <span className="text-xs text-fg-4 max-sm:hidden">Since {format(new Date(c.since), "MMM yyyy")}</span>
      <Menu>
        <MenuTrigger asChild>
          <IconButton label={`Actions for ${c.user.name}`} size="sm">
            <Ellipsis className="size-4" />
          </IconButton>
        </MenuTrigger>
        <MenuContent align="end">
          <MenuItem icon={<Mail className="size-4" />} onSelect={() => void navigator.clipboard?.writeText(c.user.email).then(() => toast.success("Email copied"))}>
            Copy email
          </MenuItem>
          <MenuItem danger icon={<UserMinus className="size-4" />} onSelect={() => setConfirm(true)}>
            Remove contact…
          </MenuItem>
        </MenuContent>
      </Menu>
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={`Remove ${c.user.name}?`}
        description="You’ll no longer be able to share new tasks with each other. Tasks already shared stay shared."
        confirm="Remove"
        onConfirm={async () => {
          try {
            await ep.contacts.remove(c.id);
            await qc.invalidateQueries({ queryKey: qk.contacts });
            toast(`${c.user.name} was removed from your contacts`);
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
    </li>
  );
}

function Requests({ data }: { data: ContactsPayload }) {
  if (!data.incoming.length && !data.outgoing.length) {
    return (
      <EmptyState art={<ContactsIllo />} title="No pending requests">
        Requests you send and receive wait here until someone answers.
      </EmptyState>
    );
  }
  return (
    <div className="flex flex-col gap-6">
      {data.incoming.length ? (
        <section>
          <h2 className="mb-2 px-1 text-[13px] font-semibold text-fg">Waiting for you</h2>
          <Card>
            {data.incoming.map((r) => (
              <RequestRow key={r.id} r={r} incoming />
            ))}
          </Card>
        </section>
      ) : null}
      {data.outgoing.length ? (
        <section>
          <h2 className="mb-2 px-1 text-[13px] font-semibold text-fg">Sent by you</h2>
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
  const qc = useQueryClient();
  const [busy, setBusy] = useState<string | null>(null);
  async function run(kind: "accept" | "decline" | "cancel") {
    setBusy(kind);
    try {
      if (kind === "accept") {
        await ep.contacts.accept(r.id);
        toast.success(`You and ${r.user.name} are now contacts`);
      } else if (kind === "decline") {
        await ep.contacts.decline(r.id);
        toast(`Declined ${r.user.name}’s request`);
      } else {
        await ep.contacts.cancel(r.id);
        toast("Request cancelled");
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
          {r.user.email} · {formatDistanceToNowStrict(new Date(r.createdAt), { addSuffix: true })}
        </p>
      </div>
      {incoming ? (
        <div className="flex items-center gap-1.5 max-sm:w-full max-sm:justify-end">
          <Button size="sm" variant="ghost" loading={busy === "decline"} disabled={!!busy} onClick={() => run("decline")}>
            Decline
          </Button>
          <Button size="sm" variant="primary" loading={busy === "accept"} disabled={!!busy} onClick={() => run("accept")}>
            Accept
          </Button>
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <Chip tone="warn">Pending</Chip>
          <Button size="sm" variant="ghost" loading={busy === "cancel"} onClick={() => run("cancel")}>
            Cancel
          </Button>
        </div>
      )}
    </li>
  );
}

function WithdrawButton({ id, email }: { id: string; email: string }) {
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
          toast(`Withdrew the invitation to ${email}`);
        } catch (err) {
          toast.error(errorMessage(err));
        } finally {
          setBusy(false);
        }
      }}
    >
      Withdraw
    </Button>
  );
}

function Invites({ data, onAdd }: { data: ContactsPayload; onAdd: () => void }) {
  if (!data.invites.length) {
    return (
      <EmptyState art={<ContactsIllo />} title="No invitations sent" action={<Button variant="secondary" onClick={onAdd}>Invite someone</Button>}>
        Invite someone who isn’t on Todo yet — they’ll become a contact when they join.
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
            <p className="truncate text-xs text-fg-3">Sent {formatDistanceToNowStrict(new Date(i.createdAt), { addSuffix: true })}</p>
          </div>
          <Chip tone={i.status === "pending" ? "warn" : i.status === "accepted" ? "success" : "neutral"} className="capitalize">
            {i.status === "accepted" ? "Joined" : i.status}
          </Chip>
          {i.status === "pending" ? <WithdrawButton id={i.id} email={i.email} /> : null}
        </li>
      ))}
    </Card>
  );
}
