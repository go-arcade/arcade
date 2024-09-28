package router

import (
	"github.com/cnlesscode/gotool/gintool"
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/app/engine/consts"
	"github.com/go-arcade/arcade/internal/app/engine/logic"
	"github.com/go-arcade/arcade/internal/app/engine/model"
	"github.com/go-arcade/arcade/internal/app/engine/repo"
	"github.com/go-arcade/arcade/pkg/httpx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:41
 * @file: router_agent.go
 * @description: agent router
 */

func (ar *Router) addAgent(r *gin.Context) {
	var addAgentReq *model.AddAgentReq
	agentRepo := repo.NewAgentRepo(ar.Ctx)
	agentLogic := logic.NewAgentLogic(agentRepo, addAgentReq)

	if err := r.BindJSON(&addAgentReq); err != nil {
		httpx.WithRepErrMsg(r, httpx.Failed.Code, httpx.Failed.Msg, r.Request.URL.Path)
		return
	}

	if err := agentLogic.AddAgent(addAgentReq); err != nil {
		httpx.WithRepErrMsg(r, httpx.Failed.Code, httpx.Failed.Msg, r.Request.URL.Path)
		return
	}

	r.Set(consts.OPERATION, "")
}

func (ar *Router) listAgent(r *gin.Context) {
	var agentReq *model.AddAgentReq
	agentRepo := repo.NewAgentRepo(ar.Ctx)
	agentLogic := logic.NewAgentLogic(agentRepo, agentReq)

	pageNum := queryInt(r, "pageNum")   // default 1
	pageSize := queryInt(r, "pageSize") // default 10
	agents, count, err := agentLogic.ListAgent(pageNum, pageSize)
	if err != nil {
		httpx.WithRepErrMsg(r, httpx.Failed.Code, httpx.Failed.Msg, r.Request.URL.Path)
		return
	}

	result := make(map[string]interface{})
	result["agents"] = agents
	result["count"] = count
	r.Set(consts.DETAIL, result)
}

func queryInt(r *gin.Context, key string) int {
	value, ok := gintool.QueryInt(r, key)
	if !ok {
		return 0
	}
	return value
}
