import { describe, expect, it } from "vitest";
import type { Task } from "../src/api/types";
import { dueAt, formatDue, groupByCompletion, groupByDue, timeAgo } from "../src/lib/dates";

// Thursday, September 24 2026, 10:00 local time.
const NOW = new Date(2026, 8, 24, 10, 0);
const me = { id: "user_me", name: "Me", email: "me@example.com" };
let n = 0;
function task(p: Partial<Task>): Task {
  n++;
  return {
    id: `task_${n}`,
    title: `Task ${n}`,
    notes: "",
    priority: 0,
    status: "open",
    owner: me,
    sharedWith: [],
    createdAt: new Date(2026, 8, 1, 9, n).toISOString(),
    updatedAt: new Date(2026, 8, 1, 9, n).toISOString(),
    can: { edit: true, delete: true, share: true },
    ...p,
  };
}
const day = (d: number, h?: number, m = 0) =>
  (h === undefined ? dueAt(new Date(2026, 8, d)) : new Date(2026, 8, d, h, m)).toISOString();

describe("groupByDue", () => {
  it("puts every task in its section, in order", () => {
    const list = [
      task({ title: "none" }),
      task({ title: "later", due: day(30) }),
      task({ title: "week", due: day(27) }),
      task({ title: "tomorrow", due: day(25, 9) }),
      task({ title: "today", due: day(24) }),
      task({ title: "overdue", due: day(22), status: "overdue" }),
      task({ title: "passed this morning", due: day(24, 8) }),
      task({ title: "done", status: "done", completedAt: day(24, 9) }),
      task({ title: "archived", status: "archived" }),
    ];
    const sections = groupByDue(list, NOW);
    expect(sections.map((s) => [s.key, s.tasks.map((t) => t.title)])).toEqual([
      ["overdue", ["overdue", "passed this morning"]],
      ["today", ["today"]],
      ["tomorrow", ["tomorrow"]],
      ["week", ["week"]],
      ["later", ["later"]],
      ["none", ["none"]],
      ["completed", ["done"]],
    ]);
    expect(sections[0]!.tone).toBe("danger");
  });

  it("sorts by due, then priority, urgent first and none last", () => {
    const list = [
      task({ title: "b-none", due: day(24, 15), priority: 0 }),
      task({ title: "a-low", due: day(24, 15), priority: 4 }),
      task({ title: "c-urgent", due: day(24, 15), priority: 1 }),
      task({ title: "early", due: day(24, 11), priority: 4 }),
    ];
    expect(groupByDue(list, NOW)[0]!.tasks.map((t) => t.title)).toEqual(["early", "c-urgent", "a-low", "b-none"]);
  });

  it("undated tasks: priority first, newest first", () => {
    const list = [task({ title: "old" }), task({ title: "new" }), task({ title: "high", priority: 2 })];
    expect(groupByDue(list, NOW)[0]!.tasks.map((t) => t.title)).toEqual(["high", "new", "old"]);
  });

  it("the week ends on Sunday", () => {
    const sunday = task({ title: "sun", due: day(27) });
    const monday = task({ title: "mon", due: day(28) });
    const s = groupByDue([sunday, monday], NOW);
    expect(s.map((x) => x.key)).toEqual(["week", "later"]);
  });
});

describe("groupByCompletion", () => {
  it("groups by day, then Earlier", () => {
    const list = [
      task({ title: "t", status: "done", completedAt: day(24, 9) }),
      task({ title: "y", status: "done", completedAt: day(23, 18) }),
      task({ title: "mon", status: "done", completedAt: day(21, 12) }),
      task({ title: "old", status: "done", completedAt: day(2, 12) }),
    ];
    expect(groupByCompletion(list, NOW).map((s) => s.label)).toEqual(["Today", "Yesterday", "Monday", "Earlier"]);
  });
});

describe("formatDue", () => {
  it("speaks like a person", () => {
    expect(formatDue(day(24), NOW)).toMatchObject({ label: "Today", overdue: false, soon: true });
    expect(formatDue(day(25), NOW).label).toBe("Tomorrow");
    expect(formatDue(day(26), NOW).label).toBe("Sat");
    expect(formatDue(day(23), NOW)).toMatchObject({ label: "Yesterday", overdue: true });
    expect(formatDue(new Date(2026, 9, 3).toISOString(), NOW).label).toBe("Oct 3");
    expect(formatDue(new Date(2027, 0, 3).toISOString(), NOW).label).toBe("Jan 3, 2027");
    expect(formatDue(day(25, 17), NOW).label).toMatch(/^Tomorrow \d/);
  });
});

describe("timeAgo", () => {
  it("is short", () => {
    expect(timeAgo(new Date(2026, 8, 24, 9, 59, 40).toISOString(), NOW)).toBe("just now");
    expect(timeAgo(new Date(2026, 8, 24, 9, 55).toISOString(), NOW)).toBe("5m ago");
    expect(timeAgo(new Date(2026, 8, 24, 7, 0).toISOString(), NOW)).toBe("3h ago");
    expect(timeAgo(new Date(2026, 8, 23, 7, 0).toISOString(), NOW)).toBe("Yesterday");
  });
});

describe("pins", () => {
  it("keep a just-completed task in its section", () => {
    const t = task({ title: "was overdue", due: day(22), status: "done", completedAt: day(24, 9) });
    expect(groupByDue([t], NOW).map((s) => s.key)).toEqual(["completed"]);
    expect(groupByDue([t], NOW, { [t.id]: "overdue" }).map((s) => s.key)).toEqual(["overdue"]);
  });
});
