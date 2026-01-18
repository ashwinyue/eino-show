// Package enhanced provides ADK wrapper for EnhancedAgent to support multi-agent transfer.
//
// 架构说明：
//   - ADKAgent 实现 adk.Agent 接口，包装 EnhancedAgent
//   - 当前架构：直接调用 EnhancedAgent.Stream()，不使用 ADK Runner
//   - 如需完整 ADK 功能（Agent 转移、中断恢复），需重构使用 ADK 标准类型
//
// Reference: Eino ADK flow.go, chatmodel.go
package enhanced

import (
	"context"
	"io"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// ADKAgentConfig ADK Agent 包装配置.
type ADKAgentConfig struct {
	// Name Agent 名称
	Name string

	// Description Agent 描述
	Description string

	// EnhancedConfig 增强 Agent 配置
	EnhancedConfig *AgentConfig

	// SubAgents 子 Agent 列表
	SubAgents []adk.Agent
}

// ADKAgent 包装 EnhancedAgent 为 ADK Agent.
type ADKAgent struct {
	name        string
	description string
	agent       *Agent
	config      *ADKAgentConfig
}

// NewADKAgent 创建 ADK Agent 包装器.
func NewADKAgent(ctx context.Context, config *ADKAgentConfig) (*ADKAgent, error) {
	// 创建增强 Agent
	enhancedAgent, err := NewAgent(ctx, config.EnhancedConfig)
	if err != nil {
		return nil, err
	}

	return &ADKAgent{
		name:        config.Name,
		description: config.Description,
		agent:       enhancedAgent,
		config:      config,
	}, nil
}

// Name 返回 Agent 名称.
func (a *ADKAgent) Name(ctx context.Context) string {
	return a.name
}

// Description 返回 Agent 描述.
func (a *ADKAgent) Description(ctx context.Context) string {
	return a.description
}

// Run 执行 Agent (实现 adk.Agent 接口).
//
// 注意：当前实现直接调用 EnhancedAgent.Stream()，不使用 ADK Runner。
// 因此返回的事件流不包含 Action 事件（TransferToAgent, Interrupted, Exit）。
// 如需完整的 ADK 功能，需重构为使用 adk.ReActAgent 或 adk.ChatModelAgent。
func (a *ADKAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	iterator, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer generator.Close()

		// 转换输入消息
		messages := convertToSchemaMessages(input.Messages)

		// 执行流式生成
		streamReader, err := a.agent.Stream(ctx, messages)
		if err != nil {
			generator.Send(&adk.AgentEvent{
				Err: err,
			})
			return
		}

		// 读取流式输出并转换为 AgentEvent
		for {
			msg, err := streamReader.Recv()
			if err != nil {
				if err != io.EOF {
					generator.Send(&adk.AgentEvent{
						Err: err,
					})
				}
				break
			}

			// 发送消息事件（使用消息自身的 Role，确保 tool 消息不会被当作 assistant）
			generator.Send(adk.EventFromMessage(msg, nil, msg.Role, ""))
		}
	}()

	return iterator
}

// convertToSchemaMessages 转换 ADK Message 到 schema.Message.
func convertToSchemaMessages(adkMessages []adk.Message) []*schema.Message {
	messages := make([]*schema.Message, len(adkMessages))
	for i, m := range adkMessages {
		messages[i] = m
	}
	return messages
}

// SubAgentManager 子 Agent 管理器 (使用 ADK flow).
type SubAgentManager struct {
	mainAgent adk.Agent
	subAgents map[string]adk.Agent
	flowAgent adk.ResumableAgent
}

// NewSubAgentManager 创建子 Agent 管理器.
func NewSubAgentManager(ctx context.Context, mainAgent adk.Agent, subAgents []adk.Agent) (*SubAgentManager, error) {
	// 使用 ADK SetSubAgents 设置子 Agent
	flowAgent, err := adk.SetSubAgents(ctx, mainAgent, subAgents)
	if err != nil {
		return nil, err
	}

	// 构建子 Agent 映射
	subAgentMap := make(map[string]adk.Agent)
	for _, sa := range subAgents {
		subAgentMap[sa.Name(ctx)] = sa
	}

	return &SubAgentManager{
		mainAgent: mainAgent,
		subAgents: subAgentMap,
		flowAgent: flowAgent,
	}, nil
}

// Run 执行多 Agent 流程.
func (m *SubAgentManager) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return m.flowAgent.Run(ctx, input, options...)
}

// GetSubAgent 获取子 Agent.
func (m *SubAgentManager) GetSubAgent(name string) (adk.Agent, bool) {
	agent, ok := m.subAgents[name]
	return agent, ok
}

// BuiltinAgentInfo 内置 Agent 信息.
type BuiltinAgentInfo struct {
	ID          string
	Name        string
	Description string
}

// ListBuiltinAgents 列出内置 Agent.
func ListBuiltinAgents() []BuiltinAgentInfo {
	return []BuiltinAgentInfo{
		{
			ID:          "builtin-quick-answer",
			Name:        "快速问答",
			Description: "基于知识库的快速问答，适合简单的检索问题",
		},
		{
			ID:          "builtin-smart-reasoning",
			Name:        "智能推理",
			Description: "ReAct 推理框架，支持多步思考和工具调用",
		},
		{
			ID:          "builtin-data-analyst",
			Name:        "数据分析师",
			Description: "专业数据分析，支持 CSV/Excel 文件的 SQL 查询",
		},
	}
}

// CreateBuiltinAgent 创建内置 Agent.
func CreateBuiltinAgent(ctx context.Context, agentID string, config *AgentConfig) (adk.Agent, error) {
	// 根据 agentID 调整配置
	switch agentID {
	case "builtin-quick-answer":
		// 快速问答: 简单 RAG，不需要意图路由
		config.IntentRouter = nil
		config.MaxStep = 3
	case "builtin-smart-reasoning":
		// 智能推理: 完整增强 Agent
		config.MaxStep = 12
	case "builtin-data-analyst":
		// 数据分析: 需要特定工具
		config.MaxStep = 8
	}

	info := getBuiltinAgentInfo(agentID)
	return NewADKAgent(ctx, &ADKAgentConfig{
		Name:           info.Name,
		Description:    info.Description,
		EnhancedConfig: config,
	})
}

func getBuiltinAgentInfo(agentID string) BuiltinAgentInfo {
	for _, info := range ListBuiltinAgents() {
		if info.ID == agentID {
			return info
		}
	}
	return BuiltinAgentInfo{ID: agentID, Name: agentID}
}

// CreateMultiAgentFlow 创建多 Agent 流程.
func CreateMultiAgentFlow(ctx context.Context, mainAgentConfig *ADKAgentConfig, subAgentConfigs []*ADKAgentConfig) (*SubAgentManager, error) {
	// 创建主 Agent
	mainAgent, err := NewADKAgent(ctx, mainAgentConfig)
	if err != nil {
		return nil, err
	}

	// 创建子 Agent
	subAgents := make([]adk.Agent, 0, len(subAgentConfigs))
	for _, cfg := range subAgentConfigs {
		subAgent, err := NewADKAgent(ctx, cfg)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}

	return NewSubAgentManager(ctx, mainAgent, subAgents)
}

// BuildTransferToolDescription 构建转移工具的描述.
func BuildTransferToolDescription(subAgents []BuiltinAgentInfo) string {
	desc := "Transfer a task to another specialized agent. Available agents:\n"
	for _, agent := range subAgents {
		desc += "- " + agent.ID + ": " + agent.Description + "\n"
	}
	return desc
}
