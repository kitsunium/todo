import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage } from "../../api/errors";
import { qk } from "../../api/queries";
import type { User } from "../../api/types";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Field, Input } from "../../components/ui/input";
import { Spinner } from "../../components/ui/spinner";
import { toast } from "../../components/ui/toast";
import { AuthHeading } from "./AuthLayout";
import { checkEmail } from "./forms";

// A token is single use: StrictMode's double effect must not spend it twice.
const inflight = new Map<string, Promise<User>>();
function verifyOnce(token: string): Promise<User> {
  let p = inflight.get(token);
  if (!p) {
    p = ep.auth.verify(token);
    inflight.set(token, p);
  }
  return p;
}

export function VerifyPage() {
  const [params] = useSearchParams();
  const token = params.get("token") ?? "";
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [error, setError] = useState<unknown>(token ? null : new Error("missing"));

  useEffect(() => {
    if (!token) return;
    let live = true;
    verifyOnce(token).then(
      (user) => {
        if (!live) return;
        qc.setQueryData(qk.me, user);
        toast.success(`Email confirmed — welcome to Todo, ${user.name.split(" ")[0]}.`);
        navigate("/app/today", { replace: true });
      },
      (err: unknown) => live && setError(err),
    );
    return () => {
      live = false;
    };
  }, [token, navigate, qc]);

  if (!error) {
    return (
      <div className="flex flex-col items-center py-10 text-center" role="status">
        <Spinner className="size-6 text-accent" />
        <p className="mt-4 text-[15px] font-medium text-fg">Confirming your email…</p>
      </div>
    );
  }
  return <Expired reason={token ? errorMessage(error) : "This link is incomplete."} />;
}

function Expired({ reason }: { reason: string }) {
  const [email, setEmail] = useState("");
  const [busy, setBusy] = useState(false);
  const [touched, setTouched] = useState(false);
  const navigate = useNavigate();
  const invalid = checkEmail(email);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setTouched(true);
    if (invalid) return;
    setBusy(true);
    try {
      await ep.auth.resend(email.trim().toLowerCase());
      navigate(`/check-inbox?email=${encodeURIComponent(email.trim().toLowerCase())}`);
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <AuthHeading title="This link didn’t work">It may have expired or already been used. Get a fresh one below.</AuthHeading>
      <Callout tone="warn" className="mb-5">
        {reason}
      </Callout>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field label="Email" error={touched ? invalid : undefined}>
          {(p) => (
            <Input {...p} inputSize="lg" type="email" autoComplete="email" placeholder="you@example.com" value={email} onChange={(e) => setEmail(e.target.value)} />
          )}
        </Field>
        <Button type="submit" variant="primary" size="lg" loading={busy} className="w-full">
          Send a new link
        </Button>
      </form>
      <p className="mt-6 text-center text-sm text-fg-3">
        Already confirmed?{" "}
        <Link to="/login" className="font-medium text-fg underline-offset-4 hover:underline">
          Sign in
        </Link>
      </p>
    </>
  );
}
