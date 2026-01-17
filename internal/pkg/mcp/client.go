// Package mcp 提供 MCP (Model Context Protocol) 客户端实现.
// 支持 SSE 和 HTTP 传输方式，实现 JSON-RPC 2.0 协议.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// JSONRPCRequest JSON-RPC 2.0 请求.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse JSON-RPC 2.0 响应.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError JSON-RPC 2.0 错误.
type JSONRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error 实现 error 接口.
func (e *JSONRPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (data: %s)", e.Code, e.Message, string(e.Data))
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// InitializeParams 初始化参数.
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

// ClientCapabilities 客户端能力.
type ClientCapabilities struct {
	Roots     *RootsCapability     `json:"roots,omitempty"`
	Sampling  *SamplingCapability  `json:"sampling,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// RootsCapability 根目录能力.
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability 采样能力.
type SamplingCapability struct{}

// ResourcesCapability 资源能力.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability 工具能力.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability 提示能力.
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ClientInfo 客户端信息.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult 初始化结果.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Meta            map[string]any     `json:"meta,omitempty"`
}

// ServerCapabilities 服务器能力.
type ServerCapabilities struct {
	Roots     *RootsCapability     `json:"roots,omitempty"`
	Sampling  *SamplingCapability  `json:"sampling,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
}

// LoggingCapability 日志能力.
type LoggingCapability struct{}

// ServerInfo 服务器信息.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListToolsResult 列出工具结果.
type ListToolsResult struct {
	Tools []MCPToolInfo `json:"tools"`
}

// MCPToolInfo MCP 工具信息.
type MCPToolInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// CallToolResult 调用工具结果.
type CallToolResult struct {
	Content []interface{}  `json:"content"`
	IsError bool           `json:"isError,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// ListResourcesResult 列出资源结果.
type ListResourcesResult struct {
	Resources []MCPResourceInfo `json:"resources"`
}

// MCPResourceInfo MCP 资源信息.
type MCPResourceInfo struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Client MCP 客户端接口.
type Client interface {
	// Initialize 初始化连接
	Initialize(ctx context.Context) (*InitializeResult, error)

	// ListTools 列出可用工具
	ListTools(ctx context.Context) ([]*model.MCPTool, error)

	// CallTool 调用工具
	CallTool(ctx context.Context, toolName string, args map[string]any) (*model.MCPToolResult, error)

	// ListResources 列出可用资源
	ListResources(ctx context.Context) ([]*model.MCPResource, error)

	// IsConnected 检查连接状态
	IsConnected() bool

	// Disconnect 断开连接
	Disconnect() error

	// GetService 获取关联的服务配置
	GetService() *model.MCPServiceM
}

// NewClient 创建 MCP 客户端.
// 使用基于 Eino 官方组件的 EinoClient 实现.
func NewClient(service *model.MCPServiceM) (Client, error) {
	return NewEinoClient(service)
}
