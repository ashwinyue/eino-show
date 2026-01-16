// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package model

import (
	"time"
)

const TableNameCustomAgentM = "custom_agents"

// CustomAgentM 自定义 Agent 模型
type CustomAgentM struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	Name      string    `gorm:"column:name;not null;type:varchar(255);comment:Agent 名称" json:"name"`
	Description string  `gorm:"column:description;type:text;comment:Agent 描述" json:"description"`
	Avatar    string    `gorm:"column:avatar;type:varchar(64);comment:Agent 头像" json:"avatar,omitempty"`
	IsBuiltin bool     `gorm:"column:is_builtin;not null;default:false;comment:是否为内置 Agent" json:"is_builtin"`
	TenantID  uint64   `gorm:"column:tenant_id;not null;index:idx_tenant;comment:租户ID" json:"tenant_id"`
	CreatedBy string   `gorm:"column:created_by;type:varchar(36);comment:创建者 ID" json:"created_by,omitempty"`
	// Config Agent 配置（JSON 格式）
	// 包含：instruction, temperature, max_iterations, tools, knowledge_bases 等
	Config     *CustomAgentConfig `gorm:"column:config;type:jsonb;not null;comment:Agent 配置" json:"config"`
	CreatedAt  time.Time          `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time          `gorm:"column:updated_at;not null;default:current_timestamp;comment:更新时间" json:"updated_at"`
	DeletedAt  *time.Time         `gorm:"column:deleted_at;index:idx_deleted;comment:删除时间" json:"deleted_at,omitempty"`
}

// TableName CustomAgentM's table name
func (*CustomAgentM) TableName() string {
	return TableNameCustomAgentM
}

// CustomAgentConfig Agent 配置
type CustomAgentConfig struct {
	// Instruction 系统提示词
	Instruction string `json:"instruction,omitempty"`
	// Temperature 温度参数
	Temperature float64 `json:"temperature,omitempty"`
	// MaxIterations 最大迭代次数
	MaxIterations int `json:"max_iterations,omitempty"`
	// Tools 可用工具列表
	Tools []string `json:"tools,omitempty"`
	// KnowledgeBases 关联的知识库 ID 列表
	KnowledgeBases []string `json:"knowledge_bases,omitempty"`
	// ModelID 使用的模型 ID
	ModelID string `json:"model_id,omitempty"`
	// EnableWebSearch 是否启用网络搜索
	EnableWebSearch bool `json:"enable_web_search,omitempty"`
}
