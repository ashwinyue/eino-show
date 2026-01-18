// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== AgentStep 类型定义（对齐 WeKnora）=====

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	Success bool                   `json:"success"`
	Output  string                 `json:"output"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// ToolCall 工具调用（对齐 WeKnora）
type ToolCall struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Args       string          `json:"args"`
	Result     *ToolCallResult `json:"result,omitempty"`
	Reflection string          `json:"reflection,omitempty"`
	Duration   int64           `json:"duration"`
}

// AgentStep 代表 ReAct 循环的一次迭代（对齐 WeKnora）
type AgentStep struct {
	Iteration int        `json:"iteration"`
	Thought   string     `json:"thought"`
	ToolCalls []ToolCall `json:"tool_calls"`
	Timestamp time.Time  `json:"timestamp"`
}

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

// ListAgentsRequest Agent 列表请求
type ListAgentsRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ListAgentsResponse Agent 列表响应（对齐 WeKnora 格式）
type ListAgentsResponse struct {
	Success bool             `json:"success"`
	Data    []*AgentResponse `json:"data"`
}

// GetAgentRequest 获取 Agent 请求
type GetAgentRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetAgentResponse 获取 Agent 响应（对齐 WeKnora 格式）
type GetAgentResponse struct {
	Success bool           `json:"success"`
	Data    *AgentResponse `json:"data"`
}

// CreateAgentResponse 创建 Agent 响应（对齐 WeKnora 格式）
type CreateAgentResponse struct {
	Success bool           `json:"success"`
	Data    *AgentResponse `json:"data"`
}

// UpdateAgentResponse 更新 Agent 响应（对齐 WeKnora 格式）
type UpdateAgentResponse struct {
	Success bool           `json:"success"`
	Data    *AgentResponse `json:"data"`
}

// DeleteAgentRequest 删除 Agent 请求
type DeleteAgentRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteAgentResponse 删除 Agent 响应（对齐 WeKnora 格式）
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

// CopyAgentResponse 复制 Agent 响应（对齐 WeKnora 格式）
type CopyAgentResponse struct {
	Success bool           `json:"success"`
	Data    *AgentResponse `json:"data"`
}

// PlaceholdersResponse 占位符响应（对齐 WeKnora 格式）
type PlaceholdersResponse struct {
	Success bool              `json:"success"`
	Data    *PlaceholdersData `json:"data"`
}

// PlaceholdersData 占位符数据
type PlaceholdersData struct {
	All                 []Placeholder `json:"all"`
	SystemPrompt        []Placeholder `json:"system_prompt"`
	AgentSystemPrompt   []Placeholder `json:"agent_system_prompt"`
	ContextTemplate     []Placeholder `json:"context_template"`
	RewriteSystemPrompt []Placeholder `json:"rewrite_system_prompt"`
	RewritePrompt       []Placeholder `json:"rewrite_prompt"`
	FallbackPrompt      []Placeholder `json:"fallback_prompt"`
}

// Placeholder 占位符定义
type Placeholder struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}
