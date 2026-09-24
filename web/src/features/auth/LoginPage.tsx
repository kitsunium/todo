import { useQueryClient } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import { ApiError } from "../../api/client";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { qk } from "../../api/queries";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Field, Input, PasswordInput } from "../../components/ui/input";
import { toast } from "../../components/ui/toast";
import { safeNext } from "../../lib/nav";
import { AuthHeading } from "./AuthLayout";
import { checkEmail, useTouched } from "./forms";

export function LoginPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [failure, setFailure] = useState<ApiError | Error | null>(null);
  const [resending, setResending] = useState(false);
  const t = useTouched<"email" | "password">();
  const server = fieldErrors(failure);
  const errors = {
    email: checkEmail(email) ?? server.email,
    password: password ? server.password : "Enter your password.",
  };

  async function submit(e: FormEvent) {
    e.preventDefault();
    t.submit();
    if (checkEmail(email) || !password) return;
    setBusy(true);
    setFailure(null);
    try {
      const user = await ep.auth.login({ email: email.trim().toLowerCase(), password });
      qc.setQueryData(qk.me, user);
      navigate(safeNext(params.get("next")), { replace: true });
    } catch (err) {
      setFailure(err instanceof Error ? err : new Error(String(err)));
    } finally {
      setBusy(false);
    }
  }

  async function resend() {
    setResending(true);
    try {
      await ep.auth.resend(email.trim().toLowerCase());
      navigate(`/check-inbox?email=${encodeURIComponent(email.trim().toLowerCase())}`);
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setResending(false);
    }
  }

  const code = failure instanceof ApiError ? failure.code : "";
  return (
    <>
      <AuthHeading title="Welcome back">Sign in to pick up where you left off.</AuthHeading>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field label="Email" error={t.shown("email") ? errors.email : undefined}>
          {(p) => (
            <Input
              {...p}
              inputSize="lg"
              type="email"
              name="email"
              autoComplete="email"
              autoFocus={!email}
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onBlur={() => email && t.touch("email")}
            />
          )}
        </Field>
        <Field
          label="Password"
          error={t.shown("password") ? errors.password : undefined}
          aside={
            <Link
              to={`/forgot${email ? `?email=${encodeURIComponent(email.trim())}` : ""}`}
              className="text-xs font-medium text-fg-3 hover:text-fg focus-visible:outline-2 focus-visible:outline-ring"
            >
              Forgot password?
            </Link>
          }
        >
          {(p) => (
            <PasswordInput
              {...p}
              inputSize="lg"
              name="password"
              autoComplete="current-password"
              autoFocus={!!email}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          )}
        </Field>

        {failure && code === "email_unverified" ? (
          <Callout
            tone="warn"
            title="Confirm your email first"
            action={
              <Button size="sm" variant="secondary" loading={resending} onClick={resend}>
                Send a new link
              </Button>
            }
          >
            We sent a confirmation link to {email.trim()} when you signed up.
          </Callout>
        ) : failure && code === "account_locked" ? (
          <Callout
            tone="danger"
            title="Account locked for 15 minutes"
            action={
              <Link
                to={`/forgot?email=${encodeURIComponent(email.trim())}`}
                className="text-sm font-medium underline underline-offset-4"
              >
                Reset your password now
              </Link>
            }
          >
            Too many attempts with the wrong password.
          </Callout>
        ) : failure && !Object.keys(server).length ? (
          <Callout tone="danger">{errorMessage(failure)}</Callout>
        ) : null}

        <Button type="submit" variant="primary" size="lg" loading={busy} className="mt-1 w-full">
          Sign in
        </Button>
      </form>
      <MockHint onFill={(e, p) => (setEmail(e), setPassword(p))} />
    </>
  );
}

/** In mock mode, the demo account — one click to fill the form. */
function MockHint({ onFill }: { onFill: (email: string, password: string) => void }) {
  if (import.meta.env.MODE !== "mock") return null;
  return (
    <button
      type="button"
      onClick={() => onFill("camille@example.com", "correct-horse-42")}
      className="mt-6 w-full rounded-lg border border-dashed border-line-strong px-3.5 py-2.5 text-left text-xs text-fg-3 transition-colors hover:border-accent hover:text-fg-2"
    >
      <span className="mr-2 rounded bg-accent-soft px-1.5 py-0.5 text-2xs font-semibold tracking-wide text-accent-ink uppercase">Mock</span>
      camille@example.com · correct-horse-42 — also try unverified@ or locked@example.com
    </button>
  );
}
