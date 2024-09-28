package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:19
 * @file: model_user.go
 * @description: user model
 */

type User struct {
	BaseModel
	UserId   string `gorm:"column:user_id" json:"userId"`
	Username string `gorm:"column:user_name" json:"username"`
	Nickname string `gorm:"column:nick_name" json:"nickname"`
	Password string `gorm:"column:password" json:"password"`
	Avatar   string `gorm:"column:avatar" json:"avatar"`
	Email    string `gorm:"column:email" json:"email"`
	Phone    string `gorm:"column:phone" json:"phone"`
	IsEnable int    `gorm:"column:is_enable" json:"is_enable"`
}

func (User) TableName() string {
	return "user"
}
