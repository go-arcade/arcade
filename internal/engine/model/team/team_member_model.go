package team

import (
	"github.com/go-arcade/arcade/internal/engine/model"
)


// TeamMemberRole 团队成员角色（兼容旧版，保留用于快速角色名称引用）
const (
	TeamRoleOwner      = "owner"      // 所有者(完全控制团队)
	TeamRoleMaintainer = "maintainer" // 维护者(管理团队成员和项目)
	TeamRoleDeveloper  = "developer"  // 开发者(参与开发)
	TeamRoleReporter   = "reporter"   // 报告者(报告问题)
	TeamRoleGuest      = "guest"      // 访客(仅查看)
)

// TeamMember 团队成员表
type TeamMember struct {
	model.BaseModel
	TeamId string `gorm:"column:team_id;not null;index:idx_team_user,unique" json:"teamId"`
	UserId string `gorm:"column:user_id;not null;index:idx_team_user,unique;index:idx_user" json:"userId"`
	RoleId string `gorm:"column:role_id;not null;index" json:"roleId"` // 角色ID（引用 t_role 表）
}

func (TeamMember) TableName() string {
	return "t_team_member"
}
