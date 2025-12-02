// Package plugin RPC client implementation
package plugin

import (
	"encoding/json"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Client is the RPC plugin client
type Client struct {
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
func (c *Client) Ping() error {
	if c.client == nil {
		return fmt.Errorf("RPC client is not initialized")
	}
	var result string
	return c.client.Call("Plugin.Ping", "", &result)
}

// GetInfo retrieves plugin information
func (c *Client) GetInfo() (PluginInfo, error) {
	if c.client == nil {
		return c.info, nil
	}
	var info PluginInfo
	err := c.client.Call("Plugin.GetInfo", "", &info)
	return info, err
}

// GetMetrics retrieves plugin runtime metrics
func (c *Client) GetMetrics() (PluginMetrics, error) {
	if c.client == nil {
		return PluginMetrics{}, fmt.Errorf("RPC client is not initialized")
	}
	var metrics PluginMetrics
	err := c.client.Call("Plugin.GetMetrics", "", &metrics)
	return metrics, err
}

// CallMethod calls a plugin method by name
// All plugin methods use unified signature: MethodName(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
// Example: CallMethod("Send", params, opts) calls Plugin.Send(params, opts)
func (c *Client) CallMethod(method string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	if c.client == nil {
		return nil, fmt.Errorf("RPC client is not initialized")
	}
	args := &MethodArgs{
		Params: params,
		Opts:   opts,
	}
	var result MethodResult
	// RPC method name format: "Plugin.MethodName"
	rpcMethod := fmt.Sprintf("Plugin.%s", method)
	if err := c.client.Call(rpcMethod, args, &result); err != nil {
		return nil, err
	}
	if result.Error != "" {
		return nil, fmt.Errorf("%s", result.Error)
	}
	return result.Result, nil
}

// Call invokes a generic RPC method
func (c *Client) Call(method string, args any, reply any) error {
	if c.client == nil {
		return fmt.Errorf("RPC client is not initialized")
	}
	return c.client.Call(method, args, reply)
}

// Close closes the plugin client and releases resources
func (c *Client) Close() error {
	if c.pluginClient != nil {
		c.pluginClient.Kill()
	}
	return nil
}
