package logic

import (
	"github.com/go-arcade/arcade/internal/app/engine/model"
	"github.com/go-arcade/arcade/internal/app/engine/repo"
	"github.com/go-arcade/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:41
 * @file: logic_agent.go
 * @description: agent logic
 */

type AgentLogic struct {
	agentRepo *repo.AgentRepo
	agentReq  *model.AddAgentReq
}

func NewAgentLogic(agentRepo *repo.AgentRepo, agentReq *model.AddAgentReq) *AgentLogic {
	return &AgentLogic{
		agentRepo: agentRepo,
		agentReq:  agentReq,
	}
}

func (al *AgentLogic) AddAgent(addAgentReq *model.AddAgentReq) error {

	var err error

	if err = al.agentRepo.AddAgent(addAgentReq); err != nil {
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
