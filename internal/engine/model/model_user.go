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
	UserId       string `gorm:"column:user_id" json:"userId"`
	Username     string `gorm:"column:username" json:"username"`
	FirstName    string `gorm:"column:first_name" json:"firstName"`
	LastName     string `gorm:"column:last_name" json:"lastName"`
	Password     string `gorm:"column:password" json:"password"`
	Avatar       string `gorm:"column:avatar" json:"avatar"`
	Email        string `gorm:"column:email" json:"email"`
	Phone        string `gorm:"column:phone" json:"phone"`
	IsEnabled    int    `gorm:"column:is_enabled" json:"isEnabled"`                 // 0: disabled, 1: enabled
	IsSuperAdmin int    `gorm:"column:is_superadmin;default:0" json:"isSuperAdmin"` // 0: normal user, 1: super admin
}

func (User) TableName() string {
	return "t_user"
}

type Register struct {
	UserId     string    `json:"userId"`
	Username   string    `json:"username"`
	FirstName  string    `gorm:"column:first_name" json:"firstName"`
	LastName   string    `gorm:"column:last_name" json:"lastName"`
	Email      string    `json:"email"`
	Avatar     string    `gorm:"column:avatar" json:"avatar"`
	Password   string    `json:"password"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
}

type Login struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResp struct {
	UserInfo UserInfo          `json:"userInfo"`
	Token    map[string]string `json:"token"`
	Role     map[string]string `json:"role"`
	ExpireAt int64             `json:"-"`
	CreateAt int64             `json:"-"`
}

type UserInfo struct {
	UserId    string `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type AddUserReq struct {
	UserId     string    `json:"userId"`
	Username   string    `json:"username"`
	FirstName  string    `gorm:"column:first_name" json:"firstName"`
	LastName   string    `gorm:"column:last_name" json:"lastName"`
	Password   string    `json:"password"`
	Avatar     string    `json:"avatar"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	IsEnabled  int       `gorm:"column:is_enabled" json:"isEnabled"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
}

// TokenInfo token information stored in Redis
type TokenInfo struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpireAt     int64  `json:"expireAt"`
	CreateAt     int64  `json:"createAt"`
}

// ResetPasswordReq reset password request (for forgot password scenario)
type ResetPasswordReq struct {
	NewPassword string `json:"newPassword"` // new password (base64 encoded)
}
