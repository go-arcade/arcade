package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/11 21:52
 * @file: model_role_relation.go
 * @description: role relation model
 */

type RoleRelation struct {
	BaseModel
	RoleId  string `gorm:"column:role_id" json:"roleId"`
	UserId  string `gorm:"column:user_id" json:"userId"`
	GroupId string `gorm:"column:group_id" json:"groupId"`
}

func (rr *RoleRelation) TableName() string {
	return "t_role_relation"
}
