<!-- updated: 2026-11-16 -->
# design — the design, the one source

- `product.yaml`: the product, its context (the actors `member` and
  `operator`, the system `mail-relay`), its evidence, and the binary `todo`
  with its roles `server` (profile server) and `load` (profile cli, command
  `load`), each with the components it mounts.
- `tasks.yaml`, `activity.yaml`, `notify.yaml`, `stats.yaml`, `load.yaml`:
  one domain each, its components, and its sequences.
- `generated.lock` is written by every generation: the design's digest and
  the files it wrote. Never edit it.

## Rules

- Component codes are `64.<domain>.<component>`, never reused: tasks is 1,
  activity 2, notify 3, stats 4, load 5.
- A description in a flow mapping (`{name: …, description: …}`) that holds a
  comma or a colon is quoted.
- A connection is declared on the adapter that opens the socket, with its
  default port: `Relay` dials the mail relay on 587, `Client` dials the
  server on 4000. The server's own listening socket is the framework's, no
  adapter's, so the design declares none for it.
- The product is closed: a new package or exported symbol is declared here
  first, and `product.exceptions` stays empty.
