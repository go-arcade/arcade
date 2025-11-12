package orgs

import "github.com/go-arcade/arcade/internal/engine/model"

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: model_organization_member.go
 * @description: 组织成员表模型
 */

// OrganizationMember 组织成员表
type OrganizationMember struct {
	model.BaseModel
	OrgId     string `gorm:"column:org_id" json:"orgId"`         // 组织ID
	UserId    string `gorm:"column:user_id" json:"userId"`       // 用户ID
	RoleId    string `gorm:"column:role_id;index" json:"roleId"` // 角色ID（引用 t_role 表）
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
