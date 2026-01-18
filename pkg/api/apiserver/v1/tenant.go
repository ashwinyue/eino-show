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

// WebSearchConfig 网络搜索配置（对齐 WeKnora）
type WebSearchConfig struct {
	Enabled           bool     `json:"enabled"`            // 是否启用
	Provider          string   `json:"provider"`           // 搜索引擎提供商ID
	APIKey            string   `json:"api_key"`            // API密钥（如果需要）
	MaxResults        int      `json:"max_results"`        // 最大搜索结果数
	SearchEngineID    string   `json:"search_engine_id"`   // 搜索引擎ID
	IncludeDate       bool     `json:"include_date"`       // 是否包含日期
	CompressionMethod string   `json:"compression_method"` // 压缩方法：none, summary, extract, rag
	Blacklist         []string `json:"blacklist"`          // 黑名单规则列表
	// RAG压缩相关配置
	EmbeddingModelID   string `json:"embedding_model_id,omitempty"`  // 嵌入模型ID
	EmbeddingDimension int    `json:"embedding_dimension,omitempty"` // 嵌入维度
	RerankModelID      string `json:"rerank_model_id,omitempty"`     // 重排模型ID
	DocumentFragments  int    `json:"document_fragments,omitempty"`  // 文档片段数量
}

// ConversationConfig 对话配置（对齐 WeKnora）
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
	// 模型配置
	SummaryModelID string `json:"summary_model_id"`
	RerankModelID  string `json:"rerank_model_id"`
	// 降级策略
	FallbackStrategy string `json:"fallback_strategy"`
	FallbackResponse string `json:"fallback_response"`
	FallbackPrompt   string `json:"fallback_prompt"`
	// 重写提示词
	RewritePromptSystem string `json:"rewrite_prompt_system"`
	RewritePromptUser   string `json:"rewrite_prompt_user"`
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

// TenantKVResponse KV 配置响应（对齐前端期望格式）
type TenantKVResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

// UpdateTenantKVRequest 更新 KV 配置请求
type UpdateTenantKVRequest struct {
	Key   string      `uri:"key" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// UpdateTenantKVResponse 更新 KV 配置响应（对齐 WeKnora）
type UpdateTenantKVResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
