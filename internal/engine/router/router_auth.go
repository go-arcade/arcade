package router

import (
	http2 "net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/middleware"
)

func (rt *Router) authRouter(r fiber.Router, auth fiber.Handler) {
	authGroup := r.Group("/auth")
	{

		// 获取提供者列表 公共
		authGroup.Get("/getProvider/:type", auth, rt.getProvider)
		authGroup.Get("/getProviderList", auth, rt.getProviderList)
		authGroup.Get("/getProvider/:name", auth, rt.getProvider)
		authGroup.Get("/getProviderTypeList", auth, rt.getProviderTypeList)
		authGroup.Get("/callback/:provider", rt.callback)
		// authGroup.Post("/revise", auth, rt.updateUser)

		authGroup.Get("/redirect/:provider", rt.redirect)

		authGroup.Post("/ldap/login/:provider", rt.ldapLogin)
	}
}

func (rt *Router) redirect(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := c.Params("provider")
	if providerName == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	url, err := authService.Redirect(providerName)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	return c.Redirect(url, http2.StatusTemporaryRedirect)
}

// callback 统一的 OAuth/OIDC 回调处理
func (rt *Router) callback(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
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

	c.Locals(middleware.DETAIL, userInfo)
	return nil
}

// getProvider 获取提供者列表
func (rt *Router) getProvider(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerType := c.Params("type")
	if providerType == "" {
		return http.WithRepErrMsg(c, http.ProviderTypeIsRequired.Code, http.ProviderTypeIsRequired.Msg, c.Path())
	}

	ssoProviders, err := authService.GetProviderByType(providerType)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, ssoProviders)
	return nil
}

// getProviderList 获取提供者列表
func (rt *Router) getProviderList(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	ssoProviders, err := authService.GetProviderList()
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.DETAIL, ssoProviders)
	return nil
}

// getProviderTypeList 获取提供者类型列表
func (rt *Router) getProviderTypeList(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerTypes, err := authService.GetProviderTypeList()
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.DETAIL, providerTypes)
	return nil
}

func (rt *Router) oidcLogin(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := c.Params("provider")
	if providerName == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	url, err := authService.OIDCLogin(providerName)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	return c.Redirect(url, http2.StatusTemporaryRedirect)
}

// LDAP 路由处理函数

func (rt *Router) ldapLogin(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := c.Params("provider")
	if providerName == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	var req service.LDAPLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, http.BadRequest.Msg, c.Path())
	}

	if req.Username == "" || req.Password == "" {
		return http.WithRepErrMsg(c, http.UsernameArePasswordIsRequired.Code, http.UsernameArePasswordIsRequired.Msg, c.Path())
	}

	userInfo, err := authService.LDAPLogin(providerName, req.Username, req.Password)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, userInfo)
	return nil
}
