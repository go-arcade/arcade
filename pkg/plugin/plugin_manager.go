package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"github.com/observabil/arcade/pkg/log"
)

// PluginRepository 插件数据库仓库接口
type PluginRepository interface {
	GetEnabledPlugins() ([]PluginModel, error)
	GetPluginByID(pluginID string) (*PluginModel, error)
	GetDefaultPluginConfig(pluginID string) (*PluginConfigModel, error)
}

// PluginModel 插件数据模型（从数据库读取）
type PluginModel struct {
	PluginId      string
	Name          string
	Version       string
	PluginType    string
	EntryPoint    string
	ConfigSchema  json.RawMessage
	DefaultConfig json.RawMessage
	IsEnabled     int
	InstallPath   string
	Checksum      string // SHA256 校验和
}

// PluginConfigModel 插件配置数据模型
type PluginConfigModel struct {
	ConfigId    string
	PluginId    string
	ConfigItems json.RawMessage
}

type Manager struct {
	mu sync.RWMutex

	// 分类索引
	ci       map[string]CIPlugin
	cd       map[string]CDPlugin
	security map[string]SecurityPlugin
	notify   map[string]NotifyPlugin
	storage  map[string]StoragePlugin
	approval map[string]ApprovalPlugin
	custom   map[string]CustomPlugin

	// 统一索引（便于 AntiRegister / 列表）
	all map[string]BasePlugin

	// 记录来源与专有配置
	meta map[string]struct {
		Path   string
		Config any
		Type   PluginType
	}

	// 自动监控器
	watcher *Watcher
	ctx     context.Context

	// 数据库插件仓库（可选）
	pluginRepo PluginRepository
}

var (
	name        string = "plugin"
	description string = "plugin manager"
	version     string = "1.0"
)

func NewManager() *Manager {
	return &Manager{
		ci:       map[string]CIPlugin{},
		cd:       map[string]CDPlugin{},
		security: map[string]SecurityPlugin{},
		notify:   map[string]NotifyPlugin{},
		storage:  map[string]StoragePlugin{},
		approval: map[string]ApprovalPlugin{},
		custom:   map[string]CustomPlugin{},

		all: map[string]BasePlugin{},
		meta: map[string]struct {
			Path   string
			Config any
			Type   PluginType
		}{},
	}
}

func (m *Manager) Name() string {
	return name
}

func (m *Manager) Description() string {
	return description
}

func (m *Manager) Version() string {
	return version
}

// Register 注册插件：给 .so 路径
func (m *Manager) Register(path string, pluginName string, config any) error {
	inst, base, err := openAndNew(path)
	if err != nil {
		return err
	}
	name := base.Name()
	if pluginName != "" {
		name = pluginName
	}
	if name == "" {
		return fmt.Errorf("plugin %s returns empty Name()", path)
	}
	return m.registerInternal(name, base, inst, path, config, base.Type())
}

// LoadPluginsFromDir 从目录装载所有 .so 插件
func (m *Manager) LoadPluginsFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir failed: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// 只加载 .so 文件
		if !strings.HasSuffix(entry.Name(), ".so") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := m.Register(path, "", nil); err != nil {
			log.Warnf("failed to load plugin %s: %v", path, err)
		} else {
			log.Infof("loaded plugin %s", path)
		}
	}
	return nil
}

// SetPluginRepository 设置插件数据库仓库
func (m *Manager) SetPluginRepository(repo PluginRepository) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pluginRepo = repo
}

// LoadPluginsFromDatabase 从数据库加载插件
func (m *Manager) LoadPluginsFromDatabase() error {
	if m.pluginRepo == nil {
		return fmt.Errorf("plugin repository not set")
	}

	// 从数据库获取所有启用的插件
	plugins, err := m.pluginRepo.GetEnabledPlugins()
	if err != nil {
		return fmt.Errorf("failed to get plugins from database: %w", err)
	}

	log.Infof("loading %d plugins", len(plugins))

	var loadErrors []string
	successCount := 0

	// 获取程序运行目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Warnf("failed to get working directory: %v", err)
		workDir = "."
	}

	for _, dbPlugin := range plugins {
		// 使用 install_path 或 entry_point 作为插件文件路径
		pluginPath := dbPlugin.InstallPath
		if pluginPath == "" {
			pluginPath = dbPlugin.EntryPoint
		}

		// 将相对路径转换为基于程序运行目录的绝对路径
		var absPath string
		if filepath.IsAbs(pluginPath) {
			// 如果已经是绝对路径，直接使用
			absPath = pluginPath
		} else {
			// 相对路径，基于程序运行目录
			absPath = filepath.Join(workDir, pluginPath)
		}

		// 安全检查：验证插件文件校验和
		if dbPlugin.Checksum != "" {
			if err := verifyPluginChecksum(absPath, dbPlugin.Checksum); err != nil {
				errMsg := fmt.Sprintf("checksum verification failed for plugin %s: %v", dbPlugin.PluginId, err)
				log.Error(errMsg)
				loadErrors = append(loadErrors, errMsg)
				continue
			}
			log.Infof("checksum verified for plugin %s", dbPlugin.PluginId)
		} else {
			log.Warnf("plugin %s has no checksum, skipping verification (security risk!)", dbPlugin.PluginId)
		}

		// 解析插件配置
		var config interface{}
		if len(dbPlugin.DefaultConfig) > 0 {
			if err := json.Unmarshal(dbPlugin.DefaultConfig, &config); err != nil {
				log.Warnf("failed to unmarshal default config for plugin %s: %v", dbPlugin.PluginId, err)
			}
		}

		// 尝试加载插件
		if err := m.Register(absPath, dbPlugin.PluginId, config); err != nil {
			errMsg := fmt.Sprintf("failed to load plugin %s from %s: %v", dbPlugin.PluginId, absPath, err)
			log.Warn(errMsg)
			loadErrors = append(loadErrors, errMsg)
			continue
		}

		log.Infof("loaded plugin %s (v%s): %s (relative: %s)", dbPlugin.PluginId, dbPlugin.Version, absPath, pluginPath)
		successCount++
	}

	log.Infof("successfully loaded %d/%d plugins", successCount, len(plugins))

	if len(loadErrors) > 0 {
		log.Warnf("failed to load %d plugins:", len(loadErrors))
		for _, errMsg := range loadErrors {
			log.Warn("  - " + errMsg)
		}
	}

	return nil
}

// Init 所有插件（传入各自的专有 config）
func (m *Manager) Init(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, p := range m.all {
		meta := m.meta[name]
		if err := p.Init(ctx, meta.Config); err != nil {
			return fmt.Errorf("init plugin %s (%s): %w", name, meta.Path, err)
		}
	}
	return nil
}

// Cleanup
func (m *Manager) Cleanup() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var firstErr error
	for name, p := range m.all {
		if err := p.Cleanup(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("cleanup %s: %w", name, err)
		}
	}
	return firstErr
}

// AntiRegister 只删除已存在的那个；Go 的 plugin 不能卸载 .so，本函数仅移除索引
func (m *Manager) AntiRegister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.all[name]; !ok {
		return fmt.Errorf("plugin %q not found", name)
	}

	switch m.meta[name].Type {
	case TypeCI:
		delete(m.ci, name)
	case TypeCD:
		delete(m.cd, name)
	case TypeSecurity:
		delete(m.security, name)
	case TypeNotify:
		delete(m.notify, name)
	case TypeStorage:
		delete(m.storage, name)
	case TypeApproval:
		delete(m.approval, name)
	case TypeCustom:
		delete(m.custom, name)
	}
	delete(m.all, name)
	delete(m.meta, name)
	return nil
}

// GetXXXPlugin 按类型获取插件实例
func (m *Manager) GetCIPlugin(name string) (CIPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.ci[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("CI plugin %q not found", name)
}
func (m *Manager) GetCDPlugin(name string) (CDPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.cd[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("CD plugin %q not found", name)
}
func (m *Manager) GetSecurityPlugin(name string) (SecurityPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.security[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("security plugin %q not found", name)
}

func (m *Manager) GetNotifyPlugin(name string) (NotifyPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.notify[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("notify plugin %q not found", name)
}
func (m *Manager) GetStoragePlugin(name string) (StoragePlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.storage[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("storage plugin %q not found", name)
}
func (m *Manager) GetApprovalPlugin(name string) (ApprovalPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.approval[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("approval plugin %q not found", name)
}
func (m *Manager) GetCustomPlugin(name string) (CustomPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.custom[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("custom plugin %q not found", name)
}

// registerInternal 内部注册逻辑
func (m *Manager) registerInternal(name string, base BasePlugin, _ any, path string, config any, pluginType PluginType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.all[name]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}
	switch base.Type() {
	case TypeCI:
		if p, ok := base.(CIPlugin); ok {
			m.ci[name] = p
		} else {
			return fmt.Errorf("%s claims CI but not implements CIPlugin", name)
		}
	case TypeCD:
		if p, ok := base.(CDPlugin); ok {
			m.cd[name] = p
		} else {
			return fmt.Errorf("%s claims CD but not implements CDPlugin", name)
		}
	case TypeSecurity:
		if p, ok := base.(SecurityPlugin); ok {
			m.security[name] = p
		} else {
			return fmt.Errorf("%s claims Security but not implements SecurityPlugin", name)
		}
	case TypeNotify:
		if p, ok := base.(NotifyPlugin); ok {
			m.notify[name] = p
		} else {
			return fmt.Errorf("%s claims Notify but not implements NotifyPlugin", name)
		}
	case TypeStorage:
		if p, ok := base.(StoragePlugin); ok {
			m.storage[name] = p
		} else {
			return fmt.Errorf("%s claims Storage but not implements StoragePlugin", name)
		}
	case TypeApproval:
		if p, ok := base.(ApprovalPlugin); ok {
			m.approval[name] = p
		} else {
			return fmt.Errorf("%s claims Approval but not implements ApprovalPlugin", name)
		}
	case TypeCustom:
		if p, ok := base.(CustomPlugin); ok {
			m.custom[name] = p
		} else {
			return fmt.Errorf("%s claims Custom but not implements CustomPlugin", name)
		}
	default:
		return fmt.Errorf("unknown plugin type %q from %s", base.Type(), name)
	}

	// 统一索引与元数据
	m.all[name] = base
	m.meta[name] = struct {
		Path   string
		Config any
		Type   PluginType
	}{
		Path:   filepath.Clean(path),
		Config: config,
		Type:   pluginType,
	}

	return nil
}

// openAndNew 支持多种导出风格：
// 1) var Plugin <interface>
func openAndNew(path string) (any, BasePlugin, error) {
	so, err := plugin.Open(path)
	if err != nil {
		return nil, nil, err
	}
	if sym, err := so.Lookup("Plugin"); err == nil {
		if bp, ok := sym.(BasePlugin); ok {
			return sym, bp, nil
		}
		// 有些人会导出为具体类型实例 any
		if anyInst, ok := sym.(any); ok {
			if bp, ok2 := anyInst.(BasePlugin); ok2 {
				return anyInst, bp, nil
			}
		}
	}

	// 兼容工厂方法 func NewPlugin() ...
	sym, err := so.Lookup("NewPlugin")
	if err != nil {
		return nil, nil, errors.New("neither symbol Plugin/BasePlugin nor function NewPlugin found")
	}

	// func() BasePlugin
	if fn, ok := sym.(func() BasePlugin); ok {
		return fn, fn(), nil
	}
	// func() any （再断言）
	if fn, ok := sym.(func() any); ok {
		v := fn()
		if bp, ok2 := v.(BasePlugin); ok2 {
			return fn, bp, nil
		}
		return fn, nil, errors.New("NewPlugin() returns any but not BasePlugin")
	}

	return nil, nil, errors.New("NewPlugin has unsupported signature")
}

// ListPlugins 列出所有已注册插件（带类型/路径/版本）
type PluginInfo struct {
	Name        string
	Type        PluginType
	Version     string
	Path        string
	Description string
}

func (m *Manager) ListPlugins() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]PluginInfo, 0, len(m.all))
	for name, p := range m.all {
		meta := m.meta[name]
		out = append(out, PluginInfo{
			Name:        name,
			Type:        meta.Type,
			Version:     p.Version(),
			Path:        meta.Path,
			Description: p.Description(),
		})
	}
	return out
}

// StartAutoWatch 启动自动监控
func (m *Manager) StartAutoWatch(dirs []string) error {
	if m.watcher != nil {
		return fmt.Errorf("watcher already started")
	}

	watcher, err := NewWatcher(m)
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	// 添加要监控的目录
	for _, dir := range dirs {
		if err := watcher.AddWatchDir(dir); err != nil {
			watcher.Stop()
			return fmt.Errorf("add watch dir %s: %w", dir, err)
		}
	}

	m.watcher = watcher
	watcher.Start()

	log.Info("plugin auto watch started")
	return nil
}

// StopAutoWatch 停止自动监控
func (m *Manager) StopAutoWatch() {
	if m.watcher != nil {
		m.watcher.Stop()
		m.watcher = nil
	}
}

// ReloadPlugin 热重载指定插件
func (m *Manager) ReloadPlugin(name string) error {
	m.mu.RLock()
	meta, exists := m.meta[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %q not found", name)
	}

	// 先调用清理
	if p, ok := m.all[name]; ok {
		if err := p.Cleanup(); err != nil {
			log.Warnf("cleanup plugin %s: %v", name, err)
		}
	}

	// 卸载旧插件
	if err := m.AntiRegister(name); err != nil {
		return fmt.Errorf("unregister plugin %s: %w", name, err)
	}

	log.Infof("plugin %s unloaded", name)

	// 重新加载插件
	if err := m.Register(meta.Path, name, meta.Config); err != nil {
		return fmt.Errorf("reload plugin %s from %s: %w", name, meta.Path, err)
	}

	// 初始化插件
	if m.ctx == nil {
		m.ctx = context.Background()
	}
	if err := m.Init(m.ctx); err != nil {
		return fmt.Errorf("init reloaded plugin %s: %w", name, err)
	}

	log.Infof("plugin %s reloaded successfully", name)
	return nil
}

// SetContext 设置全局上下文
func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// verifyPluginChecksum 验证插件文件的SHA256校验和
func verifyPluginChecksum(pluginPath string, expectedChecksum string) error {
	if expectedChecksum == "" {
		return fmt.Errorf("checksum not provided")
	}

	// 打开插件文件
	file, err := os.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin file: %w", err)
	}
	defer file.Close()

	// 计算 SHA256
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	// 比较校验和
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// CalculatePluginChecksum 计算插件文件的SHA256校验和（用于生成checksum）
func CalculatePluginChecksum(pluginPath string) (string, error) {
	file, err := os.Open(pluginPath)
	if err != nil {
		return "", fmt.Errorf("failed to open plugin file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
