import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Field, Input, PasswordInput } from "../../components/ui/input";
import { passwordStrength } from "../../lib/password";
import { AuthHeading } from "./AuthLayout";
import { checkEmail, StrengthMeter, useTouched } from "./forms";

export function SignupPage() {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [name, setName] = useState("");
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [failure, setFailure] = useState<unknown>(null);
  const t = useTouched<"name" | "email" | "password">();
  const server = fieldErrors(failure);
  const strength = passwordStrength(password);
  const errors = {
    name: !name.trim() ? "Tell us what to call you." : name.trim().length > 80 ? "Keep it under 80 characters." : server.name,
    email: checkEmail(email) ?? server.email,
    password: !password ? "Choose a password." : server.password,
  };

  async function submit(e: FormEvent) {
    e.preventDefault();
    t.submit();
    if (errors.name || checkEmail(email) || !strength.ok) return;
    setBusy(true);
    setFailure(null);
    try {
      const r = await ep.auth.signup({ name: name.trim(), email: email.trim().toLowerCase(), password });
      navigate(`/check-inbox?email=${encodeURIComponent(r.email ?? email.trim().toLowerCase())}`);
    } catch (err) {
      setFailure(err);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <AuthHeading title="Create your account">A calm place for your tasks — and the people you share them with.</AuthHeading>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field label="Name" error={t.shown("name") ? errors.name : undefined}>
          {(p) => (
            <Input
              {...p}
              inputSize="lg"
              name="name"
              autoComplete="name"
              autoFocus
              placeholder="Camille Martin"
              value={name}
              maxLength={80}
              onChange={(e) => setName(e.target.value)}
              onBlur={() => name && t.touch("name")}
            />
          )}
        </Field>
        <Field label="Email" error={t.shown("email") ? errors.email : undefined}>
          {(p) => (
            <Input
              {...p}
              inputSize="lg"
              type="email"
              name="email"
              autoComplete="email"
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
          hint={<StrengthMeter password={password} invalid={t.shown("password") && !strength.ok} />}
        >
          {(p) => (
            <PasswordInput
              {...p}
              inputSize="lg"
              name="password"
              autoComplete="new-password"
              value={password}
              maxLength={128}
              onChange={(e) => setPassword(e.target.value)}
              onBlur={() => password && t.touch("password")}
            />
          )}
        </Field>
        {failure && !Object.keys(server).length ? <Callout tone="danger">{errorMessage(failure)}</Callout> : null}
        <Button type="submit" variant="primary" size="lg" loading={busy} className="mt-1 w-full">
          Create account
        </Button>
        <p className="text-center text-xs text-fg-4">We’ll email you a link to confirm your address.</p>
      </form>
    </>
  );
}
