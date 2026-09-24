import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage } from "../../api/errors";
import { qk, useInvitations } from "../../api/queries";
import type { Invitation } from "../../api/types";
import { Avatar } from "../../components/ui/avatar";
import { Button } from "../../components/ui/button";
import { GroupDot } from "../../components/ui/group-color";
import { toast } from "../../components/ui/toast";
import { cn } from "../../lib/cn";

/** Pending group invitations, answerable in place. */
export function InvitationBanner({ className }: { className?: string }) {
  const { data } = useInvitations();
  if (!data?.length) return null;
  return (
    <div className={cn("mb-5 flex flex-col gap-2", className)}>
      {data.map((inv) => (
        <InvitationCard key={inv.id} inv={inv} />
      ))}
    </div>
  );
}

export function InvitationCard({ inv, onAccepted }: { inv: Invitation; onAccepted?: () => void }) {
  const qc = useQueryClient();
  const navigate = useNavigate();
  const [busy, setBusy] = useState<"accept" | "decline" | null>(null);

  async function answer(accept: boolean) {
    setBusy(accept ? "accept" : "decline");
    try {
      if (accept) {
        const g = await ep.invitations.accept(inv.id);
        qc.setQueryData(qk.group(g.id), g);
        if (onAccepted) {
          toast.success(`Welcome to ${g.name}`);
          onAccepted();
        } else {
          toast.success(`You joined ${g.name}`, { action: { label: "Open", onClick: () => navigate(`/app/groups/${g.id}`) } });
        }
      } else {
        await ep.invitations.decline(inv.id);
        toast(`Declined the invitation to ${inv.group.name}`);
      }
      qc.setQueryData<Invitation[]>(qk.invitations, (l) => l?.filter((x) => x.id !== inv.id));
      await Promise.all([
        qc.invalidateQueries({ queryKey: qk.groups }),
        qc.invalidateQueries({ queryKey: qk.invitations }),
        qc.invalidateQueries({ queryKey: qk.counts }),
        qc.invalidateQueries({ queryKey: qk.tasksRoot }),
      ]);
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="flex items-center gap-3 rounded-xl bg-surface p-3 pl-3.5 shadow-card animate-rise max-sm:flex-wrap">
      <Avatar user={inv.inviter} size="lg" />
      <p className="min-w-0 flex-1 text-sm text-fg-2">
        <span className="font-medium text-fg">{inv.inviter.name}</span> invited you to join{" "}
        <span className="font-medium whitespace-nowrap text-fg">
          <GroupDot color={inv.group.color} className="mr-1.5 mb-px align-middle" />
          {inv.group.name}
        </span>
      </p>
      <div className="flex shrink-0 items-center gap-1.5 max-sm:w-full max-sm:justify-end">
        <Button size="sm" variant="ghost" loading={busy === "decline"} disabled={!!busy} onClick={() => answer(false)}>
          Decline
        </Button>
        <Button size="sm" variant="primary" loading={busy === "accept"} disabled={!!busy} onClick={() => answer(true)}>
          Join group
        </Button>
      </div>
    </div>
  );
}
