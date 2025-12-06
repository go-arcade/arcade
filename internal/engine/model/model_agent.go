package model

import (
	"time"

	"gorm.io/datatypes"
)

type Agent struct {
	BaseModel
	AgentId   string `gorm:"column:agent_id" json:"agentId"`
	AgentName string `gorm:"column:agent_name" json:"agentName"`
	// todo: type? proxy?
	Address           string         `gorm:"column:address" json:"address"`
	Port              string         `gorm:"column:port" json:"port"`
	OS                string         `gorm:"column:os" json:"os"`
	Arch              string         `gorm:"column:arch" json:"arch"`
	Version           string         `gorm:"column:version" json:"version"`
	Status            int            `gorm:"column:status" json:"status"` // 0: unknown, 1: online, 2: offline, 3: busy, 4: idle
	Labels            datatypes.JSON `gorm:"column:labels" json:"labels"`
	Metrics           datatypes.JSON `gorm:"column:metrics" json:"metrics"`
	LastHeartbeat     time.Time      `gorm:"column:last_heartbeat" json:"lastHeartbeat"`
	IsEnabled         int            `gorm:"column:is_enabled" json:"isEnabled"` // 0: disable, 1: enable
}

func (a *Agent) TableName() string {
	return "t_agent"
}

type AddAgentReq struct {
	AgentName string `json:"agentName"`
	Address   string `json:"address"`
	Port      string `json:"port"`
}

type AddAgentReqRepo struct {
	AgentId string `json:"agentId"`
	*AddAgentReq
	IsEnabled  int       `gorm:"column:is_enabled" json:"isEnabled"`
	CreateTime time.Time `gorm:"column:create_time" json:"creatAt"`
}
