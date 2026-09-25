import type { UserRef } from "../../api/types";
import { cn } from "../../lib/cn";
import { hueOf, initials } from "../../lib/people";
import { Tooltip } from "./tooltip";

const SIZES = {
  xs: "size-[18px] text-[8px]",
  sm: "size-5 text-[9px]",
  md: "size-6 text-[10px]",
  lg: "size-8 text-xs",
  xl: "size-10 text-sm",
} as const;

export type AvatarSize = keyof typeof SIZES;

export function Avatar({
  user,
  size = "md",
  className,
  ring,
  title,
}: {
  user: Pick<UserRef, "id" | "name" | "email">;
  size?: AvatarSize;
  className?: string;
  ring?: boolean;
  title?: boolean;
}) {
  const hue = hueOf(user.id);
  const el = (
    <span
      className={cn(
        "relative inline-flex shrink-0 items-center justify-center rounded-full font-semibold tracking-[0.02em] select-none",
        "bg-[oklch(0.9_0.055_var(--h))] text-[oklch(0.42_0.11_var(--h))]",
        ring && "ring-[1.5px] ring-sheet",
        SIZES[size],
        className,
      )}
      style={{ ["--h" as string]: String(hue) }}
      aria-label={title ? undefined : user.name || user.email}
      role={title ? undefined : "img"}
    >
      {initials(user.name, user.email)}
    </span>
  );
  if (!title) return el;
  return <Tooltip content={user.name || user.email}>{el}</Tooltip>;
}

/** Overlapping avatars; "+3" past max. */
export function AvatarStack({
  users,
  max = 3,
  size = "sm",
  className,
}: {
  users: Pick<UserRef, "id" | "name" | "email">[];
  max?: number;
  size?: AvatarSize;
  className?: string;
}) {
  if (users.length === 0) return null;
  const shown = users.slice(0, max);
  const rest = users.length - shown.length;
  const names = users.map((u) => u.name || u.email).join(", ");
  return (
    <span className={cn("inline-flex items-center", className)} role="img" aria-label={names}>
      {shown.map((u, i) => (
        <Avatar key={u.id} user={u} size={size} ring className={i > 0 ? "-ml-1" : ""} />
      ))}
      {rest > 0 ? (
        <span
          className={cn(
            "-ml-1 inline-flex items-center justify-center rounded-full bg-inset font-semibold text-fg-3 ring-2 ring-sheet",
            SIZES[size],
          )}
        >
          +{rest}
        </span>
      ) : null}
    </span>
  );
}
