package router

import (
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/service"
	httpx "github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/15
 * @file: router_storage.go
 * @description: storage configuration router
 */

// RegisterStorageRoutes 注册存储配置相关路由
func RegisterStorageRoutes(r fiber.Router, storageService *service.StorageService) {
	storageGroup := r.Group("/storage")
	{
		// 存储配置管理
		storageGroup.Post("/configs", createStorageConfig(storageService))
		storageGroup.Get("/configs", listStorageConfigs(storageService))
		storageGroup.Get("/configs/:id", getStorageConfig(storageService))
		storageGroup.Put("/configs/:id", updateStorageConfig(storageService))
		storageGroup.Delete("/configs/:id", deleteStorageConfig(storageService))
		storageGroup.Post("/configs/:id/default", setDefaultStorageConfig(storageService))
		storageGroup.Get("/configs/default", getDefaultStorageConfig(storageService))
	}
}

// createStorageConfig 创建存储配置
func createStorageConfig(storageService *service.StorageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req service.CreateStorageConfigRequest
		if err := c.BodyParser(&req); err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "Invalid request parameters", c.Path())
		}

		storageConfig, err := storageService.CreateStorageConfig(&req)
		if err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusInternalServerError, "Failed to create storage config", c.Path())
		}

		c.Locals("detail", storageConfig)
		return httpx.WithRepJSON(c, storageConfig)
	}
}

// listStorageConfigs 获取存储配置列表
func listStorageConfigs(storageService *service.StorageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		storageConfigs, err := storageService.ListStorageConfigs()
		if err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusInternalServerError, "Failed to get storage configs", c.Path())
		}

		c.Locals("detail", storageConfigs)
		return httpx.WithRepJSON(c, storageConfigs)
	}
}

// getStorageConfig 获取存储配置
func getStorageConfig(storageService *service.StorageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		storageID := c.Params("id")
		if storageID == "" {
			return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "Storage ID is required", c.Path())
		}

		storageConfig, err := storageService.GetStorageConfig(storageID)
		if err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusNotFound, "Storage config not found", c.Path())
		}

		c.Locals("detail", storageConfig)
		return httpx.WithRepJSON(c, storageConfig)
	}
}

// updateStorageConfig 更新存储配置
func updateStorageConfig(storageService *service.StorageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		storageID := c.Params("id")
		if storageID == "" {
			return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "Storage ID is required", c.Path())
		}

		var req service.UpdateStorageConfigRequest
		if err := c.BodyParser(&req); err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "Invalid request parameters", c.Path())
		}

		req.StorageId = storageID
		storageConfig, err := storageService.UpdateStorageConfig(&req)
		if err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusInternalServerError, "Failed to update storage config", c.Path())
		}

		c.Locals("detail", storageConfig)
		return httpx.WithRepJSON(c, storageConfig)
	}
}

// deleteStorageConfig 删除存储配置
func deleteStorageConfig(storageService *service.StorageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		storageID := c.Params("id")
		if storageID == "" {
			return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "Storage ID is required", c.Path())
		}

		err := storageService.DeleteStorageConfig(storageID)
		if err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusInternalServerError, "Failed to delete storage config", c.Path())
		}

		c.Locals("operation", "delete")
		return httpx.WithRepNotDetail(c)
	}
}

// setDefaultStorageConfig 设置默认存储配置
func setDefaultStorageConfig(storageService *service.StorageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		storageID := c.Params("id")
		if storageID == "" {
			return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "Storage ID is required", c.Path())
		}

		err := storageService.SetDefaultStorageConfig(storageID)
		if err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusInternalServerError, "Failed to set default storage config", c.Path())
		}

		c.Locals("operation", "set_default")
		return httpx.WithRepNotDetail(c)
	}
}

// getDefaultStorageConfig 获取默认存储配置
func getDefaultStorageConfig(storageService *service.StorageService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		storageConfig, err := storageService.GetDefaultStorageConfig()
		if err != nil {
			return httpx.WithRepErrMsg(c, fiber.StatusNotFound, "Default storage config not found", c.Path())
		}

		c.Locals("detail", storageConfig)
		return httpx.WithRepJSON(c, storageConfig)
	}
}

// uploadFile uploads a file to default storage
func (rt *Router) uploadFile(c *fiber.Ctx) error {
	storageRepo := repo.NewStorageRepo(rt.Ctx)
	uploadService := service.NewUploadService(rt.Ctx, storageRepo)

	// get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "file is required", c.Path())
	}

	// get optional custom path from query parameter
	customPath := c.Query("path")

	// upload file
	response, err := uploadService.UploadFile(file, "", customPath)
	if err != nil {
		return httpx.WithRepErrMsg(c, fiber.StatusInternalServerError, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, response)
	c.Locals(middleware.OPERATION, "upload file")
	return nil
}

// uploadFileWithStorage uploads a file to specific storage
func (rt *Router) uploadFileWithStorage(c *fiber.Ctx) error {
	storageRepo := repo.NewStorageRepo(rt.Ctx)
	uploadService := service.NewUploadService(rt.Ctx, storageRepo)

	// get storage ID from path parameter
	storageId := c.Params("storageId")
	if storageId == "" {
		return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "storageId is required", c.Path())
	}

	// get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return httpx.WithRepErrMsg(c, fiber.StatusBadRequest, "file is required", c.Path())
	}

	// get optional custom path from query parameter
	customPath := c.Query("path")

	// upload file
	response, err := uploadService.UploadFile(file, storageId, customPath)
	if err != nil {
		return httpx.WithRepErrMsg(c, fiber.StatusInternalServerError, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, response)
	c.Locals(middleware.OPERATION, "upload file")
	return nil
}
