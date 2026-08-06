package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheDir_IsNextToExecutable(t *testing.T) {
	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("CacheDir: %v", err)
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}

	wantParent := filepath.Dir(exe)
	if gotParent := filepath.Dir(dir); gotParent != wantParent {
		t.Errorf("CacheDir parent = %q, want %q (next to the executable)", gotParent, wantParent)
	}

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Errorf("CacheDir() = %q was not created as a directory: %v", dir, err)
	}
}
