package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-arcade/arcade/internal/engine/consts"
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/go-arcade/arcade/pkg/log"
	"gorm.io/gorm"
)

type IUserRepository interface {
	AddUser(addUserReq *model.AddUserReq) error
	UpdateUser(userId string, u *model.User) error
	FetchUserInfo(userId string) (*model.UserInfo, error)
	GetUserByUsername(username string) (string, error)
	Login(login *model.Login) (*model.User, error)
	Register(register *model.Register) error
	Logout(userKey string) error
	GetUserList(offset int, pageSize int) ([]UserWithExtension, int64, error)
	SetToken(userId, aToken string, auth http.Auth) (string, error)
	SetLoginRespInfo(accessExpire time.Duration, loginResp *model.LoginResp) error
	GetToken(key string) (string, error)
	DelToken(key string) error
	GetUserPassword(userId string) (string, error)
	ResetPassword(userId, newPasswordHash string) error
	UpdateAvatar(userId, avatarUrl string) error
	GetUserAvatar(userId string) (string, error)
}

type UserRepo struct {
	db        database.IDatabase
	cache     cache.ICache
	userModel *model.User
}

func NewUserRepo(db database.IDatabase, cache cache.ICache) IUserRepository {
	return &UserRepo{
		db:        db,
		cache:     cache,
		userModel: &model.User{},
	}
}

func (ur *UserRepo) AddUser(addUserReq *model.AddUserReq) error {
	return ur.db.Database().Create(addUserReq).Error
}

// UpdateUser updates user information (user_id, username, password, created_at cannot be updated)
func (ur *UserRepo) UpdateUser(userId string, u *model.User) error {
	return ur.db.Database().Table(ur.userModel.TableName()).
		Where("user_id = ?", userId).
		Omit("user_id", "username", "password", "created_at").
		Updates(u).Error
}

func (ur *UserRepo) FetchUserInfo(userId string) (*model.UserInfo, error) {
	ctx := context.Background()
	key := consts.UserInfoKey + userId
	u := &model.UserInfo{UserId: userId}

	// 从 Redis 获取用户信息
	if ur.cache != nil {
		userInfoStr, err := ur.cache.Get(ctx, key).Result()
		if err == nil && userInfoStr != "" {
			if err := sonic.UnmarshalString(userInfoStr, u); err != nil {
				log.Errorw("failed to unmarshal user info from Redis", "userId", userId, "error", err)
			} else {
				return u, nil
			}
		}
	}

	// fetch from database
	err := ur.db.Database().Table(ur.userModel.TableName()).
		Select("user_id, username, first_name, last_name, avatar, email, phone").
		Where("user_id = ?", userId).First(u).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// cache to Redis
	if ur.cache != nil {
		userInfoJson, err := sonic.MarshalString(u)
		if err != nil {
			log.Errorw("failed to marshal user info", "userId", userId, "error", err)
		} else {
			if err := ur.cache.Set(ctx, key, userInfoJson, time.Hour).Err(); err != nil {
				log.Errorw("failed to cache user info", "userId", userId, "error", err)
			}
		}
	}

	return u, nil
}

func (ur *UserRepo) GetUserByUsername(username string) (string, error) {
	var u = &model.User{}
	err := ur.db.Database().Table(ur.userModel.TableName()).Select("user_id").Where("username = ?", username).
		First(u).Error
	return u.UserId, err
}

func (ur *UserRepo) Login(login *model.Login) (*model.User, error) {
	var u = &model.User{}
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Table(ur.userModel.TableName()).Select("user_id, username, first_name, last_name, avatar, email, phone, password")
	}

	err := ur.db.Database().Scopes(scope).Where(
		"(username = ? OR email = ?)",
		login.Username, login.Email,
	).First(&u).Error

	if err != nil {
		return nil, errors.New(http.UserNotExist.Msg)
	}
	return u, err
}

func (ur *UserRepo) Register(register *model.Register) error {
	var u model.User
	err := ur.db.Database().Table(ur.userModel.TableName()).Select("username").
		Where("username = ?", register.Username).
		First(&u).Error
	if err == nil {
		return errors.New(http.UserAlreadyExist.Msg)
	}
	return ur.db.Database().Table(ur.userModel.TableName()).Create(register).Error
}

func (ur *UserRepo) Logout(userKey string) error {
	if ur.cache == nil {
		return nil
	}
	ctx := context.Background()
	return ur.cache.Del(ctx, userKey).Err()
}

// UserWithExtension combines user and extension information
type UserWithExtension struct {
	model.User
	LastLoginAt      *time.Time `gorm:"column:last_login_at" json:"lastLoginAt"`
	InvitationStatus string     `gorm:"column:invitation_status" json:"invitationStatus"`
}

func (ur *UserRepo) GetUserList(offset int, pageSize int) ([]UserWithExtension, int64, error) {
	var users []UserWithExtension
	var count int64

	// join with user extension table to get last login time and invitation status
	err := ur.db.Database().Table(ur.userModel.TableName() + " AS u").
		Select(`u.user_id, u.username, u.first_name, u.last_name, u.avatar, u.email, u.phone,
			u.is_enabled, u.is_superadmin,
			ue.last_login_at,
			COALESCE(ue.invitation_status, 'accepted') AS invitation_status`).
		Joins("LEFT JOIN t_user_ext AS ue ON u.user_id = ue.user_id").
		Offset(offset).
		Limit(pageSize).
		Order("u.created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	err = ur.db.Database().Model(&model.User{}).Count(&count).Error
	return users, count, err
}

func (ur *UserRepo) SetToken(userId, aToken string, auth http.Auth) (string, error) {
	if ur.cache == nil {
		return "", fmt.Errorf("cache not available")
	}
	ctx := context.Background()

	// 构建 TokenInfo 结构
	now := time.Now()
	tokenInfo := middleware.TokenInfo{
		AccessToken:  aToken,
		RefreshToken: "", // refresh token 在这个方法中不需要更新
		ExpireAt:     now.Add(auth.AccessExpire * time.Second).Unix(),
		CreateAt:     now.Unix(),
	}

	// 序列化 token 信息为 JSON
	tokenInfoJson, err := sonic.MarshalString(&tokenInfo)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token info: %w", err)
	}

	key := consts.UserTokenKey + userId
	if err := ur.cache.Set(ctx, key, tokenInfoJson, auth.AccessExpire*time.Second).Err(); err != nil {
		return "", fmt.Errorf("failed to set token in Redis: %w", err)
	}
	return key, nil
}

func (ur *UserRepo) SetLoginRespInfo(accessExpire time.Duration, loginResp *model.LoginResp) error {
	if ur.cache == nil {
		return fmt.Errorf("cache not available")
	}
	ctx := context.Background()

	pipe := ur.cache.Pipeline()

	tokenInfo := middleware.TokenInfo{
		AccessToken:  loginResp.Token["accessToken"],
		RefreshToken: loginResp.Token["refreshToken"],
		ExpireAt:     loginResp.ExpireAt,
		CreateAt:     loginResp.CreateAt,
	}

	tokenInfoJson, err := sonic.Marshal(&tokenInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %w", err)
	}

	tokenKey := consts.UserTokenKey + loginResp.UserInfo.UserId
	if err := pipe.Set(ctx, tokenKey, tokenInfoJson, accessExpire*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set token in Redis: %w", err)
	}

	userInfoJson, err := sonic.Marshal(&loginResp.UserInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %w", err)
	}

	userInfoKey := consts.UserInfoKey + loginResp.UserInfo.UserId
	if err := pipe.Set(ctx, userInfoKey, userInfoJson, accessExpire*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set user info in Redis: %w", err)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}
	return nil
}

func (ur *UserRepo) GetToken(key string) (string, error) {
	if ur.cache == nil {
		return "", fmt.Errorf("cache not available")
	}
	ctx := context.Background()
	token, err := ur.cache.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get token from Redis: %w", err)
	}
	return token, nil
}

func (ur *UserRepo) DelToken(key string) error {
	if ur.cache == nil {
		return nil
	}
	ctx := context.Background()
	if err := ur.cache.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete token from Redis: %w", err)
	}
	return nil
}

// GetUserPassword gets user password hash by user ID
func (ur *UserRepo) GetUserPassword(userId string) (string, error) {
	var u model.User
	err := ur.db.Database().Table(ur.userModel.TableName()).
		Select("password").
		Where("user_id = ?", userId).
		First(&u).Error
	if err != nil {
		return "", err
	}
	return u.Password, nil
}

// ResetPassword resets user password
func (ur *UserRepo) ResetPassword(userId, newPasswordHash string) error {
	return ur.db.Database().Table(ur.userModel.TableName()).
		Where("user_id = ?", userId).
		Update("password", newPasswordHash).Error
}

// UpdateAvatar updates user avatar URL
func (ur *UserRepo) UpdateAvatar(userId, avatarUrl string) error {
	result := ur.db.Database().Table(ur.userModel.TableName()).
		Where("user_id = ?", userId).
		Update("avatar", avatarUrl)

	if result.Error != nil {
		return result.Error
	}

	// log if no rows were affected (user not found)
	if result.RowsAffected == 0 {
		log.Warnw("no rows updated for user avatar", "userId", userId)
	}

	return nil
}

// GetUserAvatar gets user avatar URL by user ID
func (ur *UserRepo) GetUserAvatar(userId string) (string, error) {
	var u model.User
	err := ur.db.Database().Table(ur.userModel.TableName()).
		Select("avatar").
		Where("user_id = ?", userId).
		First(&u).Error
	if err != nil {
		return "", err
	}
	return u.Avatar, nil
}
