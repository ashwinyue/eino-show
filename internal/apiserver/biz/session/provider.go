// Package session 提供 Session 业务层的依赖注入.
package session

import (
	"context"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/redis/go-redis/v9"

	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/agent/enhanced"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
	"github.com/ashwinyue/eino-show/internal/pkg/stream"
)

// SessionConfig 会话服务配置.
type SessionConfig struct {
	// Store 数据存储
	Store store.IStore

	// RedisClient Redis 客户端 (可选)
	RedisClient *redis.Client

	// ChatModelConfig 聊天模型配置 (可选，用于 QA)
	ChatModelConfig *agentmodel.Config

	// EmbeddingConfig 嵌入模型配置 (可选)
	EmbeddingConfig *agentmodel.EmbeddingConfig

	// EnableEnhanced 启用增强模式 (意图路由/动态Prompt/经验)
	EnableEnhanced bool

	// EnableMultiAgent 启用多 Agent 模式
	EnableMultiAgent bool
}

// NewSessionBizWithConfig 使用完整配置创建 SessionBiz.
func NewSessionBizWithConfig(ctx context.Context, cfg *SessionConfig) (SessionBiz, error) {
	if cfg == nil || cfg.Store == nil {
		return New(cfg.Store), nil
	}

	// 如果没有启用增强模式，使用简单模式
	if !cfg.EnableEnhanced && cfg.ChatModelConfig == nil {
		if cfg.RedisClient != nil {
			return NewWithRedis(cfg.Store, cfg.RedisClient), nil
		}
		return New(cfg.Store), nil
	}

	// 构建 QA 配置
	qaCfg := &QAConfig{
		Store:            cfg.Store,
		ChatModelConfig:  cfg.ChatModelConfig,
		EmbeddingConfig:  cfg.EmbeddingConfig,
		EnableMultiAgent: cfg.EnableMultiAgent,
	}

	// 初始化增强组件 (如果启用)
	if cfg.EnableEnhanced && cfg.ChatModelConfig != nil {
		if err := initEnhancedComponents(ctx, qaCfg, cfg); err != nil {
			// 增强组件初始化失败，回退到基础模式
			return NewWithQA(ctx, cfg.Store, qaCfg)
		}
	}

	return NewWithQA(ctx, cfg.Store, qaCfg)
}

// initEnhancedComponents 初始化增强组件.
// NOTE: 动态提示词、经验管理、意图路由暂时禁用，与 WeKnora 保持一致，便于测试.
func initEnhancedComponents(ctx context.Context, qaCfg *QAConfig, cfg *SessionConfig) error {
	// 1. 创建聊天模型 (用于意图路由)
	// NOTE: 暂时禁用，与 WeKnora 保持一致
	// chatModel, err := createChatModel(ctx, cfg.ChatModelConfig)
	// if err != nil {
	// 	return err
	// }
	_ = ctx // avoid unused
	_ = cfg // avoid unused

	// 2. 初始化意图路由器
	// NOTE: 暂时禁用，与 WeKnora 保持一致
	// intentRouter, err := router.NewIntentRouter(ctx, &router.RouterConfig{
	// 	ChatModel: chatModel,
	// 	FastIntentRules: []router.FastIntentRule{
	// 		{Patterns: []string{"你好", "hi", "hello"}, Response: "你好！有什么可以帮助你的吗？", Intent: router.IntentFastPath},
	// 		{Patterns: []string{"谢谢", "感谢"}, Response: "不客气！还有其他问题吗？", Intent: router.IntentFastPath},
	// 	},
	// })
	// if err == nil {
	// 	qaCfg.IntentRouter = intentRouter
	// }

	// 3. 初始化动态 Prompt 构建器
	// NOTE: 暂时禁用，与 WeKnora 保持一致
	// dynamicPromptBuilder := router.NewDefaultDynamicPromptBuilder(
	// 	"You are a helpful AI assistant with access to various tools and knowledge bases.",
	// )
	// qaCfg.DynamicPromptBuilder = dynamicPromptBuilder

	// 4. 初始化经验管理器
	// NOTE: 暂时禁用，与 WeKnora 保持一致
	// experienceManager, err := router.NewExperienceManager(&router.ExperienceManagerConfig{
	// 	MaxExperiences: 100,
	// })
	// if err == nil {
	// 	qaCfg.ExperienceManager = experienceManager
	// }

	// 5. 初始化子 Agent 配置 (如果启用多 Agent)
	if cfg.EnableMultiAgent {
		qaCfg.SubAgentConfigs = createBuiltinSubAgentConfigs(cfg)
	}

	return nil
}

// createChatModel 创建聊天模型.
func createChatModel(ctx context.Context, cfg *agentmodel.Config) (model.ChatModel, error) {
	if cfg == nil {
		return nil, nil
	}
	return agentmodel.NewChatModel(ctx, cfg)
}

// createBuiltinSubAgentConfigs 创建内置子 Agent 配置.
func createBuiltinSubAgentConfigs(cfg *SessionConfig) []*enhanced.ADKAgentConfig {
	return []*enhanced.ADKAgentConfig{
		{
			Name:        "快速问答",
			Description: "基于知识库的快速问答，适合简单的检索问题",
			EnhancedConfig: &enhanced.AgentConfig{
				SystemPrompt: "You are a quick Q&A assistant. Answer questions concisely based on the knowledge base.",
				MaxStep:      3,
			},
		},
		{
			Name:        "智能推理",
			Description: "ReAct 推理框架，支持多步思考和工具调用",
			EnhancedConfig: &enhanced.AgentConfig{
				SystemPrompt: "You are a smart reasoning assistant. Think step by step and use tools when needed.",
				MaxStep:      12,
			},
		},
	}
}

// StreamServices 流式服务集合.
type StreamServices struct {
	StreamManager         stream.StreamManager
	WebSearchStateService stream.WebSearchStateService
}

// NewStreamServices 创建流式服务.
func NewStreamServices(redisClient *redis.Client) *StreamServices {
	if redisClient == nil {
		return &StreamServices{
			StreamManager:         stream.NewMemoryStreamManager(),
			WebSearchStateService: stream.NewMemoryWebSearchStateService(),
		}
	}

	// 使用 Redis 实现
	streamMgr, err := stream.NewRedisStreamManager(&stream.RedisStreamConfig{
		Client: redisClient,
		Prefix: "sse:",
		TTL:    24 * time.Hour,
	})
	if err != nil {
		streamMgr = nil
	}

	webSearchState, err := stream.NewRedisWebSearchStateService(&stream.WebSearchStateConfig{
		Client: redisClient,
		Prefix: "websearch:",
		TTL:    time.Hour,
	})
	if err != nil {
		webSearchState = nil
	}

	services := &StreamServices{}
	if streamMgr != nil {
		services.StreamManager = streamMgr
	} else {
		services.StreamManager = stream.NewMemoryStreamManager()
	}
	if webSearchState != nil {
		services.WebSearchStateService = webSearchState
	} else {
		services.WebSearchStateService = stream.NewMemoryWebSearchStateService()
	}

	return services
}
