# todo

A task list for a team, served over HTTP by one process — and a product
written design first: `design/` says its structure, a generator writes the
shells from it, and the bodies beside them are written by hand. It runs on
the SDK's framework
([`github.com/kitsunium/sdk/framework/app`](https://github.com/kitsunium/sdk))
and imports the SDK alone.

A member signs up with an address and a password, proves the address by
the link they are mailed, signs in, and has a task
list of their own. Every mail the product sends is written in the language of
the member who reads it, French or English, as HTML with a plain text
alternative, and a due date in it is written in that member's time zone. Ten kinds of mail go out: the five the
account and the reminders need (the link that proves an address, the one that chooses a new
password, an address that already has an account, an invitation to sign up, a
reminder of a task due soon) and the five that tell a member what another did,
a contact request received or accepted, a group invitation, a task shared with
them or assigned to them, each in the language and the time zone of the member
it is written to and naming who did it; a task a member assigns to themselves
is mailed to nobody, and an event delivered twice is mailed once. One limit is known and kept: a mail leaves before its event is marked as mailed, so a process that dies between the two sends that mail a second time when the event is delivered again, as it does for a reminder; the other order would lose a mail instead. They add tasks with a priority, an assignee and a due date, to their own list or to a group's, and move
them through their lifecycle: open, done, archived. Every move is published on a topic, and lands in their activity feed a
moment later. A task done for a day is archived on its own, the owner of a task
due within a day gets one reminder by mail, and the task list
is counted every thirty seconds. Two members become contacts by a
request one sends to the other's address and the other accepts; a request
nobody answers expires after fourteen days. An address nobody has is
invited by mail to sign up instead: the invitation waits, ninety days at most,
and when the address is proven, whoever signs up with it finds a request from
each member who invited it. Members form groups: a group has an owner, admins and members, its owner
and admins invite their contacts and may withdraw an invitation still waiting, and an invitation nobody answers expires
after seven days. A member shares a task with a
contact, who then sees it, edits it and moves it, and may leave it. A task added to a group's
list is seen, edited and moved by every member of the group; a task is assigned to a member, who then sees
it. When a member leaves a group or is removed from it, the tasks they were assigned in it are theirs no more, and
the tasks of a group deleted go back to their owners' lists. The server serves a web app at `/`,
in French or English, that does everything the API does: the account and its settings and sessions,
contacts and their requests, groups with their members and invitations, the tasks with their views, counts,
sharing and assignment, the feed and the vital signs.

The former todo, written on the framework by hand, is kept whole in
[`legacy/`](legacy/), a Go module of its own, as a reference: its richer web
app no longer does anything this one does not.

## The design

| File | Domain | Holds |
|---|---|---|
| `design/product.yaml` | — | the product, its context (a member, an operator, the mail relay) and the binary `todo` with its two process roles, `server` and `load` |
| `design/identity.yaml` | identity | the account and its lock, the session, four use cases and their five ports, the stores and the topic `accounts` behind them (which tells the id of the member whose address was proven, never the address), the HTTP API and the handler that says who calls a route, the sequence `sign-in` |
| `design/contacts.yaml` | contacts | the link between two members and its request, seven use cases and their six ports, the stores (links and invitations) and the topic behind them, the subscription to the accounts' announcements, the HTTP API, the sequence `request-contact` |
| `design/groups.yaml` | groups | the group, its roles and its invitations, the use cases and their ports (the invitations a group waits on are listed and withdrawn by its owner and admins), the stores and the topic behind them, the HTTP API, the queries the rest of the product asks |
| `design/tasks.yaml` | tasks | the task and its lifecycle, its use cases and their ports (the directory among them, asked for a name alone), the store behind them, the HTTP API, the sequence `complete-task` |
| `design/activity.yaml` | activity | the feed, written from what the use cases record, what was read of it, and the query that counts what was not |
| `design/notify.yaml` | notify | the reminders and the mails about what others did (two use cases, their ports and the log that keeps each event mailed once), the mailer with its outbox and its SMTP connection, the command every mail goes through, the subscriptions to the tasks', contacts' and groups' topics, the sequence `remind-due` |
| `design/stats.yaml` | stats | the vital signs, sampled by a job |
| `design/web.yaml` | web | the web app's static files |
| `design/load.yaml` | load | a load driven against a running server over its HTTP API |

Each component of the design is a Go package, `<domain>/<component>`, whose
`gen_*.go` files are generated and never edited; `cmd/todo` is the binary.

## The API

| Route | Does |
|---|---|
| `POST /api/auth/signup` | creates an account: `email`, `name`, `password` (10 to 128 characters), and optionally `locale` (`fr` or `en`; when it is not sent, the language the `Accept-Language` header prefers, French by default) and `timeZone` (an IANA name like `Europe/Paris`; empty is UTC), and mails the link that proves the address, in that language; answers the same for an address that already has an account, whose member is told in their own language; 400 for a `locale` that is neither `fr` nor `en` or a `timeZone` that is not a zone of the database |
| `POST /api/auth/verify` | proves an address with the `token` of the link, and signs its member in |
| `POST /api/auth/verify/resend` | mails a new link to an `email` not verified yet |
| `POST /api/auth/password/forgot` | mails an `email` the link that chooses a new password |
| `POST /api/auth/password/reset` | chooses a new `password` with the `token` of that link, and ends every session |
| `POST /api/auth/password` | changes the caller's password: `current`, `password`; ends their other sessions |
| `POST /api/auth/login` | opens a session: `email`, `password`, and optionally `timeZone`, which becomes the account's; sets the `todo_session` cookie and answers the same secret as `token`; 400 for a `timeZone` that is not a zone of the database |
| `POST /api/auth/logout` | ends the caller's session |
| `GET /api/auth/me` | the caller: `id`, `email`, `name`, `locale`, `timeZone` |
| `PATCH /api/auth/me` | changes `name`, `locale` and `timeZone`, each only when it is sent (an empty `timeZone` is UTC), and answers the caller; the next mail is in the new language; 400 for an empty `name`, a `locale` that is neither `fr` nor `en` or a `timeZone` that is not a zone of the database, and for any other member of the request |
| `GET /api/auth/sessions` | the caller's open sessions, with the browser and the address each signed in from |
| `DELETE /api/auth/sessions/{id}` | ends one of the caller's sessions |
| `GET /api/contacts` | the caller's contacts, the requests they received and those they sent, and `invites`, their invitations still waiting for an account, the newest first |
| `POST /api/contacts` | asks the member an `email` belongs to to be a contact and answers `{"kind":"request","request":{…}}`; an address nobody has is invited to sign up by mail and answers `{"kind":"invite","invite":{"id","email","createdAt"}}`; 400 for an address that is none, 409 for a pair already linked or an address already invited |
| `POST /api/contacts/{id}/accept` `decline` | answers a request: its addressee accepts or declines |
| `POST /api/contacts/{id}/cancel` | its requester cancels a request and the answer is the request; the id is a request's only: an invitation's id, anyone else's request and an unknown id are a 404, a request already answered is a 409 |
| `DELETE /api/contacts/invites/{id}` | a member withdraws an invitation of their own that still waits for an account and the answer is the invitation `{"id","email","createdAt"}`; 404 for anyone else's, for one already joined and for an unknown id |
| `DELETE /api/contacts/{id}` | ends a contact, for either member |
| `GET /api/groups` `POST /api/groups` | the caller's groups; creates one: `name`, `color` |
| `GET` `PATCH` `DELETE /api/groups/{id}` | reads a group, renames or recolours it (owner and admins), deletes it (owner) |
| `POST /api/groups/{id}/invitations` | invites a contact: `userId`, for the owner and the admins |
| `GET /api/groups/{id}/invitations` | the invitations the group is waiting on, newest first, `{"invitations":[{"id","user":{"id","name"},"inviter":{"id","name"},"createdAt"}]}`, never an address; for the owner and the admins as the invite route (403 for a plain member, 404 for who is not in the group, as `GET /api/groups/{id}`) |
| `DELETE /api/groups/{id}/invitations/{invitationId}` | withdraws an invitation the group is waiting on, for the owner and the admins (403 for a plain member, 404 for who is not in the group); 404 for an invitation of another group, one already answered or one nobody made. The invitation moves once, under the store's `Change`: a withdrawal and an acceptance at the same instant leave one winner, the withdrawal is answered 200 and the acceptance 409 (no longer waiting), or the acceptance 200 and the withdrawal 404 |
| `GET /api/invitations`, `POST /api/invitations/{id}/accept` `decline` | the invitations the caller received, and their answer |
| `PATCH` `DELETE /api/groups/{id}/members/{userId}` | sets a member's role (owner), removes a member or leaves the group |
| `POST /api/tasks` | adds a task: `title`, and `notes`, `priority` (`low`, `normal`, `high`), `groupId` (a group the caller is in), `assigneeId` (its owner, or a member of the group for a group's task), `due` (RFC 3339) |
| `GET /api/tasks` | lists the tasks the caller sees, theirs, shared with them, assigned to them and on their groups' lists, not archived, the most urgent first, and answers `{"tasks":[…],"people":[{"id","name"}]}`, the `people` being every owner, assignee and member a task is shared with, each once, by name and never by address (a member the directory does not know is left out); `?state=open\|done\|archived` lists one state; `?view=inbox\|today\|upcoming\|shared\|assigned\|completed\|group` lists a view (a view ignores `state`), `?group=<id>` narrows to a group (the `group` view needs it, a group the caller is not in is a 404), `?tz=<IANA zone>` says when the day ends (UTC by default) |
| `GET /api/tasks/counts` | what the sidebar counts for the caller: `inbox`, `today`, `upcoming`, `shared`, `assigned`, `unread` (the entries of their feed not read yet), `overdue`, `completedThisWeek` and the open tasks of each of their `groups`; `?tz=<IANA zone>` |
| `GET /api/tasks/{id}` | one task the caller sees, with `people`: `[{"id","name"}]`, its owner, assignee and the members it is shared with, each once (every route that answers one task, adding, editing, sharing, moving, answers it the same) |
| `PATCH /api/tasks/{id}` | changes a task: any of `title`, `notes`, `priority`, `assigneeId` (its owner, a member it is shared with or a member of its group; empty clears it), `groupId` (its owner moves it to another of their groups, or to none), `due`, each only when sent |
| `DELETE /api/tasks/{id}` | deletes a task, for its owner, or for an owner or an admin of its group |
| `POST /api/tasks/{id}/share` | shares a task with a contact: `userId`, for its owner |
| `DELETE /api/tasks/{id}/share/{userId}` | ends the sharing with a member, for the owner or that member |
| `POST /api/tasks/{id}/complete` `reopen` `archive` `restore` | moves a task through its lifecycle |
| `GET /api/activity` | the caller's feed, newest first; `?limit=N`; every entry says whether it was `read` |
| `POST /api/activity/read` | marks every entry of the caller's feed read (no body) and answers `{"unread":0}`; what is recorded after is unread |
| `GET /api/stats` | the latest sample of the task list and the recent history |

Every route but those a stranger needs — signing up, the links, signing
in — and `/api/stats` asks for a member: the
session cookie, or `Authorization: Bearer <token>`; without one it answers
401. A member sees a task when they own it, it is shared with them, it is assigned to
them, or it is on the list of a group they are in, and they see their feed;
a task they do not see is answered as if it did not exist, and what only
its owner does (sharing it) is a 403 for the others, as deleting it is for a member of its group
who is neither its owner nor an owner or an admin of the group.

Every route that writes is limited per client — the member, or the address
of a stranger: two calls a second with a burst of five for signing up,
signing in and asking for a contact, ten a second with a burst of twenty
for the rest. A call over the limit is answered 429.

A request the API refuses says why: 400 with the field at fault, 404 for a
task or a group that does not exist, 409 for a move the task's state does not allow,
400 `invalid_token` for a link that is unknown, used or expired, 401
`invalid_credentials` for a wrong password or an unknown address alike, 403
`email_unverified` until the address is proven, 403 `account_locked` for
fifteen minutes after five wrong passwords in a row. What a stranger may
ask — signing up, a new link, a forgotten password — is answered the same
whether an address has an account or not.

## Run it

```sh
SERVER_ENV=dev go run ./cmd/todo            # http://127.0.0.1:4000
SERVER_ADDR=127.0.0.1:8080 go run ./cmd/todo
```

The server listens where `SERVER_ADDR` says. In dev it defaults to
`127.0.0.1:4000`; in production the framework's default is `:4000`, every
interface, so a deployment names its address. Its data lives in
`SERVER_DATA_DIR`.

| Setting | Default | Says |
|---|---|---|
| `SERVER_ARCHIVE_AFTER` | `24h` | how long a done task stays listed before it is archived |
| `SERVER_REMIND_WITHIN` | `24h` | how long before its due date a task's owner is reminded |
| `SERVER_BASE_URL` | `http://localhost:4000` | where members reach the product: every link of a mail is this address and a path of the web app |
| `SERVER_VERIFY_ADDRESSES` | `true`, `false` in dev | whether a new account proves its address before it signs in; `false` where the deployment has no mail relay |
| `SERVER_SMTP_URL` | unset | where the mailer delivers: `smtp://user:pass@host:587?tls=starttls`; unset, the mails are captured for the dev tool and not delivered |

What the stores keep of a person has an end: who a task is assigned to is erased a
year after the task was last written, a feed entry is deleted after ninety
days, an account untouched for three years is erased,
and a session is deleted a day after it expired. A password is kept
stretched (PBKDF2-SHA256), a session's secret as its SHA-256.

A role's configuration files are embedded in the binary:
`cmd/todo/<role>/config/config.json` for every environment,
`<env>.json` for one. The environment's variables win over them.

`go run ./cmd/todo config` prints every setting and where its value comes
from.

## The web app

`web/site/page` holds the files of the web app, written by hand — `index.html`,
`app.css` and a few plain scripts (`lang.js`, the words of each language;
`tasks.js`; `circle.js`, contacts and groups; `app.js`, the router and the
account) — with no build step and no dependency; the binary embeds them. The
page speaks French or English, by the account's language, and writes dates in
its time zone. It talks to the API above with the session cookie. The server
answers the page for any path that names no file, and its content policy allows
scripts and styles of its own origin only: nothing inline.

## Drive a load

The binary's second role calls a running server's API — it signs up and in
as a member of its own, `load@example.com`, adds tasks,
completes, archives and reopens them, and reads the list, the feed and the
vital signs — so that there is something to watch:

```sh
LOAD_TARGET=127.0.0.1:4000 go run ./cmd/todo load -for 2m -rate 5
```

The load reads no mail: the server it is driven against trusts addresses,
as one run in dev does (`cmd/todo/server/config/dev.json`), or its member's
address was proven once by hand.

`-for` is how long it lasts (default 30s), `-rate` how many calls a second
(default 5). It prints how many calls it made and how many failed, and exits
1 when some did.

## Test it

```sh
go build ./... && go vet ./...
go test -race ./...
gofmt -l .
```

The use cases are tested by the scenarios the design names, on ports in
memory; the three sequences by the order of their calls; the task's and the
account's rules by properties. `cmd/todo/server` runs the whole server role
in memory, over HTTP, on a clock the test moves, and the load role against
it.
