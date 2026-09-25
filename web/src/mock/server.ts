// The HTTP contract, in memory: every /api route the backend serves, with
// the same shapes, codes and rules (as far as a mock can tell). Requests
// reach it through the fetch interceptor in install.ts.
import { endOfDay, startOfWeek } from "date-fns";
import type { Priority, User, UserRef } from "../api/types";
import type { Locale } from "../i18n/types";
import { load, save, token, typeid, type DB, type MGroup, type MTask, type MUser } from "./db";

export class MockError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
    readonly violations?: { path: string; rule: string; message: string }[],
  ) {
    super(message);
  }
}

type Ctx = { params: Record<string, string>; query: URLSearchParams; body: Record<string, unknown>; db: DB; me: MUser | null };
type Result = { status: number; body?: unknown };
type Handler = (c: Ctx) => Result;

const routes: { method: string; parts: string[]; handler: Handler }[] = [];
function route(method: string, pattern: string, handler: Handler) {
  routes.push({ method, parts: pattern.split("/").filter(Boolean), handler });
}

const ok = (body?: unknown): Result => (body === undefined ? { status: 204 } : { status: 200, body });
const invalid = (path: string, rule: string, message: string) =>
  new MockError(400, "invalid_argument", "the request is invalid", [{ path, rule, message }]);

function need(c: Ctx): MUser {
  if (!c.me) throw new MockError(401, "unauthenticated", "sign in to continue");
  return c.me;
}

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
function str(c: Ctx, key: string): string {
  const v = c.body[key];
  return typeof v === "string" ? v : "";
}
function email(c: Ctx): string {
  const e = str(c, "email").trim().toLowerCase();
  if (!EMAIL.test(e)) throw invalid("email", "email", "must be an email address");
  return e;
}
function password(c: Ctx, key = "password"): string {
  const p = str(c, key);
  const n = Array.from(p).length;
  if (n < 10) throw invalid(key, "minlen", "must be at least 10 characters");
  if (n > 128) throw invalid(key, "maxlen", "must be at most 128 characters");
  return p;
}

// ── Projections ───────────────────────────────────────────────────────

const ref = (u: MUser): UserRef => ({ id: u.id, name: u.name, email: u.email });
const self = (u: MUser): User => ({ ...ref(u), createdAt: u.createdAt, locale: u.locale });

/** An optional "locale" member: exactly "fr" or "en", as the server wants it. */
function localeOf(c: Ctx): Locale | undefined {
  if (!("locale" in c.body)) return undefined;
  const l = c.body.locale;
  if (l !== "fr" && l !== "en") throw invalid("locale", "one_of", "must be one of: fr, en");
  return l;
}

// The mails the mock "sends", in the language of the account they go to.
const SUBJECTS: Record<"verify" | "reset" | "exists" | "invite", Record<Locale, (name: string) => string>> = {
  verify: { fr: () => "Confirmez votre adresse e-mail", en: () => "Confirm your email address" },
  reset: { fr: () => "Réinitialisez votre mot de passe Todo", en: () => "Reset your Todo password" },
  exists: { fr: () => "Vous avez déjà un compte Todo", en: () => "You already have a Todo account" },
  invite: { fr: (n) => `${n} vous invite sur Todo`, en: (n) => `${n} invited you to Todo` },
};
function user(db: DB, id: string): MUser {
  const u = db.users.find((x) => x.id === id);
  if (!u) throw new MockError(404, "not_found", "no such user");
  return u;
}
function roleIn(g: MGroup | undefined, uid: string) {
  return g?.members.find((m) => m.userId === uid)?.role ?? "";
}
function visible(db: DB, t: MTask, uid: string): boolean {
  if (t.ownerId === uid || t.sharedWith.includes(uid)) return true;
  const g = db.groups.find((x) => x.id === t.groupId);
  return !!roleIn(g, uid);
}
function statusOf(t: MTask, now = Date.now()) {
  if (t.status === "open" && t.due && Date.parse(t.due) < now) return "overdue" as const;
  return t.status;
}
function taskOut(db: DB, t: MTask, uid: string) {
  const g = db.groups.find((x) => x.id === t.groupId);
  const role = roleIn(g, uid);
  return {
    id: t.id,
    title: t.title,
    notes: t.notes,
    priority: t.priority,
    status: statusOf(t),
    ...(t.due ? { due: t.due } : {}),
    owner: ref(user(db, t.ownerId)),
    ...(t.assigneeId ? { assignee: ref(user(db, t.assigneeId)) } : {}),
    ...(g ? { group: { id: g.id, name: g.name, color: g.color } } : {}),
    sharedWith: t.sharedWith.map((id) => ref(user(db, id))),
    createdAt: t.createdAt,
    updatedAt: t.updatedAt,
    ...(t.completedAt ? { completedAt: t.completedAt } : {}),
    ...(t.completedById ? { completedBy: ref(user(db, t.completedById)) } : {}),
    can: {
      edit: true,
      delete: t.ownerId === uid || role === "owner" || role === "admin",
      share: t.ownerId === uid,
    },
  };
}
function openCount(db: DB, g: MGroup) {
  return db.tasks.filter((t) => t.groupId === g.id && (t.status === "open")).length;
}
function groupOut(db: DB, g: MGroup, uid: string) {
  return {
    id: g.id,
    name: g.name,
    color: g.color,
    role: roleIn(g, uid),
    members: g.members.map((m) => ({ user: ref(user(db, m.userId)), role: m.role, joinedAt: m.joinedAt })),
    createdAt: g.createdAt,
    openTasks: openCount(db, g),
  };
}
function areContacts(db: DB, a: string, b: string) {
  return db.links.some((l) => (l.a === a && l.b === b) || (l.a === b && l.b === a));
}
function task(c: Ctx, uid: string): MTask {
  const t = c.db.tasks.find((x) => x.id === c.params.id);
  if (!t || !visible(c.db, t, uid)) throw new MockError(404, "not_found", "no such task");
  return t;
}
function group(c: Ctx, uid: string, id = c.params.id): MGroup {
  const g = c.db.groups.find((x) => x.id === id);
  if (!g || !roleIn(g, uid)) throw new MockError(404, "not_found", "no such group");
  return g;
}
function mail(db: DB, to: string, kind: "verify" | "reset" | "exists" | "invite", subject: string, link?: string) {
  db.mails.unshift({ id: typeid("mail"), to, kind, subject, ...(link ? { link } : {}), at: new Date().toISOString() });
}
function signIn(db: DB, u: MUser): void {
  const s = {
    id: typeid("session"),
    token: token(),
    userId: u.id,
    createdAt: new Date().toISOString(),
    lastSeenAt: new Date().toISOString(),
    userAgent: navigator.userAgent,
    ip: "127.0.0.1",
  };
  db.sessions.push(s);
  db.cookie = s.token;
}

// Verification and reset links, single use.
function issue(db: DB, userId: string, kind: "verify" | "reset", ttlHours: number): string {
  const t = token();
  db.linkTokens.push({ token: t, userId, kind, expires: Date.now() + ttlHours * 3600e3 });
  return t;
}
function redeem(db: DB, t: string, kind: "verify" | "reset") {
  const found = db.linkTokens.find((x) => x.token === t && x.kind === kind && x.expires > Date.now());
  if (!found) throw new MockError(400, "invalid_token", "this link is invalid or has expired: ask for a new one");
  return found;
}
function spend(db: DB, t: string) {
  db.linkTokens = db.linkTokens.filter((x) => x.token !== t);
}

// ── identity ──────────────────────────────────────────────────────────

route("POST", "/api/auth/signup", (c) => {
  const e = email(c);
  const name = str(c, "name").trim();
  if (!name) throw invalid("name", "required", "is required");
  if (name.length > 80) throw invalid("name", "maxlen", "must be at most 80 characters");
  const pw = password(c);
  const locale = localeOf(c) ?? "fr";
  const existing = c.db.users.find((u) => u.email === e);
  if (existing) mail(c.db, e, "exists", SUBJECTS.exists[existing.locale](""));
  else {
    const u: MUser = { id: typeid("user"), name, email: e, locale, password: pw, verified: false, locked: false, failed: 0, createdAt: new Date().toISOString() };
    c.db.users.push(u);
    const t = issue(c.db, u.id, "verify", 24);
    mail(c.db, e, "verify", SUBJECTS.verify[locale](""), `/verify?token=${t}`);
  }
  return ok({ status: "verification_sent", email: e });
});

route("POST", "/api/auth/verify", (c) => {
  const t = redeem(c.db, str(c, "token"), "verify");
  spend(c.db, t.token);
  const u = user(c.db, t.userId);
  u.verified = true;
  signIn(c.db, u);
  // Pending invitations to this address become contact requests.
  for (const inv of c.db.invites.filter((i) => i.email === u.email && i.status === "pending")) {
    inv.status = "accepted";
    c.db.requests.push({ id: typeid("request"), from: inv.fromId, to: u.id, createdAt: new Date().toISOString(), status: "pending" });
  }
  return ok({ user: self(u) });
});

route("POST", "/api/auth/verify/resend", (c) => {
  const e = email(c);
  const u = c.db.users.find((x) => x.email === e && !x.verified);
  if (u) {
    const t = issue(c.db, u.id, "verify", 24);
    mail(c.db, e, "verify", SUBJECTS.verify[u.locale](""), `/verify?token=${t}`);
  }
  return ok({ status: "sent" });
});

route("POST", "/api/auth/login", (c) => {
  const e = email(c);
  const pw = str(c, "password");
  const u = c.db.users.find((x) => x.email === e);
  if (u?.locked) throw new MockError(403, "account_locked", "too many failed attempts: try again later or reset your password");
  if (!u || u.password !== pw) {
    if (u) {
      u.failed++;
      if (u.failed >= 5) u.locked = true;
    }
    throw new MockError(401, "invalid_credentials", "the email or the password is wrong");
  }
  if (!u.verified) throw new MockError(403, "email_unverified", "confirm your email address first");
  u.failed = 0;
  signIn(c.db, u);
  return ok({ user: self(u) });
});

route("POST", "/api/auth/logout", (c) => {
  need(c);
  c.db.sessions = c.db.sessions.filter((s) => s.token !== c.db.cookie);
  delete c.db.cookie;
  return ok();
});

route("GET", "/api/auth/me", (c) => {
  const me = need(c);
  return ok({ user: self(me) });
});

route("PATCH", "/api/auth/me", (c) => {
  const me = need(c);
  // {name?, locale?}: a member left out stays as it is.
  if ("name" in c.body) {
    const name = str(c, "name").trim();
    if (!name) throw invalid("name", "required", "is required");
    if (name.length > 80) throw invalid("name", "maxlen", "must be at most 80 characters");
    me.name = name;
  }
  const locale = localeOf(c);
  if (locale) me.locale = locale;
  return ok({ user: self(me) });
});

route("POST", "/api/auth/password", (c) => {
  const me = need(c);
  if (str(c, "current") !== me.password) throw invalid("current", "match", "does not match your current password");
  me.password = password(c);
  c.db.sessions = c.db.sessions.filter((s) => s.userId !== me.id || s.token === c.db.cookie);
  return ok();
});

route("POST", "/api/auth/password/forgot", (c) => {
  const e = email(c);
  const u = c.db.users.find((x) => x.email === e);
  if (u) {
    const t = issue(c.db, u.id, "reset", 1);
    mail(c.db, e, "reset", SUBJECTS.reset[u.locale](""), `/reset?token=${t}`);
  }
  return ok({ status: "sent" });
});

route("POST", "/api/auth/password/reset", (c) => {
  const t = redeem(c.db, str(c, "token"), "reset");
  const pw = password(c);
  spend(c.db, t.token);
  const u = user(c.db, t.userId);
  u.password = pw;
  u.locked = false;
  u.failed = 0;
  u.verified = true;
  c.db.sessions = c.db.sessions.filter((s) => s.userId !== u.id);
  signIn(c.db, u);
  return ok({ user: self(u) });
});

route("GET", "/api/auth/sessions", (c) => {
  const me = need(c);
  const sessions = c.db.sessions
    .filter((s) => s.userId === me.id)
    .map((s) => ({ id: s.id, createdAt: s.createdAt, lastSeenAt: s.lastSeenAt, userAgent: s.userAgent, ip: s.ip, current: s.token === c.db.cookie }))
    .sort((a, b) => Number(b.current) - Number(a.current) || b.lastSeenAt.localeCompare(a.lastSeenAt));
  return ok({ sessions });
});

route("DELETE", "/api/auth/sessions/:id", (c) => {
  const me = need(c);
  const s = c.db.sessions.find((x) => x.id === c.params.id && x.userId === me.id);
  if (!s) throw new MockError(404, "not_found", "no such session");
  c.db.sessions = c.db.sessions.filter((x) => x !== s);
  if (s.token === c.db.cookie) delete c.db.cookie;
  return ok();
});

// ── tasks ─────────────────────────────────────────────────────────────

route("GET", "/api/tasks", (c) => {
  const me = need(c);
  const view = c.query.get("view") ?? "";
  const gid = c.query.get("group") ?? "";
  const eod = endOfDay(Date.now()).getTime();
  const mine = c.db.tasks.filter((t) => visible(c.db, t, me.id));
  const active = (t: MTask) => t.status === "open";
  let list: MTask[];
  switch (view) {
    case "inbox":
      list = mine.filter((t) => t.ownerId === me.id && !t.groupId && active(t));
      break;
    case "today":
      list = mine.filter((t) => active(t) && t.due && Date.parse(t.due) <= eod);
      break;
    case "upcoming":
      list = mine.filter((t) => active(t) && t.due && Date.parse(t.due) > eod);
      break;
    case "shared":
      list = mine.filter((t) => t.sharedWith.includes(me.id) && active(t));
      break;
    case "assigned":
      list = mine.filter((t) => t.assigneeId === me.id && active(t));
      break;
    case "completed":
      list = mine
        .filter((t) => t.status === "done" || t.status === "archived")
        .sort((a, b) => (b.completedAt ?? "").localeCompare(a.completedAt ?? ""));
      break;
    case "group":
      group(c, me.id, gid);
      list = mine.filter((t) => t.groupId === gid && t.status !== "archived");
      break;
    case "":
      list = mine.filter((t) => t.status !== "archived");
      break;
    default:
      throw invalid("view", "oneof", "must be one of inbox, today, upcoming, shared, assigned, completed, group");
  }
  return ok({ tasks: list.map((t) => taskOut(c.db, t, me.id)) });
});

route("GET", "/api/tasks/counts", (c) => {
  const me = need(c);
  const eod = endOfDay(Date.now()).getTime();
  const weekStart = startOfWeek(Date.now(), { weekStartsOn: 1 }).getTime();
  const mine = c.db.tasks.filter((t) => visible(c.db, t, me.id));
  const open = mine.filter((t) => t.status === "open");
  const groups: Record<string, number> = {};
  for (const g of c.db.groups.filter((x) => roleIn(x, me.id))) groups[g.id] = open.filter((t) => t.groupId === g.id).length;
  return ok({
    inbox: open.filter((t) => t.ownerId === me.id && !t.groupId).length,
    today: open.filter((t) => t.due && Date.parse(t.due) <= eod).length,
    upcoming: open.filter((t) => t.due && Date.parse(t.due) > eod).length,
    shared: mine.filter((t) => t.sharedWith.includes(me.id) && t.status === "open").length,
    assigned: open.filter((t) => t.assigneeId === me.id).length,
    overdue: open.filter((t) => statusOf(t) === "overdue").length,
    completedThisWeek: mine.filter((t) => t.status === "done" && t.completedAt && Date.parse(t.completedAt) >= weekStart).length,
    groups,
    unread: c.db.activity.filter((a) => a.userId === me.id && !a.read).length,
  });
});

function checkAssignee(db: DB, t: MTask, uid: string) {
  const g = db.groups.find((x) => x.id === t.groupId);
  if (uid !== t.ownerId && !t.sharedWith.includes(uid) && !roleIn(g, uid)) {
    throw invalid("assigneeId", "assignable", "must be the owner, someone the task is shared with, or a member of its group");
  }
}
function title(c: Ctx, required: boolean): string | undefined {
  if (!("title" in c.body)) {
    if (required) throw invalid("title", "required", "is required");
    return undefined;
  }
  const t = str(c, "title").trim();
  if (!t) throw invalid("title", "required", "is required");
  if (t.length > 200) throw invalid("title", "maxlen", "must be at most 200 characters");
  return t;
}
function priority(c: Ctx): Priority | undefined {
  if (!("priority" in c.body)) return undefined;
  const p = c.body.priority;
  if (typeof p !== "number" || ![0, 1, 2, 3, 4].includes(p)) throw invalid("priority", "oneof", "must be 0, 1, 2, 3 or 4");
  return p as Priority;
}

route("POST", "/api/tasks", (c) => {
  const me = need(c);
  const now = new Date().toISOString();
  const t: MTask = {
    id: typeid("task"),
    title: title(c, true)!,
    notes: str(c, "notes"),
    priority: priority(c) ?? 0,
    status: "open",
    ownerId: me.id,
    sharedWith: [],
    createdAt: now,
    updatedAt: now,
  };
  const due = str(c, "due");
  if (due) {
    if (Number.isNaN(Date.parse(due))) throw invalid("due", "time", "must be an RFC 3339 time");
    t.due = new Date(due).toISOString();
  }
  const gid = str(c, "groupId");
  if (gid) t.groupId = group(c, me.id, gid).id;
  const aid = str(c, "assigneeId");
  if (aid) {
    checkAssignee(c.db, t, aid);
    t.assigneeId = aid;
  }
  c.db.tasks.push(t);
  return ok(taskOut(c.db, t, me.id));
});

route("GET", "/api/tasks/:id", (c) => {
  const me = need(c);
  return ok(taskOut(c.db, task(c, me.id), me.id));
});

route("PATCH", "/api/tasks/:id", (c) => {
  const me = need(c);
  const t = task(c, me.id);
  const tt = title(c, false);
  if (tt !== undefined) t.title = tt;
  if ("notes" in c.body) t.notes = str(c, "notes");
  const p = priority(c);
  if (p !== undefined) t.priority = p;
  if ("due" in c.body) {
    const due = str(c, "due");
    if (!due) delete t.due;
    else if (Number.isNaN(Date.parse(due))) throw invalid("due", "time", "must be an RFC 3339 time");
    else t.due = new Date(due).toISOString();
  }
  if ("groupId" in c.body && str(c, "groupId") !== (t.groupId ?? "")) {
    if (t.ownerId !== me.id) throw new MockError(403, "permission_denied", "only the owner can move this task to another group");
    const gid = str(c, "groupId");
    if (!gid) delete t.groupId;
    else t.groupId = group(c, me.id, gid).id;
  }
  if ("assigneeId" in c.body) {
    const aid = str(c, "assigneeId");
    if (!aid) delete t.assigneeId;
    else {
      checkAssignee(c.db, t, aid);
      t.assigneeId = aid;
    }
  }
  t.updatedAt = new Date().toISOString();
  return ok(taskOut(c.db, t, me.id));
});

function transition(c: Ctx, from: string[], to: MTask["status"], event: string): Result {
  const me = need(c);
  const t = task(c, me.id);
  if (!from.includes(t.status)) throw new MockError(409, "conflict", `a ${statusOf(t)} task cannot ${event}`);
  t.status = to;
  t.updatedAt = new Date().toISOString();
  if (to === "done") {
    t.completedAt = t.updatedAt;
    t.completedById = me.id;
  } else if (to === "open") {
    delete t.completedAt;
    delete t.completedById;
  }
  return ok(taskOut(c.db, t, me.id));
}
route("POST", "/api/tasks/:id/complete", (c) => transition(c, ["open"], "done", "complete"));
route("POST", "/api/tasks/:id/reopen", (c) => transition(c, ["done"], "open", "reopen"));
route("POST", "/api/tasks/:id/archive", (c) => transition(c, ["done"], "archived", "archive"));
route("POST", "/api/tasks/:id/restore", (c) => transition(c, ["archived"], "open", "restore"));

route("DELETE", "/api/tasks/:id", (c) => {
  const me = need(c);
  const t = task(c, me.id);
  if (!taskOut(c.db, t, me.id).can.delete) throw new MockError(403, "permission_denied", "only the owner or a group admin can delete this task");
  c.db.tasks = c.db.tasks.filter((x) => x !== t);
  return ok();
});

route("POST", "/api/tasks/:id/share", (c) => {
  const me = need(c);
  const t = task(c, me.id);
  if (t.ownerId !== me.id) throw new MockError(403, "permission_denied", "only the owner can share this task");
  const uid = str(c, "userId");
  if (!areContacts(c.db, me.id, uid)) throw new MockError(403, "permission_denied", "you can only share with your contacts");
  if (!t.sharedWith.includes(uid)) t.sharedWith.push(uid);
  t.updatedAt = new Date().toISOString();
  return ok(taskOut(c.db, t, me.id));
});

route("DELETE", "/api/tasks/:id/share/:userId", (c) => {
  const me = need(c);
  const t = task(c, me.id);
  const uid = c.params.userId!;
  if (t.ownerId !== me.id && uid !== me.id) throw new MockError(403, "permission_denied", "only the owner can stop sharing this task");
  t.sharedWith = t.sharedWith.filter((x) => x !== uid);
  if (t.assigneeId === uid) delete t.assigneeId;
  t.updatedAt = new Date().toISOString();
  return ok(taskOut(c.db, t, me.id));
});

// ── contacts ──────────────────────────────────────────────────────────

route("GET", "/api/contacts", (c) => {
  const me = need(c);
  const db = c.db;
  const other = (a: string, b: string) => (a === me.id ? b : a);
  return ok({
    contacts: db.links
      .filter((l) => l.a === me.id || l.b === me.id)
      .map((l) => ({ id: l.id, user: ref(user(db, other(l.a, l.b))), since: l.since }))
      .sort((x, y) => x.user.name.localeCompare(y.user.name)),
    incoming: db.requests
      .filter((r) => r.to === me.id && r.status === "pending")
      .map((r) => ({ id: r.id, user: ref(user(db, r.from)), createdAt: r.createdAt, status: r.status })),
    outgoing: db.requests
      .filter((r) => r.from === me.id && r.status === "pending")
      .map((r) => ({ id: r.id, user: ref(user(db, r.to)), createdAt: r.createdAt, status: r.status })),
    invites: db.invites
      .filter((i) => i.fromId === me.id && i.status === "pending")
      .map((i) => ({ id: i.id, email: i.email, createdAt: i.createdAt, status: i.status })),
  });
});

route("POST", "/api/contacts", (c) => {
  const me = need(c);
  const e = email(c);
  if (e === me.email) throw invalid("email", "self", "is your own address");
  const u = c.db.users.find((x) => x.email === e && x.verified);
  const now = new Date().toISOString();
  if (u) {
    if (areContacts(c.db, me.id, u.id)) throw new MockError(409, "conflict", `you and ${u.name} are already contacts`);
    if (c.db.requests.some((r) => r.status === "pending" && ((r.from === me.id && r.to === u.id) || (r.from === u.id && r.to === me.id)))) {
      throw new MockError(409, "conflict", "a request between you is already pending");
    }
    const r = { id: typeid("request"), from: me.id, to: u.id, createdAt: now, status: "pending" };
    c.db.requests.push(r);
    return ok({ kind: "request", request: { id: r.id, user: ref(u), createdAt: now, status: "pending" } });
  }
  if (c.db.invites.some((i) => i.fromId === me.id && i.email === e && i.status === "pending")) {
    throw new MockError(409, "conflict", "you already invited this address");
  }
  const inv = { id: typeid("invite"), fromId: me.id, email: e, createdAt: now, status: "pending" };
  c.db.invites.push(inv);
  // The invitee has no account yet: the invitation speaks the inviter's language.
  mail(c.db, e, "invite", SUBJECTS.invite[me.locale](me.name), "/signup");
  return ok({ kind: "invite", invite: { id: inv.id, email: e, createdAt: now, status: "pending" } });
});

function request(c: Ctx) {
  const r = c.db.requests.find((x) => x.id === c.params.id && x.status === "pending");
  if (!r) throw new MockError(404, "not_found", "no such request");
  return r;
}
route("POST", "/api/contacts/:id/accept", (c) => {
  const me = need(c);
  const r = request(c);
  if (r.to !== me.id) throw new MockError(403, "permission_denied", "only the addressee can accept");
  r.status = "accepted";
  const l = { id: typeid("contact"), a: r.from, b: r.to, since: new Date().toISOString() };
  c.db.links.push(l);
  return ok({ contact: { id: l.id, user: ref(user(c.db, r.from)), since: l.since } });
});
route("POST", "/api/contacts/:id/decline", (c) => {
  const me = need(c);
  const r = request(c);
  if (r.to !== me.id) throw new MockError(403, "permission_denied", "only the addressee can decline");
  r.status = "declined";
  return ok();
});
route("POST", "/api/contacts/:id/cancel", (c) => {
  const me = need(c);
  const inv = c.db.invites.find((i) => i.id === c.params.id && i.fromId === me.id && i.status === "pending");
  if (inv) {
    inv.status = "cancelled";
    return ok();
  }
  const r = request(c);
  if (r.from !== me.id) throw new MockError(403, "permission_denied", "only the requester can cancel");
  r.status = "cancelled";
  return ok();
});
route("DELETE", "/api/contacts/:id", (c) => {
  const me = need(c);
  const inv = c.db.invites.find((i) => i.id === c.params.id && i.fromId === me.id && i.status === "pending");
  if (inv) {
    inv.status = "cancelled";
    return ok();
  }
  const l = c.db.links.find((x) => x.id === c.params.id && (x.a === me.id || x.b === me.id));
  if (!l) throw new MockError(404, "not_found", "no such contact");
  c.db.links = c.db.links.filter((x) => x !== l);
  return ok();
});

// ── groups ────────────────────────────────────────────────────────────

const COLORS = ["slate", "red", "orange", "amber", "green", "teal", "blue", "indigo", "violet", "pink"];
function groupFields(c: Ctx, required: boolean) {
  const out: { name?: string; color?: string } = {};
  if ("name" in c.body || required) {
    const name = str(c, "name").trim();
    if (!name) throw invalid("name", "required", "is required");
    if (name.length > 60) throw invalid("name", "maxlen", "must be at most 60 characters");
    out.name = name;
  }
  if ("color" in c.body || required) {
    let color = str(c, "color");
    if (!color && required && out.name) {
      // Optional on create: derived from the name.
      let h = 0;
      for (const ch of out.name) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
      color = COLORS[h % COLORS.length]!;
    }
    if (!COLORS.includes(color)) throw invalid("color", "oneof", `must be one of ${COLORS.join(", ")}`);
    out.color = color;
  }
  return out;
}

route("GET", "/api/groups", (c) => {
  const me = need(c);
  return ok({ groups: c.db.groups.filter((g) => roleIn(g, me.id)).map((g) => groupOut(c.db, g, me.id)) });
});
route("POST", "/api/groups", (c) => {
  const me = need(c);
  const f = groupFields(c, true);
  const now = new Date().toISOString();
  const g: MGroup = { id: typeid("group"), name: f.name!, color: f.color!, createdAt: now, members: [{ userId: me.id, role: "owner", joinedAt: now }] };
  c.db.groups.push(g);
  return ok(groupOut(c.db, g, me.id));
});
route("GET", "/api/groups/:id", (c) => {
  const me = need(c);
  return ok(groupOut(c.db, group(c, me.id), me.id));
});
route("PATCH", "/api/groups/:id", (c) => {
  const me = need(c);
  const g = group(c, me.id);
  if (!["owner", "admin"].includes(roleIn(g, me.id))) throw new MockError(403, "permission_denied", "only an owner or an admin can edit the group");
  Object.assign(g, groupFields(c, false));
  return ok(groupOut(c.db, g, me.id));
});
route("DELETE", "/api/groups/:id", (c) => {
  const me = need(c);
  const g = group(c, me.id);
  if (roleIn(g, me.id) !== "owner") throw new MockError(403, "permission_denied", "only the owner can delete the group");
  c.db.groups = c.db.groups.filter((x) => x !== g);
  c.db.tasks = c.db.tasks.filter((t) => t.groupId !== g.id);
  return ok();
});
route("POST", "/api/groups/:id/invitations", (c) => {
  const me = need(c);
  const g = group(c, me.id);
  if (!["owner", "admin"].includes(roleIn(g, me.id))) throw new MockError(403, "permission_denied", "only an owner or an admin can invite");
  const uid = str(c, "userId");
  if (!areContacts(c.db, me.id, uid)) throw new MockError(403, "permission_denied", "you can only invite your contacts");
  if (roleIn(g, uid)) throw new MockError(409, "conflict", "this person is already a member");
  if (c.db.invitations.some((i) => i.groupId === g.id && i.inviteeId === uid && i.status === "pending")) {
    throw new MockError(409, "conflict", "this person is already invited");
  }
  const inv = { id: typeid("invitation"), groupId: g.id, inviterId: me.id, inviteeId: uid, createdAt: new Date().toISOString(), status: "pending" };
  c.db.invitations.push(inv);
  return ok({
    id: inv.id,
    group: { id: g.id, name: g.name, color: g.color },
    inviter: ref(me),
    invitee: ref(user(c.db, uid)),
    createdAt: inv.createdAt,
    status: inv.status,
  });
});
route("GET", "/api/invitations", (c) => {
  const me = need(c);
  return ok({
    invitations: c.db.invitations
      .filter((i) => i.inviteeId === me.id && i.status === "pending")
      .map((i) => {
        const g = c.db.groups.find((x) => x.id === i.groupId)!;
        return { id: i.id, group: { id: g.id, name: g.name, color: g.color }, inviter: ref(user(c.db, i.inviterId)), createdAt: i.createdAt, status: i.status };
      }),
  });
});
function invitation(c: Ctx, uid: string) {
  const i = c.db.invitations.find((x) => x.id === c.params.id && x.inviteeId === uid && x.status === "pending");
  if (!i) throw new MockError(404, "not_found", "no such invitation");
  return i;
}
route("POST", "/api/invitations/:id/accept", (c) => {
  const me = need(c);
  const i = invitation(c, me.id);
  i.status = "accepted";
  const g = c.db.groups.find((x) => x.id === i.groupId)!;
  g.members.push({ userId: me.id, role: "member", joinedAt: new Date().toISOString() });
  return ok(groupOut(c.db, g, me.id));
});
route("POST", "/api/invitations/:id/decline", (c) => {
  const me = need(c);
  invitation(c, me.id).status = "declined";
  return ok();
});
route("DELETE", "/api/groups/:id/members/:userId", (c) => {
  const me = need(c);
  const g = group(c, me.id);
  const uid = c.params.userId!;
  const target = roleIn(g, uid);
  if (!target) throw new MockError(404, "not_found", "no such member");
  if (target === "owner") throw new MockError(409, "conflict", "the owner cannot leave the group");
  if (uid !== me.id && !["owner", "admin"].includes(roleIn(g, me.id))) throw new MockError(403, "permission_denied", "only an owner or an admin can remove a member");
  g.members = g.members.filter((m) => m.userId !== uid);
  return ok();
});
route("PATCH", "/api/groups/:id/members/:userId", (c) => {
  const me = need(c);
  const g = group(c, me.id);
  if (roleIn(g, me.id) !== "owner") throw new MockError(403, "permission_denied", "only the owner can change roles");
  const role = str(c, "role");
  if (role !== "admin" && role !== "member") throw invalid("role", "oneof", "must be admin or member");
  const m = g.members.find((x) => x.userId === c.params.userId);
  if (!m || m.role === "owner") throw new MockError(404, "not_found", "no such member");
  m.role = role;
  return ok(groupOut(c.db, g, me.id));
});

// ── activity ──────────────────────────────────────────────────────────

route("GET", "/api/activity", (c) => {
  const me = need(c);
  const entries = c.db.activity
    .filter((a) => a.userId === me.id)
    .sort((a, b) => b.at.localeCompare(a.at))
    .map((a) => {
      const g = c.db.groups.find((x) => x.id === a.groupId);
      return {
        id: a.id,
        kind: a.kind,
        ...(a.actorId ? { actor: ref(user(c.db, a.actorId)) } : {}),
        ...(a.targetId ? { target: ref(user(c.db, a.targetId)) } : {}),
        ...(a.taskId ? { task: { id: a.taskId, title: a.taskTitle ?? "" } } : {}),
        ...(g ? { group: { id: g.id, name: g.name, color: g.color } } : {}),
        text: a.text,
        at: a.at,
        read: a.read,
      };
    });
  return ok({ entries });
});
route("POST", "/api/activity/read", (c) => {
  const me = need(c);
  for (const a of c.db.activity) if (a.userId === me.id) a.read = true;
  return ok();
});

// ── dispatch ──────────────────────────────────────────────────────────

let db: DB | null = null;
export function mockDB(): DB {
  db ??= load();
  return db;
}
export function resetDB(next: DB) {
  db = next;
  save(db);
}

export function handle(method: string, url: URL, bodyText: string | undefined): Result {
  const state = mockDB();
  const parts = url.pathname.split("/").filter(Boolean);
  for (const r of routes) {
    if (r.method !== method || r.parts.length !== parts.length) continue;
    const params: Record<string, string> = {};
    let match = true;
    for (let i = 0; i < parts.length; i++) {
      const p = r.parts[i]!;
      if (p.startsWith(":")) params[p.slice(1)] = decodeURIComponent(parts[i]!);
      else if (p !== parts[i]) {
        match = false;
        break;
      }
    }
    if (!match) continue;
    let body: Record<string, unknown> = {};
    if (bodyText) {
      try {
        const parsed: unknown = JSON.parse(bodyText);
        if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) throw new Error("not an object");
        body = parsed as Record<string, unknown>;
      } catch {
        return { status: 400, body: { error: { code: "invalid_argument", message: "the request body is not valid JSON" } } };
      }
    }
    const session = state.sessions.find((s) => s.token === state.cookie);
    const me = session ? (state.users.find((u) => u.id === session.userId) ?? null) : null;
    if (session) session.lastSeenAt = new Date().toISOString();
    try {
      const res = r.handler({ params, query: url.searchParams, body, db: state, me });
      if (method !== "GET") save(state);
      return res;
    } catch (err) {
      if (err instanceof MockError) {
        if (method !== "GET") save(state);
        return {
          status: err.status,
          body: { error: { code: err.code, message: err.message, ...(err.violations ? { violations: err.violations } : {}) } },
        };
      }
      console.error("mock handler failed", err);
      return { status: 500, body: { error: { code: "internal", message: "internal error" } } };
    }
  }
  return { status: 404, body: { error: { code: "not_found", message: `no route for ${method} ${url.pathname}` } } };
}

