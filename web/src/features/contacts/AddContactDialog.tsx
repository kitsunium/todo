import { useQueryClient } from "@tanstack/react-query";
import { MailPlus, UserRoundPlus } from "lucide-react";
import { useEffect, useState, type FormEvent } from "react";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { qk } from "../../api/queries";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Dialog, DialogBody, DialogContent, DialogFooter, DialogHeader } from "../../components/ui/dialog";
import { Field, Input } from "../../components/ui/input";
import { toast } from "../../components/ui/toast";
import { checkEmail } from "../auth/forms";

export function AddContactDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (o: boolean) => void }) {
  const qc = useQueryClient();
  const [email, setEmail] = useState("");
  const [touched, setTouched] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);

  useEffect(() => {
    if (open) {
      setEmail("");
      setTouched(false);
      setError(null);
    }
  }, [open]);

  const invalid = checkEmail(email);
  async function submit(e: FormEvent) {
    e.preventDefault();
    setTouched(true);
    if (invalid) return;
    setBusy(true);
    setError(null);
    try {
      const r = await ep.contacts.add(email.trim().toLowerCase());
      if (r.kind === "request") toast.success(`Request sent to ${r.request.user.name}`, { description: "They’ll show up in your contacts once they accept." });
      else toast.success(`Invitation emailed to ${r.invite.email}`, { description: "You’ll be connected as soon as they join." });
      await qc.invalidateQueries({ queryKey: qk.contacts });
      onOpenChange(false);
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  }
  const server = fieldErrors(error).email;
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <form onSubmit={submit} noValidate>
          <DialogHeader
            icon={
              <span className="flex size-9 shrink-0 items-center justify-center rounded-[10px] bg-accent-soft text-accent-ink">
                <UserRoundPlus className="size-[18px]" />
              </span>
            }
            title="Add a contact"
            description="Contacts can share tasks with each other and join the same groups."
          />
          <DialogBody className="flex flex-col gap-4">
            <Field label="Email address" error={(touched && invalid) || server}>
              {(p) => (
                <Input {...p} inputSize="lg" type="email" autoFocus autoComplete="off" placeholder="sam@example.com" value={email} onChange={(e) => setEmail(e.target.value)} />
              )}
            </Field>
            <div className="grid gap-2 text-xs text-fg-3 sm:grid-cols-2">
              <div className="flex gap-2 rounded-lg bg-inset p-2.5">
                <UserRoundPlus className="mt-px size-3.5 shrink-0 text-fg-4" />
                <p><span className="font-medium text-fg-2">Already on Todo?</span> They get a request to accept.</p>
              </div>
              <div className="flex gap-2 rounded-lg bg-inset p-2.5">
                <MailPlus className="mt-px size-3.5 shrink-0 text-fg-4" />
                <p><span className="font-medium text-fg-2">Not yet?</span> We email them an invitation to join.</p>
              </div>
            </div>
            {error && !server ? <Callout tone="danger">{errorMessage(error)}</Callout> : null}
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" loading={busy}>
              Send
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
