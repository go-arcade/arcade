package model

import (
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:19
 * @file: model_user.go
 * @description: user model
 */

type User struct {
	BaseModel
	UserId    string `gorm:"column:user_id" json:"userId"`
	Username  string `gorm:"column:username" json:"username"`
	Nickname  string `gorm:"column:nick_name" json:"nickname"`
	Password  string `gorm:"column:password" json:"password"`
	Avatar    string `gorm:"column:avatar" json:"avatar"`
	Email     string `gorm:"column:email" json:"email"`
	Phone     string `gorm:"column:phone" json:"phone"`
	IsEnabled int    `gorm:"column:is_enabled" json:"isEnabled"` // 1: enable, 0: disable，default value is 1
}

func (User) TableName() string {
	return "user"
}

type Register struct {
	UserId     string    `json:"userId"`
	Username   string    `json:"username"`
	Nickname   string    `gorm:"column:nick_name" json:"nickname"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	IsEnabled  int       `gorm:"column:is_enabled" json:"isEnabled"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
}

type Login struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token map[string]string `json:"token"`
	Role  map[string]string `json:"role"`
}

type AddUserReq struct {
	UserId     string    `json:"userId"`
	Username   string    `json:"username"`
	Nickname   string    `gorm:"column:nick_name" json:"nickname"`
	Password   string    `json:"password"`
	Avatar     string    `json:"avatar"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	IsEnabled  int       `gorm:"column:is_enabled" json:"isEnabled"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
}
