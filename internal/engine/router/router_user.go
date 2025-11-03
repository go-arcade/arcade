package router

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/internal/engine/tool"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func (rt *Router) userRouter(r fiber.Router, auth fiber.Handler) {
	userGroup := r.Group("/users")
	{
		// 认证相关路由（无需认证）
		userGroup.Post("/login", rt.login)       // POST /users/login - 登录
		userGroup.Post("/register", rt.register) // POST /users/register - 注册

		// 会话相关路由（需要认证）
		userGroup.Post("/logout", auth, rt.logout)   // POST /users/logout - 登出
		userGroup.Post("/refresh", auth, rt.refresh) // POST /users/refresh - 刷新token

		// 用户资源路由（需要认证）
		userGroup.Get("/me", auth, rt.fetchUserInfo) // GET /users/me - 获取当前用户信息
		userGroup.Post("/invite", auth, rt.addUser)  // POST /users/invite - 邀请用户
		userGroup.Put("/:id", auth, rt.updateUser)   // PUT /users/:id - 更新用户信息
		// userGroup.Get("/", auth, rt.getUserList)            // GET /users - 获取用户列表（待实现）
		// userGroup.Delete("/:id", auth, rt.deleteUser)       // DELETE /users/:id - 删除用户（待实现）
	}
}

func (rt *Router) login(c *fiber.Ctx) error {
	var login *model.Login
	userRepo := repo.NewUserRepo(rt.Ctx)
	userService := service.NewUserService(rt.Ctx, userRepo)

	if err := c.BodyParser(&login); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	user, err := userService.Login(login, rt.Http.Auth)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	result := make(map[string]interface{})
	result["token"] = user.Token
	result["role"] = nil

	c.Locals(middleware.DETAIL, user)
	return nil
}

func (rt *Router) register(c *fiber.Ctx) error {
	var register *model.Register
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := service.NewUserService(rt.Ctx, userRepo)
	if err := c.BodyParser(&register); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	if err := userLogic.Register(register); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "")
	return nil
}

func (rt *Router) refresh(c *fiber.Ctx) error {
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := service.NewUserService(rt.Ctx, userRepo)
	userId := c.Query("userId")
	refreshToken := c.Query("refreshToken")

	token, err := userLogic.Refresh(userId, refreshToken, &rt.Http.Auth)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, token)
	return nil
}

func (rt *Router) logout(c *fiber.Ctx) error {
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := service.NewUserService(rt.Ctx, userRepo)

	claims, err := tool.ParseAuthorizationToken(c, rt.Http.Auth.SecretKey)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	if err := userLogic.Logout(claims.UserId); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "")
	return nil
}

func (rt *Router) addUser(c *fiber.Ctx) error {
	var addUserReq *model.AddUserReq
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := service.NewUserService(rt.Ctx, userRepo)
	if err := c.BodyParser(&addUserReq); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	if err := userLogic.AddUser(*addUserReq); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.OPERATION, "")
	return nil
}

func (rt *Router) updateUser(c *fiber.Ctx) error {
	var user *model.User
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := service.NewUserService(rt.Ctx, userRepo)
	if err := c.BodyParser(&user); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	userId := c.Params("id")
	if err := userLogic.UpdateUser(userId, user); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.OPERATION, "")
	return nil
}

func (rt *Router) fetchUserInfo(c *fiber.Ctx) error {
	var user *model.UserInfo
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := service.NewUserService(rt.Ctx, userRepo)

	claims, err := tool.ParseAuthorizationToken(c, rt.Http.Auth.SecretKey)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	user, err = userLogic.FetchUserInfo(claims.UserId)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, user)
	return nil
}

//func (rt *Router) getUserList(r *gin.Context) {
//
//	userRepo := repo.NewUserRepo(rt.Ctx)
//	userLogic := service.NewUserService(userRepo)
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
//	r.Set(constant.DETAIL, result)
//}
