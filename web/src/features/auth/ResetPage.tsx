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
import { passwordStrength } from "../../lib/password";
import { AuthHeading } from "./AuthLayout";
import { StrengthMeter } from "./forms";

export function ResetPage() {
  const [params] = useSearchParams();
  const token = params.get("token") ?? "";
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [busy, setBusy] = useState(false);
  const [failure, setFailure] = useState<unknown>(null);
  const strength = passwordStrength(password);
  const server = fieldErrors(failure);
  const mismatch = confirm && confirm !== password ? "The passwords don’t match." : undefined;

  async function submit(e: FormEvent) {
    e.preventDefault();
    setSubmitted(true);
    if (!strength.ok || confirm !== password) return;
    setBusy(true);
    setFailure(null);
    try {
      const user = await ep.auth.reset({ token, password });
      qc.setQueryData(qk.me, user);
      toast.success("Password updated. You’re signed in.");
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
        <AuthHeading title="This link is incomplete">Open the link from the email again, or ask for a new one.</AuthHeading>
        <Button asChild variant="primary" size="lg" className="w-full">
          <Link to="/forgot">Get a new link</Link>
        </Button>
      </>
    );
  }

  const tokenProblem = failure && !server.password;
  return (
    <>
      <AuthHeading title="Choose a new password">Use at least 10 characters. Every other device will be signed out.</AuthHeading>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field
          label="New password"
          error={server.password}
          hint={<StrengthMeter password={password} invalid={submitted && !strength.ok} />}
        >
          {(p) => (
            <PasswordInput {...p} inputSize="lg" autoComplete="new-password" autoFocus value={password} maxLength={128} onChange={(e) => setPassword(e.target.value)} />
          )}
        </Field>
        <Field label="Confirm password" error={(submitted || confirm.length >= password.length) && mismatch ? mismatch : undefined}>
          {(p) => (
            <PasswordInput {...p} inputSize="lg" autoComplete="new-password" value={confirm} maxLength={128} onChange={(e) => setConfirm(e.target.value)} />
          )}
        </Field>
        {tokenProblem ? (
          <Callout
            tone="danger"
            action={
              <Link to="/forgot" className="text-sm font-medium underline underline-offset-4">
                Get a new link
              </Link>
            }
          >
            {errorMessage(failure)}
          </Callout>
        ) : null}
        <Button type="submit" variant="primary" size="lg" loading={busy} className="mt-1 w-full">
          Update password and sign in
        </Button>
      </form>
    </>
  );
}
