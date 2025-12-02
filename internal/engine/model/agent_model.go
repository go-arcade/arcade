package model

import (
	"time"
)

type Agent struct {
	BaseModel
	AgentId   string `gorm:"column:agent_id" json:"agentId"`
	AgentName string `gorm:"column:agent_name" json:"agentName"`
	// todo: type? proxy?
	Address   string `gorm:"column:address" json:"address"`
	Port      string `gorm:"column:port" json:"port"`
	Username  string `gorm:"column:username" json:"username"`
	AuthType  int    `gorm:"column:auth_type" json:"authType"`   // 0: password, 1: key
	IsEnabled int    `gorm:"column:is_enabled" json:"isEnabled"` // 0: disable, 1: enable
}

func (a *Agent) TableName() string {
	return "t_agent"
}

type AddAgentReq struct {
	AgentName string `json:"agentName"`
	Address   string `json:"address"`
	Port      string `json:"port"`
	Username  string `json:"username"`
	AuthType  int    `gorm:"column:auth_type" json:"authType"`
}

type AddAgentReqRepo struct {
	AgentId string `json:"agentId"`
	*AddAgentReq
	IsEnabled  int       `gorm:"column:is_enabled" json:"isEnabled"`
	CreateTime time.Time `gorm:"column:create_time" json:"creatAt"`
}
