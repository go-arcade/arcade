// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import "time"

// PluginInstallRecords 插件安装任务
type PluginInstallRecords struct {
	TaskID        string     `bson:"task_id" json:"taskId"`                                   // 任务ID（唯一标识）
	PluginName    string     `bson:"plugin_name" json:"pluginName"`                           // 插件名称
	Version       string     `bson:"version" json:"version"`                                  // 版本
	Status        string     `bson:"status" json:"status"`                                    // 状态: pending/running/completed/failed
	Progress      int        `bson:"progress" json:"progress"`                                // 进度 0-100
	Message       string     `bson:"message" json:"message"`                                  // 消息
	Error         string     `bson:"error,omitempty" json:"error,omitempty"`                  // 错误信息
	PluginID      string     `bson:"plugin_id,omitempty" json:"pluginId,omitempty"`           // 安装成功后的插件ID
	Source        string     `bson:"source" json:"source"`                                    // 来源: local/market
	CreateTime    time.Time  `bson:"create_time" json:"createTime"`                           // 创建时间
	UpdateTime    time.Time  `bson:"update_time" json:"updateTime"`                           // 更新时间
	CompletedTime *time.Time `bson:"completed_time,omitempty" json:"completedTime,omitempty"` // 完成时间
	Duration      int64      `bson:"duration,omitempty" json:"duration,omitempty"`            // 安装耗时（秒）
}

// CollectionName 返回集合名称
func (PluginInstallRecords) CollectionName() string {
	return "c_plugin_install_records"
}

// TaskStatus 任务状态常量
const (
	TaskStatusPending   = "pending"   // 等待中
	TaskStatusRunning   = "running"   // 执行中
	TaskStatusCompleted = "completed" // 已完成
	TaskStatusFailed    = "failed"    // 失败
)
