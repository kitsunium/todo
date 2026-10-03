# todo

The reference product of [kit](https://github.com/kitsunium/platform): a task
list for people who work together — and a product whose architecture is its
own live diagram. It is written on the SDK's framework
([`github.com/kitsunium/sdk/framework`](https://github.com/kitsunium/sdk)) and
imports the SDK alone; kit is the tool that runs it and reads its source.

Sign up, confirm your address, and your tasks have priorities (urgent, high,
medium, low), due dates, notes and views (Inbox, Today, Upcoming, Shared with
me, Assigned to me, Completed). Add people as contacts, share tasks with them
and assign them work; make groups with owners, admins and members who share a
list. Everyone concerned hears about it — in their activity feed, and by mail,
in their own language (French first, or English): a verification link, a
password reset, contact requests, invitations, shared and assigned tasks, and
a reminder fifteen minutes before a task is due.

| Today | A task |
|---|---|
| ![Today](docs/app.png) | ![A task](docs/app-task.png) |

## Eight services, one binary

| Service | Owns | Its daemon |
|---|---|---|
| **identity** | accounts and passwords (SDK PBKDF2), sessions, one-time links; the `session` auth handler every other API sits behind | the `accounts` and `tokens` workflows lock an account after five wrong passwords, unlock it fifteen minutes later, expire links; `session-reaper`, a loop written by hand (a ticker and a `select`), deletes expired sessions |
| **contacts** | who works with whom: requests, and invitations by email for people without an account | `requests` expire after 14 days; `claim-invites` turns an invitation into a request once its address is verified |
| **groups** | teams, roles, invitations | `invitations` expire after 7 days |
| **tasks** | the list: priorities, due dates, sharing, assignment, views and counts | `lifecycle` marks a task overdue the moment its due date passes and archives a done task after a day; `group-changes` follows deleted groups and departures |
| **notify** | every mail, in its recipient's language: one outbox, one light design, HTML and plain text, the words in `notify/locales/` | `task-mail`, `contact-mail`, `group-mail`; `reminders`, a loop kit runs, sleeps until the next reminder is due and wakes early when a task changes |
| **activity** | a feed per user | three subscriptions write it from the events of tasks, contacts and groups |
| **stats** | the list's vital signs | a job samples the census every 30 seconds |
| **web** | the single page application (`web/dist`, embedded) | — |

Services talk through commands, queries and topics, never through each
other's stores — the diagram shows every call. A service asks another's
queries and dispatches its commands in process; nobody exposes them, so no
route leads to them. Identity's mails go through notify's `send` command,
synchronously, so a one-time link never travels on a topic.

## The product, as its diagram

In dev, the kit Studio draws the product live, from what the product serves
about itself at `/_kit/` (its dev build links that API, `kitdev.go`): its
context, its containers, its services and their building blocks, the code of
each, what every entry point does step by step, every request as a trace, the
workflows and their populations, the daemon — its lifecycle, its loops, its
HTTP server — and the mails it sent.

| **Overview** — the process's own numbers, then its connectors' widgets | **Context** — who uses the product, what it talks to (C4 level 1) |
|---|---|
| ![Overview](docs/studio-overview.png) | ![Context](docs/studio-context.png) |
| **Infrastructure** — its services, their ports and flows (C4 level 2) | **Domains** — every bounded context as a peer, the flows between them (C4 level 3) |
| ![Infrastructure](docs/studio-infrastructure.png) | ![Domains](docs/studio-domain.png) |
| **A domain** — a request replayed through its components and its sequence | **Code** — what a node runs, what it touches (C4 level 4) |
| ![A domain](docs/studio-domain-replay.png) | ![Code](docs/studio-code.png) |
| **Sequence** — what an entry point does, step by step, from its code | **Requests** — every request as a waterfall, a sequence, its payloads |
| ![Sequence](docs/studio-sequence.png) | ![Requests](docs/studio-requests.png) |
| **States** — a workflow, its population and its instances | **The daemon** — its phases and its boot sequence |
| ![States](docs/studio-states.png) | ![Daemon](docs/studio-daemon.png) |
| **Runtime** — every loop, by who wrote it: a library, kit, or the product | **The HTTP server** — net/http's accept loop and the middleware |
| ![Runtime](docs/studio-runtime.png) | ![HTTP server](docs/studio-http.png) |
| **Mail** — every mail the product sent, in dev | |
| ![Mail](docs/studio-mail.png) | |

Generated: the product's own `graph -format mermaid`, given kit's static
analysis of its source, and kept current by `TestTheReadmeDiagramIsCurrent`.
Solid arrows exist by construction; dashed ones were found in the code by
kit's static analysis, which the product does not link — `go run . graph
-format mermaid` draws the solid ones alone. In the Studio, arrows also
carry what the running product observed.

```mermaid
flowchart LR
  n0(("Clients"))
  subgraph n1["activity"]
    n2["GET /api/activity · auth"]
    n3["POST /api/activity/read · auth"]
    n4{"unread"}
    n5[("entries")]
    n6[/"contacts"/]
    n7[/"groups"/]
    n8[/"tasks"/]
  end
  subgraph n9["contacts"]
    n10["POST /api/contacts/{id}/accept · auth"]
    n11["POST /api/contacts · auth"]
    n12["POST /api/contacts/{id}/cancel · auth"]
    n13["POST /api/contacts/{id}/decline · auth"]
    n14["GET /api/contacts · auth"]
    n15["DELETE /api/contacts/{id} · auth"]
    n16{"are-contacts"}
    n17[("invites")]
    n18[("links")]
    n19[/"claim-invites"/]
    n20>"events"]
    n21{{"requests"}}
  end
  subgraph n22["groups"]
    n23["POST /api/invitations/{id}/accept · auth"]
    n24["POST /api/groups · auth"]
    n25["POST /api/invitations/{id}/decline · auth"]
    n26["DELETE /api/groups/{id} · auth"]
    n27["GET /api/groups/{id} · auth"]
    n28["POST /api/groups/{id}/invitations · auth"]
    n29["GET /api/groups · auth"]
    n30["GET /api/invitations · auth"]
    n31["DELETE /api/groups/{id}/members/{userId} · auth"]
    n32["PATCH /api/groups/{id}/members/{userId} · auth"]
    n33["PATCH /api/groups/{id} · auth"]
    n34{"members"}
    n35{"memberships"}
    n36{"refs"}
    n37{"role"}
    n38[("groups")]
    n39[("invitations")]
    n40>"events"]
    n41{{"invitations"}}
  end
  subgraph n42["identity"]
    n43[/"session"\]
    n44["POST /api/auth/password · auth"]
    n45["POST /api/auth/password/forgot"]
    n46["GET /api/auth/sessions · auth"]
    n47["POST /api/auth/login"]
    n48["POST /api/auth/logout · auth"]
    n49["GET /api/auth/me · auth"]
    n50["POST /api/auth/verify/resend"]
    n51["POST /api/auth/password/reset"]
    n52["DELETE /api/auth/sessions/{id} · auth"]
    n53["POST /api/auth/signup"]
    n54["PATCH /api/auth/me · auth"]
    n55["POST /api/auth/verify"]
    n56((("session-reaper")))
    n57{"email-owner"}
    n58{"users"}
    n59[("accounts")]
    n60[("sessions")]
    n61[("tokens")]
    n62>"accounts"]
    n63{{"accounts"}}
    n64{{"tokens"}}
  end
  subgraph n65["notify"]
    n66[\"send"\]
    n67(("reminders"))
    n68[["mail"]]
    n69[("reminders")]
    n70[/"contact-mail"/]
    n71[/"group-mail"/]
    n72[/"task-mail"/]
    n73[/"track-due"/]
  end
  subgraph n74["stats"]
    n75["GET /api/stats"]
    n76(["sample"])
    n77[("snapshots")]
  end
  subgraph n78["tasks"]
    n79["POST /api/tasks/{id}/archive · auth"]
    n80["POST /api/tasks/{id}/complete · auth"]
    n81["GET /api/tasks/counts · auth"]
    n82["POST /api/tasks · auth"]
    n83["DELETE /api/tasks/{id} · auth"]
    n84["GET /api/tasks/{id} · auth"]
    n85["GET /api/tasks · auth"]
    n86["POST /api/tasks/{id}/reopen · auth"]
    n87["POST /api/tasks/{id}/restore · auth"]
    n88["POST /api/tasks/{id}/share · auth"]
    n89["DELETE /api/tasks/{id}/share/{userId} · auth"]
    n90["PATCH /api/tasks/{id} · auth"]
    n91{"audience"}
    n92{"census"}
    n93{"open-by-group"}
    n94[("tasks")]
    n95[/"group-changes"/]
    n96>"events"]
    n97{{"lifecycle"}}
  end
  subgraph n98["web"]
    n99["app"]
  end
  n2 -.->|reads| n5
  n3 -.->|reads| n5
  n3 -.->|writes| n5
  n4 -.->|reads| n5
  n6 -.->|asks| n58
  n6 -.->|reads| n5
  n6 -.->|writes| n5
  n7 -.->|asks| n58
  n7 -.->|reads| n5
  n7 -.->|writes| n5
  n8 -.->|asks| n36
  n8 -.->|asks| n58
  n8 -.->|reads| n5
  n8 -.->|writes| n5
  n10 -.->|asks| n58
  n10 -.->|reads| n18
  n10 -.->|transitions accept| n21
  n11 -.->|asks| n57
  n11 -.->|publishes| n20
  n11 -.->|reads| n17
  n11 -.->|reads| n18
  n11 -.->|transitions create| n21
  n11 -.->|writes| n17
  n11 -.->|writes| n18
  n12 -.->|reads| n17
  n12 -.->|reads| n18
  n12 -.->|transitions cancel| n21
  n12 -.->|writes| n17
  n13 -.->|reads| n18
  n13 -.->|transitions decline| n21
  n14 -.->|asks| n58
  n14 -.->|reads| n17
  n14 -.->|reads| n18
  n15 -.->|reads| n17
  n15 -.->|reads| n18
  n15 -.->|writes| n17
  n15 -.->|writes| n18
  n16 -.->|reads| n18
  n19 -.->|reads| n17
  n19 -.->|reads| n18
  n19 -.->|transitions create| n21
  n19 -.->|writes| n17
  n19 -.->|writes| n18
  n20 -->|delivers| n6
  n20 -->|delivers| n70
  n21 -->|persists| n18
  n21 -.->|publishes| n20
  n0 -->|calls| n99
  n23 -.->|asks| n58
  n23 -.->|asks| n93
  n23 -.->|reads| n38
  n23 -.->|reads| n39
  n23 -.->|transitions accept| n41
  n24 -.->|asks| n58
  n24 -.->|asks| n93
  n24 -.->|writes| n38
  n25 -.->|reads| n39
  n25 -.->|transitions decline| n41
  n26 -.->|publishes| n40
  n26 -.->|reads| n38
  n26 -.->|reads| n39
  n26 -.->|transitions revoke| n41
  n26 -.->|writes| n38
  n27 -.->|asks| n58
  n27 -.->|asks| n93
  n27 -.->|reads| n38
  n28 -.->|asks| n16
  n28 -.->|asks| n58
  n28 -.->|reads| n38
  n28 -.->|reads| n39
  n28 -.->|transitions create| n41
  n30 -.->|asks| n58
  n30 -.->|reads| n38
  n30 -.->|reads| n39
  n29 -.->|asks| n58
  n29 -.->|asks| n93
  n29 -.->|reads| n38
  n31 -.->|publishes| n40
  n31 -.->|reads| n38
  n31 -.->|writes| n38
  n32 -.->|asks| n58
  n32 -.->|asks| n93
  n32 -.->|reads| n38
  n32 -.->|writes| n38
  n33 -.->|asks| n58
  n33 -.->|asks| n93
  n33 -.->|reads| n38
  n33 -.->|writes| n38
  n35 -.->|reads| n38
  n34 -.->|reads| n38
  n36 -.->|reads| n38
  n37 -.->|reads| n38
  n40 -->|delivers| n7
  n40 -->|delivers| n71
  n40 -->|delivers| n95
  n41 -->|persists| n39
  n41 -.->|publishes| n40
  n41 -.->|reads| n38
  n41 -.->|writes| n38
  n43 -.->|reads| n60
  n43 -.->|writes| n60
  n44 -.->|publishes| n62
  n44 -.->|reads| n59
  n44 -.->|reads| n60
  n44 -.->|writes| n59
  n44 -.->|writes| n60
  n45 -.->|dispatches| n66
  n45 -.->|reads| n59
  n45 -.->|transitions create| n64
  n46 -.->|reads| n60
  n47 -.->|reads| n59
  n47 -.->|writes| n59
  n47 -.->|writes| n60
  n48 -.->|writes| n60
  n49 -.->|reads| n59
  n50 -.->|dispatches| n66
  n50 -.->|reads| n59
  n50 -.->|transitions create| n64
  n51 -.->|publishes| n62
  n51 -.->|reads| n60
  n51 -.->|reads| n61
  n51 -.->|transitions reset| n63
  n51 -.->|transitions use| n64
  n51 -.->|writes| n59
  n51 -.->|writes| n60
  n52 -.->|reads| n60
  n52 -.->|writes| n60
  n53 -.->|dispatches| n66
  n53 -.->|reads| n59
  n53 -.->|transitions create| n63
  n53 -.->|transitions create| n64
  n54 -.->|reads| n59
  n54 -.->|writes| n59
  n55 -.->|reads| n61
  n55 -.->|transitions verify| n63
  n55 -.->|transitions use| n64
  n55 -.->|writes| n60
  n56 -.->|reads| n60
  n56 -.->|reads| n61
  n56 -.->|writes| n60
  n56 -.->|writes| n61
  n57 -.->|reads| n59
  n58 -.->|reads| n59
  n62 -->|delivers| n19
  n63 -->|persists| n59
  n63 -.->|publishes| n62
  n64 -->|persists| n61
  n66 -.->|sends| n68
  n67 -.->|asks| n58
  n67 -.->|asks| n91
  n67 -.->|reads| n69
  n67 -.->|sends| n68
  n67 -.->|writes| n69
  n70 -.->|asks| n58
  n70 -.->|sends| n68
  n71 -.->|asks| n58
  n71 -.->|sends| n68
  n72 -.->|asks| n58
  n72 -.->|sends| n68
  n73 -.->|reads| n69
  n73 -.->|writes| n69
  n75 -.->|reads| n77
  n76 -.->|asks| n92
  n76 -.->|reads| n77
  n76 -.->|writes| n77
  n79 -.->|asks| n36
  n79 -.->|asks| n37
  n79 -.->|asks| n58
  n79 -.->|reads| n94
  n79 -.->|transitions archive| n97
  n80 -.->|asks| n36
  n80 -.->|asks| n37
  n80 -.->|asks| n58
  n80 -.->|reads| n94
  n80 -.->|transitions complete| n97
  n81 -.->|asks| n4
  n81 -.->|asks| n35
  n81 -.->|reads| n94
  n82 -.->|asks| n34
  n82 -.->|asks| n36
  n82 -.->|asks| n37
  n82 -.->|asks| n58
  n82 -.->|publishes| n96
  n82 -.->|transitions create| n97
  n83 -.->|asks| n34
  n83 -.->|asks| n37
  n83 -.->|publishes| n96
  n83 -.->|reads| n94
  n83 -.->|writes| n94
  n84 -.->|asks| n36
  n84 -.->|asks| n37
  n84 -.->|asks| n58
  n84 -.->|reads| n94
  n85 -.->|asks| n35
  n85 -.->|asks| n36
  n85 -.->|asks| n58
  n85 -.->|reads| n94
  n86 -.->|asks| n36
  n86 -.->|asks| n37
  n86 -.->|asks| n58
  n86 -.->|reads| n94
  n86 -.->|transitions reopen| n97
  n87 -.->|asks| n36
  n87 -.->|asks| n37
  n87 -.->|asks| n58
  n87 -.->|reads| n94
  n87 -.->|transitions restore| n97
  n88 -.->|asks| n16
  n88 -.->|asks| n34
  n88 -.->|asks| n36
  n88 -.->|asks| n37
  n88 -.->|asks| n58
  n88 -.->|publishes| n96
  n88 -.->|reads| n94
  n88 -.->|writes| n94
  n89 -.->|asks| n34
  n89 -.->|asks| n36
  n89 -.->|asks| n37
  n89 -.->|asks| n58
  n89 -.->|publishes| n96
  n89 -.->|reads| n94
  n89 -.->|writes| n94
  n90 -.->|asks| n34
  n90 -.->|asks| n36
  n90 -.->|asks| n37
  n90 -.->|asks| n58
  n90 -.->|publishes| n96
  n90 -.->|reads| n94
  n90 -.->|transitions reschedule| n97
  n90 -.->|writes| n94
  n91 -.->|reads| n94
  n92 -.->|reads| n94
  n93 -.->|reads| n94
  n95 -.->|reads| n94
  n95 -.->|writes| n94
  n96 -->|delivers| n8
  n96 -->|delivers| n72
  n96 -->|delivers| n73
  n96 -->|wakes| n67
  n97 -.->|asks| n34
  n97 -->|persists| n94
  n97 -.->|publishes| n96
```

## Run it

The product builds with the Go toolchain alone, on the SDK's published
modules (`go build .`). kit, the platform's tool, runs it in dev; install it
from a platform checkout:

```sh
git clone https://github.com/kitsunium/platform ../platform
(cd ../platform && go install ./cmd/kit)

kit dev    # http://localhost:4000, the app; kit prints the Studio's link
```

`kit dev` rebuilds and restarts the product on every save; data lives in
`.kit/data`. Without an SMTP relay, mails are captured: open them in the
Studio's **Mail** view — the verification link of a new account is there.

## A task's life

```mermaid
stateDiagram-v2
  [*] --> open: create
  open --> done: complete
  overdue --> done: complete
  done --> open: reopen
  done --> archived: archive
  archived --> open: restore
  overdue --> open: reschedule (a later due date)
  open --> overdue: overdue (timer: at the due date)
  done --> archived: auto-archive (timer: the setting archive-after, 24h)
```

## An account's life

```mermaid
stateDiagram-v2
  [*] --> unverified: sign up
  unverified --> active: verify (the link of the first mail)
  unverified --> active: reset (a reset link proves the address too)
  active --> locked: lock (guard: 5 wrong passwords in a row)
  locked --> active: unlock (timer: 15m)
  locked --> active: reset
```

## Languages

The product speaks **French first**, and English. Every account has a
language, `fr` or `en`, and every mail it receives is written in it —
subject, text, button, footer, dates ("jeudi 24 septembre à 17:00 UTC",
"Thursday, September 24 at 17:00 UTC") and plurals.

| When | The language is |
|---|---|
| signing up | the form's `locale`; without one, the browser's `Accept-Language` negotiated over `fr`, `en` (`en-US,en;q=0.9` → `en`); when it names neither, or is absent: **French** |
| later | whatever `PATCH /api/auth/me {"locale": "en"}` sets — the next mails follow |
| an account from before languages | French |
| mailing someone without an account (an invitation to join) | the inviter's |

The words live in `notify/locales/fr.json` and `en.json`, one flat catalogue
per language rendered by the SDK's `i18n` (placeholders like `{actor}`, CLDR
plural forms for counts). The process refuses to start when a key is missing
from one language or a count lacks a plural form its language needs.

The activity feed is written by the web app in its reader's language, from
each entry's `kind`, `actor`, `target` (the other person it is about: the
sharee, the assignee, the invitee, the removed member — absent when the
actor acted on themselves), `task` and `group`. Its `text` is an English
sentence kept for the API's sake.

## API

Every route but sign-up, sign-in, recovery and `/api/stats` needs a session:
the `todo_session` cookie (HttpOnly, SameSite=Lax, Secure outside dev, 30
days, sliding), or the same secret as `Authorization: Bearer`. Errors are
`{"error":{"code","message","violations?"}}`. The user the account routes
answer carries its `locale`; sign-up takes an optional `locale` and reads
`Accept-Language`; `PATCH /api/auth/me` takes `{name?, locale?}` — a
language other than `fr` or `en` is a violation on `locale`.

| Area | Routes |
|---|---|
| account | `POST /api/auth/signup` `verify` `verify/resend` `login` `logout` `password` `password/forgot` `password/reset` · `GET PATCH /api/auth/me` · `GET /api/auth/sessions` · `DELETE /api/auth/sessions/{id}` |
| contacts | `GET POST /api/contacts` · `POST /api/contacts/{id}/accept` `decline` `cancel` · `DELETE /api/contacts/{id}` |
| groups | `GET POST /api/groups` · `GET PATCH DELETE /api/groups/{id}` · `POST /api/groups/{id}/invitations` · `GET /api/invitations` · `POST /api/invitations/{id}/accept` `decline` · `PATCH DELETE /api/groups/{id}/members/{userId}` |
| tasks | `GET /api/tasks?view=&group=&tz=` · `GET /api/tasks/counts?tz=` · `POST /api/tasks` · `GET PATCH DELETE /api/tasks/{id}` · `POST /api/tasks/{id}/complete` `reopen` `archive` `restore` · `POST /api/tasks/{id}/share` · `DELETE /api/tasks/{id}/share/{userId}` |
| activity | `GET /api/activity` · `POST /api/activity/read` |
| stats | `GET /api/stats` |

The Studio's API view lists every route with its request and response
schema, and sends requests.

## Deploy

One static binary, anywhere:

```sh
KIT_DATA_DIR=/var/lib/todo \
TODO_BASE_URL=https://todo.example.com \
KIT_SMTP_URL='smtp://user:password@smtp.example.com:587?tls=starttls' \
./todo                                    # production: no Studio, data on disk
./todo healthcheck                        # exit 0 when ready: for any supervisor
./todo config                             # every setting, and where it came from
./todo graph                              # the product graph, as JSON
```

The todo declares two settings: `base-url` (service notify), where users
reach the app, for the links in its mails (default `http://localhost:4000`),
and `archive-after` (service tasks), how long a done task stays on the lists
(default `24h`). Each is set by its variable — `TODO_BASE_URL`,
`TODO_ARCHIVE_AFTER` — or, for every deployment of an environment, in the
product's configuration file `config/<env>.yaml`, embedded in the binary;
`./todo config` prints every setting and where its value came from.
`KIT_SMTP_URL` is the relay — without it, mails are only captured and a
warning says so. With Docker:

```sh
docker build -t todo .
docker run -p 4000:4000 -v todo-data:/data -e TODO_BASE_URL=… -e KIT_SMTP_URL=… todo
```

The image is distroless, runs as non-root, keeps its data in the `/data`
volume, and checks its own health with `todo healthcheck`.

## Test

```sh
go test -race ./...
```

Every test runs the whole product in-process, on a manual clock it moves:
accounts from sign-up to sign-out and the language each reads, the lock and
its timer, resets, contacts and invitations by email, groups and roles, the
task list with its views, counts and timers, sharing, and the reminders
loop. Every mail is rendered in both languages, and no word of the English
catalogue may appear in a French mail.
`TestTheDiagramMatchesTheCode` uses the product end to end with its static
analysis on, then checks that the process runs exactly the nodes the source
declares, that every edge the code proves is drawn, and that every edge the
running product took is explained by the code. `TestTheReadmeDiagramIsCurrent`
fails when the diagram above no longer matches the code: replace it with the
diagram the test prints.

The static analysis is kit's: the tests run `kit graph -static` and give its
graph to the product through the framework's `kit.Analyzer`, for the product
links no analyzer and imports nothing from the platform. With kit on `PATH`,
or `KIT` set to its binary, both tests run whole:

```sh
KIT=$(command -v kit) go test -race -v -run 'TestTheDiagramMatchesTheCode|TestTheReadmeDiagramIsCurrent' .
```

Without kit, each checks what the product says of itself — the running
product's nodes, loops, authentication, mailer and diagnostics; the
diagram's solid arrows — and skips the rest, naming it (`go test -v` shows
it).
