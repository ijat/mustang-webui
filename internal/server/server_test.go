package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFrontendHandler_ServesEmbeddedDist(t *testing.T) {
	handler, err := newFrontendHandler("")
	if err != nil {
		t.Fatalf("newFrontendHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<html") {
		t.Fatalf("body does not look like html: %q", rec.Body.String())
	}
}

func TestFrontendHandler_DevProxiesToVite(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("vite dev server"))
	}))
	defer upstream.Close()

	handler, err := newFrontendHandler(upstream.URL)
	if err != nil {
		t.Fatalf("newFrontendHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "vite dev server" {
		t.Fatalf("body = %q, want proxied response", got)
	}
}
