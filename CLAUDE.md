<!-- updated: 2026-11-16 -->
# kitsunium/todo

A task list for a team, written design first on the SDK's framework. Read
`README.md` for what it does, its API and how to run it.

One Go module, `github.com/kitsunium/todo`, go 1.27.3, on
`github.com/kitsunium/sdk` v0.57.0 and its `pkg/v1/<family>` modules,
published, no `replace`. `go list -m all` names the todo and the SDK's
modules, nothing else. The standard library is free.

`legacy/` is the former product, written on the framework by hand: a Go
module of its own (`github.com/kitsunium/todo/legacy`), with its own
`CLAUDE.md`. Nothing here imports it, and a change here never needs one there.

## Layout

| Path | Component (role) | Holds |
|---|---|---|
| `design/` | — | the design, the one source of the structure: `product.yaml` and one file per domain; `generated.lock` is written by every generation |
| `identity/account` | account (domain) | `Account` (with its `Locale` and `TimeZone`) and its lock, `Session`, `NormalizeEmail`, the errors |
| `identity/access` | access (application) | the use cases `SignUp` `VerifyEmail` `ResendVerification` `SignIn` `Authenticate` `SignOut` `ListSessions` `RevokeSession` `ForgotPassword` `ResetPassword` `ChangePassword` `UpdateProfile`, the ports `Accounts` `Sessions` `Links` `Courier` `Announcements` `Passwords` `Secrets` `Clock`; `links.go` is what the one-time links share |
| `identity/vault` | vault (adapter) | the stores `accounts`, `sessions` and `links` with their indexes, `Postman` (the mails of an account, sent through `post`), the topic `accounts` (an address proven, told by the member's id alone, never the address) and its publisher `Herald`, the queries `member` and `member-by-email` (the directory, with each member's language and time zone), `Stretcher` (PBKDF2), `Drawer` (secrets), `WallClock` |
| `identity/authapi` | authapi (presentation) | the routes of `/api/auth` (`PATCH /api/auth/me` changes the name, the language and the time zone; the sign-up negotiates the language from `Accept-Language`), the session handler `Authenticate`, the cookie `todo_session`, the setting `verify-addresses` |
| `contacts/link` | link (domain) | `Link`, `Invite` (an invitation to an address with no account), `PairOf`, the `Requests` workflow, the errors |
| `contacts/circle` | circle (application) | the use cases `RequestContact` (a request, or an invitation to an address nobody has) `AnswerRequest` `CancelInvite` `ClaimInvites` `RemoveContact` `ListContacts` `AreContacts`, the ports `Links` `Invites` `Inviter` `People` `Notices` `Clock` |
| `contacts/linkstore` | linkstore (adapter) | the stores `links` and `invites` (deleted after ninety days), its workflow `Waiting` (a request expires after fourteen days), the topic `events`, `Links`, `Invitations`, `Courier` (mails an invitation through `post`), `Directory` (asks the identity's queries), `Announcer`, `WallClock` |
| `contacts/contactsapi` | contactsapi (presentation) | the routes of `/api/contacts` (`POST /api/contacts/{id}/cancel` is a request's, `DELETE /api/contacts/invites/{id}` an invitation's), the query `are-contacts`, the subscription `accounts` (ClaimInvites on a member whose address was proven, the address read from the directory) |
| `groups/team` | team (domain) | `Group` (with its `Membership`s and `Role`), `Invitation`, the `Inviting` workflow, `Manages` `CleanName` `ColorFor` `ParseRole` `Ranked`, the errors |
| `groups/roster` | roster (application) | the use cases `CreateGroup` `EditGroup` `DeleteGroup` `ListGroups` `GetGroup` `Invite` `AnswerInvitation` `ListInvitations` `ListGroupInvitations` `WithdrawInvitation` `RemoveMember` `SetRole` `RoleOf` `Memberships` `MembersOf`, the ports `Groups` `Invitations` `People` `Circle` `Notices` `Clock`; `present.go` is what the views share |
| `groups/groupstore` | groupstore (adapter) | the stores `groups` and `invitations`, the workflow `Lapsing` (an invitation expires after a week), the topic `events`, `Groups`, `Invitations`, `Directory` (asks the identity's query), `Friends` (asks the contacts' query), `Announcer` (one event for each member told), `WallClock` |
| `groups/groupsapi` | groupsapi (presentation) | the routes of `/api/groups` (the invitations a group waits on included) and `/api/invitations`, the queries `role`, `memberships` and `members` |
| `tasks/task` | task (domain) | `Task` (with its `GroupID` and `AssigneeID`) and `Task.SeenBy`, the one rule of who sees it, `Priority`, the `Lifecycle` workflow, the errors |
| `tasks/planner` | planner (application) | the use cases `AddTask` `GetTask` `EditTask` `DeleteTask` `MoveTask` `ShareTask` `UnshareTask` `ListTasks` `CountViews` `NamePeople` `FollowGroup` `CountTasks` `DueSoon`, the ports `TaskStore` `Circle` `Groups` `People` `Clock` `Journal`; `lifecycle.go` is what they share, `access.go` is the role a member has in a task's group and who may be assigned, `views.go` the member's day and week and the views of the list |
| `tasks/taskstore` | taskstore (adapter) | the store `tasks`, its workflow `Shelf` (a done task is archived after `archive-after`), the topic `events` (a task shared or assigned names its actor and the member it concerns), `Tasks`, `Announcer` (the use cases' `Journal`, which publishes, to everyone who sees the task), `Friends` (their `Circle`, which asks the contacts' query), `Groups` (their `Groups`, which asks the groups' queries `role` and `memberships`), `Directory` (their `People`, which asks the identity's query `member` for a name), `AppClock`, and the subscription `groups`, by which the tasks follow a group deleted or a member who left it |
| `tasks/api` | api (presentation) | the routes of `/api/tasks` (views and `counts` included, `counts` carrying the `unread` of the feed; every answer that carries tasks carries their `people`), the query `census` |
| `activity/feed` | feed (adapter) | the store `entries` (each `Read` or not, marked read by the store's own `Update`), the subscriptions `tasks`, `contacts` and `groups` to their events, `GET /api/activity`, `POST /api/activity/read`, the query `unread` (which the tasks' API asks for `counts`) |
| `notify/reminders` | reminders (application) | the use cases `RemindDue` and `MailAction` (a contact request received or accepted, a group invitation, a task shared or assigned: who is mailed of what another did, once for each event), the ports `Mailer`, `Notifier`, `Directory` (who a member is written to: address, name, language, time zone), `ReminderLog` and `ActionLog` |
| `notify/post` | post (adapter) | the mailer `mail` (the framework's outbox and transport), the command `send` every mail goes through (a letter: recipient, kind, parameters, action path), the words (`locales/fr.json`, `locales/en.json`) and the design (`templates/`, HTML and text) of every mail, the setting `base-url`, `Carrier` (the reminders' `Mailer` and the `Notifier` of what others did, its SMTP connection declared in the design); ten kinds of mail in all |
| `notify/outbox` | outbox (adapter) | the stores `reminded` and `acted` (events mailed, deleted after ninety days), `Members` (the reminders' `Directory`, which asks the identity's query), `Ledger` and `Register` (the two logs) |
| `notify/watch` | watch (presentation) | the job `remind-due`, the setting `remind-within`, the subscriptions `tasks`, `contacts` and `groups` that mail what others did |
| `stats/census` | census (presentation) | the store `snapshots`, the job `sample`, `GET /api/stats` |
| `web/site` | site (presentation) | the web app, French and English: `page/index.html`, `app.css`, `lang.js` (words), `tasks.js`, `circle.js`, `app.js`, embedded and served at `/` |
| `load/driver` | driver (application) | `Drive`, the port `API` |
| `load/httpapi` | httpapi (adapter) | `Client`, the API over HTTP, the setting `target` |
| `load/loadcli` | loadcli (presentation) | the command `run` |
| `cmd/todo` | — | the binary, and its roles' wiring: `server/`, `load/` (generated), and `server/server_test.go`, the role run whole |

## Rules

- **The design is law on the structure.** A package, an exported symbol, a
  framework declaration (a store, a route, a job, a setting, a command), a
  port, a use case and who calls whom exist because `design/*.yaml` says
  so. Change the design, regenerate, then write the bodies. Never edit a
  `gen_*.go` file or `design/generated.lock`.
- **A hand-written file holds bodies.** The generator writes a skeleton
  once, then never again: the file is yours. A handler or a method the
  design gains later gets no skeleton in a file that exists — write its twin
  there, with the name its shell calls. A body that still answers
  `errNotImplemented` is a stub.
- **Nothing exported outside the design, and nothing unexported either.**
  A helper is unexported, in its component, and declared among the
  component's `internals` in the design: the dev tool derives that block
  from the code, then the check holds the code to it. A role's package
  (`cmd/todo/server`) holds no function of its own but its tests. A component imports the components its `uses` names, and
  nothing else of the module.
- **A use case and an adapter call what their `calls` draw.** A call to a
  port or a use case the design does not draw fails the check: add the link
  to the design first. A sequence's test holds the order of its steps.
- **A scenario, a sequence and a property the design names is a test to
  write**, in the file the generator opened for it. A domain method or pure
  function with a body needs a property.
- **A store holds an entity of its own component**: `taskstore.Record`
  is `task.Task` as the store keeps it, field for field. Add a field to
  both, and to `toRecord` and `fromRecord`.
- **Who sees a task is decided in one place**: `Task.SeenBy(user, inGroup)`
  (owner, shared with, assignee, or member of its group). A use case never
  tests ownership or sharing by hand: it reads the task and the caller's role
  in its group through `load` (`planner/access.go`), a list through the
  caller's `Groups.Memberships`. A task a member does not see is
  `ErrNotFound`. The groups are asked through their queries, never read.
- **Who calls is the framework's to say**: a route that needs a member
  says `auth: required` in the design, and its handler reads
  `app.UserID(ctx)`. A use case takes the owner as an input and never trusts
  an id alone: another member's task is `ErrNotFound`, never a 403.
- **A secret is never kept, answered or logged**: a password is stretched
  by `Stretcher`, a session's secret kept as its SHA-256, and both fields
  are tagged `app:"secret"`. What a stranger may ask is answered the same whether an
  address has an account or not (sign-up, a new link, a forgotten
  password); an unknown address and a wrong password are answered alike, and cost the same time (`Stretcher.verify` on an empty
  hash).
- **A record is changed where it is kept, never read then written back**:
  a use case that changes a task, a group, an invitation, a link or an
  account calls its store port's `Change(id, fn)`, and decides inside `fn`,
  on the record as it is at that instant. The adapter runs it under the
  store's `Update`, alone among the store's writers, so two changes made at
  once both stay (`cmd/todo/server` holds it: acceptances, edits, requests
  and sign-ups at the same instant). What needs another component (a role,
  a contact) is asked before, and `fn` holds it to the record it then sees.
  `fn` reads and writes nothing else. `Save` is for a record just made;
  a unique index answers two creations at once.
- **A component's data is its own**: another component reaches it through
  a query the owner declares (`vault.MemberQuery.Ask`,
  `api.CensusQuery.Ask`), a topic it publishes or a port, never its store.
- **What time does to a record is a workflow on its store** (`Shelf`,
  `Waiting`): the framework fires the timed transition, whoever wrote the
  state. What a member does stays the domain's pure workflow, in a use
  case. The two name the same states: add one to both.
- **The end-to-end tests of the server role stretch real passwords**: about
  a minute and a half under `-race`, ten seconds without. Iterate with
  `-run` and without `-race`; run the whole package under `-race` before a
  commit. The link that proves an address signs its member in, so a test
  member costs one stretching, not two.
- **A subscription is delivered at least once, and later**: it keys what
  it writes by the event's own id and writes nothing the second time
  (`feed.record`). A test that reads what a subscription wrote moves the
  clock until it is there.
- **Time is the framework's**: `app.Now(ctx)`, in UTC, to the millisecond
  (`AppClock`, and the adapters that stamp a record). A test moves it.
- **Ask "is there one?" with `Filter`, not `Get`**: a missing key is a
  failed read (`Ledger.Seen`). `Tasks.Get` reads the key because a task
  that does not exist is the caller's error, answered 404.
- **Every mail goes through `post`**, in the reader's language (French or
  English, French by default), as HTML with a plain text alternative, its
  subject one line: a component that writes to a member dispatches
  `post.SendCommand` with a kind, its parameters and the member's `Locale` and
  `TimeZone`, or goes through `Carrier`, never a transport of its own. The
  words are in `notify/post/locales`, one catalogue per language with the same
  keys (a test holds it), and the layout in `notify/post/templates`; a link is
  `base-url` and a path. A test reads what was sent with
  `post.MailMailer.Captured()`.
- **Logs name a task by its id**, never an assignee: `AssigneeID`, a mail's
  `To` are tagged `app:"personal"`, and `app:"personal,subject"` in the
  entity a store keeps. A store that keeps personal data says for how long
  in the design (`eraseAfter`, `deleteAfter`).
- **No file names the tool** that generates, checks or runs the product in
  dev: this file, the README and the code say "the generator", "the
  check", "the dev tool". The product imports the SDK only.
- **The web app is three files and no build**: no inline script or style
  (the framework's content policy refuses them), every word of the page in
  French, every call through `call` in `app.js`. A route the page uses
  that changes shape changes `app.js` in the same commit.
- The server role's App is named `server`, the load's `load`: their
  settings are `SERVER_<NAME>` and `LOAD_<NAME>`.
- **A route that writes says its limit in the design** (`rateLimit`, per
  client): 2 a second, burst 5, where a stranger makes the product work
  (sign-up, sign-in, a contact request), 10 and 20 elsewhere. The buckets
  refill on the framework's clock: a test's client moves it on a 429.
- **A setting's value for an environment is a file of the role**,
  `cmd/todo/<role>/config/<env>.json`, embedded; a key no component of the
  role declares stops the start. Dev trusts addresses there; a test of the
  server role sets `verify-addresses` back to true.
- **A request is decoded strictly**: a member the route does not declare
  is a 400. A client sends each route its own fields.

## Develop

```sh
GOWORK=off go build ./... && GOWORK=off go vet ./...
GOWORK=off go test -race ./...
gofmt -l .                                   # empty
SERVER_ENV=dev GOWORK=off go run ./cmd/todo  # 127.0.0.1:4000
```

After a change of the design: regenerate, build, write what the new shells
call, run the tests, then the structure check; commit the design, the
generated files and the lock together.

Commits: conventional, authored as `kodflow` (the `post-commit` gate checks
it), no AI attribution. Never delete `.github/workflows/post-commit.yml`.
