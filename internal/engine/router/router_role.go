package router

import (
	"encoding/json"

	rolemodel "github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/tool"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func (rt *Router) roleRouter(r fiber.Router, auth fiber.Handler) {
	roleGroup := r.Group("/roles")
	{
		// Role management (authentication required)
		roleGroup.Get("/", auth, rt.listRoles)                                // GET /roles - list all roles (supports filters)
		roleGroup.Post("/", auth, rt.createRole)                              // POST /roles - create a custom role
		roleGroup.Get("/:roleId", auth, rt.getRole)                           // GET /roles/:roleId - get role details
		roleGroup.Put("/:roleId", auth, rt.updateRole)                        // PUT /roles/:roleId - update role
		roleGroup.Delete("/:roleId", auth, rt.deleteRole)                     // DELETE /roles/:roleId - delete role
		roleGroup.Put("/:roleId/toggle", auth, rt.toggleRole)                 // PUT /roles/:roleId/toggle - toggle enabled status
		roleGroup.Get("/:roleId/permissions", auth, rt.getRolePermissions)    // GET /roles/:roleId/permissions - get role permissions
		roleGroup.Put("/:roleId/permissions", auth, rt.updateRolePermissions) // PUT /roles/:roleId/permissions - update role permissions
	}
}

// listRoles lists all roles with pagination and filters
func (rt *Router) listRoles(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	// parse query parameters
	req := &rolemodel.ListRolesRequest{
		PageNum:  queryInt(c, "pageNum"),
		PageSize: queryInt(c, "pageSize"),
		Scope:    rolemodel.RoleScope(c.Query("scope")),
		OrgId:    c.Query("orgId"),
		Name:     c.Query("name"),
	}

	// parse optional filters
	if c.Query("isBuiltin") != "" {
		isBuiltin := queryInt(c, "isBuiltin")
		req.IsBuiltin = &isBuiltin
	}
	if c.Query("isEnabled") != "" {
		isEnabled := queryInt(c, "isEnabled")
		req.IsEnabled = &isEnabled
	}

	resp, err := roleService.ListRoles(req)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	// build response without timestamps
	var roleResponses []rolemodel.RoleResponse
	for _, role := range resp.Roles {
		var permissions []string
		if role.Permissions != "" {
			json.Unmarshal([]byte(role.Permissions), &permissions)
		}

		roleResponses = append(roleResponses, rolemodel.RoleResponse{
			RoleId:      role.RoleId,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			Scope:       role.Scope,
			OrgId:       role.OrgId,
			IsBuiltin:   role.IsBuiltin,
			IsEnabled:   role.IsEnabled,
			Priority:    role.Priority,
			Permissions: permissions,
			CreatedBy:   role.CreatedBy,
		})
	}

	result := map[string]interface{}{
		"roles":    roleResponses,
		"total":    resp.Total,
		"pageNum":  resp.PageNum,
		"pageSize": resp.PageSize,
	}

	c.Locals(middleware.DETAIL, result)
	return nil
}

// createRole creates a custom role
func (rt *Router) createRole(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	var req rolemodel.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	// validate required fields
	if req.Name == "" || req.Scope == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "name and scope are required fields", c.Path())
	}

	// get creator from JWT token
	userId, err := tool.ParseAuthorizationToken(c, rt.Http.Auth.SecretKey)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}
	req.CreatedBy = userId.UserId

	role, err := roleService.CreateRole(&req)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, role)
	c.Locals(middleware.OPERATION, "create role")
	return nil
}

// getRole gets a role by ID
func (rt *Router) getRole(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	roleId := c.Params("roleId")
	if roleId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "roleId is required", c.Path())
	}

	role, err := roleService.GetRole(roleId)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	// build response without timestamps
	var permissions []string
	if role.Permissions != "" {
		json.Unmarshal([]byte(role.Permissions), &permissions)
	}

	roleResp := rolemodel.RoleResponse{
		RoleId:      role.RoleId,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		Scope:       role.Scope,
		OrgId:       role.OrgId,
		IsBuiltin:   role.IsBuiltin,
		IsEnabled:   role.IsEnabled,
		Priority:    role.Priority,
		Permissions: permissions,
		CreatedBy:   role.CreatedBy,
	}

	c.Locals(middleware.DETAIL, roleResp)
	return nil
}

// updateRole updates a role (only custom roles can be updated)
func (rt *Router) updateRole(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	roleId := c.Params("roleId")
	if roleId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "roleId is required", c.Path())
	}

	var req rolemodel.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	if err := roleService.UpdateRole(roleId, &req); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "update role")
	return nil
}

// deleteRole deletes a role (only custom roles can be deleted)
func (rt *Router) deleteRole(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	roleId := c.Params("roleId")
	if roleId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "roleId is required", c.Path())
	}

	if err := roleService.DeleteRole(roleId); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "delete role")
	return nil
}

// toggleRole toggles the enabled status of a role
func (rt *Router) toggleRole(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	roleId := c.Params("roleId")
	if roleId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "roleId is required", c.Path())
	}

	if err := roleService.ToggleRole(roleId); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "toggle role status")
	return nil
}

// getRolePermissions gets a role's permissions
func (rt *Router) getRolePermissions(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	roleId := c.Params("roleId")
	if roleId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "roleId is required", c.Path())
	}

	permissions, err := roleService.GetRolePermissions(roleId)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, map[string]interface{}{
		"roleId":      roleId,
		"permissions": permissions,
	})
	return nil
}

// updateRolePermissions updates a role's permissions
func (rt *Router) updateRolePermissions(c *fiber.Ctx) error {
	roleService := rt.Services.Role

	roleId := c.Params("roleId")
	if roleId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "roleId is required", c.Path())
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	if err := roleService.UpdateRolePermissions(roleId, req.Permissions); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "update role permissions")
	return nil
}
