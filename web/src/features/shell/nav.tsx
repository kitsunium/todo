import { CalendarDays, CircleCheck, Inbox, Star, UserRoundCheck, UsersRound, type LucideIcon } from "lucide-react";
import type { Counts, TaskView } from "../../api/types";

export type ViewDef = {
  view: Exclude<TaskView, "group">;
  path: string;
  label: string;
  icon: LucideIcon;
  keys: string[];
  count: (c: Counts) => number | undefined;
  blurb: string;
  /** The view's own color: its icon in the page header and when active. */
  color: string;
};

export const VIEWS: ViewDef[] = [
  { view: "inbox", path: "/app/inbox", label: "Inbox", icon: Inbox, keys: ["G", "I"], count: (c) => c.inbox, blurb: "Tasks that belong to no group yet.", color: "#3e63dd" },
  { view: "today", path: "/app/today", label: "Today", icon: Star, keys: ["G", "T"], count: (c) => c.today, blurb: "", color: "#f5a524" },
  { view: "upcoming", path: "/app/upcoming", label: "Upcoming", icon: CalendarDays, keys: ["G", "U"], count: (c) => c.upcoming, blurb: "Everything with a date after today.", color: "#e5484d" },
  { view: "shared", path: "/app/shared", label: "Shared with me", icon: UsersRound, keys: ["G", "S"], count: (c) => c.shared, blurb: "Tasks other people shared with you.", color: "#12a594" },
  { view: "assigned", path: "/app/assigned", label: "Assigned to me", icon: UserRoundCheck, keys: ["G", "A"], count: (c) => c.assigned, blurb: "Tasks someone handed to you — or you took on.", color: "#8e4ec6" },
  { view: "completed", path: "/app/completed", label: "Completed", icon: CircleCheck, keys: ["G", "C"], count: () => undefined, blurb: "", color: "#30a46c" },
];
