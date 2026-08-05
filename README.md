# mustang-webui

A local-first, single-binary web UI for [mustangproject](https://github.com/ZUGFeRD/mustangproject) — validate and inspect ZUGFeRD/Factur-X/XRechnung e-invoices (hybrid PDF/A-3 + embedded XML) without anything ever leaving your machine.

Run the binary, it opens a browser tab, you drop in a PDF, it tells you whether the PDF/A and the embedded invoice XML are valid — and shows you the invoice data in a form a human can actually read, with the raw XML one click away.

## Status

Early scaffold. Architecture is in place and verified end to end; the actual validate/extract UI doesn't exist yet. See `CLAUDE.md` for the full design.

## Architecture

```
mustang-webui (Go)  →  provisions + spawns  →  sidecar.jar (JVM, calls mustangproject directly)
       │
       └─ serves the frontend + reverse-proxies /api/* to the sidecar, on 127.0.0.1 only
```

- **`cmd/mustang-webui`, `internal/`** — the Go orchestrator: single binary, provisions a JVM + the sidecar jar on first run, spawns it, serves the frontend, proxies API calls.
- **`sidecar/`** — a thin Java wrapper (Maven) around mustangproject's `library` and `validator` modules. Calls their public API directly — no CLI subprocess, no reimplemented invoice parsing.
- **`web/frontend/`** — the Svelte 5 + Tailwind + Motion frontend, built and embedded into the Go binary.

## Development

Three terminals (see `Makefile`):

```sh
make build-sidecar && make dev-sidecar   # terminal 1: the JVM sidecar on :8765
make dev-frontend                        # terminal 2: Vite dev server with HMR
go run ./cmd/mustang-webui --dev         # terminal 3: orchestrator, opens the browser
```

`make build` produces a single `bin/mustang-webui` with the frontend embedded (still requires a JVM on the machine — see `CLAUDE.md` for why, and the plan for a self-provisioning release build).

## License

TBD — mustangproject itself is Apache-2.0; see `CLAUDE.md` for the licensing notes on bundled dependencies.
