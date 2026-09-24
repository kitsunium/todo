// The HTTP contract of the todo backend (plan "The todo — domain and HTTP
// contract"). JSON is camelCase, timestamps are RFC 3339 UTC, IDs are kit
// TypeIDs (user_…, task_…, group_…).

export type UserRef = { id: string; name: string; email: string };
export type User = UserRef & { createdAt: string };

/** 0 none, 1 urgent, 2 high, 3 medium, 4 low — Linear's order. */
export type Priority = 0 | 1 | 2 | 3 | 4;
export const PRIORITIES: readonly Priority[] = [1, 2, 3, 4, 0];

export type Status = "open" | "overdue" | "done" | "archived";

export const GROUP_COLORS = [
  "slate",
  "red",
  "orange",
  "amber",
  "green",
  "teal",
  "blue",
  "indigo",
  "violet",
  "pink",
] as const;
export type GroupColor = (typeof GROUP_COLORS)[number];

export type GroupRef = { id: string; name: string; color: string };

export type Task = {
  id: string;
  title: string;
  notes: string;
  priority: Priority;
  status: Status;
  due?: string;
  owner: UserRef;
  assignee?: UserRef;
  group?: GroupRef;
  sharedWith: UserRef[];
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
  completedBy?: UserRef;
  can: { edit: boolean; delete: boolean; share: boolean };
};

export type TaskView = "inbox" | "today" | "upcoming" | "shared" | "assigned" | "completed" | "group";

export type Counts = {
  inbox: number;
  today: number;
  upcoming: number;
  shared: number;
  assigned: number;
  overdue: number;
  completedThisWeek: number;
  groups: Record<string, number>;
  unread: number;
};

export type NewTask = {
  title: string;
  notes?: string;
  priority?: Priority;
  due?: string;
  groupId?: string;
  assigneeId?: string;
};

/** A PATCH: absent members are unchanged; "" clears due, groupId and assigneeId. */
export type TaskPatch = {
  title?: string;
  notes?: string;
  priority?: Priority;
  due?: string;
  groupId?: string;
  assigneeId?: string;
};

export type Session = {
  id: string;
  createdAt: string;
  lastSeenAt: string;
  userAgent: string;
  ip: string;
  current: boolean;
};

export type ContactRequest = { id: string; user: UserRef; createdAt: string; status: string };
export type Contact = { id: string; user: UserRef; since: string };
export type Invite = { id: string; email: string; createdAt: string; status: string };
export type ContactsPayload = {
  contacts: Contact[];
  incoming: ContactRequest[];
  outgoing: ContactRequest[];
  invites: Invite[];
};
export type AddContactResult =
  | { kind: "request"; request: ContactRequest }
  | { kind: "invite"; invite: Invite };

export type Role = "owner" | "admin" | "member";
export type Member = { user: UserRef; role: Role; joinedAt: string };
export type Group = {
  id: string;
  name: string;
  color: string;
  role: Role;
  members: Member[];
  createdAt: string;
  openTasks: number;
};
export type Invitation = {
  id: string;
  group: GroupRef;
  inviter: UserRef;
  /** Present on the invitation a group owner or admin just sent. */
  invitee?: UserRef;
  createdAt: string;
  status: string;
};

export type ActivityEntry = {
  id: string;
  kind: string;
  actor?: UserRef;
  task?: { id: string; title: string };
  group?: GroupRef;
  text: string;
  at: string;
  read: boolean;
};

export type Violation = { path: string; rule: string; message: string };
export type ErrorBody = { error: { code: string; message: string; violations?: Violation[] } };
