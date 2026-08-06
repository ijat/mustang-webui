package orchestrator

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{500, "500 B"},
		{1_000, "1.0 KB"},
		{68_400_000, "68.4 MB"},
		{1_500_000_000, "1.5 GB"},
	}
	for _, c := range cases {
		if got := humanBytes(c.n); got != c.want {
			t.Errorf("humanBytes(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestHumanDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{7 * time.Second, "00:07"},
		{65 * time.Second, "01:05"},
	}
	for _, c := range cases {
		if got := humanDuration(c.d); got != c.want {
			t.Errorf("humanDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestReporter_BlankLineBetweenSiblingAssets(t *testing.T) {
	var out bytes.Buffer
	r := NewReporter(&out)

	r.Section("Setting up (first run only)…")
	first := r.Asset("Runtime (JRE)", "https://example.com/jre.tar.gz")
	first.SubOk("checksum verified")
	first.SubOk("extracted")
	second := r.Asset("Sidecar", "https://example.com/sidecar.jar")
	second.SubOk("checksum verified")

	got := out.String()
	want := "\nSetting up (first run only)…\n\n" +
		"  Runtime (JRE)\n    https://example.com/jre.tar.gz\n    ✓ checksum verified\n    ✓ extracted\n" +
		"\n" +
		"  Sidecar\n    https://example.com/sidecar.jar\n    ✓ checksum verified\n"
	if got != want {
		t.Errorf("asset spacing mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestReporter_SectionAlwaysPrecededByBlankLine(t *testing.T) {
	var out bytes.Buffer
	r := NewReporter(&out)

	fmt.Fprintln(&out, "mustang-webui")
	r.Section("Setting up (first run only)…")
	r.Ok("Downloaded runtime")
	r.Section("Starting…")
	r.Ok("Sidecar ready")

	got := out.String()
	want := "mustang-webui\n\nSetting up (first run only)…\n\n  ✓ Downloaded runtime\n\nStarting…\n\n  ✓ Sidecar ready\n"
	if got != want {
		t.Errorf("Section/Ok output mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestReporter_NoColorUsesPlainMarkers(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	r := NewReporter(&out)
	r.Ok("done")
	r.Fail("nope")

	got := out.String()
	if strings.Contains(got, "✓") || strings.Contains(got, "✗") {
		t.Errorf("NO_COLOR should suppress unicode markers, got: %q", got)
	}
	if !strings.Contains(got, "[ok]") || !strings.Contains(got, "[x]") {
		t.Errorf("NO_COLOR should use plain [ok]/[x] markers, got: %q", got)
	}
}

func TestSpinner_Cycles(t *testing.T) {
	s := newSpinner()
	first := s.next()
	for i := 0; i < len(s.frames)-1; i++ {
		s.next()
	}
	if got := s.next(); got != first {
		t.Errorf("spinner should cycle back to its first frame after a full period, got %q want %q", got, first)
	}
}
