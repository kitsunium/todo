import { GROUP_COLORS } from "../../api/types";
import { cn } from "../../lib/cn";

const HEX: Record<string, string> = {
  slate: "#64748b",
  red: "#e5484d",
  orange: "#f76b15",
  amber: "#f5a524",
  green: "#30a46c",
  teal: "#12a594",
  blue: "#0090ff",
  indigo: "#3e63dd",
  violet: "#8e4ec6",
  pink: "#d6409f",
};

export function groupHex(color: string): string {
  return HEX[color] ?? HEX.slate!;
}

export const GROUP_COLOR_LIST = GROUP_COLORS.map((c) => ({ name: c, hex: groupHex(c) }));

export function GroupDot({ color, className }: { color: string; className?: string }) {
  return (
    <span
      className={cn("inline-block size-2 shrink-0 rounded-full", className)}
      style={{ backgroundColor: groupHex(color) }}
      aria-hidden="true"
    />
  );
}

/** A rounded square with the group's initial: the group's "avatar". */
export function GroupBadge({ name, color, size = "md" }: { name: string; color: string; size?: "sm" | "md" | "lg" }) {
  const hex = groupHex(color);
  const s = size === "lg" ? "size-9 text-base rounded-[10px]" : size === "sm" ? "size-5 text-[10px] rounded-[6px]" : "size-7 text-xs rounded-lg";
  return (
    <span
      className={cn("inline-flex shrink-0 items-center justify-center font-semibold text-white select-none", s)}
      style={{
        background: `linear-gradient(180deg, color-mix(in oklab, ${hex} 88%, white) 0%, ${hex} 100%)`,
        boxShadow: `inset 0 1px 0 rgb(255 255 255 / 0.22), 0 1px 2px color-mix(in oklab, ${hex} 40%, transparent)`,
      }}
      aria-hidden="true"
    >
      {Array.from(name.trim())[0]?.toUpperCase() ?? "#"}
    </span>
  );
}
