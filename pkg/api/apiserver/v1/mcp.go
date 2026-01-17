// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== MCP Service 请求/响应类型 =====

// CreateMCPServiceRequest 创建 MCP 服务请求
type CreateMCPServiceRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Type        string                 `json:"type" binding:"required"` // "stdio" or "http"
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
}

// CreateMCPServiceResponse 创建 MCP 服务响应
type CreateMCPServiceResponse struct {
	MCPService *MCPServiceResponse `json:"mcp_service"`
}

// GetMCPServiceRequest 获取 MCP 服务请求
type GetMCPServiceRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetMCPServiceResponse 获取 MCP 服务响应
type GetMCPServiceResponse struct {
	MCPService *MCPServiceResponse `json:"mcp_service"`
}

// ListMCPServicesRequest MCP 服务列表请求
type ListMCPServicesRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ListMCPServicesResponse MCP 服务列表响应
type ListMCPServicesResponse struct {
	MCPServices []*MCPServiceResponse `json:"mcp_services"`
	Total       int64                 `json:"total"`
}

// UpdateMCPServiceRequest 更新 MCP 服务请求
type UpdateMCPServiceRequest struct {
	Id          string                 `uri:"id" binding:"required"`
	Name        *string                `json:"name"`
	Type        *string                `json:"type"`
	Command     *string                `json:"command"`
	Args        []string               `json:"args"`
	URL         *string                `json:"url"`
	Description *string                `json:"description"`
	Config      map[string]interface{} `json:"config"`
}

// UpdateMCPServiceResponse 更新 MCP 服务响应
type UpdateMCPServiceResponse struct {
	MCPService *MCPServiceResponse `json:"mcp_service"`
}

// DeleteMCPServiceRequest 删除 MCP 服务请求
type DeleteMCPServiceRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteMCPServiceResponse 删除 MCP 服务响应
type DeleteMCPServiceResponse struct {
	Success bool `json:"success"`
}

// TestMCPServiceRequest 测试 MCP 服务请求
type TestMCPServiceRequest struct {
	Id string `uri:"id" binding:"required"`
}

// TestMCPServiceResponse 测试 MCP 服务响应
type TestMCPServiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetMCPServiceToolsRequest 获取 MCP 服务工具请求
type GetMCPServiceToolsRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetMCPServiceToolsResponse 获取 MCP 服务工具响应
type GetMCPServiceToolsResponse struct {
	Tools []*MCPToolResponse `json:"tools"`
}

// MCPServiceResponse MCP 服务响应
type MCPServiceResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	TenantID    uint64                 `json:"tenant_id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// MCPToolResponse MCP 工具响应
type MCPToolResponse struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}
