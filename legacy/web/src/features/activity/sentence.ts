// An activity entry as a sentence in the reader's language. Pure: unit-tested.
import type { ActivityEntry } from "../../api/types";
import type { Key } from "../../i18n";

export type SentenceKey = Extract<Key, `activity.${"task" | "contact" | "group"}.${string}`>;

/** Who a share, an assignment or a removal was about: the reader, or someone named. */
type Person = { you: true } | { you: false; name: string };

/**
 * The sentence of an entry, in the reader's language, from its kind and its
 * references. The person a share, an assignment or a removal was about is the
 * entry's target; an entry written before the server sent one only says it in
 * its English text ("with you", "to Sam Rivera"), which is read as a last resort.
 */
export function sentenceOf(e: ActivityEntry, meId?: string): { key: SentenceKey; target?: string } | null {
  const fromText = (name: string | null | undefined): Person | null =>
    !name ? null : name === "you" ? { you: true } : { you: false, name };
  const after = (marker: string) => {
    const i = e.text.lastIndexOf(marker);
    return i < 0 ? null : e.text.slice(i + marker.length).replace(/\.$/, "").trim();
  };
  const person = (legacy: () => string | null | undefined): Person | null =>
    e.target ? (e.target.id === meId ? { you: true } : { you: false, name: e.target.name || e.target.email }) : fromText(legacy());
  const about = (you: SentenceKey, other: SentenceKey, p: Person | null) =>
    p === null ? null : p.you ? { key: you } : { key: other, target: p.name };

  if (e.kind.startsWith("task.") && !e.task) return null;
  if (e.kind.startsWith("group.") && !e.group) return null;
  switch (e.kind) {
    case "task.created":
      return { key: e.group ? "activity.task.createdIn" : "activity.task.created" };
    case "task.updated":
    case "task.completed":
    case "task.reopened":
    case "task.archived":
    case "task.restored":
    case "task.overdue":
    case "task.deleted":
    case "contact.requested":
    case "contact.accepted":
    case "group.invited":
    case "group.joined":
    case "group.left":
    case "group.deleted":
      return { key: `activity.${e.kind}` as SentenceKey };
    case "task.shared":
      return about("activity.task.sharedYou", "activity.task.shared", person(() => after("” with ")));
    case "task.unshared":
      // No target: the actor stopped sharing it with themselves — they left the task.
      if (!e.target && (!e.actor || e.text.startsWith(`${e.actor.name} left `))) return { key: "activity.task.left" };
      return about("activity.task.unsharedYou", "activity.task.unshared", person(() => after("” with ")));
    case "task.assigned":
      return about("activity.task.assignedYou", "activity.task.assigned", person(() => after("” to ")));
    case "group.removed":
      return about("activity.group.removedYou", "activity.group.removed", person(() => / removed (.+) from /.exec(e.text)?.[1]));
    default:
      return null;
  }
}
