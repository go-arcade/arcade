package logic

import (
	"github.com/go-arcade/arcade/internal/app/basic/models"
	"github.com/go-arcade/arcade/pkg/id"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:41
 * @file: logic_agent.go
 * @description: agent logic
 */

func AddAgent(agent *models.AddAgentReq) error {
	a := models.Agent{}
	agent.AgentId = id.GetUild()
	agent.IsEnable = 1
	agent.CreatAt = time.Now()
	var err error
	if err = a.Add(agent); err != nil {
		return err
	}
	return err
}

func UpdateAgent(agent *models.Agent) error {
	var err error
	if err = agent.Update(); err != nil {
		return err
	}
	return err
}

func ListAgent(pageNum, pageSize int) ([]models.Agent, int64, error) {
	a := models.Agent{}

	offset := (pageNum - 1) * pageSize
	agents, count, err := a.List(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return agents, count, err
}
