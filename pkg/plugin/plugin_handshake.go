// Package plugin provides a plugin system implementation based on HashiCorp go-plugin
// Supports multiple plugin types: CI/CD, security, notification, storage, etc.
package plugin

import (
	"encoding/json"
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
	Impl       interface{}
	DbAccessor DatabaseAccessor
}

// Server returns the server-side plugin
func (h *RPCPluginHandler) Server(*plugin.MuxBroker) (interface{}, error) {
	// 如果 impl 是 RPCPluginServer，直接使用
	if server, ok := h.Impl.(*RPCPluginServer); ok {
		return server, nil
	}
	// 否则创建包装器
	return &RPCPluginServerWrapper{
		impl:       h.Impl,
		dbAccessor: h.DbAccessor,
	}, nil
}

// Client returns the client-side plugin
func (h *RPCPluginHandler) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCPluginClientWrapper{client: c}, nil
}

// RPCPluginServerWrapper is the RPC plugin server wrapper
type RPCPluginServerWrapper struct {
	impl       interface{}
	dbAccessor DatabaseAccessor // 新接口
}

// 实现 RPCPluginServer 的方法，将调用转发到 impl
func (w *RPCPluginServerWrapper) Ping(args string, reply *string) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.Ping(args, reply)
}

func (w *RPCPluginServerWrapper) GetInfo(args string, reply *PluginInfo) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.GetInfo(args, reply)
}

func (w *RPCPluginServerWrapper) GetMetrics(args string, reply *PluginMetrics) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.GetMetrics(args, reply)
}

func (w *RPCPluginServerWrapper) Init(config json.RawMessage, reply *string) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.Init(config, reply)
}

func (w *RPCPluginServerWrapper) Cleanup(args string, reply *string) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.Cleanup(args, reply)
}

func (w *RPCPluginServerWrapper) QueryConfig(args *QueryConfigArgs, reply *string) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.QueryConfig(args, reply)
}

func (w *RPCPluginServerWrapper) QueryConfigByKey(args *QueryConfigByKeyArgs, reply *string) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.QueryConfigByKey(args, reply)
}

func (w *RPCPluginServerWrapper) ListConfigs(args string, reply *string) error {
	server := &RPCPluginServer{
		info:       PluginInfo{},
		instance:   w.impl,
		dbAccessor: w.dbAccessor,
	}
	return server.ListConfigs(args, reply)
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
