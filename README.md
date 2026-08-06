# mustang-webui

A local-first, single-binary web UI for [mustangproject](https://github.com/ZUGFeRD/mustangproject) — validate and inspect ZUGFeRD/Factur-X/XRechnung e-invoices (hybrid PDF/A-3 + embedded XML) without anything ever leaving your machine.

Run the binary, it opens a browser tab, you drop in a PDF, it tells you whether the PDF/A structure and the embedded invoice XML are valid against the real EN 16931/XRechnung/ZUGFeRD/Factur-X specs — and shows you the invoice data in a form a human can actually read, with the raw XML and the original PDF one click away.

## Status

The core loop works end to end: drop a PDF, it's validated by mustangproject and rendered in the UI. Release packaging (a single downloadable binary per platform, self-provisioning its own JVM) is wired up via CI — see [Releasing](#releasing) below.

## Architecture

```
mustang-webui (Go)  →  provisions + spawns  →  sidecar.jar (JVM, calls mustangproject directly)
       │
       └─ serves the frontend + reverse-proxies /api/* to the sidecar, on 127.0.0.1 only
```

- **`cmd/mustang-webui`, `internal/`** — the Go orchestrator: single binary, provisions a JVM + the sidecar jar on first run, spawns it, serves the frontend, proxies API calls.
- **`sidecar/`** — a thin Java wrapper (Maven) around mustangproject's `library` and `validator` modules. Calls their public API directly — no CLI subprocess, no reimplemented invoice parsing. `POST /api/inspect` is the one real endpoint so far.
- **`web/frontend/`** — the Svelte 5 + Tailwind + Motion frontend, built and embedded into the Go binary.

See `CLAUDE.md` for the full architecture, conventions, and non-negotiables.

## Development

Three terminals (see `Makefile`):

```sh
make build-sidecar && make dev-sidecar   # terminal 1: the JVM sidecar on :8765
make dev-frontend                        # terminal 2: Vite dev server with HMR
go run ./cmd/mustang-webui --dev         # terminal 3: orchestrator, opens the browser
```

`make build` produces a single `bin/mustang-webui` with the frontend embedded — but built this way, it still requires a JVM already on the machine (`--dev`-style resolution). A real self-provisioning binary — no JVM required upfront — only comes out of a tagged release; see below.

## Releasing

Push a tag matching `v*.*.*` (e.g. `v0.1.0`) and `.github/workflows/release.yml` takes it from there:

1. Runs the full test suite (Go, sidecar, frontend) as a gate — nothing below happens if it fails.
2. Builds the shaded sidecar jar.
3. Builds a minimal JRE per platform natively on that platform's own runner (`jdeps` auto-detects the exact JDK modules the jar needs, `jlink` produces the runtime, then the sidecar is actually started under it and health-checked before anything ships — a broken module set fails the build instead of shipping broken).
4. Assembles `internal/orchestrator/manifest.json` from the checksums of what was just built, pointing at this release's own GitHub asset URLs, and cross-compiles the Go binary for every platform with that manifest embedded via `go:embed`.
5. Publishes everything — the binaries, the JRE archives, the sidecar jar, and their `.sha256` files — as assets on a GitHub Release for that tag.

The committed `internal/orchestrator/manifest.json` is a placeholder (`{}`); only the CI-generated one (built fresh into the checkout before `go build`, never committed) makes a binary capable of the non-`--dev` self-provisioning path. Platforms covered: linux-amd64, darwin-amd64, darwin-arm64, windows-amd64.

## License

TBD — mustangproject itself is Apache-2.0; see `CLAUDE.md` for the licensing notes on bundled dependencies.
