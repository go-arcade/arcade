package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/11 21:54
 * @file: model_role.go
 * @description: role model
 */

type Role struct {
	BaseModel
	RoleId   string `gorm:"column:role_id" json:"roleId"`
	RoleName string `gorm:"column:role_name" json:"roleName"`
	RoleCode string `gorm:"column:role_code" json:"roleCode"`
	RoleDesc string `gorm:"column:role_desc" json:"roleDesc"`
}

func (r *Role) TableName() string {
	return "t_role"
}
