// Package plugin RPC client implementation
package plugin

import (
	"encoding/json"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// RPCPluginClient is the RPC plugin client
type RPCPluginClient struct {
	// Plugin information
	info PluginInfo
	// Plugin configuration
	config PluginConfig
	// Plugin path
	pluginPath string
	// RPC client
	client *rpc.Client
	// go-plugin client
	pluginClient *plugin.Client
	// Plugin instance
	instance any
	// Connection status
	connected bool
	// Last heartbeat time
	lastHeartbeat int64
	// Error count
	errorCount int
}

// Ping checks the plugin connection status
func (c *RPCPluginClient) Ping() error {
	if c.client == nil {
		return fmt.Errorf("RPC client is not initialized")
	}
	var result string
	return c.client.Call("Plugin.Ping", "", &result)
}

// GetInfo retrieves plugin information
func (c *RPCPluginClient) GetInfo() (PluginInfo, error) {
	if c.client == nil {
		return c.info, nil
	}
	var info PluginInfo
	err := c.client.Call("Plugin.GetInfo", "", &info)
	return info, err
}

// GetMetrics retrieves plugin runtime metrics
func (c *RPCPluginClient) GetMetrics() (PluginMetrics, error) {
	if c.client == nil {
		return PluginMetrics{}, fmt.Errorf("RPC client is not initialized")
	}
	var metrics PluginMetrics
	err := c.client.Call("Plugin.GetMetrics", "", &metrics)
	return metrics, err
}

// Send sends a notification message
func (c *RPCPluginClient) Send(message json.RawMessage, opts json.RawMessage) error {
	if c.client == nil {
		return fmt.Errorf("RPC client is not initialized")
	}
	args := &NotifySendArgs{
		Message: message,
		Opts:    opts,
	}
	var result string
	return c.client.Call("Plugin.Send", args, &result)
}

// SendTemplate sends a notification using a template
func (c *RPCPluginClient) SendTemplate(template string, data json.RawMessage, opts json.RawMessage) error {
	if c.client == nil {
		return fmt.Errorf("RPC client is not initialized")
	}
	args := &NotifyTemplateArgs{
		Template: template,
		Data:     data,
		Opts:     opts,
	}
	var result string
	return c.client.Call("Plugin.SendTemplate", args, &result)
}

// Call invokes a generic RPC method
func (c *RPCPluginClient) Call(method string, args interface{}, reply interface{}) error {
	if c.client == nil {
		return fmt.Errorf("RPC client is not initialized")
	}
	return c.client.Call(method, args, reply)
}

// Close closes the plugin client and releases resources
func (c *RPCPluginClient) Close() error {
	if c.pluginClient != nil {
		c.pluginClient.Kill()
	}
	return nil
}
