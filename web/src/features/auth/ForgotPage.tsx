import { useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { MailIllo } from "../../components/brand/illustrations";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Field, Input } from "../../components/ui/input";
import { AuthHeading } from "./AuthLayout";
import { DevMailbox } from "./DevMailbox";
import { checkEmail } from "./forms";

export function ForgotPage() {
  const [params] = useSearchParams();
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [touched, setTouched] = useState(false);
  const [busy, setBusy] = useState(false);
  const [sent, setSent] = useState("");
  const [failure, setFailure] = useState<unknown>(null);
  const invalid = checkEmail(email) ?? fieldErrors(failure).email;

  async function submit(e: FormEvent) {
    e.preventDefault();
    setTouched(true);
    if (checkEmail(email)) return;
    setBusy(true);
    setFailure(null);
    try {
      const address = email.trim().toLowerCase();
      await ep.auth.forgot(address);
      setSent(address);
    } catch (err) {
      setFailure(err);
    } finally {
      setBusy(false);
    }
  }

  if (sent) {
    return (
      <div className="flex flex-col items-center text-center">
        <MailIllo className="-mt-4 mb-2" />
        <h1 className="text-2xl font-semibold tracking-[-0.025em] text-fg">Check your inbox</h1>
        <p className="mt-2 max-w-[360px] text-[15px] leading-[22px] text-fg-3">
          If an account exists for
          <span className="my-1 block font-medium [overflow-wrap:anywhere] text-fg">{sent}</span>
          a link to reset your password is on its way. It expires in one hour.
        </p>
        <p className="mt-6 text-sm text-fg-3">
          <Link to={`/login?email=${encodeURIComponent(sent)}`} className="font-medium text-fg underline-offset-4 hover:underline">
            Back to sign in
          </Link>
        </p>
        <div className="w-full text-left">
          <DevMailbox email={sent} />
        </div>
      </div>
    );
  }

  return (
    <>
      <AuthHeading title="Reset your password">Enter the email you signed up with. We’ll send you a link to choose a new password.</AuthHeading>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field label="Email" error={touched ? invalid : undefined}>
          {(p) => (
            <Input
              {...p}
              inputSize="lg"
              type="email"
              autoComplete="email"
              autoFocus
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onBlur={() => email && setTouched(true)}
            />
          )}
        </Field>
        {failure && !fieldErrors(failure).email ? <Callout tone="danger">{errorMessage(failure)}</Callout> : null}
        <Button type="submit" variant="primary" size="lg" loading={busy} className="mt-1 w-full">
          Send the reset link
        </Button>
      </form>
      <p className="mt-6 text-center text-sm text-fg-3">
        Remembered it?{" "}
        <Link to="/login" className="font-medium text-fg underline-offset-4 hover:underline">
          Sign in
        </Link>
      </p>
    </>
  );
}
