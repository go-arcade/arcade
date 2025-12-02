package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IAgentRepository interface {
	AddAgent(addAgentReqRepo *model.AddAgentReqRepo) error
	UpdateAgent(a *model.Agent) error
	ListAgent(pageNum, pageSize int) ([]model.Agent, int64, error)
}

type AgentRepo struct {
	db         database.IDatabase
	AgentModel model.Agent
}

func NewAgentRepo(db database.IDatabase) IAgentRepository {
	return &AgentRepo{
		db:         db,
		AgentModel: model.Agent{},
	}
}

func (ar *AgentRepo) AddAgent(addAgentReqRepo *model.AddAgentReqRepo) error {
	var err error
	if err = ar.db.Database().Table(ar.AgentModel.TableName()).Create(addAgentReqRepo).Error; err != nil {
		return err
	}
	return err
}

func (ar *AgentRepo) UpdateAgent(a *model.Agent) error {
	var err error
	if err = ar.db.Database().Model(a).Updates(a).Error; err != nil {
		return err
	}
	return err
}

func (ar *AgentRepo) ListAgent(pageNum, pageSize int) ([]model.Agent, int64, error) {
	var agents []model.Agent
	var count int64
	var err error
	offset := (pageNum - 1) * pageSize

	if err = ar.db.Database().Table(ar.AgentModel.TableName()).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err = ar.db.Database().Select("id, agent_id, agent_name, address, port, username, auth_type, is_enabled").
		Table(ar.AgentModel.TableName()).
		Offset(offset).Limit(pageSize).Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, count, nil
}
