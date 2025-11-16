// Package plugin RPC server implementation
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
)

// Server is the RPC plugin server implementation
// Handles RPC calls from clients and forwards them to the actual plugin instance
// Uses method names + json.RawMessage for all plugin operations
type Server struct {
	// Plugin basic information
	info PluginInfo
	// Plugin instance (object that implements Plugin interface)
	instance any
	// Database accessor for config-related actions
	dbAccessor DatabaseAccessor
}

// NewServer creates a new RPC plugin server
func NewServer(info PluginInfo, instance any, dbAccessor DatabaseAccessor) *Server {
	return &Server{
		info:       info,
		instance:   instance,
		dbAccessor: dbAccessor,
	}
}

// Ping checks the plugin status
func (s *Server) Ping(args string, reply *string) error {
	*reply = "pong"
	return nil
}

// GetInfo retrieves plugin information
func (s *Server) GetInfo(args string, reply *PluginInfo) error {
	*reply = s.info
	return nil
}

// GetMetrics retrieves plugin metrics
func (s *Server) GetMetrics(args string, reply *PluginMetrics) error {
	*reply = PluginMetrics{
		Name:    s.info.Name,
		Type:    s.info.Type,
		Version: s.info.Version,
		Status:  "running",
	}
	return nil
}

// Init initializes the plugin
func (s *Server) Init(config json.RawMessage, reply *string) error {
	if initPlugin, ok := s.instance.(interface{ Init(json.RawMessage) error }); ok {
		if err := initPlugin.Init(config); err != nil {
			return fmt.Errorf("plugin init failed: %w", err)
		}
	}
	*reply = "initialized"
	return nil
}

// Cleanup cleans up the plugin
func (s *Server) Cleanup(args string, reply *string) error {
	if cleanupPlugin, ok := s.instance.(interface{ Cleanup() error }); ok {
		if err := cleanupPlugin.Cleanup(); err != nil {
			return fmt.Errorf("plugin cleanup failed: %w", err)
		}
	}
	*reply = "cleaned up"
	return nil
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
func (s *Server) ConfigQuery(args *MethodArgs, reply *MethodResult) error {
	if s.dbAccessor == nil {
		reply.Error = "database accessor is not available"
		return nil
	}
	var params ConfigQueryArgs
	if err := UnmarshalParams(args.Params, &params); err != nil {
		reply.Error = fmt.Sprintf("invalid params: %v", err)
		return nil
	}
	result, err := s.dbAccessor.QueryConfig(context.Background(), params.PluginID)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	reply.Result = json.RawMessage(result)
	return nil
}

// ConfigQueryByKey queries plugin config by key
func (s *Server) ConfigQueryByKey(args *MethodArgs, reply *MethodResult) error {
	if s.dbAccessor == nil {
		reply.Error = "database accessor is not available"
		return nil
	}
	var params ConfigQueryByKeyArgs
	if err := UnmarshalParams(args.Params, &params); err != nil {
		reply.Error = fmt.Sprintf("invalid params: %v", err)
		return nil
	}
	result, err := s.dbAccessor.QueryConfigByKey(context.Background(), params.PluginID, params.Key)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	reply.Result = json.RawMessage(result)
	return nil
}

// ConfigList lists all plugin configs
func (s *Server) ConfigList(args *MethodArgs, reply *MethodResult) error {
	if s.dbAccessor == nil {
		reply.Error = "database accessor is not available"
		return nil
	}
	var params ConfigListArgs
	if err := UnmarshalParams(args.Params, &params); err != nil {
		reply.Error = fmt.Sprintf("invalid params: %v", err)
		return nil
	}
	result, err := s.dbAccessor.ListConfigs(context.Background())
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	reply.Result = json.RawMessage(result)
	return nil
}
