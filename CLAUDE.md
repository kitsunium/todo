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
| `identity/account` | account (domain) | `Account` and its lock, `Session`, `NormalizeEmail`, the errors |
| `identity/access` | access (application) | the use cases `SignUp` `VerifyEmail` `ResendVerification` `SignIn` `Authenticate` `SignOut` `ForgotPassword` `ResetPassword` `ChangePassword`, the ports `Accounts` `Sessions` `Links` `Courier` `Passwords` `Secrets` `Clock`; `links.go` is what the one-time links share |
| `identity/vault` | vault (adapter) | the stores `accounts`, `sessions` and `links` with their indexes, `Postman` (the mails of an account, sent through `post`), the setting `base-url`, the queries `member` and `member-by-email` (the directory), `Stretcher` (PBKDF2), `Drawer` (secrets), `WallClock` |
| `identity/authapi` | authapi (presentation) | the routes of `/api/auth`, the session handler `Authenticate`, the cookie `todo_session`, the setting `verify-addresses` |
| `contacts/link` | link (domain) | `Link`, `PairOf`, the `Requests` workflow, the errors |
| `contacts/circle` | circle (application) | the use cases `RequestContact` `AnswerRequest` `RemoveContact` `ListContacts` `AreContacts`, the ports `Links` `People` `Notices` `Clock` |
| `contacts/linkstore` | linkstore (adapter) | the store `links`, the topic `events`, `Links`, `Directory` (asks the identity's queries), `Announcer`, `WallClock` |
| `contacts/contactsapi` | contactsapi (presentation) | the routes of `/api/contacts` |
| `tasks/task` | task (domain) | `Task`, `Priority`, the `Lifecycle` workflow, the errors |
| `tasks/planner` | planner (application) | the use cases `AddTask` `MoveTask` `ListTasks` `ArchiveDone` `CountTasks` `DueSoon`, the ports `TaskStore` `Clock` `Journal`; `lifecycle.go` is what they share |
| `tasks/taskstore` | taskstore (adapter) | the store `tasks`, the topic `events`, `Tasks`, `Announcer` (the use cases' `Journal`, which publishes) and `AppClock` |
| `tasks/api` | api (presentation) | the routes of `/api/tasks`, the query `census`, the job `archive-done`, the setting `archive-after` |
| `activity/feed` | feed (adapter) | the store `entries`, the subscriptions `tasks` and `contacts` to their events, `GET /api/activity` |
| `notify/reminders` | reminders (application) | `RemindDue`, the ports `Mailer` and `ReminderLog` |
| `notify/post` | post (adapter) | the mailer `mail` (the framework's outbox and transport), the command `send` every mail goes through, `Carrier` (the reminders' `Mailer`, its SMTP connection declared in the design) |
| `notify/outbox` | outbox (adapter) | the store `reminded`, `Members` (the reminders' `Directory`, which asks the identity's query) and `Ledger` |
| `notify/watch` | watch (presentation) | the job `remind-due`, the setting `remind-within` |
| `stats/census` | census (presentation) | the store `snapshots`, the job `sample`, `GET /api/stats` |
| `web/site` | site (presentation) | the web app: `dist/index.html`, `app.css`, `app.js`, embedded and served at `/` |
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
- **A component's data is its own**: another component reaches it through
  a query the owner declares (`vault.MemberQuery.Ask`,
  `api.CensusQuery.Ask`), a topic it publishes or a port, never its store.
- **A subscription is delivered at least once, and later**: it keys what
  it writes by the event's own id and writes nothing the second time
  (`feed.record`). A test that reads what a subscription wrote moves the
  clock until it is there.
- **Time is the framework's**: `app.Now(ctx)`, in UTC, to the millisecond
  (`AppClock`, and the adapters that stamp a record). A test moves it.
- **Ask "is there one?" with `Filter`, not `Get`**: a missing key is a
  failed read (`Ledger.Seen`). `Tasks.Get` reads the key because a task
  that does not exist is the caller's error, answered 404.
- **Every mail goes through `post`**, in French, as plain text, its
  subject one line: a component that writes to a member dispatches
  `post.SendCommand` or goes through `Carrier`, never a transport of its
  own. A test reads what was sent with `post.MailMailer.Captured()`.
- **Logs name a task by its id**, never an assignee: `Assignee`, a mail's
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
