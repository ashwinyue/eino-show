// Package model 提供 Session 配置类型定义（对齐 WeKnora）.
package model

import (
	"database/sql/driver"
	"encoding/json"
)

// SummaryConfig 总结模型配置
type SummaryConfig struct {
	MaxTokens           int     `json:"max_tokens"`
	RepeatPenalty       float64 `json:"repeat_penalty"`
	TopK                int     `json:"top_k"`
	TopP                float64 `json:"top_p"`
	FrequencyPenalty    float64 `json:"frequency_penalty"`
	PresencePenalty     float64 `json:"presence_penalty"`
	Prompt              string  `json:"prompt"`
	ContextTemplate     string  `json:"context_template"`
	NoMatchPrefix       string  `json:"no_match_prefix"`
	Temperature         float64 `json:"temperature"`
	Seed                int     `json:"seed"`
	MaxCompletionTokens int     `json:"max_completion_tokens"`
	Thinking            *bool   `json:"thinking"`
}

// Value 实现 driver.Valuer 接口
func (c SummaryConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *SummaryConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// ContextCompressionStrategy 上下文压缩策略
type ContextCompressionStrategy string

const (
	// ContextCompressionSlidingWindow 滑动窗口策略
	ContextCompressionSlidingWindow ContextCompressionStrategy = "sliding_window"
	// ContextCompressionSmart 智能压缩策略
	ContextCompressionSmart ContextCompressionStrategy = "smart"
)

// SessionContextConfig 会话级上下文配置
type SessionContextConfig struct {
	MaxTokens           int                        `json:"max_tokens"`
	CompressionStrategy ContextCompressionStrategy `json:"compression_strategy"`
	RecentMessageCount  int                        `json:"recent_message_count"`
	SummarizeThreshold  int                        `json:"summarize_threshold"`
}

// Value 实现 driver.Valuer 接口
func (c SessionContextConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *SessionContextConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// SessionAgentConfig 会话级 Agent 配置
type SessionAgentConfig struct {
	Enabled             bool     `json:"enabled"`
	AgentID             string   `json:"agent_id"`
	KnowledgeBases      []string `json:"knowledge_bases"`
	KnowledgeIDs        []string `json:"knowledge_ids"`
	WebSearchEnabled    bool     `json:"web_search_enabled"`
	MCPServices         []string `json:"mcp_services"`
	Temperature         float64  `json:"temperature"`
	MaxIterations       int      `json:"max_iterations"`
	MaxCompletionTokens int      `json:"max_completion_tokens"`
}

// Value 实现 driver.Valuer 接口
func (c SessionAgentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *SessionAgentConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// EnsureDefaults 设置默认值
func (c *SessionAgentConfig) EnsureDefaults() {
	if c.Temperature == 0 {
		c.Temperature = 0.7
	}
	if c.MaxIterations == 0 {
		c.MaxIterations = 10
	}
	if c.MaxCompletionTokens == 0 {
		c.MaxCompletionTokens = 2048
	}
}

// GetDefaultSummaryConfig 获取默认总结配置
func GetDefaultSummaryConfig() SummaryConfig {
	return SummaryConfig{
		MaxTokens:           4096,
		Temperature:         0.7,
		MaxCompletionTokens: 2048,
		Prompt:              "",
		ContextTemplate: `请根据以下参考资料回答用户问题。

参考资料：
{{contexts}}

用户问题：{{query}}`,
	}
}

// GetDefaultSessionContextConfig 获取默认会话上下文配置
func GetDefaultSessionContextConfig() SessionContextConfig {
	return SessionContextConfig{
		MaxTokens:           8192,
		CompressionStrategy: ContextCompressionSlidingWindow,
		RecentMessageCount:  10,
		SummarizeThreshold:  20,
	}
}
