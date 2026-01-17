// Package mcp 提供基于 Eino 官方 MCP 组件的客户端实现.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	mcptool "github.com/cloudwego/eino-ext/components/tool/mcp"
)

// EinoClient 基于 Eino 官方组件的 MCP 客户端.
type EinoClient struct {
	service   *model.MCPServiceM
	mcpClient client.MCPClient
	tools     []tool.BaseTool
	mu        sync.RWMutex
	connected bool
}

// NewEinoClient 创建基于 Eino 的 MCP 客户端.
func NewEinoClient(service *model.MCPServiceM) (*EinoClient, error) {
	if service == nil {
		return nil, fmt.Errorf("service config is required")
	}

	return &EinoClient{
		service: service,
	}, nil
}

// Connect 连接到 MCP 服务器.
func (c *EinoClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	url := ""
	if c.service.URL != nil {
		url = *c.service.URL
	}
	if url == "" {
		return fmt.Errorf("service URL is empty")
	}

	// 根据传输类型创建客户端
	switch c.service.TransportType {
	case string(model.MCPTransportSSE), string(model.MCPTransportHTTPStreamable):
		sseCli, err := client.NewSSEMCPClient(url)
		if err != nil {
			return fmt.Errorf("create SSE client: %w", err)
		}

		// 启动 SSE 客户端
		if err := sseCli.Start(ctx); err != nil {
			return fmt.Errorf("start SSE client: %w", err)
		}

		// 初始化连接
		initReq := mcp.InitializeRequest{}
		initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initReq.Params.ClientInfo = mcp.Implementation{
			Name:    "eino-show",
			Version: "1.0.0",
		}

		if _, err := sseCli.Initialize(ctx, initReq); err != nil {
			return fmt.Errorf("initialize MCP: %w", err)
		}

		c.mcpClient = sseCli

	case string(model.MCPTransportStdio):
		// Stdio 模式需要解析配置
		var stdioConfig model.MCPStdioConfig
		if c.service.StdioConfig != nil && *c.service.StdioConfig != "" {
			if err := json.Unmarshal([]byte(*c.service.StdioConfig), &stdioConfig); err != nil {
				return fmt.Errorf("parse stdio config: %w", err)
			}
		}
		if stdioConfig.Command == "" {
			return fmt.Errorf("stdio command is required")
		}

		// 解析环境变量
		var envVars []string
		if c.service.EnvVars != nil && *c.service.EnvVars != "" {
			var envMap model.MCPEnvVars
			if err := json.Unmarshal([]byte(*c.service.EnvVars), &envMap); err == nil {
				for k, v := range envMap {
					envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}

		stdioCli, err := client.NewStdioMCPClient(
			stdioConfig.Command,
			envVars,
			stdioConfig.Args...,
		)
		if err != nil {
			return fmt.Errorf("create Stdio client: %w", err)
		}

		// 初始化连接
		initReq := mcp.InitializeRequest{}
		initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initReq.Params.ClientInfo = mcp.Implementation{
			Name:    "eino-show",
			Version: "1.0.0",
		}

		if _, err := stdioCli.Initialize(ctx, initReq); err != nil {
			return fmt.Errorf("initialize MCP: %w", err)
		}

		c.mcpClient = stdioCli

	default:
		return fmt.Errorf("unsupported transport type: %s", c.service.TransportType)
	}

	c.connected = true

	return nil
}

// GetTools 获取 MCP 工具列表（作为 Eino tool.BaseTool）.
func (c *EinoClient) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	c.mu.RLock()
	if c.tools != nil {
		tools := c.tools
		c.mu.RUnlock()
		return tools, nil
	}
	c.mu.RUnlock()

	// 确保已连接
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 构建自定义 headers
	customHeaders := make(map[string]string)
	if c.service.Headers != nil && *c.service.Headers != "" {
		var headers model.MCPHeaders
		if err := json.Unmarshal([]byte(*c.service.Headers), &headers); err == nil {
			for k, v := range headers {
				customHeaders[k] = v
			}
		}
	}

	// 添加认证头
	if c.service.AuthConfig != nil && *c.service.AuthConfig != "" {
		var authConfig model.MCPAuthConfig
		if err := json.Unmarshal([]byte(*c.service.AuthConfig), &authConfig); err == nil {
			if authConfig.APIKey != "" {
				customHeaders["Authorization"] = "Bearer " + authConfig.APIKey
			}
			if authConfig.Token != "" {
				customHeaders["Authorization"] = authConfig.Token
			}
			for k, v := range authConfig.CustomHeaders {
				customHeaders[k] = v
			}
		}
	}

	// 使用 Eino 官方组件获取工具
	tools, err := mcptool.GetTools(ctx, &mcptool.Config{
		Cli:           c.mcpClient,
		CustomHeaders: customHeaders,
	})
	if err != nil {
		return nil, fmt.Errorf("get MCP tools: %w", err)
	}

	c.tools = tools
	return tools, nil
}

// ListTools 列出工具信息（兼容旧接口）.
func (c *EinoClient) ListTools(ctx context.Context) ([]*model.MCPTool, error) {
	tools, err := c.GetTools(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.MCPTool, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}

		// 转换参数 schema
		var inputSchema []byte
		if info.ParamsOneOf != nil {
			inputSchema, _ = json.Marshal(info.ParamsOneOf)
		}

		result = append(result, &model.MCPTool{
			Name:        info.Name,
			Description: info.Desc,
			InputSchema: inputSchema,
		})
	}

	return result, nil
}

// CallTool 调用工具（兼容旧接口）.
func (c *EinoClient) CallTool(ctx context.Context, toolName string, args map[string]any) (*model.MCPToolResult, error) {
	tools, err := c.GetTools(ctx)
	if err != nil {
		return nil, err
	}

	// 查找工具
	var targetTool tool.InvokableTool
	for _, t := range tools {
		info, _ := t.Info(ctx)
		if info != nil && info.Name == toolName {
			if invokable, ok := t.(tool.InvokableTool); ok {
				targetTool = invokable
				break
			}
		}
	}

	if targetTool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	// 序列化参数
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal arguments: %w", err)
	}

	// 调用工具
	result, err := targetTool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return nil, err
	}

	return &model.MCPToolResult{
		Output: result,
	}, nil
}

// IsConnected 检查连接状态.
func (c *EinoClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Disconnect 断开连接.
func (c *EinoClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.mcpClient != nil {
		// mark3labs/mcp-go 的客户端没有显式的 Close 方法
		// 但我们可以清理状态
		c.mcpClient = nil
	}
	c.connected = false
	c.tools = nil

	return nil
}

// GetService 获取关联的服务配置.
func (c *EinoClient) GetService() *model.MCPServiceM {
	return c.service
}

// Initialize 初始化连接（实现 Client 接口）.
func (c *EinoClient) Initialize(ctx context.Context) (*InitializeResult, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return &InitializeResult{
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		ServerInfo: ServerInfo{
			Name:    c.service.Name,
			Version: "1.0.0",
		},
	}, nil
}

// ListResources 列出资源（实现 Client 接口）.
func (c *EinoClient) ListResources(ctx context.Context) ([]*model.MCPResource, error) {
	// MCP 资源功能暂不实现
	return []*model.MCPResource{}, nil
}

// 确保 EinoClient 实现了 Client 接口.
var _ Client = (*EinoClient)(nil)
