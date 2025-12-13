package model

import (
	"gorm.io/datatypes"
)

type Agent struct {
	BaseModel
	AgentId   string         `gorm:"column:agent_id" json:"agentId"`
	AgentName string         `gorm:"column:agent_name" json:"agentName"`
	Address   string         `gorm:"column:address" json:"address"`
	Port      string         `gorm:"column:port" json:"port"`
	OS        string         `gorm:"column:os" json:"os"`
	Arch      string         `gorm:"column:arch" json:"arch"`
	Version   string         `gorm:"column:version" json:"version"`
	Status    int            `gorm:"column:status" json:"status"` // 0: unknown, 1: online, 2: offline, 3: busy, 4: idle
	Labels    datatypes.JSON `gorm:"column:labels" json:"labels"`
	Metrics   string         `gorm:"column:metrics" json:"metrics"`
	IsEnabled int            `gorm:"column:is_enabled" json:"isEnabled"` // 0: disable, 1: enable
}

func (a *Agent) TableName() string {
	return "t_agent"
}

// CreateAgentReq request for creating agent
type CreateAgentReq struct {
	AgentName string         `json:"agentName"`
	Status    int            `json:"status"`
	Labels    datatypes.JSON `json:"labels"`
}

// CreateAgentResp response for creating agent
type CreateAgentResp struct {
	Agent
	Token string `json:"token"` // Token for agent communication authentication
}

// UpdateAgentReq request for updating agent (AgentId is not allowed to be modified)
type UpdateAgentReq struct {
	AgentName *string        `json:"agentName,omitempty"`
	Address   *string        `json:"address,omitempty"`
	Port      *string        `json:"port,omitempty"`
	OS        *string        `json:"os,omitempty"`
	Arch      *string        `json:"arch,omitempty"`
	Version   *string        `json:"version,omitempty"`
	Status    *int           `json:"status,omitempty"`
	Labels    datatypes.JSON `json:"labels,omitempty"`
	Metrics   datatypes.JSON `json:"metrics,omitempty"`
	IsEnabled *int           `json:"isEnabled,omitempty"`
}

// AgentDetail response for agent detail
type AgentDetail struct {
	Agent
}
