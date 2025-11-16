package plugin

import "encoding/json"

// MethodArgs contains params and opts for RPC method calls
// All plugin methods use this unified signature: MethodName(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
// Example: Send(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
type MethodArgs struct {
	Params json.RawMessage `json:"params"`
	Opts   json.RawMessage `json:"opts"`
}

// MethodResult contains the result of RPC method call
type MethodResult struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error,omitempty"`
}

// ========== Host-Provided Config Actions Arguments ==========
// These are provided by the Host, not by plugins

// ConfigQueryArgs contains arguments for querying plugin config
type ConfigQueryArgs struct {
	PluginID string `json:"plugin_id"`
}

// ConfigQueryByKeyArgs contains arguments for querying plugin config by key
type ConfigQueryByKeyArgs struct {
	PluginID string `json:"plugin_id"`
	Key      string `json:"key"`
}

// ConfigListArgs contains arguments for listing all plugin configs
// Empty struct, no parameters needed
type ConfigListArgs struct{}

// ========== Helper Functions for Type-Safe Execute ==========

// MarshalParams marshals a parameter struct to json.RawMessage
func MarshalParams(v interface{}) (json.RawMessage, error) {
	return json.Marshal(v)
}

// UnmarshalParams unmarshals json.RawMessage to a parameter struct
func UnmarshalParams(data json.RawMessage, v interface{}) error {
	return json.Unmarshal(data, v)
}
