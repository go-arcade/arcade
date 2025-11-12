package router

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
)

func (rt *Router) agentRouter(r fiber.Router, auth fiber.Handler) {
	agentGroup := r.Group("/agent", auth)
	{
		agentGroup.Post("/add", rt.addAgent)
		agentGroup.Get("/list", rt.listAgent)
	}
}

func (rt *Router) addAgent(c *fiber.Ctx) error {
	var addAgentReq *model.AddAgentReq
	agentLogic := rt.Services.Agent

	if err := c.BodyParser(&addAgentReq); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	if err := agentLogic.AddAgent(addAgentReq); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.OPERATION, "")
	return nil
}

func (rt *Router) listAgent(c *fiber.Ctx) error {
	agentLogic := rt.Services.Agent

	pageNum := queryInt(c, "pageNum")   // default 1
	pageSize := queryInt(c, "pageSize") // default 10
	agents, count, err := agentLogic.ListAgent(pageNum, pageSize)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	result := make(map[string]any)
	result["agents"] = agents
	result["count"] = count
	c.Locals(middleware.DETAIL, result)
	return nil
}
