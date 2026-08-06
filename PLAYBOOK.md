# Playbook

Status and handoff notes for whoever (human or agent) picks this project up
next. `README.md` is the pitch, `CLAUDE.md` is the architecture/conventions
bible; this file is "where things actually stand right now" and "what's
pending" — update it as state changes, don't let it go stale the way
`README.md`'s status line once did.

## What's built and verified

- **Go orchestrator** (`cmd/mustang-webui`, `internal/`): provisioning
  (download+checksum+extract), sidecar process management, the pretty
  stdout `Reporter` + `log/slog` diagnostic layer, loopback-only HTTP
  server + `/api` proxy. Tested (`go test ./...`, `-race` clean) and
  manually verified end to end, including a full dry run of the real
  (non-`--dev`) provisioning path — see "How the release manifest was
  verified" below.
- **Java sidecar** (`sidecar/`): `POST /api/inspect` calls mustangproject's
  `ZUGFeRDInvoiceImporter`/`ZUGFeRDValidator` directly and maps the result
  to JSON. Smoke-tested against real ZUGFeRD/XRechnung invoice fixtures
  from mustangproject's own test resources, plus edge cases (non-PDF
  input, corrupt/empty files, oversized uploads, missing/garbled auth).
- **Svelte frontend** (`web/frontend/`): app shell, dropzone, loading/error
  states, results workspace (validation rail + human-readable/raw-XML/
  PDF-preview tabs), dark-mode toggle — matches the finalized design
  tokens in `web/frontend/src/app.css`. Verified with `svelte-check`, a
  production build, and a real Playwright-driven run against the actual
  running app.
- **Release pipeline** (`.github/workflows/release.yml` — see below for
  why it's not actually at that path in git): builds the sidecar jar,
  jlinks a minimal JRE per platform natively on that platform's own
  runner (module set auto-detected via `jdeps`, smoke-tested by actually
  starting the sidecar under it before anything ships), assembles
  `internal/orchestrator/manifest.json` from that run's own checksums,
  cross-compiles the Go binary per platform with the manifest embedded.
  Platforms: linux-amd64, darwin-amd64, darwin-arm64, windows-amd64.
- **License**: `LICENSE` (Business Source License 1.1 — see its own
  "why this license" note below).
- **`main` branch**: exists on GitHub, currently identical to
  `claude/mustangproject-web-ui-zz2j2r`. GitHub's repo-level "default
  branch" setting still points at the `claude/...` branch — nobody has
  flipped that yet (see "Pending manual actions").

## How the release manifest was verified

Nobody has actually run `.github/workflows/release.yml` in real GitHub
Actions yet (it can't be pushed there — see below). Before writing it, the
whole mechanism was verified locally instead, since CI failures on this
particular pipeline would be expensive to debug blind:

1. Ran `jdeps --multi-release 21 --print-module-deps --ignore-missing-deps`
   against the real shaded sidecar jar to get the module list, `jlink`'d a
   runtime from it, and actually started the sidecar under that runtime
   against a real invoice fixture — confirms the auto-detected module set
   (plus `jdk.crypto.ec`/`jdk.localedata`, added defensively) is sufficient
   before trusting it in CI.
2. Packaged that JRE as a tar.gz the same way CI will (contents at archive
   root), served it plus the sidecar jar over a local `python3 -m
   http.server`, hand-wrote a `manifest.json` pointing at that local
   server, dropped it in at `internal/orchestrator/manifest.json`
   (overwriting the placeholder), rebuilt the Go binary, and ran it for
   real — first run (download/verify/extract/start), a cached second run,
   and a corrupted-checksum failure. All three matched the designed
   transcripts. This is what caught two real bugs (Windows needing
   `java.exe` not `java`, and a missing blank line between sibling asset
   blocks in the CLI output) — see git log around commit `77efa9a` for the
   detail.
3. `internal/orchestrator/manifest.json` was restored to its placeholder
   (`{}`) afterward. Never commit a real manifest to that path — it's
   meant to be generated fresh into the checkout by CI immediately before
   `go build`, not checked in.

None of this is a substitute for actually watching the workflow run once
it's live — the GitHub-hosted-runner-specific bits (exact `jlink`/`jdeps`
behavior per OS, `actions/upload-artifact` matrix aggregation, the
`softprops/action-gh-release` step) were validated by reading, not by
execution. First real tag push should be treated as the true first test.

## Pending manual actions

### 1. ~~Promote `.github/release.yml` to `.github/workflows/release.yml`~~ — done

Done 2026-08-06: a later session's push credentials did carry the
`workflow` OAuth scope, so `git mv .github/release.yml
.github/workflows/release.yml` + `git push` went through cleanly
(commit `b33ea43`). `gh workflow list` now shows `Release` as `active`.
The scope restriction noted below was real for the session that hit it,
just not a durable constraint — check current credentials before
assuming it still applies.

### 2. ~~Set `main` as the GitHub repo's default branch~~ — done

Done 2026-08-06 via `gh repo edit --default-branch main`. The
`claude/mustangproject-web-ui-zz2j2r` branch (which was both the
default branch and identical to `main`) has since been deleted, both
locally and on `origin`.

### 3. First real release

Now unblocked: push a tag matching `v*.*.*` and watch it actually run —
see the "How the release manifest was verified" caveat above about what
was and wasn't exercised for real. Still nobody has watched a live run,
so treat the first tag push as the true first test.

## Other loose ends worth knowing about

- **`part: "fx"`/`"ox"`** in `/api/inspect` findings (Factur-X/Order-X) are
  inferred by the frontend from mustangproject's own naming convention,
  not spelled out anywhere the sidecar author and frontend author both
  read — see `web/frontend/src/lib/findings.ts`'s comment. Only `pdf`,
  `xr` are documented in the original contract. Worth confirming against
  real Order-X output if that ever comes up; low risk either way since
  it's just a display label, not something the validation logic depends
  on.
- **In-app "licenses" page**: `CLAUDE.md`'s Licensing section calls this
  out as a pre-public-release TODO (carrying forward mustangproject's
  `NOTICE` plus the bundled-dependency attributions). Not built yet.
- **`--offline` release mode**: `CLAUDE.md` non-negotiable #2 mentions
  this as a future addition (pre-provisioned assets, for fully airgapped
  installs). Not built — `--dev` is currently the only way to skip the
  download path.
- **jlink runner labels** (`macos-15-intel`, `macos-14` in the release
  workflow's matrix): GitHub's hosted-runner label list changes over time
  (an earlier draft of this workflow used `macos-13`, which `actionlint`
  caught as already retired). If a future release run fails with an
  "unknown runner label" type error, that's almost certainly what
  happened again — check `actionlint`'s current label list or GitHub's
  own runner-images docs for the current names.
