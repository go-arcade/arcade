package service

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	pluginrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"gorm.io/datatypes"
)

// PluginSource 插件来源
type PluginSource string

const (
	PluginSourceMarket PluginSource = "market" // 插件市场
	PluginSourceLocal  PluginSource = "local"  // 本地上传
)

// PluginStatus 插件状态
type PluginStatus int

const (
	PluginStatusDisabled PluginStatus = 0 // 禁用
	PluginStatusEnabled  PluginStatus = 1 // 启用
	PluginStatusError    PluginStatus = 2 // 错误
)

const (
	pluginCachePath = "plugins"
)

// PluginManifest 插件资源清单
type PluginManifest struct {
	Name          string           `json:"name"`
	Version       string           `json:"version"`
	Description   string           `json:"description"`
	Author        string           `json:"author"`
	Homepage      string           `json:"homepage,omitempty"`
	Repository    string           `json:"repository,omitempty"`
	Sha256        string           `json:"sha256,omitempty"`
	PluginType    string           `json:"pluginType"`
	EntryPoint    string           `json:"entryPoint"`
	Dependencies  []string         `json:"dependencies,omitempty"`
	Config        json.RawMessage  `json:"config,omitempty"`
	Params        json.RawMessage  `json:"params,omitempty"`
	DefaultConfig json.RawMessage  `json:"defaultConfig,omitempty"`
	Icon          string           `json:"icon,omitempty"`
	Tags          []string         `json:"tags,omitempty"`
	MinVersion    string           `json:"minVersion,omitempty"` // 最低平台版本要求
	Resources     *PluginResources `json:"resources,omitempty"`
}

// PluginResources 插件资源定义
type PluginResources struct {
	CPU    string `json:"cpu,omitempty"`    // CPU要求 e.g. "1000m"
	Memory string `json:"memory,omitempty"` // 内存要求 e.g. "512Mi"
	Disk   string `json:"disk,omitempty"`   // 磁盘要求 e.g. "100Mi"
}

// PluginService 插件管理服务
type PluginService struct {
	ctx             *ctx.Context
	pluginRepo      pluginrepo.IPluginRepository
	pluginManager   *pluginpkg.Manager
	storageProvider storage.StorageProvider
}

// NewPluginService 创建插件管理服务
func NewPluginService(
	ctx *ctx.Context,
	pluginRepo pluginrepo.IPluginRepository,
	pluginManager *pluginpkg.Manager,
	storageProvider storage.StorageProvider,
) *PluginService {
	if err := os.MkdirAll(getLocalCachePath(), 0755); err != nil {
		log.Errorf("failed to create plugin cache dir: %v", err)
	}

	return &PluginService{
		ctx:             ctx,
		pluginRepo:      pluginRepo,
		pluginManager:   pluginManager,
		storageProvider: storageProvider,
	}
}

// InstallPluginRequest 安装插件请求
type InstallPluginRequest struct {
	Source   PluginSource          `json:"source"`             // 插件来源
	File     *multipart.FileHeader `json:"-"`                  // 本地上传文件
	Manifest *PluginManifest       `json:"manifest"`           // 插件清单
	MarketID string                `json:"marketId,omitempty"` // 市场插件ID
}

// InstallPluginResponse 安装插件响应
type InstallPluginResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	PluginID string `json:"pluginId"`
	Version  string `json:"version"`
}

// InstallPluginAsyncResponse 异步安装插件响应
type InstallPluginAsyncResponse struct {
	TaskID  string `json:"taskId"`
	Message string `json:"message"`
}

// UpdatePluginConfigRequest 更新插件配置请求
type UpdatePluginConfigRequest struct {
	PluginID string          `json:"pluginId"`
	Params   json.RawMessage `json:"params"`
	Config   json.RawMessage `json:"config"`
}

// PluginConfigResponse 插件配置响应
type PluginConfigResponse struct {
	PluginID  string          `json:"pluginId"`
	Params    json.RawMessage `json:"params"`
	Config    json.RawMessage `json:"config"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
}

// InstallPlugin 安装插件
func (s *PluginService) InstallPlugin(req *InstallPluginRequest) (*InstallPluginResponse, error) {
	log.Infof("[PluginService] installing plugin from source: %s", req.Source)

	var (
		pluginData []byte
		manifest   *PluginManifest
		err        error
	)

	switch req.Source {
	case PluginSourceLocal:
		// 从本地上传的文件安装
		pluginData, manifest, err = s.installFromLocal(req)
		if err != nil {
			return &InstallPluginResponse{
				Success: false,
				Message: fmt.Sprintf("failed to install from local: %v", err),
			}, err
		}

	case PluginSourceMarket:
		// 从插件市场安装
		pluginData, manifest, err = s.installFromMarket(req.MarketID)
		if err != nil {
			return &InstallPluginResponse{
				Success: false,
				Message: fmt.Sprintf("failed to install from market: %v", err),
			}, err
		}

	default:
		return &InstallPluginResponse{
			Success: false,
			Message: fmt.Sprintf("unsupported plugin source: %s", req.Source),
		}, fmt.Errorf("unsupported plugin source: %s", req.Source)
	}

	// 生成插件ID
	pluginID := id.GetXid()

	// 计算插件数据的校验和（SHA256）
	checksum := s.calculateSha256(pluginData)

	// 校验manifest的sha256和pluginData的sha256是否一致
	if manifest.Sha256 != "" && manifest.Sha256 != checksum {
		return &InstallPluginResponse{
			Success: false,
			Message: fmt.Sprintf("manifest sha256 mismatch: expected %s, got %s", manifest.Sha256, checksum),
		}, fmt.Errorf("manifest sha256 mismatch: expected %s, got %s", manifest.Sha256, checksum)
	}

	// 保存到本地缓存
	localPath, err := s.saveToLocalCache(manifest.Name, manifest.Version, pluginData)
	if err != nil {
		return &InstallPluginResponse{
			Success: false,
			Message: fmt.Sprintf("failed to save to local cache: %v", err),
		}, err
	}

	// 上传到对象存储（使用插件名）- 后台异步上传
	var s3Path string
	go func() {
		uploadedPath, uploadErr := s.uploadToStorage(manifest.Name, manifest.Version, localPath)
		if uploadErr != nil {
			log.Errorf("[PluginService] failed to upload to storage: %v", uploadErr)
		} else {
			s3Path = uploadedPath
			log.Infof("[PluginService] plugin uploaded to storage: %s", s3Path)
		}
	}()

	// 将 manifest 序列化为 JSON
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return &InstallPluginResponse{
			Success: false,
			Message: fmt.Sprintf("failed to marshal manifest: %v", err),
		}, err
	}

	// 保存到数据库
	pluginModel := &model.Plugin{
		PluginId:      pluginID,
		Name:          manifest.Name,
		Version:       manifest.Version,
		Description:   manifest.Description,
		Author:        manifest.Author,
		PluginType:    manifest.PluginType,
		EntryPoint:    manifest.EntryPoint,
		Icon:          manifest.Icon,
		Repository:    manifest.Repository,
		Documentation: "",                       // 从 manifest 中提取或设置为空
		IsEnabled:     int(PluginStatusEnabled), // 默认启用
		Checksum:      checksum,
		Source:        string(req.Source),
		S3Path:        s3Path,
		Manifest:      datatypes.JSON(manifestJSON),
		InstallTime:   time.Now(),
	}

	log.Infof("[PluginService] creating plugin record in database: %s", pluginID)
	if err := s.pluginRepo.CreatePlugin(pluginModel); err != nil {
		// 回滚：删除本地文件和S3文件
		log.Errorf("[PluginService] failed to save plugin to database: %v", err)
		s.cleanup(localPath, s3Path)
		return &InstallPluginResponse{
			Success: false,
			Message: fmt.Sprintf("failed to save plugin to database: %v", err),
		}, err
	}
	log.Infof("[PluginService] plugin record created in database successfully")

	// 自动创建插件配置（存储Schema信息到配置表）
	if len(manifest.Params) > 0 || len(manifest.Config) > 0 {
		pluginConfig := &model.PluginConfig{
			PluginId: pluginID,
			Params:   datatypes.JSON(manifest.Params),
			Config:   datatypes.JSON(manifest.Config),
		}

		if err := s.pluginRepo.CreatePluginConfig(pluginConfig); err != nil {
			log.Warnf("[PluginService] failed to create config for plugin %s: %v", pluginID, err)
			// 配置创建失败不影响插件安装
		} else {
			log.Infof("[PluginService] created config for plugin: %s", pluginID)
		}
	}

	// 热加载到内存（使用DefaultConfig）
	var defaultConfigData datatypes.JSON
	if len(manifest.DefaultConfig) > 0 {
		defaultConfigData = datatypes.JSON(manifest.DefaultConfig)
	}
	if err := s.hotReloadPlugin(manifest.Name, localPath, defaultConfigData); err != nil {
		log.Warnf("[PluginService] failed to hot reload plugin: %v", err)
		// 热加载失败不影响安装流程，但要记录错误
	}

	log.Infof("[PluginService] plugin installed successfully: %s (v%s)", manifest.Name, manifest.Version)

	return &InstallPluginResponse{
		Success:  true,
		Message:  "plugin installed successfully",
		PluginID: pluginID,
		Version:  manifest.Version,
	}, nil
}

// InstallPluginAsync 异步安装插件
func (s *PluginService) InstallPluginAsync(req *InstallPluginRequest) (*InstallPluginAsyncResponse, error) {
	// 先解析manifest获取插件名称和版本
	var manifest *PluginManifest

	switch req.Source {
	case PluginSourceLocal:
		if req.File == nil {
			return nil, fmt.Errorf("file is required")
		}

		// 验证文件类型
		if filepath.Ext(req.File.Filename) != ".zip" {
			return nil, fmt.Errorf("invalid file type, expected .zip file")
		}

		// 读取zip文件内容
		file, err := req.File.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		zipData, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}

		// 解压获取manifest
		_, manifest, err = s.extractZipPackage(zipData, req.File.Size)
		if err != nil {
			return nil, fmt.Errorf("failed to extract zip package: %v", err)
		}

	case PluginSourceMarket:
		return nil, fmt.Errorf("market installation not yet supported for async mode")

	default:
		return nil, fmt.Errorf("unsupported plugin source: %s", req.Source)
	}

	// 创建任务
	taskManager := GetTaskManager()
	task := taskManager.CreateTask(manifest.Name, manifest.Version)

	log.Infof("[PluginService] created async install task: %s for plugin %s v%s", task.TaskID, manifest.Name, manifest.Version)

	// 启动后台安装
	go s.executeInstallTask(task.TaskID, req)

	return &InstallPluginAsyncResponse{
		TaskID:  task.TaskID,
		Message: fmt.Sprintf("插件 %s v%s 正在后台安装，请使用任务ID查询进度", manifest.Name, manifest.Version),
	}, nil
}

// executeInstallTask 执行安装任务
func (s *PluginService) executeInstallTask(taskID string, req *InstallPluginRequest) {
	taskManager := GetTaskManager()

	// 更新为运行中
	taskManager.UpdateTask(taskID, TaskStatusRunning, 10, "开始解析插件包")

	// 执行安装
	resp, err := s.InstallPlugin(req)
	if err != nil {
		log.Errorf("[PluginService] async install task %s failed: %v", taskID, err)
		taskManager.UpdateTaskError(taskID, err)
		return
	}

	if !resp.Success {
		log.Errorf("[PluginService] async install task %s failed: %s", taskID, resp.Message)
		taskManager.UpdateTaskError(taskID, fmt.Errorf("%s", resp.Message))
		return
	}

	// 更新为成功
	taskManager.UpdateTaskSuccess(taskID, resp.PluginID)
	log.Infof("[PluginService] async install task %s completed successfully", taskID)
}

// GetInstallTask 获取安装任务状态
func (s *PluginService) GetInstallTask(taskID string) *PluginInstallTask {
	taskManager := GetTaskManager()
	return taskManager.GetTask(taskID)
}

// ListInstallTasks 列出所有安装任务
func (s *PluginService) ListInstallTasks() []*PluginInstallTask {
	taskManager := GetTaskManager()
	return taskManager.ListTasks()
}

// UninstallPlugin 卸载插件
func (s *PluginService) UninstallPlugin(pluginID string) error {
	log.Infof("[PluginService] uninstalling plugin: %s", pluginID)

	// 1. 从数据库获取插件信息
	pluginModel, err := s.pluginRepo.GetPluginByID(pluginID)
	if err != nil {
		return fmt.Errorf("plugin not found: %v", err)
	}

	// 2. 从内存中卸载
	if err := s.pluginManager.UnregisterPlugin(pluginID); err != nil {
		log.Warnf("[PluginService] failed to unregister plugin from memory: %v", err)
	}

	// 3. 删除S3文件（使用插件名）
	s3Path := s.getS3Path(pluginModel.Name, pluginModel.Version)
	if err := s.storageProvider.Delete(s.ctx, s3Path); err != nil {
		log.Warnf("[PluginService] failed to delete S3 file %s: %v", s3Path, err)
	} else {
		log.Infof("[PluginService] deleted S3 file: %s", s3Path)
	}

	// 5. 删除数据库记录
	if err := s.pluginRepo.DeletePlugin(pluginID); err != nil {
		return fmt.Errorf("failed to delete plugin from database: %v", err)
	}

	// 6. 删除相关配置
	if err := s.pluginRepo.DeletePluginConfigs(pluginID); err != nil {
		log.Warnf("[PluginService] failed to delete plugin configs: %v", err)
	}

	log.Infof("[PluginService] plugin uninstalled successfully: %s", pluginID)
	return nil
}

// EnablePlugin 启用插件
func (s *PluginService) EnablePlugin(pluginID string) error {
	log.Infof("[PluginService] enabling plugin: %s", pluginID)

	// 1. 更新数据库状态
	if err := s.pluginRepo.UpdatePluginStatus(pluginID, int(PluginStatusEnabled)); err != nil {
		return fmt.Errorf("failed to update plugin status: %v", err)
	}

	// 2. 获取插件信息
	pluginModel, err := s.pluginRepo.GetPluginByID(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin info: %v", err)
	}

	// 3. 热加载到内存（使用插件名）
	localPath := s.getLocalPath(pluginModel.Name, pluginModel.Version)
	// 从配置表获取配置（如果存在）
	var configData datatypes.JSON
	if pluginConfig, err := s.pluginRepo.GetPluginConfig(pluginID); err == nil && len(pluginConfig.Config) > 0 {
		configData = pluginConfig.Config
	}
	if err := s.hotReloadPlugin(pluginModel.Name, localPath, configData); err != nil {
		return fmt.Errorf("failed to hot reload plugin: %v", err)
	}

	log.Infof("[PluginService] plugin enabled successfully: %s", pluginID)
	return nil
}

// DisablePlugin 禁用插件
func (s *PluginService) DisablePlugin(pluginID string) error {
	log.Infof("[PluginService] disabling plugin: %s", pluginID)

	// 1. 从内存中卸载
	if err := s.pluginManager.UnregisterPlugin(pluginID); err != nil {
		log.Warnf("[PluginService] failed to unregister plugin from memory: %v", err)
	}

	// 2. 更新数据库状态
	if err := s.pluginRepo.UpdatePluginStatus(pluginID, int(PluginStatusDisabled)); err != nil {
		return fmt.Errorf("failed to update plugin status: %v", err)
	}

	log.Infof("[PluginService] plugin disabled successfully: %s", pluginID)
	return nil
}

// ListPlugins 列出所有插件
func (s *PluginService) ListPlugins(pluginType string, isEnabled int) ([]model.Plugin, error) {
	return s.pluginRepo.ListPlugins(pluginType, isEnabled)
}

// GetPluginDetailByID 根据plugin_id获取插件详情
func (s *PluginService) GetPluginDetailByID(pluginID string) (*model.Plugin, error) {
	return s.pluginRepo.GetPluginByID(pluginID)
}

// ======================== 私有方法 ========================

// installFromLocal 从本地上传的zip包安装
func (s *PluginService) installFromLocal(req *InstallPluginRequest) ([]byte, *PluginManifest, error) {
	if req.File == nil {
		return nil, nil, fmt.Errorf("file is required")
	}

	// 验证文件类型（必须是.zip文件）
	if filepath.Ext(req.File.Filename) != ".zip" {
		return nil, nil, fmt.Errorf("invalid file type, expected .zip file")
	}

	// 读取zip文件内容
	file, err := req.File.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	zipData, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %v", err)
	}

	// 解压zip包并提取内容
	pluginData, manifest, err := s.extractZipPackage(zipData, req.File.Size)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract zip package: %v", err)
	}

	return pluginData, manifest, nil
}

// ExtractZipPackage 解压zip包并提取插件和清单（公开方法供路由调用）
func (s *PluginService) ExtractZipPackage(zipData []byte, size int64) ([]byte, *PluginManifest, error) {
	return s.extractZipPackage(zipData, size)
}

// extractZipPackage 解压zip包并提取插件和清单
func (s *PluginService) extractZipPackage(zipData []byte, size int64) ([]byte, *PluginManifest, error) {
	// 创建zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), size)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create zip reader: %v", err)
	}

	var pluginData []byte
	var manifestData []byte
	var soFilename string

	// Log all files in the zip for debugging
	log.Infof("[PluginService] extracting zip package with %d files", len(zipReader.File))
	for _, file := range zipReader.File {
		log.Debugf("[PluginService] found file in zip: %s (dir: %v, size: %d)", file.Name, file.FileInfo().IsDir(), file.FileInfo().Size())
	}

	// 遍历zip包中的文件
	for _, file := range zipReader.File {
		// 跳过目录
		if file.FileInfo().IsDir() {
			log.Debugf("[PluginService] skipping directory: %s", file.Name)
			continue
		}

		// 获取文件名（去除路径）
		filename := filepath.Base(file.Name)
		lowerFilename := strings.ToLower(filename)

		log.Infof("[PluginService] processing file: %s (size: %d)", filename, file.FileInfo().Size())

		// 查找 manifest.json
		if lowerFilename == "manifest.json" {
			rc, err := file.Open()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to open manifest.json: %v", err)
			}
			manifestData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read manifest.json: %v", err)
			}
			log.Debug("[PluginService] found manifest.json in zip package")
			continue
		}

		// Skip documentation files
		if lowerFilename == "readme.md" ||
			lowerFilename == "readme.txt" ||
			lowerFilename == "license" ||
			lowerFilename == "license.txt" ||
			lowerFilename == "changelog.md" {
			log.Debugf("[PluginService] skipping documentation file: %s", filename)
			continue
		}

		// Look for executable plugin binary
		// Accept: no extension, .bin, .exe, .so (for backward compatibility)
		ext := strings.ToLower(filepath.Ext(filename))
		log.Infof("[PluginService] checking file %s with extension '%s'", filename, ext)

		isExecutable := ext == "" || ext == ".bin" || ext == ".exe" || ext == ".so"

		if isExecutable {
			if len(pluginData) > 0 {
				log.Warnf("[PluginService] found multiple executable files, using first one, skipping: %s", filename)
				continue
			}

			rc, err := file.Open()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to open plugin file %s: %v", filename, err)
			}
			pluginData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read plugin file %s: %v", filename, err)
			}
			soFilename = filename
			log.Infof("[PluginService] ✓ found plugin executable: %s (size: %d bytes)", filename, len(pluginData))
		} else {
			log.Infof("[PluginService] skipping non-executable file: %s (ext: %s)", filename, ext)
		}
	}

	// 验证必需文件
	if len(manifestData) == 0 {
		return nil, nil, fmt.Errorf("manifest.json not found in zip package")
	}
	if len(pluginData) == 0 {
		return nil, nil, fmt.Errorf("plugin executable file not found in zip package")
	}

	// 解析 manifest
	var manifest PluginManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, nil, fmt.Errorf("failed to parse manifest.json: %v", err)
	}

	// 验证 manifest
	if err := s.ValidateManifest(&manifest); err != nil {
		return nil, nil, fmt.Errorf("invalid manifest: %v", err)
	}

	// 更新 entryPoint 为实际的可执行文件名
	if manifest.EntryPoint == "" || manifest.EntryPoint != soFilename {
		log.Infof("[PluginService] updating entryPoint from '%s' to '%s'", manifest.EntryPoint, soFilename)
		manifest.EntryPoint = soFilename
	}

	log.Infof("[PluginService] extracted plugin package: %s v%s", manifest.Name, manifest.Version)
	return pluginData, &manifest, nil
}

// installFromMarket 从插件市场安装
func (s *PluginService) installFromMarket(marketID string) ([]byte, *PluginManifest, error) {
	// TODO: 实现从插件市场下载逻辑
	// 1. 调用插件市场API获取插件信息和下载地址
	// 2. 下载插件文件
	// 3. 验证签名
	// 4. 解析清单
	return nil, nil, fmt.Errorf("plugin market not implemented yet")
}

// saveToLocalCache saves plugin to local cache (using plugin name)
func (s *PluginService) saveToLocalCache(pluginName, version string, data []byte) (string, error) {
	// RPC plugins are executable binaries, not .so files
	filename := fmt.Sprintf("%s_%s", pluginName, version)
	localPath := filepath.Join(getLocalCachePath(), filename)

	// Write with executable permissions
	if err := os.WriteFile(localPath, data, 0755); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	log.Infof("[PluginService] saved to local cache: %s", localPath)
	return localPath, nil
}

// uploadToStorage 上传到对象存储（使用插件名和本地文件路径）
func (s *PluginService) uploadToStorage(pluginName, version string, localFilePath string) (string, error) {
	if s.storageProvider == nil {
		return "", fmt.Errorf("storage provider not configured")
	}

	s3Path := s.getS3Path(pluginName, version)

	// 创建一个可用的multipart.FileHeader
	fileHeader, err := s.createFileHeaderFromLocalFile(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file header: %v", err)
	}
	defer os.Remove(fileHeader.Filename) // 清理临时文件

	// 上传到storage
	url, err := s.storageProvider.Upload(s.ctx, s3Path, fileHeader, "application/octet-stream")
	if err != nil {
		return "", fmt.Errorf("failed to upload to storage: %v", err)
	}

	log.Infof("[PluginService] uploaded to storage: %s", url)
	return s3Path, nil
}

// createFileHeaderFromLocalFile create a usable multipart.FileHeader from local file
func (s *PluginService) createFileHeaderFromLocalFile(localFilePath string) (*multipart.FileHeader, error) {
	// read file content
	data, err := os.ReadFile(localFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// create a multipart buffer
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	// create form file
	part, err := writer.CreateFormFile("file", filepath.Base(localFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	// write data
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to form file: %v", err)
	}

	// close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}

	// parse multipart form
	reader := multipart.NewReader(buf, writer.Boundary())
	form, err := reader.ReadForm(int64(len(data)) + 1024)
	if err != nil {
		return nil, fmt.Errorf("failed to read form: %v", err)
	}

	// get FileHeader
	if len(form.File["file"]) == 0 {
		return nil, fmt.Errorf("no file found in form")
	}

	fileHeader := form.File["file"][0]

	// note: do not call form.RemoveAll(), we need to keep temporary file until upload completed

	return fileHeader, nil
}

// hotReloadPlugin hot reload plugin to memory
func (s *PluginService) hotReloadPlugin(pluginName, localPath string, defaultConfig datatypes.JSON) error {
	// parse config to json.RawMessage
	var configJSON json.RawMessage
	if len(defaultConfig) > 0 {
		configJSON = json.RawMessage(defaultConfig)
	} else {
		configJSON = json.RawMessage("{}")
	}

	// try to unload old plugin (if exists)
	_ = s.pluginManager.UnregisterPlugin(pluginName)

	// create plugin config (basic information will be retrieved from plugin itself)
	pluginConfig := pluginpkg.PluginConfig{
		Name:   pluginName,
		Config: configJSON,
		// Type and Version will be retrieved from plugin instance, here no need to set
	}

	// 获取插件记录以更新注册状态
	pluginRecord, err := s.pluginRepo.GetPluginByName(pluginName)
	if err == nil && pluginRecord != nil {
		// 更新为注册中
		_ = s.pluginRepo.UpdatePluginRegistrationStatus(pluginRecord.PluginId, model.PluginRegisterStatusRegistering, "")
	}

	// register new plugin (directly use original path, RPC plugin has no path limitation)
	if err := s.pluginManager.RegisterPlugin(pluginName, localPath, pluginConfig); err != nil {
		// 注册失败，更新数据库状态
		if pluginRecord != nil {
			_ = s.pluginRepo.UpdatePluginRegistrationStatus(pluginRecord.PluginId, model.PluginRegisterStatusFailed, err.Error())
		}
		return fmt.Errorf("failed to register plugin: %v", err)
	}

	// 注册成功，更新数据库状态
	if pluginRecord != nil {
		_ = s.pluginRepo.UpdatePluginRegistrationStatus(pluginRecord.PluginId, model.PluginRegisterStatusRegistered, "")
	}

	log.Infof("[PluginService] hot reloaded plugin: %s from %s", pluginName, localPath)
	return nil
}

// calculateSha256 计算SHA256校验和
func (s *PluginService) calculateSha256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// getLocalPath gets local cache path (dynamically generated, not stored in database, using plugin name)
func (s *PluginService) getLocalPath(pluginName, version string) string {
	// RPC plugins are executable binaries
	filename := fmt.Sprintf("%s_%s", pluginName, version)
	return filepath.Join(getLocalCachePath(), filename)
}

// getS3Path gets S3 storage path (using plugin name)
func (s *PluginService) getS3Path(pluginName, version string) string {
	// RPC plugins are executable binaries
	return fmt.Sprintf("plugins/%s/%s/%s", pluginName, version, pluginName)
}

// cleanup 清理文件
func (s *PluginService) cleanup(localPath, s3Path string) {
	if localPath != "" {
		if err := os.Remove(localPath); err != nil {
			log.Warnf("[PluginService] failed to remove local file %s: %v", localPath, err)
		}
	}
	if s3Path != "" && s.storageProvider != nil {
		if err := s.storageProvider.Delete(s.ctx, s3Path); err != nil {
			log.Warnf("[PluginService] failed to remove S3 file %s: %v", s3Path, err)
		}
	}
}

// UpdatePlugin 更新插件（版本升级）
func (s *PluginService) UpdatePlugin(pluginID string, req *InstallPluginRequest) (*InstallPluginResponse, error) {
	log.Infof("[PluginService] updating plugin: %s", pluginID)

	// 1. 检查插件是否存在
	oldPlugin, err := s.pluginRepo.GetPluginByID(pluginID)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %v", err)
	}

	// 2. 先禁用旧插件
	if err := s.DisablePlugin(pluginID); err != nil {
		log.Warnf("[PluginService] failed to disable old plugin: %v", err)
	}

	// 3. 安装新版本
	installResp, err := s.InstallPlugin(req)
	if err != nil {
		// 安装失败，恢复旧插件
		_ = s.EnablePlugin(pluginID)
		return nil, fmt.Errorf("failed to install new version: %v", err)
	}

	// 4. 清理旧版本文件（使用插件名）
	oldLocalPath := s.getLocalPath(oldPlugin.Name, oldPlugin.Version)
	s.cleanup(oldLocalPath, s.getS3Path(oldPlugin.Name, oldPlugin.Version))

	log.Infof("[PluginService] plugin updated successfully: %s", pluginID)
	return installResp, nil
}

// GetPluginConfig 获取插件配置
func (s *PluginService) GetPluginConfig(pluginID string) (*PluginConfigResponse, error) {
	log.Infof("[PluginService] getting plugin config: %s", pluginID)

	// 验证插件是否存在
	_, err := s.pluginRepo.GetPluginByID(pluginID)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %v", err)
	}

	// 获取插件配置
	config, err := s.pluginRepo.GetPluginConfig(pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin config: %v", err)
	}

	return &PluginConfigResponse{
		PluginID: config.PluginId,
		Params:   json.RawMessage(config.Params),
		Config:   json.RawMessage(config.Config),
	}, nil
}

// CreatePluginConfig 创建插件配置
func (s *PluginService) CreatePluginConfig(req *UpdatePluginConfigRequest) (*PluginConfigResponse, error) {
	log.Infof("[PluginService] creating plugin config for: %s", req.PluginID)

	// 验证插件是否存在
	_, err := s.pluginRepo.GetPluginByID(req.PluginID)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %v", err)
	}

	// 检查配置是否已存在
	existingConfig, _ := s.pluginRepo.GetPluginConfig(req.PluginID)
	if existingConfig != nil {
		return nil, fmt.Errorf("plugin config already exists, use update instead")
	}

	// 创建配置
	config := &model.PluginConfig{
		PluginId: req.PluginID,
		Params:   datatypes.JSON(req.Params),
		Config:   datatypes.JSON(req.Config),
	}

	if err := s.pluginRepo.CreatePluginConfig(config); err != nil {
		return nil, fmt.Errorf("failed to create plugin config: %v", err)
	}

	log.Infof("[PluginService] plugin config created for: %s", req.PluginID)

	return &PluginConfigResponse{
		PluginID: config.PluginId,
		Params:   json.RawMessage(config.Params),
		Config:   json.RawMessage(config.Config),
	}, nil
}

// UpdatePluginConfig 更新插件配置
func (s *PluginService) UpdatePluginConfig(req *UpdatePluginConfigRequest) (*PluginConfigResponse, error) {
	log.Infof("[PluginService] updating plugin config for: %s", req.PluginID)

	// 验证插件是否存在
	_, err := s.pluginRepo.GetPluginByID(req.PluginID)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %v", err)
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if len(req.Params) > 0 {
		updates["params_schema"] = datatypes.JSON(req.Params)
	}
	if len(req.Config) > 0 {
		updates["config_schema"] = datatypes.JSON(req.Config)
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// 更新配置
	if err := s.pluginRepo.UpdatePluginConfig(req.PluginID, updates); err != nil {
		return nil, fmt.Errorf("failed to update plugin config: %v", err)
	}

	// 重新获取更新后的配置
	updatedConfig, err := s.pluginRepo.GetPluginConfig(req.PluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated config: %v", err)
	}

	log.Infof("[PluginService] plugin config updated for: %s", req.PluginID)

	return &PluginConfigResponse{
		PluginID: updatedConfig.PluginId,
		Params:   json.RawMessage(updatedConfig.Params),
		Config:   json.RawMessage(updatedConfig.Config),
	}, nil
}

// ValidateManifest 验证插件清单
func (s *PluginService) ValidateManifest(manifest *PluginManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.PluginType == "" {
		return fmt.Errorf("plugin type is required")
	}
	if manifest.EntryPoint == "" {
		return fmt.Errorf("plugin entry point is required")
	}
	return nil
}

// getLocalCachePath 获取本地缓存目录
func getLocalCachePath() string {
	localCachePath, err := os.Getwd()
	if err != nil {
		log.Errorf("failed to get current working directory: %v", err)
		return ""
	}
	return filepath.Join(localCachePath, pluginCachePath)
}
