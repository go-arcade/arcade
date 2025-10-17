// Package plugin RPC server implementation
package plugin

import (
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
}

// NewRPCPluginServer creates a new RPC plugin server
func NewRPCPluginServer(info PluginInfo, instance interface{}) *RPCPluginServer {
	return &RPCPluginServer{
		info:     info,
		instance: instance,
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
