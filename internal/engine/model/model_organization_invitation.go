package model

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
