// Avatars are initials on a hue derived from the person's ID: stable across
// sessions and devices, different for two people with the same initials.
import type { UserRef } from "../api/types";

export function hueOf(id: string): number {
  // FNV-1a, 32 bits.
  let h = 0x811c9dc5;
  for (let i = 0; i < id.length; i++) {
    h ^= id.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return (h >>> 0) % 360;
}

export function initials(name: string, email = ""): string {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (words.length >= 2) return (first(words[0]!) + first(words[words.length - 1]!)).toUpperCase();
  if (words.length === 1) return first(words[0]!).toUpperCase();
  return first(email).toUpperCase() || "?";
}

function first(s: string): string {
  return Array.from(s)[0] ?? "";
}

export function firstName(u: Pick<UserRef, "name" | "email">): string {
  return u.name.trim().split(/\s+/)[0] || u.email.split("@")[0] || "Someone";
}

/** "You" for the signed-in person, their name otherwise. */
export function displayName(u: UserRef, meId?: string): string {
  return u.id === meId ? "You" : u.name || u.email;
}
