// Package plugin type definitions
package plugin

import "encoding/json"

// PluginType is the plugin type enumeration
type PluginType string

const (
	// Source plugin type
	TypeSource PluginType = "source"
	// Build plugin type
	TypeBuild PluginType = "build"
	// Test plugin type
	TypeTest PluginType = "test"
	// Deploy plugin type
	TypeDeploy PluginType = "deploy"
	// Security plugin type
	TypeSecurity PluginType = "security"
	// Notify plugin type
	TypeNotify PluginType = "notify"
	// Approval plugin type
	TypeApproval PluginType = "approval"
	// Storage plugin type
	TypeStorage PluginType = "storage"
	// Analytics plugin type
	TypeAnalytics PluginType = "analytics"
	// Integration plugin type
	TypeIntegration PluginType = "integration"
	// Custom plugin type
	TypeCustom PluginType = "custom"
)

// PluginConfig represents the plugin configuration
type PluginConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Type        string            `json:"type"`
	Config      json.RawMessage   `json:"config"`
	Environment map[string]string `json:"environment"`
	TaskID      string            `json:"task_id"` // 任务ID，用于日志关联
	LogHandlers []LogHandler      `json:"-"`       // 日志处理器（不序列化）
}

// PluginInfo contains plugin information
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Homepage    string `json:"homepage,omitempty"`
}

// PluginMetrics contains plugin runtime metrics
type PluginMetrics struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Version       string                 `json:"version"`
	Status        string                 `json:"status"`
	Uptime        int64                  `json:"uptime"`
	CallCount     int64                  `json:"call_count"`
	ErrorCount    int64                  `json:"error_count"`
	LastError     string                 `json:"last_error,omitempty"`
	LastCallTime  int64                  `json:"last_call_time"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
}

// RPCError represents an RPC error structure
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// Error implements the error interface
func (e *RPCError) Error() string {
	return e.Message
}

// AllPluginTypes returns all supported plugin types
func AllPluginTypes() []PluginType {
	return []PluginType{
		TypeSource,
		TypeBuild,
		TypeTest,
		TypeDeploy,
		TypeSecurity,
		TypeNotify,
		TypeApproval,
		TypeStorage,
		TypeAnalytics,
		TypeIntegration,
		TypeCustom,
	}
}

// IsValidPluginType checks if a plugin type is valid
func IsValidPluginType(t string) bool {
	for _, validType := range AllPluginTypes() {
		if string(validType) == t {
			return true
		}
	}
	return false
}

// GetPluginTypeDescription returns a description for a plugin type
func GetPluginTypeDescription(t PluginType) string {
	descriptions := map[PluginType]string{
		TypeSource:      "源码管理插件，用于代码仓库操作（克隆、拉取、分支切换等）",
		TypeBuild:       "构建插件，用于编译和构建项目（编译、打包、生成产物等）",
		TypeTest:        "测试插件，用于运行测试和生成报告（单元测试、集成测试、覆盖率等）",
		TypeDeploy:      "部署插件，用于应用部署和管理（部署、回滚、扩缩容等）",
		TypeSecurity:    "安全插件，用于安全扫描和审计（漏洞扫描、合规检查等）",
		TypeNotify:      "通知插件，用于发送各类通知（邮件、Webhook、即时消息等）",
		TypeApproval:    "审批插件，用于审批流程管理（创建审批、批准、拒绝等）",
		TypeStorage:     "存储插件，用于数据存储和管理（保存、加载、删除、列表等）",
		TypeAnalytics:   "分析插件，用于数据分析和报告（事件追踪、查询、指标、报告等）",
		TypeIntegration: "集成插件，用于第三方服务集成（连接、调用、订阅等）",
		TypeCustom:      "自定义插件，用于特殊用途的定制化功能",
	}
	return descriptions[t]
}

// String returns the string representation of PluginType
func (pt PluginType) String() string {
	return string(pt)
}

// Validate validates the PluginType
func (pt PluginType) Validate() error {
	if !IsValidPluginType(string(pt)) {
		return &RPCError{
			Code:    400,
			Message: "invalid plugin type",
			Data:    string(pt),
		}
	}
	return nil
}
