package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jvzantvoort/webui/internal/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 3110, Index: "README.md"},
	}
	s, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return s
}

func TestHandleIndex(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleNotFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestStaticAssets(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /static/css/style.css status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestShutdownEndpointMethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/shutdown", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	// GET is not registered; Go 1.22 mux returns 405 for wrong method on a known path.
	if w.Code == http.StatusOK {
		t.Error("GET /api/shutdown should not return 200")
	}
}
