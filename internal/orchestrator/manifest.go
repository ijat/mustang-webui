package orchestrator

import (
	_ "embed"
	"encoding/json"
	"errors"
	"runtime"
)

// Asset is a downloadable, checksum-verified release artifact.
type Asset struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// Manifest pins the exact runtime and sidecar jar assets for a release.
// It points at this project's own GitHub Release assets — a jlink'd
// minimal JRE per platform plus the shaded sidecar jar built from
// /sidecar. Nothing is fetched from Maven Central or Adoptium at runtime;
// the release process resolves and re-hosts pinned, checksummed copies so
// end users only ever talk to one origin.
type Manifest struct {
	// Runtime maps "GOOS-GOARCH" (e.g. "linux-amd64") to a jlink runtime archive.
	Runtime map[string]Asset `json:"runtime"`
	Sidecar Asset            `json:"sidecar"`
}

// manifestJSON embeds internal/orchestrator/manifest.json as it exists at
// build time. The file committed to the repo is a placeholder ("{}") — a
// tagged-release CI build overwrites it with the real, computed asset
// URLs and checksums for that release *before* running `go build`, so
// go:embed picks up the real content. Any build that isn't a release
// build (a dev checkout, a plain `go build`, a PR build) embeds the
// placeholder and ends up with releaseManifest == nil, same as before —
// --dev is the only way to run such a build.
//
//go:embed manifest.json
var manifestJSON []byte

var releaseManifest = parseManifest(manifestJSON)

// parseManifest returns nil for invalid or placeholder ("{}", no runtime
// assets) JSON, rather than an error — an unconfigured manifest is a
// normal, expected state for any non-release build.
func parseManifest(data []byte) *Manifest {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil || len(m.Runtime) == 0 {
		return nil
	}
	return &m
}

// ErrManifestUnconfigured is returned by ProvisionRuntime when this build
// has no pinned release assets to download, i.e. any build that isn't a
// tagged release. Use --dev instead.
var ErrManifestUnconfigured = errors.New("orchestrator: no release manifest in this build; run with --dev")

func platformKey() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

// javaExecutableName is "java" everywhere except Windows.
func javaExecutableName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}
