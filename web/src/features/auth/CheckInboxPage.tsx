import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router";
import * as ep from "../../api/endpoints";
import { errorMessage } from "../../api/errors";
import { MailIllo } from "../../components/brand/illustrations";
import { Button } from "../../components/ui/button";
import { toast } from "../../components/ui/toast";
import { tr, useT } from "../../i18n";
import { DevMailbox } from "./DevMailbox";

const COOLDOWN = 30;

export function CheckInboxPage() {
  const t = useT();
  const [params] = useSearchParams();
  const email = params.get("email") ?? "";
  const [left, setLeft] = useState(0);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (left <= 0) return;
    const timer = setTimeout(() => setLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [left]);

  async function resend() {
    setBusy(true);
    try {
      await ep.auth.resend(email);
      toast.success(tr()("inbox.sent"));
      setLeft(COOLDOWN);
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col items-center text-center">
      <MailIllo className="-mt-4 mb-2" />
      <h1 className="text-2xl font-semibold tracking-[-0.025em] text-fg">{t("inbox.title")}</h1>
      <p className="mt-2 max-w-[360px] text-[15px] leading-[22px] text-fg-3">
        {t.rich("inbox.body", {
          email: <span className="my-1 block font-medium [overflow-wrap:anywhere] text-fg">{email || t("inbox.yourAddress")}</span>,
        })}
      </p>
      <div className="mt-7 flex w-full flex-col gap-2">
        {email ? (
          <Button variant="secondary" size="lg" className="w-full" loading={busy} disabled={left > 0} onClick={resend}>
            {left > 0 ? t("inbox.resendIn", { seconds: left }) : t("inbox.resend")}
          </Button>
        ) : null}
        <p className="mt-3 text-sm text-fg-3">
          <span className="whitespace-nowrap">
            {t("inbox.wrongAddress")}{" "}
            <Link to="/signup" className="font-medium text-fg underline-offset-4 hover:underline">
              {t("inbox.startOver")}
            </Link>
          </span>{" "}
          ·{" "}
          <span className="whitespace-nowrap">
            {t("inbox.alreadyConfirmed")}{" "}
            <Link to={`/login${email ? `?email=${encodeURIComponent(email)}` : ""}`} className="font-medium text-fg underline-offset-4 hover:underline">
              {t("auth.signIn")}
            </Link>
          </span>
        </p>
      </div>
      <div className="w-full text-left">
        <DevMailbox {...(email ? { email } : {})} />
      </div>
    </div>
  );
}
