package orchestrator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownload_Success(t *testing.T) {
	content := []byte("pretend-jre-archive-bytes")
	sum := sha256.Sum256(content)
	checksum := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "asset.bin")

	var out bytes.Buffer
	reporter := NewReporter(&out)
	ap := reporter.Asset("Runtime (JRE)", srv.URL)

	asset := Asset{URL: srv.URL, SHA256: checksum}
	if err := download(context.Background(), asset, dest, ap); err != nil {
		t.Fatalf("download: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("downloaded content = %q, want %q", got, content)
	}

	if _, err := os.Stat(dest + ".part"); !os.IsNotExist(err) {
		t.Errorf(".part file should have been renamed away, stat err = %v", err)
	}

	log := out.String()
	if !strings.Contains(log, "Runtime (JRE)") {
		t.Errorf("log missing asset label:\n%s", log)
	}
	if !strings.Contains(log, srv.URL) {
		t.Errorf("log missing exact source URL:\n%s", log)
	}
	if !strings.Contains(log, "downloaded") {
		t.Errorf("log missing completion line:\n%s", log)
	}
}

func TestDownload_ChecksumMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("actual-content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "asset.bin")

	var out bytes.Buffer
	reporter := NewReporter(&out)
	ap := reporter.Asset("Sidecar", srv.URL)

	asset := Asset{URL: srv.URL, SHA256: strings.Repeat("0", 64)}
	err := download(context.Background(), asset, dest, ap)
	if err == nil {
		t.Fatal("expected checksum error, got nil")
	}

	var mismatch *ChecksumError
	if !errors.As(err, &mismatch) {
		t.Fatalf("error is not a *ChecksumError: %v", err)
	}
	if mismatch.URL != srv.URL {
		t.Errorf("ChecksumError.URL = %q, want %q", mismatch.URL, srv.URL)
	}

	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Error("dest file should not exist after a checksum failure")
	}
	if _, statErr := os.Stat(dest + ".part"); !os.IsNotExist(statErr) {
		t.Error(".part file should be cleaned up after a checksum failure")
	}

	if !strings.Contains(out.String(), "checksum verification failed") {
		t.Errorf("log missing failure line:\n%s", out.String())
	}
}

func TestExplainError_ChecksumMismatch(t *testing.T) {
	err := &ChecksumError{URL: "https://example.com/a.tar.gz", Got: "aaa", Want: "bbb"}
	msg := ExplainError(err)
	if !strings.Contains(msg, "https://example.com/a.tar.gz") {
		t.Errorf("ExplainError should reference the URL that failed, got: %s", msg)
	}
}

func TestExplainError_Generic(t *testing.T) {
	msg := ExplainError(errors.New("connection refused"))
	if msg == "" {
		t.Error("ExplainError should never return an empty string")
	}
}

func TestEnsureFile_SkipsWhenAlreadyValid(t *testing.T) {
	content := []byte("cached-content")
	sum := sha256.Sum256(content)
	checksum := hex.EncodeToString(sum[:])

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write(content)
	}))
	defer srv.Close()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sidecar.jar"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	reporter := NewReporter(&out)
	asset := Asset{URL: srv.URL, SHA256: checksum}

	announced := false
	if _, err := ensureFile(context.Background(), dir, "sidecar.jar", "Sidecar", asset, reporter, func() { announced = true }); err != nil {
		t.Fatalf("ensureFile: %v", err)
	}

	if calls != 0 {
		t.Errorf("expected no HTTP calls for an already-valid cached file, got %d", calls)
	}
	if announced {
		t.Error("section header should not be announced when nothing needs downloading")
	}
	if out.Len() != 0 {
		t.Errorf("expected no reporter output for a cache hit, got:\n%s", out.String())
	}
}
