package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/model/entity"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/id"
	"github.com/observabil/arcade/pkg/log"
	"gorm.io/gorm"
)

type TeamService struct {
	ctx      *ctx.Context
	teamRepo *repo.TeamRepo
}

func NewTeamService(ctx *ctx.Context, teamRepo *repo.TeamRepo) *TeamService {
	return &TeamService{
		ctx:      ctx,
		teamRepo: teamRepo,
	}
}

// CreateTeam 创建团队
func (s *TeamService) CreateTeam(req *model.CreateTeamReq, createdBy string) (*model.TeamResp, error) {
	// 1. 验证组织是否存在
	if req.OrgId == "" {
		return nil, errors.New("organization id cannot be empty")
	}

	// 2. 检查团队名称是否已存在
	exists, err := s.teamRepo.CheckTeamNameExists(req.OrgId, req.Name)
	if err != nil {
		log.Errorf("check team name failed: %v", err)
		return nil, fmt.Errorf("check team name failed: %w", err)
	}
	if exists {
		return nil, errors.New("team name already exists")
	}

	// 3. 构建团队路径和层级
	path, level, err := s.teamRepo.BuildTeamPath(req.ParentTeamId)
	if err != nil {
		log.Errorf("build team path failed: %v", err)
		return nil, fmt.Errorf("build team path failed: %w", err)
	}

	// 4. 处理 settings
	settingsJSON, err := repo.ConvertSettingsToJSON(req.Settings)
	if err != nil {
		log.Errorf("convert settings failed: %v", err)
		return nil, fmt.Errorf("convert settings failed: %w", err)
	}

	// 5. 创建团队实体
	team := &entity.Team{
		TeamId:       id.GetUUID(),
		OrgId:        req.OrgId,
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		Avatar:       req.Avatar,
		ParentTeamId: req.ParentTeamId,
		Path:         path,
		Level:        level,
		Settings:     settingsJSON,
		Visibility:   req.Visibility,
		IsEnabled:    1,
	}

	// 设置显示名称默认值
	if team.DisplayName == "" {
		team.DisplayName = team.Name
	}

	// 6. 保存到数据库
	if err := s.teamRepo.CreateTeam(team); err != nil {
		log.Errorf("create team failed: %v", err)
		return nil, fmt.Errorf("create team failed: %w", err)
	}

	log.Infof("success create team: %s (ID: %s)", team.Name, team.TeamId)

	// 7. 返回响应
	return model.EntityToTeamResp(team), nil
}

// UpdateTeam 更新团队
func (s *TeamService) UpdateTeam(teamId string, req *model.UpdateTeamReq) (*model.TeamResp, error) {
	// 1. 检查团队是否存在
	team, err := s.teamRepo.GetTeamById(teamId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team not found")
		}
		return nil, fmt.Errorf("check team name failed: %w", err)
	}

	// 2. 构建更新数据
	updates := make(map[string]interface{})

	if req.Name != nil && *req.Name != "" {
		// 检查新名称是否与其他团队冲突
		exists, err := s.teamRepo.CheckTeamNameExists(team.OrgId, *req.Name, teamId)
		if err != nil {
			return nil, fmt.Errorf("check team name failed: %w", err)
		}
		if exists {
			return nil, errors.New("team name already exists")
		}
		updates["name"] = *req.Name
	}

	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}

	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	if req.Settings != nil {
		settingsJSON, err := repo.ConvertSettingsToJSON(req.Settings)
		if err != nil {
			return nil, fmt.Errorf("convert settings failed: %w", err)
		}
		updates["settings"] = settingsJSON
	}

	// 3. 执行更新
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.teamRepo.UpdateTeam(teamId, updates); err != nil {
			log.Errorf("update team failed: %v", err)
			return nil, fmt.Errorf("update team failed: %w", err)
		}
	}

	// 4. 查询更新后的团队信息
	updatedTeam, err := s.teamRepo.GetTeamById(teamId)
	if err != nil {
		return nil, fmt.Errorf("get updated team failed: %w", err)
	}

	log.Infof("success update team: %s (ID: %s)", updatedTeam.Name, teamId)

	return model.EntityToTeamResp(updatedTeam), nil
}

// DeleteTeam 删除团队
func (s *TeamService) DeleteTeam(teamId string) error {
	// 1. 检查团队是否存在
	team, err := s.teamRepo.GetTeamById(teamId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("team not found")
		}
		return fmt.Errorf("get team failed: %w", err)
	}

	// 2. 检查是否有子团队
	subTeams, err := s.teamRepo.GetSubTeams(teamId)
	if err != nil {
		return fmt.Errorf("get sub teams failed: %w", err)
	}
	if len(subTeams) > 0 {
		return errors.New("sub teams exist, cannot delete")
	}

	// 3. 检查是否有成员（可选）
	if team.TotalMembers > 0 {
		return fmt.Errorf("team has %d members, cannot delete", team.TotalMembers)
	}

	// 4. 检查是否有项目（可选）
	if team.TotalProjects > 0 {
		return fmt.Errorf("team has %d projects, cannot delete", team.TotalProjects)
	}

	// 5. 执行删除
	if err := s.teamRepo.DeleteTeam(teamId); err != nil {
		log.Errorf("delete team failed: %v", err)
		return fmt.Errorf("delete team failed: %w", err)
	}

	log.Infof("success delete team: %s (ID: %s)", team.Name, teamId)

	return nil
}

// GetTeamById 根据ID获取团队
func (s *TeamService) GetTeamById(teamId string) (*model.TeamResp, error) {
	team, err := s.teamRepo.GetTeamById(teamId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team not found")
		}
		return nil, fmt.Errorf("get team failed: %w", err)
	}

	return model.EntityToTeamResp(team), nil
}

// ListTeams 查询团队列表
func (s *TeamService) ListTeams(query *model.TeamQueryReq) (*model.TeamListResp, error) {
	// 设置默认分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 查询团队列表
	teams, total, err := s.teamRepo.ListTeams(query)
	if err != nil {
		log.Errorf("list teams failed: %v", err)
		return nil, fmt.Errorf("list teams failed: %w", err)
	}

	// 转换为响应格式
	teamResps := make([]*model.TeamResp, 0, len(teams))
	for _, team := range teams {
		teamResps = append(teamResps, model.EntityToTeamResp(team))
	}

	// 计算总页数
	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	return &model.TeamListResp{
		Teams:      teamResps,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTeamsByOrgId 获取组织下的所有团队
func (s *TeamService) GetTeamsByOrgId(orgId string) ([]*model.TeamResp, error) {
	teams, err := s.teamRepo.GetTeamsByOrgId(orgId)
	if err != nil {
		return nil, fmt.Errorf("get teams by org id failed: %w", err)
	}

	teamResps := make([]*model.TeamResp, 0, len(teams))
	for _, team := range teams {
		teamResps = append(teamResps, model.EntityToTeamResp(team))
	}

	return teamResps, nil
}

// GetSubTeams 获取子团队
func (s *TeamService) GetSubTeams(parentTeamId string) ([]*model.TeamResp, error) {
	teams, err := s.teamRepo.GetSubTeams(parentTeamId)
	if err != nil {
		return nil, fmt.Errorf("get sub teams failed: %w", err)
	}

	teamResps := make([]*model.TeamResp, 0, len(teams))
	for _, team := range teams {
		teamResps = append(teamResps, model.EntityToTeamResp(team))
	}

	return teamResps, nil
}

// GetTeamsByUserId 获取用户所属的所有团队
func (s *TeamService) GetTeamsByUserId(userId string) ([]*model.TeamResp, error) {
	teams, err := s.teamRepo.GetTeamsByUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("get teams by user id failed: %w", err)
	}

	teamResps := make([]*model.TeamResp, 0, len(teams))
	for _, team := range teams {
		teamResps = append(teamResps, model.EntityToTeamResp(team))
	}

	return teamResps, nil
}

// UpdateTeamStatistics 更新团队统计信息
func (s *TeamService) UpdateTeamStatistics(teamId string) error {
	return s.teamRepo.UpdateTeamStatistics(teamId)
}

// EnableTeam 启用团队
func (s *TeamService) EnableTeam(teamId string) error {
	updates := map[string]interface{}{
		"is_enabled": 1,
		"updated_at": time.Now(),
	}
	return s.teamRepo.UpdateTeam(teamId, updates)
}

// DisableTeam 禁用团队
func (s *TeamService) DisableTeam(teamId string) error {
	updates := map[string]interface{}{
		"is_enabled": 0,
		"updated_at": time.Now(),
	}
	return s.teamRepo.UpdateTeam(teamId, updates)
}
