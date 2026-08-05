package orchestrator

import (
	"errors"
	"runtime"
)

// Asset is a downloadable, checksum-verified release artifact.
type Asset struct {
	URL    string
	SHA256 string
}

// Manifest pins the exact runtime and sidecar jar assets for a release.
// It is populated per-release (see release tooling under /sidecar and CI)
// and points at this project's own GitHub Release assets — a jlink'd
// minimal JRE per platform plus the shaded sidecar jar built from
// /sidecar. Nothing is fetched from Maven Central or Adoptium at runtime;
// the release process resolves and re-hosts pinned, checksummed copies so
// end users only ever talk to one origin.
type Manifest struct {
	// Runtime maps "GOOS-GOARCH" (e.g. "linux-amd64") to a jlink runtime archive.
	Runtime map[string]Asset
	Sidecar Asset
}

// releaseManifest is nil until wired up by the release build (expected via
// go:generate or -ldflags at build time). Until then, ProvisionRuntime
// only supports --dev mode.
var releaseManifest *Manifest

// ErrManifestUnconfigured is returned by ProvisionRuntime when this build
// has no pinned release assets to download, i.e. any build that isn't a
// tagged release. Use --dev instead.
var ErrManifestUnconfigured = errors.New("orchestrator: no release manifest in this build; run with --dev")

func platformKey() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}
