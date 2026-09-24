import { useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { MailIllo } from "../../components/brand/illustrations";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Field, Input } from "../../components/ui/input";
import { useT } from "../../i18n";
import { AuthHeading } from "./AuthLayout";
import { DevMailbox } from "./DevMailbox";
import { checkEmail } from "./forms";

export function ForgotPage() {
  const t = useT();
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
        <h1 className="text-2xl font-semibold tracking-[-0.025em] text-fg">{t("inbox.title")}</h1>
        <p className="mt-2 max-w-[360px] text-[15px] leading-[22px] text-fg-3">
          {t.rich("forgot.sentBody", {
            email: <span className="my-1 block font-medium [overflow-wrap:anywhere] text-fg">{sent}</span>,
          })}
        </p>
        <p className="mt-6 text-sm text-fg-3">
          <Link to={`/login?email=${encodeURIComponent(sent)}`} className="font-medium text-fg underline-offset-4 hover:underline">
            {t("forgot.back")}
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
      <AuthHeading title={t("forgot.title")}>{t("forgot.subtitle")}</AuthHeading>
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field label={t("common.email")} error={touched ? invalid : undefined}>
          {(p) => (
            <Input
              {...p}
              inputSize="lg"
              type="email"
              autoComplete="email"
              autoFocus
              placeholder={t("common.emailPlaceholder")}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              onBlur={() => email && setTouched(true)}
            />
          )}
        </Field>
        {failure && !fieldErrors(failure).email ? <Callout tone="danger">{errorMessage(failure)}</Callout> : null}
        <Button type="submit" variant="primary" size="lg" loading={busy} className="mt-1 w-full">
          {t("forgot.submit")}
        </Button>
      </form>
      <p className="mt-6 text-center text-sm text-fg-3">
        {t("forgot.remembered")}{" "}
        <Link to="/login" className="font-medium text-fg underline-offset-4 hover:underline">
          {t("auth.signIn")}
        </Link>
      </p>
    </>
  );
}
