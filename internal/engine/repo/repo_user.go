package repo

import (
	"errors"
	"fmt"
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/httpx"
	"github.com/go-arcade/arcade/pkg/server"
	"gorm.io/gorm"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/4 10:30
 * @file: repo_user.go
 * @description: user repository
 */

type UserRepo struct {
	*ctx.Context
	userModel *model.User
}

func NewUserRepo(ctx *ctx.Context) *UserRepo {
	return &UserRepo{
		Context:   ctx,
		userModel: &model.User{},
	}
}

func (ur *UserRepo) AddUser(addUserReq *model.AddUserReq) error {
	return ur.Context.GetDB().Create(addUserReq).Error
}

func (ur *UserRepo) UpdateUser(userId string, user *model.User) error {
	return ur.Context.GetDB().Where("user_id = ?", userId).Model(user).Updates(user).Error
}

func (ur *UserRepo) GetUserById(userId string) (*model.User, error) {
	user := &model.User{}
	err := ur.Context.GetDB().Where("user_id = ?", userId).First(user).Error
	return user, err
}

func (ur *UserRepo) GetUserByUsername(username string) (string, error) {
	var user = &model.User{}
	err := ur.Context.GetDB().Table(ur.userModel.TableName()).Select("user_id").Where("username = ?", username).First(user).Error
	return user.UserId, err
}

func (ur *UserRepo) Login(login *model.Login) (*model.User, error) {
	var user = &model.User{}
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Table(ur.userModel.TableName()).Select("user_id, username, nick_name, avatar, email, phone, password")
	}

	err := ur.Context.GetDB().Scopes(scope).Where(
		"(username = ? OR email = ?)",
		login.Username, login.Email,
	).First(&user).Error

	if err != nil {
		return nil, errors.New(httpx.UserNotExist.Msg)
	}
	return user, err
}

func (ur *UserRepo) Register(register *model.Register) error {
	var user model.User
	err := ur.Context.GetDB().Table(ur.userModel.TableName()).Select("username").Where("username = ?", register.Username).First(&user).Error
	if err == nil {
		return errors.New(httpx.UserAlreadyExist.Msg)
	}
	return ur.Context.GetDB().Table(ur.userModel.TableName()).Create(register).Error
}

func (ur *UserRepo) GetUserList(offset int, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var count int64
	err := ur.Context.GetDB().Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	err = ur.Context.GetDB().Model(&model.User{}).Count(&count).Error
	return users, count, err
}

func (ur *UserRepo) SetToken(userId, aToken string, auth server.Auth) (string, error) {

	key := auth.RedisKeyPrefix + userId
	if err := ur.GetRedis().Set(ur.Ctx, key, aToken, auth.AccessExpire*time.Second).Err(); err != nil {
		return "", fmt.Errorf("failed to set token in Redis: %w", err)
	}
	return key, nil
}
