// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package model

import (
	"time"
)

const TableNameSessionM = "sessions"

// SessionM 会话模型，用于管理用户与 Agent 的对话会话
type SessionM struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID  uint64    `gorm:"column:tenant_id;not null;index:idx_tenant;comment:租户ID" json:"tenant_id"`
	Title     string    `gorm:"column:title;type:varchar(255);comment:会话标题" json:"title"`
	Description string  `gorm:"column:description;type:text;comment:会话描述" json:"description"`
	AgentID   string    `gorm:"column:agent_id;type:varchar(36);index:idx_agent;comment:关联的Agent ID" json:"agent_id"`
	// AgentConfig 会话级别的 Agent 配置（JSON 格式）
	// 包含：agent_id, mode, knowledge_bases, temperature, max_iterations, web_search_enabled 等
	AgentConfig   *SessionAgentConfig `gorm:"column:agent_config;type:jsonb;comment:Agent 配置" json:"agent_config,omitempty"`
	// ContextConfig LLM 上下文管理配置（JSON 格式）
	// 包含：max_messages, compression_threshold 等
	ContextConfig *SessionContextConfig `gorm:"column:context_config;type:jsonb;comment:上下文配置" json:"context_config,omitempty"`
	CreatedAt     time.Time           `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time           `gorm:"column:updated_at;not null;default:current_timestamp;comment:更新时间" json:"updated_at"`
}

// TableName SessionM's table name
func (*SessionM) TableName() string {
	return TableNameSessionM
}

// SessionAgentConfig 会话级别的 Agent 配置
type SessionAgentConfig struct {
	AgentID          string   `json:"agent_id,omitempty"`           // Agent ID
	Mode             string   `json:"mode,omitempty"`              // 模式：rag, react, chat
	KnowledgeBases   []string `json:"knowledge_bases,omitempty"`    // 关联的知识库 ID 列表
	Temperature      float64  `json:"temperature,omitempty"`       // 温度参数
	MaxIterations    int      `json:"max_iterations,omitempty"`    // 最大迭代次数
	WebSearchEnabled bool     `json:"web_search_enabled,omitempty"` // 是否启用网络搜索
}

// SessionContextConfig LLM 上下文管理配置
type SessionContextConfig struct {
	MaxMessages           int     `json:"max_messages,omitempty"`             // 最大消息数量
	CompressionThreshold  int     `json:"compression_threshold,omitempty"`   // 压缩阈值
	EnableContextCompression bool `json:"enable_context_compression,omitempty"` // 是否启用上下文压缩
}
