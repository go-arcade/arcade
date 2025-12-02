package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
)

// Server is the gRPC plugin server implementation
// Handles gRPC calls from clients and forwards them to the actual plugin instance
type Server struct {
	pluginv1.UnimplementedPluginServiceServer
	// Plugin basic information
	info *PluginInfo
	// Plugin instance (object that implements Plugin interface)
	instance any
	// Database accessor for config-related actions
	dbAccessor DB
}

// NewServer creates a new gRPC plugin server
func NewServer(info *PluginInfo, instance any, dbAccessor DB) *Server {
	return &Server{
		info:       info,
		instance:   instance,
		dbAccessor: dbAccessor,
	}
}

// Ping checks the plugin status
func (s *Server) Ping(ctx context.Context, req *pluginv1.PingRequest) (*pluginv1.PingResponse, error) {
	return &pluginv1.PingResponse{Message: "pong"}, nil
}

// GetInfo retrieves plugin information
func (s *Server) GetInfo(ctx context.Context, req *pluginv1.GetInfoRequest) (*pluginv1.GetInfoResponse, error) {
	if s.info == nil {
		s.info = &PluginInfo{}
	}
	return &pluginv1.GetInfoResponse{
		Info: s.info,
	}, nil
}

// GetMetrics retrieves plugin metrics
func (s *Server) GetMetrics(ctx context.Context, req *pluginv1.GetMetricsRequest) (*pluginv1.GetMetricsResponse, error) {
	metrics := &pluginv1.PluginMetrics{
		Name:    s.info.GetName(),
		Type:    s.info.GetType(),
		Version: s.info.GetVersion(),
		Status:  "running",
	}
	return &pluginv1.GetMetricsResponse{Metrics: metrics}, nil
}

// Init initializes the plugin
func (s *Server) Init(ctx context.Context, req *pluginv1.InitRequest) (*pluginv1.InitResponse, error) {
	if initPlugin, ok := s.instance.(interface{ Init(json.RawMessage) error }); ok {
		if err := initPlugin.Init(req.Config); err != nil {
			return nil, fmt.Errorf("plugin init failed: %w", err)
		}
	}
	return &pluginv1.InitResponse{Message: "initialized"}, nil
}

// Cleanup cleans up the plugin
func (s *Server) Cleanup(ctx context.Context, req *pluginv1.CleanupRequest) (*pluginv1.CleanupResponse, error) {
	if cleanupPlugin, ok := s.instance.(interface{ Cleanup() error }); ok {
		if err := cleanupPlugin.Cleanup(); err != nil {
			return nil, fmt.Errorf("plugin cleanup failed: %w", err)
		}
	}
	return &pluginv1.CleanupResponse{Message: "cleaned up"}, nil
}

// Execute executes an action using the plugin instance
func (s *Server) Execute(ctx context.Context, req *pluginv1.ExecuteRequest) (*pluginv1.ExecuteResponse, error) {
	result, err := s.callPluginMethod(req.Action, req.Params, req.Opts)
	if err != nil {
		var rpcErr *RPCError
		// Try to get RPCError from wrapper
		if wrapper, ok := err.(interface{ GetRPCError() *RPCError }); ok {
			rpcErr = wrapper.GetRPCError()
		} else {
			rpcErr = &RPCError{
				Code:    500,
				Message: err.Error(),
			}
		}
		return &pluginv1.ExecuteResponse{
			Error: rpcErr,
		}, nil
	}
	return &pluginv1.ExecuteResponse{Result: result}, nil
}

// callPluginMethod is a helper to call plugin methods dynamically
// It tries to call the method on plugin instance, or falls back to Execute(action, ...) if available
func (s *Server) callPluginMethod(method string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	// Try to call the method directly on plugin instance
	// For plugins that implement Execute(action, ...), we use that
	if plugin, ok := s.instance.(interface {
		Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
	}); ok {
		return plugin.Execute(method, params, opts)
	}

	return nil, fmt.Errorf("plugin instance does not support method: %s", method)
}

// ========== Host-Provided Config Methods ==========

// ConfigQuery queries plugin config
func (s *Server) ConfigQuery(ctx context.Context, req *pluginv1.ConfigQueryRequest) (*pluginv1.ConfigQueryResponse, error) {
	if s.dbAccessor == nil {
		return &pluginv1.ConfigQueryResponse{
			Error: &pluginv1.RPCError{
				Code:    500,
				Message: "database accessor is not available",
			},
		}, nil
	}
	result, err := s.dbAccessor.QueryConfig(ctx, req.PluginId)
	if err != nil {
		return &pluginv1.ConfigQueryResponse{
			Error: &pluginv1.RPCError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}
	return &pluginv1.ConfigQueryResponse{Config: []byte(result)}, nil
}

// ConfigQueryByKey queries plugin config by key
func (s *Server) ConfigQueryByKey(ctx context.Context, req *pluginv1.ConfigQueryByKeyRequest) (*pluginv1.ConfigQueryByKeyResponse, error) {
	if s.dbAccessor == nil {
		return &pluginv1.ConfigQueryByKeyResponse{
			Error: &pluginv1.RPCError{
				Code:    500,
				Message: "database accessor is not available",
			},
		}, nil
	}
	result, err := s.dbAccessor.QueryConfigByKey(ctx, req.PluginId, req.Key)
	if err != nil {
		return &pluginv1.ConfigQueryByKeyResponse{
			Error: &pluginv1.RPCError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}
	return &pluginv1.ConfigQueryByKeyResponse{Value: []byte(result)}, nil
}

// ConfigList lists all plugin configs
func (s *Server) ConfigList(ctx context.Context, req *pluginv1.ConfigListRequest) (*pluginv1.ConfigListResponse, error) {
	if s.dbAccessor == nil {
		return &pluginv1.ConfigListResponse{
			Error: &pluginv1.RPCError{
				Code:    500,
				Message: "database accessor is not available",
			},
		}, nil
	}
	result, err := s.dbAccessor.ListConfigs(ctx)
	if err != nil {
		return &pluginv1.ConfigListResponse{
			Error: &pluginv1.RPCError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}
	return &pluginv1.ConfigListResponse{Configs: []byte(result)}, nil
}
