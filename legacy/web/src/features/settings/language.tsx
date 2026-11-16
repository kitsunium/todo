// The interface's language, tied to the account: signed in, the account's
// locale is the interface's; switching it saves it (optimistically) with
// PATCH /api/auth/me, so the mails the server sends follow too.
import { useQueryClient } from "@tanstack/react-query";
import { useLayoutEffect } from "react";
import * as ep from "../../api/endpoints";
import { errorMessage } from "../../api/errors";
import { qk } from "../../api/queries";
import type { User } from "../../api/types";
import { toast } from "../../components/ui/toast";
import {
  getLocale,
  LOCALES,
  setLocale,
  tr,
  useLocale,
  useT,
  type Locale,
} from "../../i18n";
import { cn } from "../../lib/cn";

/** Signed in: the account's language becomes the interface's, before the first paint. */
export function useLocaleSync(user: User) {
  const locale = user.locale;
  useLayoutEffect(() => setLocale(locale), [locale]);
}

// Only the last switch counts: an older answer arriving late must not undo it.
let latest = 0;

/** Switches the language — and, signed in, the account's. */
export function useSwitchLocale(): (next: Locale) => void {
  const qc = useQueryClient();
  return (next) => {
    const before = getLocale();
    const me = qc.getQueryData<User>(qk.me);
    setLocale(next);
    if (!me || me.locale === next) return;
    const mine = ++latest;
    qc.setQueryData<User>(qk.me, { ...me, locale: next });
    ep.auth.updateMe({ locale: next }).then(
      (user) => {
        if (mine === latest) qc.setQueryData(qk.me, user);
      },
      (err: unknown) => {
        if (mine !== latest) return;
        qc.setQueryData<User>(qk.me, (u) =>
          u ? { ...u, locale: me.locale } : u,
        );
        setLocale(before);
        toast.error(tr(before)("lang.failed"), {
          description: errorMessage(err, { locale: before }),
        });
      },
    );
  };
}

/** FR | EN: a compact switch, each language named in its own. */
export function LanguageSwitch({ className }: { className?: string }) {
  const locale = useLocale();
  const t = useT();
  const change = useSwitchLocale();
  return (
    <div
      role="group"
      aria-label={t("lang.label")}
      className={cn(
        "inline-flex items-center rounded-md bg-inset p-0.5",
        className,
      )}
    >
      {LOCALES.map((l) => (
        <button
          key={l}
          type="button"
          lang={l}
          aria-pressed={l === locale}
          title={tr(l)(`lang.${l}`)}
          onClick={() => change(l)}
          className={cn(
            "h-6 min-w-8 rounded-[5px] px-2 text-xs font-semibold tracking-[0.02em] outline-none transition-colors",
            "focus-visible:ring-2 focus-visible:ring-ring",
            l === locale
              ? "bg-surface text-fg shadow-card"
              : "text-fg-3 hover:text-fg",
          )}
        >
          <span aria-hidden="true">{t(`lang.${l}.short`)}</span>
          <span className="sr-only">{tr(l)(`lang.${l}`)}</span>
        </button>
      ))}
    </div>
  );
}

/** The Settings choice: a card per language, with a sample in that language. */
export function LanguageCards() {
  const locale = useLocale();
  const t = useT();
  const change = useSwitchLocale();
  return (
    <div
      role="group"
      aria-label={t("lang.label")}
      className="grid gap-3 sm:grid-cols-2"
    >
      {LOCALES.map((l) => {
        const on = l === locale;
        return (
          <button
            key={l}
            type="button"
            lang={l}
            aria-pressed={on}
            onClick={() => change(l)}
            className={cn(
              "flex items-center gap-3 rounded-xl bg-surface p-3.5 text-left outline-none transition-shadow focus-visible:ring-2 focus-visible:ring-ring",
              on
                ? "shadow-[0_0_0_2px_var(--accent)]"
                : "shadow-card hover:shadow-[0_0_0_1px_var(--line-strong)]",
            )}
          >
            <span
              className={cn(
                "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold tracking-[0.04em]",
                on ? "bg-accent-soft text-accent-ink" : "bg-inset text-fg-3",
              )}
              aria-hidden="true"
            >
              {tr(l)(`lang.${l}.short`)}
            </span>
            <span className="min-w-0 flex-1">
              <span className="block text-sm font-medium text-fg">
                {tr(l)(`lang.${l}`)}
              </span>
              <span className="block truncate text-xs text-fg-3">
                {tr(l)("settings.sample")}
              </span>
            </span>
            <span
              className={cn(
                "flex size-4 shrink-0 items-center justify-center rounded-full transition-colors",
                on
                  ? "bg-accent"
                  : "shadow-[inset_0_0_0_1.5px_var(--line-strong)]",
              )}
              aria-hidden="true"
            >
              {on ? <span className="size-1.5 rounded-full bg-white" /> : null}
            </span>
          </button>
        );
      })}
    </div>
  );
}
