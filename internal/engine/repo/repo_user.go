package repo

import (
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/observabil/arcade/internal/engine/consts"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/log"
	"gorm.io/gorm"
)

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
	return ur.Context.DBSession().Create(addUserReq).Error
}

func (ur *UserRepo) UpdateUser(userId string, user *model.User) error {
	return ur.Context.DBSession().Where("user_id = ?", userId).Model(user).Updates(user).Error
}

func (ur *UserRepo) GetUserInfo(userId string) (*model.UserInfo, error) {

	key := consts.UserInfoKey + userId
	user := &model.UserInfo{UserId: userId}

	// 从 Redis 获取用户信息
	userInfoStr, err := ur.RedisSession().Get(ur.Ctx, key).Result()
	if err == nil && userInfoStr != "" {
		if err := sonic.UnmarshalString(userInfoStr, user); err != nil {
			log.Errorf("failed to unmarshal user info from Redis: %v", err)
		} else {
			return user, nil
		}
	}

	// 从数据库获取
	err = ur.Context.DBSession().Table(ur.userModel.TableName()).
		Select("user_id, username, nick_name AS nickname, avatar, email, phone").
		Where("user_id = ?", userId).First(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 缓存到 Redis
	userInfoJson, err := sonic.MarshalString(user)
	if err != nil {
		log.Errorf("failed to marshal user info: %v", err)
	} else {
		if err := ur.RedisSession().Set(ur.Ctx, key, userInfoJson, time.Hour).Err(); err != nil {
			log.Errorf("failed to cache user info: %v", err)
		}
	}

	return user, nil
}

func (ur *UserRepo) GetUserByUsername(username string) (string, error) {
	var user = &model.User{}
	err := ur.Context.DBSession().Table(ur.userModel.TableName()).Select("user_id").Where("username = ?", username).
		First(user).Error
	return user.UserId, err
}

func (ur *UserRepo) Login(login *model.Login) (*model.User, error) {
	var user = &model.User{}
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Table(ur.userModel.TableName()).Select("user_id, username, nick_name, avatar, email, phone, password")
	}

	err := ur.Context.DBSession().Scopes(scope).Where(
		"(username = ? OR email = ?)",
		login.Username, login.Email,
	).First(&user).Error

	if err != nil {
		return nil, errors.New(http.UserNotExist.Msg)
	}
	return user, err
}

func (ur *UserRepo) Register(register *model.Register) error {
	var user model.User
	err := ur.Context.DBSession().Table(ur.userModel.TableName()).Select("username").
		Where("username = ?", register.Username).
		First(&user).Error
	if err == nil {
		return errors.New(http.UserAlreadyExist.Msg)
	}
	return ur.Context.DBSession().Table(ur.userModel.TableName()).Create(register).Error
}

func (ur *UserRepo) Logout(userKey string) error {

	return ur.Context.RedisSession().Del(ur.Ctx, userKey).Err()
}

func (ur *UserRepo) GetUserList(offset int, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var count int64
	err := ur.Context.DBSession().Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	err = ur.Context.DBSession().Model(&model.User{}).Count(&count).Error
	return users, count, err
}

func (ur *UserRepo) SetToken(userId, aToken string, auth http.Auth) (string, error) {

	key := consts.UserInfoKey + userId
	if err := ur.RedisSession().Set(ur.Ctx, key, aToken, auth.AccessExpire*time.Second).Err(); err != nil {
		return "", fmt.Errorf("failed to set token in Redis: %w", err)
	}
	return key, nil
}

func (ur *UserRepo) SetLoginRespInfo(accessExpire time.Duration, loginResp *model.LoginResp) error {

	pipe := ur.RedisSession().Pipeline()

	// 设置 token
	key := consts.UserInfoKey + loginResp.UserInfo.UserId
	if err := pipe.
		Set(ur.Ctx, key, loginResp.Token["accessToken"], accessExpire*time.Minute).
		Err(); err != nil {
		return fmt.Errorf("failed to set token in Redis: %w", err)
	}

	// 序列化用户信息为 JSON
	userInfoJson, err := sonic.MarshalString(&loginResp.UserInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %w", err)
	}

	// 设置用户信息
	if err := pipe.Set(ur.Ctx, key, userInfoJson, accessExpire*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set user info in Redis: %w", err)
	}

	if _, err := pipe.Exec(ur.Ctx); err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}
	return nil
}

func (ur *UserRepo) GetToken(key string) (string, error) {
	token, err := ur.RedisSession().Get(ur.Ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get token from Redis: %w", err)
	}
	return token, nil
}

func (ur *UserRepo) DelToken(key string) error {
	if err := ur.RedisSession().Del(ur.Ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete token from Redis: %w", err)
	}
	return nil
}
