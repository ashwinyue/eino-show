// Package tool 提供 MCP 工具包装器.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	mcpclient "github.com/ashwinyue/eino-show/internal/pkg/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

// MCPClient 定义 MCP 客户端接口.
// TODO: 实现完整的 MCP 协议客户端
type MCPClient interface {
	// ListTools 列出服务提供的工具
	ListTools(ctx context.Context) ([]*model.MCPTool, error)

	// CallTool 调用工具
	CallTool(ctx context.Context, toolName string, args map[string]any) (*model.MCPToolResult, error)

	// IsConnected 检查连接状态
	IsConnected() bool

	// Disconnect 断开连接
	Disconnect() error
}

// MCPToolConfig MCP 工具配置.
type MCPToolConfig struct {
	// Service MCP 服务配置
	Service *model.MCPServiceM

	// MCPTool MCP 工具定义
	MCPTool *model.MCPTool

	// Client MCP 客户端
	Client MCPClient
}

// mcpTool MCP 工具包装器，实现 Eino InvokableTool 接口.
type mcpTool struct {
	config *MCPToolConfig
}

// NewMCPTool 创建 MCP 工具.
func NewMCPTool(cfg *MCPToolConfig) tool.InvokableTool {
	return &mcpTool{config: cfg}
}

// Info 返回工具信息.
func (t *mcpTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	serviceName := sanitizeName(t.config.Service.Name)
	toolName := sanitizeName(t.config.MCPTool.Name)

	info := &schema.ToolInfo{
		Name: fmt.Sprintf("mcp.%s.%s", serviceName, toolName),
		Desc: t.formatDescription(),
	}

	// 如果有 JSON Schema，解析并设置参数
	if len(t.config.MCPTool.InputSchema) > 0 {
		var jsSchema jsonschema.Schema
		if err := json.Unmarshal(t.config.MCPTool.InputSchema, &jsSchema); err == nil {
			info.ParamsOneOf = schema.NewParamsOneOfByJSONSchema(&jsSchema)
		}
	}

	return info, nil
}

// InvokableRun 执行工具.
func (t *mcpTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var args map[string]any
	if argumentsInJSON != "" && argumentsInJSON != "{}" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	// 调用 MCP 工具
	result, err := t.config.Client.CallTool(ctx, t.config.MCPTool.Name, args)
	if err != nil {
		return "", fmt.Errorf("MCP tool call failed: %w", err)
	}

	return result.Output, nil
}

// formatDescription 格式化工具描述.
func (t *mcpTool) formatDescription() string {
	serviceDesc := fmt.Sprintf("[MCP: %s] ", t.config.Service.Name)
	if t.config.MCPTool.Description != "" {
		return serviceDesc + t.config.MCPTool.Description
	}
	return serviceDesc + t.config.MCPTool.Name
}

// sanitizeName 清理名称为有效标识符.
func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	var result strings.Builder
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// RegisterMCPTools 注册 MCP 工具到注册表.
// 从给定的 MCP 客户端映射中获取所有可用工具并注册.
func RegisterMCPTools(
	ctx context.Context,
	registry *Registry,
	clients map[string]mcpclient.Client,
) error {
	// 遍历所有 MCP 客户端
	for _, client := range clients {
		// 列出该服务提供的工具
		tools, err := client.ListTools(ctx)
		if err != nil {
			// 跳过无法列出工具的服务
			continue
		}

		// 为每个工具创建 Eino 工具包装器并注册
		for _, mcpTool := range tools {
			cfg := &MCPToolConfig{
				Service: client.GetService(),
				MCPTool: mcpTool,
				Client:  client,
			}
			registry.Register(NewMCPTool(cfg))
		}
	}

	return nil
}

// MCPManager MCP 管理器接口.
type MCPManager interface {
	// GetOrCreateClient 获取或创建客户端
	GetOrCreateClient(service *model.MCPServiceM) (mcpclient.Client, error)

	// GetEinoClient 获取 Eino 客户端（支持直接获取工具）
	GetEinoClient(service *model.MCPServiceM) (*mcpclient.EinoClient, error)

	// GetToolsFromServices 从多个服务获取所有工具
	GetToolsFromServices(ctx context.Context, services []*model.MCPServiceM) ([]tool.BaseTool, error)

	// CloseClient 关闭客户端
	CloseClient(serviceID string) error

	// CloseAll 关闭所有客户端
	CloseAll()
}

// NewMCPManager 创建 MCP 管理器.
func NewMCPManager() MCPManager {
	return &mcpManagerImpl{
		clients:     make(map[string]mcpclient.Client),
		einoClients: make(map[string]*mcpclient.EinoClient),
	}
}

// mcpManagerImpl MCP 管理器实现.
type mcpManagerImpl struct {
	clients     map[string]mcpclient.Client
	einoClients map[string]*mcpclient.EinoClient
	mu          sync.RWMutex
}

func (m *mcpManagerImpl) GetOrCreateClient(service *model.MCPServiceM) (mcpclient.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已存在客户端，直接返回
	if client, ok := m.clients[service.ID]; ok {
		return client, nil
	}

	// 优先使用 EinoClient
	einoClient, err := mcpclient.NewEinoClient(service)
	if err != nil {
		return nil, err
	}

	m.clients[service.ID] = einoClient
	m.einoClients[service.ID] = einoClient
	return einoClient, nil
}

func (m *mcpManagerImpl) GetEinoClient(service *model.MCPServiceM) (*mcpclient.EinoClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已存在客户端，直接返回
	if client, ok := m.einoClients[service.ID]; ok {
		return client, nil
	}

	// 创建新客户端
	einoClient, err := mcpclient.NewEinoClient(service)
	if err != nil {
		return nil, err
	}

	m.clients[service.ID] = einoClient
	m.einoClients[service.ID] = einoClient
	return einoClient, nil
}

func (m *mcpManagerImpl) GetToolsFromServices(ctx context.Context, services []*model.MCPServiceM) ([]tool.BaseTool, error) {
	var allTools []tool.BaseTool

	for _, svc := range services {
		client, err := m.GetEinoClient(svc)
		if err != nil {
			continue // 跳过无法创建客户端的服务
		}

		tools, err := client.GetTools(ctx)
		if err != nil {
			continue // 跳过无法获取工具的服务
		}

		allTools = append(allTools, tools...)
	}

	return allTools, nil
}

func (m *mcpManagerImpl) CloseClient(serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[serviceID]; ok {
		client.Disconnect()
		delete(m.clients, serviceID)
		delete(m.einoClients, serviceID)
	}
	return nil
}

func (m *mcpManagerImpl) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, client := range m.clients {
		client.Disconnect()
		delete(m.clients, id)
		delete(m.einoClients, id)
	}
}
