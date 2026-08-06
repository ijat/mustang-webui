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
	"sync/atomic"
	"time"
)

// RuntimePaths locates a ready-to-run JVM and sidecar jar.
type RuntimePaths struct {
	JavaBin    string
	SidecarJar string
}

// ChecksumError means a downloaded file didn't match its expected
// checksum — kept as a distinct type (rather than a plain fmt.Errorf) so
// the CLI's error reporting can recognize it and print a tailored,
// actionable explanation instead of a generic failure message.
type ChecksumError struct {
	URL  string
	Got  string
	Want string
}

func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch for %s: got %s, want %s", e.URL, e.Got, e.Want)
}

// ProvisionRuntime resolves a JavaBin + SidecarJar pair, downloading and
// caching them under cacheDir on first use. In dev mode it uses the
// system `java` on PATH and a locally built sidecar jar instead, so the
// project is runnable without a cut release. reporter may be nil, in
// which case provisioning proceeds silently (used by tests).
func ProvisionRuntime(ctx context.Context, cacheDir string, dev bool, reporter *Reporter) (*RuntimePaths, error) {
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

	sectioned := false
	announceSection := func() {
		if !sectioned && reporter != nil {
			reporter.Section("Setting up (first run only)…")
			sectioned = true
		}
	}

	javaBin, err := ensureRuntimeArchive(ctx, cacheDir, asset, reporter, announceSection)
	if err != nil {
		return nil, fmt.Errorf("provisioning JVM: %w", err)
	}

	jarPath, err := ensureFile(ctx, cacheDir, "sidecar.jar", "Sidecar", releaseManifest.Sidecar, reporter, announceSection)
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
// path inside it. announceSection is called only if a download actually
// happens, so a fully-cached run never prints "Setting up…" at all.
func ensureRuntimeArchive(ctx context.Context, cacheDir string, asset Asset, reporter *Reporter, announceSection func()) (string, error) {
	runtimeDir := filepath.Join(cacheDir, "runtime")
	javaBin := filepath.Join(runtimeDir, "bin", "java")

	if _, err := os.Stat(javaBin); err == nil {
		return javaBin, nil
	}

	announceSection()
	var ap *AssetProgress
	if reporter != nil {
		ap = reporter.Asset("Runtime (JRE)", asset.URL)
	}

	archivePath := filepath.Join(cacheDir, "runtime.tar.gz")
	if err := download(ctx, asset, archivePath, ap); err != nil {
		return "", err
	}
	defer os.Remove(archivePath)
	if ap != nil {
		ap.SubOk("checksum verified")
	}

	if err := extractTarGz(archivePath, runtimeDir); err != nil {
		if ap != nil {
			ap.Fail("extraction failed")
		}
		return "", err
	}
	if ap != nil {
		ap.SubOk("extracted")
	}

	if _, err := os.Stat(javaBin); err != nil {
		return "", fmt.Errorf("extracted runtime has no bin/java: %w", err)
	}
	return javaBin, nil
}

// ensureFile downloads a single checksummed file into cacheDir/name if not
// already present and valid.
func ensureFile(ctx context.Context, cacheDir, name, label string, asset Asset, reporter *Reporter, announceSection func()) (string, error) {
	dest := filepath.Join(cacheDir, name)

	if sum, err := sha256File(dest); err == nil && sum == asset.SHA256 {
		return dest, nil
	}

	announceSection()
	var ap *AssetProgress
	if reporter != nil {
		ap = reporter.Asset(label, asset.URL)
	}

	if err := download(ctx, asset, dest, ap); err != nil {
		return "", err
	}
	if ap != nil {
		ap.SubOk("checksum verified")
	}
	return dest, nil
}

// download fetches asset.URL to dest, verifying its checksum before the
// file is considered valid. progress may be nil.
func download(ctx context.Context, asset Asset, dest string, progress *AssetProgress) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if progress != nil {
			progress.Fail("download failed")
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if progress != nil {
			progress.Fail("download failed")
		}
		return fmt.Errorf("downloading %s: unexpected status %s", asset.URL, resp.Status)
	}

	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	h := sha256.New()
	counted := &countingWriter{w: io.MultiWriter(f, h)}
	total := resp.ContentLength

	stopTicker := func() {}
	if progress != nil {
		done := make(chan struct{})
		stopTicker = func() { close(done) }
		go reportProgressUntil(done, progress, counted, total)
	}

	_, copyErr := io.Copy(counted, resp.Body)
	stopTicker()
	if copyErr != nil {
		f.Close()
		os.Remove(tmp)
		if progress != nil {
			progress.Fail("download failed")
		}
		return copyErr
	}
	f.Close()

	got := hex.EncodeToString(h.Sum(nil))
	if got != asset.SHA256 {
		os.Remove(tmp)
		checksumErr := &ChecksumError{URL: asset.URL, Got: got, Want: asset.SHA256}
		if progress != nil {
			progress.Fail("checksum verification failed")
		}
		return checksumErr
	}

	if progress != nil {
		progress.Done(counted.n.Load())
	}

	return os.Rename(tmp, dest)
}

// countingWriter tracks bytes written so a concurrent ticker can report
// progress without the download loop itself needing to know about the
// terminal. n is accessed from both the copy loop and the ticker
// goroutine, hence atomic rather than a plain int64.
type countingWriter struct {
	w io.Writer
	n atomic.Int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n.Add(int64(n))
	return n, err
}

func reportProgressUntil(done <-chan struct{}, progress *AssetProgress, counted *countingWriter, total int64) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			progress.Update(counted.n.Load(), total)
		}
	}
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
