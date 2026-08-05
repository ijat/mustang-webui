// Package appdir resolves where mustang-webui caches its provisioned JVM
// runtime and sidecar jar between runs.
package appdir

import (
	"os"
	"path/filepath"
)

const dirName = "mustang-webui"

// CacheDir returns the platform-appropriate cache directory for downloaded
// runtime assets (JRE, sidecar jar), creating it if necessary.
func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
