// Package supervisor 提供基于 Eino ADK 的 Supervisor 多 Agent 模式实现.
package supervisor

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// Config Supervisor 配置.
type Config struct {
	// Name Supervisor 名称
	Name string

	// ChatModel Supervisor 使用的聊天模型
	ChatModel model.ToolCallingChatModel

	// SystemPrompt Supervisor 系统提示词
	SystemPrompt string

	// SubAgents 子 Agent 列表
	SubAgents []*SubAgentConfig
}

// SubAgentConfig 子 Agent 配置.
type SubAgentConfig struct {
	// Name 子 Agent 名称（唯一标识）
	Name string

	// Description 子 Agent 描述（用于 Supervisor 选择）
	Description string

	// Agent Eino Agent 实例
	Agent adk.Agent
}

// Supervisor 多 Agent 协调器.
type Supervisor struct {
	name      string
	agent     adk.ResumableAgent
	chatModel model.ToolCallingChatModel
}

// New 创建 Supervisor.
func New(ctx context.Context, cfg *Config) (*Supervisor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if len(cfg.SubAgents) == 0 {
		return nil, fmt.Errorf("at least one sub agent is required")
	}

	// 创建 Supervisor Agent
	supervisorAgent := NewSupervisorAgent(ctx, &SupervisorAgentConfig{
		Name:         cfg.Name,
		ChatModel:    cfg.ChatModel,
		SystemPrompt: cfg.SystemPrompt,
		SubAgents:    cfg.SubAgents,
	})

	// 构建子 Agent 列表
	subAgents := make([]adk.Agent, 0, len(cfg.SubAgents))
	for _, sub := range cfg.SubAgents {
		subAgents = append(subAgents, sub.Agent)
	}

	// 使用 Eino 官方 Supervisor 模式
	resumableAgent, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: supervisorAgent,
		SubAgents:  subAgents,
	})
	if err != nil {
		return nil, fmt.Errorf("create supervisor: %w", err)
	}

	return &Supervisor{
		name:      cfg.Name,
		agent:     resumableAgent,
		chatModel: cfg.ChatModel,
	}, nil
}

// Run 执行 Supervisor.
func (s *Supervisor) Run(ctx context.Context, input string) (*schema.Message, error) {
	// 构建输入消息
	messages := []adk.Message{
		schema.UserMessage(input),
	}

	// 运行 Agent
	iter := s.agent.Run(ctx, &adk.AgentInput{
		Messages: messages,
	})

	// 收集结果
	var lastMessage *schema.Message
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if msg, _, err := adk.GetMessage(event); err == nil {
			lastMessage = msg
		}
	}

	return lastMessage, nil
}

// Stream 流式执行 Supervisor.
func (s *Supervisor) Stream(ctx context.Context, input string) *adk.AsyncIterator[*adk.AgentEvent] {
	// 构建输入消息
	messages := []adk.Message{
		schema.UserMessage(input),
	}

	// 运行 Agent
	return s.agent.Run(ctx, &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: true,
	})
}

// Name 返回 Supervisor 名称.
func (s *Supervisor) Name() string {
	return s.name
}

// GetAgent 返回底层的 ResumableAgent.
func (s *Supervisor) GetAgent() adk.ResumableAgent {
	return s.agent
}
