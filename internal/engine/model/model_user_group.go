package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 21:54
 * @file: model_user_group.go
 * @description: user group model
 */

type UserGroup struct {
	BaseModel
	GroupId   string `gorm:"column:group_id" json:"groupId"`
	GroupName string `gorm:"column:group_name" json:"groupName"`
}

func (UserGroup) TableName() string {
	return "t_user_group"
}
