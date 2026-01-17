// Package session 提供会话相关业务逻辑，包括流式问答.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	apimodel "github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/agent"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/enhanced"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/prompts"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/react"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/router"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	"github.com/ashwinyue/eino-show/internal/pkg/dedup"
	"github.com/ashwinyue/eino-show/pkg/store/where"
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
}

// ADKAgentResult ADK Agent 结果（用于 handler 层流式处理）.
type ADKAgentResult struct {
	Agent     *enhanced.ADKAgent // ADK Agent
	Messages  []*schema.Message  // 输入消息
	SessionID string             // 会话 ID
	MessageID string             // 消息 ID
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
			WebSearchEnabled: false,
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

	// 从数据库获取模型配置
	// 优先级：tenant.agent_config > session.agent.config > 默认模型
	var defaultModel *apimodel.LLMModelM
	tenantID := contextx.TenantID(ctx)

	// 1. 尝试从 tenant.agent_config 获取 model_id
	tenant, _ := e.store.Tenant().GetByID(ctx, uint64(tenantID))
	if tenant != nil && tenant.AgentConfig != nil && *tenant.AgentConfig != "" && *tenant.AgentConfig != "null" {
		var tenantAgentCfg struct {
			ModelID string `json:"model_id"`
		}
		if json.Unmarshal([]byte(*tenant.AgentConfig), &tenantAgentCfg) == nil && tenantAgentCfg.ModelID != "" {
			defaultModel, _ = e.store.Model().GetByID(ctx, tenantAgentCfg.ModelID)
		}
	}

	// 2. 如果 tenant 没有配置，尝试从 session 关联的 agent 获取
	if defaultModel == nil {
		session, _ := e.store.Session().Get(ctx, where.F("id", req.SessionID))
		if session != nil && session.AgentID != nil && *session.AgentID != "" {
			agent, _ := e.store.CustomAgent().Get(ctx, where.F("id", *session.AgentID))
			if agent != nil && agent.Config != "" && agent.Config != "{}" {
				var agentCfg struct {
					ModelID string `json:"model_id"`
				}
				if json.Unmarshal([]byte(agent.Config), &agentCfg) == nil && agentCfg.ModelID != "" {
					defaultModel, _ = e.store.Model().GetByID(ctx, agentCfg.ModelID)
				}
			}
		}
	}

	// 3. 如果都没有找到，使用默认模型
	if defaultModel == nil {
		var err error
		defaultModel, err = e.store.Model().GetDefault(ctx, "MODEL_TYPE_LLM")
		if err != nil {
			defaultModel, err = e.store.Model().GetDefault(ctx, "llm")
			if err != nil {
				return nil, fmt.Errorf("get default model failed: %w", err)
			}
		}
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

	// 包装为 ADK Agent
	adkAgent, err := enhanced.NewADKAgent(ctx, &enhanced.ADKAgentConfig{
		Name:        "react_agent",
		Description: "ReAct Agent for Q&A",
		EnhancedConfig: &enhanced.AgentConfig{
			ToolCallingModel: toolCallingModel,
			SystemPrompt:     systemPrompt,
			MaxStep:          reactCfg.MaxIterations,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create ADK agent failed: %w", err)
	}

	// 生成消息 ID
	messageID := fmt.Sprintf("%s-%d", req.SessionID, time.Now().UnixNano())

	return &ADKAgentResult{
		Agent:     adkAgent,
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
