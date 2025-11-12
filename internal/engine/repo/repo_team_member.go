package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: repo_team_member.go
 * @description: 团队成员仓储
 */

type ITeamMemberRepository interface {
	GetTeamMember(teamId, userId string) (*model.TeamMember, error)
	ListTeamMembers(teamId string) ([]model.TeamMember, error)
	ListUserTeams(userId string) ([]model.TeamMember, error)
	AddTeamMember(member *model.TeamMember) error
	UpdateTeamMemberRole(teamId, userId, role string) error
	RemoveTeamMember(teamId, userId string) error
}

type TeamMemberRepo struct {
	db database.DB
}

func NewTeamMemberRepo(db database.DB) ITeamMemberRepository {
	return &TeamMemberRepo{db: db}
}

// GetTeamMember 获取团队成员
func (r *TeamMemberRepo) GetTeamMember(teamId, userId string) (*model.TeamMember, error) {
	var member model.TeamMember
	err := r.db.DB().Where("team_id = ? AND user_id = ?", teamId, userId).First(&member).Error
	return &member, err
}

// ListTeamMembers 列出团队成员
func (r *TeamMemberRepo) ListTeamMembers(teamId string) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := r.db.DB().Where("team_id = ?", teamId).Find(&members).Error
	return members, err
}

// ListUserTeams 列出用户所在的团队
func (r *TeamMemberRepo) ListUserTeams(userId string) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := r.db.DB().Where("user_id = ?", userId).Find(&members).Error
	return members, err
}

// AddTeamMember 添加团队成员
func (r *TeamMemberRepo) AddTeamMember(member *model.TeamMember) error {
	return r.db.DB().Create(member).Error
}

// UpdateTeamMemberRole 更新团队成员角色
func (r *TeamMemberRepo) UpdateTeamMemberRole(teamId, userId, role string) error {
	return r.db.DB().Model(&model.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamId, userId).
		Update("role", role).Error
}

// RemoveTeamMember 移除团队成员
func (r *TeamMemberRepo) RemoveTeamMember(teamId, userId string) error {
	return r.db.DB().Where("team_id = ? AND user_id = ?", teamId, userId).
		Delete(&model.TeamMember{}).Error
}
