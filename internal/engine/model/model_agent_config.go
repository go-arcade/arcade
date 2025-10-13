package model

import "github.com/observabil/arcade/pkg/datatype"

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: model_agent_config.go
 * @description: agent config model
 */

// AgentConfig Agent配置表（每个Agent一条记录）
type AgentConfig struct {
	BaseModel
	AgentId     string        `gorm:"column:agent_id" json:"agentId"`
	ConfigItems datatype.JSON `gorm:"column:config_items;type:json" json:"configItems"` // 所有配置在一个JSON中
	Description string        `gorm:"column:description" json:"description"`
}

func (AgentConfig) TableName() string {
	return "t_agent_config"
}

// AgentConfigItems Agent 配置项结构（解析 config_items 的结构）
type AgentConfigItems struct {
	HeartbeatInterval int               `json:"heartbeat_interval"`     // 心跳间隔(秒)
	MaxConcurrentJobs int               `json:"max_concurrent_jobs"`    // 最大并发任务数
	JobTimeout        int               `json:"job_timeout"`            // 任务超时时间(秒)
	WorkspaceDir      string            `json:"workspace_dir"`          // 工作目录
	TempDir           string            `json:"temp_dir"`               // 临时目录
	LogLevel          string            `json:"log_level"`              // 日志级别
	EnableDocker      bool              `json:"enable_docker"`          // 是否启用Docker
	DockerNetwork     string            `json:"docker_network"`         // Docker网络模式
	ResourceLimits    ResourceLimits    `json:"resource_limits"`        // 资源限制
	AllowedCommands   []string          `json:"allowed_commands"`       // 允许执行的命令白名单
	EnvVars           map[string]string `json:"env_vars"`               // 环境变量
	SSHKey            string            `json:"ssh_key,omitempty"`      // SSH私钥(加密)
	SSHPassword       string            `json:"ssh_password,omitempty"` // SSH密码(加密)
	ProxyURL          string            `json:"proxy_url,omitempty"`    // 代理地址
	CacheDir          string            `json:"cache_dir"`              // 缓存目录
	CleanupPolicy     CleanupPolicy     `json:"cleanup_policy"`         // 清理策略
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	CPU    string `json:"cpu"`    // CPU限制，如 "2" 表示2核
	Memory string `json:"memory"` // 内存限制，如 "4G"
}

// CleanupPolicy 清理策略
type CleanupPolicy struct {
	MaxAgeDays int `json:"max_age_days"` // 最大保留天数
	MaxSizeGB  int `json:"max_size_gb"`  // 最大占用空间(GB)
}

// AgentConfigKey Agent 配置键常量
const (
	AgentConfigHeartbeatInterval = "heartbeat_interval"  // 心跳间隔(秒)
	AgentConfigMaxConcurrentJobs = "max_concurrent_jobs" // 最大并发任务数
	AgentConfigJobTimeout        = "job_timeout"         // 任务超时时间(秒)
	AgentConfigWorkspaceDir      = "workspace_dir"       // 工作目录
	AgentConfigTempDir           = "temp_dir"            // 临时目录
	AgentConfigLogLevel          = "log_level"           // 日志级别
	AgentConfigEnableDocker      = "enable_docker"       // 是否启用Docker
	AgentConfigDockerNetwork     = "docker_network"      // Docker网络模式
	AgentConfigResourceLimits    = "resource_limits"     // 资源限制(JSON)
	AgentConfigAllowedCommands   = "allowed_commands"    // 允许执行的命令白名单(JSON)
	AgentConfigEnvVars           = "env_vars"            // 环境变量(JSON)
	AgentConfigSSHKey            = "ssh_key"             // SSH私钥(加密)
	AgentConfigSSHPassword       = "ssh_password"        // SSH密码(加密)
	AgentConfigProxyURL          = "proxy_url"           // 代理地址
	AgentConfigCacheDir          = "cache_dir"           // 缓存目录
	AgentConfigCleanupPolicy     = "cleanup_policy"      // 清理策略(JSON)
)
