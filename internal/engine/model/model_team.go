package model

import "gorm.io/datatypes"

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: model_team.go
 * @description: 团队表模型
 */

// Team 团队表
type Team struct {
	BaseModel
	TeamId       string         `gorm:"column:team_id" json:"teamId"`              // 团队唯一标识
	OrgId        string         `gorm:"column:org_id" json:"orgId"`                // 所属组织ID
	Name         string         `gorm:"column:name" json:"name"`                   // 团队名称
	DisplayName  string         `gorm:"column:display_name" json:"displayName"`    // 团队显示名称
	Description  string         `gorm:"column:description" json:"description"`     // 团队描述
	Avatar       string         `gorm:"column:avatar" json:"avatar"`               // 团队头像
	ParentTeamId string         `gorm:"column:parent_team_id" json:"parentTeamId"` // 父团队ID(支持嵌套)
	Path         string         `gorm:"column:path" json:"path"`                   // 团队路径(用于层级关系)
	Level        int            `gorm:"column:level" json:"level"`                 // 团队层级
	Settings     datatypes.JSON `gorm:"column:settings;type:json" json:"settings"` // 团队设置
	Visibility   int            `gorm:"column:visibility" json:"visibility"`       // 可见性: 0-私有, 1-内部, 2-公开
	IsEnabled    int            `gorm:"column:is_enabled" json:"isEnabled"`        // 是否启用: 0-禁用, 1-启用

	// 统计字段
	TotalMembers  int `gorm:"column:total_members" json:"totalMembers"`   // 成员总数
	TotalProjects int `gorm:"column:total_projects" json:"totalProjects"` // 项目总数
}

func (Team) TableName() string {
	return "t_team"
}

// TeamSettings 团队设置结构
type TeamSettings struct {
	DefaultRole       string `json:"default_role"`        // 默认角色
	AllowMemberInvite bool   `json:"allow_member_invite"` // 允许成员邀请
	RequireApproval   bool   `json:"require_approval"`    // 需要审批
	MaxMembers        int    `json:"max_members"`         // 最大成员数
}

// TeamVisibility 团队可见性枚举
const (
	TeamVisibilityPrivate  = 0 // 私有(仅成员可见)
	TeamVisibilityInternal = 1 // 内部(组织内可见)
	TeamVisibilityPublic   = 2 // 公开(所有人可见)
)
