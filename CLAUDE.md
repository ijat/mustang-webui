# mustang-webui

A local-first, single-binary web UI for [mustangproject](https://github.com/ZUGFeRD/mustangproject): validate and inspect ZUGFeRD/Factur-X/XRechnung e-invoices (hybrid PDF/A-3 + embedded XML) entirely on the user's machine. No invoice data, no PDF content, and no extracted XML ever leaves the device.

## Non-negotiables

These are the constraints the whole design serves. Any change that violates one of these needs a real conversation first, not a quiet workaround:

1. **mustangproject is the source of truth.** We never reimplement CII/UBL parsing, PDF/A validation, or Schematron rule logic. The Java sidecar calls mustangproject's public library API directly and does nothing more than translate the result to JSON.
2. **Everything after setup is offline.** The orchestrator and the sidecar bind to `127.0.0.1` only — never `0.0.0.0`, never a LAN-visible address. The only network call in the entire app is the one-time asset download on first run (see below); disable that path (`--dev`, or a future `--offline` release mode with pre-provisioned assets) and the app makes zero network calls.
3. **Single binary, self-provisioning.** The user runs one executable. It downloads what it needs (a minimal JRE + the sidecar jar, from this project's own release assets — never Maven Central or Adoptium directly at runtime, so there's one origin to trust) into a per-user cache, and every subsequent run is instant. A JVM dependency is unavoidable given (1); "single binary" means single *thing the user runs*, not literally zero runtime dependencies. Don't try to make this literal by vendoring a fat 150MB+ binary unless we explicitly decide that trade-off is worth it.
4. **No Claude/Anthropic visual identity.** This product has its own design language. Don't reach for Claude's oranges/creams or its typefaces out of habit.

## Architecture

```
mustang-webui (Go, single binary)
  │
  ├─ on first run: downloads + checksums a jlink JRE and sidecar.jar into
  │  the user cache dir, from this project's own release assets
  │
  ├─ spawns sidecar.jar as a subprocess (127.0.0.1, random port, random
  │  per-launch bearer token) and waits for /healthz
  │
  └─ serves the frontend + reverse-proxies /api/* to the sidecar
     (injecting the bearer token — the browser never sees it)
              │
              ▼
     opens the system browser to http://127.0.0.1:<port>
```

The sidecar is a thin adapter, not a reimplementation: it calls `ZUGFeRDInvoiceImporter`, `ZUGFeRDValidator`, `ZUGFeRDVisualizer` etc. from mustangproject's `library`/`validator` modules directly (in-process, not via mustang's CLI — the CLI's text/XML stdout is not something we parse) and serializes the results with Jackson.

## Repository layout

```
cmd/mustang-webui/    Go entry point — flag parsing, lifecycle, signal handling
internal/appdir/      per-user cache dir resolution
internal/orchestrator/ runtime provisioning (download+verify+extract) and sidecar process management
internal/server/      HTTP server: static frontend + /api reverse proxy, loopback only
web/embed.go          go:embed of web/dist into the binary
web/dist/             built frontend (committed — see "On committing web/dist" below)
web/frontend/         Svelte source (Vite project)
sidecar/              Java wrapper around mustangproject (Maven)
```

## Go orchestrator conventions

- **Stdlib first.** This component's job is I/O glue (download, spawn, proxy, serve static files) — it does not need a web framework, a DI container, or a config library. `net/http`, `flag`, `log/slog`, `os/exec` are enough. Reach for a dependency only when the stdlib genuinely can't do it.
- **No global mutable state.** Everything flows through explicit values (`RuntimePaths`, `*Sidecar`, `Options`) passed to constructors — no package-level vars holding live state, no init-time side effects beyond the `releaseManifest` data itself.
- **`context.Context` everywhere it matters.** Provisioning, downloads, and process startup all take a `ctx` and respect cancellation — the app must shut down cleanly on Ctrl-C without leaving an orphaned JVM running.
- **Errors: wrap, don't swallow.** `fmt.Errorf("doing X: %w", err)` at every layer boundary so failures are traceable from the top-level log line. No `_ = err` outside of genuinely-don't-care cleanup paths (and even then, prefer logging it).
- **`log/slog` for structured logging**, not `fmt.Println`/`log.Printf`. Key-value pairs, not string interpolation.
- **Table-driven tests with `net/http/httptest`.** No mocking framework — the interfaces here (an `http.Handler`, a `*Sidecar`) are small enough that real implementations or trivial fakes are simpler than a mock.
- **Every download is checksum-verified before use** (`internal/orchestrator/runtime.go`). Never relax this to "trust the URL."
- **Every archive extraction guards against zip-slip/path escape** (see `isWithinDir`). This is not optional hardening — it's handling untrusted (network-delivered) input.
- Run `gofmt -l .`, `go vet ./...`, and `go test ./...` clean before considering any Go change done.

## Java sidecar conventions

- **This module does not grow business logic.** If you find yourself writing invoice-field-parsing or PDF-structure logic here instead of calling into mustangproject, stop — that logic belongs upstream in mustangproject, or the sidecar is doing something wrong. The sidecar's only allowed jobs are: (a) call mustang's public API, (b) shape the result into JSON, (c) enforce the bearer-token check.
- **No web framework.** `com.sun.net.httpserver.HttpServer` is enough for a single-user loopback API with a handful of endpoints. Don't add Spring/Javalin/etc. for this.
- **Jackson (`jackson-databind`, already a dependency) for all JSON** — don't hand-roll JSON string building beyond the trivial fixed strings in health/error responses.
- **Every non-`/healthz` route must go through the bearer-token filter.** New endpoints get added under `/api`, which already carries the filter — don't create new top-level contexts that bypass it.
- **Never let this process become network-reachable beyond loopback**, and never let it make outbound network calls — mustangproject's own XSD/Schematron resolution is already fully offline (classloader resources, not HTTP); keep it that way when adding endpoints.
- Pin mustang and Jackson versions explicitly in `pom.xml`; don't float on `LATEST`/`RELEASE`.
- `maven-shade-plugin` filters must keep excluding `META-INF/*.SF|.DSA|.RSA` — without it the shaded jar fails signature verification at runtime (bit us once already; see git history if it recurs).

## Frontend conventions

Stack: **Svelte 5 (runes) + Tailwind v4 + Motion (motion.dev) + GSAP for complex sequences.** Svelte's compiled, no-vdom output means less main-thread work competing with animation frames — that's the whole reason it's the pick over React/Solid for a "60fps everywhere" requirement.

**Typography**

- Pick one text typeface and one (optional) display typeface, both variable fonts if available — fewer HTTP requests, smoother weight transitions for animated headings. Self-host or bundle; don't pull from Google Fonts at runtime (violates offline-first, and this app has no internet after setup anyway).
- Use a proper modular type scale (e.g. a 1.2–1.25 ratio), not ad hoc pixel values. Define it once as Tailwind theme tokens, reference tokens everywhere.
- Tabular/monospaced figures for anything numeric that updates or aligns in a column (invoice amounts, validation counts) — default proportional figures make columns of numbers visually jitter.
- Body text: real line-height (1.5–1.6 for body copy), a measure (line length) capped around 60–75 characters for prose, not stretched full-width.

**Motion**

- Animate `transform` and `opacity` only for anything that needs to hit 60fps — never animate `width`/`height`/`top`/`left`/box-shadow spread on interactive/frequent transitions, they force layout/paint.
- Respect `prefers-reduced-motion` — this is not optional polish, wire it in from the first animated component, not retrofitted later.
- Motion should communicate state change (this validated / this failed / this is loading), not decorate for its own sake. If an animation doesn't help the user understand what just happened, cut it.
- Spring-based transitions (Motion's default) for anything user-triggered and interruptible (panel open/close, hover states); timeline-based (GSAP) for fixed multi-step sequences (e.g. a validation-progress choreography) where you need precise sequencing.

**General UI**

- This is a document-inspection tool, not a marketing site — information density and scan-ability matter more than spectacle. Animate the chrome (transitions, feedback, loading states), not the content the user is trying to read.
- Every validation result needs both the human-readable view and one click to raw data (XML/JSON) — never make raw data harder to reach than the pretty view, that's a trust issue for a tool whose whole job is telling people whether their document is correct.
- Accessible by default: semantic HTML, visible focus states, keyboard-operable everything (this will be used by accountants and back-office staff, not just developers).

*(A concrete palette, typeface choices, and layout system are a separate design proposal — see the discussion following this scaffold. Once agreed, they get documented here as the canonical design system rather than living only in code.)*

## Clean code, generally

- No abstraction ahead of a second real use case. Three similar lines beat a premature helper.
- No comments explaining *what* code does — name things well instead. A comment is only earned by a non-obvious *why* (see the shade-plugin note above for an example of the kind of thing that's worth one).
- No half-finished endpoints/features pretending to work — a route that isn't implemented yet returns a real `501`, not a hardcoded fake success response.
- No unused abstractions "for later" — the provisioning/download machinery in `internal/orchestrator` exists now because release packaging genuinely needs it soon, not as speculative infrastructure.

## Licensing

mustangproject core is Apache-2.0. Bundled/transitive pieces we redistribute: PDFBox (Apache-2.0), veraPDF (MPL 2.0), Saxon-HE (MPL 2.0), ph-schematron (Apache-2.0), DOM4j (BSD), Apache FOP (Apache-2.0) — all permissive. **The EN16931 Schematron validation artefacts are EUPL-2.0** (the one non-Apache/BSD/MPL license in the tree) — permits redistribution, just needs its own attribution. The bundled JRE (Adoptium/OpenJDK builds) is GPLv2+Classpath-Exception, which explicitly permits redistributing a custom `jlink` runtime image. Before a public release: ship an in-app "licenses" page carrying forward mustangproject's `NOTICE` file plus these attributions.

## Build & dev commands

See `README.md` for the day-to-day dev loop and `Makefile` for the exact commands (`make build-sidecar`, `make dev-frontend`, `make build`, `make test`).
