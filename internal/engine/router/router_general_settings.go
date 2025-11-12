package router

import (
	"strconv"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
)

// generalSettingsRouter registers general settings related routes
func (rt *Router) generalSettingsRouter(r fiber.Router, auth fiber.Handler) {
	settingsGroup := r.Group("/general-settings")
	{
		// General settings routes (authentication required)
		// Note: General settings are pre-defined system configurations, only updates are allowed
		settingsGroup.Get("/", auth, rt.getGeneralSettingsList)                          // GET /general-settings - list general settings
		settingsGroup.Get("/categories", auth, rt.getCategories)                         // GET /general-settings/categories - get all categories
		settingsGroup.Get("/:id", auth, rt.getGeneralSettings)                           // GET /general-settings/:id - get general settings
		settingsGroup.Put("/:id", auth, rt.updateGeneralSettings)                        // PUT /general-settings/:id - update general settings
		settingsGroup.Get("/by-name/:category/:name", auth, rt.getGeneralSettingsByName) // GET /general-settings/by-name/:category/:name - get by name
	}
}

// updateGeneralSettings updates a general settings
func (rt *Router) updateGeneralSettings(c *fiber.Ctx) error {
	generalSettingsService := rt.Services.GeneralSettings

	// get settings ID from path parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid settings id", c.Path())
	}

	var settings model.GeneralSettings
	if err := c.BodyParser(&settings); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request body", c.Path())
	}

	if err := generalSettingsService.UpdateGeneralSettings(id, &settings); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, settings)
	c.Locals(middleware.OPERATION, "update general settings")
	return nil
}

// getGeneralSettings gets a general settings by ID
func (rt *Router) getGeneralSettings(c *fiber.Ctx) error {
	generalSettingsService := rt.Services.GeneralSettings

	// get settings ID from path parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid settings id", c.Path())
	}

	settings, err := generalSettingsService.GetGeneralSettingsByID(id)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, settings)
	c.Locals(middleware.OPERATION, "get general settings")
	return nil
}

// getGeneralSettingsByName gets a general settings by category and name
func (rt *Router) getGeneralSettingsByName(c *fiber.Ctx) error {
	generalSettingsService := rt.Services.GeneralSettings

	category := c.Params("category")
	name := c.Params("name")

	if category == "" || name == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "category and name are required", c.Path())
	}

	settings, err := generalSettingsService.GetGeneralSettingsByName(category, name)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, settings)
	c.Locals(middleware.OPERATION, "get general settings by name")
	return nil
}

// getGeneralSettingsList gets general settings list with pagination and filters
func (rt *Router) getGeneralSettingsList(c *fiber.Ctx) error {
	generalSettingsService := rt.Services.GeneralSettings

	// get query parameters
	pageNum, _ := strconv.Atoi(c.Query("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	category := c.Query("category", "")

	settingsList, total, err := generalSettingsService.GetGeneralSettingsList(pageNum, pageSize, category)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	// construct response
	response := map[string]interface{}{
		"list":     settingsList,
		"total":    total,
		"pageNum":  pageNum,
		"pageSize": pageSize,
	}

	c.Locals(middleware.DETAIL, response)
	c.Locals(middleware.OPERATION, "get general settings list")
	return nil
}

// getCategories gets all distinct categories
func (rt *Router) getCategories(c *fiber.Ctx) error {
	generalSettingsService := rt.Services.GeneralSettings

	categories, err := generalSettingsService.GetCategories()
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, map[string]interface{}{"categories": categories})
	c.Locals(middleware.OPERATION, "get categories")
	return nil
}
