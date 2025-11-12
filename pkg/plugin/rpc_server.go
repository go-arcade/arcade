// Package plugin RPC server implementation
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
)

// RPCPluginServer is the RPC plugin server implementation
// Handles RPC calls from clients and forwards them to the actual plugin instance
type RPCPluginServer struct {
	// Plugin basic information
	info PluginInfo
	// Plugin instance (object that implements specific functionality)
	instance interface{}
	// Database accessor (新的接口实现)
	dbAccessor DatabaseAccessor
}

// NewRPCPluginServer creates a new RPC plugin server
func NewRPCPluginServer(info PluginInfo, instance interface{}, dbAccessor DatabaseAccessor) *RPCPluginServer {
	return &RPCPluginServer{
		info:       info,
		instance:   instance,
		dbAccessor: dbAccessor,
	}
}

// Ping checks the plugin status
func (s *RPCPluginServer) Ping(args string, reply *string) error {
	*reply = "pong"
	return nil
}

// GetInfo retrieves plugin information
func (s *RPCPluginServer) GetInfo(args string, reply *PluginInfo) error {
	*reply = s.info
	return nil
}

// GetMetrics retrieves plugin metrics
func (s *RPCPluginServer) GetMetrics(args string, reply *PluginMetrics) error {
	*reply = PluginMetrics{
		Name:    s.info.Name,
		Type:    s.info.Type,
		Version: s.info.Version,
		Status:  "running",
	}
	return nil
}

// Init initializes the plugin
func (s *RPCPluginServer) Init(config json.RawMessage, reply *string) error {
	if initPlugin, ok := s.instance.(interface{ Init(json.RawMessage) error }); ok {
		if err := initPlugin.Init(config); err != nil {
			return fmt.Errorf("plugin init failed: %w", err)
		}
	}
	*reply = "initialized"
	return nil
}

// Cleanup cleans up the plugin
func (s *RPCPluginServer) Cleanup(args string, reply *string) error {
	if cleanupPlugin, ok := s.instance.(interface{ Cleanup() error }); ok {
		if err := cleanupPlugin.Cleanup(); err != nil {
			return fmt.Errorf("plugin cleanup failed: %w", err)
		}
	}
	*reply = "cleaned up"
	return nil
}

// Send sends a notification
func (s *RPCPluginServer) Send(args *NotifySendArgs, reply *string) error {
	if notifyPlugin, ok := s.instance.(NotifyPluginRPCInterface); ok {
		if err := notifyPlugin.Send(args.Message, args.Opts); err != nil {
			return fmt.Errorf("send notification failed: %w", err)
		}
	} else {
		return fmt.Errorf("plugin does not support notify interface")
	}
	*reply = "sent"
	return nil
}

// SendTemplate sends a template notification
func (s *RPCPluginServer) SendTemplate(args *NotifyTemplateArgs, reply *string) error {
	if notifyPlugin, ok := s.instance.(NotifyPluginRPCInterface); ok {
		if err := notifyPlugin.SendTemplate(args.Template, args.Data, args.Opts); err != nil {
			return fmt.Errorf("send template notification failed: %w", err)
		}
	} else {
		return fmt.Errorf("plugin does not support notify interface")
	}
	*reply = "sent"
	return nil
}

// QueryConfig 查询插件配置（RPC 方法）
func (s *RPCPluginServer) QueryConfig(args *QueryConfigArgs, reply *string) error {
	if s.dbAccessor == nil {
		return fmt.Errorf("database accessor is not available")
	}
	result, err := s.dbAccessor.QueryConfig(context.Background(), args.PluginID)
	if err != nil {
		return fmt.Errorf("query config failed: %w", err)
	}
	*reply = result
	return nil
}

// QueryConfigByKey 根据配置键查询配置值（RPC 方法）
func (s *RPCPluginServer) QueryConfigByKey(args *QueryConfigByKeyArgs, reply *string) error {
	if s.dbAccessor == nil {
		return fmt.Errorf("database accessor is not available")
	}
	result, err := s.dbAccessor.QueryConfigByKey(context.Background(), args.PluginID, args.Key)
	if err != nil {
		return fmt.Errorf("query config by key failed: %w", err)
	}
	*reply = result
	return nil
}

// ListConfigs 列出所有插件配置（RPC 方法）
func (s *RPCPluginServer) ListConfigs(args string, reply *string) error {
	if s.dbAccessor == nil {
		return fmt.Errorf("database accessor is not available")
	}
	result, err := s.dbAccessor.ListConfigs(context.Background())
	if err != nil {
		return fmt.Errorf("list configs failed: %w", err)
	}
	*reply = result
	return nil
}
