// Package llmcontext 提供消息压缩策略.
package llmcontext

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/eino-show/internal/pkg/retriever"
)

// CompressionStrategy 压缩策略接口.
type CompressionStrategy interface {
	// Compress 压缩消息列表，使其符合 token 限制
	Compress(ctx context.Context, messages []*schema.Message, maxTokens int) ([]*schema.Message, error)

	// EstimateTokens 估算消息列表的 token 数量
	EstimateTokens(messages []*schema.Message) int
}

// ProgressiveCompressionStrategy 渐进式压缩策略 (使用 LLM 生成摘要).
type ProgressiveCompressionStrategy struct {
	compressor *retriever.ProgressiveCompressor
}

// NewProgressiveCompressionStrategy 创建渐进式压缩策略.
func NewProgressiveCompressionStrategy(chatModel model.ChatModel, sessionID string, summaryStore retriever.SummaryStore) (CompressionStrategy, error) {
	cfg := retriever.NewDefaultProgressiveConfig()
	cfg.ChatModel = chatModel
	cfg.SessionID = sessionID
	cfg.SummaryStore = summaryStore

	compressor, err := retriever.NewProgressiveCompressor(cfg)
	if err != nil {
		return nil, err
	}

	return &ProgressiveCompressionStrategy{compressor: compressor}, nil
}

func (s *ProgressiveCompressionStrategy) EstimateTokens(messages []*schema.Message) int {
	return s.compressor.EstimateTokens(messages)
}

func (s *ProgressiveCompressionStrategy) Compress(ctx context.Context, messages []*schema.Message, maxTokens int) ([]*schema.Message, error) {
	// 使用渐进式压缩
	result, err := s.compressor.Compress(ctx, messages)
	if err != nil {
		return nil, err
	}
	return result.Messages, nil
}

// SlidingWindowStrategy 滑动窗口压缩策略.
// 保留系统消息和最近的 N 条消息.
type SlidingWindowStrategy struct {
	// TokensPerChar 每个字符估算的 token 数（中文约 0.5，英文约 0.25）
	TokensPerChar float64
}

// NewSlidingWindowStrategy 创建滑动窗口压缩策略.
func NewSlidingWindowStrategy() CompressionStrategy {
	return &SlidingWindowStrategy{
		TokensPerChar: 0.5, // 默认值，适合中英文混合
	}
}

func (s *SlidingWindowStrategy) EstimateTokens(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		// 估算: 内容长度 * 系数 + 角色开销
		total += int(float64(len(msg.Content))*s.TokensPerChar) + 4
	}
	return total
}

func (s *SlidingWindowStrategy) Compress(ctx context.Context, messages []*schema.Message, maxTokens int) ([]*schema.Message, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	// 当前 token 数
	currentTokens := s.EstimateTokens(messages)
	if currentTokens <= maxTokens {
		return messages, nil
	}

	// 分离系统消息和其他消息
	var systemMsg *schema.Message
	var otherMsgs []*schema.Message

	for _, msg := range messages {
		if msg.Role == schema.System {
			systemMsg = msg
		} else {
			otherMsgs = append(otherMsgs, msg)
		}
	}

	// 系统消息的 token 数
	systemTokens := 0
	if systemMsg != nil {
		systemTokens = int(float64(len(systemMsg.Content))*s.TokensPerChar) + 4
	}

	// 可用于其他消息的 token 数
	availableTokens := maxTokens - systemTokens
	if availableTokens < 100 {
		availableTokens = 100 // 至少保留 100 tokens
	}

	// 从最新的消息开始保留
	var kept []*schema.Message
	usedTokens := 0

	for i := len(otherMsgs) - 1; i >= 0; i-- {
		msg := otherMsgs[i]
		msgTokens := int(float64(len(msg.Content))*s.TokensPerChar) + 4

		if usedTokens+msgTokens > availableTokens {
			break
		}

		kept = append([]*schema.Message{msg}, kept...)
		usedTokens += msgTokens
	}

	// 重新组合：系统消息 + 保留的消息
	var result []*schema.Message
	if systemMsg != nil {
		result = append(result, systemMsg)
	}
	result = append(result, kept...)

	return result, nil
}
