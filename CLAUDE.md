# mustang-webui

A local-first, single-binary web UI for [mustangproject](https://github.com/ZUGFeRD/mustangproject): validate and inspect ZUGFeRD/Factur-X/XRechnung e-invoices (hybrid PDF/A-3 + embedded XML) entirely on the user's machine. No invoice data, no PDF content, and no extracted XML ever leaves the device.

## Non-negotiables

These are the constraints the whole design serves. Any change that violates one of these needs a real conversation first, not a quiet workaround:

1. **mustangproject is the source of truth.** We never reimplement CII/UBL parsing, PDF/A validation, or Schematron rule logic. The Java sidecar calls mustangproject's public library API directly and does nothing more than translate the result to JSON.
2. **Everything after setup is offline.** The orchestrator and the sidecar bind to `127.0.0.1` only — never `0.0.0.0`, never a LAN-visible address. The only network call in the entire app is the one-time asset download on first run (see below); disable that path (`--dev`, or a future `--offline` release mode with pre-provisioned assets) and the app makes zero network calls.
3. **Single binary, self-provisioning.** The user runs one executable. It downloads what it needs (a minimal JRE + the sidecar jar, from this project's own release assets — never Maven Central or Adoptium directly at runtime, so there's one origin to trust) into a subfolder next to the binary itself, and every subsequent run is instant. A JVM dependency is unavoidable given (1); "single binary" means single *thing the user runs*, not literally zero runtime dependencies. Don't try to make this literal by vendoring a fat 150MB+ binary unless we explicitly decide that trade-off is worth it.
4. **No Claude/Anthropic visual identity.** This product has its own design language. Don't reach for Claude's oranges/creams or its typefaces out of habit.
5. **Portable, not OS-integrated.** Everything the app writes — the provisioned runtime, partial downloads, extraction staging — stays inside `<binary dir>/mustang-webui-data/`. Never `os.UserCacheDir()`, never `os.TempDir()`, never anywhere outside the binary's own directory. Moving or deleting that one folder should take the whole install with it, with nothing left behind elsewhere on the machine.

## Architecture

```
mustang-webui (Go, single binary)
  │
  ├─ on first run: downloads + checksums a jlink JRE and sidecar.jar into
  │  <binary dir>/mustang-webui-data/, from this project's own release assets
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
.github/workflows/release.yml  tag-triggered release pipeline (see "Releasing" in README.md)
cmd/mustang-webui/    Go entry point — flag parsing, lifecycle, signal handling
internal/appdir/      resolves <binary dir>/mustang-webui-data/ (never an OS-wide cache dir)
internal/orchestrator/ runtime provisioning (download+verify+extract), sidecar process management,
                       manifest.go/manifest.json (placeholder in the repo; a release build overwrites
                       manifest.json with real asset URLs+checksums before go:embed picks it up)
internal/server/       HTTP server: static frontend + /api reverse proxy, loopback only
web/embed.go            go:embed of web/dist into the binary
web/dist/               built frontend — a generated artifact, but committed anyway so `go build` works
                        without a Node toolchain present; `make build-frontend` regenerates it
web/frontend/           Svelte source (Vite project)
sidecar/                Java wrapper around mustangproject (Maven)
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
- **Two output layers, never merged.** `internal/orchestrator/progress.go`'s `Reporter` is the pretty, human-facing stdout output for the default run (section headers, per-asset progress with the exact source URL always visible, checkmarks) — it's what a non-technical user sees. `log/slog` is the structured stderr layer, quiet by default (`Warn`+) and switched to `Debug` only by `--verbose`. Don't let a `slog.Info` call leak routine chatter into the default run, and don't hand-roll `fmt.Println` diagnostics that should be `slog` instead — each layer has exactly one job.
- Run `gofmt -l .`, `go vet ./...`, and `go test ./...` clean before considering any Go change done.

## Java sidecar conventions

- **This module does not grow business logic.** If you find yourself writing invoice-field-parsing or PDF-structure logic here instead of calling into mustangproject, stop — that logic belongs upstream in mustangproject, or the sidecar is doing something wrong. The sidecar's only allowed jobs are: (a) call mustang's public API, (b) shape the result into JSON, (c) enforce the bearer-token check.
- **No web framework.** `com.sun.net.httpserver.HttpServer` is enough for a single-user loopback API with a handful of endpoints. Don't add Spring/Javalin/etc. for this.
- **Jackson (`jackson-databind`, already a dependency) for all JSON** — don't hand-roll JSON string building beyond the trivial fixed strings in health/error responses.
- **Every non-`/healthz` route must go through the bearer-token filter.** New endpoints get added under `/api`, which already carries the filter — don't create new top-level contexts that bypass it.
- **`POST /api/inspect`** (`InspectHandler.java`) is the reference example for new endpoints: raw PDF bytes as the request body (not multipart — there's only ever one file and no other fields, and multipart parsing isn't worth a dependency), filename via an `X-Filename` header, a hand-built DTO layer mapping mustang's `Invoice`/`CalculatedInvoice`/`ValidationContext` objects to our own JSON shape (never serialize mustang's classes directly — their shape is theirs to change). A 50MB body cap is enforced before buffering. `invoice`/`rawXml` being `null` in the response is a normal, valid outcome (a PDF with no embedded e-invoice XML), not an error — don't conflate "no embedded invoice" with "request failed."
- **Never let this process become network-reachable beyond loopback**, and never let it make outbound network calls — mustangproject's own XSD/Schematron resolution is already fully offline (classloader resources, not HTTP); keep it that way when adding endpoints.
- Pin mustang and Jackson versions explicitly in `pom.xml`; don't float on `LATEST`/`RELEASE`.
- `maven-shade-plugin` filters must keep excluding `META-INF/*.SF|.DSA|.RSA` — without it the shaded jar fails signature verification at runtime (bit us once already; see git history if it recurs).

## Frontend conventions

Stack: **Svelte 5 (runes) + Tailwind v4 + Motion (motion.dev) + GSAP for complex sequences.** Svelte's compiled, no-vdom output means less main-thread work competing with animation frames — that's the whole reason it's the pick over React/Solid for a "60fps everywhere" requirement.

**The design system** (finalized; tokens live in `web/frontend/src/app.css` as CSS custom properties, wired into Tailwind's `@theme inline`, so both plain CSS and utility classes read from the same source):

- **Color.** Plain white (`#ffffff`) is the ground, not a warm/cream tint — deliberately Swiss/International-Typographic-Style rather than an "AI-default" cream+serif look. Ink `#191919`, muted `#5c5c5c`, faint `#8c8c8c`, hairline border `#dcdcdc`. One accent, spent sparingly (active tab, focus rings, links): a desaturated "seal" blue `#1e4b7a`. Semantic severity color (`success`/`warning`/`critical`) is a separate set of tokens from the accent — never reuse the accent hue for validation state.
- **Dark mode is opt-in only** (`lib/theme.svelte.ts`), toggled in the titlebar, persisted to `localStorage`, applied via `data-theme="dark"` scoped to the app shell — **never** `prefers-color-scheme`. Light is what the app opens to, always, regardless of OS setting.
- **Typography.** Arimo (400/500/700, self-hosted via `@fontsource/arimo`) for everything — display, body, UI chrome — a free, metric-compatible Helvetica/Arial alternative, not Inter/Space Grotesk (both flagged as generic "AI-default" faces). Cousine (400, `@fontsource/cousine`) is reserved *only* for the raw-XML view — don't let a monospace face leak into labels/chips/wordmarks the way an early draft of this design did.
- Tabular figures (`font-variant-numeric: tabular-nums`, `.tabular-nums` utility) for anything numeric that aligns in a column — invoice totals, validation counts.
- **Glass chrome over an ambient background.** The titlebar/rail/content panels use `backdrop-filter: blur(...) saturate(...)` over a `color-mix()`-based translucent background, layered above a handful of large, slowly-drifting blurred color fields (`transform`-only `@keyframes`, disabled under `prefers-reduced-motion`). The translucency is what proves the blur is real — don't make a "glass" panel that's actually opaque.

**Motion**

- Animate `transform` and `opacity` only for anything that needs to hit 60fps — never animate `width`/`height`/`top`/`left`/box-shadow spread on interactive/frequent transitions, they force layout/paint.
- Respect `prefers-reduced-motion` — this is not optional polish, wire it in from the first animated component, not retrofitted later.
- Motion should communicate state change (this validated / this failed / this is loading), not decorate for its own sake. If an animation doesn't help the user understand what just happened, cut it.
- Spring-based transitions (Motion's default) for anything user-triggered and interruptible (panel open/close, hover states); timeline-based (GSAP) for fixed multi-step sequences (e.g. a validation-progress choreography) where you need precise sequencing.

**General UI**

- This is a document-inspection tool, not a marketing site — information density and scan-ability matter more than spectacle. Animate the chrome (transitions, feedback, loading states), not the content the user is trying to read.
- Every validation result needs both the human-readable view and one click to raw data (XML/JSON) — never make raw data harder to reach than the pretty view, that's a trust issue for a tool whose whole job is telling people whether their document is correct.
- Accessible by default: semantic HTML, visible focus states, keyboard-operable everything (this will be used by accountants and back-office staff, not just developers).

## Clean code, generally

- No abstraction ahead of a second real use case. Three similar lines beat a premature helper.
- No comments explaining *what* code does — name things well instead. A comment is only earned by a non-obvious *why* (see the shade-plugin note above for an example of the kind of thing that's worth one).
- No half-finished endpoints/features pretending to work — a route that isn't implemented yet returns a real `501`, not a hardcoded fake success response.
- No unused abstractions "for later" — the provisioning/download machinery in `internal/orchestrator` exists now because release packaging genuinely needs it soon, not as speculative infrastructure.

## Licensing

mustangproject core is Apache-2.0. Bundled/transitive pieces we redistribute: PDFBox (Apache-2.0), veraPDF (MPL 2.0), Saxon-HE (MPL 2.0), ph-schematron (Apache-2.0), DOM4j (BSD), Apache FOP (Apache-2.0) — all permissive. **The EN16931 Schematron validation artefacts are EUPL-2.0** (the one non-Apache/BSD/MPL license in the tree) — permits redistribution, just needs its own attribution. The bundled JRE (Adoptium/OpenJDK builds) is GPLv2+Classpath-Exception, which explicitly permits redistributing a custom `jlink` runtime image. Before a public release: ship an in-app "licenses" page carrying forward mustangproject's `NOTICE` file plus these attributions.

## Build & dev commands

See `README.md` for the day-to-day dev loop and `Makefile` for the exact commands (`make build-sidecar`, `make dev-frontend`, `make build`, `make test`).
