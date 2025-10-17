package router

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/repo"
	service_plugin "github.com/observabil/arcade/internal/engine/service/plugin"
	httpx "github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/middleware"
	"github.com/observabil/arcade/pkg/log"
	"github.com/observabil/arcade/pkg/storage"
)

func (rt *Router) pluginRouter(r fiber.Router, auth fiber.Handler) {
	pluginGroup := r.Group("/plugins")
	{
		// 插件列表
		pluginGroup.Get("", auth, rt.listPlugins)

		// ===== 安装相关路由（放在前面，避免被 /:pluginId 捕获）=====
		// 安装插件（异步）
		pluginGroup.Post("/install", auth, rt.installPlugin)
		// 列出所有安装任务
		pluginGroup.Get("/tasks", auth, rt.listInstallTasks)
		// 查询安装任务状态
		pluginGroup.Get("/tasks/:taskId", auth, rt.getInstallTask)
		// 验证插件清单
		pluginGroup.Post("/validate-manifest", auth, rt.validateManifest)

		// ===== 插件管理路由（/:pluginId 通配符路由放在后面）=====
		// 插件详情
		pluginGroup.Get("/:pluginId", auth, rt.getPluginDetailByID)
		// 卸载插件
		pluginGroup.Delete("/:pluginId", auth, rt.uninstallPlugin)
		// 启用插件
		pluginGroup.Post("/:pluginId/enable", auth, rt.enablePlugin)
		// 禁用插件
		pluginGroup.Post("/:pluginId/disable", auth, rt.disablePlugin)
		// 更新插件
		pluginGroup.Put("/:pluginId", auth, rt.updatePlugin)
		// 插件配置管理
		pluginGroup.Get("/:pluginId/config", auth, rt.getPluginConfig)
		// 创建插件配置
		pluginGroup.Post("/:pluginId/config", auth, rt.createPluginConfig)
		// 更新插件配置
		pluginGroup.Put("/:pluginId/config", auth, rt.updatePluginConfig)
	}
}

// getPluginService 获取插件服务实例
func (rt *Router) getPluginService() *service_plugin.PluginService {
	pluginRepo := repo.NewPluginRepo(rt.Ctx)

	// Use shared plugin manager from router
	if rt.PluginManager == nil {
		log.Errorf("[PluginRouter] plugin manager is nil")
		return nil
	}

	// 获取存储提供者
	storageRepo := repo.NewStorageRepo(rt.Ctx)
	storageDBProvider, err := storage.NewStorageDBProvider(rt.Ctx, storageRepo)
	if err != nil {
		log.Errorf("[PluginRouter] failed to create storage DB provider: %v", err)
		return nil
	}

	storageProvider, err := storageDBProvider.GetStorageProvider()
	if err != nil {
		log.Errorf("[PluginRouter] failed to get storage provider: %v", err)
		return nil
	}

	return service_plugin.NewPluginService(rt.Ctx, pluginRepo, rt.PluginManager, storageProvider)
}

// listPlugins 列出所有插件
func (rt *Router) listPlugins(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginType := c.Query("pluginType")
	isEnabled := c.Query("isEnabled")

	var isEnabledInt int
	if isEnabled != "" {
		isEnabledInt = 1
	}

	plugins, err := pluginService.ListPlugins(pluginType, isEnabledInt)
	if err != nil {
		log.Errorf("[PluginRouter] failed to list plugins: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, plugins)
	return nil
}

// getPluginDetail 获取插件详情
func (rt *Router) getPluginDetailByID(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	plugin, err := pluginService.GetPluginDetailByID(pluginID)
	if err != nil {
		log.Errorf("[PluginRouter] failed to get plugin detail: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, plugin)
	return nil
}

// installPlugin 异步安装插件
func (rt *Router) installPlugin(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	source := c.FormValue("source")
	if source == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "source is required", c.Path())
	}

	req := &service_plugin.InstallPluginRequest{
		Source: service_plugin.PluginSource(source),
	}

	// 根据来源处理
	switch service_plugin.PluginSource(source) {
	case service_plugin.PluginSourceLocal:
		// 获取上传的zip文件
		file, err := c.FormFile("file")
		if err != nil {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("file is required for local source: %v", err), c.Path())
		}
		// 验证文件类型
		if filepath.Ext(file.Filename) != ".zip" {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "file must be a .zip package", c.Path())
		}
		req.File = file

		// marketplace source 暂未实现
	case service_plugin.PluginSourceMarket:
		// 	// 获取市场插件ID
		// 	marketID := c.FormValue("marketId")
		// 	if marketID == "" {
		// 		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "marketId is required for market source", c.Path())
		// 	}
		// 	req.MarketID = marketID

	default:
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("unsupported source: %s", source), c.Path())
	}

	// 异步安装插件
	resp, err := pluginService.InstallPluginAsync(req)
	if err != nil {
		log.Errorf("[PluginRouter] failed to start async install: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, resp)
	return nil
}

// getInstallTask 获取安装任务状态
func (rt *Router) getInstallTask(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	taskID := c.Params("taskId")
	if taskID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "taskId is required", c.Path())
	}

	task := pluginService.GetInstallTask(taskID)
	if task == nil {
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, "task not found", c.Path())
	}

	c.Locals(middleware.DETAIL, task)
	return nil
}

// listInstallTasks 列出所有安装任务
func (rt *Router) listInstallTasks(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	tasks := pluginService.ListInstallTasks()
	c.Locals(middleware.DETAIL, tasks)
	return nil
}

// uninstallPlugin 卸载插件
func (rt *Router) uninstallPlugin(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	if err := pluginService.UninstallPlugin(pluginID); err != nil {
		log.Errorf("[PluginRouter] failed to uninstall plugin: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "uninstall plugin")
	return nil
}

// enablePlugin 启用插件
func (rt *Router) enablePlugin(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	if err := pluginService.EnablePlugin(pluginID); err != nil {
		log.Errorf("[PluginRouter] failed to enable plugin: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "enable plugin")
	return nil
}

// disablePlugin 禁用插件
func (rt *Router) disablePlugin(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	if err := pluginService.DisablePlugin(pluginID); err != nil {
		log.Errorf("[PluginRouter] failed to disable plugin: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "disable plugin")
	return nil
}

// updatePlugin 更新插件
func (rt *Router) updatePlugin(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	source := c.FormValue("source")
	if source == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "source is required", c.Path())
	}

	req := &service_plugin.InstallPluginRequest{
		Source: service_plugin.PluginSource(source),
		File:   nil, // 从上传文件获取
	}

	// 更新插件
	resp, err := pluginService.UpdatePlugin(pluginID, req)
	if err != nil {
		log.Errorf("[PluginRouter] failed to update plugin: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, resp)
	c.Locals(middleware.OPERATION, "update plugin")
	return nil
}

func (rt *Router) validateManifest(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	// 可以接收JSON格式的manifest，或者上传zip包进行验证
	contentType := c.Get("Content-Type")

	var manifest service_plugin.PluginManifest

	if strings.Contains(contentType, "application/json") {
		// JSON格式
		if err := c.BodyParser(&manifest); err != nil {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("invalid manifest: %v", err), c.Path())
		}
	} else {
		// 从zip包中提取manifest
		file, err := c.FormFile("file")
		if err != nil {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "file is required", c.Path())
		}

		if filepath.Ext(file.Filename) != ".zip" {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "file must be a .zip package", c.Path())
		}

		// 读取并解析zip
		fileContent, err := file.Open()
		if err != nil {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("failed to open file: %v", err), c.Path())
		}
		defer fileContent.Close()

		zipData, err := io.ReadAll(fileContent)
		if err != nil {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("failed to read file: %v", err), c.Path())
		}

		// 只提取并验证manifest
		_, extractedManifest, err := pluginService.ExtractZipPackage(zipData, file.Size)
		if err != nil {
			return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("failed to extract manifest: %v", err), c.Path())
		}
		manifest = *extractedManifest
	}

	if err := pluginService.ValidateManifest(&manifest); err != nil {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, err.Error(), c.Path())
	}

	result := map[string]interface{}{
		"message":  "manifest is valid",
		"valid":    true,
		"manifest": manifest,
	}
	c.Locals(middleware.DETAIL, result)
	return nil
}

// getPluginConfig 获取插件配置
func (rt *Router) getPluginConfig(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	config, err := pluginService.GetPluginConfig(pluginID)
	if err != nil {
		log.Errorf("[PluginRouter] failed to get plugin config: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, config)
	return nil
}

// createPluginConfig 创建插件配置
func (rt *Router) createPluginConfig(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	var req service_plugin.UpdatePluginConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("invalid request: %v", err), c.Path())
	}

	// 设置pluginID
	req.PluginID = pluginID

	resp, err := pluginService.CreatePluginConfig(&req)
	if err != nil {
		log.Errorf("[PluginRouter] failed to create plugin config: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, resp)
	c.Locals(middleware.OPERATION, "create plugin config")
	return nil
}

// updatePluginConfig 更新插件配置
func (rt *Router) updatePluginConfig(c *fiber.Ctx) error {
	pluginService := rt.getPluginService()
	if pluginService == nil {
		return httpx.WithRepErrMsg(c, httpx.InternalError.Code, "failed to initialize plugin service", c.Path())
	}

	pluginID := c.Params("pluginId")
	if pluginID == "" {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, "pluginId is required", c.Path())
	}

	var req service_plugin.UpdatePluginConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return httpx.WithRepErrMsg(c, httpx.BadRequest.Code, fmt.Sprintf("invalid request: %v", err), c.Path())
	}

	// 设置pluginID
	req.PluginID = pluginID

	resp, err := pluginService.UpdatePluginConfig(&req)
	if err != nil {
		log.Errorf("[PluginRouter] failed to update plugin config: %v", err)
		return httpx.WithRepErrMsg(c, httpx.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, resp)
	c.Locals(middleware.OPERATION, "update plugin config")
	return nil
}
