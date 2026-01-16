// Package react 提供 ReAct Agent 封装，基于 Eino 内置实现.
package react

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	einoReact "github.com/cloudwego/eino/flow/agent/react"
)

// Config ReAct Agent 配置.
type Config struct {
	// ChatModel 对话模型
	ChatModel model.ToolCallingChatModel

	// Tools 可用工具列表
	Tools []tool.InvokableTool

	// MaxIterations 最大迭代次数（默认 12）
	MaxIterations int

	// Temperature 温度参数（可选）
	Temperature *float32

	// TopP Top-P 采样参数（可选）
	TopP *float32
}

// NewAgent 创建 ReAct Agent.
// 基于 Eino 内置的 react.Agent 实现.
func NewAgent(ctx context.Context, cfg *Config) (*einoReact.Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}

	// 构建工具列表
	var tools []tool.BaseTool
	if cfg.Tools != nil {
		tools = make([]tool.BaseTool, len(cfg.Tools))
		for i, t := range cfg.Tools {
			tools[i] = t
		}
	}

	// 设置默认最大迭代次数
	maxIterations := cfg.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 12
	}

	// 构建配置
	agentCfg := &einoReact.AgentConfig{
		ToolCallingModel: cfg.ChatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
		},
		MaxStep: maxIterations,
	}

	// 创建 Agent
	return einoReact.NewAgent(ctx, agentCfg)
}

// NewSimpleAgent 创建简单的 ReAct Agent（使用默认配置）.
func NewSimpleAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools ...tool.InvokableTool) (*einoReact.Agent, error) {
	return NewAgent(ctx, &Config{
		ChatModel:     chatModel,
		Tools:         tools,
		MaxIterations: 12,
	})
}

// WithTools 创建工具配置选项.
func WithTools(ctx context.Context, tools ...tool.InvokableTool) ([]agent.AgentOption, error) {
	baseTools := make([]tool.BaseTool, len(tools))
	for i, t := range tools {
		baseTools[i] = t
	}
	return einoReact.WithTools(ctx, baseTools...)
}
