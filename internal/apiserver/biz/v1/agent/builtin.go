// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"
	"fmt"

	agentpkg "github.com/ashwinyue/eino-show/internal/pkg/agent"
	chatagent "github.com/ashwinyue/eino-show/internal/pkg/agent/chat"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/react"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// BuiltinAgent 内置 Agent 类型.
type BuiltinAgent struct {
	ID          string
	Type        string
	Name        string
	Description string
	Create      func(ctx context.Context, factory *agentpkg.Factory) (interface{}, error)
}

// BuiltinAgents 内置 Agent 注册表.
//
// IMPORTANT: Agent Name and Description here are for UI display only (Chinese).
// The actual prompts sent to LLM must be in English (see SystemPrompt).
var BuiltinAgents = []*BuiltinAgent{
	{
		ID:          "quick-answer",
		Type:        "chat",
		Name:        "Quick Answer",
		Description: "Directly answer user questions without using tools. Suitable for simple knowledge Q&A.",
		Create: func(ctx context.Context, factory *agentpkg.Factory) (interface{}, error) {
			cfg := &chatagent.Config{
				SystemPrompt: "You are an intelligent assistant. Answer user questions concisely and accurately.",
			}
			return factory.CreateChatAgent(ctx, cfg)
		},
	},
	{
		ID:          "smart-reasoning",
		Type:        "react",
		Name:        "Smart Reasoning",
		Description: "Use knowledge base search and tools for reasoning and analysis. Suitable for complex problems.",
		Create: func(ctx context.Context, factory *agentpkg.Factory) (interface{}, error) {
			cfg := &react.Config{
				MaxIterations: 12,
			}
			return factory.CreateReactAgent(ctx, cfg)
		},
	},
}

// GetBuiltinAgent 获取内置 Agent.
func GetBuiltinAgent(id string) *BuiltinAgent {
	for _, a := range BuiltinAgents {
		if a.ID == id {
			return a
		}
	}
	return nil
}

// CreateBuiltinAgent 创建内置 Agent 实例.
func CreateBuiltinAgent(ctx context.Context, factory *agentpkg.Factory, id string) (interface{}, error) {
	agent := GetBuiltinAgent(id)
	if agent == nil {
		return nil, fmt.Errorf("unknown builtin agent: %s", id)
	}
	return agent.Create(ctx, factory)
}

// ExecuteBuiltinChatAgent 执行内置 Chat Agent.
func ExecuteBuiltinChatAgent(ctx context.Context, agent *chatagent.Agent, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return agent.Generate(ctx, messages, opts...)
}

// ExecuteBuiltinChatAgentStream 执行内置 Chat Agent (流式).
func ExecuteBuiltinChatAgentStream(ctx context.Context, agent *chatagent.Agent, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return agent.StreamGenerate(ctx, messages, opts...)
}

// CreateAgentFactoryConfig 创建 Agent 工厂配置（从环境变量）.
func CreateAgentFactoryConfig() *agentpkg.FactoryConfig {
	return &agentpkg.FactoryConfig{
		ChatModelConfig: agentmodel.DefaultConfig(),
		EmbeddingConfig: agentmodel.DefaultEmbeddingConfig(),
	}
}
