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
	Id        int    `gorm:"primaryKey" json:"id"`
	AgentId   string `gorm:"column:agent_id" json:"agentId"`
	AgentName string `gorm:"column:agent_name" json:"agentName"`
	// todo: type? proxy?
	Address   string    `gorm:"column:address" json:"address"`
	Port      string    `gorm:"column:port" json:"port"`
	Username  string    `gorm:"column:username" json:"username"`
	Password  string    `gorm:"column:password" json:"password"`
	PublicKey string    `gorm:"column:public_key" json:"publicKey"`
	AuthType  int       `gorm:"column:auth_type" json:"authType"` // 0: password, 1: key
	IsEnable  int       `gorm:"column:is_enable" json:"isEnable"` // 0: disable, 1: enable
	CreatAt   time.Time `gorm:"column:creat_time" json:"creatAt"`
	UpdateAt  time.Time `gorm:"column:update_time" json:"updateAt"`
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
	IsEnable  int       `gorm:"column:is_enable" json:"isEnable"`
	CreatAt   time.Time `gorm:"column:creat_time" json:"creatAt"`
}
