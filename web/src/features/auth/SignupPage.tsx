import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Field, Input, PasswordInput } from "../../components/ui/input";
import { useT } from "../../i18n";
import { passwordStrength } from "../../lib/password";
import { AuthHeading } from "./AuthLayout";
import { checkEmail, StrengthMeter, useTouched } from "./forms";

export function SignupPage() {
  const t = useT();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [name, setName] = useState("");
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [failure, setFailure] = useState<unknown>(null);
  const touched = useTouched<"name" | "email" | "password">();
  const server = fieldErrors(failure);
  const strength = passwordStrength(password, t.locale);
  const errors = {
    name: !name.trim() ? t("signup.nameEmpty") : name.trim().length > 80 ? t("signup.nameTooLong") : server.name,
    email: checkEmail(email) ?? server.email,
    password: !password ? t("signup.passwordEmpty") : server.password,
  };

  async function submit(e: FormEvent) {
    e.preventDefault();
    touched.submit();
    if (errors.name || checkEmail(email) || !strength.ok) return;
    setBusy(true);
    setFailure(null);
    try {
      // The account starts in the language of the page it was created from.
      const r = await ep.auth.signup({ name: name.trim(), email: email.trim().toLowerCase(), password, locale: t.locale });
      navigate(`/check-inbox?email=${encodeURIComponent(r.email ?? email.trim().toLowerCase())}`);
    } catch (err) {
      setFailure(err);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <AuthHeading title={t("signup.title")}>{t("signup.subtitle")}</AuthHeading>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field label={t("common.name")} error={touched.shown("name") ? errors.name : undefined}>
          {(p) => (
            <Input
              {...p}
              inputSize="lg"
              name="name"
              autoComplete="name"
              autoFocus
              placeholder={t("signup.namePlaceholder")}
              value={name}
              maxLength={80}
              onChange={(e) => setName(e.target.value)}
              onBlur={() => name && touched.touch("name")}
            />
          )}
        </Field>
        <Field label={t("common.email")} error={touched.shown("email") ? errors.email : undefined}>
          {(p) => (
            <Input
              {...p}
              inputSize="lg"
              type="email"
              name="email"
              autoComplete="email"
              placeholder={t("common.emailPlaceholder")}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onBlur={() => email && touched.touch("email")}
            />
          )}
        </Field>
        <Field
          label={t("common.password")}
          error={touched.shown("password") ? errors.password : undefined}
          hint={<StrengthMeter password={password} invalid={touched.shown("password") && !strength.ok} />}
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
              onBlur={() => password && touched.touch("password")}
            />
          )}
        </Field>
        {failure && !Object.keys(server).length ? <Callout tone="danger">{errorMessage(failure)}</Callout> : null}
        <Button type="submit" variant="primary" size="lg" loading={busy} className="mt-1 w-full">
          {t("signup.submit")}
        </Button>
        <p className="text-center text-xs text-fg-4">{t("signup.footnote")}</p>
      </form>
    </>
  );
}
