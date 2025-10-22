package repo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/model/entity"
	"github.com/observabil/arcade/pkg/ctx"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TeamRepo struct {
	ctx *ctx.Context
}

func NewTeamRepo(ctx *ctx.Context) *TeamRepo {
	return &TeamRepo{ctx: ctx}
}

// CreateTeam 创建团队
func (r *TeamRepo) CreateTeam(team *entity.Team) error {
	return r.ctx.DB.Create(team).Error
}

// UpdateTeam 更新团队
func (r *TeamRepo) UpdateTeam(teamId string, updates map[string]interface{}) error {
	return r.ctx.DB.Model(&entity.Team{}).
		Where("team_id = ?", teamId).
		Updates(updates).Error
}

// DeleteTeam 删除团队（软删除或硬删除）
func (r *TeamRepo) DeleteTeam(teamId string) error {
	return r.ctx.DB.Where("team_id = ?", teamId).Delete(&entity.Team{}).Error
}

// GetTeamById 根据团队ID获取团队信息
func (r *TeamRepo) GetTeamById(teamId string) (*entity.Team, error) {
	var team entity.Team
	err := r.ctx.DB.Where("team_id = ?", teamId).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetTeamByName 根据团队名称和组织ID获取团队
func (r *TeamRepo) GetTeamByName(orgId, name string) (*entity.Team, error) {
	var team entity.Team
	err := r.ctx.DB.Where("org_id = ? AND name = ?", orgId, name).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// ListTeams 查询团队列表
func (r *TeamRepo) ListTeams(query *model.TeamQueryReq) ([]*entity.Team, int64, error) {
	var teams []*entity.Team
	var total int64

	db := r.ctx.DB.Model(&entity.Team{})

	// 条件查询
	if query.OrgId != "" {
		db = db.Where("org_id = ?", query.OrgId)
	}
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.ParentTeamId != "" {
		db = db.Where("parent_team_id = ?", query.ParentTeamId)
	}
	if query.Visibility != nil {
		db = db.Where("visibility = ?", *query.Visibility)
	}
	if query.IsEnabled != nil {
		db = db.Where("is_enabled = ?", *query.IsEnabled)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	} else {
		// 默认分页
		db = db.Limit(100)
	}

	// 查询结果，按创建时间降序
	err := db.Order("created_at DESC").Find(&teams).Error
	return teams, total, err
}

// GetTeamsByOrgId 获取组织下的所有团队
func (r *TeamRepo) GetTeamsByOrgId(orgId string) ([]*entity.Team, error) {
	var teams []*entity.Team
	err := r.ctx.DB.Where("org_id = ? AND is_enabled = ?", orgId, 1).
		Order("level ASC, created_at DESC").
		Find(&teams).Error
	return teams, err
}

// GetSubTeams 获取子团队
func (r *TeamRepo) GetSubTeams(parentTeamId string) ([]*entity.Team, error) {
	var teams []*entity.Team
	err := r.ctx.DB.Where("parent_team_id = ? AND is_enabled = ?", parentTeamId, 1).
		Order("created_at DESC").
		Find(&teams).Error
	return teams, err
}

// CheckTeamExists 检查团队是否存在
func (r *TeamRepo) CheckTeamExists(teamId string) (bool, error) {
	var count int64
	err := r.ctx.DB.Model(&entity.Team{}).
		Where("team_id = ?", teamId).
		Count(&count).Error
	return count > 0, err
}

// CheckTeamNameExists 检查团队名称在组织内是否已存在
func (r *TeamRepo) CheckTeamNameExists(orgId, name string, excludeTeamId ...string) (bool, error) {
	query := r.ctx.DB.Model(&entity.Team{}).
		Where("org_id = ? AND name = ?", orgId, name)

	if len(excludeTeamId) > 0 && excludeTeamId[0] != "" {
		query = query.Where("team_id != ?", excludeTeamId[0])
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// UpdateTeamPath 更新团队路径
func (r *TeamRepo) UpdateTeamPath(teamId, path string, level int) error {
	return r.ctx.DB.Model(&entity.Team{}).
		Where("team_id = ?", teamId).
		Updates(map[string]interface{}{
			"path":  path,
			"level": level,
		}).Error
}

// IncrementTeamMembers 增加团队成员数量
func (r *TeamRepo) IncrementTeamMembers(teamId string, delta int) error {
	return r.ctx.DB.Model(&entity.Team{}).
		Where("team_id = ?", teamId).
		Update("total_members", gorm.Expr("total_members + ?", delta)).Error
}

// IncrementTeamProjects 增加团队项目数量
func (r *TeamRepo) IncrementTeamProjects(teamId string, delta int) error {
	return r.ctx.DB.Model(&entity.Team{}).
		Where("team_id = ?", teamId).
		Update("total_projects", gorm.Expr("total_projects + ?", delta)).Error
}

// UpdateTeamStatistics 更新团队统计信息
func (r *TeamRepo) UpdateTeamStatistics(teamId string) error {
	// 更新成员数量
	var memberCount int64
	if err := r.ctx.DB.Model(&model.TeamMember{}).
		Where("team_id = ?", teamId).
		Count(&memberCount).Error; err != nil {
		return err
	}

	// 更新项目数量（假设有团队项目关联表）
	var projectCount int64
	r.ctx.DB.Table("t_project_team_relation").
		Where("team_id = ?", teamId).
		Count(&projectCount)

	return r.ctx.DB.Model(&entity.Team{}).
		Where("team_id = ?", teamId).
		Updates(map[string]interface{}{
			"total_members":  memberCount,
			"total_projects": projectCount,
		}).Error
}

// BuildTeamPath 构建团队路径
func (r *TeamRepo) BuildTeamPath(parentTeamId string) (string, int, error) {
	if parentTeamId == "" {
		return "/", 0, nil
	}

	parent, err := r.GetTeamById(parentTeamId)
	if err != nil {
		return "", 0, fmt.Errorf("parent team not found: %w", err)
	}

	path := strings.TrimSuffix(parent.Path, "/") + "/" + parentTeamId + "/"
	level := parent.Level + 1

	return path, level, nil
}

// ConvertSettingsToJSON 将 settings map 转换为 JSON
func ConvertSettingsToJSON(settings map[string]interface{}) (datatypes.JSON, error) {
	if settings == nil {
		return datatypes.JSON("{}"), nil
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(data), nil
}

// BatchGetTeams 批量获取团队信息
func (r *TeamRepo) BatchGetTeams(teamIds []string) ([]*entity.Team, error) {
	if len(teamIds) == 0 {
		return []*entity.Team{}, nil
	}

	var teams []*entity.Team
	err := r.ctx.DB.Where("team_id IN ?", teamIds).Find(&teams).Error
	return teams, err
}

// GetTeamsByUserId 获取用户所属的所有团队
func (r *TeamRepo) GetTeamsByUserId(userId string) ([]*entity.Team, error) {
	var teams []*entity.Team
	err := r.ctx.DB.Table("t_team t").
		Select("t.*").
		Joins("JOIN t_team_member tm ON t.team_id = tm.team_id").
		Where("tm.user_id = ? AND t.is_enabled = ?", userId, 1).
		Order("t.created_at DESC").
		Find(&teams).Error
	return teams, err
}
