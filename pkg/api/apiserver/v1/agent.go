// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

// ===== Agent 请求/响应类型 =====

// CreateAgentRequest 创建 Agent 请求
type CreateAgentRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Avatar      string                 `json:"avatar"`
	Config      map[string]interface{} `json:"config"`
}

// UpdateAgentRequest 更新 Agent 请求
type UpdateAgentRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Avatar      *string                `json:"avatar"`
	Config      map[string]interface{} `json:"config"`
}

// AgentResponse Agent 响应
type AgentResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Avatar      string                 `json:"avatar"`
	Config      map[string]interface{} `json:"config"`
	TenantID    uint64                 `json:"tenant_id"`
	IsBuiltin   bool                   `json:"is_builtin"`
}

// ListAgentsResponse Agent 列表响应
type ListAgentsResponse struct {
	Agents []*AgentResponse `json:"agents"`
	Total  int64            `json:"total"`
}

// ListAgentsRequest Agent 列表请求
type ListAgentsRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// GetAgentRequest 获取 Agent 请求
type GetAgentRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetAgentResponse 获取 Agent 响应
type GetAgentResponse struct {
	Agent *AgentResponse `json:"agent"`
}

// CreateAgentResponse 创建 Agent 响应
type CreateAgentResponse struct {
	Agent *AgentResponse `json:"agent"`
}

// UpdateAgentResponse 更新 Agent 响应
type UpdateAgentResponse struct {
	Agent *AgentResponse `json:"agent"`
}

// DeleteAgentRequest 删除 Agent 请求
type DeleteAgentRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteAgentResponse 删除 Agent 响应
type DeleteAgentResponse struct {
	Success bool `json:"success"`
}

// BuiltinAgent 内置 Agent
type BuiltinAgent struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Type        string `json:"type"`
}

// CopyAgentRequest 复制 Agent 请求
type CopyAgentRequest struct {
	Id string `uri:"id" binding:"required"`
}

// CopyAgentResponse 复制 Agent 响应
type CopyAgentResponse struct {
	Agent *AgentResponse `json:"agent"`
}

// PlaceholdersResponse 占位符响应
type PlaceholdersResponse struct {
	All             []Placeholder `json:"all"`
	SystemPrompt    []Placeholder `json:"system_prompt"`
	ContextTemplate []Placeholder `json:"context_template"`
}

// Placeholder 占位符定义
type Placeholder struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
