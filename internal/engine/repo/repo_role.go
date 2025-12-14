// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IRoleRepository interface {
	GetRole(roleId string) (*model.Role, error)
	GetRolesByRoleIds(roleIds []string) ([]model.Role, error)
	ListRoles() ([]model.Role, error)
}

type RoleRepo struct {
	database.IDatabase
}

func NewRoleRepo(db database.IDatabase) IRoleRepository {
	return &RoleRepo{
		IDatabase: db,
	}
}

// GetRole 获取角色
func (r *RoleRepo) GetRole(roleId string) (*model.Role, error) {
	var role model.Role
	err := r.Database().Select("id", "role_id", "name", "display_name", "description", "is_enabled", "created_at", "updated_at").
		Where("role_id = ? AND is_enabled = ?", roleId, 1).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRolesByRoleIds 根据角色ID列表获取角色
func (r *RoleRepo) GetRolesByRoleIds(roleIds []string) ([]model.Role, error) {
	if len(roleIds) == 0 {
		return []model.Role{}, nil
	}
	var roles []model.Role
	err := r.Database().Select("id", "role_id", "name", "display_name", "description", "is_enabled", "created_at", "updated_at").
		Where("role_id IN ? AND is_enabled = ?", roleIds, 1).Find(&roles).Error
	return roles, err
}

// ListRoles 列出所有启用的角色
func (r *RoleRepo) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := r.Database().Select("id", "role_id", "name", "display_name", "description", "is_enabled", "created_at", "updated_at").
		Where("is_enabled = ?", 1).Find(&roles).Error
	return roles, err
}
