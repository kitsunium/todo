import type { Task, TaskView } from "../../api/types";

/**
 * Whether a task belongs on the screen of a view, by status — the server's
 * rule: the to-do views list open and overdue tasks, Completed lists done and
 * archived ones, a group (and the default list) hides only the archived.
 */
export function viewAccepts(view: TaskView | undefined, t: Task): boolean {
  switch (view) {
    case "inbox":
    case "today":
    case "upcoming":
    case "shared":
    case "assigned":
      return t.status === "open" || t.status === "overdue";
    case "completed":
      return t.status === "done" || t.status === "archived";
    default:
      return t.status !== "archived";
  }
}
