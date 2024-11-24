package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/engine/constant"
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/pkg/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:41
 * @file: router_agent.go
 * @description: agent router
 */

func (rt *Router) addAgent(r *gin.Context) {
	var addAgentReq *model.AddAgentReq
	agentRepo := repo.NewAgentRepo(rt.Ctx)
	agentLogic := service.NewAgentService(agentRepo, addAgentReq)

	if err := r.BindJSON(&addAgentReq); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	if err := agentLogic.AddAgent(addAgentReq); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	r.Set(constant.OPERATION, "")
}

func (rt *Router) listAgent(r *gin.Context) {
	var agentReq *model.AddAgentReq
	agentRepo := repo.NewAgentRepo(rt.Ctx)
	agentLogic := service.NewAgentService(agentRepo, agentReq)

	pageNum := queryInt(r, "pageNum")   // default 1
	pageSize := queryInt(r, "pageSize") // default 10
	agents, count, err := agentLogic.ListAgent(pageNum, pageSize)
	if err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	result := make(map[string]interface{})
	result["agents"] = agents
	result["count"] = count
	r.Set(constant.DETAIL, result)
}
