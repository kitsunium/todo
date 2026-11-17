// The circle around a member: contacts and their requests, groups, their
// members and their invitations. Loaded before app.js, whose helpers it calls.
"use strict";

const groupColors = [
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
];

function person(user) {
  const box = el("span", "grow");
  box.append(
    el("span", "title", user.name || user.email),
    el("span", "meta", user.email || ""),
  );
  return box;
}

async function contactsPage() {
  const list = await call("GET", "/api/contacts");
  const answer = (id, verb, said) =>
    button(t("common." + verb), async () => {
      await call(
        "POST",
        "/api/contacts/" + encodeURIComponent(id) + "/" + verb,
      );
      await render();
      word(t(said));
    });

  const add = form(
    [field(t("f.email"), input("email", { type: "email", required: true }))],
    t("contacts.ask"),
    async (v) => {
      const res = await call("POST", "/api/contacts", { email: v.email });
      await render();
      word(
        t(res.kind === "invite" ? "contacts.invited" : "contacts.requested", {
          email: v.email,
        }),
      );
    },
  );

  const incoming = list.incoming.map((c) => [
    person(c.user),
    actions(
      answer(c.id, "accept", "contacts.accepted"),
      answer(c.id, "decline", "contacts.declined"),
    ),
  ]);
  const outgoing = list.outgoing.map((c) => [
    person(c.user),
    actions(answer(c.id, "cancel", "contacts.cancelled")),
  ]);
  const invites = list.invites.map((i) => {
    const head = el("span", "grow");
    head.append(
      el("span", "title", i.email),
      el("span", "meta", when(i.createdAt)),
    );
    return [
      head,
      actions(
        button(t("common.cancel"), async () => {
          await call(
            "DELETE",
            "/api/contacts/invites/" + encodeURIComponent(i.id),
          );
          await render();
          word(t("contacts.withdrawn"));
        }),
      ),
    ];
  });
  const contacts = list.contacts.map((c) => {
    const head = person(c.user);
    head.append(
      el(
        "span",
        "meta",
        t("contacts.since", { date: when(c.since, { dateStyle: "medium" }) }),
      ),
    );
    return [
      head,
      actions(
        twoStep(t("common.remove"), t("common.sure"), async () => {
          await call("DELETE", "/api/contacts/" + encodeURIComponent(c.id));
          await render();
          word(t("contacts.removed"));
        }),
      ),
    ];
  });

  const parts = [section(t("contacts.add"), add)];
  if (incoming.length)
    parts.push(section(t("contacts.incoming"), rows(incoming)));
  if (outgoing.length)
    parts.push(section(t("contacts.outgoing"), rows(outgoing)));
  if (invites.length)
    parts.push(
      section(
        t("contacts.invites"),
        el("p", "soft", t("contacts.invitesHint")),
        rows(invites),
      ),
    );
  parts.push(
    section(
      t("contacts.title"),
      contacts.length ? rows(contacts) : empty(t("contacts.empty")),
    ),
  );
  return parts;
}

function swatch(color) {
  return el(
    "span",
    "swatch c-" + (groupColors.includes(color) ? color : "slate"),
  );
}

async function groupsPage() {
  const [mine, received] = await Promise.all([
    call("GET", "/api/groups"),
    call("GET", "/api/invitations"),
  ]);
  const parts = [];
  if (received.invitations.length) {
    const answer = (id, verb, said) =>
      button(t("common." + verb), async () => {
        await call(
          "POST",
          "/api/invitations/" + encodeURIComponent(id) + "/" + verb,
        );
        await render();
        word(t(said));
      });
    parts.push(
      section(
        t("groups.invitations"),
        rows(
          received.invitations.map((i) => {
            const head = el("span", "grow");
            head.append(
              el("span", "title", i.group.name),
              el(
                "span",
                "meta",
                t("groups.invitedBy", { name: i.inviter.name }),
              ),
            );
            return [
              swatch(i.group.color),
              head,
              actions(
                answer(i.id, "accept", "groups.joined"),
                answer(i.id, "decline", "groups.refused"),
              ),
            ];
          }),
        ),
      ),
    );
  }

  const create = form(
    [
      field(t("f.name"), input("name", { required: true, maxLength: 60 })),
      field(
        t("f.color"),
        select("color", [
          ["", t("groups.colorAuto")],
          ...groupColors.map((c) => [c, t("color." + c)]),
        ]),
      ),
    ],
    t("groups.create"),
    async (v) => {
      const group = await call("POST", "/api/groups", {
        name: v.name.trim(),
        color: v.color,
      });
      await go("/group/" + encodeURIComponent(group.id));
      word(t("groups.created"));
    },
  );
  parts.push(section(t("groups.new"), create));

  const items = mine.groups.map((g) => {
    const head = el("span", "grow");
    head.append(
      link(g.name, "/group/" + encodeURIComponent(g.id), "title"),
      el(
        "span",
        "meta",
        t("groups.members", { count: g.members.length }) +
          " · " +
          t("role." + g.role),
      ),
    );
    return [swatch(g.color), head];
  });
  parts.push(
    section(
      t("groups.title"),
      items.length ? rows(items) : empty(t("groups.empty")),
    ),
  );
  return parts;
}

// groupPage shows a group with its members. An action is offered to who may
// do it by their role; the API refuses any other and its message is shown.
async function groupPage(id) {
  const path = "/api/groups/" + encodeURIComponent(id);
  const [group, contacts] = await Promise.all([
    call("GET", path),
    call("GET", "/api/contacts"),
  ]);
  const manager = group.role === "owner" || group.role === "admin";
  const isOwner = group.role === "owner";

  const back = el("p");
  back.append(
    link(t("groups.back"), "/groups"),
    " · ",
    link(t("groups.tasks"), pathOf("group", "", group.id)),
  );
  const head = el("div", "head");
  head.append(
    swatch(group.color),
    el("h2", "", group.name),
    badge(t("role." + group.role)),
  );
  const parts = [back, head];

  if (manager) {
    parts.push(
      section(
        t("groups.edit"),
        form(
          [
            field(
              t("f.name"),
              input("name", {
                value: group.name,
                required: true,
                maxLength: 60,
              }),
            ),
            field(
              t("f.color"),
              select(
                "color",
                groupColors.map((c) => [c, t("color." + c)]),
                group.color,
              ),
            ),
          ],
          t("groups.save"),
          async (v) => {
            await call("PATCH", path, { name: v.name.trim(), color: v.color });
            await render();
            word(t("groups.saved"));
          },
        ),
      ),
    );

    const inside = new Set(group.members.map((m) => m.user.id));
    const choices = contacts.contacts
      .filter((c) => !inside.has(c.user.id))
      .map((c) => [c.user.id, c.user.name]);
    parts.push(
      section(
        t("groups.invite"),
        choices.length
          ? form(
              [field(t("groups.contact"), select("userId", choices))],
              t("common.invite"),
              async (v) => {
                await call("POST", path + "/invitations", { userId: v.userId });
                await render();
                word(t("groups.invited"));
              },
            )
          : empty(t("groups.noContacts")),
      ),
    );
  }

  if (manager) {
    const waiting = await call("GET", path + "/invitations");
    const lines = waiting.invitations.map((i) => {
      const head = el("span", "grow");
      head.append(
        el("span", "title", i.user.name),
        el(
          "span",
          "meta",
          t("groups.invitedBy", { name: i.inviter.name }) +
            " · " +
            when(i.createdAt),
        ),
      );
      return [
        head,
        actions(
          twoStep(t("groups.withdraw"), t("common.sure"), async () => {
            await call(
              "DELETE",
              path + "/invitations/" + encodeURIComponent(i.id),
            );
            await render();
            word(t("groups.withdrawn"));
          }),
        ),
      ];
    });
    parts.push(
      section(
        t("groups.waiting"),
        el("p", "soft", t("groups.waitingHint")),
        lines.length ? rows(lines) : empty(t("groups.waitingEmpty")),
      ),
    );
  }

  const members = group.members.map((m) => {
    const head = el("span", "grow");
    head.append(
      el(
        "span",
        "title",
        m.user.id === me.id
          ? t("you.name", { name: m.user.name })
          : m.user.name,
      ),
      el(
        "span",
        "meta",
        t("groups.joinedOn", {
          date: when(m.joinedAt, { dateStyle: "medium" }),
        }),
      ),
    );
    const line = [head];
    const self = m.user.id === me.id;
    const memberPath = path + "/members/" + encodeURIComponent(m.user.id);
    if (isOwner && !self && m.role !== "owner") {
      const role = select(
        "role",
        [
          ["admin", t("role.admin")],
          ["member", t("role.member")],
        ],
        m.role,
      );
      role.setAttribute(
        "aria-label",
        t("groups.roleOf", { name: m.user.name }),
      );
      role.addEventListener("change", () =>
        attempt(async () => {
          await call("PATCH", memberPath, { role: role.value });
          await render();
          word(t("groups.roleSet"));
        }),
      );
      line.push(role);
    } else {
      line.push(badge(t("role." + m.role)));
    }
    if (self && m.role !== "owner") {
      line.push(
        actions(
          twoStep(t("groups.leave"), t("common.sure"), async () => {
            await call("DELETE", memberPath);
            await go("/groups", true);
            word(t("groups.left"));
          }),
        ),
      );
    } else if (
      !self &&
      ((isOwner && m.role !== "owner") ||
        (group.role === "admin" && m.role === "member"))
    ) {
      line.push(
        actions(
          twoStep(t("common.remove"), t("common.sure"), async () => {
            await call("DELETE", memberPath);
            await render();
            word(t("groups.removed"));
          }),
        ),
      );
    }
    return line;
  });
  parts.push(section(t("groups.membersTitle"), rows(members)));

  if (isOwner) {
    parts.push(
      section(
        t("groups.danger"),
        twoStep(t("groups.delete"), t("groups.deleteSure"), async () => {
          await call("DELETE", path);
          await go("/groups", true);
          word(t("groups.deleted"));
        }),
      ),
    );
  }
  return parts;
}
