// Mock mode only: the mails the in-memory backend "sent", with their links,
// so the signup → verify and forgot → reset flows can be clicked through.
import { Mail } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router";
import { useT } from "../i18n";
import { distance } from "../lib/dates";
import type { MMail } from "./db";
import { mockDB } from "./server";

function read(email?: string): MMail[] {
  return mockDB().mails.filter((m) => !email || m.to === email.trim().toLowerCase()).slice(0, 4);
}

export default function MockMailbox({ email }: { email?: string }) {
  const t = useT();
  const [mails, setMails] = useState(() => read(email));
  useEffect(() => {
    const t = setInterval(() => setMails(read(email)), 600);
    return () => clearInterval(t);
  }, [email]);
  return (
    <section className="mt-6 overflow-hidden rounded-xl border border-dashed border-line-strong" aria-label={t("mock.mailboxLabel")}>
      <header className="flex items-center justify-between border-b border-dashed border-line-strong bg-inset/60 px-3.5 py-2">
        <span className="flex items-center gap-2 text-xs font-medium text-fg-2">
          <span className="rounded bg-accent-soft px-1.5 py-0.5 text-2xs font-semibold tracking-wide text-accent-ink uppercase">Mock</span>
          {t("mock.mailbox")}
        </span>
        <span className="text-2xs text-fg-4">vite --mode mock</span>
      </header>
      {mails.length === 0 ? (
        <p className="px-3.5 py-3 text-sm text-fg-3">{t("mock.noMail")}</p>
      ) : (
        <ul className="divide-y divide-line-soft">
          {mails.map((m) => (
            <li key={m.id} className="flex items-center gap-3 px-3.5 py-2.5">
              <Mail className="size-4 shrink-0 text-fg-4" aria-hidden="true" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-fg">{m.subject}</p>
                <p className="truncate text-xs text-fg-3">
                  {t("mock.to", { email: m.to, ago: distance(m.at, t.locale) })}
                </p>
              </div>
              {m.link ? (
                <Link to={m.link} className="shrink-0 rounded-md bg-surface px-2.5 py-1 text-xs font-medium text-fg shadow-card hover:bg-inset">
                  {m.kind === "verify" ? t("mock.verify") : m.kind === "reset" ? t("mock.reset") : t("common.open")}
                </Link>
              ) : null}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
