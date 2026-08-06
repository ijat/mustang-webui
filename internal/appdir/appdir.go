// Package appdir resolves where mustang-webui stores its provisioned JVM
// runtime and sidecar jar between runs.
package appdir

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirName = "mustang-webui-data"

// CacheDir returns the directory where mustang-webui stores everything it
// downloads (JRE, sidecar jar) and every intermediate it creates while
// doing so (partial downloads, extraction staging), creating it if
// necessary. This is deliberately a subfolder next to the running binary,
// not an OS-wide cache dir: the app is meant to be portable — nothing it
// writes ever leaves its own directory, so moving or deleting that
// directory takes the binary and everything it provisioned with it.
func CacheDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}

	dir := filepath.Join(filepath.Dir(exe), dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}
	return dir, nil
}
