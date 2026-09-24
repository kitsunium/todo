<!-- updated: 2026-09-24T00:19:47Z -->
# kitsunium/todo

The first product built with kit (`github.com/kitsunium/platform`) — and its
end-to-end test. Read `README.md` for what it does.

## Layout

| Path | Role |
|---|---|
| `main.go` | the App: four services, one binary; `App.Main` gives serve / graph / healthcheck |
| `todos/` | the list, its store, the `changes` topic and the `lifecycle` workflow; the API |
| `activity/` | the feed, a subscription to `todos.changes` |
| `stats/` | a job sampling the list every 10 s through `todos.CensusAPI` (private endpoint) |
| `web/` | the UI (`web/assets`, no build step), served by `Service.Static` |
| `main_test.go` | end-to-end tests, including `TestTheDiagramMatchesTheCode` |

## Rules

- Declare building blocks as package-level variables: the static analysis
  only reads those.
- A service's data is its own: another service goes through an endpoint
  (`CensusAPI.Call`), never through the store. The diagram shows the call.
- Subscriptions are at-least-once: key what they write by the message's own
  ID (`Change.ID`), never by something two messages can share.
- The UI renders user text with `textContent` only; the frontend CSP forbids
  inline scripts.
- The README's Mermaid diagram is generated: after a change to the graph,
  replace it with the output of `go run . graph -format mermaid`
  (`TestTheReadmeDiagramIsCurrent` checks it).

## Develop

The platform is not published yet. `go.mod` replaces it with `../platform`;
on a machine where the platform branch lives elsewhere, an untracked `go.work`
with a `replace` points at it.

```sh
kit dev              # :4000, Studio at /_kit/, rebuild on save
go test -race ./...  # end to end, with a manual clock
docker build --build-context platform=../platform -t todo .
```

Commits: conventional, authored as `kodflow` (the `post-commit` gate checks
it), no AI attribution. Never delete `.github/workflows/post-commit.yml`.
