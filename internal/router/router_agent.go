package router

import (
	"github.com/arcade/arcade/internal/app/basic/logic"
	"github.com/arcade/arcade/internal/app/basic/models"
	"github.com/arcade/arcade/pkg/httpx"
	"github.com/cnlesscode/gotool/gintool"
	"github.com/gin-gonic/gin"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:41
 * @file: router_agent.go
 * @description: agent router
 */

func addAgent(r *gin.Context) {

	var agent *models.AddAgentReq

	if err := r.BindJSON(&agent); err != nil {
		httpx.WithRepErrMsg(r, httpx.Failed.Code, httpx.Failed.Msg, r.Request.URL.Path)
		return
	}

	if err := logic.AddAgent(agent); err != nil {
		httpx.WithRepErrMsg(r, httpx.Failed.Code, httpx.Failed.Msg, r.Request.URL.Path)
		return
	}

	httpx.WithRepJSON(r, nil)

}

func listAgent(r *gin.Context) {

	pageNum, _ := gintool.QueryInt(r, "pageNum")   // default 1
	pageSize, _ := gintool.QueryInt(r, "pageSize") // default 10

	agents, count, err := logic.ListAgent(pageNum, pageSize)
	if err != nil {
		httpx.WithRepErrMsg(r, httpx.Failed.Code, httpx.Failed.Msg, r.Request.URL.Path)
		return
	}

	result := make(map[string]interface{})
	result["agents"] = agents
	result["count"] = count
	httpx.WithRepJSON(r, result)

}
