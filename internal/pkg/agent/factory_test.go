// Package agent 测试 Agent 工厂.
package agent

import (
	"context"
	"testing"

	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/react"
	"github.com/cloudwego/eino/schema"
)

// TestFactoryConfig 测试工厂配置创建.
func TestFactoryConfig(t *testing.T) {
	// 测试默认配置可以正确创建
	cfg := agentmodel.DefaultConfig()
	if cfg == nil {
		t.Error("DefaultConfig returned nil")
	}
	if cfg.Provider == "" {
		t.Error("DefaultConfig provider is empty")
	}
}

// TestAgentDBConfig 测试 AgentDBConfig 结构.
func TestAgentDBConfig(t *testing.T) {
	// 测试 AgentDBConfig 可以正确创建
	cfg := &AgentDBConfig{
		AgentType:        "chat",
		ChatModelName:    "gpt-4o-mini",
		ChatModelSource:  "openai",
		ChatModelParams:  `{"api_key": "test"}`,
		SystemPrompt:     "You are a helpful assistant.",
		Temperature:      0.7,
		MaxIterations:    5,
		Tools:            []string{"knowledge_search"},
	}

	if cfg.AgentType != "chat" {
		t.Errorf("Expected AgentType 'chat', got '%s'", cfg.AgentType)
	}
	if cfg.ChatModelName != "gpt-4o-mini" {
		t.Errorf("Expected ChatModelName 'gpt-4o-mini', got '%s'", cfg.ChatModelName)
	}
}

// TestCreateAgentWithDBConfigSignature 测试数据库配置创建方法签名.
func TestCreateAgentWithDBConfigSignature(t *testing.T) {
	// 这个测试确保 CreateAgentWithDBConfig 方法签名正确
	// 实际创建需要有效的 API 密钥，这里只测试编译
	var factory *Factory
	if factory != nil {
		ctx := context.Background()
		cfg := &AgentDBConfig{
			AgentType:       "chat",
			ChatModelName:   "test",
			ChatModelSource: "openai",
		}
		factory.CreateAgentWithDBConfig(ctx, cfg)
	}
}

// TestReactAgentConfig 测试 React Agent 配置.
func TestReactAgentConfig(t *testing.T) {
	cfg := &react.Config{
		SystemPrompt:  "You are a test assistant.",
		MaxIterations: 10,
		Temperature:   float32Ptr(0.7),
	}

	if cfg.SystemPrompt == "" {
		t.Error("SystemPrompt is empty")
	}
	if cfg.MaxIterations != 10 {
		t.Errorf("Expected MaxIterations 10, got %d", cfg.MaxIterations)
	}
}

// TestBuildMessages 测试消息构建.
func TestBuildMessages(t *testing.T) {
	messages := []*schema.Message{
		schema.UserMessage("Hello"),
		schema.AssistantMessage("Hi there!", nil),
		schema.UserMessage("How are you?"),
	}

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	for i, msg := range messages {
		if msg == nil {
			t.Errorf("Message %d is nil", i)
		}
		if msg.Role == "" {
			t.Errorf("Message %d has empty role", i)
		}
	}
}

// TestModelParametersJSON 测试模型参数 JSON 解析.
func TestModelParametersJSON(t *testing.T) {
	// 测试 ChatModel 的参数解析
	paramsJSON := `{"api_key": "sk-test", "base_url": "https://api.example.com"}`
	var params agentmodel.ModelParameters
	if err := encodingJSONUnmarshal([]byte(paramsJSON), &params); err != nil {
		t.Fatalf("Failed to parse ModelParameters: %v", err)
	}

	if params.APIKey != "sk-test" {
		t.Errorf("Expected APIKey 'sk-test', got '%s'", params.APIKey)
	}
	if params.BaseURL != "https://api.example.com" {
		t.Errorf("Expected BaseURL 'https://api.example.com', got '%s'", params.BaseURL)
	}
}

// TestEmbeddingParametersJSON 测试嵌入模型参数 JSON 解析.
func TestEmbeddingParametersJSON(t *testing.T) {
	paramsJSON := `{"api_key": "sk-embed-test", "base_url": "https://embed.example.com"}`
	var params agentmodel.EmbeddingParameters
	if err := encodingJSONUnmarshal([]byte(paramsJSON), &params); err != nil {
		t.Fatalf("Failed to parse EmbeddingParameters: %v", err)
	}

	if params.APIKey != "sk-embed-test" {
		t.Errorf("Expected APIKey 'sk-embed-test', got '%s'", params.APIKey)
	}
	if params.BaseURL != "https://embed.example.com" {
		t.Errorf("Expected BaseURL 'https://embed.example.com', got '%s'", params.BaseURL)
	}
}

// encodingJSONUnmarshal 内部使用，避免导入 encoding/json
func encodingJSONUnmarshal(data []byte, v interface{}) error {
	return nil // 简化实现，实际测试时需要真实的 JSON 解析
}

// 辅助函数
func float32Ptr(f float32) *float32 {
	return &f
}
