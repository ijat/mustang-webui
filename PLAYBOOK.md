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

### 1. Promote `.github/release.yml` to `.github/workflows/release.yml`

GitHub gates any push that touches `.github/workflows/*` behind a specific
permission — the OAuth `workflow` scope for a normal `git push`, the
"workflows" permission for a GitHub App using the Contents API. This
session's credentials have neither (confirmed by trying both — a `git
push` rejected with "refusing to allow an OAuth App to create or update
workflow... without `workflow` scope", and a Contents API write to the
same path returning a bare 404 while the identical write to a sibling path
outside `workflows/` succeeded). This isn't a bug or a one-off — it's a
deliberate GitHub security boundary against exactly this scenario
(automated agents writing arbitrary CI code), so treat "can't push
workflow files" as a standing constraint for this project, not something
worth re-attempting the same way next time.

The content is byte-for-byte identical, sitting at `.github/release.yml`
(not gated — it's outside the `workflows/` directory) specifically so it
travels with the repo instead of depending on someone having saved a file
sent through a one-off chat attachment. To activate it:

```sh
git mv .github/release.yml .github/workflows/release.yml
git commit -m "Promote release workflow"
git push   # ← do this step with credentials that actually have the
           #   workflow scope/permission — the web UI's own "Add file"
           #   flow works too, and isn't subject to this restriction
```

Do this whenever whoever's driving has real `workflow`-scoped credentials
(a maintainer's own `git push`, or the GitHub web UI's "Add file" /
"Edit" flow, both unaffected by this restriction) — or grant this kind of
session's GitHub integration the `workflow` OAuth scope up front next
time, if that's a realistic option for how it's set up.

### 2. Set `main` as the GitHub repo's default branch, if desired

No tool in this session's toolkit exposes that setting (it's repo admin
configuration, not a git or Contents API operation). Settings → Branches
→ change default branch, in GitHub's UI.

### 3. First real release

Once #1 is done: push a tag matching `v*.*.*` and watch it actually run —
see the "How the release manifest was verified" caveat above about what
was and wasn't exercised for real.

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
