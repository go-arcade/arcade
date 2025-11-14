package team

import (
	"github.com/go-arcade/arcade/internal/engine/model/team"
	"github.com/go-arcade/arcade/pkg/database"
)


type ITeamMemberRepository interface {
	GetTeamMember(teamId, userId string) (*team.TeamMember, error)
	ListTeamMembers(teamId string) ([]team.TeamMember, error)
	ListUserTeams(userId string) ([]team.TeamMember, error)
	AddTeamMember(member *team.TeamMember) error
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
func (r *TeamMemberRepo) GetTeamMember(teamId, userId string) (*team.TeamMember, error) {
	var member team.TeamMember
	err := r.db.DB().Where("team_id = ? AND user_id = ?", teamId, userId).First(&member).Error
	return &member, err
}

// ListTeamMembers 列出团队成员
func (r *TeamMemberRepo) ListTeamMembers(teamId string) ([]team.TeamMember, error) {
	var members []team.TeamMember
	err := r.db.DB().Where("team_id = ?", teamId).Find(&members).Error
	return members, err
}

// ListUserTeams 列出用户所在的团队
func (r *TeamMemberRepo) ListUserTeams(userId string) ([]team.TeamMember, error) {
	var members []team.TeamMember
	err := r.db.DB().Where("user_id = ?", userId).Find(&members).Error
	return members, err
}

// AddTeamMember 添加团队成员
func (r *TeamMemberRepo) AddTeamMember(member *team.TeamMember) error {
	return r.db.DB().Create(member).Error
}

// UpdateTeamMemberRole 更新团队成员角色
func (r *TeamMemberRepo) UpdateTeamMemberRole(teamId, userId, role string) error {
	return r.db.DB().Model(&team.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamId, userId).
		Update("role", role).Error
}

// RemoveTeamMember 移除团队成员
func (r *TeamMemberRepo) RemoveTeamMember(teamId, userId string) error {
	return r.db.DB().Where("team_id = ? AND user_id = ?", teamId, userId).
		Delete(&team.TeamMember{}).Error
}
