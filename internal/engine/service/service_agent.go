package service

import (
	"time"

	agentmodel "github.com/go-arcade/arcade/internal/engine/model"
	agentrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
)

type AgentService struct {
	agentRepo        agentrepo.IAgentRepository
	addAgentReq      *agentmodel.AddAgentReq
	addAgentReqagent *agentmodel.AddAgentReq
}

func NewAgentService(agentagent agentrepo.IAgentRepository, agentReq *agentmodel.AddAgentReq) *AgentService {
	return &AgentService{
		agentRepo:   agentagent,
		addAgentReq: agentReq,
	}
}

func (al *AgentService) AddAgent(addAgentReq *agentmodel.AddAgentReq) error {

	var err error
	addAgentReqagent := &agentmodel.AddAgentReqRepo{
		AddAgentReq: addAgentReq,
		AgentId:     id.GetUild(),
		IsEnabled:   1,
		CreateTime:  time.Now(),
	}
	if err = al.agentRepo.AddAgent(addAgentReqagent); err != nil {
		log.Errorf("add agent err: %v", err)
		return err
	}
	return err
}

func (al *AgentService) UpdateAgent(agent *agentmodel.Agent) error {
	var err error
	if err = al.agentRepo.UpdateAgent(agent); err != nil {
		return err
	}
	return err
}

func (al *AgentService) ListAgent(pageNum, pageSize int) ([]agentmodel.Agent, int64, error) {

	offset := (pageNum - 1) * pageSize
	agents, count, err := al.agentRepo.ListAgent(offset, pageSize)

	if err != nil {
		log.Errorf("list agent err: %v", err)
		return nil, 0, err
	}
	return agents, count, err
}
