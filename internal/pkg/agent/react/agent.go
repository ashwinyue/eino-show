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
	"github.com/cloudwego/eino/schema"
)

// Config ReAct Agent 配置.
type Config struct {
	// ChatModel 对话模型
	ChatModel model.ToolCallingChatModel

	// Tools 可用工具列表
	Tools []tool.InvokableTool

	// SystemPrompt 系统提示词（可选）
	SystemPrompt string

	// MaxIterations 最大迭代次数（默认 12）
	MaxIterations int

	// Temperature 温度参数（可选）
	Temperature *float32

	// TopP Top-P 采样参数（可选）
	TopP *float32
}

// Agent ReAct Agent 包装器，支持系统提示词.
type Agent struct {
	agent        *einoReact.Agent
	systemPrompt string
}

// NewAgent 创建 ReAct Agent 并返回包装器.
func NewAgent(ctx context.Context, cfg *Config) (*Agent, error) {
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

	// 创建 Eino Agent
	einoAgent, err := einoReact.NewAgent(ctx, agentCfg)
	if err != nil {
		return nil, err
	}

	return &Agent{
		agent:        einoAgent,
		systemPrompt: cfg.SystemPrompt,
	}, nil
}

// NewSimpleAgent 创建简单的 ReAct Agent（使用默认配置）.
func NewSimpleAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools ...tool.InvokableTool) (*Agent, error) {
	return NewAgent(ctx, &Config{
		ChatModel:     chatModel,
		Tools:         tools,
		MaxIterations: 12,
	})
}

// Generate 生成回复（非流式）.
func (a *Agent) Generate(ctx context.Context, messages []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
	// 添加系统提示词
	messages = a.prependSystemPrompt(messages)

	return a.agent.Generate(ctx, messages, opts...)
}

// Stream 返回流式生成器.
func (a *Agent) Stream(ctx context.Context, messages []*schema.Message, opts ...agent.AgentOption) (*schema.StreamReader[*schema.Message], error) {
	// 添加系统提示词
	messages = a.prependSystemPrompt(messages)

	return a.agent.Stream(ctx, messages, opts...)
}

// prependSystemPrompt 预置系统提示词.
func (a *Agent) prependSystemPrompt(messages []*schema.Message) []*schema.Message {
	if a.systemPrompt == "" {
		return messages
	}

	// 检查是否已有系统消息
	for _, msg := range messages {
		if msg.Role == schema.System {
			return messages
		}
	}

	// 在开头添加系统消息
	result := make([]*schema.Message, 0, len(messages)+1)
	result = append(result, schema.SystemMessage(a.systemPrompt))
	result = append(result, messages...)
	return result
}

// WithTools 创建工具配置选项.
func WithTools(ctx context.Context, tools ...tool.InvokableTool) ([]agent.AgentOption, error) {
	baseTools := make([]tool.BaseTool, len(tools))
	for i, t := range tools {
		baseTools[i] = t
	}
	return einoReact.WithTools(ctx, baseTools...)
}

// WithMessageFuture 返回一个 agent option 和 MessageFuture 接口.
// 用于获取 Agent 执行过程中的中间消息（thinking、tool calls 等）.
func WithMessageFuture() (agent.AgentOption, einoReact.MessageFuture) {
	return einoReact.WithMessageFuture()
}
