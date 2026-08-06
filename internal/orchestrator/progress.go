package orchestrator

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// Reporter prints the human-facing provisioning/startup progress to an
// io.Writer (normally os.Stdout). It is the pretty layer for the default
// run; log/slog remains the structured layer for --verbose/diagnostics —
// the two are deliberately separate outputs for separate audiences, not
// one compromise format.
type Reporter struct {
	w       io.Writer
	tty     bool
	noColor bool
	mu      sync.Mutex
}

// NewReporter builds a Reporter for w, auto-detecting whether it's an
// interactive terminal (enables the in-place spinner/progress line) and
// whether color/unicode should be suppressed (NO_COLOR, or w isn't a
// *os.File at all).
func NewReporter(w io.Writer) *Reporter {
	tty := false
	if f, ok := w.(*os.File); ok {
		tty = term.IsTerminal(int(f.Fd()))
	}
	_, noColor := os.LookupEnv("NO_COLOR")
	return &Reporter{w: w, tty: tty, noColor: noColor}
}

// Blank prints a single blank line, e.g. to separate sections.
func (r *Reporter) Blank() {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprintln(r.w)
}

// Section prints a section header ("Setting up (first run only)…",
// "Starting…"), preceded by a blank line to separate it from whatever
// came before — the caller is always expected to have printed something
// first (at minimum the "mustang-webui" banner), so this is unconditional.
func (r *Reporter) Section(title string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprintln(r.w)
	fmt.Fprintln(r.w, title)
	fmt.Fprintln(r.w)
}

// Ok prints a top-level (2-space indent) completed step, e.g.
// "Sidecar ready on 127.0.0.1:53211".
func (r *Reporter) Ok(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprintf(r.w, "  %s %s\n", r.check(), label)
}

// Fail prints a top-level (2-space indent) failed step.
func (r *Reporter) Fail(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprintf(r.w, "  %s %s\n", r.cross(), label)
}

// Asset begins reporting progress for one downloaded asset: prints its
// label and the exact URL it's coming from (a static line, kept in
// scrollback — the app's whole trust pitch is "one known origin," so
// that origin is never hidden behind a vague "Downloading…").
func (r *Reporter) Asset(label, url string) *AssetProgress {
	r.mu.Lock()
	fmt.Fprintf(r.w, "  %s\n", label)
	fmt.Fprintf(r.w, "    %s\n", url)
	r.mu.Unlock()

	return &AssetProgress{r: r, start: time.Now(), spinner: newSpinner()}
}

func (r *Reporter) check() string {
	if r.noColor {
		return "[ok]"
	}
	return "✓"
}

func (r *Reporter) cross() string {
	if r.noColor {
		return "[x]"
	}
	return "✗"
}

// AssetProgress tracks one in-flight download.
type AssetProgress struct {
	r          *Reporter
	start      time.Time
	spinner    *spinner
	lastRender time.Time
	rendered   bool
}

const progressThrottle = 100 * time.Millisecond

// Update reports downloaded/total bytes so far. Only draws on an
// interactive terminal (throttled) — on a non-TTY (redirected to a file,
// running under a service manager) it's a no-op, so logs don't fill up
// with thousands of \r-only lines that never actually appear as separate
// lines in a file.
func (a *AssetProgress) Update(downloaded, total int64) {
	if !a.r.tty {
		return
	}
	if !a.lastRender.IsZero() && time.Since(a.lastRender) < progressThrottle {
		return
	}
	a.lastRender = time.Now()
	a.rendered = true

	elapsed := time.Since(a.start).Seconds()
	rate := float64(0)
	if elapsed > 0 {
		rate = float64(downloaded) / elapsed
	}

	a.r.mu.Lock()
	defer a.r.mu.Unlock()
	line := fmt.Sprintf("    %s %s / %s   %s/s", a.spinner.next(), humanBytes(downloaded), humanBytes(total), humanBytes(int64(rate)))
	fmt.Fprint(a.r.w, "\r"+clearLine+line)
}

// Done marks the download complete and prints a static summary line,
// replacing any in-place progress line.
func (a *AssetProgress) Done(total int64) {
	a.r.mu.Lock()
	defer a.r.mu.Unlock()
	if a.rendered {
		fmt.Fprint(a.r.w, "\r"+clearLine)
	}
	fmt.Fprintf(a.r.w, "    %s downloaded   %s   %s\n", a.r.check(), humanBytes(total), humanDuration(time.Since(a.start)))
}

// SubOk prints a completed sub-step under the current asset (checksum
// verified, extracted), at the same 4-space indent as the URL/progress.
func (a *AssetProgress) SubOk(label string) {
	a.r.mu.Lock()
	defer a.r.mu.Unlock()
	fmt.Fprintf(a.r.w, "    %s %s\n", a.r.check(), label)
}

// Fail marks the download/verification as failed, replacing any in-place
// progress line with a static failure line.
func (a *AssetProgress) Fail(label string) {
	a.r.mu.Lock()
	defer a.r.mu.Unlock()
	if a.rendered {
		fmt.Fprint(a.r.w, "\r"+clearLine)
	}
	fmt.Fprintf(a.r.w, "    %s %s\n", a.r.cross(), label)
}

// clearLine erases whatever's currently on the terminal line before we
// redraw it — needed because the new line isn't guaranteed to be at
// least as long as the old one (e.g. shrinking "MB" widths).
const clearLine = "\x1b[2K"

type spinner struct {
	frames []rune
	i      int
}

func newSpinner() *spinner {
	return &spinner{frames: []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")} // braille dots
}

func (s *spinner) next() string {
	r := s.frames[s.i%len(s.frames)]
	s.i++
	return string(r)
}

func humanBytes(n int64) string {
	const unit = 1000
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	units := "KMGT"[exp]
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), units)
}

func humanDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ExplainError turns a small set of known provisioning failures into a
// plain, actionable sentence for the default (non-verbose) run. Anything
// it does not recognize falls back to a generic retry suggestion — the
// full wrapped error is still available via %w for --verbose/log output.
func ExplainError(err error) string {
	var mismatch *ChecksumError
	if errors.As(err, &mismatch) {
		return fmt.Sprintf(
			"The download from %s doesn't match its expected checksum — it may\n"+
				"have been corrupted or intercepted in transit. Nothing was installed.\n"+
				"Run mustang-webui again; if this keeps happening, check your network\n"+
				"connection.",
			mismatch.URL,
		)
	}

	return "Run mustang-webui again; if this keeps happening, check your network\nconnection or your machine's available disk space."
}
