// Package plugin provides a plugin system implementation based on HashiCorp go-plugin
// Supports multiple plugin types: CI/CD, security, notification, storage, etc.
package plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// RPCHandshake is the RPC handshake configuration (recommended new version)
// Protocol version 2, provides better security and functionality
var RPCHandshake = plugin.HandshakeConfig{
	ProtocolVersion:  2,
	MagicCookieKey:   "ARCADE_RPC_PLUGIN",
	MagicCookieValue: "arcade-rpc-plugin-protocol",
}

// RPCPluginHandler is the RPC plugin handler
type RPCPluginHandler struct {
	Impl interface{}
}

// Server returns the server-side plugin
func (h *RPCPluginHandler) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCPluginServerWrapper{impl: h.Impl}, nil
}

// Client returns the client-side plugin
func (h *RPCPluginHandler) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCPluginClientWrapper{client: c}, nil
}

// RPCPluginServerWrapper is the RPC plugin server wrapper
type RPCPluginServerWrapper struct {
	impl interface{}
}

// RPCPluginClientWrapper is the RPC plugin client wrapper
type RPCPluginClientWrapper struct {
	client *rpc.Client
}

// GetClient returns the underlying RPC client
func (w *RPCPluginClientWrapper) GetClient() *rpc.Client {
	return w.client
}

// PluginMap is the plugin mapping
var PluginMap = map[string]plugin.Plugin{
	"plugin": &RPCPluginHandler{},
}
