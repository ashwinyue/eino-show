// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package agent 提供根据配置创建 Agent 的工厂方法.
package agent

import (
	"context"
	"fmt"

	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	chatagent "github.com/ashwinyue/eino-show/internal/pkg/agent/chat"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/react"
	agenttool "github.com/ashwinyue/eino-show/internal/pkg/agent/tool"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	einoReact "github.com/cloudwego/eino/flow/agent/react"
)

// FactoryConfig Agent 工厂配置.
type FactoryConfig struct {
	// Store 数据存储接口
	Store store.IStore

	// ChatModelConfig LLM 模型配置
	ChatModelConfig *agentmodel.Config

	// EmbeddingConfig 向量模型配置
	EmbeddingConfig *agentmodel.EmbeddingConfig
}

// Factory Agent 工厂，负责创建各种类型的 Agent.
type Factory struct {
	cfg   *FactoryConfig
	tcm   model.ToolCallingChatModel
	cm    model.ChatModel
	embed embedding.Embedder
	tools []*agenttool.Registry
}

// NewFactory 创建 Agent 工厂.
func NewFactory(ctx context.Context, cfg *FactoryConfig) (*Factory, error) {
	if cfg == nil {
		return nil, fmt.Errorf("factory config is nil")
	}
	if cfg.Store == nil {
		return nil, fmt.Errorf("store is required")
	}
	if cfg.ChatModelConfig == nil {
		return nil, fmt.Errorf("chat model config is required")
	}

	// 创建 ChatModel
	cm, err := agentmodel.NewChatModel(ctx, cfg.ChatModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// 确保 ChatModel 实现了 ToolCallingChatModel
	tcm, ok := cm.(model.ToolCallingChatModel)
	if !ok {
		return nil, fmt.Errorf("chat model must implement ToolCallingChatModel")
	}

	f := &Factory{
		cfg:   cfg,
		tcm:   tcm,
		cm:    cm,
		tools: []*agenttool.Registry{agenttool.NewRegistry()},
	}

	// 如果配置了 Embedding，创建 Embedder
	if cfg.EmbeddingConfig != nil {
		embed, err := agentmodel.NewEmbedder(ctx, cfg.EmbeddingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedder: %w", err)
		}
		f.embed = embed

		// 注册默认工具
		if err := f.registerDefaultTools(ctx); err != nil {
			return nil, fmt.Errorf("failed to register default tools: %w", err)
		}
	}

	return f, nil
}

// registerDefaultTools 注册默认工具.
func (f *Factory) registerDefaultTools(ctx context.Context) error {
	registry := f.tools[0]

	// 注册 knowledge_search 工具
	if f.embed != nil {
		knowledgeSearch := agenttool.NewKnowledgeSearch(f.cfg.Store, f.embed)
		registry.Register(knowledgeSearch)
	}

	// 注册 grep_chunks 工具
	grepChunks := agenttool.NewGrepChunks(f.cfg.Store)
	registry.Register(grepChunks)

	// 注册 web_search 工具（占位符）
	webSearch, err := agenttool.NewWebSearch(ctx)
	if err != nil {
		return err
	}
	registry.Register(webSearch)

	return nil
}

// CreateReactAgent 创建 ReAct Agent.
func (f *Factory) CreateReactAgent(ctx context.Context, cfg *react.Config) (*einoReact.Agent, error) {
	if cfg == nil {
		cfg = &react.Config{}
	}

	// 设置默认 ChatModel
	if cfg.ChatModel == nil {
		cfg.ChatModel = f.tcm
	}

	// 如果没有指定工具，使用注册表中的所有工具
	if cfg.Tools == nil && len(f.tools) > 0 {
		allTools := f.tools[0].List()
		cfg.Tools = make([]einotool.InvokableTool, len(allTools))
		for i, t := range allTools {
			cfg.Tools[i] = t
		}
	}

	return react.NewAgent(ctx, cfg)
}

// CreateChatAgent 创建纯对话 Agent.
func (f *Factory) CreateChatAgent(ctx context.Context, cfg *chatagent.Config) (*chatagent.Agent, error) {
	if cfg == nil {
		cfg = &chatagent.Config{}
	}

	// 设置默认 ChatModel
	if cfg.ChatModel == nil {
		cfg.ChatModel = f.cm
	}

	return chatagent.NewAgent(ctx, cfg)
}

// CreateAgent 根据类型创建 Agent.
// agentType: "react" 或 "chat"
func (f *Factory) CreateAgent(ctx context.Context, agentType string, config interface{}) (interface{}, error) {
	switch agentType {
	case "react":
		cfg, ok := config.(*react.Config)
		if !ok {
			return f.CreateReactAgent(ctx, nil)
		}
		return f.CreateReactAgent(ctx, cfg)

	case "chat":
		cfg, ok := config.(*chatagent.Config)
		if !ok {
			return f.CreateChatAgent(ctx, nil)
		}
		return f.CreateChatAgent(ctx, cfg)

	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

// GetToolRegistry 获取工具注册表.
func (f *Factory) GetToolRegistry() *agenttool.Registry {
	if len(f.tools) > 0 {
		return f.tools[0]
	}
	return nil
}

// RegisterTool 注册自定义工具.
func (f *Factory) RegisterTool(t einotool.InvokableTool) {
	if len(f.tools) > 0 {
		f.tools[0].Register(t)
	}
}

// GetTools 根据名称列表获取工具.
func (f *Factory) GetTools(names []string) []einotool.InvokableTool {
	if len(f.tools) == 0 {
		return nil
	}
	return f.tools[0].GetToolsByNames(names)
}
