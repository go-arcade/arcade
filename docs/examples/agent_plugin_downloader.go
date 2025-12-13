package examples

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	"github.com/go-arcade/arcade/pkg/log"
)

// PluginDownloader Agent端插件下载管理器
type PluginDownloader struct {
	agentClient agentv1.AgentServiceClient
	agentID     string
	pluginDir   string
	cache       map[string]*CachedPlugin // key: pluginID@version
	mu          sync.RWMutex
	maxRetries  int
	timeout     time.Duration
}

// CachedPlugin 缓存的插件信息
type CachedPlugin struct {
	PluginID string
	Version  string
	Path     string
	Checksum string
	Size     int64
	CachedAt int64
}

// NewPluginDownloader 创建插件下载器
func NewPluginDownloader(agentClient agentv1.AgentServiceClient, agentID string, pluginDir string) *PluginDownloader {
	return &PluginDownloader{
		agentClient: agentClient,
		agentID:     agentID,
		pluginDir:   pluginDir,
		cache:       make(map[string]*CachedPlugin),
		maxRetries:  3,
		timeout:     5 * time.Minute,
	}
}

// Init 初始化插件目录并扫描已缓存的插件
func (d *PluginDownloader) Init() error {
	log.Info("[PluginDownloader] initializing plugin downloader")

	// 创建目录结构
	dirs := []string{
		filepath.Join(d.pluginDir, "builtin"),    // 预装插件
		filepath.Join(d.pluginDir, "downloaded"), // 下载的插件
		filepath.Join(d.pluginDir, "temp"),       // 临时文件
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}

	// 扫描已缓存的插件
	if err := d.scanCachedPlugins(); err != nil {
		log.Warn("[PluginDownloader] failed to scan cached plugins: %v", err)
		// 继续执行，不返回错误
	}

	log.Info("[PluginDownloader] initialization completed, cached plugins: %d", len(d.cache))
	return nil
}

// scanCachedPlugins 扫描本地缓存的插件
func (d *PluginDownloader) scanCachedPlugins() error {
	// 扫描builtin目录
	if err := d.scanDirectory(filepath.Join(d.pluginDir, "builtin"), "builtin"); err != nil {
		log.Warn("[PluginDownloader] failed to scan builtin plugins: %v", err)
	}

	// 扫描downloaded目录
	if err := d.scanDirectory(filepath.Join(d.pluginDir, "downloaded"), "downloaded"); err != nil {
		log.Warn("[PluginDownloader] failed to scan downloaded plugins: %v", err)
	}

	return nil
}

// scanDirectory 扫描指定目录
func (d *PluginDownloader) scanDirectory(dir string, source string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".so" {
			continue
		}

		pluginPath := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			log.Warn("[PluginDownloader] failed to get file info: %v", err)
			continue
		}

		// 计算校验和
		checksum, err := calculateChecksumFromFile(pluginPath)
		if err != nil {
			log.Warn("[PluginDownloader] failed to calculate checksum for %s: %v", pluginPath, err)
			continue
		}

		// 解析插件ID和版本
		// 假设文件名格式：pluginID@version.so 或 pluginID.so
		name := entry.Name()[:len(entry.Name())-3] // 去掉.so

		cached := &CachedPlugin{
			Path:     pluginPath,
			Checksum: checksum,
			Size:     info.Size(),
			CachedAt: info.ModTime().Unix(),
		}

		// 尝试解析版本
		// 如果文件名包含@，则分割为 pluginID@version
		// 否则整个文件名就是pluginID
		cacheKey := name

		d.mu.Lock()
		d.cache[cacheKey] = cached
		d.mu.Unlock()

		log.Info("[PluginDownloader] found cached plugin: %s (source=%s, checksum=%s...)",
			name, source, checksum[:8])
	}

	return nil
}

// EnsurePlugins 确保所有需要的插件都已就绪
func (d *PluginDownloader) EnsurePlugins(ctx context.Context, plugins []*agentv1.PluginInfo) ([]string, error) {
	if len(plugins) == 0 {
		return []string{}, nil
	}

	log.Info("[PluginDownloader] ensuring %d plugins are ready", len(plugins))

	pluginPaths := make([]string, 0, len(plugins))

	for _, plugin := range plugins {
		path, err := d.ensurePlugin(ctx, plugin)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure plugin %s: %w", plugin.PluginId, err)
		}
		pluginPaths = append(pluginPaths, path)
	}

	log.Info("[PluginDownloader] all %d plugins are ready", len(pluginPaths))
	return pluginPaths, nil
}

// ensurePlugin 确保单个插件就绪
func (d *PluginDownloader) ensurePlugin(ctx context.Context, plugin *agentv1.PluginInfo) (string, error) {
	cacheKey := fmt.Sprintf("%s@%s", plugin.PluginId, plugin.Version)

	// 检查缓存
	d.mu.RLock()
	cached, exists := d.cache[cacheKey]
	d.mu.RUnlock()

	if exists {
		// 验证校验和
		if cached.Checksum == plugin.Checksum {
			log.Info("[PluginDownloader] plugin %s already cached at %s", cacheKey, cached.Path)
			return cached.Path, nil
		}
		log.Warn("[PluginDownloader] plugin %s checksum mismatch (cached=%s, expected=%s), re-downloading",
			cacheKey, cached.Checksum[:8], plugin.Checksum[:8])
	}

	// 下载插件（带重试）
	log.Info("[PluginDownloader] downloading plugin %s (v%s)", plugin.PluginId, plugin.Version)
	return d.downloadPluginWithRetry(ctx, plugin)
}

// downloadPluginWithRetry 下载插件（带重试）
func (d *PluginDownloader) downloadPluginWithRetry(ctx context.Context, plugin *agentv1.PluginInfo) (string, error) {
	var lastErr error

	for i := 0; i < d.maxRetries; i++ {
		if i > 0 {
			log.Info("[PluginDownloader] retry downloading plugin %s (attempt %d/%d)",
				plugin.PluginId, i+1, d.maxRetries)
			// 指数退避
			time.Sleep(time.Second * time.Duration(1<<uint(i)))
		}

		path, err := d.downloadPlugin(ctx, plugin)
		if err == nil {
			return path, nil
		}

		lastErr = err
		log.Warn("[PluginDownloader] download attempt %d failed: %v", i+1, err)
	}

	return "", fmt.Errorf("download failed after %d retries: %w", d.maxRetries, lastErr)
}

// downloadPlugin 从服务端下载插件
func (d *PluginDownloader) downloadPlugin(ctx context.Context, plugin *agentv1.PluginInfo) (string, error) {
	// 创建带超时的上下文
	downloadCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	// 调用gRPC下载插件
	resp, err := d.agentClient.DownloadPlugin(downloadCtx, &agentv1.DownloadPluginRequest{
		AgentId:  d.agentID,
		PluginId: plugin.PluginId,
		Version:  plugin.Version,
	})
	if err != nil {
		return "", fmt.Errorf("grpc call failed: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("download failed: %s", resp.Message)
	}

	// 验证校验和
	actualChecksum := calculateChecksumFromBytes(resp.PluginData)
	if actualChecksum != resp.Checksum {
		return "", fmt.Errorf("checksum verification failed: expected %s, got %s",
			resp.Checksum, actualChecksum)
	}

	// 额外验证（如果plugin中也有checksum）
	if plugin.Checksum != "" && actualChecksum != plugin.Checksum {
		return "", fmt.Errorf("checksum mismatch with plugin info: expected %s, got %s",
			plugin.Checksum, actualChecksum)
	}

	// 保存到本地
	filename := fmt.Sprintf("%s@%s.so", plugin.PluginId, plugin.Version)
	pluginPath := filepath.Join(d.pluginDir, "downloaded", filename)

	// 先写入临时文件
	tempPath := filepath.Join(d.pluginDir, "temp", filename+".tmp")
	if err := os.WriteFile(tempPath, resp.PluginData, 0755); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// 重命名为最终文件（原子操作）
	if err := os.Rename(tempPath, pluginPath); err != nil {
		os.Remove(tempPath) // 清理临时文件
		return "", fmt.Errorf("failed to rename plugin file: %w", err)
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("%s@%s", plugin.PluginId, plugin.Version)
	d.mu.Lock()
	d.cache[cacheKey] = &CachedPlugin{
		PluginID: plugin.PluginId,
		Version:  plugin.Version,
		Path:     pluginPath,
		Checksum: resp.Checksum,
		Size:     resp.Size,
		CachedAt: time.Now().Unix(),
	}
	d.mu.Unlock()

	log.Info("[PluginDownloader] plugin %s (v%s) downloaded successfully: %s (size=%d bytes)",
		plugin.PluginId, plugin.Version, pluginPath, resp.Size)

	return pluginPath, nil
}

// GetInstalledPlugins 获取已安装插件列表（用于Agent注册时上报）
func (d *PluginDownloader) GetInstalledPlugins() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	installed := make([]string, 0, len(d.cache))
	for key := range d.cache {
		installed = append(installed, key)
	}

	return installed
}

// GetPluginPath 获取插件路径
func (d *PluginDownloader) GetPluginPath(pluginID, version string) (string, bool) {
	cacheKey := fmt.Sprintf("%s@%s", pluginID, version)

	d.mu.RLock()
	defer d.mu.RUnlock()

	if cached, exists := d.cache[cacheKey]; exists {
		return cached.Path, true
	}

	return "", false
}

// CleanCache 清理缓存（可选实现，用于定期清理旧版本插件）
func (d *PluginDownloader) CleanCache(maxAge time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().Unix()
	cleaned := 0

	for key, cached := range d.cache {
		// 只清理downloaded目录中的插件，保留builtin
		if filepath.Dir(cached.Path) != filepath.Join(d.pluginDir, "downloaded") {
			continue
		}

		age := now - cached.CachedAt
		if age > int64(maxAge.Seconds()) {
			// 删除文件
			if err := os.Remove(cached.Path); err != nil {
				log.Warn("[PluginDownloader] failed to remove old plugin %s: %v", key, err)
				continue
			}

			// 从缓存中删除
			delete(d.cache, key)
			cleaned++
			log.Info("[PluginDownloader] cleaned old plugin: %s (age=%d seconds)", key, age)
		}
	}

	if cleaned > 0 {
		log.Info("[PluginDownloader] cache cleanup completed, removed %d plugins", cleaned)
	}

	return nil
}

// ListAvailablePlugins 从服务端查询可用插件列表
func (d *PluginDownloader) ListAvailablePlugins(ctx context.Context, pluginType string) ([]*agentv1.PluginInfo, error) {
	resp, err := d.agentClient.ListAvailablePlugins(ctx, &agentv1.ListAvailablePluginsRequest{
		AgentId:    d.agentID,
		PluginType: pluginType,
	})

	if err != nil {
		return nil, fmt.Errorf("grpc call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("list failed: %s", resp.Message)
	}

	return resp.Plugins, nil
}

// 辅助函数

// calculateChecksumFromFile 从文件计算校验和
func calculateChecksumFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return calculateChecksumFromBytes(data), nil
}

// calculateChecksumFromBytes 从字节数组计算校验和
func calculateChecksumFromBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
