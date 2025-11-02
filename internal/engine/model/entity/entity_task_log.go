package entity

import streamv1 "github.com/go-arcade/arcade/api/stream/v1"

type TaskLog struct {
	BaseModel
	TaskID  string               `gorm:"column:task_id" json:"taskId"`
	AgentID string               `gorm:"column:agent_id" json:"agentId"`
	Logs    []*streamv1.LogChunk `gorm:"column:logs" json:"logs"`
}
