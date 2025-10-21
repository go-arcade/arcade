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
		// Provider 资源管理（需要认证）
		authGroup.Get("/providers", auth, rt.listProviders)           // 获取所有 providers（支持 ?type=xxx 过滤）
		authGroup.Get("/providers/types", auth, rt.listProviderTypes) // 获取所有 provider 类型
		authGroup.Get("/providers/:name", auth, rt.getProvider)       // 获取指定 provider

		// 认证流程（无需认证）
		authGroup.Get("/authorize/:provider", rt.authorize)   // 发起授权（OAuth/OIDC）
		authGroup.Get("/callback/:provider", rt.callback)     // 授权回调
		authGroup.Post("/ldap/login/:provider", rt.ldapLogin) // 登录（LDAP 或其他）
	}
}

// authorize 发起授权（OAuth/OIDC）
func (rt *Router) authorize(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerName := c.Params("provider")
	if providerName == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	url, err := authService.Authorize(providerName)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	return c.Redirect(url, http2.StatusTemporaryRedirect)
}

// callback OAuth/OIDC 授权回调
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

// listProviders 获取所有 providers（支持 ?type=xxx 过滤）
func (rt *Router) listProviders(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	// 支持通过 query 参数过滤类型
	providerType := c.Query("type")

	var err error
	var ssoProviders any

	if providerType != "" {
		ssoProviders, err = authService.GetProviderByType(providerType)
	} else {
		ssoProviders, err = authService.GetProviderList()
	}

	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, ssoProviders)
	return nil
}

// getProvider 获取指定的 provider
func (rt *Router) getProvider(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	name := c.Params("name")
	if name == "" {
		return http.WithRepErrMsg(c, http.ProviderIsRequired.Code, http.ProviderIsRequired.Msg, c.Path())
	}

	provider, err := authService.GetProvider(name)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, provider)
	return nil
}

// listProviderTypes 获取所有 provider 类型
func (rt *Router) listProviderTypes(c *fiber.Ctx) error {
	authRepo := repo.NewSSORepo(rt.Ctx)
	userRepo := repo.NewUserRepo(rt.Ctx)
	authService := service.NewAuthService(authRepo, userRepo)

	providerTypes, err := authService.GetProviderTypeList()
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, providerTypes)
	return nil
}

// login 统一登录入口（LDAP 等需要用户名密码的认证方式）
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
