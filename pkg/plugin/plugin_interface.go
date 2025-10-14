package plugin

import "context"

type Option func(*Options)
type Options struct {
	// ctx context.Context
}

// PluginType 插件类型
type PluginType string

const (
	// TypeCI CI类型插件
	TypeCI PluginType = "ci"
	// TypeCD CD类型插件
	TypeCD PluginType = "cd"
	// TypeSecurity 安全类型插件
	TypeSecurity PluginType = "security"
	// TypeNotify 通知类型插件
	TypeNotify PluginType = "notify"
	// TypeStorage 存储类型插件
	TypeStorage PluginType = "storage"
	// TypeApproval 审批类型插件
	TypeApproval PluginType = "approval"
	// TypeCustom 自定义类型插件
	TypeCustom PluginType = "custom"
)

// BasePlugin 基础插件接口
type BasePlugin interface {
	// Name plugin name
	Name() string
	// Description plugin description
	Description() string
	// Version plugin version
	Version() string
	// Type plugin type
	Type() PluginType
	// Init plugin init
	Init(ctx context.Context, config any) error
	// Cleanup plugin cleanup
	Cleanup() error
}

// CIPlugin CI类型插件接口
type CIPlugin interface {
	BasePlugin
	// Build 构建项目
	Build(ctx context.Context, projectConfig any, opts ...Option) error
	// Test 运行测试
	Test(ctx context.Context, projectConfig any, opts ...Option) error
	// Lint 代码检查
	Lint(ctx context.Context, projectConfig any, opts ...Option) error
}

// CDPlugin CD类型插件接口
type CDPlugin interface {
	BasePlugin
	// Deploy 部署应用
	Deploy(ctx context.Context, projectConfig any, opts ...Option) error
	// Rollback 回滚部署
	Rollback(ctx context.Context, projectConfig any, opts ...Option) error
}

// SecurityPlugin 安全类型插件接口
type SecurityPlugin interface {
	BasePlugin
	// Scan 安全扫描
	Scan(ctx context.Context, projectConfig any, opts ...Option) error
	// Audit 安全审计
	Audit(ctx context.Context, projectConfig any, opts ...Option) error
}

// NotifyPlugin 通知类型插件接口
type NotifyPlugin interface {
	BasePlugin
	// Send 发送通知
	Send(ctx context.Context, message any, opts ...Option) error
	// SendTemplate 发送模板通知
	SendTemplate(ctx context.Context, template string, data any, opts ...Option) error
}

// StoragePlugin 存储类型插件接口
type StoragePlugin interface {
	BasePlugin
	// Save 保存数据
	Save(ctx context.Context, key string, data any, opts ...Option) error
	// Load 加载数据
	Load(ctx context.Context, key string, opts ...Option) (any, error)
	// Delete 删除数据
	Delete(ctx context.Context, key string, opts ...Option) error
}

// ApprovalPlugin 审批类型插件接口
type ApprovalPlugin interface {
	BasePlugin
	// Approve 审批
	Approve(ctx context.Context, projectConfig any, opts ...Option) error
}

// CustomPlugin 自定义类型插件接口
type CustomPlugin interface {
	BasePlugin
	// Execute 执行自定义操作
	Execute(ctx context.Context, action string, params any, opts ...Option) (any, error)
}
