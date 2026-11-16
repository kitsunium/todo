import {
  CalendarDays,
  CircleCheck,
  Inbox,
  Star,
  UserRoundCheck,
  UsersRound,
  type LucideIcon,
} from "lucide-react";
import type { Counts, TaskView } from "../../api/types";
import type { Key } from "../../i18n";

export type ViewDef = {
  view: Exclude<TaskView, "group">;
  path: string;
  /** The view's name, a dictionary key: "Aujourd’hui", "Today". */
  label: Key;
  icon: LucideIcon;
  /** "G" then a letter. The letters stay the same in every language: they are muscle memory. */
  keys: string[];
  count: (c: Counts) => number | undefined;
  blurb?: Key;
  /** The view's own color: its icon in the page header and when active. */
  color: string;
};

export const VIEWS: ViewDef[] = [
  {
    view: "inbox",
    path: "/app/inbox",
    label: "view.inbox",
    icon: Inbox,
    keys: ["G", "I"],
    count: (c) => c.inbox,
    blurb: "view.inbox.blurb",
    color: "#3e63dd",
  },
  {
    view: "today",
    path: "/app/today",
    label: "view.today",
    icon: Star,
    keys: ["G", "T"],
    count: (c) => c.today,
    color: "#f5a524",
  },
  {
    view: "upcoming",
    path: "/app/upcoming",
    label: "view.upcoming",
    icon: CalendarDays,
    keys: ["G", "U"],
    count: (c) => c.upcoming,
    blurb: "view.upcoming.blurb",
    color: "#e5484d",
  },
  {
    view: "shared",
    path: "/app/shared",
    label: "view.shared",
    icon: UsersRound,
    keys: ["G", "S"],
    count: (c) => c.shared,
    blurb: "view.shared.blurb",
    color: "#12a594",
  },
  {
    view: "assigned",
    path: "/app/assigned",
    label: "view.assigned",
    icon: UserRoundCheck,
    keys: ["G", "A"],
    count: (c) => c.assigned,
    blurb: "view.assigned.blurb",
    color: "#8e4ec6",
  },
  {
    view: "completed",
    path: "/app/completed",
    label: "view.completed",
    icon: CircleCheck,
    keys: ["G", "C"],
    count: () => undefined,
    color: "#30a46c",
  },
];
