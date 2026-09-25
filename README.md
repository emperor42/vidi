# VIDI

VIDI is a small, dependency-free browser library for rendering records from a
JSON or XML endpoint as escaped cards, with pagination and CSV export.

The root of this component contains two separate things:

- **`vidi.js`** is the browser API. It does not need a Node.js runtime, a
  bundler, or a Go server.
- **`main.go`, `static/`, and `templates/`** are an optional Go template-editor
  demo. They are not required to use `vidi.js`.

The archived material in [`legacy/`](legacy/) is retained for reference and is
not part of the current build or API.

## Browser API

### Include the library

Serve `vidi.js` as a normal browser asset:

```html
<script src="/path/to/vidi.js"></script>
<div id="vidi-cards-container"></div>
<div id="vidi-pagination"></div>
```

Construct an instance after those elements exist. Construction starts a fetch
immediately; there is no `init()`, automatic singleton, or automatic DOM scan:

```html
<script>
  const view = new Vidi({
    dataSource: "/s/data/azzurrotech/posts",
    pageSize: 12,
    fields: ["title", "body", "link"],
    titleField: "title",
    onRender: function (rows, container) {
      console.log("rendered " + rows.length + " rows", container);
    },
    onError: function (error) {
      console.error(error);
    }
  });
</script>
```

`window.Vidi` and `window.vidi` both refer to the constructor. A CommonJS
consumer can use `const { Vidi } = require("./vidi.js")`.

### Constructor options

| Option | Default | Description |
| --- | --- | --- |
| `dataSource` | required | URL passed to `fetch`. |
| `container` | `#vidi-cards-container` | CSS selector for the cards element. |
| `pagination` | `#vidi-pagination` | CSS selector for the pager element. |
| `pageSize` | `12` | Number of records shown per page. |
| `fields` | all non-bookkeeping fields | Non-empty array of field names to display. |
| `titleField` | automatic | Preferred title field. The automatic order is `title`, `name`, `subject`, then `label`. |
| `onRender` | none | Called with `(rows, container)` after a render. Callback errors are ignored. |
| `onError` | none | Called when loading fails. |

The instance exposes `rows`, `page`, `total`, `loading`, `container`, and
`pagination`. Calling `load()` fetches again, unless a load is already in
progress. The fetch uses `credentials: "same-origin"`; cross-origin requests
still require a CORS policy from the data server.

### Accepted data shapes

JSON can be a bare array, `{ "records": [...] }`, `{ "items": [...] }`, or a
single `{ "record": {...} }`. A pod-style row such as
`{ "id": "42", "fields": { "title": "Hello" } }` is normalized into a top-level row.
An `XML` response is also supported when it has this general shape:

```xml
<records>
  <record id="42">
    <field name="title">Hello</field>
    <field name="body">A record body</field>
  </record>
</records>
```

### Rendering and URL safety

Card content is built with DOM text nodes. Text values are HTML-escaped before
line breaks are added, so a record cannot inject markup through a normal field.

A value is rendered as an `<a>` only when it is a candidate for a link **and**
uses one of these schemes:

- `http:`
- `https:`
- `mailto:`
- `tel:`

The `link` field is a link candidate regardless of its position among the
selected fields; the final selected field is also a candidate. Candidates with
another scheme, a missing scheme, or a relative value are rendered as text,
not placed in an `href`. In particular, `javascript:`, `data:`, and
protocol-relative values do not become executable links.

The pager changes pages in memory and the **Export CSV** button downloads all
currently loaded rows. CSV is data, not a sanitized spreadsheet format; if it
will be opened by a spreadsheet application, treat formulas and other active
content as a separate consumer-side concern.

## Optional Go template demo

The Go program is a deliberately small demonstration of storing and serving
HTML snippets. It is not a required backend for the browser library.

### Run locally

From this directory (the module root):

```bash
go run .
```

The server listens on **127.0.0.1:8084** by default. Templates are stored as
JSON files in `templates_data/`, which is created on startup. The editor page
is served from `templates/index.html`. A non-loopback listener is refused at
startup unless `VIDI_API_TOKEN` is set; when set, the token is required for
all non-static routes (bearer or `X-Vidi-Token`).

### Routes

| Method | Route | Behavior |
| --- | --- | --- |
| `GET` | `/` | Visual editor and list of stored templates. |
| `GET` | `/<name>` | Return the stored HTML for a name. |
| `GET` | `/static/...` | Serve the demo's static assets. |
| `GET` | `/api/templates` | Return `var templates = [...]` for the editor/export UI. |
| `POST` | `/api/templates` | Save JSON `{ "name": "...", "html": "..." }`. |
| `GET` | `/api/templates/<name>` | Return one stored HTML document. |
| `DELETE` | `/api/templates/<name>` | Delete one stored template. |

POST requests are limited to 1 MiB and malformed JSON is rejected. Template
names are restricted to a single filename component made from ASCII letters,
digits, `.`, `_`, and `-` (maximum 128 characters). Both `/` and `\` are
rejected, including names supplied in the JSON body, so a name cannot traverse
outside `templates_data/`. Filesystem read, parse, delete, and write failures
are returned as server errors rather than being silently ignored.

### Docker

The image uses the Go version declared by `go.mod` (`1.21`) and exposes the
same port as the server:

```bash
docker build -t vidi-demo .
docker run --rm -p 127.0.0.1:8084:8084 \
  -e VIDI_API_TOKEN='replace-with-a-long-random-token' \
  -v "$PWD/templates_data:/app/templates_data" \
  vidi-demo
```

## Security limitations

The browser library is a rendering component, not an authorization or data
security boundary:

- It trusts the data source's authentication, authorization, availability, and
  CORS policy. It does not provide user accounts, sessions, encryption, secret
  storage, or audit logging.
- Escaping protects the generated card text, but it cannot make an arbitrary
  page, script, or trusted HTML fragment safe. Treat data and any separately
  injected HTML as untrusted according to the hosting application's policy.
- CSV values are not rewritten to remove spreadsheet formula syntax.

The Go demo is intended for local development. In the default loopback mode it
has no token and is not an authorization boundary. If `VIDI_API_TOKEN` is set,
all non-static routes require a constant-time bearer/`X-Vidi-Token` check; a
non-loopback bind without that token is refused. Stored HTML is still returned
verbatim and can execute in the server's origin, so keep the demo behind TLS,
request limits, and an isolated origin. It has no CSRF protection, per-user
authorization, or tenant isolation.

## Development checks

For a container or other non-loopback deployment, provide `VIDI_API_TOKEN` and
send it as `Authorization: Bearer …` (or `X-Vidi-Token`) on every non-static
request. The token is a demo guard, not a user/session system.

There is no npm build step for `vidi.js`. From this directory, run:

```bash
gofmt -w main.go main_test.go
go test ./...
node --check vidi.js
node --test vidi_test.js
```

The optional Go tests cover template persistence, route validation, malformed
input, traversal rejection, method handling, and write-failure reporting.

**License:** MIT License © Matthew Salvatore Giancola.
