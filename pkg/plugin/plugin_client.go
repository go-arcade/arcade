package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
	"google.golang.org/grpc"
)

// Client is the gRPC plugin client
type Client struct {
	// Plugin information
	info *PluginInfo
	// Plugin configuration
	config *RuntimePluginConfig
	// Plugin path
	pluginPath string
	// gRPC client connection
	conn *grpc.ClientConn
	// gRPC service client
	client pluginv1.PluginServiceClient
	// go-plugin client (for process management)
	pluginClient interface {
		Kill()
		Exited() bool
	}
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
		return fmt.Errorf("gRPC client is not initialized")
	}
	ctx := context.Background()
	_, err := c.client.HealthCheck(ctx, &pluginv1.HealthCheckRequest{
		PluginId:  c.info.Name,
		Message:   "ping",
		Timestamp: time.Now().Unix(),
	})
	return err
}

// GetInfo retrieves plugin information
func (c *Client) GetInfo() (*PluginInfo, error) {
	if c.client == nil {
		return c.info, nil
	}
	ctx := context.Background()
	resp, err := c.client.GetInfo(ctx, &pluginv1.GetInfoRequest{})
	if err != nil {
		return c.info, err
	}
	if resp.Info != nil {
		c.info = resp.Info
	}
	return c.info, nil
}

// GetMetrics retrieves plugin runtime metrics
func (c *Client) GetMetrics() (*PluginMetrics, error) {
	if c.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	ctx := context.Background()
	resp, err := c.client.GetMetrics(ctx, &pluginv1.GetMetricsRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Metrics == nil {
		return nil, fmt.Errorf("metrics is nil")
	}
	return resp.Metrics, nil
}

// CallMethod calls a plugin method by name
// All plugin methods use unified signature: MethodName(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
// Example: CallMethod("Send", params, opts) calls Plugin.Execute(action="Send", params, opts)
func (c *Client) CallMethod(method string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	if c.client == nil {
		return nil, fmt.Errorf("gRPC client is not initialized")
	}
	ctx := context.Background()
	req := &pluginv1.ExecuteRequest{
		Action: method,
		Params: params,
		Opts:   opts,
	}
	resp, err := c.client.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, RPCErrorToError(resp.Error)
	}
	return resp.Result, nil
}

// Call invokes a generic gRPC method (for backward compatibility)
// This method is deprecated, use specific methods instead
func (c *Client) Call(method string, args any, reply any) error {
	// Convert to Execute call for backward compatibility
	if c.client == nil {
		return fmt.Errorf("gRPC client is not initialized")
	}
	params, err := json.Marshal(args)
	if err != nil {
		return err
	}
	result, err := c.CallMethod(method, params, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(result, reply)
}

// Close closes the plugin client and releases resources
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetConn returns the gRPC connection (for advanced usage)
func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}

// GetClient returns the gRPC service client (for advanced usage)
func (c *Client) GetClient() pluginv1.PluginServiceClient {
	return c.client
}
