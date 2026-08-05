package orchestrator

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RuntimePaths locates a ready-to-run JVM and sidecar jar.
type RuntimePaths struct {
	JavaBin    string
	SidecarJar string
}

// ProvisionRuntime resolves a JavaBin + SidecarJar pair, downloading and
// caching them under cacheDir on first use. In dev mode it uses the
// system `java` on PATH and a locally built sidecar jar instead, so the
// project is runnable without a cut release.
func ProvisionRuntime(ctx context.Context, cacheDir string, dev bool) (*RuntimePaths, error) {
	if dev {
		return devRuntime()
	}

	if releaseManifest == nil {
		return nil, ErrManifestUnconfigured
	}

	asset, ok := releaseManifest.Runtime[platformKey()]
	if !ok {
		return nil, fmt.Errorf("orchestrator: no runtime asset for platform %s", platformKey())
	}

	javaBin, err := ensureRuntimeArchive(ctx, cacheDir, asset)
	if err != nil {
		return nil, fmt.Errorf("provisioning JVM: %w", err)
	}

	jarPath, err := ensureFile(ctx, cacheDir, "sidecar.jar", releaseManifest.Sidecar)
	if err != nil {
		return nil, fmt.Errorf("provisioning sidecar jar: %w", err)
	}

	return &RuntimePaths{JavaBin: javaBin, SidecarJar: jarPath}, nil
}

func devRuntime() (*RuntimePaths, error) {
	javaBin, err := exec.LookPath("java")
	if err != nil {
		return nil, fmt.Errorf("dev mode requires a JDK on PATH: %w", err)
	}

	matches, err := filepath.Glob("sidecar/target/sidecar-*.jar")
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("dev mode requires a built sidecar jar; run `make build-sidecar` first")
	}

	return &RuntimePaths{JavaBin: javaBin, SidecarJar: matches[len(matches)-1]}, nil
}

// ensureRuntimeArchive downloads and extracts a gzipped tar JRE archive
// into cacheDir/runtime if not already present, returning the java binary
// path inside it.
func ensureRuntimeArchive(ctx context.Context, cacheDir string, asset Asset) (string, error) {
	runtimeDir := filepath.Join(cacheDir, "runtime")
	javaBin := filepath.Join(runtimeDir, "bin", "java")

	if _, err := os.Stat(javaBin); err == nil {
		return javaBin, nil
	}

	archivePath := filepath.Join(cacheDir, "runtime.tar.gz")
	if err := download(ctx, asset, archivePath); err != nil {
		return "", err
	}
	defer os.Remove(archivePath)

	if err := extractTarGz(archivePath, runtimeDir); err != nil {
		return "", err
	}

	if _, err := os.Stat(javaBin); err != nil {
		return "", fmt.Errorf("extracted runtime has no bin/java: %w", err)
	}
	return javaBin, nil
}

// ensureFile downloads a single checksummed file into cacheDir/name if not
// already present and valid.
func ensureFile(ctx context.Context, cacheDir, name string, asset Asset) (string, error) {
	dest := filepath.Join(cacheDir, name)

	if sum, err := sha256File(dest); err == nil && sum == asset.SHA256 {
		return dest, nil
	}

	if err := download(ctx, asset, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func download(ctx context.Context, asset Asset, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: unexpected status %s", asset.URL, resp.Status)
	}

	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(f, h), resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()

	if got := hex.EncodeToString(h.Sum(nil)); got != asset.SHA256 {
		os.Remove(tmp)
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", asset.URL, got, asset.SHA256)
	}

	return os.Rename(tmp, dest)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, filepath.Clean(hdr.Name))
		if !isWithinDir(destDir, target) {
			return fmt.Errorf("archive entry escapes destination: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
}

func isWithinDir(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
