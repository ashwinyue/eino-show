// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"
	"fmt"

	agentpkg "github.com/ashwinyue/eino-show/internal/pkg/agent"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/react"
	chatagent "github.com/ashwinyue/eino-show/internal/pkg/agent/chat"
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
var BuiltinAgents = []*BuiltinAgent{
	{
		ID:          "quick-answer",
		Type:        "chat",
		Name:        "快速回答",
		Description: "直接回答用户问题，不使用工具。适合简单的知识问答。",
		Create: func(ctx context.Context, factory *agentpkg.Factory) (interface{}, error) {
			cfg := &chatagent.Config{
				SystemPrompt: "你是一个智能助手，请用简洁、准确的语言回答用户的问题。",
			}
			return factory.CreateChatAgent(ctx, cfg)
		},
	},
	{
		ID:          "smart-reasoning",
		Type:        "react",
		Name:        "智能推理",
		Description: "使用知识库搜索和工具进行推理分析。适合复杂问题。",
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
		ChatModelConfig:  agentmodel.DefaultConfig(),
		EmbeddingConfig:  agentmodel.DefaultEmbeddingConfig(),
	}
}
