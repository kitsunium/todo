// One function per route of the contract. Nothing here knows about React.
import { getLocale, isLocale, type Locale } from "../i18n";
import { api, seg } from "./client";
import type {
  ActivityEntry,
  AddContactResult,
  Contact,
  ContactsPayload,
  Counts,
  Group,
  Invitation,
  NewTask,
  Role,
  Session,
  Task,
  TaskPatch,
  TaskView,
  User,
} from "./types";

type UserBody = { user: User };

// The server may send an empty list as null: every list the UI reads is
// normalized here, once, so no component ever meets a null.
const arr = <T,>(v: readonly T[] | null | undefined): T[] => (Array.isArray(v) ? [...v] : []);
const taskIn = (t: Task): Task => ({ ...t, notes: t.notes ?? "", sharedWith: arr(t.sharedWith) });
const groupIn = (g: Group): Group => ({ ...g, members: arr(g.members) });
// A server that predates the locale field answers none: the interface keeps its own.
const userIn = (u: User): User => ({ ...u, locale: isLocale(u.locale) ? u.locale : getLocale() });

export const auth = {
  signup: (b: { email: string; name: string; password: string; locale: Locale }) =>
    api.post<{ status: string; email: string }>("/api/auth/signup", b),
  verify: (token: string) => api.post<UserBody>("/api/auth/verify", { token }).then((r) => userIn(r.user)),
  resend: (email: string) => api.post<{ status: string }>("/api/auth/verify/resend", { email }),
  login: (b: { email: string; password: string }) => api.post<UserBody>("/api/auth/login", b).then((r) => userIn(r.user)),
  logout: () => api.post<void>("/api/auth/logout"),
  me: (signal?: AbortSignal) => api.get<UserBody>("/api/auth/me", { signal }).then((r) => userIn(r.user)),
  /** PATCH /api/auth/me: absent members are unchanged. */
  updateMe: (b: { name?: string; locale?: Locale }) => api.patch<UserBody>("/api/auth/me", b).then((r) => userIn(r.user)),
  changePassword: (b: { current: string; password: string }) => api.post<void>("/api/auth/password", b),
  forgot: (email: string) => api.post<{ status: string }>("/api/auth/password/forgot", { email }),
  reset: (b: { token: string; password: string }) =>
    api.post<UserBody>("/api/auth/password/reset", b).then((r) => userIn(r.user)),
  sessions: () => api.get<{ sessions: Session[] }>("/api/auth/sessions").then((r) => arr(r.sessions)),
  revoke: (id: string) => api.del(`/api/auth/sessions/${seg(id)}`),
};

/** The browser's IANA time zone: where "today" ends for the server. */
export function timeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}

export const tasks = {
  list: (view?: TaskView, group?: string, signal?: AbortSignal) => {
    const q = new URLSearchParams();
    if (view) q.set("view", view);
    if (group) q.set("group", group);
    q.set("tz", timeZone());
    return api.get<{ tasks: Task[] }>(`/api/tasks?${q.toString()}`, { signal }).then((r) => arr(r.tasks).map(taskIn));
  },
  counts: () =>
    api.get<Counts>(`/api/tasks/counts?tz=${encodeURIComponent(timeZone())}`).then((c) => ({ ...c, groups: c.groups ?? {} })),
  get: (id: string) => api.get<Task>(`/api/tasks/${seg(id)}`).then(taskIn),
  create: (b: NewTask) => api.post<Task>("/api/tasks", b).then(taskIn),
  update: (id: string, patch: TaskPatch) => api.patch<Task>(`/api/tasks/${seg(id)}`, patch).then(taskIn),
  complete: (id: string) => api.post<Task>(`/api/tasks/${seg(id)}/complete`).then(taskIn),
  reopen: (id: string) => api.post<Task>(`/api/tasks/${seg(id)}/reopen`).then(taskIn),
  archive: (id: string) => api.post<Task>(`/api/tasks/${seg(id)}/archive`).then(taskIn),
  restore: (id: string) => api.post<Task>(`/api/tasks/${seg(id)}/restore`).then(taskIn),
  remove: (id: string) => api.del(`/api/tasks/${seg(id)}`),
  share: (id: string, userId: string) => api.post<Task>(`/api/tasks/${seg(id)}/share`, { userId }).then(taskIn),
  unshare: (id: string, userId: string) => api.del<Task>(`/api/tasks/${seg(id)}/share/${seg(userId)}`).then(taskIn),
};

export const contacts = {
  list: () =>
    api.get<ContactsPayload>("/api/contacts").then((d) => ({
      contacts: arr(d.contacts),
      incoming: arr(d.incoming),
      outgoing: arr(d.outgoing),
      invites: arr(d.invites),
    })),
  add: (email: string) => api.post<AddContactResult>("/api/contacts", { email }),
  accept: (id: string) => api.post<{ contact: Contact }>(`/api/contacts/${seg(id)}/accept`).then((r) => r.contact),
  decline: (id: string) => api.post<void>(`/api/contacts/${seg(id)}/decline`),
  /** Cancels a pending request — or withdraws an email invitation (invite_… ID). */
  cancel: (id: string) => api.post<void>(`/api/contacts/${seg(id)}/cancel`),
  remove: (id: string) => api.del(`/api/contacts/${seg(id)}`),
};

export const groups = {
  list: () => api.get<{ groups: Group[] }>("/api/groups").then((r) => arr(r.groups).map(groupIn)),
  get: (id: string) => api.get<Group>(`/api/groups/${seg(id)}`).then(groupIn),
  create: (b: { name: string; color?: string }) => api.post<Group>("/api/groups", b).then(groupIn),
  update: (id: string, b: { name?: string; color?: string }) => api.patch<Group>(`/api/groups/${seg(id)}`, b).then(groupIn),
  remove: (id: string) => api.del(`/api/groups/${seg(id)}`),
  invite: (id: string, userId: string) => api.post<Invitation>(`/api/groups/${seg(id)}/invitations`, { userId }),
  removeMember: (id: string, userId: string) => api.del(`/api/groups/${seg(id)}/members/${seg(userId)}`),
  setRole: (id: string, userId: string, role: Exclude<Role, "owner">) =>
    api.patch<Group>(`/api/groups/${seg(id)}/members/${seg(userId)}`, { role }).then(groupIn),
};

export const invitations = {
  list: () => api.get<{ invitations: Invitation[] }>("/api/invitations").then((r) => arr(r.invitations)),
  accept: (id: string) => api.post<Group>(`/api/invitations/${seg(id)}/accept`).then(groupIn),
  decline: (id: string) => api.post<void>(`/api/invitations/${seg(id)}/decline`),
};

export const activity = {
  list: () => api.get<{ entries: ActivityEntry[] }>("/api/activity").then((r) => arr(r.entries)),
  markRead: () => api.post<void>("/api/activity/read"),
};

/** Whether the kit Studio answers: the app runs in dev, with a dev mailbox. */
export async function studioAvailable(): Promise<boolean> {
  try {
    const res = await fetch("/_kit/api/graph", { credentials: "same-origin", headers: { Accept: "application/json" } });
    if (!res.ok) return false;
    const type = res.headers.get("content-type") ?? "";
    await res.body?.cancel();
    return type.includes("json");
  } catch {
    return false;
  }
}
