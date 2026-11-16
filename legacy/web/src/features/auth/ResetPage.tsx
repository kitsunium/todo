import { useQueryClient } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { qk } from "../../api/queries";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Field, PasswordInput } from "../../components/ui/input";
import { toast } from "../../components/ui/toast";
import { setLocale, tr, useT } from "../../i18n";
import { passwordStrength } from "../../lib/password";
import { AuthHeading } from "./AuthLayout";
import { StrengthMeter } from "./forms";

export function ResetPage() {
  const t = useT();
  const [params] = useSearchParams();
  const token = params.get("token") ?? "";
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [busy, setBusy] = useState(false);
  const [failure, setFailure] = useState<unknown>(null);
  const strength = passwordStrength(password, t.locale);
  const server = fieldErrors(failure);
  const mismatch = confirm && confirm !== password ? t("reset.mismatch") : undefined;

  async function submit(e: FormEvent) {
    e.preventDefault();
    setSubmitted(true);
    if (!strength.ok || confirm !== password) return;
    setBusy(true);
    setFailure(null);
    try {
      const user = await ep.auth.reset({ token, password });
      qc.setQueryData(qk.me, user);
      setLocale(user.locale);
      toast.success(tr(user.locale)("reset.done"));
      navigate("/app/today", { replace: true });
    } catch (err) {
      setFailure(err);
    } finally {
      setBusy(false);
    }
  }

  if (!token) {
    return (
      <>
        <AuthHeading title={t("reset.incompleteTitle")}>{t("reset.incompleteBody")}</AuthHeading>
        <Button asChild variant="primary" size="lg" className="w-full">
          <Link to="/forgot">{t("reset.newLink")}</Link>
        </Button>
      </>
    );
  }

  const tokenProblem = failure && !server.password;
  return (
    <>
      <AuthHeading title={t("reset.title")}>{t("reset.subtitle")}</AuthHeading>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field
          label={t("reset.new")}
          error={server.password}
          hint={<StrengthMeter password={password} invalid={submitted && !strength.ok} />}
        >
          {(p) => (
            <PasswordInput {...p} inputSize="lg" autoComplete="new-password" autoFocus value={password} maxLength={128} onChange={(e) => setPassword(e.target.value)} />
          )}
        </Field>
        <Field label={t("reset.confirm")} error={(submitted || confirm.length >= password.length) && mismatch ? mismatch : undefined}>
          {(p) => (
            <PasswordInput {...p} inputSize="lg" autoComplete="new-password" value={confirm} maxLength={128} onChange={(e) => setConfirm(e.target.value)} />
          )}
        </Field>
        {tokenProblem ? (
          <Callout
            tone="danger"
            action={
              <Link to="/forgot" className="text-sm font-medium underline underline-offset-4">
                {t("reset.newLink")}
              </Link>
            }
          >
            {errorMessage(failure)}
          </Callout>
        ) : null}
        <Button type="submit" variant="primary" size="lg" loading={busy} className="mt-1 w-full">
          {t("reset.submit")}
        </Button>
      </form>
    </>
  );
}
