// Empty-state illustrations: paper cards, soft ink, one warm accent.
// Colors come from CSS variables, so they follow the theme.
import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

const PAPER = "var(--illo-paper)";
const LINE = "var(--illo-line)";
const INK = "var(--illo-ink)";
const INK2 = "var(--illo-ink-2)";
const ACCENT = "var(--accent)";
const DONE = "var(--done)";

function Frame({ children, className, label }: { children: ReactNode; className?: string; label: string }) {
  return (
    <svg viewBox="0 0 160 120" className={cn("h-[120px] w-[160px]", className)} role="img" aria-label={label}>
      <ellipse cx="80" cy="108" rx="54" ry="5" fill="var(--illo-shadow)" />
      {children}
    </svg>
  );
}

function Card({ x, y, w, h, r = 10, rotate = 0 }: { x: number; y: number; w: number; h: number; r?: number; rotate?: number }) {
  return (
    <rect
      x={x}
      y={y}
      width={w}
      height={h}
      rx={r}
      fill={PAPER}
      stroke={LINE}
      strokeWidth="1.2"
      transform={rotate ? `rotate(${rotate} ${x + w / 2} ${y + h / 2})` : undefined}
    />
  );
}

function Bar({ x, y, w, h = 5, fill = INK }: { x: number; y: number; w: number; h?: number; fill?: string }) {
  return <rect x={x} y={y} width={w} height={h} rx={h / 2} fill={fill} />;
}

function Tick({ cx, cy, r = 7, color = DONE }: { cx: number; cy: number; r?: number; color?: string }) {
  return (
    <g>
      <circle cx={cx} cy={cy} r={r} fill={color} />
      <path
        d={`M${cx - r * 0.42} ${cy + r * 0.02} l${r * 0.3} ${r * 0.3} l${r * 0.56} -${r * 0.6}`}
        fill="none"
        stroke="#fff"
        strokeWidth={r * 0.26}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </g>
  );
}

function Ring({ cx, cy, r = 6, color = INK2 }: { cx: number; cy: number; r?: number; color?: string }) {
  return <circle cx={cx} cy={cy} r={r} fill="none" stroke={color} strokeWidth="1.6" />;
}

function Spark({ x, y, s = 1, fill = ACCENT }: { x: number; y: number; s?: number; fill?: string }) {
  return (
    <path
      transform={`translate(${x} ${y}) scale(${s})`}
      d="M0-6 1.6-1.6 6 0 1.6 1.6 0 6-1.6 1.6-6 0-1.6-1.6Z"
      fill={fill}
    />
  );
}

export function InboxIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="An empty inbox tray">
      <Card x={46} y={20} w={68} h={46} rotate={-6} />
      <Bar x={56} y={34} w={30} fill={INK} />
      <Bar x={56} y={44} w={42} fill={LINE} />
      <path d="M28 64h28l6 12h36l6-12h28v30a8 8 0 0 1-8 8H36a8 8 0 0 1-8-8Z" fill={PAPER} stroke={LINE} strokeWidth="1.2" />
      <path d="M28 64l12-22h80l12 22" fill="none" stroke={LINE} strokeWidth="1.2" strokeLinejoin="round" />
      <Tick cx={114} cy={30} r={9} />
      <Spark x={36} y={30} s={0.8} />
    </Frame>
  );
}

export function TodayIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="A clear day: the sun over a finished list">
      <circle cx="80" cy="46" r="24" fill={ACCENT} opacity="0.14" />
      <circle cx="80" cy="46" r="15" fill={ACCENT} opacity="0.9" />
      {Array.from({ length: 8 }, (_, i) => {
        const a = (i * Math.PI) / 4;
        return (
          <line
            key={i}
            x1={80 + Math.cos(a) * 21}
            y1={46 + Math.sin(a) * 21}
            x2={80 + Math.cos(a) * 27}
            y2={46 + Math.sin(a) * 27}
            stroke={ACCENT}
            strokeWidth="2"
            strokeLinecap="round"
            opacity="0.55"
          />
        );
      })}
      <Card x={30} y={62} w={100} h={38} />
      <Tick cx={46} cy={75} r={6} />
      <Bar x={58} y={72.5} w={44} fill={LINE} />
      <Tick cx={46} cy={89} r={6} />
      <Bar x={58} y={86.5} w={32} fill={LINE} />
    </Frame>
  );
}

export function UpcomingIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="A calendar with nothing planned">
      <Card x={36} y={18} w={88} h={80} r={12} />
      <path d="M36 30a12 12 0 0 1 12-12h64a12 12 0 0 1 12 12v8H36Z" fill={ACCENT} opacity="0.9" />
      <rect x="54" y="12" width="4" height="12" rx="2" fill={INK2} />
      <rect x="102" y="12" width="4" height="12" rx="2" fill={INK2} />
      {Array.from({ length: 12 }, (_, i) => (
        <rect key={i} x={46 + (i % 4) * 18} y={48 + Math.floor(i / 4) * 15} width="10" height="8" rx="2.5" fill={i === 6 ? ACCENT : LINE} opacity={i === 6 ? 0.9 : 1} />
      ))}
    </Frame>
  );
}

export function SharedIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="Two people passing a task between them">
      <Card x={44} y={34} w={72} h={44} />
      <Ring cx={58} cy={50} color={ACCENT} />
      <Bar x={70} y={47.5} w={34} />
      <Bar x={70} y={60} w={24} fill={LINE} />
      <circle cx="30" cy="56" r="13" fill="oklch(0.88 0.06 250)" />
      <circle cx="130" cy="56" r="13" fill="oklch(0.88 0.06 150)" />
      <path d="M44 88c10 10 62 10 72 0" fill="none" stroke={INK2} strokeWidth="1.4" strokeDasharray="3 4" strokeLinecap="round" />
      <path d="m112 84 4 4-5 2" fill="none" stroke={INK2} strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
    </Frame>
  );
}

export function AssignedIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="A task card with a name tag">
      <Card x={34} y={26} w={92} h={62} rotate={3} />
      <Ring cx={50} cy={46} color={ACCENT} />
      <Bar x={62} y={43.5} w={46} />
      <Bar x={62} y={56} w={30} fill={LINE} />
      <rect x="96" y="66" width="40" height="18" rx="9" fill={PAPER} stroke={LINE} strokeWidth="1.2" />
      <circle cx="106" cy="75" r="5.5" fill="oklch(0.86 0.07 20)" />
      <Bar x={114} y={72.5} w={16} fill={INK} />
      <Spark x={124} y={24} s={0.8} />
    </Frame>
  );
}

export function CompletedIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="A big check mark and confetti">
      <circle cx="80" cy="56" r="30" fill={DONE} opacity="0.12" />
      <Tick cx={80} cy={56} r={20} />
      <Spark x={40} y={34} s={0.9} />
      <Spark x={122} y={30} s={0.7} fill={DONE} />
      <rect x="116" y="74" width="6" height="6" rx="1.5" fill="var(--medium)" transform="rotate(20 119 77)" />
      <rect x="36" y="76" width="6" height="6" rx="1.5" fill="var(--low)" transform="rotate(-15 39 79)" />
      <circle cx="128" cy="52" r="2.5" fill={ACCENT} />
      <circle cx="30" cy="54" r="2.5" fill={DONE} />
    </Frame>
  );
}

export function ContactsIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="Two people, not yet connected">
      <Card x={18} y={34} w={54} h={62} r={12} />
      <circle cx="45" cy="56" r="11" fill="oklch(0.88 0.06 250)" />
      <Bar x={32} y={74} w={26} fill={INK} />
      <Bar x={36} y={83} w={18} fill={LINE} />
      <Card x={88} y={34} w={54} h={62} r={12} />
      <circle cx="115" cy="56" r="11" fill="oklch(0.88 0.06 150)" />
      <Bar x={102} y={74} w={26} fill={INK} />
      <Bar x={106} y={83} w={18} fill={LINE} />
      <circle cx="80" cy="30" r="10" fill={ACCENT} />
      <path d="M80 25.5v9M75.5 30h9" stroke="#fff" strokeWidth="2" strokeLinecap="round" />
    </Frame>
  );
}

export function ActivityIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="A quiet bell">
      <path d="M62 78V56a18 18 0 0 1 36 0v22l6 8H56Z" fill={PAPER} stroke={LINE} strokeWidth="1.2" strokeLinejoin="round" />
      <path d="M73 90a7 7 0 0 0 14 0" fill="none" stroke={INK2} strokeWidth="1.4" strokeLinecap="round" />
      <circle cx="80" cy="36" r="3" fill={INK2} />
      <path d="M44 50a36 36 0 0 1 8-16M116 50a36 36 0 0 0-8-16" fill="none" stroke={INK} strokeWidth="1.6" strokeLinecap="round" />
      <path d="M36 54a46 46 0 0 1 10-24M124 54a46 46 0 0 0-10-24" fill="none" stroke={LINE} strokeWidth="1.6" strokeLinecap="round" />
      <circle cx="98" cy="46" r="5" fill={ACCENT} />
    </Frame>
  );
}

export function GroupIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="A stack of shared lists">
      <Card x={38} y={22} w={84} h={30} rotate={-4} />
      <Card x={34} y={40} w={92} h={30} rotate={2} />
      <Card x={30} y={60} w={100} h={36} />
      <rect x="30" y="60" width="6" height="36" rx="3" fill={ACCENT} />
      <Ring cx={50} cy={78} />
      <Bar x={62} y={75.5} w={44} />
      <circle cx="116" cy="30" r="7" fill="oklch(0.88 0.06 300)" stroke={PAPER} strokeWidth="2" />
      <circle cx="126" cy="36" r="7" fill="oklch(0.88 0.06 200)" stroke={PAPER} strokeWidth="2" />
    </Frame>
  );
}

export function MailIllo({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 160 120" className={cn("h-[120px] w-[160px]", className)} role="img" aria-label="An envelope with a letter">
      <ellipse cx="80" cy="108" rx="54" ry="5" fill="var(--illo-shadow)" />
      <rect x="44" y="18" width="72" height="60" rx="8" fill={PAPER} stroke={LINE} strokeWidth="1.2" />
      <Bar x={54} y={30} w={36} fill={INK} />
      <Bar x={54} y={40} w={52} fill={LINE} />
      <Bar x={54} y={49} w={44} fill={LINE} />
      <rect x="54" y="58" width="30" height="10" rx="5" fill={ACCENT} />
      <path d="M28 56l52 30 52-30v38a8 8 0 0 1-8 8H36a8 8 0 0 1-8-8Z" fill={PAPER} stroke={LINE} strokeWidth="1.2" strokeLinejoin="round" />
      <path d="M28 100l38-26M132 100 94 74" fill="none" stroke={LINE} strokeWidth="1.2" />
      <Spark x={126} y={26} s={0.9} />
      <Spark x={34} y={40} s={0.6} fill="var(--medium)" />
    </svg>
  );
}

export function LostIllo({ className }: { className?: string }) {
  return (
    <Frame className={className} label="A fox looking for a missing page">
      <Card x={52} y={30} w={56} h={68} rotate={-8} />
      <Bar x={62} y={46} w={30} />
      <Bar x={62} y={56} w={22} fill={LINE} />
      <g transform="translate(92 40) scale(1.3)">
        <path d="M5.5 4.5 12.5 10h7l7-5.5 1 11L16 27.5 4.5 15.5Z" fill={ACCENT} stroke={ACCENT} strokeWidth="2.4" strokeLinejoin="round" />
        <circle cx="12" cy="16" r="1.6" fill="#fff" />
        <circle cx="20" cy="16" r="1.6" fill="#fff" />
      </g>
      <text x="44" y="30" fontSize="18" fontWeight="700" fill={INK2} fontFamily="Inter Variable, system-ui">?</text>
    </Frame>
  );
}
