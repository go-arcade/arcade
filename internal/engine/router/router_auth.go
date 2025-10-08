package router

import (
	http2 "net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/constant"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/pkg/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/9 16:11
 * @file: router_auth.go
 * @description: router auth
 */

func (rt *Router) authRouter(r fiber.Router, auth fiber.Handler) {
	authGroup := r.Group("/auth")
	{
		authGroup.Get("/oauth/:provider", rt.oauth)
		authGroup.Get("/callback/:provider", rt.callback)
		authGroup.Post("/revise", auth, rt.updateUser)
		authGroup.Get("/getProvider/:provider", auth, rt.getOauthProvider)
		authGroup.Get("/getProviderList", auth, rt.getOauthProviderList)
	}
}

func (rt *Router) oauth(c *fiber.Ctx) error {
	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := c.Params("provider")
	if providerName == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	url, err := authService.Oauth(providerName)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	return c.Redirect(url, http2.StatusTemporaryRedirect)
}

func (rt *Router) callback(c *fiber.Ctx) error {
	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := c.Params("provider")
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" || providerName == "" {
		return http.WithRepErrMsg(c, http.InvalidStatusParameter.Code, http.InvalidStatusParameter.Msg, c.Path())
	}

	userInfo, err := authService.Callback(providerName, state, code)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(constant.DETAIL, userInfo)
	return nil
}

func (rt *Router) getOauthProvider(c *fiber.Ctx) error {
	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	name := c.Params("provider")
	if name == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	authConfig, err := authService.GetOauthProvider(name)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(constant.DETAIL, authConfig)
	return nil
}

func (rt *Router) getOauthProviderList(c *fiber.Ctx) error {
	authRepo := repo.NewAuthRepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	authConfigs, err := authService.GetOauthProviderList()
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(constant.DETAIL, authConfigs)
	return nil
}
