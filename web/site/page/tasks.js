// The tasks: the views and filters, the counts, the task page, the feed and
// the vital signs. Loaded before app.js, whose helpers it calls.
"use strict";

const taskViews = [
  "inbox",
  "today",
  "upcoming",
  "shared",
  "assigned",
  "completed",
];
const taskStates = [
  ["", "state.all"],
  ["open", "state.open"],
  ["done", "state.done"],
  ["archived", "state.archived"],
];
const taskMoves = {
  open: ["complete"],
  done: ["reopen", "archive"],
  archived: ["restore"],
};
const priorities = ["normal", "high", "low"];

// pathOf builds the address of a tasks view.
function pathOf(view, state, group) {
  const q = new URLSearchParams();
  if (view) q.set("view", view);
  else if (state) q.set("state", state);
  if (view === "group" && group) q.set("group", group);
  const s = q.toString();
  return s ? "/?" + s : "/";
}

// peopleOf is every member's name the page knows, by id: first the people the
// answer names (every owner, assignee and member a task is shared with, which
// no contact or group list may hold), then the member, the contacts and the
// members of the groups.
function peopleOf(answered, contacts, groups) {
  const names = new Map();
  for (const p of answered || []) names.set(p.id, p.name);
  const know = (id, name) => {
    if (!names.has(id)) names.set(id, name);
  };
  for (const g of groups)
    for (const m of g.members || []) know(m.user.id, m.user.name);
  for (const c of (contacts && contacts.contacts) || [])
    know(c.user.id, c.user.name);
  know(me.id, me.name);
  return names;
}

// assignees are who a task may be assigned to: the member, those it is
// shared with, the members of its group.
function assignees(names, groups, groupId, shared) {
  const ids = new Map([[me.id, names.get(me.id)]]);
  for (const id of shared || []) ids.set(id, names.get(id) || id);
  const g = groups.find((x) => x.id === groupId);
  if (g) for (const m of g.members) ids.set(m.user.id, m.user.name);
  return [["", t("task.unassigned")], ...ids.entries()];
}

function groupChoices(groups) {
  return [["", t("task.noGroup")], ...groups.map((g) => [g.id, g.name])];
}

function priorityChoices() {
  return priorities.map((p) => [p, t("priority." + p)]);
}

async function tasksPage(q) {
  const view = q.get("view") || "";
  const state = view ? "" : q.get("state") || "";
  const group = q.get("group") || "";
  const params = new URLSearchParams();
  if (view) params.set("view", view);
  else if (state) params.set("state", state);
  if (view === "group" && group) params.set("group", group);
  params.set("tz", zone());
  const [list, counts, mine, contacts] = await Promise.all([
    call("GET", "/api/tasks?" + params),
    call("GET", "/api/tasks/counts?tz=" + encodeURIComponent(zone())),
    call("GET", "/api/groups"),
    call("GET", "/api/contacts"),
  ]);
  const groups = mine.groups;
  const names = peopleOf(list.people, contacts, groups);

  const here = view === "group" ? "group:" + group : view;
  const entry = (label, href, key, count) => {
    const a = link(label, href);
    if (key === here) a.setAttribute("aria-current", "page");
    if (count !== undefined) a.append(" ", el("span", "count", String(count)));
    return [a];
  };
  const side = [entry(t("view.all"), pathOf("", ""), "", undefined)];
  for (const v of taskViews)
    side.push(
      entry(
        t("view." + v),
        pathOf(v),
        v,
        v === "completed" ? counts.completedThisWeek : counts[v],
      ),
    );
  for (const g of groups)
    side.push(
      entry(
        g.name,
        pathOf("group", "", g.id),
        "group:" + g.id,
        (counts.groups || {})[g.id] || 0,
      ),
    );
  const nav = el("nav", "side");
  nav.setAttribute("aria-label", t("view.label"));
  nav.append(rows(side));
  if (counts.overdue)
    nav.append(
      el("p", "meta late", t("tasks.overdue", { count: counts.overdue })),
    );

  const body = el("div", "main");
  body.append(addTaskForm(groups, names, view === "group" ? group : ""));
  const title =
    view === "group"
      ? (groups.find((g) => g.id === group) || { name: t("view.group") }).name
      : t("view." + (view || "all"));
  body.append(el("h2", "", title));
  if (!view) {
    const bar = el("div", "chips");
    for (const [value, key] of taskStates) {
      const a = link(t(key), pathOf("", value));
      if (value === state) a.setAttribute("aria-current", "page");
      bar.append(a);
    }
    body.append(bar);
  }
  const items = list.tasks;
  body.append(
    items.length ? taskList(items, groups, names) : empty(t("tasks.empty")),
  );

  const wrap = el("div", "cols");
  wrap.append(nav, body);
  return [wrap];
}

function addTaskForm(groups, names, group) {
  const assignee = select("assigneeId", assignees(names, groups, group, []));
  const where = select("groupId", groupChoices(groups), group);
  where.addEventListener("change", () =>
    fill(assignee, assignees(names, groups, where.value, []), assignee.value),
  );
  const f = form(
    [
      field(t("f.title"), input("title", { required: true, maxLength: 120 })),
      field(t("f.notes"), area("notes")),
      field(t("f.priority"), select("priority", priorityChoices(), "normal")),
      field(t("f.due"), input("due", { type: "datetime-local" })),
      field(t("f.group"), where),
      field(t("f.assignee"), assignee),
    ],
    t("tasks.add"),
    async (v) => {
      const body = { title: v.title.trim(), priority: v.priority };
      if (v.notes) body.notes = v.notes;
      if (v.due) body.due = toInstant(v.due, zone());
      if (v.groupId) body.groupId = v.groupId;
      if (v.assigneeId) body.assigneeId = v.assigneeId;
      await call("POST", "/api/tasks", body);
      await render();
      word(t("tasks.added"));
    },
  );
  const box = el("details", "add");
  box.append(el("summary", "", t("tasks.new")), f);
  return box;
}

function taskList(items, groups, names) {
  const now = Date.now();
  return rows(
    items.map((task) => {
      const head = el("span", "grow");
      head.append(
        link(
          task.title,
          "/task/" + encodeURIComponent(task.id),
          "title " + task.state,
        ),
      );
      const meta = [];
      if (task.state !== "open") meta.push(t("state." + task.state));
      meta.push(t("priority." + task.priority));
      const late =
        task.state === "open" &&
        isDate(task.due) &&
        new Date(task.due).getTime() < now;
      if (isDate(task.due))
        meta.push(
          t(late ? "task.late" : "task.dueOn", { date: when(task.due) }),
        );
      const g = groups.find((x) => x.id === task.groupId);
      if (g) meta.push(g.name);
      if (task.assigneeId)
        meta.push(
          t("task.assignedTo", {
            name:
              task.assigneeId === me.id
                ? t("you")
                : names.get(task.assigneeId) || "?",
          }),
        );
      head.append(el("span", "meta" + (late ? " late" : ""), meta.join(" · ")));
      return [
        head,
        actions(
          ...(taskMoves[task.state] || []).map((ev) => moveButton(task.id, ev)),
        ),
      ];
    }),
  );
}

function moveButton(id, event) {
  return button(t("task." + event), async () => {
    await call("POST", "/api/tasks/" + encodeURIComponent(id) + "/" + event);
    await render();
  });
}

// taskPage opens one task: its fields to edit, its moves, who it is shared
// with.
async function taskPage(id) {
  const path = "/api/tasks/" + encodeURIComponent(id);
  const [task, mine, contacts] = await Promise.all([
    call("GET", path),
    call("GET", "/api/groups"),
    call("GET", "/api/contacts"),
  ]);
  const groups = mine.groups;
  const names = peopleOf(task.people, contacts, groups);
  const owner = task.ownerId === me.id;
  const shared = task.sharedWith || [];
  const was = toLocal(task.due, zone());

  const assignee = select(
    "assigneeId",
    assignees(names, groups, task.groupId, shared),
    task.assigneeId,
  );
  if (task.assigneeId && assignee.value !== task.assigneeId) {
    const o = el("option", "", names.get(task.assigneeId) || task.assigneeId);
    o.value = task.assigneeId;
    assignee.append(o);
    assignee.value = task.assigneeId;
  }
  const where = select("groupId", groupChoices(groups), task.groupId);
  where.addEventListener("change", () =>
    fill(
      assignee,
      assignees(names, groups, where.value, shared),
      assignee.value,
    ),
  );

  const edit = form(
    [
      field(
        t("f.title"),
        input("title", { value: task.title, required: true, maxLength: 120 }),
      ),
      field(t("f.notes"), area("notes", { value: task.notes || "" })),
      field(
        t("f.priority"),
        select("priority", priorityChoices(), task.priority),
      ),
      field(t("f.due"), input("due", { type: "datetime-local", value: was })),
      field(t("f.group"), where),
      field(t("f.assignee"), assignee),
    ],
    t("task.save"),
    async (v) => {
      const body = {};
      if (v.title.trim() !== task.title) body.title = v.title.trim();
      if (v.notes !== (task.notes || "")) body.notes = v.notes;
      if (v.priority !== task.priority) body.priority = v.priority;
      if (v.due !== was) body.due = v.due ? toInstant(v.due, zone()) : zeroTime;
      if (v.groupId !== (task.groupId || "")) body.groupId = v.groupId;
      if (v.assigneeId !== (task.assigneeId || ""))
        body.assigneeId = v.assigneeId;
      if (!Object.keys(body).length) {
        word(t("task.nochange"));
        return;
      }
      await call("PATCH", path, body);
      await render();
      word(t("task.saved"));
    },
  );

  const head = el("div", "head");
  head.append(el("h2", "", task.title), badge(t("state." + task.state)));
  const moves = actions(
    ...(taskMoves[task.state] || []).map((ev) =>
      button(t("task." + ev), async () => {
        await call("POST", path + "/" + ev);
        await render();
      }),
    ),
    twoStep(t("task.delete"), t("task.deleteSure"), async () => {
      await call("DELETE", path);
      await go("/", true);
      word(t("task.deleted"));
    }),
  );
  const meta = el(
    "p",
    "meta",
    [
      t("task.created", { date: when(task.createdAt) }),
      owner
        ? t("task.yours")
        : t("task.ownedBy", { name: names.get(task.ownerId) || "?" }),
      isDate(task.doneAt) ? t("task.doneOn", { date: when(task.doneAt) }) : "",
    ]
      .filter(Boolean)
      .join(" · "),
  );

  const sharing = rows(
    shared.map((uid) => [
      el("span", "grow", names.get(uid) || uid),
      actions(
        button(t("task.unshare"), async () => {
          await call("DELETE", path + "/share/" + encodeURIComponent(uid));
          await render();
          word(t("task.unshared"));
        }),
      ),
    ]),
  );
  const choices = ((contacts && contacts.contacts) || [])
    .filter((c) => !shared.includes(c.user.id))
    .map((c) => [c.user.id, c.user.name]);
  const children = [shared.length ? sharing : empty(t("task.notShared"))];
  if (choices.length) {
    children.push(
      form(
        [field(t("task.shareWith"), select("userId", choices))],
        t("task.share"),
        async (v) => {
          await call("POST", path + "/share", { userId: v.userId });
          await render();
          word(t("task.shared"));
        },
      ),
    );
  } else {
    children.push(empty(t("task.noContacts")));
  }

  const back = el("p");
  back.append(link(t("page.back"), "/"));
  return [
    back,
    head,
    meta,
    moves,
    section(t("task.edit"), edit),
    section(t("task.sharing"), ...children),
  ];
}

// activityPage is the feed, newest first, and the vital signs.
async function activityPage() {
  const [feed, stats] = await Promise.all([
    call("GET", "/api/activity?limit=50"),
    call("GET", "/api/stats"),
  ]);
  const items = feed.entries.map((e) => {
    const key = "kind." + e.kind;
    const what = hasWord(key)
      ? t(key, { title: e.title })
      : e.title + " — " + e.kind;
    const line = el("span", e.read ? "grow" : "grow unread");
    if (e.taskId && e.kind !== "deleted" && hasWord(key)) {
      const [before, after] = what.split(e.title);
      line.append(
        before || "",
        link(e.title, "/task/" + encodeURIComponent(e.taskId)),
        after || "",
      );
    } else {
      line.textContent = what;
    }
    const state = e.read ? [] : [el("span", "sr-only", t("activity.unread"))];
    return [...state, el("span", "meta", when(e.at)), line];
  });
  const markAll = feed.entries.some((e) => !e.read)
    ? actions(
        button(t("activity.markRead"), async () => {
          const res = await call("POST", "/api/activity/read");
          unread = res.unread || 0;
          await render();
        }),
      )
    : null;
  return [
    section(
      t("activity.title"),
      ...(markAll ? [markAll] : []),
      items.length ? rows(items) : empty(t("activity.empty")),
    ),
    section(t("stats.title"), statsView(stats)),
  ];
}

function statsView(report) {
  const s = report.latest;
  if (!s) return empty(t("stats.none"));
  const counts = Object.entries(s.counts || {}).map(([state, n]) => [
    el("span", "grow", hasWord("state." + state) ? t("state." + state) : state),
    el("span", "count", String(n)),
  ]);
  const sum = el(
    "p",
    "meta",
    [
      t("stats.total", { count: s.total }),
      t("stats.overdue", { count: s.overdue }),
      t("stats.rate", { rate: Math.round(s.completionRate * 100) }),
      t("stats.at", { date: when(s.at) }),
      t("stats.samples", { count: (report.history || []).length }),
    ].join(" · "),
  );
  return fragment(sum, rows(counts));
}

function fragment(...nodes) {
  const box = el("div");
  box.append(...nodes);
  return box;
}
