package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/21 20:13
 * @file: repo_agent.go
 * @description: agent repo
 */

type AgentRepo struct {
	Ctx        *ctx.Context
	AgentModel model.Agent
}

func NewAgentRepo(ctx *ctx.Context) *AgentRepo {
	return &AgentRepo{
		Ctx:        ctx,
		AgentModel: model.Agent{},
	}
}

func (ar *AgentRepo) AddAgent(addAgentReqRepo *model.AddAgentReqRepo) error {

	var err error
	if err = ar.Ctx.GetDB().Table(ar.AgentModel.TableName()).Create(addAgentReqRepo).Error; err != nil {
		return err
	}
	return err
}

func (ar *AgentRepo) UpdateAgent(agent *model.Agent) error {
	var err error
	if err = ar.Ctx.GetDB().Model(agent).Updates(agent).Error; err != nil {
		return err
	}
	return err
}

func (ar *AgentRepo) ListAgent(pageNum, pageSize int) ([]model.Agent, int64, error) {

	var agents []model.Agent
	var count int64
	var err error
	offset := (pageNum - 1) * pageSize

	if err = ar.Ctx.GetDB().Table(ar.AgentModel.TableName()).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err = ar.Ctx.GetDB().Select("id, agent_id, agent_name, address, port, username, auth_type, is_enabled").
		Table(ar.AgentModel.TableName()).
		Offset(offset).Limit(pageSize).Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, count, nil
}
