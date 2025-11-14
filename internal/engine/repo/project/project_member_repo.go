package project

import (
	"github.com/go-arcade/arcade/internal/engine/model/project"
	"github.com/go-arcade/arcade/pkg/database"
)


type IProjectMemberRepository interface {
	GetProjectMember(projectId, userId string) (*project.ProjectMember, error)
	ListProjectMembers(projectId string) ([]project.ProjectMember, error)
	AddProjectMember(member *project.ProjectMember) error
	UpdateProjectMemberRole(projectId, userId, role string) error
	RemoveProjectMember(projectId, userId string) error
	GetUserProjects(userId string) ([]project.ProjectMember, error)
}

type ProjectMemberRepo struct {
	db database.DB
}

func NewProjectMemberRepo(db database.DB) IProjectMemberRepository {
	return &ProjectMemberRepo{db: db}
}

// GetProjectMember 获取项目成员
func (r *ProjectMemberRepo) GetProjectMember(projectId, userId string) (*project.ProjectMember, error) {
	var member project.ProjectMember
	err := r.db.DB().Where("project_id = ? AND user_id = ?", projectId, userId).First(&member).Error
	return &member, err
}

// ListProjectMembers 列出项目成员
func (r *ProjectMemberRepo) ListProjectMembers(projectId string) ([]project.ProjectMember, error) {
	var members []project.ProjectMember
	err := r.db.DB().Where("project_id = ?", projectId).Find(&members).Error
	return members, err
}

// AddProjectMember 添加项目成员
func (r *ProjectMemberRepo) AddProjectMember(member *project.ProjectMember) error {
	return r.db.DB().Create(member).Error
}

// UpdateProjectMemberRole 更新项目成员角色
func (r *ProjectMemberRepo) UpdateProjectMemberRole(projectId, userId, role string) error {
	return r.db.DB().Model(&project.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectId, userId).
		Update("role", role).Error
}

// RemoveProjectMember 移除项目成员
func (r *ProjectMemberRepo) RemoveProjectMember(projectId, userId string) error {
	return r.db.DB().Where("project_id = ? AND user_id = ?", projectId, userId).
		Delete(&project.ProjectMember{}).Error
}

// GetUserProjects 获取用户的所有项目
func (r *ProjectMemberRepo) GetUserProjects(userId string) ([]project.ProjectMember, error) {
	var members []project.ProjectMember
	err := r.db.DB().Where("user_id = ?", userId).Find(&members).Error
	return members, err
}
