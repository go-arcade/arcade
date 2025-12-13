package pprof

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for pprof
var ProviderSet = wire.NewSet(
	NewPprofServer,
)

// NewPprofServer creates a new pprof server from config
func NewPprofServer(config PprofConfig) *Server {
	return NewServer(config)
}
