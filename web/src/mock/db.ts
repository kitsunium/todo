// The mock backend's state: the whole product in one JSON object, persisted
// in localStorage so a reload (or a screenshot run) keeps it. Only loaded in
// `vite --mode mock`.
import { addDays, format, startOfDay, subDays, subHours, subMinutes } from "date-fns";
import type { Priority } from "../api/types";

export type MUser = {
  id: string;
  name: string;
  email: string;
  password: string;
  verified: boolean;
  locked: boolean;
  failed: number;
  createdAt: string;
};
export type MSession = { id: string; token: string; userId: string; createdAt: string; lastSeenAt: string; userAgent: string; ip: string };
export type MTask = {
  id: string;
  title: string;
  notes: string;
  priority: Priority;
  status: "open" | "done" | "archived";
  due?: string;
  ownerId: string;
  assigneeId?: string;
  groupId?: string;
  sharedWith: string[];
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
  completedById?: string;
};
export type MLink = { id: string; a: string; b: string; since: string };
export type MRequest = { id: string; from: string; to: string; createdAt: string; status: string };
export type MInvite = { id: string; fromId: string; email: string; createdAt: string; status: string };
export type MMember = { userId: string; role: "owner" | "admin" | "member"; joinedAt: string };
export type MGroup = { id: string; name: string; color: string; createdAt: string; members: MMember[] };
export type MInvitation = { id: string; groupId: string; inviterId: string; inviteeId: string; createdAt: string; status: string };
export type MActivity = {
  id: string;
  userId: string;
  kind: string;
  actorId?: string;
  taskId?: string;
  taskTitle?: string;
  groupId?: string;
  text: string;
  at: string;
  read: boolean;
};
export type MLinkToken = { token: string; userId: string; kind: "verify" | "reset"; expires: number };
export type MMail = { id: string; to: string; subject: string; kind: "verify" | "reset" | "exists" | "invite"; link?: string; at: string };

export type DB = {
  version: number;
  seededOn: string;
  users: MUser[];
  sessions: MSession[];
  tasks: MTask[];
  links: MLink[];
  requests: MRequest[];
  invites: MInvite[];
  groups: MGroup[];
  invitations: MInvitation[];
  activity: MActivity[];
  mails: MMail[];
  linkTokens: MLinkToken[];
  /** The browser's todo_session cookie. */
  cookie?: string;
};

const KEY = "todo.mock.db";
const VERSION = 4;

const ALPHABET = "0123456789abcdefghjkmnpqrstvwxyz";
export function typeid(prefix: string): string {
  let s = "01j";
  for (let i = 0; i < 23; i++) s += ALPHABET[Math.floor(Math.random() * 32)];
  return `${prefix}_${s}`;
}

export function token(): string {
  const bytes = new Uint8Array(24);
  crypto.getRandomValues(bytes);
  return btoa(String.fromCharCode(...bytes)).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export const DEMO_EMAIL = "camille@example.com";
export const DEMO_PASSWORD = "correct-horse-42";

/** A date-only due: that day at 23:59 local time. */
function day(d: Date): string {
  const x = startOfDay(d);
  x.setHours(23, 59, 0, 0);
  return x.toISOString();
}
function at(d: Date, h: number, m = 0): string {
  const x = startOfDay(d);
  x.setHours(h, m, 0, 0);
  return x.toISOString();
}

export function seed(now = new Date()): DB {
  const iso = (d: Date) => d.toISOString();
  const long = subDays(now, 120);
  const u = (id: string, name: string, email: string, extra: Partial<MUser> = {}): MUser => ({
    id,
    name,
    email,
    password: DEMO_PASSWORD,
    verified: true,
    locked: false,
    failed: 0,
    createdAt: iso(long),
    ...extra,
  });
  const me = u("user_01j9camillemartin00000000", "Camille Martin", DEMO_EMAIL);
  const sam = u("user_01j9samrivera000000000000", "Sam Rivera", "sam@example.com");
  const noor = u("user_01j9noorhaddad00000000000", "Noor Haddad", "noor@example.com");
  const lea = u("user_01j9leadubois0000000000000", "Léa Dubois", "lea@example.com");
  const tom = u("user_01j9tombecker0000000000000", "Tom Becker", "tom@example.com");
  const priya = u("user_01j9priyanair000000000000", "Priya Nair", "priya@example.com");
  const ava = u("user_01j9avachen00000000000000", "Ava Chen", "ava@example.com");
  const marco = u("user_01j9marcorossi000000000000", "Marco Rossi", "marco@example.com");
  const unverified = u("user_01j9unverified00000000000", "Una Verified", "unverified@example.com", { verified: false });
  const locked = u("user_01j9lockedaccount000000000", "Lock Smith", "locked@example.com", { locked: true });

  const today = startOfDay(now);
  const g = (id: string, name: string, color: string, members: [MUser, MMember["role"], number][]): MGroup => ({
    id,
    name,
    color,
    createdAt: iso(subDays(now, 60)),
    members: members.map(([user, role, ago]) => ({ userId: user.id, role, joinedAt: iso(subDays(now, ago)) })),
  });
  const launch = g("group_01j9launch0000000000000000", "Launch", "orange", [
    [me, "owner", 40],
    [sam, "admin", 38],
    [noor, "member", 30],
    [lea, "member", 12],
  ]);
  const home = g("group_01j9home00000000000000000", "Home", "green", [
    [me, "owner", 90],
    [tom, "member", 88],
  ]);
  const crit = g("group_01j9designcrit000000000000", "Design crit", "violet", [
    [priya, "owner", 50],
    [me, "member", 20],
    [lea, "member", 18],
  ]);
  const book = g("group_01j9bookclub00000000000000", "Book club", "pink", [
    [noor, "owner", 25],
    [lea, "member", 20],
  ]);

  let n = 0;
  const t = (p: Partial<MTask> & { title: string }): MTask => {
    n++;
    const created = p.createdAt ?? iso(subHours(now, 200 - n * 7));
    return {
      id: `task_01j9seed${String(n).padStart(18, "0")}`,
      notes: "",
      priority: 0,
      status: "open",
      ownerId: me.id,
      sharedWith: [],
      createdAt: created,
      updatedAt: created,
      ...p,
    };
  };

  const tasks: MTask[] = [
    // Overdue
    t({ title: "Pay the electricity bill", priority: 1, due: day(subDays(today, 2)), notes: "Account number is on the last invoice, in the Home folder." }),
    t({ title: "Send the Q3 invoice to Studio Nord", priority: 2, due: day(subDays(today, 1)), sharedWith: [tom.id] }),
    // Today
    t({ title: "Ship the landing page", priority: 1, due: at(today, 17), groupId: launch.id, assigneeId: sam.id, notes: "Hero copy is final. Waiting on the pricing table from Noor." }),
    t({ title: "Review Sam’s pull request", priority: 2, due: day(today), groupId: launch.id, ownerId: sam.id, assigneeId: me.id }),
    t({ title: "Call the dentist", priority: 3, due: at(today, 16, 30) }),
    t({ title: "Pick up the dry cleaning", priority: 4, due: day(today) }),
    // Tomorrow
    t({ title: "Write the launch announcement", priority: 2, due: at(addDays(today, 1), 10), groupId: launch.id, assigneeId: me.id, notes: "Short. Three bullet points, one screenshot, one link." }),
    t({ title: "Draft the partnership email", priority: 3, due: day(addDays(today, 1)), ownerId: sam.id, sharedWith: [me.id] }),
    // This week / later
    t({ title: "QA the signup flow", priority: 3, due: day(addDays(today, 2)), groupId: launch.id, assigneeId: noor.id }),
    t({ title: "Book a table for Noor’s birthday", due: at(addDays(today, 2), 19, 30), sharedWith: [lea.id] }),
    t({ title: "Prepare critique notes", priority: 3, due: day(addDays(today, 1)), groupId: crit.id, ownerId: priya.id, assigneeId: me.id }),
    t({ title: "Fix the kitchen tap", priority: 3, due: day(addDays(today, 3)), groupId: home.id, assigneeId: tom.id }),
    t({ title: "Pick a date for the book club", due: day(addDays(today, 3)), ownerId: noor.id, sharedWith: [me.id, lea.id] }),
    t({ title: "Renew passport", priority: 2, due: day(addDays(today, 19)), notes: "Photos are in the top drawer. Bring the old passport." }),
    t({ title: "Plan the Lisbon trip", due: day(addDays(today, 34)), groupId: home.id }),
    // No date
    t({ title: "Back up the photo library", priority: 3, notes: "Both the laptop and the old external drive." }),
    t({ title: "Read “The Design of Everyday Things”", priority: 4 }),
    t({ title: "Try the new ramen place on Rue Oberkampf" }),
    t({ title: "Order stickers for the launch party", priority: 4, groupId: launch.id, assigneeId: lea.id }),
    t({ title: "Buy plants for the balcony", groupId: home.id }),
    // Done
    t({ title: "Send the launch deck", priority: 2, status: "done", completedAt: iso(subHours(now, 2)), completedById: me.id, groupId: launch.id }),
    t({ title: "Set up analytics", priority: 3, status: "done", completedAt: iso(subMinutes(now, 64)), completedById: sam.id, groupId: launch.id, ownerId: sam.id }),
    t({ title: "Water the plants", status: "done", completedAt: iso(subHours(subDays(now, 1), 3)), completedById: me.id, groupId: home.id }),
    t({ title: "Renew the gym membership", status: "done", completedAt: iso(subDays(now, 3)), completedById: me.id }),
    t({ title: "Share the Figma file", status: "done", completedAt: iso(subDays(now, 4)), completedById: me.id, groupId: crit.id, ownerId: priya.id }),
    t({ title: "File the expense report", priority: 3, status: "done", completedAt: iso(subDays(now, 9)), completedById: me.id }),
    t({ title: "Old errand", status: "archived", completedAt: iso(subDays(now, 20)), completedById: me.id }),
  ];

  const link = (a: MUser, b: MUser, ago: number): MLink => ({ id: typeid("contact"), a: a.id, b: b.id, since: iso(subDays(now, ago)) });
  const links = [link(me, sam, 110), link(me, noor, 80), link(me, lea, 60), link(me, tom, 100), link(me, priya, 45)];
  const requests: MRequest[] = [
    { id: typeid("request"), from: ava.id, to: me.id, createdAt: iso(subDays(now, 2)), status: "pending" },
    { id: typeid("request"), from: me.id, to: marco.id, createdAt: iso(subDays(now, 5)), status: "pending" },
  ];
  const invites: MInvite[] = [
    { id: typeid("invite"), fromId: me.id, email: "hello@studio-nord.com", createdAt: iso(subDays(now, 6)), status: "pending" },
    { id: typeid("invite"), fromId: me.id, email: "jules@example.org", createdAt: iso(subDays(now, 21)), status: "expired" },
  ];
  const invitations: MInvitation[] = [
    { id: typeid("invitation"), groupId: book.id, inviterId: noor.id, inviteeId: me.id, createdAt: iso(subHours(now, 20)), status: "pending" },
  ];

  const task = (title: string) => tasks.find((x) => x.title === title)!;
  const a = (p: Omit<MActivity, "id" | "userId">): MActivity => ({ id: typeid("entry"), userId: me.id, ...p });
  const activity: MActivity[] = [
    a({ kind: "task.completed", actorId: sam.id, taskId: task("Set up analytics").id, taskTitle: "Set up analytics", groupId: launch.id, text: "Sam Rivera completed “Set up analytics”.", at: iso(subMinutes(now, 64)), read: false }),
    a({ kind: "task.shared", actorId: noor.id, taskId: task("Pick a date for the book club").id, taskTitle: "Pick a date for the book club", text: "Noor Haddad shared “Pick a date for the book club” with you.", at: iso(subHours(now, 3)), read: false }),
    a({ kind: "task.assigned", actorId: sam.id, taskId: task("Review Sam’s pull request").id, taskTitle: "Review Sam’s pull request", groupId: launch.id, text: "Sam Rivera assigned “Review Sam’s pull request” to you.", at: iso(subHours(now, 5)), read: false }),
    a({ kind: "group.invited", actorId: noor.id, groupId: book.id, text: "Noor Haddad invited you to Book club.", at: iso(subHours(now, 20)), read: true }),
    a({ kind: "task.assigned", actorId: priya.id, taskId: task("Prepare critique notes").id, taskTitle: "Prepare critique notes", groupId: crit.id, text: "Priya Nair assigned “Prepare critique notes” to you.", at: iso(subHours(subDays(now, 1), 6)), read: true }),
    a({ kind: "contact.requested", actorId: ava.id, text: "Ava Chen wants to add you as a contact.", at: iso(subDays(now, 2)), read: true }),
    a({ kind: "task.overdue", taskId: task("Pay the electricity bill").id, taskTitle: "Pay the electricity bill", text: "“Pay the electricity bill” is overdue.", at: iso(subDays(now, 2)), read: true }),
    a({ kind: "task.updated", actorId: tom.id, taskId: task("Fix the kitchen tap").id, taskTitle: "Fix the kitchen tap", groupId: home.id, text: "Tom Becker changed the due date of “Fix the kitchen tap”.", at: iso(subDays(now, 3)), read: true }),
    a({ kind: "contact.accepted", actorId: lea.id, text: "Léa Dubois accepted your contact request.", at: iso(subDays(now, 3)), read: true }),
    a({ kind: "group.joined", actorId: lea.id, groupId: launch.id, text: "Léa Dubois joined Launch.", at: iso(subDays(now, 12)), read: true }),
  ];

  const current: MSession = {
    id: typeid("session"),
    token: token(),
    userId: me.id,
    createdAt: iso(subDays(now, 6)),
    lastSeenAt: iso(now),
    userAgent: navigator.userAgent,
    ip: "127.0.0.1",
  };
  const sessions: MSession[] = [
    current,
    { id: typeid("session"), token: token(), userId: me.id, createdAt: iso(subDays(now, 12)), lastSeenAt: iso(subDays(now, 2)), userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 19_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/19.1 Mobile/15E148 Safari/604.1", ip: "82.64.12.9" },
    { id: typeid("session"), token: token(), userId: me.id, createdAt: iso(subDays(now, 30)), lastSeenAt: iso(subDays(now, 14)), userAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0", ip: "91.198.174.192" },
  ];

  return {
    version: VERSION,
    seededOn: format(now, "yyyy-MM-dd"),
    users: [me, sam, noor, lea, tom, priya, ava, marco, unverified, locked],
    sessions,
    tasks,
    links,
    requests,
    invites,
    groups: [launch, home, crit, book],
    invitations,
    activity,
    mails: [],
    linkTokens: [],
    cookie: current.token,
  };
}

export function load(): DB {
  try {
    const raw = localStorage.getItem(KEY);
    if (raw) {
      const db = JSON.parse(raw) as DB;
      if (db.version === VERSION && db.seededOn === format(new Date(), "yyyy-MM-dd")) return db;
    }
  } catch {
    /* fall through to a fresh seed */
  }
  const db = seed();
  save(db);
  return db;
}

export function save(db: DB): void {
  try {
    localStorage.setItem(KEY, JSON.stringify(db));
  } catch {
    /* storage full or disabled: the state lasts for this page */
  }
}

export function clear(): void {
  localStorage.removeItem(KEY);
}
