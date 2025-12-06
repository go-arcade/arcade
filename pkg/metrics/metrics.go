package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsConfig holds metrics server configuration
type MetricsConfig struct {
	Host   string
	Port   int
	Enable bool
}

// Server represents a metrics server
type Server struct {
	config     MetricsConfig
	server     *http.Server
	registry   *prometheus.Registry
	collectors []prometheus.Collector
	mu         sync.Mutex
}

// NewServer creates a new metrics server
func NewServer(config MetricsConfig) *Server {
	registry := prometheus.NewRegistry()
	// Register default collectors
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return &Server{
		config:     config,
		registry:   registry,
		collectors: make([]prometheus.Collector, 0),
	}
}

// RegisterCollector registers a prometheus collector
func (s *Server) RegisterCollector(collector prometheus.Collector) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.registry.Register(collector); err != nil {
		return fmt.Errorf("failed to register collector: %w", err)
	}
	s.collectors = append(s.collectors, collector)
	return nil
}

// Start starts the metrics HTTP server
func (s *Server) Start() error {
	if !s.config.Enable {
		log.Info("Metrics server is disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Infow("Metrics server started", "address", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorw("Metrics server failed", "error", err)
		}
	}()

	return nil
}

// Stop stops the metrics HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// GetRegistry returns the prometheus registry
func (s *Server) GetRegistry() *prometheus.Registry {
	return s.registry
}
