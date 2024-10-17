package logic

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:41
 * @file: logic_agent.go
 * @description: agent logic
 */

type AgentLogic struct {
	agentRepo       *repo.AgentRepo
	addAgentReq     *model.AddAgentReq
	addAgentReqRepo *model.AddAgentReqRepo
}

func NewAgentLogic(agentRepo *repo.AgentRepo, agentReq *model.AddAgentReq) *AgentLogic {
	return &AgentLogic{
		agentRepo:   agentRepo,
		addAgentReq: agentReq,
	}
}

func (al *AgentLogic) AddAgent(addAgentReq *model.AddAgentReq) error {

	var err error
	addAgentReqRepo := &model.AddAgentReqRepo{
		AddAgentReq: addAgentReq,
		AgentId:     id.GetUild(),
		IsEnabled:   1,
		CreateTime:  time.Now(),
	}
	if err = al.agentRepo.AddAgent(addAgentReqRepo); err != nil {
		log.Errorf("add agent err: %v", err)
		return err
	}
	return err
}

func (al *AgentLogic) UpdateAgent() error {

	var err error
	if err = al.agentRepo.UpdateAgent(&al.agentRepo.AgentModel); err != nil {
		return err
	}
	return err
}

func (al *AgentLogic) ListAgent(pageNum, pageSize int) ([]model.Agent, int64, error) {

	offset := (pageNum - 1) * pageSize
	agents, count, err := al.agentRepo.ListAgent(offset, pageSize)

	if err != nil {
		log.Errorf("list agent err: %v", err)
		return nil, 0, err
	}
	return agents, count, err
}
