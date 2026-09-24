// A password strength hint. The backend's rule is 10 to 128 characters; the
// score only nudges toward longer, less guessable passwords.
import { getLocale, tr, type Locale } from "../i18n";

export type Strength = { score: 0 | 1 | 2 | 3 | 4; label: string; ok: boolean };

export const PASSWORD_MIN = 10;
export const PASSWORD_MAX = 128;

const COMMON = /(password|passwd|qwerty|azerty|123456|abcdef|letmein|welcome|iloveyou|admin|todo|motdepasse|bonjour|soleil)/i;

export function passwordStrength(pw: string, loc: Locale = getLocale()): Strength {
  const t = tr(loc);
  const n = Array.from(pw).length;
  if (n === 0) return { score: 0, label: "", ok: false };
  if (n < PASSWORD_MIN) return { score: 1, label: t("password.more", { count: PASSWORD_MIN - n }), ok: false };
  if (n > PASSWORD_MAX) return { score: 1, label: t("password.tooLong"), ok: false };
  let score = n >= 16 ? 4 : n >= 12 ? 3 : 2;
  const classes = [/[a-z]/, /[A-Z]/, /\d/, /[^A-Za-z0-9]/].filter((r) => r.test(pw)).length;
  if (classes <= 1 && n < 20) score -= 1;
  if (COMMON.test(pw) || /(.)\1{3,}/.test(pw)) score -= 1;
  const s = Math.max(1, Math.min(4, score)) as 1 | 2 | 3 | 4;
  const label = t(s === 1 ? "password.weak" : s === 2 ? "password.fair" : s === 3 ? "password.good" : "password.strong");
  return { score: s, label, ok: true };
}
