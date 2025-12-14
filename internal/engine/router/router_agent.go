// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"errors"
	"strconv"

	agentmodel "github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func (rt *Router) agentRouter(r fiber.Router, auth fiber.Handler) {
	agentGroup := r.Group("/agent", auth)
	{
		// RESTful API
		agentGroup.Post("", rt.createAgent)       // POST /agent - create agent
		agentGroup.Get("", rt.listAgent)          // GET /agent - list agents
		agentGroup.Get("/:id", rt.getAgent)       // GET /agent/:id - get agent by id
		agentGroup.Put("/:id", rt.updateAgent)    // PUT /agent/:id - update agent
		agentGroup.Delete("/:id", rt.deleteAgent) // DELETE /agent/:id - delete agent
	}
}

// createAgent POST /agent - create a new agent
func (rt *Router) createAgent(c *fiber.Ctx) error {
	var createAgentReq *agentmodel.CreateAgentReq
	agentLogic := rt.Services.Agent

	if err := c.BodyParser(&createAgentReq); err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	agent, err := agentLogic.CreateAgent(createAgentReq)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.DETAIL, agent)
	return nil
}

// listAgent GET /agent - list agents with pagination
func (rt *Router) listAgent(c *fiber.Ctx) error {
	agentLogic := rt.Services.Agent

	pageNum := queryInt(c, "pageNum")
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := queryInt(c, "pageSize")
	if pageSize <= 0 {
		pageSize = 10
	}

	agents, count, err := agentLogic.ListAgent(pageNum, pageSize)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	result := make(map[string]any)
	result["agents"] = agents
	result["count"] = count
	result["pageNum"] = pageNum
	result["pageSize"] = pageSize
	c.Locals(middleware.DETAIL, result)
	return nil
}

// getAgent GET /agent/:id - get agent by id
func (rt *Router) getAgent(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "agent id is required", c.Path())
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid agent id", c.Path())
	}

	agentLogic := rt.Services.Agent
	agent, err := agentLogic.GetAgentById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.WithRepErrMsg(c, http.NotFound.Code, "agent not found", c.Path())
		}
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.DETAIL, agent)
	return nil
}

// updateAgent PUT /agent/:id - update agent
func (rt *Router) updateAgent(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "agent id is required", c.Path())
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid agent id", c.Path())
	}

	var updateReq *agentmodel.UpdateAgentReq
	if err := c.BodyParser(&updateReq); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid request body", c.Path())
	}

	agentLogic := rt.Services.Agent
	if err := agentLogic.UpdateAgent(id, updateReq); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.WithRepErrMsg(c, http.NotFound.Code, "agent not found", c.Path())
		}
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	// Get updated agent
	updatedAgent, err := agentLogic.GetAgentById(id)
	if err != nil {
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.DETAIL, updatedAgent)
	return nil
}

// deleteAgent DELETE /agent/:id - delete agent
func (rt *Router) deleteAgent(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "agent id is required", c.Path())
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "invalid agent id", c.Path())
	}

	agentLogic := rt.Services.Agent
	if err := agentLogic.DeleteAgent(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.WithRepErrMsg(c, http.NotFound.Code, "agent not found", c.Path())
		}
		return http.WithRepErrMsg(c, http.Failed.Code, http.Failed.Msg, c.Path())
	}

	c.Locals(middleware.OPERATION, "delete agent")
	return nil
}
