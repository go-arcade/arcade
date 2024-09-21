package model

import "time"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:19
 * @file: model_user.go
 * @description: user model
 */

type User struct {
	ID       int       `gorm:"primaryKey" json:"id"`
	UserId   string    `gorm:"column:user_id" json:"userId"`
	Username string    `gorm:"column:username" json:"username"`
	Password string    `gorm:"column:password" json:"password"`
	Nickname string    `gorm:"column:nickname" json:"nickname"`
	Email    string    `gorm:"column:email" json:"email"`
	Phone    string    `gorm:"column:phone" json:"phone"`
	IsEnable int       `gorm:"column:enable" json:"isEnable"`
	UserRole string    `gorm:"column:user_role" json:"userRole"`
	CreatAt  time.Time `gorm:"column:creat_time" json:"creatAt"`
	UpdateAt time.Time `gorm:"column:update_time" json:"updateAt"`
}

func (User) TableName() string {
	return "user"
}
