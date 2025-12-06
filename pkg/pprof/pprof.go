package pprof

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/go-arcade/arcade/pkg/log"
)

// PprofConfig holds pprof server configuration
type PprofConfig struct {
	Host   string
	Port   int
	Enable bool
}

// Server represents a pprof server
type Server struct {
	config PprofConfig
	server *http.Server
}

// NewServer creates a new pprof server
func NewServer(config PprofConfig) *Server {
	return &Server{
		config: config,
	}
}

// Start starts the pprof HTTP server
func (s *Server) Start() error {
	if !s.config.Enable {
		log.Info("Pprof server is disabled")
		return nil
	}

	mux := http.NewServeMux()
	// Register pprof routes
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Infow("Pprof server started", "address", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorw("Pprof server failed", "error", err)
		}
	}()

	return nil
}

// Stop stops the pprof HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
