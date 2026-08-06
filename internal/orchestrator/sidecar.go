package orchestrator

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Sidecar is a running JVM subprocess exposing the mustang wrapper API on
// loopback only.
type Sidecar struct {
	cmd     *exec.Cmd
	BaseURL string
	Token   string
	Port    int
}

// StartSidecar launches the sidecar jar on a free loopback port, waits for
// it to report healthy, and returns a handle to it. The caller must call
// Stop when done. reporter may be nil (used by tests).
//
// The JVM's own stdout/stderr is captured rather than connected directly
// to ours — it's Java's own log noise, not part of the CLI's output, and
// piping it straight through would interleave with the pretty progress
// lines above. It's only surfaced (as part of the returned error) if the
// sidecar fails to become healthy, where it's genuinely useful for
// diagnosing why.
func StartSidecar(ctx context.Context, rt *RuntimePaths, reporter *Reporter) (*Sidecar, error) {
	port, err := freePort()
	if err != nil {
		return nil, fmt.Errorf("finding a free port: %w", err)
	}

	token, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("generating auth token: %w", err)
	}

	cmd := exec.CommandContext(ctx, rt.JavaBin,
		"-jar", rt.SidecarJar,
		"--port", strconv.Itoa(port),
		"--token", token,
	)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting sidecar process: %w", err)
	}

	s := &Sidecar{
		cmd:     cmd,
		BaseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		Token:   token,
		Port:    port,
	}

	if err := s.waitHealthy(ctx); err != nil {
		_ = s.Stop()
		if output.Len() > 0 {
			return nil, fmt.Errorf("%w\n\nsidecar output:\n%s", err, output.String())
		}
		return nil, err
	}

	if reporter != nil {
		reporter.Ok(fmt.Sprintf("Sidecar ready on 127.0.0.1:%d", port))
	}

	return s, nil
}

func (s *Sidecar) waitHealthy(ctx context.Context) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.BaseURL+"/healthz", nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(150 * time.Millisecond):
		}
	}
	return fmt.Errorf("sidecar did not become healthy within 30s")
}

// Stop terminates the sidecar process, escalating to a kill if it doesn't
// exit promptly.
func (s *Sidecar) Stop() error {
	if s.cmd.Process == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() { done <- s.cmd.Wait() }()

	_ = s.cmd.Process.Signal(os.Interrupt)

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		_ = s.cmd.Process.Kill()
		return <-done
	}
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
