package repo

import (
	"github.com/go-arcade/arcade/pkg/ctx"
	"time"

	"github.com/go-arcade/arcade/internal/app/engine/model"
	"github.com/go-arcade/arcade/pkg/id"
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

func (ar *AgentRepo) AddAgent(addAgentReq *model.AddAgentReq) error {

	addAgentReq.AgentId = id.GetUild()
	addAgentReq.IsEnable = 1
	addAgentReq.CreatAt = time.Now()
	var err error
	if err = ar.Ctx.GetDB().Table(ar.AgentModel.TableName()).Create(addAgentReq).Error; err != nil {
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
	offset := (pageNum - 1) * pageSize

	if err := ar.Ctx.GetDB().Table(ar.AgentModel.TableName()).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := ar.Ctx.GetDB().Table(ar.AgentModel.TableName()).Offset(offset).Limit(pageSize).Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, count, nil
}
