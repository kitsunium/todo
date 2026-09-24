import { describe, expect, it } from "vitest";
import { mentionAt, mentionText, parseQuickAdd, parseTime, resolve } from "../src/lib/quickadd";

// Thursday, September 24 2026, 10:00 local time.
const NOW = new Date(2026, 8, 24, 10, 0);
const groups = [
  { id: "group_launch", name: "Launch" },
  { id: "group_home", name: "Home" },
  { id: "group_book", name: "Book club" },
];
const contacts = [
  { id: "user_sam", name: "Sam Rivera", email: "sam@example.com" },
  { id: "user_noor", name: "Noor Haddad", email: "noor.h@example.com" },
  { id: "user_sara", name: "Sara Kim", email: "skim@example.com" },
];
const parse = (s: string, ignore?: Set<"priority" | "due" | "group" | "contact">) =>
  parseQuickAdd(s, NOW, { groups, contacts, ...(ignore ? { ignore } : {}) });

const ymdhm = (d: Date) => [d.getFullYear(), d.getMonth() + 1, d.getDate(), d.getHours(), d.getMinutes()];

describe("priority", () => {
  it.each([
    ["!urgent", 1],
    ["!high", 2],
    ["!medium", 3],
    ["!low", 4],
    ["!1", 1],
    ["!4", 4],
    ["!HIGH", 2],
  ])("%s", (tok, p) => {
    const r = parse(`Ship it ${tok}`);
    expect(r.priority).toBe(p);
    expect(r.title).toBe("Ship it");
  });

  it("needs a word boundary", () => {
    expect(parse("Wow!high").priority).toBeUndefined();
    expect(parse("Ship !highest").priority).toBeUndefined();
  });
});

describe("due", () => {
  it("today is date-only, at 23:59", () => {
    const r = parse("Pay rent today");
    expect(r.title).toBe("Pay rent");
    expect(r.due?.hasTime).toBe(false);
    expect(ymdhm(r.due!.date)).toEqual([2026, 9, 24, 23, 59]);
  });

  it("tomorrow at 5pm", () => {
    const r = parse("Call mom tomorrow at 5pm");
    expect(r.title).toBe("Call mom");
    expect(r.due?.hasTime).toBe(true);
    expect(ymdhm(r.due!.date)).toEqual([2026, 9, 25, 17, 0]);
  });

  it("a 24-hour time before the day", () => {
    const r = parse("Standup 9:30 tomorrow");
    expect(r.title).toBe("Standup");
    expect(ymdhm(r.due!.date)).toEqual([2026, 9, 25, 9, 30]);
  });

  it("weekdays are the next occurrence, never today", () => {
    expect(ymdhm(parse("Review fri").due!.date)).toEqual([2026, 9, 25, 23, 59]);
    expect(ymdhm(parse("Review monday").due!.date)).toEqual([2026, 9, 28, 23, 59]);
    // Today is a Thursday: "thu" is next week's.
    expect(ymdhm(parse("Review thu").due!.date)).toEqual([2026, 10, 1, 23, 59]);
  });

  it("next week is next Monday", () => {
    const r = parse("Plan the sprint next week");
    expect(r.title).toBe("Plan the sprint");
    expect(ymdhm(r.due!.date)).toEqual([2026, 9, 28, 23, 59]);
  });

  it("in N days and weeks", () => {
    expect(ymdhm(parse("Renew in 3 days").due!.date)).toEqual([2026, 9, 27, 23, 59]);
    expect(ymdhm(parse("Renew in 2 weeks").due!.date)).toEqual([2026, 10, 8, 23, 59]);
  });

  it("a time alone is its next occurrence", () => {
    expect(ymdhm(parse("Gym at 18:00").due!.date)).toEqual([2026, 9, 24, 18, 0]);
    expect(ymdhm(parse("Breakfast at 8am").due!.date)).toEqual([2026, 9, 25, 8, 0]);
  });

  it("does not eat ordinary words and numbers", () => {
    expect(parse("Review today's numbers").due).toBeUndefined();
    expect(parse("Buy 5 apples").due).toBeUndefined();
    expect(parse("Plan next weekend").due).toBeUndefined();
    expect(parse("Book the monthly review").due).toBeUndefined();
  });

  it("an impossible time keeps the day alone", () => {
    const r = parse("Call tomorrow at 27");
    expect(r.due?.hasTime).toBe(false);
    expect(r.title).toBe("Call at 27");
  });

  it("the last date wins", () => {
    const r = parse("Move from mon to fri");
    expect(ymdhm(r.due!.date)).toEqual([2026, 9, 25, 23, 59]);
    expect(r.title).toBe("Move from mon to");
  });
});

describe("group and person", () => {
  it("resolves #group ignoring case, spaces and dashes", () => {
    const r = parse("Pick the next novel #book-club");
    expect(r.group?.id).toBe("group_book");
    expect(r.title).toBe("Pick the next novel");
  });

  it("resolves @person by first name, prefix or email", () => {
    expect(parse("Send deck @sam").contact?.id).toBe("user_sam");
    expect(parse("Send deck @noor.h").contact?.id).toBe("user_noor");
    expect(parse("Send deck @Riv").contact?.id).toBe("user_sam");
  });

  it("an ambiguous prefix resolves nothing", () => {
    const r = parse("Send deck @sa");
    expect(r.contact).toBeUndefined();
    expect(r.contactQuery).toBe("sa");
    expect(r.title).toBe("Send deck @sa");
  });

  it("an unknown #group stays in the title", () => {
    const r = parse("Tag this #nowhere");
    expect(r.group).toBeUndefined();
    expect(r.groupQuery).toBe("nowhere");
    expect(r.title).toBe("Tag this #nowhere");
    expect(r.tokens[0]?.resolved).toBe(false);
  });

  it("an email address is not a mention", () => {
    expect(parse("Write to sam@example.com").contact).toBeUndefined();
  });
});

describe("everything at once", () => {
  it("parses and cleans the title", () => {
    const r = parse("Call the printer tomorrow at 5pm !high #launch @sam");
    expect(r.title).toBe("Call the printer");
    expect(r.priority).toBe(2);
    expect(r.group?.id).toBe("group_launch");
    expect(r.contact?.id).toBe("user_sam");
    expect(r.tokens.map((t) => t.kind)).toEqual(["due", "priority", "group", "contact"]);
    const raw = r.tokens.map((t) => t.raw);
    expect(raw).toEqual(["tomorrow at 5pm", "!high", "#launch", "@sam"]);
  });

  it("an ignored kind stays text", () => {
    const r = parse("Enjoy the sun !low", new Set(["due"]));
    expect(r.due).toBeUndefined();
    expect(r.title).toBe("Enjoy the sun");
  });
});

describe("helpers", () => {
  it("parseTime", () => {
    expect(parseTime("5pm")).toEqual({ h: 17, m: 0 });
    expect(parseTime("12am")).toEqual({ h: 0, m: 0 });
    expect(parseTime("12pm")).toEqual({ h: 12, m: 0 });
    expect(parseTime("at 17:30")).toEqual({ h: 17, m: 30 });
    expect(parseTime("13pm")).toBeNull();
    expect(parseTime("24:00")).toBeNull();
  });

  it("resolve prefers an exact match", () => {
    expect(resolve("home", groups)?.id).toBe("group_home");
  });

  it("mentionAt finds the word at the caret", () => {
    expect(mentionAt("Plan #lau", 9)).toEqual({ kind: "group", query: "lau", start: 5 });
    expect(mentionAt("Ask @", 5)).toEqual({ kind: "contact", query: "", start: 4 });
    expect(mentionAt("Plan launch", 11)).toBeNull();
  });

  it("mentionText", () => {
    expect(mentionText("group", { id: "g", name: "Book club" })).toBe("#Book-club ");
    expect(mentionText("contact", { id: "u", name: "Sam Rivera" })).toBe("@Sam ");
  });
});
