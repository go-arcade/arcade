package model

import "github.com/observabil/arcade/pkg/datatype"

// Organization 组织表
type Organization struct {
	BaseModel
	OrgId       string        `gorm:"column:org_id" json:"orgId"`                // 组织唯一标识
	Name        string        `gorm:"column:name" json:"name"`                   // 组织名称
	DisplayName string        `gorm:"column:display_name" json:"displayName"`    // 组织显示名称
	Description string        `gorm:"column:description" json:"description"`     // 组织描述
	Logo        string        `gorm:"column:logo" json:"logo"`                   // 组织Logo URL
	Website     string        `gorm:"column:website" json:"website"`             // 组织官网
	Email       string        `gorm:"column:email" json:"email"`                 // 组织联系邮箱
	Phone       string        `gorm:"column:phone" json:"phone"`                 // 组织联系电话
	Address     string        `gorm:"column:address" json:"address"`             // 组织地址
	Settings    datatype.JSON `gorm:"column:settings;type:json" json:"settings"` // 组织设置
	Plan        string        `gorm:"column:plan" json:"plan"`                   // 订阅计划(free/pro/enterprise)
	Status      int           `gorm:"column:status" json:"status"`               // 状态: 0-未激活, 1-正常, 2-冻结, 3-已删除
	OwnerUserId string        `gorm:"column:owner_user_id" json:"ownerUserId"`   // 组织所有者用户ID
	IsEnabled   int           `gorm:"column:is_enabled" json:"isEnabled"`        // 是否启用: 0-禁用, 1-启用

	// 统计字段
	TotalMembers  int `gorm:"column:total_members" json:"totalMembers"`   // 成员总数
	TotalTeams    int `gorm:"column:total_teams" json:"totalTeams"`       // 团队总数
	TotalProjects int `gorm:"column:total_projects" json:"totalProjects"` // 项目总数
}

func (Organization) TableName() string {
	return "t_organization"
}

// OrganizationSettings 组织设置结构
type OrganizationSettings struct {
	AllowPublicProjects bool     `json:"allow_public_projects"` // 允许公开项目
	RequireTwoFactor    bool     `json:"require_two_factor"`    // 要求双因素认证
	AllowedDomains      []string `json:"allowed_domains"`       // 允许的邮箱域名
	MaxMembers          int      `json:"max_members"`           // 最大成员数
	MaxProjects         int      `json:"max_projects"`          // 最大项目数
	MaxTeams            int      `json:"max_teams"`             // 最大团队数
	MaxAgents           int      `json:"max_agents"`            // 最大Agent数
	DefaultVisibility   string   `json:"default_visibility"`    // 默认项目可见性
	EnableSAML          bool     `json:"enable_saml"`           // 启用SAML认证
	EnableLDAP          bool     `json:"enable_ldap"`           // 启用LDAP认证
}

// OrganizationStatus 组织状态枚举
const (
	OrgStatusInactive = 0 // 未激活
	OrgStatusActive   = 1 // 正常
	OrgStatusFrozen   = 2 // 冻结
	OrgStatusDeleted  = 3 // 已删除
)

// OrganizationPlan 订阅计划
const (
	PlanFree       = "free"       // 免费版
	PlanPro        = "pro"        // 专业版
	PlanEnterprise = "enterprise" // 企业版
)

// Team 团队表
type Team struct {
	BaseModel
	TeamId       string        `gorm:"column:team_id" json:"teamId"`              // 团队唯一标识
	OrgId        string        `gorm:"column:org_id" json:"orgId"`                // 所属组织ID
	Name         string        `gorm:"column:name" json:"name"`                   // 团队名称
	DisplayName  string        `gorm:"column:display_name" json:"displayName"`    // 团队显示名称
	Description  string        `gorm:"column:description" json:"description"`     // 团队描述
	Avatar       string        `gorm:"column:avatar" json:"avatar"`               // 团队头像
	ParentTeamId string        `gorm:"column:parent_team_id" json:"parentTeamId"` // 父团队ID(支持嵌套)
	Path         string        `gorm:"column:path" json:"path"`                   // 团队路径(用于层级关系)
	Level        int           `gorm:"column:level" json:"level"`                 // 团队层级
	Settings     datatype.JSON `gorm:"column:settings;type:json" json:"settings"` // 团队设置
	Visibility   int           `gorm:"column:visibility" json:"visibility"`       // 可见性: 0-私有, 1-内部, 2-公开
	IsEnabled    int           `gorm:"column:is_enabled" json:"isEnabled"`        // 是否启用: 0-禁用, 1-启用

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

// OrganizationMember 组织成员表
type OrganizationMember struct {
	BaseModel
	OrgId     string `gorm:"column:org_id" json:"orgId"`         // 组织ID
	UserId    string `gorm:"column:user_id" json:"userId"`       // 用户ID
	Role      string `gorm:"column:role" json:"role"`            // 组织角色(owner/admin/member)
	Username  string `gorm:"column:username" json:"username"`    // 用户名(冗余)
	Email     string `gorm:"column:email" json:"email"`          // 邮箱(冗余)
	InvitedBy string `gorm:"column:invited_by" json:"invitedBy"` // 邀请人用户ID
	Status    int    `gorm:"column:status" json:"status"`        // 状态: 0-待接受, 1-正常, 2-禁用
}

func (OrganizationMember) TableName() string {
	return "t_organization_member"
}

// OrganizationMemberRole 组织成员角色
const (
	OrgRoleOwner  = "owner"  // 所有者(最高权限)
	OrgRoleAdmin  = "admin"  // 管理员(管理组织、成员、团队)
	OrgRoleMember = "member" // 普通成员
)

// OrganizationMemberStatus 组织成员状态
const (
	OrgMemberStatusPending  = 0 // 待接受
	OrgMemberStatusActive   = 1 // 正常
	OrgMemberStatusDisabled = 2 // 禁用
)

// TeamMember 团队成员表
type TeamMember struct {
	BaseModel
	TeamId    string `gorm:"column:team_id" json:"teamId"`       // 团队ID
	UserId    string `gorm:"column:user_id" json:"userId"`       // 用户ID
	Role      string `gorm:"column:role" json:"role"`            // 团队角色(owner/maintainer/developer/reporter/guest)
	Username  string `gorm:"column:username" json:"username"`    // 用户名(冗余)
	InvitedBy string `gorm:"column:invited_by" json:"invitedBy"` // 邀请人用户ID
}

func (TeamMember) TableName() string {
	return "t_team_member"
}

// TeamMemberRole 团队成员角色
const (
	TeamRoleOwner      = "owner"      // 所有者
	TeamRoleMaintainer = "maintainer" // 维护者
	TeamRoleDeveloper  = "developer"  // 开发者
	TeamRoleReporter   = "reporter"   // 报告者
	TeamRoleGuest      = "guest"      // 访客
)

// ProjectTeamRelation 项目团队关联表
type ProjectTeamRelation struct {
	BaseModel
	ProjectId string `gorm:"column:project_id" json:"projectId"` // 项目ID
	TeamId    string `gorm:"column:team_id" json:"teamId"`       // 团队ID
	Access    string `gorm:"column:access" json:"access"`        // 访问权限(read/write/admin)
}

func (ProjectTeamRelation) TableName() string {
	return "t_project_team_relation"
}

// ProjectTeamAccess 项目团队访问权限
const (
	AccessRead  = "read"  // 只读
	AccessWrite = "write" // 读写
	AccessAdmin = "admin" // 管理员
)

// OrganizationInvitation 组织邀请表
type OrganizationInvitation struct {
	BaseModel
	InvitationId string `gorm:"column:invitation_id" json:"invitationId"` // 邀请唯一标识
	OrgId        string `gorm:"column:org_id" json:"orgId"`               // 组织ID
	Email        string `gorm:"column:email" json:"email"`                // 被邀请人邮箱
	Role         string `gorm:"column:role" json:"role"`                  // 角色
	Token        string `gorm:"column:token" json:"token"`                // 邀请令牌
	InvitedBy    string `gorm:"column:invited_by" json:"invitedBy"`       // 邀请人用户ID
	Status       int    `gorm:"column:status" json:"status"`              // 状态: 0-待接受, 1-已接受, 2-已拒绝, 3-已过期
	ExpiresAt    string `gorm:"column:expires_at" json:"expiresAt"`       // 过期时间
}

func (OrganizationInvitation) TableName() string {
	return "t_organization_invitation"
}

// InvitationStatus 邀请状态
const (
	InvitationStatusPending  = 0 // 待接受
	InvitationStatusAccepted = 1 // 已接受
	InvitationStatusRejected = 2 // 已拒绝
	InvitationStatusExpired  = 3 // 已过期
)
