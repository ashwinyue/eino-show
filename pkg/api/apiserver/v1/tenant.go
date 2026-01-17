// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== Tenant 请求/响应类型（对齐 WeKnora）=====

// Tenant 租户信息（完整版，对齐 WeKnora）
type TenantFull struct {
	ID                 uint64              `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	APIKey             string              `json:"api_key"`
	Status             string              `json:"status"`
	Business           string              `json:"business"`
	StorageQuota       int64               `json:"storage_quota"`
	StorageUsed        int64               `json:"storage_used"`
	RetrieverEngines   *RetrieverEngines   `json:"retriever_engines,omitempty"`
	AgentConfig        *AgentConfig        `json:"agent_config,omitempty"`
	ContextConfig      *ContextConfig      `json:"context_config,omitempty"`
	WebSearchConfig    *WebSearchConfig    `json:"web_search_config,omitempty"`
	ConversationConfig *ConversationConfig `json:"conversation_config,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// RetrieverEngines 检索引擎配置
type RetrieverEngines struct {
	Engines []RetrieverEngineParams `json:"engines"`
}

// RetrieverEngineParams 检索引擎参数
type RetrieverEngineParams struct {
	RetrieverType       string `json:"retriever_type"`
	RetrieverEngineType string `json:"retriever_engine_type"`
}

// AgentConfig 租户级 Agent 配置
type AgentConfig struct {
	MaxIterations         int      `json:"max_iterations"`
	ReflectionEnabled     bool     `json:"reflection_enabled"`
	AllowedTools          []string `json:"allowed_tools"`
	Temperature           float64  `json:"temperature"`
	SystemPrompt          string   `json:"system_prompt,omitempty"`
	UseCustomSystemPrompt bool     `json:"use_custom_system_prompt"`
	WebSearchEnabled      bool     `json:"web_search_enabled"`
	WebSearchMaxResults   int      `json:"web_search_max_results"`
	MultiTurnEnabled      bool     `json:"multi_turn_enabled"`
	HistoryTurns          int      `json:"history_turns"`
}

// ContextConfig 全局上下文配置
type ContextConfig struct {
	GlobalContext        string `json:"global_context"`
	EnableForAllSessions bool   `json:"enable_for_all_sessions"`
}

// WebSearchConfig 网络搜索配置
type WebSearchConfig struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"`
	MaxResults     int    `json:"max_results"`
	SearchEngineID string `json:"search_engine_id"`
}

// ConversationConfig 对话配置
type ConversationConfig struct {
	Prompt               string  `json:"prompt"`
	ContextTemplate      string  `json:"context_template"`
	Temperature          float64 `json:"temperature"`
	MaxCompletionTokens  int     `json:"max_completion_tokens"`
	MaxRounds            int     `json:"max_rounds"`
	EmbeddingTopK        int     `json:"embedding_top_k"`
	KeywordThreshold     float64 `json:"keyword_threshold"`
	VectorThreshold      float64 `json:"vector_threshold"`
	RerankTopK           int     `json:"rerank_top_k"`
	RerankThreshold      float64 `json:"rerank_threshold"`
	EnableRewrite        bool    `json:"enable_rewrite"`
	EnableQueryExpansion bool    `json:"enable_query_expansion"`
	FallbackStrategy     string  `json:"fallback_strategy"`
	FallbackResponse     string  `json:"fallback_response"`
	FallbackPrompt       string  `json:"fallback_prompt"`
}

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name             string            `json:"name" binding:"required"`
	Description      string            `json:"description"`
	Business         string            `json:"business"`
	StorageQuota     int64             `json:"storage_quota"`
	RetrieverEngines *RetrieverEngines `json:"retriever_engines,omitempty"`
}

// CreateTenantResponse 创建租户响应
type CreateTenantResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Tenant  *TenantFull `json:"tenant,omitempty"`
}

// GetTenantRequest 获取租户请求
type GetTenantRequest struct {
	ID uint64 `uri:"id" binding:"required"`
}

// GetTenantResponse 获取租户响应
type GetTenantResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Tenant  *TenantFull `json:"tenant,omitempty"`
}

// ListTenantsRequest 租户列表请求
type ListTenantsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Name     string `form:"name"`
}

// ListTenantsResponse 租户列表响应
type ListTenantsResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Tenants []*TenantFull `json:"tenants"`
	Total   int64         `json:"total"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	ID               uint64            `uri:"id" binding:"required"`
	Name             *string           `json:"name"`
	Description      *string           `json:"description"`
	Business         *string           `json:"business"`
	Status           *string           `json:"status"`
	StorageQuota     *int64            `json:"storage_quota"`
	RetrieverEngines *RetrieverEngines `json:"retriever_engines,omitempty"`
}

// UpdateTenantResponse 更新租户响应
type UpdateTenantResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Tenant  *TenantFull `json:"tenant,omitempty"`
}

// DeleteTenantRequest 删除租户请求
type DeleteTenantRequest struct {
	ID uint64 `uri:"id" binding:"required"`
}

// DeleteTenantResponse 删除租户响应
type DeleteTenantResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// SearchTenantsRequest 搜索租户请求
type SearchTenantsRequest struct {
	Query    string `form:"query"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// SearchTenantsResponse 搜索租户响应
type SearchTenantsResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Tenants []*TenantFull `json:"tenants"`
	Total   int64         `json:"total"`
}

// TenantKVRequest KV 配置请求
type TenantKVRequest struct {
	Key string `uri:"key" binding:"required"`
}

// TenantKVResponse KV 配置响应
type TenantKVResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
}

// UpdateTenantKVRequest 更新 KV 配置请求
type UpdateTenantKVRequest struct {
	Key   string      `uri:"key" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// UpdateTenantKVResponse 更新 KV 配置响应
type UpdateTenantKVResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
