package router

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func (rt *Router) userExtensionRouter(r fiber.Router, auth fiber.Handler) {
	userExtGroup := r.Group("/users/:userId/extension", auth)
	{
		userExtGroup.Get("/", rt.getUserExtension)                 // GET /users/:userId/extension - get user extension info
		userExtGroup.Put("/", rt.updateUserExtension)              // PUT /users/:userId/extension - update user extension info
		userExtGroup.Put("/timezone", rt.updateTimezone)           // PUT /users/:userId/extension/timezone - update timezone
		userExtGroup.Put("/invitation", rt.updateInvitationStatus) // PUT /users/:userId/extension/invitation - update invitation status
	}
}

// getUserExtension gets user extension information
func (rt *Router) getUserExtension(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "user id is required", c.Path())
	}

	userExtRepo := repo.NewUserExtensionRepo(rt.Ctx)
	userExtService := service.NewUserExtensionService(userExtRepo)

	extension, err := userExtService.GetUserExtension(userId)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.DETAIL, extension)
	return nil
}

// updateUserExtension updates user extension information
func (rt *Router) updateUserExtension(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "user id is required", c.Path())
	}

	var extension model.UserExtension
	if err := c.BodyParser(&extension); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	userExtRepo := repo.NewUserExtensionRepo(rt.Ctx)
	userExtService := service.NewUserExtensionService(userExtRepo)

	if err := userExtService.UpdateUserExtension(userId, &extension); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "update user extension")
	return nil
}

// updateTimezone updates user timezone
func (rt *Router) updateTimezone(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "user id is required", c.Path())
	}

	type TimezoneReq struct {
		Timezone string `json:"timezone"`
	}

	var req TimezoneReq
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	if req.Timezone == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "timezone is required", c.Path())
	}

	userExtRepo := repo.NewUserExtensionRepo(rt.Ctx)
	userExtService := service.NewUserExtensionService(userExtRepo)

	if err := userExtService.UpdateTimezone(userId, req.Timezone); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "update user timezone")
	return nil
}

// updateInvitationStatus updates invitation status
func (rt *Router) updateInvitationStatus(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "user id is required", c.Path())
	}

	type InvitationStatusReq struct {
		Status string `json:"status"` // pending, accepted, expired, revoked
	}

	var req InvitationStatusReq
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request parameters", c.Path())
	}

	if req.Status == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "status is required", c.Path())
	}

	userExtRepo := repo.NewUserExtensionRepo(rt.Ctx)
	userExtService := service.NewUserExtensionService(userExtRepo)

	if err := userExtService.UpdateInvitationStatus(userId, req.Status); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, err.Error(), c.Path())
	}

	c.Locals(middleware.OPERATION, "update invitation status")
	return nil
}
