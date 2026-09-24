// The todo list's user interface. It only talks to the product's public API
// — /todos, /activity, /stats — the same way any client would; kit sees each
// call arrive from this page and draws it as an edge of the diagram.
//
// Everything a user typed is rendered with textContent, never as HTML.
"use strict";

const $ = (id) => document.getElementById(id);
const state = { filter: "", todos: [], busy: false, lastError: 0 };

// Rows are rebuilt on every refresh; only the ones not shown before animate
// in, so the list does not flicker every two seconds.
const seen = { todos: new Set(), activity: new Set() };
function isNew(kind, id) {
  const fresh = seen[kind].size > 0 && !seen[kind].has(id);
  seen[kind].add(id);
  return fresh;
}

async function api(method, path, body) {
  const init = { method, headers: {} };
  if (body !== undefined) {
    init.headers["Content-Type"] = "application/json";
    init.body = JSON.stringify(body);
  }
  const res = await fetch(path, init);
  if (res.status === 204) return null;
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    const err = new Error(
      (data && data.error && data.error.message) || res.statusText,
    );
    err.details = data && data.error;
    throw err;
  }
  return data;
}

// ── Rendering ────────────────────────────────────────────────────────────────

function relative(date) {
  const s = Math.round((new Date(date).getTime() - Date.now()) / 1000);
  const abs = Math.abs(s);
  const units = [
    ["day", 86400],
    ["hour", 3600],
    ["minute", 60],
    ["second", 1],
  ];
  for (const [unit, size] of units) {
    if (abs >= size || unit === "second") {
      return new Intl.RelativeTimeFormat("en", { numeric: "auto" }).format(
        Math.round(s / size),
        unit,
      );
    }
  }
  return "";
}

function renderTodos() {
  const list = $("list");
  const shown = state.todos.filter(
    (t) => !state.filter || t.status === state.filter,
  );
  const tpl = $("todo-row");
  list.replaceChildren(
    ...shown.map((t) => {
      const row = tpl.content.firstElementChild.cloneNode(true);
      row.classList.add(t.status);
      row.dataset.id = t.id;
      if (isNew("todos", t.id)) row.classList.add("new");
      row.querySelector(".title").textContent = t.title;
      const meta = row.querySelector(".meta");
      if (t.status === "done" && t.doneAt) {
        meta.textContent =
          "done " + relative(t.doneAt) + " · archives itself soon";
      } else if (t.due) {
        meta.textContent =
          (t.status === "overdue" ? "was due " : "due ") + relative(t.due);
        meta.classList.toggle("late", t.status === "overdue");
      } else {
        meta.textContent = "added " + relative(t.createdAt);
      }
      const pill = row.querySelector(".pill");
      pill.textContent = t.status;
      pill.classList.add(t.status);
      const check = row.querySelector(".check");
      check.setAttribute(
        "aria-label",
        t.status === "done" ? "Reopen" : "Complete",
      );
      check.disabled = t.status === "archived";
      check.addEventListener("click", () =>
        act(t.status === "done" ? "reopen" : "complete", t.id),
      );
      const archive = row.querySelector(".archive");
      archive.hidden = t.status !== "done";
      archive.addEventListener("click", () => act("archive", t.id));
      row
        .querySelector(".delete")
        .addEventListener("click", () => remove(t.id));
      return row;
    }),
  );
  $("empty").hidden = shown.length > 0;
  const counts = { "": state.todos.length };
  for (const t of state.todos) counts[t.status] = (counts[t.status] || 0) + 1;
  for (const span of document.querySelectorAll("[data-count]")) {
    span.textContent = String(counts[span.dataset.count] || 0);
  }
}

const eventWords = {
  create: "Added",
  complete: "Completed",
  reopen: "Reopened",
  archive: "Archived",
  "auto-archive": "Archived automatically",
  overdue: "Became overdue",
};

function renderActivity(entries) {
  $("activity").replaceChildren(
    ...entries.map((e) => {
      const li = document.createElement("li");
      if (isNew("activity", e.id)) li.classList.add("new");
      const dot = document.createElement("span");
      dot.className = "dot " + e.event;
      const text = document.createElement("div");
      const what = document.createElement("div");
      what.className = "what";
      what.append(
        document.createTextNode((eventWords[e.event] || e.event) + " · "),
      );
      const title = document.createElement("strong");
      title.textContent = e.title;
      what.append(title);
      if (e.event === "auto-archive" || e.event === "overdue") {
        const chip = document.createElement("span");
        chip.className = "chip daemon";
        chip.textContent = "daemon";
        what.append(" ", chip);
      }
      const when = document.createElement("div");
      when.className = "when";
      when.textContent =
        relative(e.at) + (e.from ? " · " + e.from + " → " + e.to : "");
      text.append(what, when);
      li.append(dot, text);
      return li;
    }),
  );
}

function renderStats(stats) {
  const latest = stats.latest;
  const counts = (latest && latest.counts) || {};
  for (const k of ["open", "overdue", "done", "archived"]) {
    $("stat-" + k).textContent = latest ? String(counts[k] || 0) : "–";
  }
  $("stat-rate").textContent = latest
    ? Math.round(latest.completionRate * 100) + "%"
    : "–";
  const svg = $("sparkline");
  const points = (stats.history || []).map((s) => s.completionRate);
  svg.replaceChildren();
  if (points.length < 2) return;
  const ns = "http://www.w3.org/2000/svg";
  const step = 120 / (points.length - 1);
  const coords = points.map((p, i) => [i * step, 30 - p * 28]);
  const line = coords
    .map(([x, y], i) => (i ? "L" : "M") + x.toFixed(1) + " " + y.toFixed(1))
    .join(" ");
  const area = document.createElementNS(ns, "path");
  area.setAttribute("class", "area");
  area.setAttribute("d", line + " L120 32 L0 32 Z");
  const path = document.createElementNS(ns, "path");
  path.setAttribute("d", line);
  svg.append(area, path);
}

// ── Actions ──────────────────────────────────────────────────────────────────

async function refresh() {
  try {
    const [todos, activity, stats] = await Promise.all([
      api("GET", "/todos"),
      api("GET", "/activity"),
      api("GET", "/stats"),
    ]);
    state.todos = todos.todos;
    renderTodos();
    renderActivity(activity.entries);
    renderStats(stats);
    setConn(true);
  } catch {
    setConn(false);
  }
}

async function act(event, id) {
  try {
    await api("POST", "/todos/" + encodeURIComponent(id) + "/" + event);
  } catch (err) {
    showError(err.message);
  }
  refresh();
}

async function remove(id) {
  try {
    await api("DELETE", "/todos/" + encodeURIComponent(id));
  } catch (err) {
    showError(err.message);
  }
  refresh();
}

function showError(message) {
  const el = $("form-error");
  el.textContent = message;
  el.hidden = false;
  state.lastError = Date.now();
  setTimeout(() => {
    if (Date.now() - state.lastError >= 4000) el.hidden = true;
  }, 4000);
}

function setConn(ok) {
  const el = $("conn");
  el.textContent = ok ? "live" : "reconnecting…";
  el.className = "conn " + (ok ? "ok" : "down");
}

$("create").addEventListener("submit", async (ev) => {
  ev.preventDefault();
  if (state.busy) return;
  const title = $("title").value.trim();
  const dueRaw = $("due").value;
  const body = { title };
  if (dueRaw) body.due = new Date(dueRaw).toISOString();
  state.busy = true;
  ev.submitter && (ev.submitter.disabled = true);
  try {
    await api("POST", "/todos", body);
    $("title").value = "";
    $("due").value = "";
    $("form-error").hidden = true;
  } catch (err) {
    const v = err.details && err.details.violations;
    showError(
      v && v.length
        ? v.map((x) => x.path + ": " + x.message).join(", ")
        : err.message,
    );
  } finally {
    state.busy = false;
    ev.submitter && (ev.submitter.disabled = false);
  }
  refresh();
});

for (const button of document.querySelectorAll(".filters button")) {
  button.addEventListener("click", () => {
    state.filter = button.dataset.filter;
    for (const b of document.querySelectorAll(".filters button"))
      b.classList.toggle("active", b === button);
    renderTodos();
  });
}

// In dev, kit serves the product's diagram at /_kit/: show the way there.
fetch("/_kit/api/graph", { method: "HEAD" })
  .then((res) => {
    if (res.ok) $("diagram-link").hidden = false;
  })
  .catch(() => {});

refresh();
setInterval(() => {
  if (!document.hidden) refresh();
}, 2000);
