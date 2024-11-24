package service

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
 * @file: service_agent.go
 * @description: agent service
 */

type AgentService struct {
	agentRepo       *repo.AgentRepo
	addAgentReq     *model.AddAgentReq
	addAgentReqRepo *model.AddAgentReqRepo
}

func NewAgentService(agentRepo *repo.AgentRepo, agentReq *model.AddAgentReq) *AgentService {
	return &AgentService{
		agentRepo:   agentRepo,
		addAgentReq: agentReq,
	}
}

func (al *AgentService) AddAgent(addAgentReq *model.AddAgentReq) error {

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

func (al *AgentService) UpdateAgent() error {

	var err error
	if err = al.agentRepo.UpdateAgent(&al.agentRepo.AgentModel); err != nil {
		return err
	}
	return err
}

func (al *AgentService) ListAgent(pageNum, pageSize int) ([]model.Agent, int64, error) {

	offset := (pageNum - 1) * pageSize
	agents, count, err := al.agentRepo.ListAgent(offset, pageSize)

	if err != nil {
		log.Errorf("list agent err: %v", err)
		return nil, 0, err
	}
	return agents, count, err
}
