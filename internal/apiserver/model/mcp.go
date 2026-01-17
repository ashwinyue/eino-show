// Package model 提供 MCP 服务相关的扩展类型和方法.
package model

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
)

// MCPTransportType MCP 传输类型.
type MCPTransportType string

const (
	// MCPTransportSSE Server-Sent Events 传输
	MCPTransportSSE MCPTransportType = "sse"
	// MCPTransportHTTPStreamable HTTP 流式传输
	MCPTransportHTTPStreamable MCPTransportType = "http-streamable"
	// MCPTransportStdio 标准输入输出传输
	MCPTransportStdio MCPTransportType = "stdio"
)

// MCPHeaders HTTP 请求头.
type MCPHeaders map[string]string

// MCPAuthConfig MCP 认证配置.
type MCPAuthConfig struct {
	APIKey        string            `json:"api_key,omitempty"`
	Token         string            `json:"token,omitempty"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// MCPAdvancedConfig MCP 高级配置.
type MCPAdvancedConfig struct {
	Timeout    int `json:"timeout"`     // 超时时间（秒），默认: 30
	RetryCount int `json:"retry_count"` // 重试次数，默认: 3
	RetryDelay int `json:"retry_delay"` // 重试延迟（秒），默认: 1
}

// MCPStdioConfig Stdio 传输配置.
type MCPStdioConfig struct {
	Command string   `json:"command"` // 命令: "uvx" 或 "npx"
	Args    []string `json:"args"`    // 命令参数数组
}

// MCPEnvVars 环境变量.
type MCPEnvVars map[string]string

// MCPTool MCP 工具定义.
type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"` // JSON Schema 参数定义
}

// MCPResource MCP 资源定义.
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// MCPTestResult MCP 测试结果.
type MCPTestResult struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message,omitempty"`
	Tools     []*MCPTool    `json:"tools,omitempty"`
	Resources []*MCPResource `json:"resources,omitempty"`
}

// MCPToolResult MCP 工具执行结果.
type MCPToolResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// BeforeCreate GORM 创建前钩子扩展.
func BeforeCreateMCPService(m *MCPServiceM) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// Value 实现 driver.Valuer 接口 (MCPHeaders).
func (h MCPHeaders) Value() (driver.Value, error) {
	if h == nil {
		return nil, nil
	}
	return json.Marshal(h)
}

// Scan 实现 sql.Scanner 接口 (MCPHeaders).
func (h *MCPHeaders) Scan(value interface{}) error {
	if value == nil {
		*h = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, h)
}

// Value 实现 driver.Valuer 接口 (MCPAuthConfig).
func (c *MCPAuthConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口 (MCPAuthConfig).
func (c *MCPAuthConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer 接口 (MCPAdvancedConfig).
func (c *MCPAdvancedConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口 (MCPAdvancedConfig).
func (c *MCPAdvancedConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer 接口 (MCPStdioConfig).
func (c *MCPStdioConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口 (MCPStdioConfig).
func (c *MCPStdioConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// Value 实现 driver.Valuer 接口 (MCPEnvVars).
func (e MCPEnvVars) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return json.Marshal(e)
}

// Scan 实现 sql.Scanner 接口 (MCPEnvVars).
func (e *MCPEnvVars) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, e)
}

// GetDefaultAdvancedConfig 返回默认高级配置.
func GetDefaultAdvancedConfig() *MCPAdvancedConfig {
	return &MCPAdvancedConfig{
		Timeout:    30,
		RetryCount: 3,
		RetryDelay: 1,
	}
}

// MaskSensitiveData 隐藏敏感信息用于显示.
func MaskSensitiveDataMCPService(m *MCPServiceM) {
	// 解析 JSON 字段进行掩码处理
	if m.Headers != nil {
		var headers MCPHeaders
		if err := json.Unmarshal([]byte(*m.Headers), &headers); err == nil {
			for key := range headers {
				if key == "api_key" || key == "authorization" || key == "token" {
					headers[key] = maskString(headers[key])
				}
			}
			if b, err := json.Marshal(headers); err == nil {
				s := string(b)
				m.Headers = &s
			}
		}
	}
	if m.AuthConfig != nil {
		var config MCPAuthConfig
		if err := json.Unmarshal([]byte(*m.AuthConfig), &config); err == nil {
			if config.APIKey != "" {
				config.APIKey = maskString(config.APIKey)
			}
			if config.Token != "" {
				config.Token = maskString(config.Token)
			}
			if b, err := json.Marshal(config); err == nil {
				s := string(b)
				m.AuthConfig = &s
			}
		}
	}
}

// maskString 隐藏字符串，只显示前 4 和后 4 个字符.
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
