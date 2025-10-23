package service

import (
	"errors"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/tool"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/jwt"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

type LoginService interface {
	Login(login *model.Login, auth http.Auth) (*model.LoginResp, error)
}

type UserService struct {
	ctx      *ctx.Context
	userRepo *repo.UserRepo
}

func NewUserService(ctx *ctx.Context, userRepo *repo.UserRepo) *UserService {
	return &UserService{
		ctx:      ctx,
		userRepo: userRepo,
	}
}

func (ul *UserService) Login(login *model.Login, auth http.Auth) (*model.LoginResp, error) {
	pwd, err := tool.DecodeBase64(login.Password)
	if err != nil {
		log.Errorf("failed to decode password: %v", err)
		return nil, errors.New(http.UserIncorrectPassword.Msg)
	}

	// 用户名和密码登录
	user, err := ul.userRepo.Login(login)
	if err != nil {
		log.Errorf("login failed for user: %v", err)
		return nil, err
	}
	if user == nil || user.Username == "" || user.Username != login.Username {
		log.Error("user not found")
		return nil, errors.New(http.UserNotExist.Msg)
	}

	// 比较存储的密码哈希与提供的密码
	if !comparePassword(user.Password, string(pwd)) {
		log.Error("incorrect password provided")
		return nil, errors.New(http.UserIncorrectPassword.Msg)
	}

	aToken, rToken, err := jwt.GenToken(user.UserId, []byte(auth.SecretKey), auth.AccessExpire, auth.RefreshExpire)
	if err != nil {
		log.Errorf("failed to generate tokens: %v", err)
		return nil, err
	}

	resp := &model.LoginResp{
		UserInfo: model.UserInfo{
			UserId:   user.UserId,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Email:    user.Email,
			Phone:    user.Phone,
		},
		Token: map[string]string{
			"accessToken":  aToken,
			"refreshToken": rToken,
		},
		Role: map[string]string{},
	}

	go func() {
		if err = ul.userRepo.SetLoginRespInfo(auth.AccessExpire, resp); err != nil {
			log.Errorf("failed to set login response info: %v", err)
			return
		}
	}()

	return resp, nil
}

func (ul *UserService) Refresh(userId, rToken string, auth *http.Auth) (map[string]string, error) {
	token, err := jwt.RefreshToken(auth, userId, rToken)
	if err != nil {
		log.Errorf("failed to refresh token: %v", err)
		return nil, err
	}

	k, err := ul.userRepo.SetToken(userId, token["accessToken"], *auth)
	log.Debugf("token key: %v", k)
	if err != nil {
		log.Errorf("failed to set token in Redis: %v", err)
		return token, err
	}

	return token, nil
}

func (ul *UserService) Register(register *model.Register) error {

	var err error
	register.UserId = id.GetUUIDWithoutDashes()
	register.Nickname = register.Username
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

func (ul *UserService) Logout(keyPrefix, userId string) error {
	var key = keyPrefix + userId

	result, err := ul.userRepo.GetToken(key)
	if err != nil {
		log.Errorf("failed to get token from Redis: %v", err)
		return errors.New(http.TokenBeEmpty.Msg)
	}
	if result == "" {
		log.Error("token not found")
		return errors.New(http.TokenBeEmpty.Msg)
	}

	if err = ul.userRepo.DelToken(key); err != nil {
		log.Errorf("failed to logout: %v", err)
		return errors.New(http.TokenBeEmpty.Msg)
	}
	return err
}

func (ul *UserService) AddUser(addUserReq model.AddUserReq) error {

	var err error
	addUserReq.UserId = id.GetUUIDWithoutDashes()
	addUserReq.IsEnabled = 1
	addUserReq.CreateTime = time.Now()
	if err = ul.userRepo.AddUser(&addUserReq); err != nil {
		return err
	}
	return err
}

func (ul *UserService) UpdateUser(userId string, user *model.User) error {

	var err error
	if err = ul.userRepo.UpdateUser(userId, user); err != nil {
		return err
	}
	return err
}

func (ul *UserService) GetUserInfo(userId string) (*model.UserInfo, error) {

	user, err := ul.userRepo.GetUserInfo(userId)
	if err != nil {
		return nil, err
	}
	return user, err
}

//func (ul *UserService) GetUserList(pageNum, pageSize int) ([]model.User, int64, error) {
//
//	offset := (pageNum - 1) * pageSize
//	users, count, err := ul.userRepo.GetUserList(offset, pageSize)
//	if err != nil {
//		return nil, 0, err
//	}
//	return users, count, err
//}

func getPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hash, err
}

func comparePassword(oldPassword, newPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(newPassword))
	if err != nil {
		return false
	} else {
		return true
	}
}
