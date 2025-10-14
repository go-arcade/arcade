package repo

import (
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: repo_project_team_access.go
 * @description: 项目团队访问权限仓储
 */

type ProjectTeamAccessRepo struct {
	ctx *ctx.Context
}

func NewProjectTeamAccessRepo(ctx *ctx.Context) *ProjectTeamAccessRepo {
	return &ProjectTeamAccessRepo{ctx: ctx}
}

// GetProjectTeamAccess 获取项目团队访问权限
func (r *ProjectTeamAccessRepo) GetProjectTeamAccess(projectId, teamId string) (*model.ProjectTeamAccess, error) {
	var access model.ProjectTeamAccess
	err := r.ctx.DB.Where("project_id = ? AND team_id = ?", projectId, teamId).First(&access).Error
	return &access, err
}

// ListProjectTeams 列出项目的所有团队
func (r *ProjectTeamAccessRepo) ListProjectTeams(projectId string) ([]model.ProjectTeamAccess, error) {
	var accesses []model.ProjectTeamAccess
	err := r.ctx.DB.Where("project_id = ?", projectId).Find(&accesses).Error
	return accesses, err
}

// ListTeamProjects 列出团队可访问的所有项目
func (r *ProjectTeamAccessRepo) ListTeamProjects(teamId string) ([]model.ProjectTeamAccess, error) {
	var accesses []model.ProjectTeamAccess
	err := r.ctx.DB.Where("team_id = ?", teamId).Find(&accesses).Error
	return accesses, err
}

// GrantTeamAccess 授予团队访问权限
func (r *ProjectTeamAccessRepo) GrantTeamAccess(access *model.ProjectTeamAccess) error {
	return r.ctx.DB.Create(access).Error
}

// UpdateTeamAccessLevel 更新团队访问级别
func (r *ProjectTeamAccessRepo) UpdateTeamAccessLevel(projectId, teamId, accessLevel string) error {
	return r.ctx.DB.Model(&model.ProjectTeamAccess{}).
		Where("project_id = ? AND team_id = ?", projectId, teamId).
		Update("access_level", accessLevel).Error
}

// RevokeTeamAccess 撤销团队访问权限
func (r *ProjectTeamAccessRepo) RevokeTeamAccess(projectId, teamId string) error {
	return r.ctx.DB.Where("project_id = ? AND team_id = ?", projectId, teamId).
		Delete(&model.ProjectTeamAccess{}).Error
}
