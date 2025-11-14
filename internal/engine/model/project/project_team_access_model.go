package project

import (
	"github.com/go-arcade/arcade/internal/engine/model"
)


// ProjectTeamAccess 项目团队访问权限表
type ProjectTeamAccess struct {
	model.BaseModel
	ProjectId   string `gorm:"column:project_id;not null;index:idx_project_team,unique" json:"projectId"`
	TeamId      string `gorm:"column:team_id;not null;index:idx_project_team,unique;index:idx_team" json:"teamId"`
	AccessLevel string `gorm:"column:access_level;not null;type:varchar(32)" json:"accessLevel"` // read/write/admin
}

func (ProjectTeamAccess) TableName() string {
	return "t_project_team_access"
}

// ProjectTeamAccessLevel 项目团队访问权限级别
const (
	AccessLevelRead  = "read"  // 只读(查看代码和构建)
	AccessLevelWrite = "write" // 读写(提交代码、触发构建)
	AccessLevelAdmin = "admin" // 管理员(完全控制项目)
)
