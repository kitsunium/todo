# todo

A task list for a team, served over HTTP by one process — and a product
written design first: `design/` says its structure, a generator writes the
shells from it, and the bodies beside them are written by hand. It runs on
the SDK's framework
([`github.com/kitsunium/sdk/framework/app`](https://github.com/kitsunium/sdk))
and imports the SDK alone.

A member signs up with an address and a password, proves the address by
the link they are mailed, signs in, and has a task
list of their own. They add tasks with a priority, an assignee and a due date, and moves
them through their lifecycle: open, done, archived. Every move is published on a topic, and lands in their activity feed a
moment later. A task done for a day is archived on its own, the owner of a task
due within a day gets one reminder by mail, and the task list
is counted every thirty seconds. Two members become contacts by a
request one sends to the other's address and the other accepts. The server serves a small web app at `/`,
in French: sign up, sign in, add tasks, move them, read the feed.

The former todo, written on the framework by hand, is where the rest still
lives — mails in two languages and in HTML, invitations of people
who have no account, groups, shared tasks, its richer web app — until the design can say it. It is kept whole in [`legacy/`](legacy/), a Go
module of its own.

## The design

| File | Domain | Holds |
|---|---|---|
| `design/product.yaml` | — | the product, its context (a member, an operator, the mail relay) and the binary `todo` with its two process roles, `server` and `load` |
| `design/identity.yaml` | identity | the account and its lock, the session, four use cases and their five ports, the stores behind them, the HTTP API and the handler that says who calls a route, the sequence `sign-in` |
| `design/contacts.yaml` | contacts | the link between two members and its request, five use cases and their four ports, the store and the topic behind them, the HTTP API, the sequence `request-contact` |
| `design/tasks.yaml` | tasks | the task and its lifecycle, six use cases and their three ports, the store behind them, the HTTP API, the sequence `complete-task` |
| `design/activity.yaml` | activity | the feed, written from what the use cases record |
| `design/notify.yaml` | notify | the reminders, the mailer with its outbox and its SMTP connection, the command every mail goes through, the sequence `remind-due` |
| `design/stats.yaml` | stats | the vital signs, sampled by a job |
| `design/web.yaml` | web | the web app's static files |
| `design/load.yaml` | load | a load driven against a running server over its HTTP API |

Each component of the design is a Go package, `<domain>/<component>`, whose
`gen_*.go` files are generated and never edited; `cmd/todo` is the binary.

## The API

| Route | Does |
|---|---|
| `POST /api/auth/signup` | creates an account: `email`, `name`, `password` (10 to 128 characters), and mails the link that proves the address; answers the same for an address that already has an account |
| `POST /api/auth/verify` | proves an address: the `token` of the link |
| `POST /api/auth/verify/resend` | mails a new link to an `email` not verified yet |
| `POST /api/auth/password/forgot` | mails an `email` the link that chooses a new password |
| `POST /api/auth/password/reset` | chooses a new `password` with the `token` of that link, and ends every session |
| `POST /api/auth/password` | changes the caller's password: `current`, `password`; ends their other sessions |
| `POST /api/auth/login` | opens a session: `email`, `password`; sets the `todo_session` cookie and answers the same secret as `token` |
| `POST /api/auth/logout` | ends the caller's session |
| `GET /api/auth/me` | the caller |
| `GET /api/contacts` | the caller's contacts, the requests they received and those they sent |
| `POST /api/contacts` | asks the member an `email` belongs to to be a contact |
| `POST /api/contacts/{id}/accept` `decline` `cancel` | answers a request: its addressee accepts or declines, its requester cancels |
| `DELETE /api/contacts/{id}` | ends a contact, for either member |
| `POST /api/tasks` | adds a task: `title`, and `notes`, `priority` (`low`, `normal`, `high`), `assignee`, `due` (RFC 3339) |
| `GET /api/tasks` | lists the caller's tasks not archived, the most urgent first; `?state=open\|done\|archived` lists one state |
| `POST /api/tasks/{id}/complete` `reopen` `archive` `restore` | moves a task through its lifecycle |
| `GET /api/activity` | the caller's feed, newest first; `?limit=N` |
| `GET /api/stats` | the latest sample of the task list and the recent history |

Every route but those a stranger needs — signing up, the links, signing
in — and `/api/stats` asks for a member: the
session cookie, or `Authorization: Bearer <token>`; without one it answers
401. A member sees their own tasks and feed, and another member's
task is answered as if it did not exist.

Every route that writes is limited per client — the member, or the address
of a stranger: two calls a second with a burst of five for signing up,
signing in and asking for a contact, ten a second with a burst of twenty
for the rest. A call over the limit is answered 429.

A request the API refuses says why: 400 with the field at fault, 404 for a
task that does not exist, 409 for a move the task's state does not allow,
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
| `SERVER_BASE_URL` | `http://localhost:4000` | where members reach the product, for the links of its mails |
| `SERVER_VERIFY_ADDRESSES` | `true` | whether a new account proves its address before it signs in; `false` where the deployment has no mail relay |
| `SERVER_SMTP_URL` | unset | where the mailer delivers: `smtp://user:pass@host:587?tls=starttls`; unset, the mails are captured for the dev tool and not delivered |

What the stores keep of a person has an end: a task's assignee is erased a
year after the task was last written, a feed entry is deleted after ninety
days, an account untouched for three years is erased,
and a session is deleted a day after it expired. A password is kept
stretched (PBKDF2-SHA256), a session's secret as its SHA-256.

`go run ./cmd/todo config` prints every setting and where its value comes
from.

## The web app

`web/site/dist` holds three files written by hand — `index.html`, `app.css`,
`app.js` — with no build step and no dependency; the binary embeds them. The
page talks to the API above with the session cookie. The server answers the
page for any path that names no file, and its content policy allows scripts
and styles of its own origin only: nothing inline.

## Drive a load

The binary's second role calls a running server's API — it signs up and in
as a member of its own, `load@example.com`, adds tasks,
completes, archives and reopens them, and reads the list, the feed and the
vital signs — so that there is something to watch:

```sh
LOAD_TARGET=127.0.0.1:4000 go run ./cmd/todo load -for 2m -rate 5
```

The load reads no mail: the server it is driven against runs with
`SERVER_VERIFY_ADDRESSES=false`, or its member's address was proven once by
hand.

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
