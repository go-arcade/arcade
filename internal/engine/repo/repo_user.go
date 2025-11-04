package repo

import (
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-arcade/arcade/internal/engine/consts"
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
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

// UpdateUser updates user information (user_id, username, password, created_at cannot be updated)
func (ur *UserRepo) UpdateUser(userId string, user *model.User) error {
	return ur.Context.DBSession().Table(ur.userModel.TableName()).
		Where("user_id = ?", userId).
		Omit("user_id", "username", "password", "created_at").
		Updates(user).Error
}

func (ur *UserRepo) FetchUserInfo(userId string) (*model.UserInfo, error) {

	key := consts.UserInfoKey + userId
	user := &model.UserInfo{UserId: userId}

	// 从 Redis 获取用户信息
	userInfoStr, err := ur.RedisSession().Get(ur.Context.ContextIns(), key).Result()
	if err == nil && userInfoStr != "" {
		if err := sonic.UnmarshalString(userInfoStr, user); err != nil {
			log.Errorf("failed to unmarshal user info from Redis: %v", err)
		} else {
			return user, nil
		}
	}

	// fetch from database
	err = ur.Context.DBSession().Table(ur.userModel.TableName()).
		Select("user_id, username, first_name, last_name, avatar, email, phone").
		Where("user_id = ?", userId).First(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// cache to Redis
	userInfoJson, err := sonic.MarshalString(user)
	if err != nil {
		log.Errorf("failed to marshal user info: %v", err)
	} else {
		if err := ur.RedisSession().Set(ur.Context.ContextIns(), key, userInfoJson, time.Hour).Err(); err != nil {
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
		return db.Table(ur.userModel.TableName()).Select("user_id, username, first_name, last_name, avatar, email, phone, password")
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

	return ur.Context.RedisSession().Del(ur.Context.ContextIns(), userKey).Err()
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
	err := ur.Context.DBSession().Table(ur.userModel.TableName() + " AS u").
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

	err = ur.Context.DBSession().Model(&model.User{}).Count(&count).Error
	return users, count, err
}

func (ur *UserRepo) SetToken(userId, aToken string, auth http.Auth) (string, error) {

	// 构建 TokenInfo 结构
	now := time.Now()
	tokenInfo := model.TokenInfo{
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
	if err := ur.RedisSession().Set(ur.Context.ContextIns(), key, tokenInfoJson, auth.AccessExpire*time.Second).Err(); err != nil {
		return "", fmt.Errorf("failed to set token in Redis: %w", err)
	}
	return key, nil
}

func (ur *UserRepo) SetLoginRespInfo(accessExpire time.Duration, loginResp *model.LoginResp) error {

	pipe := ur.RedisSession().Pipeline()

	tokenInfo := model.TokenInfo{
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
	if err := pipe.
		Set(ur.Context.ContextIns(), tokenKey, tokenInfoJson, accessExpire*time.Minute).
		Err(); err != nil {
		return fmt.Errorf("failed to set token in Redis: %w", err)
	}

	userInfoJson, err := sonic.Marshal(&loginResp.UserInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %w", err)
	}

	userInfoKey := consts.UserInfoKey + loginResp.UserInfo.UserId
	if err := pipe.Set(ur.Context.ContextIns(), userInfoKey, userInfoJson, accessExpire*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set user info in Redis: %w", err)
	}

	if _, err := pipe.Exec(ur.Context.ContextIns()); err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}
	return nil
}

func (ur *UserRepo) GetToken(key string) (string, error) {
	token, err := ur.RedisSession().Get(ur.Context.ContextIns(), key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get token from Redis: %w", err)
	}
	return token, nil
}

func (ur *UserRepo) DelToken(key string) error {
	if err := ur.RedisSession().Del(ur.Context.ContextIns(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete token from Redis: %w", err)
	}
	return nil
}

// GetUserPassword gets user password hash by user ID
func (ur *UserRepo) GetUserPassword(userId string) (string, error) {
	var user model.User
	err := ur.Context.DBSession().Table(ur.userModel.TableName()).
		Select("password").
		Where("user_id = ?", userId).
		First(&user).Error
	if err != nil {
		return "", err
	}
	return user.Password, nil
}

// ResetPassword resets user password
func (ur *UserRepo) ResetPassword(userId, newPasswordHash string) error {
	return ur.Context.DBSession().Table(ur.userModel.TableName()).
		Where("user_id = ?", userId).
		Update("password", newPasswordHash).Error
}

// UpdateAvatar updates user avatar URL
func (ur *UserRepo) UpdateAvatar(userId, avatarUrl string) error {
	result := ur.Context.DBSession().Table(ur.userModel.TableName()).
		Where("user_id = ?", userId).
		Update("avatar", avatarUrl)

	if result.Error != nil {
		return result.Error
	}

	// log if no rows were affected (user not found)
	if result.RowsAffected == 0 {
		log.Warnf("no rows updated for user avatar, userId: %s", userId)
	}

	return nil
}

// GetUserAvatar gets user avatar URL by user ID
func (ur *UserRepo) GetUserAvatar(userId string) (string, error) {
	var user model.User
	err := ur.Context.DBSession().Table(ur.userModel.TableName()).
		Select("avatar").
		Where("user_id = ?", userId).
		First(&user).Error
	if err != nil {
		return "", err
	}
	return user.Avatar, nil
}
