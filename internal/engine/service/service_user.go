package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-arcade/arcade/internal/engine/consts"
	usermodel "github.com/go-arcade/arcade/internal/engine/model"
	userrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/tool"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/jwt"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

type LoginService interface {
	Login(login *usermodel.Login, auth http.Auth) (*usermodel.LoginResp, error)
}

type UserService struct {
	ctx      *ctx.Context
	cache    cache.ICache
	userRepo userrepo.IUserRepository
}

func NewUserService(ctx *ctx.Context, cache cache.ICache, userRepo userrepo.IUserRepository) *UserService {
	return &UserService{
		ctx:      ctx,
		cache:    cache,
		userRepo: userRepo,
	}
}

func (ul *UserService) Login(login *usermodel.Login, auth http.Auth) (*usermodel.LoginResp, error) {
	pwd, err := tool.DecodeBase64(login.Password)
	if err != nil {
		log.Error("failed to decode password: %v", err)
		return nil, errors.New(http.UserIncorrectPassword.Msg)
	}

	userInfo, err := ul.userRepo.Login(login)
	if err != nil {
		log.Error("login failed for userInfo: %v", err)
		return nil, err
	}
	if userInfo == nil || userInfo.Username == "" || userInfo.Username != login.Username {
		log.Error("userInfo not found")
		return nil, errors.New(http.UserNotExist.Msg)
	}

	// compare stored password hash with provided password
	if !comparePassword(userInfo.Password, string(pwd)) {
		log.Error("incorrect password provided")
		return nil, errors.New(http.UserIncorrectPassword.Msg)
	}

	aToken, rToken, err := jwt.GenToken(userInfo.UserId, []byte(auth.SecretKey), auth.AccessExpire, auth.RefreshExpire)
	if err != nil {
		log.Error("failed to generate tokens: %v", err)
		return nil, err
	}

	now := time.Now()
	createAt := now.Unix()
	expireAt := now.Add(auth.AccessExpire * time.Minute).Unix()

	resp := &usermodel.LoginResp{
		UserInfo: usermodel.UserInfo{
			UserId:    userInfo.UserId,
			Username:  userInfo.Username,
			FirstName: userInfo.FirstName,
			LastName:  userInfo.LastName,
			Avatar:    userInfo.Avatar,
			Email:     userInfo.Email,
			Phone:     userInfo.Phone,
		},
		Token: map[string]string{
			"accessToken":  aToken,
			"refreshToken": rToken,
			"expireAt":     fmt.Sprintf("%d", expireAt),
			"createAt":     fmt.Sprintf("%d", createAt),
		},
		Role:     map[string]string{},
		ExpireAt: expireAt,
		CreateAt: createAt,
	}

	go func() {
		if err = ul.userRepo.SetLoginRespInfo(auth.AccessExpire, resp); err != nil {
			log.Error("failed to set login response info: %v", err)
			return
		}

		// update last login time in userInfo extension
		// Note: This should be injected, but for now we'll skip it to avoid circular dependency
		// TODO: Inject UserExtensionService to avoid direct repo creation
	}()

	return resp, nil
}

func (ul *UserService) Refresh(userId, rToken string, auth *http.Auth) (map[string]string, error) {
	token, err := jwt.RefreshToken(auth, userId, rToken)
	if err != nil {
		log.Error("failed to refresh token: %v", err)
		return nil, err
	}

	// calculate token expiration time
	expireAt := time.Now().Add(auth.AccessExpire * time.Minute).Unix()
	token["expireAt"] = fmt.Sprintf("%d", expireAt)

	k, err := ul.userRepo.SetToken(userId, token["accessToken"], *auth)
	log.Debug("token key: %v", k)
	if err != nil {
		log.Error("failed to set token in Redis: %v", err)
		return token, err
	}

	return token, nil
}

func (ul *UserService) Register(register *usermodel.Register) error {

	var err error
	register.UserId = id.GetUUIDWithoutDashes()
	// set default values if not provided
	if register.FirstName == "" {
		register.FirstName = register.Username
	}
	register.CreateTime = time.Now()
	password, err := getPassword(register.Password)
	if err != nil {
		return err
	}
	register.Password = string(password)
	if err = ul.userRepo.Register(register); err != nil {
		return err
	}

	return err
}

func (ul *UserService) Logout(userId string) error {
	// delete token
	tokenKey := consts.UserTokenKey + userId
	if err := ul.userRepo.DelToken(tokenKey); err != nil {
		log.Error("failed to delete token: %v", err)
		return errors.New(http.TokenBeEmpty.Msg)
	}

	// delete user info cache
	userInfoKey := consts.UserInfoKey + userId
	if err := ul.userRepo.DelToken(userInfoKey); err != nil {
		log.Error("failed to delete user info: %v", err)
		// user info deletion failure does not affect logout
	}

	return nil
}

func (ul *UserService) AddUser(addUserReq usermodel.AddUserReq) error {

	var err error
	addUserReq.UserId = id.GetUUIDWithoutDashes()
	addUserReq.IsEnabled = 1
	addUserReq.CreateTime = time.Now()
	if err = ul.userRepo.AddUser(&addUserReq); err != nil {
		return err
	}
	return err
}

func (ul *UserService) UpdateUser(userId string, userEntity *usermodel.User) error {
	var err error
	if err = ul.userRepo.UpdateUser(userId, userEntity); err != nil {
		return err
	}

	// clear user info cache after update
	userInfoKey := consts.UserInfoKey + userId
	if err := ul.userRepo.DelToken(userInfoKey); err != nil {
		log.Warn("failed to clear user info cache: %v", err)
	}

	return err
}

func (ul *UserService) FetchUserInfo(userId string) (*usermodel.UserInfo, error) {

	userInfo, err := ul.userRepo.FetchUserInfo(userId)
	if err != nil {
		return nil, err
	}

	return userInfo, err
}

func (ul *UserService) GetUserList(pageNum, pageSize int) ([]userrepo.UserWithExtension, int64, error) {
	// set default values
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNum - 1) * pageSize
	users, count, err := ul.userRepo.GetUserList(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return users, count, err
}

func comparePassword(oldPassword, newPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(newPassword))
	if err != nil {
		return false
	} else {
		return true
	}
}

// ResetPassword resets user password (for forgot password scenario, no old password required)
func (ul *UserService) ResetPassword(userId string, req *usermodel.ResetPasswordReq) error {
	// decode new password from base64
	newPwd, err := tool.DecodeBase64(req.NewPassword)
	if err != nil {
		log.Error("failed to decode new password: %v", err)
		return errors.New("invalid new password format")
	}

	// validate new password length
	if len(newPwd) < 6 {
		return errors.New("new password must be at least 6 characters")
	}

	// hash new password
	newPasswordHash, err := getPassword(string(newPwd))
	if err != nil {
		log.Error("failed to hash new password: %v", err)
		return errors.New("failed to process new password")
	}

	// update password
	if err := ul.userRepo.ResetPassword(userId, string(newPasswordHash)); err != nil {
		log.Error("failed to reset password: %v", err)
		return errors.New("failed to reset password")
	}

	// invalidate all tokens for security
	tokenKey := consts.UserTokenKey + userId
	if err := ul.userRepo.DelToken(tokenKey); err != nil {
		log.Warn("failed to delete token after password reset: %v", err)
		// this is not critical, continue
	}

	log.Info("user password reset successfully: %s", userId)
	return nil
}

// UpdateAvatar updates user avatar URL and clears cache
func (ul *UserService) UpdateAvatar(userId, avatarUrl string) error {
	// update avatar in database
	if err := ul.userRepo.UpdateAvatar(userId, avatarUrl); err != nil {
		log.Error("failed to update user avatar: %v", err)
		return errors.New("failed to update user avatar")
	}

	// clear user info cache in Redis
	if ul.cache != nil {
		key := consts.UserInfoKey + userId
		if err := ul.cache.Del(context.Background(), key).Err(); err != nil {
			log.Warn("failed to clear user info cache: %v", err)
			// not critical, continue
		}
	}

	log.Info("user avatar updated successfully: %s, url: %s", userId, avatarUrl)
	return nil
}
