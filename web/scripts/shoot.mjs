#!/usr/bin/env node
// Screenshots of every screen of the web app, in French and in English,
// desktop and mobile, through headless Chrome and the DevTools protocol
// (Node's own WebSocket, no dependency). Not part of the bundle.
//
//   npx vite --mode mock --port 5299 &     # the app, backend in memory
//   node scripts/shoot.mjs                 # → $OUT (default /tmp/todo-web-shots)
//
//   BASE=http://localhost:5299  where the app is served
//   OUT=/tmp/todo-web-shots     where the PNGs go
//   ONLY=today,panel            only the scenes whose name contains one of these
//   LOCALES=fr,en               which languages (the account's, and the signed-out pages')
//   MOBILE=0                    skip the mobile pass
//   SCALE=2                     desktop device pixel ratio (details, typography)
//   REAL=1                      the real app (kit dev): signs in as REAL_EMAIL / REAL_PASSWORD
//                               (camille@example.com), adds demo data through the API first
//   PROD=1                      a production build (no mock): signed-out scenes only,
//                               e.g. served with the frontend CSP to catch violations
//
// Scenes drive the UI like a person would: real mouse and keyboard events,
// so Radix menus and popovers open exactly as they do in a browser. Console
// errors and uncaught exceptions are collected into $OUT/console.log; the
// script exits 1 when there is any.

import { spawn } from "node:child_process";
import fs from "node:fs";
// The interface's own words, to find buttons and tabs by what they say.
import { en } from "../src/i18n/en.ts";
import { fr } from "../src/i18n/fr.ts";

const WORDS = { fr, en };

const CHROME =
  process.env.CHROME ??
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const PORT = Number(process.env.CDP_PORT ?? 9555);
const BASE = process.env.BASE ?? "http://localhost:5299";
const OUT = process.env.OUT ?? "/tmp/todo-web-shots";
const ONLY = (process.env.ONLY ?? "").split(",").filter(Boolean);
const LOCALES = (process.env.LOCALES ?? "fr,en").split(",").filter((l) => l === "fr" || l === "en");
/** The current scene's language, and its words. */
let LOC = "fr";
const say = (key) => WORDS[LOC][key];
const MOBILE = process.env.MOBILE !== "0";

fs.mkdirSync(OUT, { recursive: true });
const profile = `${OUT}/.profile`;
fs.rmSync(profile, { recursive: true, force: true });
const chrome = spawn(
  CHROME,
  [
    "--headless=new",
    "--disable-gpu",
    "--hide-scrollbars",
    "--no-first-run",
    "--no-default-browser-check",
    `--remote-debugging-port=${PORT}`,
    `--user-data-dir=${profile}`,
    "--window-size=1440,900",
    "about:blank",
  ],
  { stdio: "ignore" },
);

const wait = (ms) => new Promise((r) => setTimeout(r, ms));
let targets = [];
for (let i = 0; i < 80; i++) {
  try {
    targets = await (await fetch(`http://127.0.0.1:${PORT}/json/list`)).json();
    if (targets.some((t) => t.type === "page")) break;
  } catch {
    /* not up yet */
  }
  await wait(150);
}
const page = targets.find((t) => t.type === "page");
if (!page) {
  console.error("Chrome did not start");
  chrome.kill();
  process.exit(2);
}

const ws = new WebSocket(page.webSocketDebuggerUrl);
await new Promise((r) => (ws.onopen = r));
let seq = 0;
const pending = new Map();
const problems = [];
let scene = "";
ws.onmessage = (m) => {
  const msg = JSON.parse(m.data);
  if (msg.id && pending.has(msg.id)) {
    pending.get(msg.id)(msg);
    pending.delete(msg.id);
    return;
  }
  if (msg.method === "Runtime.exceptionThrown") {
    const d = msg.params.exceptionDetails;
    problems.push(
      `[${scene}] EXCEPTION: ${d.exception?.description ?? d.text}`,
    );
  } else if (
    msg.method === "Runtime.consoleAPICalled" &&
    (msg.params.type === "error" || msg.params.type === "assert")
  ) {
    problems.push(
      `[${scene}] console.${msg.params.type}: ${msg.params.args.map((a) => a.value ?? a.description).join(" ")}`,
    );
  } else if (
    msg.method === "Log.entryAdded" &&
    msg.params.entry.level === "error"
  ) {
    problems.push(
      `[${scene}] LOG: ${msg.params.entry.text} ${msg.params.entry.url ?? ""}`,
    );
  }
};
const send = (method, params = {}) =>
  new Promise((resolve, reject) => {
    const id = ++seq;
    pending.set(id, (msg) =>
      msg.error
        ? reject(new Error(`${method}: ${msg.error.message}`))
        : resolve(msg.result),
    );
    ws.send(JSON.stringify({ id, method, params }));
  });

await send("Runtime.enable");
await send("Log.enable");
await send("Page.enable");

// ── helpers ───────────────────────────────────────────────────────────

let viewport = { width: 1440, height: 900, mobile: false, scale: 1 };
async function setViewport(v) {
  viewport = v;
  await send("Emulation.setDeviceMetricsOverride", {
    width: v.width,
    height: v.height,
    deviceScaleFactor: v.scale,
    mobile: v.mobile,
  });
  await send("Emulation.setTouchEmulationEnabled", { enabled: v.mobile });
}

async function motion() {
  await send("Emulation.setEmulatedMedia", {
    features: [{ name: "prefers-reduced-motion", value: "no-preference" }],
  });
}

async function evaluate(expression) {
  const r = await send("Runtime.evaluate", {
    expression,
    awaitPromise: true,
    returnByValue: true,
  });
  if (r.exceptionDetails)
    throw new Error(
      `evaluate failed: ${r.exceptionDetails.exception?.description ?? r.exceptionDetails.text}\n${expression}`,
    );
  return r.result?.value;
}

async function go(path, settle = 900) {
  await send("Page.navigate", { url: `${BASE}${path}` });
  await waitFor("document.readyState === 'complete'", 8000);
  await wait(settle);
}

async function waitFor(expr, timeout = 6000) {
  const end = Date.now() + timeout;
  while (Date.now() < end) {
    try {
      if (await evaluate(`!!(${expr})`)) return true;
    } catch {
      /* page navigating */
    }
    await wait(80);
  }
  throw new Error(`timed out waiting for ${expr}`);
}

/** The center of the first element matching a CSS selector, or of the element containing text. */
async function center(target) {
  const expr =
    typeof target === "string"
      ? `(() => { const el = document.querySelector(${JSON.stringify(target)}); if (!el) return null; el.scrollIntoView({block:"nearest"}); const r = el.getBoundingClientRect(); return {x: r.x + r.width/2, y: r.y + r.height/2}; })()`
      : `(() => { const want = ${JSON.stringify(target.text)}; const sel = ${JSON.stringify(target.within ?? "button, a, [role=menuitem], [role=option], [role=tab], label, li, span")}; const els = [...document.querySelectorAll(sel)].filter(e => e.textContent.includes(want) && e.getClientRects().length).sort((a, b) => a.textContent.length - b.textContent.length); const el = els[${Number(target.nth ?? 0)}]; if (!el) return null; el.scrollIntoView({block:"nearest"}); const r = el.getBoundingClientRect(); return {x: r.x + r.width/2, y: r.y + r.height/2}; })()`;
  const p = await evaluate(expr);
  if (!p) throw new Error(`not found: ${JSON.stringify(target)}`);
  return p;
}

async function mouse(type, x, y, extra = {}) {
  await send("Input.dispatchMouseEvent", {
    type,
    x,
    y,
    button: "left",
    clickCount: 1,
    pointerType: "mouse",
    ...extra,
  });
}

async function hover(target) {
  const { x, y } = await center(target);
  await mouse("mouseMoved", x, y, { button: "none" });
  await wait(250);
}

async function click(target, settle = 350) {
  const { x, y } = await center(target);
  if (viewport.mobile) {
    await send("Input.dispatchTouchEvent", {
      type: "touchStart",
      touchPoints: [{ x, y }],
    });
    await send("Input.dispatchTouchEvent", {
      type: "touchEnd",
      touchPoints: [],
    });
  } else {
    await mouse("mouseMoved", x, y, { button: "none" });
    await mouse("mousePressed", x, y);
    await mouse("mouseReleased", x, y);
  }
  await wait(settle);
}

async function type(text, settle = 250) {
  await send("Input.insertText", { text });
  await wait(settle);
}

const KEYS = {
  Enter: { code: "Enter", key: "Enter", windowsVirtualKeyCode: 13, text: "\r" },
  Escape: { code: "Escape", key: "Escape", windowsVirtualKeyCode: 27 },
  Tab: { code: "Tab", key: "Tab", windowsVirtualKeyCode: 9 },
  ArrowDown: { code: "ArrowDown", key: "ArrowDown", windowsVirtualKeyCode: 40 },
  ArrowUp: { code: "ArrowUp", key: "ArrowUp", windowsVirtualKeyCode: 38 },
};
async function press(name, { meta = false, settle = 300 } = {}) {
  const k = KEYS[name] ?? {
    key: name,
    code: /^[a-z]$/i.test(name) ? `Key${name.toUpperCase()}` : undefined,
    windowsVirtualKeyCode: name.toUpperCase().charCodeAt(0),
    text: name,
  };
  const modifiers = meta ? 4 : 0;
  const { text, ...rest } = k;
  await send("Input.dispatchKeyEvent", { type: text && !meta ? "keyDown" : "rawKeyDown", modifiers, ...rest, ...(text && !meta ? { text } : {}) });
  await send("Input.dispatchKeyEvent", { type: "keyUp", modifiers, ...rest });
  await wait(settle);
}

async function shot(name) {
  const r = await send("Page.captureScreenshot", { format: "png" });
  const file = `${OUT}/${name}.png`;
  fs.writeFileSync(file, Buffer.from(r.data, "base64"));
  console.log(`  ${file}`);
}

// ── modes ─────────────────────────────────────────────────────────────
//
// Mock (default): the in-memory backend, re-seeded before every scene.
// REAL=1: a running app (kit dev) with its seeded users; a one-time setup
// through the API gives every screen something to show, and scenes that
// change data put it back.
// PROD=1: a production build without a backend: signed-out scenes only.

const REAL = process.env.REAL === "1";
const PROD = process.env.PROD === "1";
const EMAIL = process.env.REAL_EMAIL ?? (REAL ? "camille@example.com" : "camille@example.com");
const PASSWORD = process.env.REAL_PASSWORD ?? (REAL ? "correct-horse-battery-staple" : "correct-horse-42");
const stamp = Date.now().toString(36);

const mock = {
  reset: () => evaluate("window.__mock.reset()"),
  signOut: () => evaluate("window.__mock.signOut()"),
};

/** A fetch from the page: same origin, the page's cookie. */
const inPage = (method, path, body) =>
  evaluate(
    `fetch(${JSON.stringify(path)}, { method: ${JSON.stringify(method)}, credentials: "same-origin", headers: ${JSON.stringify(
      body ? { "Content-Type": "application/json" } : {},
    )}, ${body ? `body: ${JSON.stringify(JSON.stringify(body))},` : ""} }).then(async (r) => ({ status: r.status, body: r.status === 204 ? null : await r.json().catch(() => null) }))`,
  );

/** A signed-in API client from this script (REAL mode). */
async function as(email, password = PASSWORD) {
  const r = await fetch(`${BASE}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!r.ok) throw new Error(`login ${email}: ${r.status}`);
  const cookie = r.headers.getSetCookie().map((c) => c.split(";")[0]).join("; ");
  return async (method, path, body) => {
    const res = await fetch(`${BASE}${path}`, {
      method,
      headers: { Cookie: cookie, ...(body ? { "Content-Type": "application/json" } : {}) },
      ...(body ? { body: JSON.stringify(body) } : {}),
    });
    const text = await res.text();
    return { status: res.status, body: text ? JSON.parse(text) : null };
  };
}

/** The link of the newest mail to `to` containing `prefix`, from the dev mailbox (REAL mode). */
async function mailLink(to, prefix, timeout = 8000) {
  const end = Date.now() + timeout;
  while (Date.now() < end) {
    const list = await (await fetch(`${BASE}/_kit/api/mail`)).json();
    for (const m of list) {
      if (!m.to.some((t) => t.includes(to))) continue;
      const full = await (await fetch(`${BASE}/_kit/api/mail/${encodeURIComponent(m.id)}`)).json();
      const hit = (full.text ?? "").match(new RegExp(`${prefix.replace("?", "\\?")}[A-Za-z0-9_=-]+`));
      if (hit) return hit[0];
    }
    await wait(300);
  }
  throw new Error(`no mail with ${prefix} to ${to}`);
}

/** A verified account of our own (REAL mode): signup, then the mailed link. */
async function verifiedUser(name, email, password) {
  const r = await fetch(`${BASE}/api/auth/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, name, password }),
  });
  if (!r.ok) throw new Error(`signup ${email}: ${r.status}`);
  const link = await mailLink(email, "/verify?token=");
  const token = new URL(link, BASE).searchParams.get("token");
  const v = await fetch(`${BASE}/api/auth/verify`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ token }) });
  if (!v.ok) throw new Error(`verify ${email}: ${v.status}`);
  return as(email, password);
}

let invitedGroup = "group_01j9bookclub00000000000000";
let mainGroup = "group_01j9launch0000000000000000";

/** REAL mode: give every screen something to show. */
async function setupReal() {
  const camille = await as(EMAIL);
  const me = (await camille("GET", "/api/auth/me")).body.user;
  const groups = (await camille("GET", "/api/groups")).body.groups;
  mainGroup = [...groups].sort((a, b) => b.members.length - a.members.length)[0]?.id ?? mainGroup;
  const hugo = await as("hugo@example.com");
  // An invitation waiting for Camille.
  const g = (await hugo("POST", "/api/groups", { name: "Book club", color: "pink" })).body;
  await hugo("POST", `/api/groups/${g.id}/invitations`, { userId: me.id });
  invitedGroup = g.id;
  // A task shared with her, one assigned to her.
  const shared = (await hugo("POST", "/api/tasks", { title: "Pick a venue for the team dinner", priority: 3, due: new Date(Date.now() + 2 * 864e5).toISOString() })).body;
  await hugo("POST", `/api/tasks/${shared.id}/share`, { userId: me.id });
  const launch = groups.find((x) => x.members.some((m) => m.user.email === "hugo@example.com"));
  if (launch) {
    await hugo("POST", "/api/tasks", { title: "Review the launch video", priority: 2, groupId: launch.id, assigneeId: me.id, due: new Date(Date.now() + 864e5).toISOString() });
  }
  // A request for her, one from her, an invitation by email.
  const jules = await verifiedUser("Jules Laurent", `jules.${stamp}@example.com`, "Jules-password-123");
  await jules("POST", "/api/contacts", { email: EMAIL });
  await verifiedUser("Marco Rossi", `marco.${stamp}@example.com`, "Marco-password-123");
  await camille("POST", "/api/contacts", { email: `marco.${stamp}@example.com` });
  await camille("POST", "/api/contacts", { email: `hello.${stamp}@studio-nord.com` });
}

async function fresh(path, signedIn = true) {
  if (PROD) return go(path);
  if (REAL) {
    await go("/favicon.svg", 100);
    const me = await inPage("GET", "/api/auth/me");
    if (signedIn && (me.status !== 200 || me.body?.user?.email !== EMAIL)) {
      const r = await inPage("POST", "/api/auth/login", { email: EMAIL, password: PASSWORD });
      if (r.status !== 200) throw new Error(`sign in: ${r.status}`);
    }
    if (!signedIn && me.status === 200) await inPage("POST", "/api/auth/logout");
    // Signed in, the account's language wins; signed out, the browser's last choice.
    if (signedIn) await inPage("PATCH", "/api/auth/me", { locale: LOC });
    await evaluate(`localStorage.setItem('todo.locale', ${JSON.stringify(LOC)})`);
    return go(path, 1300);
  }
  // A known state for every scene: the seed, the right session, a snappy mock.
  await go("/login", 300);
  await evaluate(`localStorage.setItem('todo.mock.latency', '60'); localStorage.setItem('todo.locale', ${JSON.stringify(LOC)})`);
  await mock.reset();
  if (signedIn) await inPage("PATCH", "/api/auth/me", { locale: LOC });
  else await mock.signOut();
  await go(path);
}

/** Clicks the button labelled `label` inside the nth task row. */
async function rowButton(n, label, settle = 450) {
  const pos = await evaluate(`(() => {
    const row = document.querySelectorAll("[data-task-id]")[${n}];
    if (!row) return null;
    row.scrollIntoView({ block: "center" });
    const r0 = row.getBoundingClientRect();
    return { x: r0.x + r0.width / 2, y: r0.y + r0.height / 2 };
  })()`);
  if (!pos) throw new Error(`no row ${n}`);
  await mouse("mouseMoved", pos.x, pos.y, { button: "none" });
  await wait(200);
  const b = await evaluate(`(() => {
    const row = document.querySelectorAll("[data-task-id]")[${n}];
    const btn = row?.querySelector(${JSON.stringify(`button[aria-label="${label}"]`)});
    if (!btn) return null;
    const r = btn.getBoundingClientRect();
    return { x: r.x + r.width / 2, y: r.y + r.height / 2 };
  })()`);
  if (!b) throw new Error(`no ${label} in row ${n}`);
  await mouse("mouseMoved", b.x, b.y, { button: "none" });
  await mouse("mousePressed", b.x, b.y);
  await mouse("mouseReleased", b.x, b.y);
  await wait(settle);
}

async function undoIfReal() {
  if (!REAL) return;
  await click({ text: say("common.undo"), within: "button" }, 1200).catch(() => undefined);
}

async function signupAs(email) {
  await fresh("/signup", false);
  await type("Alex Moreau");
  await click("input[name=email]");
  await type(email);
  await click("input[name=password]");
  await type("Tangerine-Otter-42");
}

// ── scenes ────────────────────────────────────────────────────────────

const desktop = [
  ["auth-login", async () => fresh("/login", false)],
  [
    "auth-login-error",
    async () => {
      await fresh("/login", false);
      await click("input[name=email]");
      await type("nobody@example.com");
      await click("input[name=password]");
      await type("not-my-password");
      await press("Enter", { settle: 900 });
    },
  ],
  [
    "auth-login-unverified",
    async () => {
      let email = "unverified@example.com";
      let password = "correct-horse-42";
      if (REAL) {
        email = `una.${stamp}.${LOC}@example.com`;
        password = "Unverified-pass-1";
        await fetch(`${BASE}/api/auth/signup`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ email, name: "Una Verified", password }) });
      }
      await fresh("/login", false);
      await click("input[name=email]");
      await type(email);
      await click("input[name=password]");
      await type(password);
      await press("Enter", { settle: 900 });
    },
  ],
  [
    "auth-login-locked",
    async () => {
      let email = "locked@example.com";
      if (REAL) {
        email = `lock.${stamp}.${LOC}@example.com`;
        await verifiedUser("Lock Smith", email, "Lock-password-123");
        for (let i = 0; i < 5; i++) {
          await fetch(`${BASE}/api/auth/login`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ email, password: `wrong-${i}-password` }) });
          await wait(600);
        }
      }
      await fresh("/login", false);
      await click("input[name=email]");
      await type(email);
      await click("input[name=password]");
      await type("whatever-it-is");
      await press("Enter", { settle: 900 });
    },
  ],
  [
    "auth-signup",
    async () => {
      await fresh("/signup", false);
      await type("Alex Moreau");
      await click("input[name=email]");
      await type("alex@");
      await click("input[name=password]");
      await type("Tangerine-otter");
      await wait(200);
    },
  ],
  [
    "auth-check-inbox",
    async () => {
      await signupAs(REAL ? `alex.${stamp}.${LOC}@example.com` : "alex@example.com");
      await press("Enter", { settle: 1600 });
    },
  ],
  ["auth-verify-expired", async () => fresh("/verify?token=expired-or-used", false)],
  ["auth-forgot", async () => fresh(`/forgot?email=${encodeURIComponent(EMAIL)}`, false)],
  ["auth-reset", async () => fresh("/reset?token=from-the-mail", false)],
  ["app-today", async () => fresh("/app/today")],
  ["app-inbox", async () => fresh("/app/inbox")],
  ["app-upcoming", async () => fresh("/app/upcoming")],
  ["app-shared", async () => fresh("/app/shared")],
  ["app-assigned", async () => fresh("/app/assigned")],
  ["app-completed", async () => fresh("/app/completed")],
  [
    "app-row-hover",
    async () => {
      await fresh("/app/inbox");
      const p = await evaluate(`(() => { const r = document.querySelectorAll("[data-task-id]")[1]?.getBoundingClientRect(); return r ? { x: r.x + r.width / 2, y: r.y + r.height / 2 } : null; })()`);
      if (p) await mouse("mouseMoved", p.x, p.y, { button: "none" });
      await wait(300);
    },
  ],
  [
    "app-keyboard-select",
    async () => {
      await fresh("/app/today");
      await press("j");
      await press("j");
      await press("j");
    },
  ],
  [
    "app-focus-ring",
    async () => {
      await fresh("/app/today");
      await click(`input[aria-label="${say("quickadd.label")}"]`);
      await press("Escape");
      for (let i = 0; i < 4; i++) await press("Tab", { settle: 120 });
      await wait(200);
    },
  ],
  [
    "app-complete-frames",
    async () => {
      await fresh("/app/today");
      await hover("[data-task-id] button[role=checkbox]");
      await shot(`${scene}-0-hover`);
      await click("[data-task-id] button[role=checkbox]", 0);
      await wait(140);
      await shot(`${scene}-1-check`);
      await wait(360);
      await shot(`${scene}-2-strike`);
      await wait(520);
      await shot(`${scene}-3-collapse`);
      await wait(500);
      await undoIfReal();
      return true;
    },
  ],
  [
    "app-quick-add",
    async () => {
      await fresh("/app/inbox");
      await click(`input[aria-label="${say("quickadd.label")}"]`);
      const who = REAL ? "#home @hugo" : "#launch @sam";
      await type(LOC === "fr" ? `Appeler l’imprimeur demain à 17h !haute ${who}` : `Call the printer tomorrow at 5pm !high ${who}`, 600);
    },
  ],
  [
    "app-quick-add-suggest",
    async () => {
      await fresh("/app/inbox");
      await click(`input[aria-label="${say("quickadd.label")}"]`);
      await type(LOC === "fr" ? "Préparer le séminaire #" : "Plan the offsite #", 400);
    },
  ],
  [
    "app-task-panel",
    async () => {
      await fresh("/app/today");
      await click("[data-task-id] a", 1000);
    },
  ],
  [
    "app-task-panel-due",
    async () => {
      await fresh("/app/today");
      await click("[data-task-id] a", 1000);
      await click({ text: say("panel.due"), within: "[role=dialog] span" }).catch(() => undefined);
      const due = await evaluate(`(() => { const row = [...document.querySelectorAll('[role=dialog] span')].find((s) => s.textContent === ${JSON.stringify(say("panel.due"))})?.closest("div.grid"); const b = row?.querySelector("button"); if (!b) return null; const r = b.getBoundingClientRect(); return { x: r.x + r.width / 2, y: r.y + r.height / 2 }; })()`);
      if (!due) throw new Error("no due button");
      await mouse("mouseMoved", due.x, due.y, { button: "none" });
      await mouse("mousePressed", due.x, due.y);
      await mouse("mouseReleased", due.x, due.y);
      await wait(500);
    },
  ],
  [
    "app-row-menu",
    async () => {
      await fresh("/app/inbox");
      await rowButton(3, say("common.moreActions"), 500);
    },
  ],
  [
    "app-palette",
    async () => {
      await fresh("/app/today");
      await press("k", { meta: true, settle: 500 });
      await type(REAL ? "the" : "pla", 600);
    },
  ],
  [
    "app-new-task",
    async () => {
      await fresh("/app/activity");
      await press("c", { settle: 600 });
      await type(LOC === "fr" ? "Réserver les vols pour Lisbonne la semaine prochaine !moyenne #home" : "Book flights to Lisbon next week !medium #home", 500);
    },
  ],
  [
    "app-complete-toast",
    async () => {
      await fresh("/app/today");
      await click("[data-task-id] button[role=checkbox]", 1300);
      await shot(scene);
      await undoIfReal();
      return true;
    },
  ],
  ["app-group", async () => fresh(`/app/groups/${mainGroup}`)],
  [
    "app-group-members",
    async () => {
      await fresh(`/app/groups/${mainGroup}`);
      await click({ text: say("group.tabMembers"), within: "[role=tab]" }, 500);
    },
  ],
  ["app-group-invited", async () => fresh(`/app/groups/${invitedGroup}`)],
  [
    "app-group-create",
    async () => {
      await fresh("/app/today");
      await click(`button[aria-label="${say("sidebar.newGroup")}"]`, 500);
      await type(LOC === "fr" ? "Jardin" : "Garden", 300);
    },
  ],
  ["app-contacts", async () => fresh("/app/contacts")],
  [
    "app-contacts-requests",
    async () => {
      await fresh("/app/contacts");
      await click({ text: say("contacts.tabRequests"), within: "[role=tab]" }, 400);
    },
  ],
  [
    "app-contacts-invites",
    async () => {
      await fresh("/app/contacts");
      await click({ text: say("contacts.tabInvites"), within: "[role=tab]" }, 400);
    },
  ],
  [
    "app-contacts-add",
    async () => {
      await fresh("/app/contacts");
      await click({ text: say("contacts.add"), within: "button" }, 500);
      await type("jules@studio-nord.com", 200);
    },
  ],
  ["app-activity", async () => fresh("/app/activity")],
  ["app-settings", async () => fresh("/app/settings")],
  [
    "app-user-menu",
    async () => {
      await fresh("/app/today");
      await click("nav .border-t button", 500);
    },
  ],
  [
    "app-shortcuts",
    async () => {
      await fresh("/app/today");
      await press("?", { settle: 500 });
    },
  ],
  [
    "app-empty-new-account",
    async () => {
      const email = REAL ? `new.${stamp}.${LOC}@example.com` : "alex@example.com";
      await signupAs(email);
      await press("Enter", { settle: 1400 });
      if (REAL) await go(await mailLink(email, "/verify?token="), 1500);
      else await click({ text: say("mock.verify"), within: "a" }, 1500);
      await waitFor("location.pathname === '/app/today'");
      await wait(1200);
    },
  ],
  ["app-404", async () => fresh("/app/nowhere")],
];

const mobile = [
  ["m-login", async () => fresh("/login", false)],
  ["m-today", async () => fresh("/app/today")],
  [
    "m-drawer",
    async () => {
      await fresh("/app/today");
      await click(`button[aria-label="${say("app.openMenu")}"]`, 600);
    },
  ],
  [
    "m-task-panel",
    async () => {
      await fresh("/app/today");
      await click("[data-task-id] a", 1000);
    },
  ],
  ["m-contacts", async () => fresh("/app/contacts")],
  ["m-activity", async () => fresh("/app/activity")],
];

async function run(list, prefix, vp) {
  await setViewport(vp);
  await motion();
  for (const loc of LOCALES) {
    LOC = loc;
    for (const [name, fn] of list) {
      if (ONLY.length && !ONLY.some((o) => name.includes(o))) continue;
      scene = `${prefix}${name}-${loc}`;
      try {
        // A scene that takes its own screenshots returns true.
        if ((await fn()) !== true) await shot(scene);
      } catch (err) {
        const why = err instanceof Error ? err.message : String(err);
        problems.push(`[${scene}] SCENE FAILED: ${why}`);
        console.log(`  ✗ ${scene}: ${why}`);
      }
    }
  }
}

console.log(`Shooting ${BASE} → ${OUT}${REAL ? " (real app)" : PROD ? " (production build)" : " (mock)"}`);
if (REAL) await setupReal();
await run(desktop, "", { width: 1440, height: 900, mobile: false, scale: Number(process.env.SCALE ?? 1) });
if (MOBILE) await run(mobile, "", { width: 390, height: 844, mobile: true, scale: 2 });
// Camille speaks French again: the seeded account's own language.
if (REAL) {
  try {
    await (await as(EMAIL))("PATCH", "/api/auth/me", { locale: "fr" });
  } catch (err) {
    problems.push(`[restore] Camille's locale: ${err instanceof Error ? err.message : err}`);
  }
}

fs.writeFileSync(`${OUT}/console.log`, problems.join("\n") + (problems.length ? "\n" : ""));
console.log(`${problems.length} problem(s)`);
for (const p of problems.slice(0, 40)) console.log(`  ${p}`);
ws.close();
chrome.kill();
process.exit(problems.length ? 1 : 0);
