package model

import (
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 20:33
 * @file: model_agent.go
 * @description: agent model
 */

type Agent struct {
	BaseModel
	AgentId   string `gorm:"column:agent_id" json:"agentId"`
	AgentName string `gorm:"column:agent_name" json:"agentName"`
	// todo: type? proxy?
	Address   string `gorm:"column:address" json:"address"`
	Port      string `gorm:"column:port" json:"port"`
	Username  string `gorm:"column:username" json:"username"`
	Password  string `gorm:"column:password" json:"password,omitempty"`
	PublicKey string `gorm:"column:public_key" json:"publicKey,omitempty"`
	AuthType  int    `gorm:"column:auth_type" json:"authType"`   // 0: password, 1: key
	IsEnabled int    `gorm:"column:is_enabled" json:"isEnabled"` // 0: disable, 1: enable
}

func (a *Agent) TableName() string {
	return "agent"
}

type AddAgentReq struct {
	AgentId   string    `json:"agentId"`
	AgentName string    `json:"agentName"`
	Address   string    `json:"address"`
	Port      string    `json:"port"`
	Username  string    `json:"username"`
	Password  string    `gorm:"column:password" json:"password"`
	PublicKey string    `gorm:"column:public_key" json:"publicKey"`
	AuthType  int       `gorm:"column:auth_type" json:"authType"`
	IsEnabled int       `gorm:"column:is_enabled" json:"isEnabled"`
	CreatAt   time.Time `gorm:"column:creat_time" json:"creatAt"`
}
