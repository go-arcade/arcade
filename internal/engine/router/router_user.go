package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/engine/common"
	"github.com/go-arcade/arcade/internal/engine/consts"
	"github.com/go-arcade/arcade/internal/engine/logic"
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/4 10:47
 * @file: router_user.go
 * @description: user router
 */

func (rt *Router) login(r *gin.Context) {

	var login *model.Login
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := logic.NewUserLogic(rt.Ctx, userRepo)

	if err := r.BindJSON(&login); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	user, err := userLogic.Login(login, rt.Http.Auth)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	result := make(map[string]interface{})
	result["token"] = user.Token
	result["role"] = nil

	r.Set(consts.DETAIL, user)
}

func (rt *Router) register(r *gin.Context) {

	//todo: 实现注册开关

	var register *model.Register
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := logic.NewUserLogic(rt.Ctx, userRepo)
	if err := r.BindJSON(&register); err != nil {
		//todo: 统一拦截
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	if err := userLogic.Register(register); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	r.Set(consts.OPERATION, "")
}

func (rt *Router) refresh(r *gin.Context) {

	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := logic.NewUserLogic(rt.Ctx, userRepo)
	userId := r.Query("userId")
	refreshToken := r.Query("refreshToken")

	token, err := userLogic.Refresh(userId, refreshToken, &rt.Http.Auth)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	r.Set(consts.DETAIL, token)
}

func (rt *Router) logout(r *gin.Context) {
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := logic.NewUserLogic(rt.Ctx, userRepo)
	userId := r.PostForm("userId")
	if err := userLogic.Logout(rt.Http.Auth.RedisKeyPrefix, userId); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	r.Set(consts.OPERATION, "")
}

func (rt *Router) oauth(r *gin.Context) {
	r.Set(consts.OPERATION, "")
}

func (rt *Router) addUser(r *gin.Context) {

	var addUserReq *model.AddUserReq
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := logic.NewUserLogic(rt.Ctx, userRepo)
	if err := r.BindJSON(&addUserReq); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	if err := userLogic.AddUser(*addUserReq); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	r.Set(consts.OPERATION, "")
}

func (rt *Router) updateUser(r *gin.Context) {

	var user *model.User
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := logic.NewUserLogic(rt.Ctx, userRepo)
	if err := r.BindJSON(&user); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	userId := r.Param("userId")
	if err := userLogic.UpdateUser(userId, user); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	r.Set(consts.OPERATION, "")
}

func (rt *Router) getUserInfo(r *gin.Context) {

	var user *model.UserInfo
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := logic.NewUserLogic(rt.Ctx, userRepo)

	claims, err := common.ParseAuthorizationToken(r, rt.Http.Auth.SecretKey)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
	}

	user, err = userLogic.GetUserInfo(claims.UserId)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	r.Set(consts.DETAIL, user)
}

//func (rt *Router) getUserList(r *gin.Context) {
//
//	userRepo := repo.NewUserRepo(rt.Ctx)
//	userLogic := logic.NewUserLogic(userRepo)
//
//	pageNum := queryInt(r, "pageNum")   // default 1
//	pageSize := queryInt(r, "pageSize") // default 10
//	users, count, err := userLogic.GetUserList(pageNum, pageSize)
//	if err != nil {
//		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
//		return
//	}
//
//	result := make(map[string]interface{})
//	result["users"] = users
//	result["count"] = count
//	r.Set(consts.DETAIL, result)
//}
