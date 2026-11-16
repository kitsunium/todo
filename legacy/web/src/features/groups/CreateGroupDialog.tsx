import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { qk } from "../../api/queries";
import type { Group } from "../../api/types";
import { Button } from "../../components/ui/button";
import { Dialog, DialogBody, DialogContent, DialogFooter, DialogHeader } from "../../components/ui/dialog";
import { GroupBadge } from "../../components/ui/group-color";
import { Field, Input } from "../../components/ui/input";
import { toast } from "../../components/ui/toast";
import { tr, useT } from "../../i18n";
import { ColorPicker } from "./ColorPicker";

const PICKS = ["orange", "blue", "green", "violet", "pink", "teal", "indigo", "amber", "red", "slate"];

/** Create a group, or edit one (name and color) when `group` is given. */
export function CreateGroupDialog({
  open,
  onOpenChange,
  group,
}: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  group?: Group;
}) {
  const t = useT();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [color, setColor] = useState("orange");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const [touched, setTouched] = useState(false);

  useEffect(() => {
    if (!open) return;
    setName(group?.name ?? "");
    setColor(group?.color ?? PICKS[Math.floor(Math.random() * 5)]!);
    setError(null);
    setTouched(false);
  }, [open, group]);

  const invalid = !name.trim() ? t("groupForm.nameEmpty") : undefined;

  async function submit(e: FormEvent) {
    e.preventDefault();
    setTouched(true);
    if (invalid) return;
    setBusy(true);
    setError(null);
    try {
      if (group) {
        const g = await ep.groups.update(group.id, { name: name.trim(), color });
        qc.setQueryData(qk.group(g.id), g);
        toast.success(tr()("groupForm.updated"));
      } else {
        const g = await ep.groups.create({ name: name.trim(), color });
        qc.setQueryData(qk.group(g.id), g);
        toast.success(tr()("groupForm.ready", { name: g.name }), { description: tr()("groupForm.readyHint") });
        navigate(`/app/groups/${g.id}`);
      }
      await Promise.all([qc.invalidateQueries({ queryKey: qk.groups }), qc.invalidateQueries({ queryKey: qk.tasksRoot })]);
      onOpenChange(false);
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  }

  const server = fieldErrors(error);
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <form onSubmit={submit} noValidate>
          <DialogHeader
            icon={<GroupBadge name={name || "?"} color={color} size="lg" />}
            title={group ? t("groupForm.editTitle") : t("groupForm.newTitle")}
            description={group ? t("groupForm.editBody") : t("groupForm.newBody")}
          />
          <DialogBody className="flex flex-col gap-5">
            <Field label={t("common.name")} error={(touched && invalid) || server.name}>
              {(p) => (
                <Input {...p} inputSize="lg" autoFocus placeholder={t("groupForm.namePlaceholder")} value={name} maxLength={60} onChange={(e) => setName(e.target.value)} />
              )}
            </Field>
            <div className="flex flex-col gap-2">
              <span className="text-sm font-medium text-fg">{t("groupForm.color")}</span>
              <ColorPicker value={color} onChange={setColor} />
            </div>
            {error && !server.name ? <p className="text-sm text-danger-ink">{errorMessage(error)}</p> : null}
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => onOpenChange(false)}>
              {t("common.cancel")}
            </Button>
            <Button type="submit" variant="primary" loading={busy}>
              {group ? t("common.save") : t("groupForm.create")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
