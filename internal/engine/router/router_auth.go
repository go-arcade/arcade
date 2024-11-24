package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/engine/constant"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/pkg/http"
	http2 "net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/9 16:11
 * @file: router_auth.go
 * @description: router auth
 */

func (rt *Router) oauth(r *gin.Context) {

	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := r.Param("provider")
	if providerName == "" {
		http.WithRepErrMsg(r, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, r.Request.URL.Path)
		return
	}

	url, err := authService.Oauth(providerName)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	r.Redirect(http2.StatusTemporaryRedirect, url)
}

func (rt *Router) callback(r *gin.Context) {

	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := r.Param("provider")
	state := r.Query("state")
	code := r.Query("code")
	if state == "" || code == "" || providerName == "" {
		http.WithRepErrMsg(r, http.InvalidStatusParameter.Code, http.InvalidStatusParameter.Msg, r.Request.URL.Path)
		return
	}

	userInfo, err := authService.Callback(providerName, state, code)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, err.Error(), r.Request.URL.Path)
		return
	}

	r.Set(constant.DETAIL, userInfo)
}

func (rt *Router) getOauthProvider(r *gin.Context) {

	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	name := r.Param("provider")
	if name == "" {
		http.WithRepErrMsg(r, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, r.Request.URL.Path)
		return
	}

	authConfig, err := authService.GetOauthProvider(name)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	r.Set(constant.DETAIL, authConfig)
}

func (rt *Router) getOauthProviderList(r *gin.Context) {

	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	authConfigs, err := authService.GetOauthProviderList()
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	r.Set(constant.DETAIL, authConfigs)
}
