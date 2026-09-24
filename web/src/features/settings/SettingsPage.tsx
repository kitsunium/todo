import { useQueryClient } from "@tanstack/react-query";
import { Laptop, Settings, Smartphone } from "lucide-react";
import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import * as ep from "../../api/endpoints";
import { errorMessage, fieldErrors } from "../../api/errors";
import { qk, useCurrentUser, useSessions } from "../../api/queries";
import type { Session } from "../../api/types";
import { Avatar } from "../../components/ui/avatar";
import { Button } from "../../components/ui/button";
import { Callout } from "../../components/ui/callout";
import { Chip } from "../../components/ui/chip";
import { ConfirmDialog } from "../../components/ui/dialog";
import { Field, Input, PasswordInput } from "../../components/ui/input";
import { Skeleton } from "../../components/ui/skeleton";
import { toast } from "../../components/ui/toast";
import { tr, useT, type Locale } from "../../i18n";
import { cn } from "../../lib/cn";
import { distance, formatDate } from "../../lib/dates";
import { passwordStrength } from "../../lib/password";
import { StrengthMeter } from "../auth/forms";
import { Page } from "../shell/Page";
import { LanguageCards } from "./language";

function Section({ title, description, children }: { title: string; description?: string; children: ReactNode }) {
  return (
    <section className="grid gap-4 border-t border-line-soft py-8 first:border-t-0 first:pt-2 md:grid-cols-[220px_minmax(0,1fr)] md:gap-8">
      <div>
        <h2 className="text-[15px] font-semibold tracking-[-0.01em] text-fg">{title}</h2>
        {description ? <p className="mt-1 text-sm text-fg-3">{description}</p> : null}
      </div>
      <div className="min-w-0">{children}</div>
    </section>
  );
}

export function SettingsPage() {
  const t = useT();
  return (
    <Page title={t("settings.title")} icon={<Settings className="text-fg-2" strokeWidth={2} />} subtitle={t("settings.subtitle")} wide>
      <Section title={t("settings.profile")} description={t("settings.profileHint")}>
        <Profile />
      </Section>
      <Section title={t("settings.language")} description={t("settings.languageHint")}>
        <LanguageCards />
      </Section>
      <Section title={t("settings.password")} description={t("settings.passwordHint")}>
        <Password />
      </Section>
      <Section title={t("settings.sessions")} description={t("settings.sessionsHint")}>
        <Sessions />
      </Section>
    </Page>
  );
}

function Profile() {
  const t = useT();
  const me = useCurrentUser();
  const qc = useQueryClient();
  const [name, setName] = useState(me.name);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  useEffect(() => setName(me.name), [me.name]);
  const dirty = name.trim() !== me.name && name.trim().length > 0;

  async function save(e: FormEvent) {
    e.preventDefault();
    if (!dirty) return;
    setBusy(true);
    setError(null);
    try {
      const u = await ep.auth.updateMe({ name: name.trim() });
      qc.setQueryData(qk.me, u);
      toast.success(tr()("settings.profileSaved"));
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={save} className="rounded-xl bg-surface p-5 shadow-card">
      <div className="mb-5 flex items-center gap-3">
        <Avatar user={{ ...me, name: name || me.name }} size="xl" />
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-fg">{name || me.name}</p>
          <p className="truncate text-xs text-fg-3">
            {t("settings.memberSince", { date: formatDate(new Date(me.createdAt), "date.pattern.monthYear", t.locale) })}
          </p>
        </div>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label={t("common.name")} error={!name.trim() ? t("settings.nameEmpty") : fieldErrors(error, t.locale).name}>
          {(p) => <Input {...p} value={name} maxLength={80} autoComplete="name" onChange={(e) => setName(e.target.value)} />}
        </Field>
        <Field label={t("common.email")} hint={t("settings.emailHint")}>
          {(p) => <Input {...p} value={me.email} readOnly className="text-fg-3" />}
        </Field>
      </div>
      {error && !fieldErrors(error).name ? (
        <Callout tone="danger" className="mt-4">
          {errorMessage(error)}
        </Callout>
      ) : null}
      <div className="mt-5 flex justify-end gap-2">
        {dirty ? (
          <Button variant="ghost" onClick={() => setName(me.name)}>
            {t("settings.discard")}
          </Button>
        ) : null}
        <Button type="submit" variant="primary" disabled={!dirty} loading={busy}>
          {t("common.save")}
        </Button>
      </div>
    </form>
  );
}

function Password() {
  const t = useT();
  const qc = useQueryClient();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const strength = passwordStrength(next, t.locale);
  const server = fieldErrors(error);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setSubmitted(true);
    if (!current || !strength.ok) return;
    setBusy(true);
    setError(null);
    try {
      await ep.auth.changePassword({ current, password: next });
      setCurrent("");
      setNext("");
      setSubmitted(false);
      toast.success(tr()("settings.passwordChanged"), { description: tr()("settings.passwordChangedHint") });
      await qc.invalidateQueries({ queryKey: qk.sessions });
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  }
  const wrongCurrent = error && (error as { code?: string }).code === "invalid_credentials";
  return (
    <form onSubmit={submit} noValidate className="rounded-xl bg-surface p-5 shadow-card">
      <div className="grid gap-4 sm:grid-cols-2">
        <Field
          label={t("settings.currentPassword")}
          error={wrongCurrent ? t("settings.wrongCurrent") : (server.current ?? (submitted && !current ? t("settings.currentEmpty") : undefined))}
        >
          {(p) => <PasswordInput {...p} value={current} autoComplete="current-password" onChange={(e) => setCurrent(e.target.value)} />}
        </Field>
        <Field label={t("settings.newPassword")} error={server.password} hint={<StrengthMeter password={next} invalid={submitted && !strength.ok} />}>
          {(p) => <PasswordInput {...p} value={next} maxLength={128} autoComplete="new-password" onChange={(e) => setNext(e.target.value)} />}
        </Field>
      </div>
      {error && !wrongCurrent && !Object.keys(server).length ? <Callout tone="danger" className="mt-4">{errorMessage(error)}</Callout> : null}
      <div className="mt-5 flex justify-end">
        <Button type="submit" variant="primary" loading={busy} disabled={!current && !next}>
          {t("settings.changePassword")}
        </Button>
      </div>
    </form>
  );
}

/** "Chrome on macOS", "Chrome sur macOS" from a user agent: enough to recognize a device. */
export function describeAgent(ua: string, loc: Locale): { label: string; mobile: boolean } {
  const t = tr(loc);
  const browser = /Edg\//.test(ua)
    ? "Edge"
    : /OPR\//.test(ua)
      ? "Opera"
      : /Firefox\//.test(ua)
        ? "Firefox"
        : /Chrome\//.test(ua)
          ? "Chrome"
          : /Safari\//.test(ua)
            ? "Safari"
            : /curl|Go-http-client|okhttp|python/i.test(ua)
              ? t("session.apiClient")
              : t("session.browser");
  const os = /iPhone|iPad|iPod/.test(ua)
    ? "iOS"
    : /Android/.test(ua)
      ? "Android"
      : /Mac OS X|Macintosh/.test(ua)
        ? "macOS"
        : /Windows/.test(ua)
          ? "Windows"
          : /Linux|X11|Ubuntu/.test(ua)
            ? "Linux"
            : "";
  return { label: os ? t("session.on", { browser, os }) : browser, mobile: /iPhone|Android.*Mobile|Mobile/.test(ua) };
}

function Sessions() {
  const q = useSessions();
  if (q.isPending) {
    return (
      <div className="rounded-xl bg-surface shadow-card">
        {[0, 1, 2].map((i) => (
          <div key={i} className="flex items-center gap-3 border-t border-line-soft px-4 py-3.5 first:border-t-0">
            <Skeleton className="size-9 rounded-lg" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-3 w-36" />
              <Skeleton className="h-2.5 w-52" />
            </div>
          </div>
        ))}
      </div>
    );
  }
  if (q.isError) return <Callout tone="danger">{errorMessage(q.error)}</Callout>;
  return (
    <ul className="overflow-hidden rounded-xl bg-surface shadow-card [&>li+li]:border-t [&>li+li]:border-line-soft">
      {q.data.map((s) => (
        <SessionRow key={s.id} s={s} />
      ))}
    </ul>
  );
}

function SessionRow({ s }: { s: Session }) {
  const t = useT();
  const qc = useQueryClient();
  const [confirm, setConfirm] = useState(false);
  const { label, mobile } = describeAgent(s.userAgent, t.locale);
  const Icon = mobile ? Smartphone : Laptop;
  const seen =
    Date.now() - Date.parse(s.lastSeenAt) < 5 * 60_000
      ? t("session.activeNow")
      : t("session.active", { ago: distance(s.lastSeenAt, t.locale) });
  return (
    <li className="flex items-center gap-3 px-4 py-3.5">
      <span className={cn("flex size-9 shrink-0 items-center justify-center rounded-lg", s.current ? "bg-accent-soft text-accent-ink" : "bg-inset text-fg-3")}>
        <Icon className="size-[18px]" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="flex items-center gap-2 truncate text-sm font-medium text-fg">
          {label}
          {s.current ? <Chip tone="success">{t("session.thisDevice")}</Chip> : null}
        </p>
        <p className="truncate text-xs text-fg-3">
          {s.ip || t("session.unknownAddress")} · {seen} · {t("session.signedIn", { ago: distance(s.createdAt, t.locale) })}
        </p>
      </div>
      {!s.current ? (
        <Button size="sm" variant="ghost" onClick={() => setConfirm(true)}>
          {t("session.revoke")}
        </Button>
      ) : null}
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title={t("session.revokeTitle")}
        description={t("session.revokeBody", { device: label })}
        confirm={t("session.revokeConfirm")}
        onConfirm={async () => {
          try {
            await ep.auth.revoke(s.id);
            await qc.invalidateQueries({ queryKey: qk.sessions });
            toast.success(tr()("session.revoked", { device: label }));
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
    </li>
  );
}
