import { useState } from "react";
import { cn } from "../../lib/cn";
import { passwordStrength } from "../../lib/password";

export const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function checkEmail(v: string): string | undefined {
  const e = v.trim();
  if (!e) return "Enter your email address.";
  if (!EMAIL_RE.test(e)) return "That doesn’t look like an email address.";
  return undefined;
}

/** Field errors that show once a field was left, or after a submit. */
export function useTouched<K extends string>() {
  const [touched, setTouched] = useState<Partial<Record<K, boolean>>>({});
  const [submitted, setSubmitted] = useState(false);
  return {
    shown: (k: K) => submitted || !!touched[k],
    touch: (k: K) => setTouched((t) => ({ ...t, [k]: true })),
    submit: () => setSubmitted(true),
  };
}

export function StrengthMeter({ password, invalid }: { password: string; invalid?: boolean }) {
  const s = passwordStrength(password);
  const color =
    s.score <= 1 ? "bg-danger" : s.score === 2 ? "bg-medium" : s.score === 3 ? "bg-[#65b43f]" : "bg-done";
  return (
    <div className="flex items-center gap-3" aria-live="polite">
      <div className="flex flex-1 gap-1" aria-hidden="true">
        {[1, 2, 3, 4].map((i) => (
          <span
            key={i}
            className={cn("h-1 flex-1 rounded-full transition-colors duration-300", i <= s.score && s.score > 0 ? color : "bg-line")}
          />
        ))}
      </div>
      <span className={cn("shrink-0 text-right text-xs whitespace-nowrap", invalid ? "text-danger-ink" : s.score <= 1 ? "text-fg-3" : "text-fg-2")}>
        {s.label || "10+ characters"}
      </span>
    </div>
  );
}
