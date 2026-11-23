# VIDI

VIDI has two independent pieces:

1. `vidi.js` — a dependency-free browser library that renders records from a JSON or XML endpoint as escaped, paginated cards with CSV export.
2. `main.go` — an optional Go template editor demo with a small API.

The browser library needs no server. The Go demo is a local editor for stored HTML templates, not a data-management platform.

Loopback and container-publishing rules for all six demos: [`../docs/LOCAL-DEMOS.md`](../docs/LOCAL-DEMOS.md). Runnable scripts and a loopback-only Compose file: [`../examples`](../examples/README.md).

## Browser library (`vidi.js`)

Construct with `new Vidi({ dataSource: "/api/pod/table/acme/items?format=json", pageSize: 12 })`. Exports `window.Vidi` and `window.vidi` (the same class), plus a CommonJS export for tests.

- Accepts `{records:[...]}`, `{items:[...]}`, a single `{record:...}`, or a bare array; pod-style `{id, fields:{...}}` rows are flattened.
- Parses XML of the form `<records><record id="..."><field name="...">value</field></record></records>`.
- Renders every value with text nodes, never `innerHTML`; URLs are only linked when the scheme is `http`, `https`, `mailto`, or `tel`.
- Provides a pager with a five-page window and an `Export CSV` button for all loaded rows.
- Calls `onRender(rows, container)` and `onError(err)` callbacks when supplied.

## Go template editor demo

```bash
go test ./...
node --test vidi_test.js
go run .                 # http://127.0.0.1:8084
```

| Variable | Default | Meaning |
| --- | --- | --- |
| `VIDI_HOST` | `127.0.0.1` | Bind address. Non-loopback binds are refused without a token. |
| `VIDI_PORT` | `8084` | Bind port. |
| `VIDI_API_TOKEN` | empty | Bearer/`X-Vidi-Token` value for API routes. Static assets are exempt. |

Routes:

| Endpoint | Purpose |
| --- | --- |
| `GET /` | Template list UI. |
| `GET /{name}` | Serve a stored template's HTML verbatim. |
| `GET /api/templates` | List template names and bodies. |
| `POST /api/templates` | Save `{name, html}` (1 MiB body cap). |
| `GET /api/templates/{name}` | Read a template. |
| `DELETE /api/templates/{name}` | Delete a template. |
| `GET /static/` | UI assets. |

Templates live as JSON files under `templates_data/` and are written atomically with mode `0600`. Names are restricted to a single conservative filename component, with a path check as defence in depth. The handler sends `nosniff`, frame-deny, and referrer headers, and uses constant-time token comparison.

**Stored templates are executable HTML.** They are served as-is without sanitization, so a caller with write access can run script on this origin. Do not share the origin with sensitive applications.

`SECURITY.md` describes the actual implementation; the previous document listed features such as AES encryption, RBAC, MFA, SSO, audit logging, and DDoS protection that do not exist.

## Tests

`go test ./...` and `node --test vidi_test.js` both pass. The JS suite covers escaping and URL-scheme handling.

## Remaining gaps

- No accounts, per-user authorization, CSRF protection, rate limiting, or TLS.
- The token is a single shared demo credential.
- Pagination, XML parsing, and CSV export live in `vidi.js`, not in the Go API.
- The stale prebuilt binary was removed; build from source with `go build .` (it is now gitignored).
