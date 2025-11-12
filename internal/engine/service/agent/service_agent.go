package agent

import (
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
)

type AgentService struct {
	AgentRepo       repo.IAgentRepository
	addAgentReq     *model.AddAgentReq
	addAgentReqRepo *model.AddAgentReqRepo
}

func NewAgentService(agentRepo repo.IAgentRepository, agentReq *model.AddAgentReq) *AgentService {
	return &AgentService{
		AgentRepo:   agentRepo,
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
	if err = al.AgentRepo.AddAgent(addAgentReqRepo); err != nil {
		log.Errorf("add agent err: %v", err)
		return err
	}
	return err
}

func (al *AgentService) UpdateAgent(agent *model.Agent) error {
	var err error
	if err = al.AgentRepo.UpdateAgent(agent); err != nil {
		return err
	}
	return err
}

func (al *AgentService) ListAgent(pageNum, pageSize int) ([]model.Agent, int64, error) {

	offset := (pageNum - 1) * pageSize
	agents, count, err := al.AgentRepo.ListAgent(offset, pageSize)

	if err != nil {
		log.Errorf("list agent err: %v", err)
		return nil, 0, err
	}
	return agents, count, err
}
