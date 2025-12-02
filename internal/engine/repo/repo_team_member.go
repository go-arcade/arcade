package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type ITeamMemberRepository interface {
	GetTeamMember(teamId, userId string) (*model.TeamMember, error)
	ListTeamMembers(teamId string) ([]model.TeamMember, error)
	ListUserTeams(userId string) ([]model.TeamMember, error)
	AddTeamMember(member *model.TeamMember) error
	UpdateTeamMemberRole(teamId, userId, role string) error
	RemoveTeamMember(teamId, userId string) error
}

type TeamMemberRepo struct {
	db database.IDatabase
}

func NewTeamMemberRepo(db database.IDatabase) ITeamMemberRepository {
	return &TeamMemberRepo{db: db}
}

// GetTeamMember 获取团队成员
func (r *TeamMemberRepo) GetTeamMember(teamId, userId string) (*model.TeamMember, error) {
	var member model.TeamMember
	err := r.db.Database().Where("team_id = ? AND user_id = ?", teamId, userId).First(&member).Error
	return &member, err
}

// ListTeamMembers 列出团队成员
func (r *TeamMemberRepo) ListTeamMembers(teamId string) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := r.db.Database().Where("team_id = ?", teamId).Find(&members).Error
	return members, err
}

// ListUserTeams 列出用户所在的团队
func (r *TeamMemberRepo) ListUserTeams(userId string) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := r.db.Database().Where("user_id = ?", userId).Find(&members).Error
	return members, err
}

// AddTeamMember 添加团队成员
func (r *TeamMemberRepo) AddTeamMember(member *model.TeamMember) error {
	return r.db.Database().Create(member).Error
}

// UpdateTeamMemberRole 更新团队成员角色
func (r *TeamMemberRepo) UpdateTeamMemberRole(teamId, userId, role string) error {
	return r.db.Database().Model(&model.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamId, userId).
		Update("role", role).Error
}

// RemoveTeamMember 移除团队成员
func (r *TeamMemberRepo) RemoveTeamMember(teamId, userId string) error {
	return r.db.Database().Where("team_id = ? AND user_id = ?", teamId, userId).
		Delete(&model.TeamMember{}).Error
}
