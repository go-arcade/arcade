package service

import (
	"encoding/json"
	"errors"

	rolemodel "github.com/go-arcade/arcade/internal/engine/model"
	rolerepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
)

type RoleService struct {
	roleRepo rolerepo.IRoleRepository
}

func NewRoleService(roleRepo rolerepo.IRoleRepository) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
	}
}

// CreateRole creates a custom role
func (rs *RoleService) CreateRole(req *rolemodel.CreateRoleRequest) (*rolemodel.Role, error) {
	// validate scope
	if req.Scope != rolemodel.RoleScopeProject && req.Scope != rolemodel.RoleScopeTeam && req.Scope != rolemodel.RoleScopeOrg {
		return nil, errors.New("invalid scope, must be project, team, or org")
	}

	// check if role with same name exists in the same scope and org
	existing, err := rs.roleRepo.GetRoleByName(req.Name, req.Scope, req.OrgId)
	if err == nil && existing != nil {
		return nil, errors.New("role with this name already exists in the specified scope")
	}

	// serialize permissions
	permJson, err := json.Marshal(req.Permissions)
	if err != nil {
		log.Errorw("failed to serialize permissions", "error", err)
		return nil, errors.New("invalid permissions format")
	}

	role := &rolemodel.Role{
		RoleId:      id.GetUUID(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Scope:       req.Scope,
		OrgId:       req.OrgId,
		IsBuiltin:   rolemodel.RoleCustom, // custom role
		IsEnabled:   rolemodel.RoleEnabled,
		Priority:    req.Priority,
		Permissions: string(permJson),
		CreatedBy:   req.CreatedBy,
	}

	if err := rs.roleRepo.CreateRole(role); err != nil {
		log.Errorw("failed to create role", "error", err)
		return nil, errors.New("failed to create role")
	}

	log.Infow("role created successfully", "roleId", role.RoleId)
	return role, nil
}

// UpdateRole updates a role (only custom roles can be updated)
func (rs *RoleService) UpdateRole(roleId string, req *rolemodel.UpdateRoleRequest) error {
	// get existing role
	role, err := rs.roleRepo.GetRole(roleId)
	if err != nil {
		log.Errorw("role not found", "roleId", roleId, "error", err)
		return errors.New("role not found")
	}

	// built-in roles cannot be updated
	if role.IsBuiltin == rolemodel.RoleBuiltin {
		return errors.New("built-in roles cannot be updated")
	}

	// update fields
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Priority != 0 {
		role.Priority = req.Priority
	}
	if req.Permissions != nil {
		permJson, err := json.Marshal(req.Permissions)
		if err != nil {
			log.Errorw("failed to serialize permissions", "roleId", roleId, "error", err)
			return errors.New("invalid permissions format")
		}
		role.Permissions = string(permJson)
	}

	if err := rs.roleRepo.UpdateRole(role); err != nil {
		log.Errorw("failed to update role", "roleId", roleId, "error", err)
		return errors.New("failed to update role")
	}

	log.Infow("role updated successfully", "roleId", roleId)
	return nil
}

// DeleteRole deletes a role (only custom roles can be deleted)
func (rs *RoleService) DeleteRole(roleId string) error {
	// get existing role
	role, err := rs.roleRepo.GetRole(roleId)
	if err != nil {
		log.Errorw("role not found", "roleId", roleId, "error", err)
		return errors.New("role not found")
	}

	// built-in roles cannot be deleted
	if role.IsBuiltin == rolemodel.RoleBuiltin {
		return errors.New("built-in roles cannot be deleted")
	}

	if err := rs.roleRepo.DeleteRole(roleId); err != nil {
		log.Errorw("failed to delete role", "roleId", roleId, "error", err)
		return errors.New("failed to delete role")
	}

	log.Infow("role deleted successfully", "roleId", roleId)
	return nil
}

// ToggleRole toggles the enabled status of a role
func (rs *RoleService) ToggleRole(roleId string) error {
	role, err := rs.roleRepo.GetRole(roleId)
	if err != nil {
		log.Errorw("role not found", "roleId", roleId, "error", err)
		return errors.New("role not found")
	}

	newStatus := rolemodel.RoleDisabled
	if role.IsEnabled == rolemodel.RoleDisabled {
		newStatus = rolemodel.RoleEnabled
	}

	if err := rs.roleRepo.EnableRole(roleId, newStatus == rolemodel.RoleEnabled); err != nil {
		log.Errorw("failed to toggle role status", "roleId", roleId, "error", err)
		return errors.New("failed to toggle role status")
	}

	log.Infow("role status toggled successfully", "roleId", roleId, "newStatus", newStatus)
	return nil
}

// GetRole gets a role by ID
func (rs *RoleService) GetRole(roleId string) (*rolemodel.Role, error) {
	role, err := rs.roleRepo.GetRole(roleId)
	if err != nil {
		log.Errorw("role not found", "roleId", roleId, "error", err)
		return nil, errors.New("role not found")
	}
	return role, nil
}

// ListRoles lists roles with pagination and filters
func (rs *RoleService) ListRoles(req *rolemodel.ListRolesRequest) (*rolemodel.ListRolesResponse, error) {
	// set default pagination
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // max page size
	}

	roles, total, err := rs.roleRepo.ListRolesWithPagination(req)
	if err != nil {
		log.Errorw("failed to list roles", "error", err)
		return nil, errors.New("failed to list roles")
	}

	return &rolemodel.ListRolesResponse{
		Roles:    roles,
		Total:    total,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	}, nil
}

// GetRolePermissions gets a role's permissions as a list
func (rs *RoleService) GetRolePermissions(roleId string) ([]string, error) {
	role, err := rs.roleRepo.GetRole(roleId)
	if err != nil {
		return nil, errors.New("role not found")
	}

	var permissions []string
	if role.Permissions != "" {
		if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
			log.Errorw("failed to parse permissions", "roleId", roleId, "error", err)
			return nil, errors.New("invalid permissions format")
		}
	}

	return permissions, nil
}

// UpdateRolePermissions updates a role's permissions
func (rs *RoleService) UpdateRolePermissions(roleId string, permissions []string) error {
	// get existing role
	role, err := rs.roleRepo.GetRole(roleId)
	if err != nil {
		return errors.New("role not found")
	}

	// built-in roles cannot be updated
	if role.IsBuiltin == rolemodel.RoleBuiltin {
		return errors.New("built-in roles' permissions cannot be modified")
	}

	if err := rs.roleRepo.UpdateRolePermissions(roleId, permissions); err != nil {
		log.Errorw("failed to update role permissions", "roleId", roleId, "error", err)
		return errors.New("failed to update role permissions")
	}

	log.Infow("role permissions updated successfully", "roleId", roleId)
	return nil
}
