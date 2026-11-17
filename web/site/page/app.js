// The todo's web app: one page that talks to the HTTP API with the session
// cookie the sign-in sets. No build step, no dependency. The words are in
// lang.js, the task views in tasks.js, contacts and groups in circle.js; this
// file holds the helpers, the router, the account pages and the start.
"use strict";

const $ = (id) => document.getElementById(id);
const zeroTime = "0001-01-01T00:00:00Z";
const publicPages = ["/verify", "/forgot", "/reset"];

let me = null; // the signed-in member's profile, null at the door
let known = false; // whether the session has been asked about
let pref = String(navigator.language || "fr")
  .toLowerCase()
  .startsWith("en")
  ? "en"
  : "fr";
let epoch = 0; // a draw that finds the epoch moved on paints nothing
let moved = false; // the member navigated: focus goes to the page
let drawn = ""; // the address the view on screen was drawn for

// ---- words and dates ------------------------------------------------------

const lang = () =>
  me && (me.locale === "en" || me.locale === "fr") ? me.locale : pref;

function hasWord(key) {
  return words[lang()][key] !== undefined || words.fr[key] !== undefined;
}

// t says a key in the page's language; a key missing there falls back to
// French, and one missing in both shows itself.
function t(key, vars) {
  let text = words[lang()][key];
  if (text === undefined) text = words.fr[key];
  if (text === undefined) return key;
  return vars
    ? text.replace(/\{(\w+)\}/g, (m, k) => (k in vars ? String(vars[k]) : m))
    : text;
}

// said is a refusal of the API in the page's language: the API words it in
// English, and the dictionaries hold every sentence it can say, by that
// sentence, and every broken rule of a field, by the field and the rule. One
// they do not hold is shown as the API said it.
function said(key, otherwise) {
  return hasWord(key) ? t(key) : otherwise;
}

function browserZone() {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  } catch {
    return "UTC";
  }
}

// zone is the account's time zone (an empty one reads UTC), or the browser's
// before anybody is signed in.
const zone = () => (me ? me.timeZone || "UTC" : browserZone());

const formats = new Map();
function formatter(opts) {
  const key = lang() + "|" + zone() + "|" + JSON.stringify(opts);
  if (!formats.has(key)) {
    let f;
    try {
      f = new Intl.DateTimeFormat(lang(), { timeZone: zone(), ...opts });
    } catch {
      f = new Intl.DateTimeFormat(lang(), { timeZone: "UTC", ...opts });
    }
    formats.set(key, f);
  }
  return formats.get(key);
}

function isDate(text) {
  const d = new Date(text);
  return Boolean(text) && !Number.isNaN(d.getTime()) && d.getFullYear() > 1;
}

// when writes an instant in the account's time zone; the zero time is "".
function when(text, opts) {
  return isDate(text)
    ? formatter(opts || { dateStyle: "medium", timeStyle: "short" }).format(
        new Date(text),
      )
    : "";
}

// zoneOffset is how far ahead of UTC the zone is at an instant, in ms.
function zoneOffset(ts, tz) {
  const parts = Object.fromEntries(
    new Intl.DateTimeFormat("en-US", {
      timeZone: tz,
      hourCycle: "h23",
      year: "numeric",
      month: "numeric",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
      second: "numeric",
    })
      .formatToParts(new Date(ts))
      .map((p) => [p.type, p.value]),
  );
  const wall = Date.UTC(
    +parts.year,
    +parts.month - 1,
    +parts.day,
    +parts.hour,
    +parts.minute,
    +parts.second,
  );
  return wall - Math.floor(ts / 1000) * 1000;
}

// toInstant reads a datetime-local value as a wall time in the account's
// zone and answers the RFC 3339 instant it is.
function toInstant(local, tz) {
  const [day, time] = local.split("T");
  const [y, mo, d] = day.split("-").map(Number);
  const [h, mi] = time.split(":").map(Number);
  const wall = Date.UTC(y, mo - 1, d, h, mi);
  let ts = wall;
  try {
    for (let i = 0; i < 2; i++) ts = wall - zoneOffset(ts, tz);
  } catch {
    ts = wall;
  }
  return new Date(ts).toISOString();
}

// toLocal is toInstant's way back: the value a datetime-local input shows.
function toLocal(text, tz) {
  if (!isDate(text)) return "";
  let ts = new Date(text).getTime();
  try {
    ts += zoneOffset(ts, tz);
  } catch {
    // an unknown zone reads as UTC
  }
  return new Date(ts).toISOString().slice(0, 16);
}

// ---- the API --------------------------------------------------------------

// call sends one request to the API and answers its JSON; a refusal is
// thrown with the API's own message and code, and a 429 says to wait.
async function call(method, path, body) {
  let res;
  try {
    res = await fetch(path, {
      method,
      credentials: "same-origin",
      headers: body ? { "Content-Type": "application/json" } : {},
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch {
    const down = new Error(t("err.network"));
    down.status = 0;
    throw down;
  }
  const data = await res.json().catch(() => ({}));
  if (res.ok) return data;
  const fault = data.error || {};
  const err = new Error(said("api." + fault.message, fault.message || res.statusText));
  err.status = res.status;
  err.code = fault.code || "";
  if (fault.violations && fault.violations.length) {
    err.message = fault.violations
      .map((v) => said("api.field." + v.path + "." + v.rule, v.message))
      .join(" ");
  }
  if (res.status === 429) err.message = t("err.wait");
  throw err;
}

function notice(text) {
  $("notice").textContent = text || "";
  $("notice").hidden = !text;
}

// word says something that went well.
function word(text) {
  $("word").textContent = text || "";
  $("word").hidden = !text;
}

// fail says what went wrong; a 401 for a signed-in member shows the door.
function fail(err) {
  if (err.status === 401 && me) {
    me = null;
    known = true;
    render();
    notice(t("app.sessionEnded"));
    return;
  }
  notice(err.message);
}

// attempt runs an action and says what went wrong, if anything did.
async function attempt(action) {
  notice("");
  word("");
  try {
    await action();
  } catch (err) {
    fail(err);
  }
}

// ---- small DOM helpers -----------------------------------------------------

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text) node.textContent = text;
  return node;
}

function button(label, run, className) {
  const b = el("button", className || "", label);
  b.type = "button";
  b.addEventListener("click", () => attempt(run));
  return b;
}

// twoStep is a button that asks again before it does something that cannot
// be undone: the first press changes its label, the second acts.
function twoStep(label, sure, run, className) {
  const b = el("button", className || "", label);
  b.type = "button";
  let armed = false;
  b.addEventListener("click", () => {
    if (!armed) {
      armed = true;
      b.textContent = sure;
      b.classList.add("danger");
      return;
    }
    disarm();
    attempt(run);
  });
  // It disarms when the member goes elsewhere, and by itself after a few
  // seconds: not every browser moves the focus to a button that is clicked.
  let timer = 0;
  const disarm = () => {
    clearTimeout(timer);
    armed = false;
    b.textContent = label;
    b.classList.remove("danger");
  };
  b.addEventListener("click", () => {
    clearTimeout(timer);
    if (armed) timer = setTimeout(disarm, 5000);
  });
  b.addEventListener("blur", disarm);
  return b;
}

function link(text, href, className) {
  const a = el("a", className || "", text);
  a.href = href;
  a.addEventListener("click", (ev) => {
    if (ev.button !== 0 || ev.metaKey || ev.ctrlKey || ev.shiftKey || ev.altKey)
      return;
    ev.preventDefault();
    go(href);
  });
  return a;
}

function input(name, attrs) {
  const i = el("input");
  i.name = name;
  Object.assign(i, attrs || {});
  i.defaultValue = i.value;
  return i;
}

function area(name, attrs) {
  const a = el("textarea");
  a.name = name;
  a.rows = 3;
  Object.assign(a, attrs || {});
  a.defaultValue = a.value;
  return a;
}

function select(name, options, value) {
  const s = el("select");
  s.name = name;
  fill(s, options, value);
  s.dataset.first = s.value;
  return s;
}

// fill (re)writes a select's options, [value, label] each, and keeps value
// when it is still offered.
function fill(s, options, value) {
  s.replaceChildren(
    ...options.map(([v, label]) => {
      const o = el("option", "", label);
      o.value = v;
      return o;
    }),
  );
  if (value !== undefined && options.some(([v]) => v === value))
    s.value = value;
}

function field(label, control) {
  const l = el("label");
  l.append(el("span", "", label), control);
  return l;
}

function fields(f) {
  return Object.fromEntries(new FormData(f).entries());
}

function form(children, submit, onSubmit) {
  const f = el("form");
  f.append(...children);
  if (submit) {
    const b = el("button", "primary", submit);
    b.type = "submit";
    f.append(b);
  }
  f.addEventListener("submit", (ev) => {
    ev.preventDefault();
    attempt(async () => {
      f.dataset.sent = "1"; // what it held is sent: a redraw starts it afresh
      try {
        await onSubmit(fields(f), f);
      } catch (err) {
        delete f.dataset.sent;
        throw err;
      }
    });
  });
  return f;
}

function section(title, ...children) {
  const s = el("section");
  s.append(el("h2", "", title), ...children);
  return s;
}

function rows(items) {
  const ul = el("ul", "rows");
  for (const item of items) {
    const li = el("li");
    li.append(...item);
    ul.append(li);
  }
  return ul;
}

function badge(text, className) {
  return el("span", "badge" + (className ? " " + className : ""), text);
}

function actions(...buttons) {
  const box = el("span", "moves");
  box.append(...buttons);
  return box;
}

function empty(text) {
  return el("p", "soft", text);
}

// ---- drafts ---------------------------------------------------------------

// drafts is what the member typed in the view and did not send, and which
// folded parts are open: an action that redraws the view must not lose it.
// Passwords are never kept.
function drafts() {
  const typed = [];
  $("view")
    .querySelectorAll("form")
    .forEach((f, at) => {
      if (f.dataset.sent) return;
      for (const c of f.elements) {
        if (!c.name || c.type === "password") continue;
        const first = c.tagName === "SELECT" ? c.dataset.first : c.defaultValue;
        if (c.value !== first) typed.push([at, c.name, c.value]);
      }
    });
  const open = [...$("view").querySelectorAll("details")].map((d) => d.open);
  return { where: drawn, typed, open };
}

// redraft puts drafts back in the view drawn anew, when it is the same view.
function redraft(kept) {
  if (kept.where !== location.pathname + location.search) return;
  const forms = $("view").querySelectorAll("form");
  for (const [at, name, value] of kept.typed) {
    const c = forms[at] && forms[at].elements[name];
    if (!c || c.type === "password") continue;
    c.value = value;
    if (c.tagName === "SELECT")
      c.dispatchEvent(new Event("change", { bubbles: true }));
  }
  $("view")
    .querySelectorAll("details")
    .forEach((d, at) => {
      if (kept.open[at]) d.open = true;
    });
}

// ---- router ---------------------------------------------------------------

function go(href, replace) {
  notice("");
  word("");
  moved = true;
  history[replace ? "replaceState" : "pushState"](null, "", href);
  return render();
}

async function render() {
  try {
    await draw();
  } catch (err) {
    fail(err);
  }
}

function chrome() {
  document.documentElement.lang = lang();
  document.title = t("app.title");
  const items = [
    ["/", "nav.tasks"],
    ["/contacts", "nav.contacts"],
    ["/groups", "nav.groups"],
    ["/activity", "nav.activity"],
    ["/settings", "nav.settings"],
  ];
  const path = location.pathname;
  const here =
    path.startsWith("/task") || path === "/"
      ? "/"
      : path.startsWith("/group")
        ? "/groups"
        : path;
  $("nav").setAttribute("aria-label", t("nav.label"));
  $("nav").replaceChildren(
    ...items.map(([href, key]) => {
      const a = link(t(key), href);
      if (href === here) a.setAttribute("aria-current", "page");
      return a;
    }),
  );
  $("nav").hidden = !me;
  $("who").hidden = !me;
  $("who-name").textContent = me ? me.name : "";
  $("logout").textContent = t("nav.logout");
  $("lang").textContent = t("lang.switch");
  $("lang").hidden = Boolean(me);
}

async function ensureMe() {
  try {
    me = await call("GET", "/api/auth/me");
  } catch {
    me = null;
  }
  known = true;
}

async function draw() {
  const mine = ++epoch;
  const path = location.pathname;
  chrome();
  let nodes;
  if (path === "/verify") {
    nodes = await verifyPage();
  } else if (path === "/forgot") {
    nodes = [forgotPage()];
  } else if (path === "/reset") {
    nodes = [resetPage()];
  } else {
    if (!known) await ensureMe();
    nodes = me ? await signedIn(path) : [door()];
  }
  if (mine !== epoch) return;
  chrome();
  const kept = drafts();
  $("view").replaceChildren(...nodes);
  drawn = location.pathname + location.search;
  redraft(kept);
  if (moved) {
    moved = false;
    $("view").focus({ preventScroll: true });
    window.scrollTo(0, 0);
  }
}

// signedIn draws the member's page for a path; a refusal that is not the end
// of the session leaves a way back.
async function signedIn(path) {
  try {
    const m = path.match(/^\/(task|group)\/([^/]+)$/);
    if (m && m[1] === "task") return await taskPage(decodeURIComponent(m[2]));
    if (m) return await groupPage(decodeURIComponent(m[2]));
    switch (path) {
      case "/contacts":
        return await contactsPage();
      case "/groups":
        return await groupsPage();
      case "/activity":
        return await activityPage();
      case "/settings":
        return await settingsPage();
      default:
        return await tasksPage(new URLSearchParams(location.search));
    }
  } catch (err) {
    if (err.status === 401) throw err;
    notice(err.message);
    return [link(t("page.back"), "/")];
  }
}

// ---- the door: sign in, sign up -------------------------------------------

async function signIn(email, password) {
  const res = await call("POST", "/api/auth/login", {
    email,
    password,
    timeZone: browserZone(),
  });
  me = res.user;
  known = true;
  const stay = publicPages.includes(location.pathname)
    ? "/"
    : location.pathname + location.search;
  await go(stay, true);
}

function door() {
  const resend = el("div", "resend");
  const offerResend = (email) => {
    resend.replaceChildren(
      el("p", "", t("door.unverified", { email })),
      button(t("door.resend"), async () => {
        await call("POST", "/api/auth/verify/resend", { email });
        word(t("door.resent", { email }));
      }),
    );
  };

  const login = form(
    [
      el("h2", "", t("door.signin")),
      field(
        t("f.email"),
        input("email", {
          type: "email",
          autocomplete: "username",
          required: true,
        }),
      ),
      field(
        t("f.password"),
        input("password", {
          type: "password",
          autocomplete: "current-password",
          required: true,
        }),
      ),
    ],
    t("door.signin"),
    async (f) => {
      try {
        await signIn(f.email, f.password);
      } catch (err) {
        if (err.code === "email_unverified") offerResend(f.email);
        throw err;
      }
    },
  );
  const forgot = el("p");
  forgot.append(link(t("door.forgot"), "/forgot"));
  login.append(forgot);

  const signup = form(
    [
      el("h2", "", t("door.signup")),
      field(
        t("f.name"),
        input("name", { autocomplete: "name", required: true, maxLength: 80 }),
      ),
      field(
        t("f.email"),
        input("email", {
          type: "email",
          autocomplete: "username",
          required: true,
        }),
      ),
      field(
        t("f.passwordNew"),
        input("password", {
          type: "password",
          autocomplete: "new-password",
          required: true,
          minLength: 10,
          maxLength: 128,
        }),
      ),
      field(
        t("f.language"),
        select(
          "locale",
          [
            ["fr", t("lang.fr")],
            ["en", t("lang.en")],
          ],
          pref,
        ),
      ),
    ],
    t("door.create"),
    async (f, node) => {
      const res = await call("POST", "/api/auth/signup", {
        email: f.email,
        name: f.name,
        password: f.password,
        locale: f.locale,
        timeZone: browserZone(),
      });
      if (res.status === "verified") {
        await signIn(f.email, f.password);
        return;
      }
      node.reset();
      word(t("door.sent", { email: res.email }));
      offerResend(res.email);
    },
  );

  const box = el("div", "door");
  box.append(login, signup, resend);
  return box;
}

function forgotPage() {
  const f = form(
    [
      el("h2", "", t("forgot.title")),
      field(
        t("f.email"),
        input("email", {
          type: "email",
          autocomplete: "username",
          required: true,
        }),
      ),
    ],
    t("forgot.send"),
    async (v, node) => {
      await call("POST", "/api/auth/password/forgot", { email: v.email });
      node.reset();
      word(t("forgot.sent"));
    },
  );
  const back = el("p");
  back.append(link(t("page.back"), "/"));
  f.append(back);
  return f;
}

function resetPage() {
  const token = new URLSearchParams(location.search).get("token") || "";
  if (!token) return el("p", "", t("link.incomplete"));
  return form(
    [
      el("h2", "", t("reset.title")),
      field(
        t("f.passwordNew"),
        input("password", {
          type: "password",
          autocomplete: "new-password",
          required: true,
          minLength: 10,
          maxLength: 128,
        }),
      ),
      field(
        t("reset.confirm"),
        input("again", {
          type: "password",
          autocomplete: "new-password",
          required: true,
          minLength: 10,
          maxLength: 128,
        }),
      ),
    ],
    t("reset.submit"),
    async (f) => {
      if (f.password !== f.again) throw new Error(t("reset.mismatch"));
      await call("POST", "/api/auth/password/reset", {
        token,
        password: f.password,
      });
      me = null;
      known = true;
      await go("/", true);
      word(t("reset.done"));
    },
  );
}

// verifyPage proves the address with the link's token and signs the member
// in; a link that does not work offers a new one.
async function verifyPage() {
  const token = new URLSearchParams(location.search).get("token") || "";
  if (!token)
    return [el("p", "", t("link.incomplete")), link(t("page.back"), "/")];
  try {
    const res = await call("POST", "/api/auth/verify", { token });
    me = res.user;
    known = true;
    history.replaceState(null, "", "/");
    word(t("verify.done", { name: me.name }));
    return await signedIn("/");
  } catch (err) {
    if (err.status === 401) throw err;
    notice(err.message);
    return [
      form(
        [
          el("h2", "", t("verify.failed")),
          field(
            t("f.email"),
            input("email", {
              type: "email",
              autocomplete: "username",
              required: true,
            }),
          ),
        ],
        t("door.resend"),
        async (f) => {
          await call("POST", "/api/auth/verify/resend", { email: f.email });
          word(t("door.resent", { email: f.email }));
        },
      ),
    ];
  }
}

// ---- settings: the account, the password, the sessions -----------------------

async function settingsPage() {
  const { sessions } = await call("GET", "/api/auth/sessions");
  const zones = el("datalist");
  zones.id = "zones";
  if (typeof Intl.supportedValuesOf === "function") {
    for (const z of Intl.supportedValuesOf("timeZone")) {
      const o = el("option");
      o.value = z;
      zones.append(o);
    }
  }
  const tz = input("timeZone", {
    value: me.timeZone || "",
    placeholder: "UTC",
  });
  tz.setAttribute("list", "zones");

  const profile = form(
    [
      field(
        t("f.name"),
        input("name", { value: me.name, required: true, maxLength: 80 }),
      ),
      field(
        t("f.language"),
        select(
          "locale",
          [
            ["fr", t("lang.fr")],
            ["en", t("lang.en")],
          ],
          lang(),
        ),
      ),
      field(t("f.timeZone"), tz),
      zones,
    ],
    t("settings.save"),
    async (f) => {
      me = await call("PATCH", "/api/auth/me", {
        name: f.name.trim(),
        locale: f.locale,
        timeZone: f.timeZone.trim(),
      });
      await render();
      word(t("settings.saved"));
    },
  );

  const password = form(
    [
      field(
        t("settings.current"),
        input("current", {
          type: "password",
          autocomplete: "current-password",
          required: true,
        }),
      ),
      field(
        t("f.passwordNew"),
        input("password", {
          type: "password",
          autocomplete: "new-password",
          required: true,
          minLength: 10,
          maxLength: 128,
        }),
      ),
    ],
    t("settings.change"),
    async (f, node) => {
      await call("POST", "/api/auth/password", {
        current: f.current,
        password: f.password,
      });
      node.reset();
      word(t("settings.changed"));
    },
  );

  const list = rows(
    sessions.map((s) => {
      const head = el("span", "grow");
      head.append(el("span", "title", s.userAgent || t("session.unknown")));
      head.append(
        el(
          "span",
          "meta",
          [
            s.ip,
            t("session.seen", { date: when(s.lastSeenAt) }),
            t("session.since", { date: when(s.createdAt) }),
          ]
            .filter(Boolean)
            .join(" · "),
        ),
      );
      const line = [head];
      if (s.current) {
        line.push(badge(t("session.current"), "ok"));
      } else {
        line.push(
          actions(
            button(t("session.revoke"), async () => {
              await call(
                "DELETE",
                "/api/auth/sessions/" + encodeURIComponent(s.id),
              );
              await render();
              word(t("session.revoked"));
            }),
          ),
        );
      }
      return line;
    }),
  );

  return [
    section(t("settings.account"), el("p", "soft", me.email), profile),
    section(t("settings.password"), password),
    section(t("settings.sessions"), list),
  ];
}

// ---- start ----------------------------------------------------------------

$("logout").addEventListener("click", () =>
  attempt(async () => {
    await call("POST", "/api/auth/logout");
    me = null;
    known = true;
    await go("/", true);
    word(t("app.signedOut"));
  }),
);

$("lang").addEventListener("click", () => {
  pref = pref === "fr" ? "en" : "fr";
  render();
});

$("brand").addEventListener("click", (ev) => {
  if (ev.button !== 0 || ev.metaKey || ev.ctrlKey || ev.shiftKey || ev.altKey)
    return;
  ev.preventDefault();
  go("/");
});

window.addEventListener("popstate", () => {
  notice("");
  word("");
  moved = true;
  render();
});

render();
