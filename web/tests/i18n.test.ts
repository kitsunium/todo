import { isValidElement, type ReactElement, type ReactNode } from "react";
import { describe, expect, it } from "vitest";
import type { ActivityEntry } from "../src/api/types";
import { sentenceOf } from "../src/features/activity/sentence";
import { en } from "../src/i18n/en";
import { fr } from "../src/i18n/fr";
import { tr, translateRich } from "../src/i18n";

const holes = (s: string) => [...s.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort();
const forms = (v: unknown) => (typeof v === "string" ? [v] : Object.values(v as Record<string, string>));

describe("the dictionaries", () => {
  it("have exactly the same keys", () => {
    expect(Object.keys(en).sort()).toEqual(Object.keys(fr).sort());
  });

  it("have the same placeholders and the same plurals, key by key", () => {
    for (const key of Object.keys(fr) as (keyof typeof fr)[]) {
      const a = fr[key];
      const b = en[key];
      expect(typeof b, key).toBe(typeof a);
      const ha = [...new Set(forms(a).flatMap(holes))].sort();
      const hb = [...new Set(forms(b).flatMap(holes))].sort();
      expect(hb, key).toEqual(ha);
    }
  });

  it("say nothing twice by accident: French is not English", () => {
    const same = (Object.keys(fr) as (keyof typeof fr)[]).filter(
      (k) => typeof fr[k] === "string" && fr[k] === en[k] && !/^(lang|date\.pattern|color|role\.admin|common\.more|signup\.namePlaceholder|login\.mockHint)/.test(k),
    );
    // Words that are the same in both languages, by design.
    expect(same.sort()).toEqual(
      [
        "activity.invitations",
        "app.navigation",
        "session.active",
        "contacts.tabContacts",
        "contacts.title",
        "date.withTime",
        "field.notes",
        "palette.actions",
        "panel.notes",
        "session.sessions",
        "settings.sessions",
        "sidebar.contacts",
        "quickadd.date",
      ]
        .filter((k) => k in fr)
        .sort(),
    );
  });
});

describe("translate", () => {
  it("interpolates and formats numbers in the language", () => {
    expect(tr("fr")("verify.welcome", { name: "Camille" })).toBe("Adresse confirmée — bienvenue sur Todo, Camille.");
    expect(tr("en")("verify.welcome", { name: "Camille" })).toBe("Email confirmed — welcome to Todo, Camille.");
    expect(tr("fr")("group.members", { count: 1200 }).replace(/ /g, " ")).toBe("1 200 membres");
  });

  it("French plurals: 0 and 1 are singular", () => {
    const t = tr("fr");
    expect(t("list.moved", { count: 0 })).toBe("0 tâche reportée à aujourd’hui");
    expect(t("list.moved", { count: 1 })).toBe("1 tâche reportée à aujourd’hui");
    expect(t("list.moved", { count: 2 })).toBe("2 tâches reportées à aujourd’hui");
    expect(t("view.completed.week", { count: 0 })).toBe("0 tâche terminée cette semaine");
  });

  it("English plurals: 1 is singular, 0 is not", () => {
    const t = tr("en");
    expect(t("list.moved", { count: 0 })).toBe("Moved 0 tasks to today");
    expect(t("list.moved", { count: 1 })).toBe("Moved 1 task to today");
    expect(t("group.members", { count: 2 })).toBe("2 members");
  });

  it("rich: React nodes in the holes, never HTML", () => {
    const node = translateRich("fr", "invitation.text", { inviter: "Noor", group: "<b>Club</b>" }) as ReactElement<{ children: ReactNode[] }>;
    expect(isValidElement(node)).toBe(true);
    expect(node.props.children).toEqual(["Noor", " vous invite à rejoindre ", "<b>Club</b>"]);
  });
});

describe("activity sentences", () => {
  const actor = { id: "user_sam", name: "Sam Rivera", email: "sam@example.com" };
  const task = { id: "task_1", title: "Ship it" };
  const group = { id: "group_1", name: "Launch", color: "orange" };
  const entry = (kind: string, text: string, extra: Partial<ActivityEntry> = {}): ActivityEntry => ({
    id: "e",
    kind,
    actor,
    text,
    at: "2026-09-24T08:00:00Z",
    read: false,
    ...extra,
  });

  it("come from the kind, not from the server's English", () => {
    expect(sentenceOf(entry("task.completed", "Sam Rivera completed “Ship it”.", { task }))).toEqual({ key: "activity.task.completed" });
    expect(sentenceOf(entry("task.created", "Sam Rivera added “Ship it” to Launch.", { task, group }))).toEqual({ key: "activity.task.createdIn" });
    expect(sentenceOf(entry("contact.requested", "Sam Rivera wants to add you as a contact."))).toEqual({ key: "activity.contact.requested" });
  });

  it("read who a share, an assignment or a removal was about", () => {
    expect(sentenceOf(entry("task.shared", "Sam Rivera shared “Ship it” with you.", { task }))).toEqual({ key: "activity.task.sharedYou" });
    expect(sentenceOf(entry("task.assigned", "Sam Rivera assigned “Ship it” to Léa Dubois.", { task }))).toEqual({
      key: "activity.task.assigned",
      target: "Léa Dubois",
    });
    expect(sentenceOf(entry("task.unshared", "Sam Rivera left “Ship it”.", { task }))).toEqual({ key: "activity.task.left" });
    expect(sentenceOf(entry("group.removed", "Sam Rivera removed you from Launch.", { group }))).toEqual({ key: "activity.group.removedYou" });
  });

  it("take the person from the entry's target: the reader, or someone named", () => {
    const me = { id: "user_me", name: "Camille Martin", email: "camille@example.com" };
    const lea = { id: "user_lea", name: "Léa Dubois", email: "lea@example.com" };
    // The text is ignored once there is a target: here it even says something else.
    expect(sentenceOf(entry("task.shared", "(whatever)", { task, target: me }), me.id)).toEqual({ key: "activity.task.sharedYou" });
    expect(sentenceOf(entry("task.assigned", "(whatever)", { task, target: lea }), me.id)).toEqual({
      key: "activity.task.assigned",
      target: "Léa Dubois",
    });
    expect(sentenceOf(entry("task.unshared", "(whatever)", { task, target: lea }), me.id)).toEqual({
      key: "activity.task.unshared",
      target: "Léa Dubois",
    });
    expect(sentenceOf(entry("group.removed", "(whatever)", { group, target: me }), me.id)).toEqual({ key: "activity.group.removedYou" });
    // No target on an unshare: the actor left the task.
    expect(sentenceOf(entry("task.unshared", "Sam Rivera left “Ship it”.", { task }), me.id)).toEqual({ key: "activity.task.left" });
  });

  it("an unknown kind falls back to the server's words", () => {
    expect(sentenceOf(entry("task.teleported", "Sam Rivera teleported “Ship it”.", { task }))).toBeNull();
  });

  it("renders in both languages", () => {
    expect(tr("fr")("activity.task.assignedYou", { actor: "Sam", task: "« Ship it »" })).toBe("Sam vous a assigné « Ship it ».");
    expect(tr("en")("activity.task.assignedYou", { actor: "Sam", task: "“Ship it”" })).toBe("Sam assigned “Ship it” to you.");
  });
});
