// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== Model 请求/响应类型 =====

// CreateModelRequest 创建模型请求（对齐 WeKnora 格式）
type CreateModelRequest struct {
	Name        string            `json:"name" binding:"required"`
	Type        string            `json:"type" binding:"required"`        // KnowledgeQA, Embedding, Rerank, VLLM
	Source      string            `json:"source" binding:"required"`      // remote, builtin
	Description string            `json:"description"`
	Parameters  ModelParameters   `json:"parameters" binding:"required"`
}

// ModelParameters 模型参数（对齐 WeKnora）
type ModelParameters struct {
	BaseURL             string                `json:"base_url"`
	APIKey              string                `json:"api_key"`
	Provider            string                `json:"provider"`
	EmbeddingParameters EmbeddingParameters   `json:"embedding_parameters,omitempty"`
}

// EmbeddingParameters Embedding 模型参数
type EmbeddingParameters struct {
	Dimension int `json:"dimension"`
}

// CreateModelResponse 创建模型响应（对齐 WeKnora 格式）
type CreateModelResponse struct {
	Success bool          `json:"success"`
	Data    *ModelResponse `json:"data"`
}

// GetModelRequest 获取模型请求
type GetModelRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetModelResponse 获取模型响应（对齐 WeKnora 格式）
type GetModelResponse struct {
	Success bool          `json:"success"`
	Data    *ModelResponse `json:"data"`
}

// ListModelsRequest 模型列表请求
type ListModelsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Provider string `form:"provider"`
}

// ListModelsResponse 模型列表响应（对齐 WeKnora 格式）
type ListModelsResponse struct {
	Success bool            `json:"success"`
	Data    []*ModelResponse `json:"data"`
	Total   int64            `json:"total,omitempty"`
}

// UpdateModelRequest 更新模型请求（对齐 WeKnora 格式）
type UpdateModelRequest struct {
	Id          string              `uri:"id" binding:"required"`
	Name        *string             `json:"name"`
	Type        *string             `json:"type"`        // KnowledgeQA, Embedding, Rerank, VLLM
	Source      *string             `json:"source"`      // remote, builtin, local
	Description *string             `json:"description"`
	Parameters  *ModelParameters    `json:"parameters"`
	IsDefault   *bool               `json:"is_default"`
}

// UpdateModelResponse 更新模型响应（对齐 WeKnora 格式）
type UpdateModelResponse struct {
	Success bool          `json:"success"`
	Data    *ModelResponse `json:"data"`
}

// DeleteModelRequest 删除模型请求
type DeleteModelRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteModelResponse 删除模型响应
type DeleteModelResponse struct {
	Success bool `json:"success"`
}

// SetDefaultModelRequest 设置默认模型请求
type SetDefaultModelRequest struct {
	Id string `uri:"id" binding:"required"`
}

// ModelResponse 模型响应（对齐 WeKnora 格式）
type ModelResponse struct {
	ID          string                 `json:"id"`
	TenantID    uint64                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`        // KnowledgeQA, Embedding, Rerank, VLLM
	Source      string                 `json:"source"`      // remote, builtin, local
	Description string                 `json:"description"`
	Parameters  *ModelParameters       `json:"parameters"`
	IsDefault   bool                   `json:"is_default"`
	IsBuiltin   bool                   `json:"is_builtin"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SetDefaultModelResponse 设置默认模型响应
type SetDefaultModelResponse struct {
	Success bool `json:"success"`
}

// ListProvidersRequest 获取模型提供商列表请求
type ListProvidersRequest struct {
	ModelType string `form:"model_type"`
}

// ListProvidersResponse 获取模型提供商列表响应（对齐 WeKnora 格式）
type ListProvidersResponse struct {
	Success bool            `json:"success"`
	Data    []*ProviderInfo `json:"data"`
}

// ProviderInfo 模型提供商信息（对齐 WeKnora，使用驼峰命名）
type ProviderInfo struct {
	Value       string            `json:"value"`
	Label       string            `json:"label"`
	Description string            `json:"description"`
	DefaultURLs map[string]string `json:"defaultUrls"`  // 驼峰命名
	ModelTypes  []string          `json:"modelTypes"`   // 驼峰命名
}

// ===== 模型测试请求/响应类型 =====

// TestChatModelRequest 测试 Chat 模型请求
type TestChatModelRequest struct {
	Provider  string `json:"provider" binding:"required"`
	ModelName string `json:"model_name"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key" binding:"required"`
}

// TestChatModelResponse 测试 Chat 模型响应
type TestChatModelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
	Latency int64  `json:"latency_ms,omitempty"` // 响应延迟（毫秒）
}

// TestEmbeddingModelRequest 测试 Embedding 模型请求
type TestEmbeddingModelRequest struct {
	Provider  string `json:"provider" binding:"required"`
	ModelName string `json:"model_name"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key" binding:"required"`
	Text      string `json:"text"`
}

// TestEmbeddingModelResponse 测试 Embedding 模型响应
type TestEmbeddingModelResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Dimension int    `json:"dimension,omitempty"` // 向量维度
	Latency   int64  `json:"latency_ms,omitempty"`
}

// TestRerankModelRequest 测试 Rerank 模型请求
type TestRerankModelRequest struct {
	Provider  string `json:"provider" binding:"required"`
	ModelName string `json:"model_name"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key" binding:"required"`
}

// TestRerankModelResponse 测试 Rerank 模型响应
type TestRerankModelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency int64  `json:"latency_ms,omitempty"`
}
