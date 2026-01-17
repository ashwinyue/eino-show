// Package supervisor 提供 Supervisor Agent 实现.
package supervisor

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// SupervisorAgentConfig Supervisor Agent 配置.
type SupervisorAgentConfig struct {
	Name         string
	ChatModel    model.ToolCallingChatModel
	SystemPrompt string
	SubAgents    []*SubAgentConfig
}

// supervisorAgent 实现 adk.Agent 接口.
type supervisorAgent struct {
	name         string
	description  string
	chatModel    model.ToolCallingChatModel
	systemPrompt string
	subAgents    []*SubAgentConfig
	tools        []tool.BaseTool
}

// NewSupervisorAgent 创建 Supervisor Agent.
func NewSupervisorAgent(ctx context.Context, cfg *SupervisorAgentConfig) adk.Agent {
	// 构建子 Agent 描述
	var subAgentDescs []string
	for _, sub := range cfg.SubAgents {
		subAgentDescs = append(subAgentDescs, fmt.Sprintf("- %s: %s", sub.Name, sub.Description))
	}

	// Default system prompt
	systemPrompt := cfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf(`You are an intelligent Supervisor responsible for delegating user requests to appropriate specialist agents.

Available specialist agents:
%s

Your responsibilities:
1. Analyze user requests and determine which specialist should handle them
2. Use the transfer_to_agent tool to delegate tasks to the appropriate specialist
3. If multiple specialists are needed, delegate tasks sequentially
4. Summarize the results from specialists and provide a final response`, strings.Join(subAgentDescs, "\n"))
	}

	// 为每个子 Agent 创建工具
	var tools []tool.BaseTool
	for _, sub := range cfg.SubAgents {
		tools = append(tools, adk.NewAgentTool(ctx, sub.Agent))
	}

	return &supervisorAgent{
		name:         cfg.Name,
		description:  "Supervisor agent that coordinates multiple specialist agents",
		chatModel:    cfg.ChatModel,
		systemPrompt: systemPrompt,
		subAgents:    cfg.SubAgents,
		tools:        tools,
	}
}

// Name 返回 Agent 名称.
func (a *supervisorAgent) Name(ctx context.Context) string {
	return a.name
}

// Description 返回 Agent 描述.
func (a *supervisorAgent) Description(ctx context.Context) string {
	return a.description
}

// Run 运行 Agent.
func (a *supervisorAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer gen.Close()

		// 构建消息
		messages := make([]*schema.Message, 0, len(input.Messages)+1)
		messages = append(messages, schema.SystemMessage(a.systemPrompt))
		for _, msg := range input.Messages {
			messages = append(messages, msg)
		}

		// 调用模型
		if input.EnableStreaming {
			// 流式输出
			stream, err := a.chatModel.Stream(ctx, messages)
			if err != nil {
				gen.Send(&adk.AgentEvent{
					Err: err,
				})
				return
			}

			gen.Send(adk.EventFromMessage(nil, stream, schema.Assistant, ""))
		} else {
			// 非流式输出
			resp, err := a.chatModel.Generate(ctx, messages)
			if err != nil {
				gen.Send(&adk.AgentEvent{
					Err: err,
				})
				return
			}

			gen.Send(adk.EventFromMessage(resp, nil, schema.Assistant, ""))
		}
	}()

	return iter
}
