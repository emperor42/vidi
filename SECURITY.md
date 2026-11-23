# VIDI security notes

## Scope

VIDI has a browser rendering library (`vidi.js`) and an optional Go template
demo. They are separate trust surfaces. Neither implements accounts,
authorization, a multi-tenant store, or a production CMS.

## Browser library

`vidi.js` renders fetched records into the page.

- Every value is written with `textContent`; the renderer does not use
  `innerHTML` for record data.
- A value is turned into a link only when its scheme is `http`, `https`,
  `mailto`, or `tel`. This limits `javascript:` URLs but is not a general URL
  or content-safety policy.
- XML is parsed with `DOMParser` and JSON is decoded directly. No code from a
  record is executed, and no remote script is loaded.
- `dataSource` is chosen by the embedding application. The library performs no
  origin checks, authentication, CORS handling, or TLS validation of its own.

A hostile page can still override globals, replace the DOM, or intercept
`fetch`. The library must not be treated as a sandbox for untrusted pages.

## Go template demo

The demo binds to `127.0.0.1:8084` by default and serves a small template
editor. Startup refuses a non-loopback bind unless `VIDI_API_TOKEN` is set.
When the token is set, all routes except `/static/` require a constant-time
bearer or `X-Vidi-Token` check.

That token is a single shared demo credential. It is not a user account,
per-user authorization, CSRF protection, or rate limit. Anyone who holds it can
list, read, create, and delete every template. Keep the process on loopback or
behind a separately reviewed authentication and network policy. There is no TLS
termination in the demo.

The handler applies these guardrails:

- method checks and `Allow` headers on API routes;
- template names restricted to a single conservative filename component, with
  both slash styles rejected, plus an absolute-path check before file access;
- a 1 MiB request body limit and rejection of trailing JSON values;
- template writes through a temporary file with mode `0600`, a sync, and an
  atomic rename;
- `X-Content-Type-Options`, `X-Frame-Options`, and `Referrer-Policy` headers.

**Stored templates are executable HTML.** The API saves a template body
verbatim and `/{name}` serves it as `text/html` without sanitization. A caller
with write access can store script, inline event handlers, or markup that runs
with the origin's privileges. The storage directory is not an authorization
boundary. Do not share this origin with sensitive applications.

Templates are plaintext JSON on local disk. The demo does not encrypt them,
replicate them, back them up, or protect them from a process that can read the
data directory. It has no audit log; request outcomes go only to the process
log. It also has no CORS policy, DDoS mitigation, MFA, SSO, roles, password
hashing, or secrets management. None of those features are implemented.

The Dockerfile sets `VIDI_HOST=0.0.0.0` so a published port works in a
container. Set `VIDI_API_TOKEN` at runtime and publish the port to host
loopback only.

## Reporting a suspected vulnerability

Report suspected vulnerabilities privately to the project maintainers. Include
the affected file, a minimal reproduction, and the deployment assumptions,
especially whether a token was set and whether the listener was non-loopback.
