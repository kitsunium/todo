// The brand panel of the sign-in pages: three task cards in a loose stack,
// the fox keeping an eye on them. Drawn by hand; text uses the page's Inter,
// in the page's language — the pills are sized to their words.
import { useEffect, useState } from "react";
import { capitalize, useT } from "../../i18n";
import { formatTime } from "../../lib/dates";

const FONT = "Inter Variable, system-ui, sans-serif";

let canvas: HTMLCanvasElement | null = null;
/** The width of a line of text in Inter, measured by the browser. */
function textWidth(text: string, size: number, weight: number): number {
  const ctx = typeof document !== "undefined" ? (canvas ??= document.createElement("canvas")).getContext("2d") : null;
  if (!ctx) return text.length * size * 0.62;
  ctx.font = `${weight} ${size}px ${FONT}`;
  return ctx.measureText(text).width;
}

/** Re-renders once the web font is in: widths measured before it are the fallback's. */
function useFontsReady(): boolean {
  const [ready, setReady] = useState(() => typeof document === "undefined" || document.fonts?.status === "loaded");
  useEffect(() => {
    if (ready || !document.fonts) return;
    let live = true;
    void document.fonts.ready.then(() => live && setReady(true));
    return () => {
      live = false;
    };
  }, [ready]);
  return ready;
}

const pillWidth = (text: string) => Math.ceil(textWidth(text, 11, 600)) + 16;

function Check({ x, y, done, ring }: { x: number; y: number; done?: boolean; ring: string }) {
  return done ? (
    <g>
      <circle cx={x} cy={y} r="9" fill="#30a46c" />
      <path d={`m${x - 4} ${y}.3 2.8 2.7 5.2-5.4`} fill="none" stroke="#fff" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </g>
  ) : (
    <circle cx={x} cy={y} r="8.25" fill={`color-mix(in srgb, ${ring} 10%, white)`} stroke={ring} strokeWidth="1.8" />
  );
}

function Pill({ x, y, w, text, color, bg }: { x: number; y: number; w: number; text: string; color: string; bg: string }) {
  return (
    <g>
      <rect x={x} y={y} width={w} height="20" rx="6" fill={bg} />
      <text x={x + 8} y={y + 14} fontSize="11" fontWeight="600" fill={color} fontFamily={FONT}>
        {text}
      </text>
    </g>
  );
}

function Face({ cx, cy, r, hue, letters }: { cx: number; cy: number; r: number; hue: number; letters: string }) {
  return (
    <g>
      <circle cx={cx} cy={cy} r={r + 2} fill="#fff" />
      <circle cx={cx} cy={cy} r={r} fill={`oklch(0.9 0.06 ${hue})`} />
      <text
        x={cx}
        y={cy + 3.6}
        textAnchor="middle"
        fontSize="10"
        fontWeight="700"
        fill={`oklch(0.42 0.11 ${hue})`}
        fontFamily={FONT}
      >
        {letters}
      </text>
    </g>
  );
}

export function AuthArt({ className }: { className?: string }) {
  const t = useT();
  useFontsReady();
  const font = FONT;
  const upper = (s: string) => s.toLocaleUpperCase(t.locale);
  const home = t("auth.art.home");
  const time = formatTime(new Date(2026, 0, 1, 17, 0), t.locale);
  const launch = t("auth.art.launch");
  const venue = t("auth.art.venue");
  const deck = t("auth.art.deck");
  const homeW = pillWidth(home);
  const timeW = pillWidth(time);
  const launchW = pillWidth(launch);
  // The venue and its pill share the middle card's line: the pill follows the words.
  const homeX = Math.min(52 + Math.ceil(textWidth(venue, 14, 500)) + 12, 304 - 16 - homeW);
  const launchX = 56 + timeW + 6;
  const facesX = launchX + launchW + 24;
  return (
    <svg viewBox="0 0 440 400" className={className} role="img" aria-label={t("auth.art.label")}>
      <defs>
        <filter id="aa-shadow" x="-30%" y="-30%" width="160%" height="170%">
          <feDropShadow dx="0" dy="18" stdDeviation="18" floodColor="#7a2a06" floodOpacity="0.28" />
          <feDropShadow dx="0" dy="2" stdDeviation="2" floodColor="#7a2a06" floodOpacity="0.12" />
        </filter>
        <linearGradient id="aa-card" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#ffffff" />
          <stop offset="1" stopColor="#fbfaf9" />
        </linearGradient>
      </defs>

      {/* back card */}
      <g transform="rotate(-7 220 200) translate(54 38)" filter="url(#aa-shadow)" opacity="0.92">
        <rect width="300" height="148" rx="18" fill="url(#aa-card)" />
        <rect x="24" y="26" width="92" height="9" rx="4.5" fill="#e7e5e4" />
        <rect x="24" y="58" width="220" height="8" rx="4" fill="#efeeec" />
        <rect x="24" y="80" width="180" height="8" rx="4" fill="#efeeec" />
        <rect x="24" y="102" width="200" height="8" rx="4" fill="#efeeec" />
      </g>

      {/* middle card */}
      <g transform="rotate(4 220 200) translate(80 96)" filter="url(#aa-shadow)">
        <rect width="304" height="150" rx="18" fill="url(#aa-card)" />
        <text x="24" y="34" fontSize="12" fontWeight="600" fill="#a8a29e" fontFamily={font} letterSpacing="0.02em">
          {upper(t("section.tomorrow"))}
        </text>
        <Check x={33} y={62} ring="#ffb224" />
        <text x="52" y="66.5" fontSize="14" fontWeight="500" fill="#1c1917" fontFamily={font}>
          {venue}
        </text>
        <Pill x={homeX} y={52} w={homeW} text={home} color="#218358" bg="#e6f6ec" />
      </g>

      {/* front card */}
      <g transform="translate(40 186)" filter="url(#aa-shadow)">
        <rect width="336" height="182" rx="20" fill="url(#aa-card)" />
        <text x="26" y="38" fontSize="12" fontWeight="600" fill="#c2410c" fontFamily={font} letterSpacing="0.02em">
          {upper(capitalize(t("section.today"), t.locale))}
          <tspan dx="9" fill="#a8a29e" letterSpacing="0">
            3
          </tspan>
        </text>

        <Check x={35} y={70} done ring="#30a46c" />
        <text x="56" y="74.5" fontSize="14" fontWeight="500" fill="#a8a29e" fontFamily={font}>
          {deck}
        </text>
        <line x1="56" y1="70" x2={56 + Math.ceil(textWidth(deck, 14, 500)) + 2} y2="70" stroke="#a8a29e" strokeWidth="1.4" />

        <Check x={35} y={110} ring="#e5484d" />
        <text x="56" y="114.5" fontSize="14" fontWeight="600" fill="#1c1917" fontFamily={font}>
          {t("auth.art.landing")}
        </text>
        <Pill x={56} y={124} w={timeW} text={time} color="#c2410c" bg="#fdeee4" />
        <Pill x={launchX} y={124} w={launchW} text={launch} color="#b24c0c" bg="#fff1e6" />
        <Face cx={facesX} cy={134} r={10} hue={250} letters="SR" />
        <Face cx={facesX + 21} cy={134} r={10} hue={150} letters="NH" />

        <Check x={35} y={160} ring="#f76b15" />
        <rect x="56" y="155" width="150" height="9" rx="4.5" fill="#e7e5e4" />
      </g>

      {/* the fox, a seal on the corner of the front card */}
      <g transform="translate(344 162) rotate(10)" filter="url(#aa-shadow)">
        <circle cx="26" cy="26" r="27" fill="#fff" />
        <g transform="translate(6.5 6.5) scale(1.22)">
          <path d="M5.5 4.5 12.5 10h7l7-5.5 1 11L16 27.5 4.5 15.5Z" fill="#f26b1d" stroke="#f26b1d" strokeWidth="2.4" strokeLinejoin="round" />
          <path d="m10.8 16.2 3.8 3.6 6.8-7" fill="none" stroke="#fff" strokeWidth="2.6" strokeLinecap="round" strokeLinejoin="round" />
        </g>
      </g>

      {/* sparkles */}
      <g fill="#fff">
        <path d="M58 150l3 7 7 3-7 3-3 7-3-7-7-3 7-3z" opacity="0.8" />
        <path d="M392 330l2 5 5 2-5 2-2 5-2-5-5-2 5-2z" opacity="0.7" />
        <circle cx="96" cy="120" r="2.5" opacity="0.6" />
        <circle cx="404" cy="96" r="3" opacity="0.5" />
      </g>
    </svg>
  );
}
