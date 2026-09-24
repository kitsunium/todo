import { useQueryClient } from "@tanstack/react-query";
import { format, formatDistanceToNowStrict } from "date-fns";
import { Laptop, Monitor, Moon, Settings, Smartphone, Sun } from "lucide-react";
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
import { cn } from "../../lib/cn";
import { passwordStrength } from "../../lib/password";
import { setTheme, useTheme, type ThemeChoice } from "../../lib/theme";
import { StrengthMeter } from "../auth/forms";
import { Page } from "../shell/Page";

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
  return (
    <Page title="Settings" icon={<Settings className="text-fg-2" strokeWidth={2} />} subtitle="Your profile, your password, your devices." wide>
      <Section title="Profile" description="How other people see you when you share tasks.">
        <Profile />
      </Section>
      <Section title="Appearance" description="Follow your system, or pick a side.">
        <Appearance />
      </Section>
      <Section title="Password" description="Changing it signs you out everywhere else.">
        <Password />
      </Section>
      <Section title="Sessions" description="Devices signed in to your account.">
        <Sessions />
      </Section>
    </Page>
  );
}

function Profile() {
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
      const u = await ep.auth.rename(name.trim());
      qc.setQueryData(qk.me, u);
      toast.success("Profile updated");
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
          <p className="truncate text-xs text-fg-3">Member since {format(new Date(me.createdAt), "MMMM yyyy")}</p>
        </div>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Name" error={!name.trim() ? "Your name can’t be empty." : fieldErrors(error).name}>
          {(p) => <Input {...p} value={name} maxLength={80} autoComplete="name" onChange={(e) => setName(e.target.value)} />}
        </Field>
        <Field label="Email" hint="Your email is your sign-in: it can’t be changed.">
          {(p) => <Input {...p} value={me.email} readOnly className="text-fg-3" />}
        </Field>
      </div>
      {error && !fieldErrors(error).name ? <Callout tone="danger" className="mt-4">{errorMessage(error)}</Callout> : null}
      <div className="mt-5 flex justify-end gap-2">
        {dirty ? (
          <Button variant="ghost" onClick={() => setName(me.name)}>
            Discard
          </Button>
        ) : null}
        <Button type="submit" variant="primary" disabled={!dirty} loading={busy}>
          Save
        </Button>
      </div>
    </form>
  );
}

function ThemeCard({ value, label, icon, current }: { value: ThemeChoice; label: string; icon: ReactNode; current: ThemeChoice }) {
  const on = value === current;
  const preview = (dark: boolean) => (
    <span className={cn("flex h-full w-full gap-1 p-1.5", dark ? "bg-[#0e0e10]" : "bg-[#f5f5f4]")}>
      <span className={cn("flex w-5 flex-col gap-1 pt-0.5")}>
        <span className="h-1 w-3 rounded-full bg-[#f26b1d]" />
        <span className={cn("h-1 w-4 rounded-full", dark ? "bg-[#3a3a40]" : "bg-[#d6d3d1]")} />
        <span className={cn("h-1 w-3.5 rounded-full", dark ? "bg-[#3a3a40]" : "bg-[#d6d3d1]")} />
      </span>
      <span className={cn("flex flex-1 flex-col gap-1.5 rounded-[4px] p-1.5", dark ? "bg-[#18181b]" : "bg-white")}>
        {[0, 1, 2].map((i) => (
          <span key={i} className="flex items-center gap-1">
            <span className={cn("size-1.5 rounded-full border", i === 0 ? "border-[#e5484d]" : dark ? "border-[#5c5c66]" : "border-[#a8a29e]")} />
            <span className={cn("h-1 rounded-full", i === 1 ? "w-8" : "w-10", dark ? "bg-[#3a3a40]" : "bg-[#e7e5e4]")} />
          </span>
        ))}
      </span>
    </span>
  );
  return (
    <button
      type="button"
      role="radio"
      aria-checked={on}
      onClick={() => setTheme(value)}
      className={cn(
        "group flex flex-col gap-2 rounded-xl p-1.5 text-left outline-none transition-shadow focus-visible:ring-2 focus-visible:ring-ring",
        on ? "shadow-[0_0_0_2px_var(--accent)]" : "shadow-card hover:shadow-[0_0_0_1px_var(--line-strong)]",
      )}
    >
      <span className="flex h-[76px] overflow-hidden rounded-lg">
        {value === "system" ? (
          <>
            <span className="w-1/2 overflow-hidden">{preview(false)}</span>
            <span className="w-1/2 overflow-hidden">{preview(true)}</span>
          </>
        ) : (
          preview(value === "dark")
        )}
      </span>
      <span className="flex items-center gap-1.5 px-1 pb-0.5 text-sm font-medium text-fg">
        <span className="text-fg-3">{icon}</span>
        {label}
      </span>
    </button>
  );
}

function Appearance() {
  const { choice } = useTheme();
  return (
    <div role="radiogroup" aria-label="Theme" className="grid grid-cols-3 gap-3">
      <ThemeCard value="light" label="Light" icon={<Sun className="size-4" />} current={choice} />
      <ThemeCard value="dark" label="Dark" icon={<Moon className="size-4" />} current={choice} />
      <ThemeCard value="system" label="System" icon={<Monitor className="size-4" />} current={choice} />
    </div>
  );
}

function Password() {
  const qc = useQueryClient();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const strength = passwordStrength(next);
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
      toast.success("Password changed", { description: "Your other devices were signed out." });
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
        <Field label="Current password" error={wrongCurrent ? "That’s not your current password." : server.current ?? (submitted && !current ? "Enter your current password." : undefined)}>
          {(p) => <PasswordInput {...p} value={current} autoComplete="current-password" onChange={(e) => setCurrent(e.target.value)} />}
        </Field>
        <Field label="New password" error={server.password} hint={<StrengthMeter password={next} invalid={submitted && !strength.ok} />}>
          {(p) => <PasswordInput {...p} value={next} maxLength={128} autoComplete="new-password" onChange={(e) => setNext(e.target.value)} />}
        </Field>
      </div>
      {error && !wrongCurrent && !Object.keys(server).length ? <Callout tone="danger" className="mt-4">{errorMessage(error)}</Callout> : null}
      <div className="mt-5 flex justify-end">
        <Button type="submit" variant="primary" loading={busy} disabled={!current && !next}>
          Change password
        </Button>
      </div>
    </form>
  );
}

/** "Chrome on macOS" from a user agent: enough to recognize a device. */
export function describeAgent(ua: string): { label: string; mobile: boolean } {
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
              ? "API client"
              : "Browser";
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
  return { label: os ? `${browser} on ${os}` : browser, mobile: /iPhone|Android.*Mobile|Mobile/.test(ua) };
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
  const qc = useQueryClient();
  const [confirm, setConfirm] = useState(false);
  const { label, mobile } = describeAgent(s.userAgent);
  const Icon = mobile ? Smartphone : Laptop;
  const seen = Date.now() - Date.parse(s.lastSeenAt) < 5 * 60_000 ? "Active now" : `Active ${formatDistanceToNowStrict(new Date(s.lastSeenAt), { addSuffix: true })}`;
  return (
    <li className="flex items-center gap-3 px-4 py-3.5">
      <span className={cn("flex size-9 shrink-0 items-center justify-center rounded-lg", s.current ? "bg-accent-soft text-accent-ink" : "bg-inset text-fg-3")}>
        <Icon className="size-[18px]" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="flex items-center gap-2 truncate text-sm font-medium text-fg">
          {label}
          {s.current ? <Chip tone="success">This device</Chip> : null}
        </p>
        <p className="truncate text-xs text-fg-3">
          {s.ip || "Unknown address"} · {seen} · signed in {formatDistanceToNowStrict(new Date(s.createdAt), { addSuffix: true })}
        </p>
      </div>
      {!s.current ? (
        <Button size="sm" variant="ghost" onClick={() => setConfirm(true)}>
          Revoke
        </Button>
      ) : null}
      <ConfirmDialog
        open={confirm}
        onOpenChange={setConfirm}
        title="Sign out this device?"
        description={`${label} will need to sign in again.`}
        confirm="Sign out device"
        onConfirm={async () => {
          try {
            await ep.auth.revoke(s.id);
            await qc.invalidateQueries({ queryKey: qk.sessions });
            toast.success(`${label} was signed out`);
          } catch (err) {
            toast.error(errorMessage(err));
          }
        }}
      />
    </li>
  );
}
