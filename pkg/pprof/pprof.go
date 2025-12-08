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
	Path   string
}

// SetDefaults sets default values for PprofConfig
func (p *PprofConfig) SetDefaults() {
	if p.Host == "" {
		p.Host = "0.0.0.0"
	}
	if p.Port == 0 {
		p.Port = 8083
	}
	if p.Path == "" {
		p.Path = "/debug/pprof"
	}
}

// Server represents a pprof server
type Server struct {
	config PprofConfig
	server *http.Server
}

// NewServer creates a new pprof server
func NewServer(config PprofConfig) *Server {
	config.SetDefaults()

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

	// Use configured path prefix (defaults should be set via SetDefaults)
	pathPrefix := s.config.Path

	mux := http.NewServeMux()
	// Register pprof routes
	mux.HandleFunc(pathPrefix+"/", pprof.Index)
	mux.HandleFunc(pathPrefix+"/cmdline", pprof.Cmdline)
	mux.HandleFunc(pathPrefix+"/profile", pprof.Profile)
	mux.HandleFunc(pathPrefix+"/symbol", pprof.Symbol)
	mux.HandleFunc(pathPrefix+"/trace", pprof.Trace)
	mux.Handle(pathPrefix+"/allocs", pprof.Handler("allocs"))
	mux.Handle(pathPrefix+"/block", pprof.Handler("block"))
	mux.Handle(pathPrefix+"/goroutine", pprof.Handler("goroutine"))
	mux.Handle(pathPrefix+"/heap", pprof.Handler("heap"))
	mux.Handle(pathPrefix+"/mutex", pprof.Handler("mutex"))
	mux.Handle(pathPrefix+"/threadcreate", pprof.Handler("threadcreate"))

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
