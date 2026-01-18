// Package model 提供租户配置类型定义（对齐 WeKnora）.
package model

import (
	"database/sql/driver"
	"encoding/json"
)

// PromptTemplate 单个提示词模板
type PromptTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

// PromptTemplatesConfig 提示词模板配置
type PromptTemplatesConfig struct {
	SystemPrompt    []PromptTemplate `json:"system_prompt"`
	ContextTemplate []PromptTemplate `json:"context_template"`
	RewriteSystem   []PromptTemplate `json:"rewrite_system"`
	RewriteUser     []PromptTemplate `json:"rewrite_user"`
	Fallback        []PromptTemplate `json:"fallback"`
}

// GetDefaultPromptTemplates 获取默认提示词模板
func GetDefaultPromptTemplates() *PromptTemplatesConfig {
	return &PromptTemplatesConfig{
		SystemPrompt: []PromptTemplate{
			{
				Name:        "default",
				Description: "默认系统提示词",
				Content:     "你是一个专业的AI助手，基于提供的知识库内容回答用户问题。请确保回答准确、简洁、有帮助。",
				IsDefault:   true,
			},
			{
				Name:        "professional",
				Description: "专业模式",
				Content:     "你是一位专业领域的专家顾问。请基于知识库中的专业资料，提供权威、详细的解答。使用专业术语时请附带解释。",
			},
			{
				Name:        "friendly",
				Description: "友好模式",
				Content:     "你是一位友好的助手。请用通俗易懂的语言回答问题，适当使用比喻帮助理解。",
			},
		},
		ContextTemplate: []PromptTemplate{
			{
				Name:        "default",
				Description: "默认上下文模板",
				Content:     "以下是相关的参考资料：\n\n{{contexts}}\n\n请根据以上资料回答问题：{{query}}",
				IsDefault:   true,
			},
			{
				Name:        "detailed",
				Description: "详细引用模式",
				Content:     "参考资料：\n{{contexts}}\n\n历史对话：\n{{history}}\n\n当前问题：{{query}}\n\n请综合以上信息作答，并标注信息来源。",
			},
		},
		RewriteSystem: []PromptTemplate{
			{
				Name:        "default",
				Description: "默认重写系统提示词",
				Content:     "你是一个查询优化专家。请将用户的问题改写为更清晰、更适合检索的形式。",
				IsDefault:   true,
			},
		},
		RewriteUser: []PromptTemplate{
			{
				Name:        "default",
				Description: "默认重写用户提示词",
				Content:     "请将以下问题改写为更适合知识库检索的查询：\n\n原始问题：{{query}}\n\n历史对话：{{history}}",
				IsDefault:   true,
			},
		},
		Fallback: []PromptTemplate{
			{
				Name:        "default",
				Description: "默认兜底提示词",
				Content:     "抱歉，我在知识库中没有找到与您问题直接相关的内容。请尝试换一种方式提问，或者提供更多上下文信息。",
				IsDefault:   true,
			},
		},
	}
}

// RetrieverType 检索器类型
type RetrieverType string

const (
	KeywordsRetrieverType RetrieverType = "keywords"
	VectorRetrieverType   RetrieverType = "vector"
)

// RetrieverEngineType 检索引擎类型
type RetrieverEngineType string

const (
	PostgresRetrieverEngineType      RetrieverEngineType = "postgres"
	ElasticsearchRetrieverEngineType RetrieverEngineType = "elasticsearch"
	QdrantRetrieverEngineType        RetrieverEngineType = "qdrant"
)

// RetrieverEngineParams 检索引擎参数
type RetrieverEngineParams struct {
	RetrieverType       RetrieverType       `json:"retriever_type"`
	RetrieverEngineType RetrieverEngineType `json:"retriever_engine_type"`
}

// RetrieverEngines 检索引擎配置
type RetrieverEngines struct {
	Engines []RetrieverEngineParams `json:"engines"`
}

// Value 实现 driver.Valuer 接口
func (c RetrieverEngines) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *RetrieverEngines) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// AgentConfig 租户级 Agent 配置（已废弃，保留兼容）
type AgentConfig struct {
	MaxIterations           int      `json:"max_iterations"`
	ReflectionEnabled       bool     `json:"reflection_enabled"`
	AllowedTools            []string `json:"allowed_tools"`
	Temperature             float64  `json:"temperature"`
	KnowledgeBases          []string `json:"knowledge_bases"`
	KnowledgeIDs            []string `json:"knowledge_ids"`
	SystemPrompt            string   `json:"system_prompt,omitempty"`
	SystemPromptWebEnabled  string   `json:"system_prompt_web_enabled,omitempty"`
	SystemPromptWebDisabled string   `json:"system_prompt_web_disabled,omitempty"`
	UseCustomSystemPrompt   bool     `json:"use_custom_system_prompt"`
	WebSearchEnabled        bool     `json:"web_search_enabled"`
	WebSearchMaxResults     int      `json:"web_search_max_results"`
	MultiTurnEnabled        bool     `json:"multi_turn_enabled"`
	HistoryTurns            int      `json:"history_turns"`
	MCPSelectionMode        string   `json:"mcp_selection_mode"`
	MCPServices             []string `json:"mcp_services"`
	SubAgents               []string `json:"sub_agents"`
}

// Value 实现 driver.Valuer 接口
func (c AgentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *AgentConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// ResolveSystemPrompt 获取系统提示词
func (c *AgentConfig) ResolveSystemPrompt(webSearchEnabled bool) string {
	if c == nil {
		return ""
	}
	if c.SystemPrompt != "" {
		return c.SystemPrompt
	}
	if webSearchEnabled {
		if c.SystemPromptWebEnabled != "" {
			return c.SystemPromptWebEnabled
		}
	} else {
		if c.SystemPromptWebDisabled != "" {
			return c.SystemPromptWebDisabled
		}
	}
	return ""
}

// ContextConfig 全局上下文配置
type ContextConfig struct {
	// GlobalContext 全局上下文内容
	GlobalContext string `json:"global_context"`
	// EnableForAllSessions 是否对所有会话启用
	EnableForAllSessions bool `json:"enable_for_all_sessions"`
}

// Value 实现 driver.Valuer 接口
func (c ContextConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ContextConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// WebSearchConfig 网络搜索配置（对齐 WeKnora）
type WebSearchConfig struct {
	// Enabled 是否启用
	Enabled bool `json:"enabled"`
	// Provider 搜索引擎提供商ID
	Provider string `json:"provider"`
	// APIKey API密钥（如果需要）
	APIKey string `json:"api_key"`
	// MaxResults 最大搜索结果数
	MaxResults int `json:"max_results"`
	// SearchEngineID 搜索引擎ID
	SearchEngineID string `json:"search_engine_id"`
	// IncludeDate 是否包含日期
	IncludeDate bool `json:"include_date"`
	// CompressionMethod 压缩方法：none, summary, extract, rag
	CompressionMethod string `json:"compression_method"`
	// Blacklist 黑名单规则列表
	Blacklist []string `json:"blacklist"`
	// RAG压缩相关配置
	// EmbeddingModelID 嵌入模型ID（用于RAG压缩）
	EmbeddingModelID string `json:"embedding_model_id,omitempty"`
	// EmbeddingDimension 嵌入维度（用于RAG压缩）
	EmbeddingDimension int `json:"embedding_dimension,omitempty"`
	// RerankModelID 重排模型ID（用于RAG压缩）
	RerankModelID string `json:"rerank_model_id,omitempty"`
	// DocumentFragments 文档片段数量（用于RAG压缩）
	DocumentFragments int `json:"document_fragments,omitempty"`
}

// Value 实现 driver.Valuer 接口
func (c WebSearchConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *WebSearchConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// FallbackStrategy 兜底策略
type FallbackStrategy string

const (
	FallbackStrategyFixed FallbackStrategy = "fixed"
	FallbackStrategyModel FallbackStrategy = "model"
)

// ConversationConfig 对话配置（普通模式）
type ConversationConfig struct {
	// Prompt 系统提示词
	Prompt string `json:"prompt"`
	// ContextTemplate 上下文模板
	ContextTemplate string `json:"context_template"`
	// Temperature 温度
	Temperature float64 `json:"temperature"`
	// MaxCompletionTokens 最大生成 token 数
	MaxCompletionTokens int `json:"max_completion_tokens"`

	// 检索策略参数
	MaxRounds            int     `json:"max_rounds"`
	EmbeddingTopK        int     `json:"embedding_top_k"`
	KeywordThreshold     float64 `json:"keyword_threshold"`
	VectorThreshold      float64 `json:"vector_threshold"`
	RerankTopK           int     `json:"rerank_top_k"`
	RerankThreshold      float64 `json:"rerank_threshold"`
	EnableRewrite        bool    `json:"enable_rewrite"`
	EnableQueryExpansion bool    `json:"enable_query_expansion"`

	// 模型配置
	SummaryModelID string `json:"summary_model_id"`
	RerankModelID  string `json:"rerank_model_id"`

	// 兜底策略
	FallbackStrategy string `json:"fallback_strategy"`
	FallbackResponse string `json:"fallback_response"`
	FallbackPrompt   string `json:"fallback_prompt"`

	// 重写提示词
	RewritePromptSystem string `json:"rewrite_prompt_system"`
	RewritePromptUser   string `json:"rewrite_prompt_user"`
}

// Value 实现 driver.Valuer 接口
func (c ConversationConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ConversationConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// GetDefaultConversationConfig 获取默认对话配置
func GetDefaultConversationConfig() ConversationConfig {
	return ConversationConfig{
		Prompt: "",
		ContextTemplate: `请根据以下参考资料回答用户问题。

参考资料：
{{contexts}}

用户问题：{{query}}`,
		Temperature:          0.7,
		MaxCompletionTokens:  2048,
		MaxRounds:            3,
		EmbeddingTopK:        10,
		KeywordThreshold:     0.3,
		VectorThreshold:      0.5,
		RerankTopK:           5,
		RerankThreshold:      0.5,
		EnableRewrite:        true,
		EnableQueryExpansion: true,
		FallbackStrategy:     string(FallbackStrategyModel),
	}
}
