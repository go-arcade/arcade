package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/engine/consts"
	"github.com/go-arcade/arcade/internal/engine/logic"
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
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
	agentLogic := logic.NewAgentLogic(agentRepo, addAgentReq)

	if err := r.BindJSON(&addAgentReq); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	if err := agentLogic.AddAgent(addAgentReq); err != nil {
		http.WithRepErrMsg(r, http.Failed.Code, http.Failed.Msg, r.Request.URL.Path)
		return
	}

	r.Set(consts.OPERATION, "")
}

func (rt *Router) listAgent(r *gin.Context) {
	var agentReq *model.AddAgentReq
	agentRepo := repo.NewAgentRepo(rt.Ctx)
	agentLogic := logic.NewAgentLogic(agentRepo, agentReq)

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
	r.Set(consts.DETAIL, result)
}
