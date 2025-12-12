package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	v1 "github.com/go-arcade/arcade/api/agent/v1"
	pluginmodel "github.com/go-arcade/arcade/internal/engine/model"
	pluginrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/log"
)

type PluginDownloadService struct {
	pluginRepo pluginrepo.IPluginRepository
}

func NewPluginDownloadService(pluginRepo pluginrepo.IPluginRepository) *PluginDownloadService {
	return &PluginDownloadService{
		pluginRepo: pluginRepo,
	}
}

// GetPluginForDownload 获取插件用于下载
func (s *PluginDownloadService) GetPluginForDownload(pluginID, version string) (*v1.DownloadPluginResponse, error) {
	log.Infow("[PluginDownload] requesting plugin", "pluginId", pluginID, "version", version)

	// 1. 从数据库获取插件信息
	plugin, err := s.pluginRepo.GetPluginByID(pluginID)
	if err != nil {
		log.Errorw("[PluginDownload] plugin not found", "pluginId", pluginID, "error", err)
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: fmt.Sprintf("plugin not found: %v", err),
		}, nil
	}

	// 2. 检查插件是否启用
	if plugin.IsEnabled != 1 {
		log.Warnw("[PluginDownload] plugin is disabled", "pluginId", pluginID)
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: "plugin is disabled",
		}, nil
	}

	// 3. 检查版本（如果指定了版本）
	if version != "" && plugin.Version != version {
		log.Warnw("[PluginDownload] version mismatch", "pluginId", pluginID, "requested", version, "available", plugin.Version)
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: fmt.Sprintf("version mismatch: requested %s, available %s", version, plugin.Version),
		}, nil
	}

	// 4. 获取插件文件路径（动态生成）
	pluginPath := s.getPluginLocalPath(pluginID, plugin.Version)

	// 5. 读取插件文件
	pluginData, err := os.ReadFile(pluginPath)
	if err != nil {
		log.Errorw("[PluginDownload] failed to read plugin file", "pluginId", pluginID, "pluginPath", pluginPath, "error", err)
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: fmt.Sprintf("failed to read plugin file: %v", err),
		}, nil
	}

	// 6. 计算校验和
	hash := sha256.Sum256(pluginData)
	checksum := hex.EncodeToString(hash[:])

	// 7. 验证校验和（如果数据库中有记录）
	if plugin.Checksum != "" && plugin.Checksum != checksum {
		log.Warnw("[PluginDownload] checksum mismatch", "pluginId", pluginID, "dbChecksum", plugin.Checksum, "fileChecksum", checksum)
		// 更新数据库中的校验和
		// TODO: 考虑是否要自动更新或者返回错误
	}

	log.Infow("[PluginDownload] plugin ready for download", "pluginId", pluginID, "version", plugin.Version, "size", len(pluginData), "checksum", checksum[:8])

	return &v1.DownloadPluginResponse{
		Success:    true,
		Message:    "plugin downloaded successfully",
		PluginData: pluginData,
		Checksum:   checksum,
		Size:       int64(len(pluginData)),
		Version:    plugin.Version,
	}, nil
}

// ListAvailablePlugins 列出可用插件
func (s *PluginDownloadService) ListAvailablePlugins(pluginType string) ([]*v1.PluginInfo, error) {
	var plugins []pluginmodel.Plugin
	var err error

	if pluginType != "" {
		log.Infow("[PluginDownload] listing plugins of type", "pluginType", pluginType)
		plugins, err = s.pluginRepo.GetPluginsByType(pluginType)
	} else {
		log.Info("[PluginDownload] listing all enabled plugins")
		plugins, err = s.pluginRepo.GetEnabledPlugins()
	}

	if err != nil {
		log.Errorw("[PluginDownload] failed to list plugins", "pluginType", pluginType, "error", err)
		return nil, err
	}

	result := make([]*v1.PluginInfo, 0, len(plugins))
	for _, p := range plugins {
		// 计算文件大小（动态生成路径）
		localPath := s.getPluginLocalPath(p.PluginId, p.Version)
		size := s.getPluginFileSize(localPath)

		result = append(result, &v1.PluginInfo{
			PluginId: p.PluginId,
			Name:     p.Name,
			Version:  p.Version,
			Checksum: p.Checksum,
			Size:     size,
			Location: v1.PluginLocation_PLUGIN_LOCATION_SERVER,
		})
	}

	log.Infow("[PluginDownload] found available plugins", "count", len(result))
	return result, nil
}

// GetPluginsForTask 获取任务所需的插件列表
func (s *PluginDownloadService) GetPluginsForTask(taskID string) ([]*v1.PluginInfo, error) {
	log.Infow("[PluginDownload] getting plugins for task", "taskId", taskID)

	// 从任务插件关联表获取
	taskPlugins, err := s.pluginRepo.GetTaskPlugins(taskID)
	if err != nil {
		log.Errorw("[PluginDownload] failed to get task plugins", "taskId", taskID, "error", err)
		return nil, err
	}

	if len(taskPlugins) == 0 {
		log.Infow("[PluginDownload] no plugins required for task", "taskId", taskID)
		return []*v1.PluginInfo{}, nil
	}

	result := make([]*v1.PluginInfo, 0, len(taskPlugins))
	for _, tp := range taskPlugins {
		plugin, err := s.pluginRepo.GetPluginByID(tp.PluginId)
		if err != nil {
			log.Warnw("[PluginDownload] failed to get plugin", "pluginId", tp.PluginId, "error", err)
			continue
		}

		// 跳过禁用的插件
		if plugin.IsEnabled != 1 {
			log.Warnw("[PluginDownload] skipping disabled plugin", "pluginId", plugin.PluginId)
			continue
		}

		// 计算文件大小（动态生成路径）
		localPath := s.getPluginLocalPath(plugin.PluginId, plugin.Version)
		size := s.getPluginFileSize(localPath)

		result = append(result, &v1.PluginInfo{
			PluginId: plugin.PluginId,
			Name:     plugin.Name,
			Version:  plugin.Version,
			Checksum: plugin.Checksum,
			Size:     size,
			Location: v1.PluginLocation_PLUGIN_LOCATION_SERVER,
		})
	}

	log.Infow("[PluginDownload] task requires plugins", "taskId", taskID, "count", len(result))
	return result, nil
}

// CheckPluginPermission 检查Agent是否有权限下载指定插件（可选实现）
func (s *PluginDownloadService) CheckPluginPermission(agentID string, pluginID string) error {
	// TODO: 实现权限检查逻辑
	// 可以根据Agent的标签、组织、项目等进行权限判断
	// 目前暂时允许所有Agent下载所有插件
	return nil
}

// getPluginLocalPath 获取插件本地缓存路径（动态生成）
func (s *PluginDownloadService) getPluginLocalPath(pluginID, version string) string {
	// TODO: 从配置中获取本地缓存目录
	localCacheDir := "/var/lib/arcade/plugins"
	filename := fmt.Sprintf("%s_%s.so", pluginID, version)
	return filepath.Join(localCacheDir, filename)
}

// getPluginFileSize 获取插件文件大小
func (s *PluginDownloadService) getPluginFileSize(pluginPath string) int64 {
	if pluginPath == "" {
		return 0
	}

	info, err := os.Stat(pluginPath)
	if err != nil {
		log.Warnw("[PluginDownload] failed to get file size", "pluginPath", pluginPath, "error", err)
		return 0
	}

	return info.Size()
}

// CalculatePluginChecksum 计算插件文件的SHA256校验和（工具方法）
func (s *PluginDownloadService) CalculatePluginChecksum(pluginPath string) (string, error) {
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

// ValidatePluginFile 验证插件文件的完整性
func (s *PluginDownloadService) ValidatePluginFile(pluginPath string, expectedChecksum string) error {
	actualChecksum, err := s.CalculatePluginChecksum(pluginPath)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// UpdatePluginChecksum 更新数据库中的插件校验和（维护工具）
func (s *PluginDownloadService) UpdatePluginChecksum(pluginID string) error {
	plugin, err := s.pluginRepo.GetPluginByID(pluginID)
	if err != nil {
		return err
	}

	// 动态生成本地路径
	pluginPath := s.getPluginLocalPath(pluginID, plugin.Version)

	checksum, err := s.CalculatePluginChecksum(pluginPath)
	if err != nil {
		return err
	}

	// TODO: 更新数据库
	// UPDATE t_plugin SET checksum = ? WHERE plugin_id = ?

	log.Infow("[PluginDownload] updated checksum for plugin", "pluginId", pluginID, "checksum", checksum)
	return nil
}

// DownloadStatistics 获取插件下载统计（可选实现，用于监控）
type DownloadStatistics struct {
	PluginID      string
	TotalDownload int64
	LastDownload  int64
	AvgSize       int64
}

func (s *PluginDownloadService) GetDownloadStatistics(pluginID string) (*DownloadStatistics, error) {
	// TODO: 实现统计逻辑，可以从日志或专门的统计表中获取
	return nil, nil
}
