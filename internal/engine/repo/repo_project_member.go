package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: repo_project_member.go
 * @description: 项目成员仓储
 */

type ProjectMemberRepo struct {
	ctx *ctx.Context
}

func NewProjectMemberRepo(ctx *ctx.Context) *ProjectMemberRepo {
	return &ProjectMemberRepo{ctx: ctx}
}

// GetProjectMember 获取项目成员
func (r *ProjectMemberRepo) GetProjectMember(projectId, userId string) (*model.ProjectMember, error) {
	var member model.ProjectMember
	err := r.ctx.DBSession().Where("project_id = ? AND user_id = ?", projectId, userId).First(&member).Error
	return &member, err
}

// ListProjectMembers 列出项目成员
func (r *ProjectMemberRepo) ListProjectMembers(projectId string) ([]model.ProjectMember, error) {
	var members []model.ProjectMember
	err := r.ctx.DBSession().Where("project_id = ?", projectId).Find(&members).Error
	return members, err
}

// AddProjectMember 添加项目成员
func (r *ProjectMemberRepo) AddProjectMember(member *model.ProjectMember) error {
	return r.ctx.DBSession().Create(member).Error
}

// UpdateProjectMemberRole 更新项目成员角色
func (r *ProjectMemberRepo) UpdateProjectMemberRole(projectId, userId, role string) error {
	return r.ctx.DBSession().Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectId, userId).
		Update("role", role).Error
}

// RemoveProjectMember 移除项目成员
func (r *ProjectMemberRepo) RemoveProjectMember(projectId, userId string) error {
	return r.ctx.DBSession().Where("project_id = ? AND user_id = ?", projectId, userId).
		Delete(&model.ProjectMember{}).Error
}

// GetUserProjects 获取用户的所有项目
func (r *ProjectMemberRepo) GetUserProjects(userId string) ([]model.ProjectMember, error) {
	var members []model.ProjectMember
	err := r.ctx.DBSession().Where("user_id = ?", userId).Find(&members).Error
	return members, err
}
