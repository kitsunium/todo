<!-- updated: 2026-09-24T09:32:49Z -->
# web — the todo's user interface

A single page application: React 19, TypeScript (strict), Vite, Tailwind CSS v4,
Radix primitives, TanStack Query, React Router, cmdk, sonner, motion, date-fns and
lucide icons, with Inter self-hosted. One light design, and an interface in French
first, English second. `npm run build` writes `dist/`, which
`web.go` embeds (`//go:embed dist`) and serves at `/` with
`Service.Static("app", "/", dist, kit.Root("dist"))`. A path without an extension
that matches no file serves `index.html`, so client routes survive a reload.

`dist/` is committed: the Go build never needs Node. Rebuild it after any change
under `src/`.

## Develop

```sh
npm ci
npm run dev:mock     # http://localhost:5173 — the whole API in memory, no backend
npm run dev          # http://localhost:5173 — /api and /_kit proxied to KIT_APP
                     # (default http://localhost:4000: run `KIT_ENV=dev go run .` first)
```

**Mock mode** (`vite --mode mock`) answers every route of the HTTP contract from
`src/mock/`, with seeded people, contacts, groups, tasks in every priority, status
and due case, activity, sessions — and a mailbox: the "Check your inbox" and
"Forgot password" pages list the mails the mock sent, with their links, so
signup → verify and forgot → reset can be clicked through. State lives in
`localStorage` (`todo.mock.db`) and is re-seeded every day.

| Mock account | Result |
|---|---|
| `camille@example.com` / `correct-horse-42` | the seeded workspace (in French: switch it in Settings) |
| `unverified@example.com` / `correct-horse-42` | `email_unverified` |
| `locked@example.com` / anything | `account_locked` |

In the browser console: `__mock.reset()`, `__mock.signOut()`, `__mock.signIn()`,
`__mock.setLatency(ms)`. Production code never imports `src/mock/`: every entry
point sits behind `import.meta.env.MODE === "mock"`, which a production build
replaces with `false` and drops — `grep -r __mock dist` finds nothing.

With a real app in dev, the "Check your inbox" page links to the Studio's dev
mailbox (`/_kit/#/mail`), shown only when `/_kit/api/graph` answers JSON.

## Languages

The interface speaks French (the default) and English, with no library:
`src/i18n/fr.ts` is the source of truth, `en.ts` must have exactly its keys
(TypeScript refuses a missing or extra one, `tests/i18n.test.ts` checks the
`{placeholders}` match). `useT()` gives a translator for the current language:
`t("view.today")`, `t("group.members", { count })` — plurals through
`Intl.PluralRules` ("0 tâche", "2 tâches") — and `t.rich(key, { name: <b/> })`
for React nodes in the holes. Dates go through date-fns locales and `Intl`
(`jeu. 24 sept.`, `17:00`, `il y a 3 h`); the French calendar starts on Monday.

Which language: signed in, the account's (`user.locale` from the API); signed
out, the last choice made in this browser (`localStorage["todo.locale"]`), else
French. The FR | EN switch lives in the user menu, in Settings, in the palette and
at the foot of the signed-out pages; signed in, it saves the account's language
with `PATCH /api/auth/me {"locale"}` (optimistic, rolled back on failure), which
also sets the language of the mails the server sends. A signup sends the page's
language. `<html lang>` follows.

API errors are shown from their code and each violation's rule, in the current
language; the server's English words are only a fallback for a code the app does
not know. Activity entries are written from their `kind`, `actor`, `task`, `group`
and `target` (the person a share, an assignment or a removal was about); the
server's English `text` is only read for an entry written before `target` existed.
The quick add understands both languages whatever the interface's:
`Appeler Marco demain à 10h !haute #lancement`, `Call Sam tomorrow at 5pm !high`.

## Check

```sh
npm run typecheck    # tsc, strict, app and node configs
npm test             # vitest: quick-add parser (FR and EN), dates, error mapping, dictionaries, passwords
npm run build        # dist/
```

## Screenshots

`scripts/shoot.mjs` drives headless Chrome through the DevTools protocol (Node's
own WebSocket; not bundled): every screen in French and in English at 1440×900,
and a 390×844 mobile pass, with real mouse and keyboard events (menus, popovers, the
panel, the palette, quick-add chips, the completion frames). Console errors and
exceptions go to `$OUT/console.log`; the script exits 1 if there is any.

```sh
npx vite --mode mock --port 5299 &
node scripts/shoot.mjs                     # → /tmp/todo-web-shots
ONLY=today,panel LOCALES=en MOBILE=0 SCALE=2 node scripts/shoot.mjs

# The real app (KIT_ENV=dev go run . — seeded users, mails in /_kit/api/mail):
REAL=1 BASE=http://127.0.0.1:4000 OUT=/tmp/todo-web-real node scripts/shoot.mjs
```

`REAL=1` signs in as `REAL_EMAIL` (default `camille@example.com`, password
`REAL_PASSWORD`) and first adds demo data through the API — an invitation to a
"Book club" group, a task shared with and one assigned to Camille, an incoming and
an outgoing contact request, an email invitation — and creates throwaway accounts
for the sign-up, unverified and locked screens (verify links come from the dev
mailbox). Scenes that complete a task undo it. Its console log also lists the
browser's network entries for the intended 4xx answers (signed-out `/api/auth/me`,
wrong passwords, a used link): they are not app errors. `PROD=1` serves the
signed-out screens of a build without a backend, to check the CSP.

## Layout

| Path | Role |
|---|---|
| `src/api/` | the contract's types, a fetch client (`ApiError{status,code,message,violations}`, 401 → sign in again), one function per route, TanStack Query hooks with optimistic updates and rollback, error wording |
| `src/i18n/` | the dictionaries (`fr.ts`, `en.ts`), the current language, `useT()`, plurals and rich messages |
| `src/lib/` | pure logic: the quick-add language, dates and sections, avatars, password strength |
| `src/components/ui/` | the design system: buttons, fields, menus, popovers, dialogs, avatars, chips, tabs, skeletons, toasts |
| `src/components/brand/` | the fox, the sign-in art, the empty-state illustrations (original SVG) |
| `src/features/` | auth pages, the shell (sidebar, palette, shortcuts), tasks (list, row, quick add, panel, pickers), groups, contacts, activity, settings |
| `src/mock/` | mock mode only |
| `tests/` | vitest |
| `scripts/shoot.mjs` | screenshots |

## Rules

- User text is rendered by React only — no `dangerouslySetInnerHTML`, no HTML
  from the API.
- The bundle stays clean for the frontend CSP (`default-src 'self'; img-src 'self'
  data:; style-src 'self' 'unsafe-inline'; …`): no inline script, no eval, no
  external URL, no inlined asset (`assetsInlineLimit: 0`; `font-src` has no
  `data:`).
- `//go:embed` skips files whose name starts with `.` or `_`: `vite.config.ts`
  renames such chunks.
- Dependencies are pinned to exact versions; `package-lock.json` is committed.
- Light only: no dark theme, no `prefers-color-scheme` switch.
- Every visible string comes from `src/i18n/` — never a literal in a component.
- Times are `17:00` in French and follow the person's clock in English (`5:00 PM`
  or `17:00`). A due without a time is sent as 23:59 local time; lists and counts pass `tz` (the
  browser's IANA zone) so "today" ends where the person is.

## Keyboard

`⌘K` palette · `/` search tasks · `C` new task · `?` shortcuts · `G` then
`I T U S A C` go to a view (the same letters in both languages) · in a list `J`/`K` move, `X` complete, `E`/`↵` open,
`1`–`4`/`0` priority, `Esc` let go · in the panel `J`/`K` next/previous task.
