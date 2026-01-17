// Package agent 提供根据配置创建 Agent 的工厂方法.
package agent

import (
	"context"
	"fmt"

	"github.com/ashwinyue/eino-show/internal/apiserver/cache"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	chatagent "github.com/ashwinyue/eino-show/internal/pkg/agent/chat"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/react"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/supervisor"
	agenttool "github.com/ashwinyue/eino-show/internal/pkg/agent/tool"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/tools"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/workflow"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
)

// FactoryConfig Agent 工厂配置.
type FactoryConfig struct {
	// Store 数据存储接口
	Store store.IStore

	// Cache 缓存接口（可选）
	Cache cache.ICache

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
func (f *Factory) CreateReactAgent(ctx context.Context, cfg *react.Config) (*react.Agent, error) {
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

	// 自动添加 thinking 工具（如果还没有）
	hasThinkingTool := false
	for _, t := range cfg.Tools {
		info, _ := t.Info(ctx)
		if info != nil && info.Name == tools.ToolThinking {
			hasThinkingTool = true
			break
		}
	}
	if !hasThinkingTool {
		cfg.Tools = append(cfg.Tools, tools.NewSequentialThinkingTool())
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

// CreateChatModelFromDB 从数据库模型配置创建 ChatModel.
// modelName: 模型名称 (如 "gpt-4o-mini")
// modelSource: 模型来源 (如 "openai", "dashscope", "ark")
// parametersJSON: 模型参数 JSON 字符串
func (f *Factory) CreateChatModelFromDB(ctx context.Context, modelName, modelSource, parametersJSON string) (model.ChatModel, error) {
	return agentmodel.NewChatModelFromDB(ctx, modelName, modelSource, parametersJSON)
}

// CreateEmbedderFromDB 从数据库模型配置创建 Embedder.
// modelName: 模型名称 (如 "text-embedding-3-small")
// modelSource: 模型来源 (如 "openai")
// parametersJSON: 模型参数 JSON 字符串
func (f *Factory) CreateEmbedderFromDB(ctx context.Context, modelName, modelSource, parametersJSON string) (embedding.Embedder, error) {
	return agentmodel.NewEmbedderFromDB(ctx, modelName, modelSource, parametersJSON)
}

// AgentDBConfig Agent 数据库配置（对齐 WeKnora AgentConfig）.
type AgentDBConfig struct {
	// AgentType Agent 类型: "react", "chat", "supervisor", "sequential", "loop", "parallel"
	AgentType string

	// ChatModelName LLM 模型名称
	ChatModelName string

	// ChatModelSource LLM 模型来源
	ChatModelSource string

	// ChatModelParams LLM 模型参数 JSON
	ChatModelParams string

	// EmbeddingModelName 向量模型名称 (可选)
	EmbeddingModelName string

	// EmbeddingModelSource 向量模型来源 (可选)
	EmbeddingModelSource string

	// EmbeddingModelParams 向量模型参数 JSON (可选)
	EmbeddingModelParams string

	// SystemPrompt 系统提示词（对齐 WeKnora 的 system_prompt）
	SystemPrompt string

	// Temperature 温度参数
	Temperature float64

	// MaxIterations 最大迭代次数 (仅 ReAct Agent)
	MaxIterations int

	// Tools 启用的工具名称列表（对齐 WeKnora 的 allowed_tools）
	Tools []string

	// KnowledgeBases 可访问的知识库 ID 列表（对齐 WeKnora）
	KnowledgeBases []string

	// KnowledgeIDs 可访问的知识（文档）ID 列表（对齐 WeKnora）
	KnowledgeIDs []string

	// ReflectionEnabled 是否启用反思（对齐 WeKnora）
	ReflectionEnabled bool

	// WebSearchEnabled 是否启用网络搜索（对齐 WeKnora）
	WebSearchEnabled bool

	// WebSearchMaxResults 网络搜索最大结果数（对齐 WeKnora）
	WebSearchMaxResults int

	// MultiTurnEnabled 是否启用多轮对话（对齐 WeKnora）
	MultiTurnEnabled bool

	// HistoryTurns 历史轮数（对齐 WeKnora）
	HistoryTurns int

	// MCPSelectionMode MCP 服务选择模式：all/selected/none（对齐 WeKnora）
	MCPSelectionMode string

	// MCPServices 选中的 MCP 服务 ID 列表（对齐 WeKnora）
	MCPServices []string

	// SubAgents 子 Agent ID 列表，用于多 Agent 协作（对齐 WeKnora）
	SubAgents []string

	// MaxLoopIterations 循环模式最大迭代次数 (仅 loop 模式)
	MaxLoopIterations int
}

// CreateAgentWithDBConfig 使用数据库配置创建 Agent（对齐 WeKnora）.
func (f *Factory) CreateAgentWithDBConfig(ctx context.Context, cfg *AgentDBConfig) (interface{}, error) {
	if cfg == nil {
		return nil, fmt.Errorf("agent db config is nil")
	}

	// 从数据库配置创建 ChatModel
	cm, err := f.CreateChatModelFromDB(ctx, cfg.ChatModelName, cfg.ChatModelSource, cfg.ChatModelParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model from db: %w", err)
	}

	// 根据 Agent 类型创建
	switch cfg.AgentType {
	case "react":
		// 确保是 ToolCallingChatModel
		tcm, ok := cm.(model.ToolCallingChatModel)
		if !ok {
			return nil, fmt.Errorf("react agent requires ToolCallingChatModel")
		}

		reactCfg := &react.Config{
			ChatModel: tcm,
		}

		// 设置系统提示词
		if cfg.SystemPrompt != "" {
			reactCfg.SystemPrompt = cfg.SystemPrompt
		}

		// 设置最大迭代次数
		if cfg.MaxIterations > 0 {
			reactCfg.MaxIterations = cfg.MaxIterations
		}

		// 设置温度参数
		if cfg.Temperature > 0 {
			temp := float32(cfg.Temperature)
			reactCfg.Temperature = &temp
		}

		// 构建工具列表（对齐 WeKnora 的工具选择逻辑）
		var toolNames []string

		// 1. 添加显式指定的工具
		toolNames = append(toolNames, cfg.Tools...)

		// 2. 根据配置自动添加工具
		if cfg.WebSearchEnabled {
			toolNames = append(toolNames, "web_search")
		}
		if len(cfg.KnowledgeBases) > 0 || len(cfg.KnowledgeIDs) > 0 {
			toolNames = append(toolNames, "knowledge_search")
		}

		// 3. 获取工具
		if len(toolNames) > 0 {
			reactCfg.Tools = f.GetTools(toolNames)
		} else {
			// 使用所有注册的工具
			reactCfg.Tools = f.GetTools(nil)
		}

		return f.CreateReactAgent(ctx, reactCfg)

	case "chat":
		chatCfg := &chatagent.Config{
			ChatModel: cm,
		}

		// 设置系统提示词
		if cfg.SystemPrompt != "" {
			chatCfg.SystemPrompt = cfg.SystemPrompt
		}

		// 设置温度参数
		if cfg.Temperature > 0 {
			temp := float32(cfg.Temperature)
			chatCfg.Temperature = &temp
		}

		return f.CreateChatAgent(ctx, chatCfg)

	case "supervisor":
		return f.createSupervisorFromDBConfig(ctx, cfg, cm)

	case "sequential", "loop", "parallel":
		return f.createWorkflowFromDBConfig(ctx, cfg)

	default:
		return nil, fmt.Errorf("unsupported agent type: %s", cfg.AgentType)
	}
}

// createSupervisorFromDBConfig creates a Supervisor from database config.
func (f *Factory) createSupervisorFromDBConfig(ctx context.Context, cfg *AgentDBConfig, cm model.ChatModel) (*supervisor.Supervisor, error) {
	tcm, ok := cm.(model.ToolCallingChatModel)
	if !ok {
		return nil, fmt.Errorf("supervisor requires ToolCallingChatModel")
	}

	if len(cfg.SubAgents) == 0 {
		return nil, fmt.Errorf("supervisor requires at least one sub agent")
	}

	// Build sub-agent configs (sub-agents need to be created separately)
	subAgentConfigs := make([]*supervisor.SubAgentConfig, 0, len(cfg.SubAgents))
	for _, subAgentID := range cfg.SubAgents {
		// For now, create a placeholder - in real usage, sub-agents should be loaded from DB
		subAgentConfigs = append(subAgentConfigs, &supervisor.SubAgentConfig{
			Name:        subAgentID,
			Description: fmt.Sprintf("Sub-agent %s", subAgentID),
			Agent:       nil, // Will be set when loading from DB
		})
	}

	return supervisor.New(ctx, &supervisor.Config{
		Name:         cfg.AgentType,
		ChatModel:    tcm,
		SystemPrompt: cfg.SystemPrompt,
		SubAgents:    subAgentConfigs,
	})
}

// createWorkflowFromDBConfig creates a Workflow from database config.
func (f *Factory) createWorkflowFromDBConfig(ctx context.Context, cfg *AgentDBConfig) (*workflow.Workflow, error) {
	if len(cfg.SubAgents) == 0 {
		return nil, fmt.Errorf("workflow requires at least one sub agent")
	}

	// Determine workflow mode
	var mode workflow.WorkflowMode
	switch cfg.AgentType {
	case "sequential":
		mode = workflow.ModeSequential
	case "loop":
		mode = workflow.ModeLoop
	case "parallel":
		mode = workflow.ModeParallel
	default:
		mode = workflow.ModeSequential
	}

	// Build sub-agents list (sub-agents need to be created separately)
	subAgents := make([]adk.Agent, 0, len(cfg.SubAgents))
	// Note: In real usage, sub-agents should be loaded and created from DB

	maxIter := cfg.MaxLoopIterations
	if maxIter <= 0 {
		maxIter = 10
	}

	return workflow.New(ctx, &workflow.Config{
		Name:          cfg.AgentType,
		Description:   cfg.SystemPrompt,
		Mode:          mode,
		SubAgents:     subAgents,
		MaxIterations: maxIter,
	})
}

// CreateSupervisor creates a Supervisor with sub-agents.
func (f *Factory) CreateSupervisor(ctx context.Context, name string, subAgents []*supervisor.SubAgentConfig) (*supervisor.Supervisor, error) {
	return supervisor.New(ctx, &supervisor.Config{
		Name:      name,
		ChatModel: f.tcm,
		SubAgents: subAgents,
	})
}

// CreateWorkflow creates a Workflow with the specified mode.
func (f *Factory) CreateWorkflow(ctx context.Context, name string, mode workflow.WorkflowMode, subAgents []adk.Agent) (*workflow.Workflow, error) {
	return workflow.New(ctx, &workflow.Config{
		Name:      name,
		Mode:      mode,
		SubAgents: subAgents,
	})
}

// CreateSequentialWorkflow creates a sequential workflow.
func (f *Factory) CreateSequentialWorkflow(ctx context.Context, name string, subAgents ...adk.Agent) (*workflow.Workflow, error) {
	return workflow.SequentialWorkflow(ctx, name, subAgents...)
}

// CreateLoopWorkflow creates a loop workflow.
func (f *Factory) CreateLoopWorkflow(ctx context.Context, name string, maxIterations int, subAgents ...adk.Agent) (*workflow.Workflow, error) {
	return workflow.LoopWorkflow(ctx, name, maxIterations, subAgents...)
}

// CreateParallelWorkflow creates a parallel workflow.
func (f *Factory) CreateParallelWorkflow(ctx context.Context, name string, subAgents ...adk.Agent) (*workflow.Workflow, error) {
	return workflow.ParallelWorkflow(ctx, name, subAgents...)
}

// GetToolCallingChatModel returns the ToolCallingChatModel.
func (f *Factory) GetToolCallingChatModel() model.ToolCallingChatModel {
	return f.tcm
}

// GetStore 获取 Store 接口.
func (f *Factory) GetStore() store.IStore {
	return f.cfg.Store
}

// GetCache 获取缓存接口（如果配置了）.
func (f *Factory) GetCache() cache.ICache {
	return f.cfg.Cache
}

// GetEmbedder 获取 Embedder 接口.
// 如果未配置 Embedding，返回 nil.
func (f *Factory) GetEmbedder() embedding.Embedder {
	return f.embed
}
