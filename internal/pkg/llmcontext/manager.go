// Package llmcontext 提供 LLM 上下文管理器.
package llmcontext

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// Manager 上下文管理器接口.
type Manager interface {
	// AddMessage 添加消息到会话上下文
	AddMessage(ctx context.Context, sessionID string, message *schema.Message) error

	// GetContext 获取会话的上下文消息
	GetContext(ctx context.Context, sessionID string) ([]*schema.Message, error)

	// ClearContext 清除会话上下文
	ClearContext(ctx context.Context, sessionID string) error

	// SetSystemPrompt 设置或更新系统提示词
	SetSystemPrompt(ctx context.Context, sessionID string, systemPrompt string) error
}

// ContextStats 上下文统计信息.
type ContextStats struct {
	MessageCount         int
	TokenCount           int
	IsCompressed         bool
	OriginalMessageCount int
}

// Config 上下文管理器配置.
type Config struct {
	MaxTokens           int                 // 最大 token 数，默认 4096
	CompressionStrategy CompressionStrategy // 压缩策略，默认滑动窗口
	Storage             ContextStorage      // 存储后端，默认内存
}

// contextManager 上下文管理器实现.
type contextManager struct {
	storage             ContextStorage
	compressionStrategy CompressionStrategy
	maxTokens           int
}

// NewManager 创建上下文管理器.
func NewManager(cfg *Config) Manager {
	if cfg == nil {
		cfg = &Config{}
	}

	storage := cfg.Storage
	if storage == nil {
		storage = NewMemoryStorage()
	}

	strategy := cfg.CompressionStrategy
	if strategy == nil {
		strategy = NewSlidingWindowStrategy()
	}

	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096 // 默认 4k tokens
	}

	return &contextManager{
		storage:             storage,
		compressionStrategy: strategy,
		maxTokens:           maxTokens,
	}
}

// NewDefaultManager 创建默认配置的上下文管理器.
func NewDefaultManager() Manager {
	return NewManager(nil)
}

func (m *contextManager) AddMessage(ctx context.Context, sessionID string, message *schema.Message) error {
	// 加载现有消息
	messages, err := m.storage.Load(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to load context: %w", err)
	}

	// 添加新消息
	messages = append(messages, message)

	// 检查是否需要压缩
	tokenCount := m.compressionStrategy.EstimateTokens(messages)
	if tokenCount > m.maxTokens {
		compressed, err := m.compressionStrategy.Compress(ctx, messages, m.maxTokens)
		if err != nil {
			return fmt.Errorf("failed to compress context: %w", err)
		}
		messages = compressed
	}

	// 保存
	if err := m.storage.Save(ctx, sessionID, messages); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	return nil
}

func (m *contextManager) GetContext(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	messages, err := m.storage.Load(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load context: %w", err)
	}
	return messages, nil
}

func (m *contextManager) ClearContext(ctx context.Context, sessionID string) error {
	if err := m.storage.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to clear context: %w", err)
	}
	return nil
}

func (m *contextManager) SetSystemPrompt(ctx context.Context, sessionID string, systemPrompt string) error {
	messages, err := m.storage.Load(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to load context: %w", err)
	}

	systemMessage := schema.SystemMessage(systemPrompt)

	// 检查第一条是否是系统消息
	if len(messages) > 0 && messages[0].Role == schema.System {
		messages[0] = systemMessage
	} else {
		messages = append([]*schema.Message{systemMessage}, messages...)
	}

	if err := m.storage.Save(ctx, sessionID, messages); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	return nil
}
