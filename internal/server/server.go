package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jvzantvoort/webui/internal/config"
	"github.com/jvzantvoort/webui/internal/content"
	"github.com/jvzantvoort/webui/internal/data"
	"github.com/jvzantvoort/webui/internal/tmpl"
	"github.com/jvzantvoort/webui/static"
)

// Server holds the HTTP mux and the underlying http.Server for graceful shutdown.
type Server struct {
	cfg      *config.Config
	mux      *http.ServeMux
	renderer *tmpl.Renderer
	httpSrv  *http.Server
}

// New creates a Server with all routes registered.
// Returns an error if template parsing fails.
func New(cfg *config.Config) (*Server, error) {
	r, err := tmpl.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("init templates: %w", err)
	}

	s := &Server{cfg: cfg, mux: http.NewServeMux(), renderer: r}
	s.registerRoutes()
	return s, nil
}

func (s *Server) registerRoutes() {
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))

	for _, item := range s.cfg.Content {
		s.mux.Handle("/"+item.Name+"/", content.NewHandler(item, s.renderer))
	}

	for _, item := range s.cfg.Data {
		s.mux.Handle("/data/"+item.Name+"/", data.NewHandler(item, s.renderer))
	}

	s.mux.HandleFunc("POST /api/shutdown", s.handleShutdown)
	s.mux.HandleFunc("/", s.handleIndex)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	counts := make(map[string]int, len(s.cfg.Data))
	for _, item := range s.cfg.Data {
		counts["/data/"+item.Name+"/"] = data.CountRows(item.Path)
	}
	s.renderer.Render(w, "index.html", tmpl.PageData{Title: "webui", Counts: counts})
}

// handleShutdown responds immediately then gracefully stops the server.
func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpSrv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()
}

// Start begins listening on the configured port.
// It returns nil when the server is stopped via Shutdown.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)
	s.httpSrv = &http.Server{Addr: addr, Handler: s.mux}

	log.Printf("listening on http://localhost%s", addr)

	if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Handler returns the underlying http.Handler (for testing).
func (s *Server) Handler() http.Handler {
	return s.mux
}
