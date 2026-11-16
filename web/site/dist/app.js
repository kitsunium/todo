// The todo's web app: one page that talks to the HTTP API with the session
// cookie the sign-in sets. No build step, no dependency.
"use strict";

const $ = (id) => document.getElementById(id);
const priorities = { high: "haute", normal: "normale", low: "basse" };
const kinds = { added: "ajoutée", complete: "faite", reopen: "rouverte", archive: "archivée", restore: "restaurée" };
const people = { "contact-requested": "vous demande en contact", "contact-accepted": "a accepté votre demande de contact" };
const moves = {
  open: [["complete", "Fait"]],
  done: [["reopen", "Rouvrir"], ["archive", "Archiver"]],
  archived: [["restore", "Restaurer"]],
};
let state = "";

// call sends one request to the API and answers its JSON; a refusal is
// thrown with the API's own message, and a 401 shows the door.
async function call(method, path, body) {
  const res = await fetch(path, {
    method,
    credentials: "same-origin",
    headers: body ? { "Content-Type": "application/json" } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await res.json().catch(() => ({}));
  if (res.ok) return data;
  const err = new Error((data.error && data.error.message) || res.statusText);
  err.status = res.status;
  if (data.error && data.error.violations && data.error.violations.length) {
    err.message = data.error.violations.map((v) => v.message).join(", ");
  }
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

// attempt runs an action and says what went wrong, if anything did.
async function attempt(action) {
  notice("");
  word("");
  try {
    await action();
  } catch (err) {
    if (err.status === 401 && $("door").hidden) show(null); // the session ended
    notice(err.message);
  }
}

function show(me, page) {
  $("door").hidden = Boolean(me) || Boolean(page);
  $("lost").hidden = page !== "lost";
  $("renew").hidden = page !== "renew";
  $("desk").hidden = !me;
  $("who").hidden = !me;
  $("who-name").textContent = me ? me.name : "";
}

function when(text) {
  const t = new Date(text);
  return t.getFullYear() < 2 ? "" : t.toLocaleString("fr-FR", { dateStyle: "medium", timeStyle: "short" });
}

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text) node.textContent = text;
  return node;
}

async function refresh() {
  const [list, feed] = await Promise.all([
    call("GET", "/api/tasks" + (state ? "?state=" + state : "")),
    call("GET", "/api/activity?limit=20"),
  ]);
  const now = Date.now();
  $("tasks").replaceChildren(...list.tasks.map((t) => {
    const li = el("li", t.state);
    li.append(el("span", "title", t.title));
    const actions = el("span", "moves");
    for (const [event, label] of moves[t.state] || []) {
      const b = el("button", "", label);
      b.type = "button";
      b.addEventListener("click", () => attempt(async () => {
        await call("POST", "/api/tasks/" + encodeURIComponent(t.id) + "/" + event);
        await refresh();
      }));
      actions.append(b);
    }
    li.append(actions);
    const due = when(t.due);
    const late = due && t.state === "open" && new Date(t.due).getTime() < now;
    const meta = ["priorité " + (priorities[t.priority] || t.priority)];
    if (due) meta.push((late ? "en retard depuis le " : "pour le ") + due);
    li.append(el("span", "meta" + (late ? " late" : ""), meta.join(" · ")));
    return li;
  }));
  $("empty").hidden = list.tasks.length > 0;
  $("feed").replaceChildren(...feed.entries.map((e) =>
    el("li", "", when(e.at) + " — " + (people[e.kind] ? e.title + " " + people[e.kind] : "« " + e.title + " » " + (kinds[e.kind] || e.kind)))));
}

function fields(form) {
  return Object.fromEntries(new FormData(form).entries());
}

async function enter(creds) {
  await call("POST", "/api/auth/login", { email: creds.email, password: creds.password });
  show(await call("GET", "/api/auth/me"));
  await refresh();
}

$("login").addEventListener("submit", (ev) => {
  ev.preventDefault();
  attempt(() => enter(fields(ev.target)));
});

$("signup").addEventListener("submit", (ev) => {
  ev.preventDefault();
  attempt(async () => {
    const f = fields(ev.target);
    const res = await call("POST", "/api/auth/signup", { email: f.email, name: f.name, password: f.password });
    if (res.status === "verified") {
      await enter(f);
      return;
    }
    ev.target.reset();
    word("Un lien de confirmation vient d’être envoyé à " + res.email + ". Ouvrez-le, puis connectez-vous.");
  });
});

$("logout").addEventListener("click", () => attempt(async () => {
  await call("POST", "/api/auth/logout");
  show(null);
}));

$("add").addEventListener("submit", (ev) => {
  ev.preventDefault();
  attempt(async () => {
    const f = fields(ev.target);
    const task = { title: f.title, priority: f.priority };
    if (f.due) task.due = new Date(f.due).toISOString();
    await call("POST", "/api/tasks", task);
    ev.target.reset();
    await refresh();
  });
});

for (const b of document.querySelectorAll("#states button")) {
  b.addEventListener("click", () => attempt(async () => {
    state = b.dataset.state;
    for (const other of document.querySelectorAll("#states button")) {
      other.setAttribute("aria-pressed", String(other === b));
    }
    await refresh();
  }));
}

$("forgot").addEventListener("submit", (ev) => {
  ev.preventDefault();
  attempt(async () => {
    await call("POST", "/api/auth/password/forgot", { email: fields(ev.target).email });
    ev.target.reset();
    word("Si cette adresse a un compte, un lien vient de lui être envoyé.");
  });
});

$("reset").addEventListener("submit", (ev) => {
  ev.preventDefault();
  attempt(async () => {
    await call("POST", "/api/auth/password/reset", { token: token, password: fields(ev.target).password });
    history.replaceState(null, "", "/");
    show(null);
    word("Votre mot de passe est changé. Connectez-vous.");
  });
});

// At load: the link of a mail does what it is for; else a session still
// open goes straight to the desk.
const token = new URLSearchParams(location.search).get("token") || "";
if (location.pathname === "/verify" && token) {
  show(null);
  attempt(async () => {
    await call("POST", "/api/auth/verify", { token: token });
    history.replaceState(null, "", "/");
    word("Votre adresse est confirmée. Connectez-vous.");
  });
} else if (location.pathname === "/reset" && token) {
  show(null, "renew");
} else if (location.pathname === "/forgot") {
  show(null, "lost");
} else {
  call("GET", "/api/auth/me").then(async (me) => {
    show(me);
    await refresh();
  }, () => show(null));
}
