package model

import (
	"time"
)

// UserExtension user extension information table
type UserExtension struct {
	BaseModel
	UserId           string     `gorm:"column:user_id;uniqueIndex" json:"userId"`                           // user ID (foreign key)
	Timezone         string     `gorm:"column:timezone;default:'UTC'" json:"timezone"`                      // user timezone (e.g., 'Asia/Shanghai', 'America/New_York')
	LastLoginAt      *time.Time `gorm:"column:last_login_at" json:"lastLoginAt"`                            // last login timestamp
	InvitationStatus string     `gorm:"column:invitation_status;default:'pending'" json:"invitationStatus"` // invitation status: pending, accepted, expired, revoked
	InvitedBy        string     `gorm:"column:invited_by" json:"invitedBy"`                                 // invited by user ID
	InvitedAt        *time.Time `gorm:"column:invited_at" json:"invitedAt"`                                 // invitation timestamp
	AcceptedAt       *time.Time `gorm:"column:accepted_at" json:"acceptedAt"`                               // invitation accepted timestamp
}

func (UserExtension) TableName() string {
	return "t_user_ext"
}

// UserInvitationStatus constants for user extension
const (
	UserInvitationStatusPending  = "pending"  // pending acceptance
	UserInvitationStatusAccepted = "accepted" // accepted
	UserInvitationStatusExpired  = "expired"  // expired
	UserInvitationStatusRevoked  = "revoked"  // revoked by admin
)
