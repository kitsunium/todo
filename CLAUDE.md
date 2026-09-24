<!-- updated: 2026-09-24T09:54:21Z -->
# kitsunium/todo

The reference product of kit (`github.com/kitsunium/platform`), and its
end-to-end test: a task list for people who work together — accounts,
contacts, groups, shared tasks, mail. Read `README.md` for what it does.

## Layout

| Path | Service | Role |
|---|---|---|
| `main.go` | — | the App: eight services, one binary; `App.Main` gives serve / graph / healthcheck |
| `identity/` | identity | accounts (workflow `accounts`) and their language, one-time links (workflow `tokens`), sessions, the auth handler `session`, the directory (`UsersAPI` → `Directory`, `People` with each user's locale; `UserByEmailAPI`), the hand-written `session-reaper` loop |
| `contacts/` | contacts | links between users (workflow `requests`), invitations by email, `claim-invites` on `identity.AccountEvents`, `CheckAPI` |
| `groups/` | groups | groups and roles, invitations (workflow `invitations`), `RoleAPI` `BatchAPI` `MembershipsAPI` `MembersAPI` |
| `tasks/` | tasks | the model other services import: `Task`, the `Tasks` store, the `Events` topic, `CensusAPI` `OpenByGroupAPI` `AudienceAPI` |
| `tasks/api/` | tasks | the `lifecycle` workflow, the whole `/api/tasks` API, views and counts, `group-changes` |
| `notify/` | notify | the `mail` mailer, `SendAPI` for identity's transactional mails, the mails, their one light design (`templates/`) and their words (`locales/fr.json`, `locales/en.json`) |
| `notify/dispatch/` | notify | who is mailed when: `task-mail` `contact-mail` `group-mail` `track-due`, the `reminders` store and declared loop |
| `activity/` | activity | the feeds: `entries` (`kind`, `actor`, `target`, `task`, `group` for a client to write in its reader's language; `text` in English), three subscriptions, `UnreadAPI` |
| `stats/` | stats | the `sample` job and `GET /api/stats` |
| `internal/wire/` | — | wire conventions: `Now` (UTC, ms), `Line` (one-line text), `Invalid` (a violation), `Is` (error code), `Locale` (`fr` first, `en`; `NegotiateLocale`, `CheckLocale`, `Resolve`) |
| `web/` | web | the SPA (`web/src` → committed `web/dist`), owned by the web agent |
| `*_test.go` | — | end-to-end tests, including `TestTheDiagramMatchesTheCode` |

## Rules

- Declare building blocks as package-level variables: the static analysis
  only reads those. Call a building block through its variable
  (`tasks.Tasks.Get`, `groups.RoleAPI.Call`), never through a value.
- A service's data is its own: another service goes through an endpoint —
  a private one, `kit.Private()` — or a topic, never through the store.
- Go imports must stay a tree while services call each other both ways.
  When a service is called by a service it must itself call or listen to,
  it declares its shared part (Service, model, store, topic, private
  endpoints) in `svc/` and the rest in a sub-package on the same
  `svc.Service` (`tasks/api`, `notify/dispatch`); `main.go` imports that
  sub-package blank.
- Fire a workflow event with a constant name at the call site
  (`Lifecycle.Fire(ctx, id, "complete")`): the analysis labels the edge
  with it. Never call another endpoint's handler function directly.
- Subscriptions are at-least-once: key what they write by the event's own
  ID, insert rather than overwrite, and do the one irreversible thing — a
  mail — last.
- Every timestamp and deadline reads `wire.Now(ctx)` (kit's clock): the
  tests and the Studio move it. Logs go through `kit.Log(ctx)` and name
  users by ID, never by address.
- Errors to callers are `kit.Error` values: `wire.Invalid` for a field,
  `kit.NotFound` for what the caller may not see (never 403 for a task, a
  group or a contact of someone else), and the codes of the contract
  (`invalid_credentials`, `email_unverified`, `account_locked`,
  `invalid_token`). Never return or log a password, a secret or a hash;
  a secret is stored as its SHA-256, a password with the SDK's
  `password.Hash`.
- Ask "is there one?" with `Find` on an index, not `Get`: a missing key is
  an error span, and the Studio paints the store red.
- The product speaks French first, and English. An account's `Locale` is
  set at sign-up (the form's `locale`, else the SDK's `i18n.Negotiator` over
  `Accept-Language`, else French) and changed by `PATCH /api/auth/me`; an
  account without one reads French (`Locale.Resolve`). A service that
  writes to a user gets the language from the directory (`identity.People`)
  or the call (`SendInput.Locale`), never from identity's store.
- Every word of a mail comes from `notify/locales/{fr,en}.json` (flat keys,
  `{name}` placeholders, CLDR plural forms for a count) through the SDK's
  `i18n` printers. A new mail adds its keys to both files: the process
  refuses to start on a key one language lacks or a missing plural form.
  French is written for French readers — vous, `’`, `\u00a0` before
  `: ? !` and inside `« »`, no elision a name could break ("envoyée par
  {actor}", not "de {actor}") — and mails stay light: no dark mode.
- The README's Mermaid diagram is generated: after a change to the graph,
  replace it with `go run . graph -format mermaid`
  (`TestTheReadmeDiagramIsCurrent`).

## Develop

The platform is not published yet. `go.mod` replaces it with `../platform`;
on a machine where the platform branch lives elsewhere, an untracked `go.work`
with a `replace` points at it.

```sh
kit dev                             # :4000, Studio at /_kit/, mails in its Mail view
go vet ./... && go test -race ./... # end to end, on a manual clock
gofmt -l .                          # empty
docker build --build-context platform=../platform -t todo .
```

Commits: conventional, authored as `kodflow` (the `post-commit` gate checks
it), no AI attribution. Never delete `.github/workflows/post-commit.yml`.
