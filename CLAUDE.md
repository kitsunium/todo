<!-- updated: 2026-10-03T19:05:00Z -->
# kitsunium/todo

The reference product of kit (`github.com/kitsunium/platform`), and its
end-to-end test: a task list for people who work together — accounts,
contacts, groups, shared tasks, mail. Read `README.md` for what it does.

It is written on the SDK's framework and imports the SDK alone:
`github.com/kitsunium/sdk/framework` (`kit`, `model`) and
`github.com/kitsunium/sdk/pkg` (`pkg/v1/*`), both v0.17.0, published, no
`replace`. kit, the platform's tool, runs the product (`kit dev`) and reads
its source (the static analysis); the product links none of it.

## Layout

| Path | Service | Role |
|---|---|---|
| `main.go` | — | the App: eight services, one binary; `App.Main` gives serve / graph / healthcheck / config / secrets. It imports `framework/kit/server`: the framework serves HTTP only in a program that links it, and refuses to start a server that does not (`profile.server-missing`) |
| `kitdev.go` | — | `//go:build kitdev`, the tag `kit dev` builds with: imports `framework/kit/studio`, the Studio's API at `/_kit/` in dev. A release build, and the tests, link none of it |
| `identity/` | identity | accounts (workflow `accounts`) and their language, one-time links (workflow `tokens`), sessions, the auth handler `session`, the directory (the queries `UsersByID` → `Directory`, `People` with each user's locale and time zone; `EmailOwner`), the hand-written `session-reaper` loop |
| `contacts/` | contacts | links between users (workflow `requests`), invitations by email, `claim-invites` on `identity.AccountEvents`, the query `AreContacts` |
| `groups/` | groups | groups and roles, invitations (workflow `invitations`), the queries `RoleOf` `GroupRefs` `UserGroups` `GroupMembers` |
| `tasks/` | tasks | the model other services import: `Task`, the `Tasks` store, the `Events` topic, the queries `TaskCensus` `OpenTasks` `TaskAudience` |
| `tasks/api/` | tasks | the `lifecycle` workflow, the whole `/api/tasks` API, views and counts, `group-changes` |
| `notify/` | notify | the `mail` mailer, the command `SendMail` for identity's transactional mails, the mails, their one light design (`templates/`) and their words (`locales/fr.json`, `locales/en.json`); the setting `base-url` (the links of the mails) |
| `notify/dispatch/` | notify | who is mailed when: `task-mail` `contact-mail` `group-mail` `track-due`, the `reminders` store and declared loop |
| `activity/` | activity | the feeds: `entries` (`kind`, `actor`, `target`, `task`, `group` for a client to write in its reader's language; `text` in English), three subscriptions, the query `UnreadCount` |
| `stats/` | stats | the `sample` job and `GET /api/stats` |
| `internal/wire/` | — | wire conventions: `Now` (UTC, ms), `Line` (one-line text), `Invalid` (a violation), `Is` (error code), `Locale` (`fr` first, `en`; `NegotiateLocale`, `CheckLocale`, `Resolve`), time zones (`CheckTimeZone`, `Zone`; the zone database is embedded, `time/tzdata`) |
| `web/` | web | the SPA (`web/src` → committed `web/dist`), owned by the web agent |
| `*_test.go` | — | end-to-end tests, including the diagram suite: `TestTheDiagramMatchesTheCode` (`diagram_test.go`) and `TestTheReadmeDiagramIsCurrent` (`main_test.go`). Their static half is kit's analysis, run as a process (`kit graph -static -format json`) and given to the product through `kit.Analyzer` |

## Rules

- Declare building blocks as package-level variables: the static analysis
  only reads those. Call a building block through its variable
  (`tasks.Tasks.Get`, `groups.RoleOf.Ask`), never through a value.
- A service's data is its own: another service asks its queries or
  dispatches its commands — internal, as long as nobody exposes them — or
  listens to its topics, never through the store.
- Go imports must stay a tree while services call each other both ways.
  When a service is called by a service it must itself call or listen to,
  it declares its shared part (Service, model, store, topic, the queries
  and commands others run) in `svc/` and the rest in a sub-package on the
  same `svc.Service` (`tasks/api`, `notify/dispatch`); `main.go` imports
  that sub-package blank.
- Fire a workflow event with a constant name at the call site
  (`Lifecycle.Fire(ctx, id, "complete")`): the analysis labels the edge
  with it. Never call another endpoint's handler function directly.
- Subscriptions are at-least-once: key what they write by the event's own
  ID, insert rather than overwrite, and do the one irreversible thing — a
  mail — last.
- Every timestamp and deadline reads `wire.Now(ctx)` (kit's clock): the
  tests move it. Logs go through `kit.Log(ctx)` and name users by ID, never
  by address.
- A transition due at a date the entity carries (a due date, a link's
  deadline) is an `At`: kit wakes the workflow's loop at that instant. A
  `When` guard sees the entity only and is checked each time it is written —
  never a condition on the time. No workflow loop runs while nothing is due.
- Errors to callers are `kit.Error` values: `wire.Invalid` for a field,
  `kit.NotFound` for what the caller may not see (never 403 for a task, a
  group or a contact of someone else), and the codes of the contract
  (`invalid_credentials`, `email_unverified`, `account_locked`,
  `invalid_token`). A kit error is matched by its wire code:
  `wire.Is(err, kit.WireNotFound)`, `kit.WireConflict`. Never return or log
  a password, a secret or a hash; a secret is stored as its SHA-256, a
  password with the SDK's `password.Hash` (`pkg/v1/crypto/password`).
- Ask "is there one?" with `Find` on an index, not `Get`: a missing key is
  an error span, and the Studio paints the store red.
- The product speaks French first, and English. An account's `Locale` is
  set at sign-up (the form's `locale`, else the SDK's `i18n.Negotiator` over
  `Accept-Language`, else French) and changed by `PATCH /api/auth/me`; an
  account without one reads French (`Locale.Resolve`). A service that
  writes to a user gets the language from the directory (`identity.People`)
  or the call (`SendInput.Locale`), never from identity's store.
- An account's `TimeZone` is the IANA zone the browser gave at sign-up and
  at each sign-in (`timeZone`, checked by `wire.CheckTimeZone`: spelled as
  the database spells it, so every OS answers alike); the times in a mail
  are written in the reader's zone (`wordsIn(l).at(zone)`, `{zone}` in
  `date.long`), UTC when they gave none.
- Settings are declared where they are used (`Service.Setting`), never read
  from the environment by hand: `base-url` (notify), `archive-after` (tasks,
  the auto-archive timer waits it). A test gives one with `kit.Set`.
- Every word of a mail comes from `notify/locales/{fr,en}.json` (flat keys,
  `{name}` placeholders, CLDR plural forms for a count) through the SDK's
  `i18n` printers. A new mail adds its keys to both files: the process
  refuses to start on a key one language lacks or a missing plural form.
  French is written for French readers — vous, `’`, `\u00a0` before
  `: ? !` and inside `« »`, no elision a name could break ("envoyée par
  {actor}", not "de {actor}") — and mails stay light: no dark mode.
- Import the SDK only — `sdk/framework/*` and `sdk/pkg/v1/*` — never
  `github.com/kitsunium/platform`: a product never imports the platform
  (its ADR 0010, D2; `kit check`'s `imports` rule). What the platform gives
  the product, it gives as a tool: `kit dev`, and the static analysis the
  diagram tests run through `kit graph -static`.
- The README's Mermaid diagram is generated: after a change to the graph,
  replace it with what `TestTheReadmeDiagramIsCurrent` prints where kit is —
  the product's `graph -format mermaid`, given kit's static analysis. `go
  run . graph -format mermaid` prints its solid arrows alone (what the
  product declares): the dashed ones are what kit finds in the code.

## Develop

The build needs the Go toolchain and the module proxy, nothing else: no
platform checkout and no `go.work` (git-ignored, machine-local: build with
`GOWORK=off` to be sure none is read). kit is installed apart, from a
platform checkout: `go install ./cmd/kit` there.

```sh
kit dev                             # :4000; kit prints the Studio's link, mails in its Mail view
GOWORK=off go build ./... && GOWORK=off go vet ./... && GOWORK=off go vet -tags kitdev ./...
GOWORK=off go test -race ./...      # end to end, on a manual clock
gofmt -l .                          # empty
docker build -t todo .
```

The diagram suite's static half needs kit: kit on `PATH`, or `KIT=<its
binary>`. Without it, `TestTheDiagramMatchesTheCode` checks what the
running product says of itself and `TestTheReadmeDiagramIsCurrent` the
diagram's solid arrows, then each skips the rest, naming it (`go test -v`
shows the skip). With it, both run whole:

```sh
KIT=$(command -v kit) GOWORK=off go test -race -v -run 'TestTheDiagramMatchesTheCode|TestTheReadmeDiagramIsCurrent' .
```

Commits: conventional, authored as `kodflow` (the `post-commit` gate checks
it), no AI attribution. Never delete `.github/workflows/post-commit.yml`.
