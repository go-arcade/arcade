package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/consts"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/internal/engine/tool"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/middleware"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/4 10:47
 * @file: router_user.go
 * @description: user router
 */

func (rt *Router) userRouter(r fiber.Router, auth fiber.Handler) {
	userGroup := r.Group("/user")
	{
		userGroup.Post("/login", rt.login)
		userGroup.Post("/register", rt.register)

		userGroup.Post("/logout", auth, rt.logout)
		userGroup.Get("/refresh", auth, rt.refresh)
		userGroup.Post("/invite", auth, rt.addUser)

		userGroup.Get("/getUserInfo", auth, rt.getUserInfo)
		//userGroup.GET("/getUserList", rt.getUserList)
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
	userId := c.FormValue("userId")
	if err := userLogic.Logout(consts.UserInfoKey, userId); err != nil {
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

	userId := c.Params("userId")
	if err := userLogic.UpdateUser(userId, user); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.OPERATION, "")
	return nil
}

func (rt *Router) getUserInfo(c *fiber.Ctx) error {
	var user *model.UserInfo
	userRepo := repo.NewUserRepo(rt.Ctx)
	userLogic := service.NewUserService(rt.Ctx, userRepo)

	claims, err := tool.ParseAuthorizationToken(c, rt.Http.Auth.SecretKey)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	user, err = userLogic.GetUserInfo(claims.UserId)
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
