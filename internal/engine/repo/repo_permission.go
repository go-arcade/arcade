package repo

import (
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/14
 * @file: repo_permission.go
 * @description: 权限仓库（基于关联表设计）
 */

type PermissionRepo struct {
	Ctx *ctx.Context
}

func NewPermissionRepo(ctx *ctx.Context) *PermissionRepo {
	return &PermissionRepo{
		Ctx: ctx,
	}
}

// GetPermissionByCode 根据权限代码获取权限点
func (r *PermissionRepo) GetPermissionByCode(code string) (*model.Permission, error) {
	var permission model.Permission
	err := r.Ctx.DB.Where("code = ? AND is_enabled = ?", code, 1).First(&permission).Error
	return &permission, err
}

// GetPermissionsByCategory 根据分类获取权限点列表
func (r *PermissionRepo) GetPermissionsByCategory(category string) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.Ctx.DB.Where("category = ? AND is_enabled = ?", category, 1).
		Order("code ASC").
		Find(&permissions).Error
	return permissions, err
}

// GetAllPermissions 获取所有权限点
func (r *PermissionRepo) GetAllPermissions() ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.Ctx.DB.Where("is_enabled = ?", 1).
		Order("category ASC, code ASC").
		Find(&permissions).Error
	return permissions, err
}

// GetRolePermissions 获取角色的所有权限
func (r *PermissionRepo) GetRolePermissions(roleId string) ([]string, error) {
	var permissionCodes []string

	err := r.Ctx.DB.Table("t_role_permission rp").
		Select("p.code").
		Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
		Where("rp.role_id = ? AND p.is_enabled = ?", roleId, 1).
		Pluck("code", &permissionCodes).Error

	if err != nil {
		log.Errorf("failed to get role permissions for %s: %v", roleId, err)
		return []string{}, err
	}

	return permissionCodes, nil
}

// GetRolePermissionsDetailed 获取角色的详细权限信息
func (r *PermissionRepo) GetRolePermissionsDetailed(roleId string) ([]model.Permission, error) {
	var permissions []model.Permission

	err := r.Ctx.DB.Table("t_permission p").
		Select("p.*").
		Joins("JOIN t_role_permission rp ON p.permission_id = rp.permission_id").
		Where("rp.role_id = ? AND p.is_enabled = ?", roleId, 1).
		Order("p.category ASC, p.code ASC").
		Find(&permissions).Error

	return permissions, err
}

// AddRolePermission 为角色添加权限
func (r *PermissionRepo) AddRolePermission(roleId, permissionId string) error {
	rolePermission := &model.RolePermission{
		RoleId:       roleId,
		PermissionId: permissionId,
	}
	return r.Ctx.DB.Create(rolePermission).Error
}

// AddRolePermissionByCode 为角色添加权限（通过权限代码）
func (r *PermissionRepo) AddRolePermissionByCode(roleId, permissionCode string) error {
	permission, err := r.GetPermissionByCode(permissionCode)
	if err != nil {
		return err
	}
	return r.AddRolePermission(roleId, permission.PermissionId)
}

// RemoveRolePermission 移除角色的权限
func (r *PermissionRepo) RemoveRolePermission(roleId, permissionId string) error {
	return r.Ctx.DB.Where("role_id = ? AND permission_id = ?", roleId, permissionId).
		Delete(&model.RolePermission{}).Error
}

// RemoveRolePermissionByCode 移除角色的权限（通过权限代码）
func (r *PermissionRepo) RemoveRolePermissionByCode(roleId, permissionCode string) error {
	permission, err := r.GetPermissionByCode(permissionCode)
	if err != nil {
		return err
	}
	return r.RemoveRolePermission(roleId, permission.PermissionId)
}

// RemoveAllRolePermissions 移除角色的所有权限
func (r *PermissionRepo) RemoveAllRolePermissions(roleId string) error {
	return r.Ctx.DB.Where("role_id = ?", roleId).
		Delete(&model.RolePermission{}).Error
}

// SetRolePermissions 设置角色的权限（替换所有现有权限）
func (r *PermissionRepo) SetRolePermissions(roleId string, permissionCodes []string) error {
	// 开启事务
	tx := r.Ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除现有权限
	if err := tx.Where("role_id = ?", roleId).Delete(&model.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 添加新权限
	for _, code := range permissionCodes {
		var permission model.Permission
		if err := tx.Where("code = ? AND is_enabled = ?", code, 1).First(&permission).Error; err != nil {
			tx.Rollback()
			return err
		}

		rolePermission := &model.RolePermission{
			RoleId:       roleId,
			PermissionId: permission.PermissionId,
		}
		if err := tx.Create(rolePermission).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// HasPermission 检查角色是否拥有指定权限
func (r *PermissionRepo) HasPermission(roleId, permissionCode string) (bool, error) {
	var count int64
	err := r.Ctx.DB.Table("t_role_permission rp").
		Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
		Where("rp.role_id = ? AND p.code = ? AND p.is_enabled = ?", roleId, permissionCode, 1).
		Count(&count).Error

	return count > 0, err
}

// GetPermissionsByIds 根据ID列表批量获取权限
func (r *PermissionRepo) GetPermissionsByIds(permissionIds []string) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.Ctx.DB.Where("permission_id IN ? AND is_enabled = ?", permissionIds, 1).
		Find(&permissions).Error
	return permissions, err
}

// CreatePermission 创建新权限点
func (r *PermissionRepo) CreatePermission(permission *model.Permission) error {
	return r.Ctx.DB.Create(permission).Error
}

// UpdatePermission 更新权限点
func (r *PermissionRepo) UpdatePermission(permission *model.Permission) error {
	return r.Ctx.DB.Model(&model.Permission{}).
		Where("permission_id = ?", permission.PermissionId).
		Updates(permission).Error
}

// DisablePermission 禁用权限点
func (r *PermissionRepo) DisablePermission(permissionId string) error {
	return r.Ctx.DB.Model(&model.Permission{}).
		Where("permission_id = ?", permissionId).
		Update("is_enabled", 0).Error
}

// EnablePermission 启用权限点
func (r *PermissionRepo) EnablePermission(permissionId string) error {
	return r.Ctx.DB.Model(&model.Permission{}).
		Where("permission_id = ?", permissionId).
		Update("is_enabled", 1).Error
}

// GetRolesWithPermission 获取拥有指定权限的所有角色
func (r *PermissionRepo) GetRolesWithPermission(permissionCode string) ([]string, error) {
	var roleIds []string

	err := r.Ctx.DB.Table("t_role_permission rp").
		Select("rp.role_id").
		Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
		Where("p.code = ? AND p.is_enabled = ?", permissionCode, 1).
		Pluck("role_id", &roleIds).Error

	return roleIds, err
}

// CountRolePermissions 统计角色的权限数量
func (r *PermissionRepo) CountRolePermissions(roleId string) (int64, error) {
	var count int64
	err := r.Ctx.DB.Table("t_role_permission rp").
		Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
		Where("rp.role_id = ? AND p.is_enabled = ?", roleId, 1).
		Count(&count).Error
	return count, err
}

// GetPermissionStatistics 获取权限统计信息
func (r *PermissionRepo) GetPermissionStatistics() (map[string]int64, error) {
	var results []struct {
		Category string
		Count    int64
	}

	err := r.Ctx.DB.Table("t_permission").
		Select("category, COUNT(*) as count").
		Where("is_enabled = ?", 1).
		Group("category").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Category] = r.Count
	}

	return stats, nil
}
