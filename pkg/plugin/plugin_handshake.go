// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package plugin provides a plugin system implementation based on HashiCorp go-plugin with gRPC
// Supports multiple plugin types: CI/CD, security, notification, storage, etc.
package plugin

import (
	"context"
	"errors"
	"net/rpc"

	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// PluginHandshake is the gRPC handshake configuration
// Protocol version 2, provides better security and functionality
var PluginHandshake = plugin.HandshakeConfig{
	ProtocolVersion:  2,
	MagicCookieKey:   "ARCADE_GRPC_PLUGIN",
	MagicCookieValue: "arcade-grpc-plugin-protocol",
}

// GRPCPlugin is the gRPC plugin handler
type GRPCPlugin struct {
	Impl any
	DB   DB
}

// Server returns the server-side plugin (required by plugin.Plugin interface, not used for gRPC)
func (h *GRPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, errors.New("RPC protocol not supported, use gRPC protocol instead")
}

// Client returns the client-side plugin (required by plugin.Plugin interface, not used for gRPC)
func (h *GRPCPlugin) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, errors.New("RPC protocol not supported, use gRPC protocol instead")
}

// GRPCServer returns the server-side plugin (for gRPC protocol)
func (h *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	var server *Server
	if srv, ok := h.Impl.(*Server); ok {
		server = srv
	} else {
		server = &Server{
			info:     &PluginInfo{},
			instance: h.Impl,
			db:       h.DB,
		}
	}
	// Register gRPC service
	pluginv1.RegisterPluginServiceServer(s, server)
	return nil
}

// GRPCClient returns the client-side plugin (for gRPC protocol)
func (h *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCPluginClientWrapper{conn: c}, nil
}

// GRPCPluginClientWrapper is the gRPC plugin client wrapper
type GRPCPluginClientWrapper struct {
	conn *grpc.ClientConn
}

// GetConn returns the underlying gRPC connection
func (w *GRPCPluginClientWrapper) GetConn() *grpc.ClientConn {
	return w.conn
}

// PluginMap is the plugin mapping
var PluginMap = map[string]plugin.Plugin{
	"plugin": &GRPCPlugin{},
}
