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
import { tr, useT } from "../../i18n";
import { checkEmail } from "../auth/forms";

export function AddContactDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (o: boolean) => void }) {
  const t = useT();
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
      const say = tr();
      if (r.kind === "request") {
        toast.success(say("addContact.requestSent", { name: r.request.user.name }), { description: say("addContact.requestHint") });
      } else {
        toast.success(say("addContact.inviteSent", { email: r.invite.email }), { description: say("addContact.inviteHint") });
      }
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
            title={t("palette.addContact")}
            description={t("addContact.body")}
          />
          <DialogBody className="flex flex-col gap-4">
            <Field label={t("addContact.email")} error={(touched && invalid) || server}>
              {(p) => (
                <Input {...p} inputSize="lg" type="email" autoFocus autoComplete="off" placeholder={t("addContact.placeholder")} value={email} onChange={(e) => setEmail(e.target.value)} />
              )}
            </Field>
            <div className="grid gap-2 text-xs text-fg-3 sm:grid-cols-2">
              <div className="flex gap-2 rounded-lg bg-inset p-2.5">
                <UserRoundPlus className="mt-px size-3.5 shrink-0 text-fg-4" />
                <p>
                  <span className="font-medium text-fg-2">{t("addContact.onTodo")}</span> {t("addContact.onTodoBody")}
                </p>
              </div>
              <div className="flex gap-2 rounded-lg bg-inset p-2.5">
                <MailPlus className="mt-px size-3.5 shrink-0 text-fg-4" />
                <p>
                  <span className="font-medium text-fg-2">{t("addContact.notYet")}</span> {t("addContact.notYetBody")}
                </p>
              </div>
            </div>
            {error && !server ? <Callout tone="danger">{errorMessage(error)}</Callout> : null}
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => onOpenChange(false)}>
              {t("common.cancel")}
            </Button>
            <Button type="submit" variant="primary" loading={busy}>
              {t("common.send")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
