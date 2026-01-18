// Package session 提供会话相关业务逻辑，包括流式问答.
package session

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/agent"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/enhanced"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/prompts"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/react"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/router"
	agenttool "github.com/ashwinyue/eino-show/internal/pkg/agent/tool"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/tools"
	"github.com/ashwinyue/eino-show/internal/pkg/dedup"
)

// QAConfig QA 服务配置.
type QAConfig struct {
	Store           store.IStore
	ChatModelConfig *agentmodel.Config
	EmbeddingConfig *agentmodel.EmbeddingConfig

	// IntentRouter 意图路由器 (可选)
	IntentRouter *router.IntentRouter

	// DynamicPromptBuilder 动态 Prompt 构建器 (可选)
	DynamicPromptBuilder *router.DynamicPromptBuilder

	// ExperienceManager 经验管理器 (可选)
	ExperienceManager *router.ExperienceManager

	// EnableMultiAgent 启用多 Agent 模式 (可选)
	EnableMultiAgent bool

	// SubAgentConfigs 子 Agent 配置 (可选)
	SubAgentConfigs []*enhanced.ADKAgentConfig
}

// AgentQARequest Agent 问答请求.
type AgentQARequest struct {
	SessionID        string
	Query            string
	AgentType        string
	KnowledgeBaseIDs []string
	KnowledgeIDs     []string
	ModelID          string
	SystemPrompt     string
	MaxIterations    int
	Tools            []string
	History          []*schema.Message
	WebSearchEnabled bool // 是否启用网络搜索
}

// ADKAgentResult ADK Agent 结果（用于 handler 层流式处理）.
type ADKAgentResult struct {
	Agent     adk.Agent         // ADK Agent（使用接口类型以支持 ChatModelAgent）
	Runner    *adk.Runner       // ADK Runner（用于正确处理工具调用事件）
	Messages  []*schema.Message // 输入消息
	SessionID string            // 会话 ID
	MessageID string            // 消息 ID
}

// qaExecutor QA 执行器（内部实现）.
type qaExecutor struct {
	store                store.IStore
	factory              *agent.Factory
	intentRouter         *router.IntentRouter
	dynamicPromptBuilder *router.DynamicPromptBuilder
	experienceManager    *router.ExperienceManager
	enableMultiAgent     bool
	subAgentManager      *enhanced.SubAgentManager
	deduplicator         *dedup.QADeduplicator
}

// newQAExecutor 创建 QA 执行器.
func newQAExecutor(ctx context.Context, cfg *QAConfig) (*qaExecutor, error) {
	if cfg == nil || cfg.Store == nil {
		return nil, fmt.Errorf("invalid QA config")
	}

	factoryCfg := &agent.FactoryConfig{
		Store:           cfg.Store,
		ChatModelConfig: cfg.ChatModelConfig,
		EmbeddingConfig: cfg.EmbeddingConfig,
	}

	factory, err := agent.NewFactory(ctx, factoryCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent factory: %w", err)
	}

	executor := &qaExecutor{
		store:                cfg.Store,
		factory:              factory,
		intentRouter:         cfg.IntentRouter,
		dynamicPromptBuilder: cfg.DynamicPromptBuilder,
		experienceManager:    cfg.ExperienceManager,
		enableMultiAgent:     cfg.EnableMultiAgent,
		deduplicator:         dedup.NewQADeduplicator(),
	}

	return executor, nil
}

// getADKAgent 获取 ADK Agent 用于流式处理.
func (e *qaExecutor) getADKAgent(ctx context.Context, req *AgentQARequest) (*ADKAgentResult, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// 构建系统提示词
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = prompts.BuildSystemPrompt(&prompts.BuildConfig{
			KnowledgeBases:   buildKBInfos(req.KnowledgeBaseIDs),
			WebSearchEnabled: req.WebSearchEnabled,
		})
	}

	// 构建输入消息
	messages := req.History
	if messages == nil {
		messages = []*schema.Message{}
	}
	messages = append([]*schema.Message{schema.SystemMessage(systemPrompt)}, messages...)
	messages = append(messages, schema.UserMessage(req.Query))

	// 创建 React Agent 配置
	reactCfg := &react.Config{
		SystemPrompt:  systemPrompt,
		MaxIterations: req.MaxIterations,
	}
	if reactCfg.MaxIterations <= 0 {
		reactCfg.MaxIterations = 20
	}

	if len(req.Tools) > 0 {
		reactCfg.Tools = e.factory.GetTools(req.Tools)
	}

	// 从数据库获取模型配置（对齐 WeKnora）
	// 必须指定 model_id，不使用默认模型
	if req.ModelID == "" {
		return nil, fmt.Errorf("model_id is required")
	}

	defaultModel, err := e.store.Model().GetByID(ctx, req.ModelID)
	if err != nil {
		return nil, fmt.Errorf("get model failed: %w", err)
	}

	// 从数据库模型配置创建 ChatModel
	chatModel, err := e.factory.CreateChatModelFromDB(ctx, defaultModel.Name, defaultModel.Source, defaultModel.Parameters)
	if err != nil {
		return nil, fmt.Errorf("create chat model from db failed: %w", err)
	}

	// 转换为 ToolCallingChatModel
	toolCallingModel, ok := chatModel.(model.ToolCallingChatModel)
	if !ok {
		return nil, fmt.Errorf("chat model does not support tool calling")
	}

	// 构建工具列表（包含 thinking 工具 + 请求指定工具）
	toolList := make([]tool.BaseTool, 0, 4)
	toolList = append(toolList, tools.NewSequentialThinkingTool())
	toolList = append(toolList, agenttool.NewTodoTool())
	if len(req.Tools) > 0 {
		for _, t := range e.factory.GetTools(req.Tools) {
			toolList = append(toolList, t)
		}
	} else {
		for _, t := range e.factory.GetTools(nil) {
			toolList = append(toolList, t)
		}
	}
	toolsConfig := adk.ToolsConfig{
		ToolsNodeConfig: compose.ToolsNodeConfig{
			Tools: toolList,
		},
	}

	// 使用 ADK 内置的 ChatModelAgent（确保正确的流式处理）
	// 对齐官方示例: a-old/old/eino-examples/adk/intro/http-sse-service/main.go
	adkAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "react_agent",
		Description: "ReAct Agent for Q&A",
		Instruction: systemPrompt,
		Model:       toolCallingModel,
		ToolsConfig: toolsConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("create ADK agent failed: %w", err)
	}

	// 创建 ADK Runner（对齐 eino-examples 最佳实践）
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           adkAgent,
	})

	// 生成消息 ID
	messageID := fmt.Sprintf("%s-%d", req.SessionID, time.Now().UnixNano())

	return &ADKAgentResult{
		Agent:     adkAgent,
		Runner:    runner,
		Messages:  messages,
		SessionID: req.SessionID,
		MessageID: messageID,
	}, nil
}

// buildKBInfos 从知识库 ID 列表构建知识库信息.
func buildKBInfos(knowledgeBaseIDs []string) []*prompts.KnowledgeBaseInfo {
	if len(knowledgeBaseIDs) == 0 {
		return nil
	}
	infos := make([]*prompts.KnowledgeBaseInfo, len(knowledgeBaseIDs))
	for i, id := range knowledgeBaseIDs {
		infos[i] = &prompts.KnowledgeBaseInfo{
			ID:   id,
			Name: fmt.Sprintf("Knowledge Base %d", i+1),
			Type: "document",
		}
	}
	return infos
}
