package metrics

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for metrics
var ProviderSet = wire.NewSet(
	NewMetricsServer,
)

// NewMetricsServer creates a new metrics server from config
func NewMetricsServer(config MetricsConfig) *Server {
	server := NewServer(config)
	// Register cron metrics
	SetupCronMetrics(server.GetRegistry())
	return server
}
