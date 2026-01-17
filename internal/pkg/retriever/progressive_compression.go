// Package retriever provides progressive compression for conversation context.
// Reference: WeKnora llmcontext/progressive_compression.go
package retriever

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// SummaryStore 摘要持久化存储接口.
type SummaryStore interface {
	// SaveSummary 保存摘要到会话
	SaveSummary(ctx context.Context, sessionID string, summary string, tokenCount int) error

	// GetSummaries 获取会话的所有摘要
	GetSummaries(ctx context.Context, sessionID string) ([]string, error)

	// ReplaceSummaries 替换所有摘要为一个合并后的摘要
	ReplaceSummaries(ctx context.Context, sessionID string, mergedSummary string, tokenCount int) error

	// ClearSummaries 清除会话的所有摘要
	ClearSummaries(ctx context.Context, sessionID string) error
}

const (
	// DefaultMaxSummaries 最大摘要层数
	DefaultMaxSummaries = 3
	// DefaultCompressThreshold 触发压缩的最少消息数
	DefaultCompressThreshold = 15
	// DefaultKeepRecentCount 压缩后保留的最近消息数
	DefaultKeepRecentCount = 5
	// DefaultSummaryTokens 摘要的最大 token 数
	DefaultSummaryTokens = 500
)

// ProgressiveCompressionConfig 渐进式压缩配置.
type ProgressiveCompressionConfig struct {
	// ChatModel LLM model for compression
	ChatModel model.ChatModel

	// SummaryStore 摘要持久化存储 (可选，不设置则使用内存)
	SummaryStore SummaryStore

	// SessionID 会话 ID (使用持久化存储时必须)
	SessionID string

	// MaxSummaries 最大摘要层数 (default: 3)
	MaxSummaries int

	// CompressThreshold 触发压缩的消息阈值 (default: 15)
	CompressThreshold int

	// KeepRecentCount 压缩后保留的最近消息数 (default: 5)
	KeepRecentCount int

	// SummaryTokens 单次摘要的最大 token 数 (default: 500)
	SummaryTokens int

	// SummaryPrompt 自定义摘要提示词 (可选)
	SummaryPrompt string
}

// NewDefaultProgressiveConfig 创建默认配置.
func NewDefaultProgressiveConfig() *ProgressiveCompressionConfig {
	return &ProgressiveCompressionConfig{
		MaxSummaries:      DefaultMaxSummaries,
		CompressThreshold: DefaultCompressThreshold,
		KeepRecentCount:   DefaultKeepRecentCount,
		SummaryTokens:     DefaultSummaryTokens,
	}
}

// ProgressiveCompressor 渐进式压缩器.
// 支持多层摘要存储，提升历史信息保留率.
type ProgressiveCompressor struct {
	cfg          *ProgressiveCompressionConfig
	summaries    []string // 已有的摘要层 (内存模式)
	sessionID    string   // 当前会话 ID
	summaryStore SummaryStore
}

// NewProgressiveCompressor 创建渐进式压缩器.
func NewProgressiveCompressor(cfg *ProgressiveCompressionConfig) (*ProgressiveCompressor, error) {
	if cfg == nil {
		cfg = NewDefaultProgressiveConfig()
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if cfg.MaxSummaries <= 0 {
		cfg.MaxSummaries = DefaultMaxSummaries
	}
	if cfg.CompressThreshold <= 0 {
		cfg.CompressThreshold = DefaultCompressThreshold
	}
	if cfg.KeepRecentCount <= 0 {
		cfg.KeepRecentCount = DefaultKeepRecentCount
	}
	if cfg.SummaryTokens <= 0 {
		cfg.SummaryTokens = DefaultSummaryTokens
	}

	return &ProgressiveCompressor{
		cfg:          cfg,
		summaries:    make([]string, 0, cfg.MaxSummaries),
		sessionID:    cfg.SessionID,
		summaryStore: cfg.SummaryStore,
	}, nil
}

// SetSessionID 设置会话 ID (用于持久化).
func (pc *ProgressiveCompressor) SetSessionID(sessionID string) {
	pc.sessionID = sessionID
}

// loadSummaries 从存储加载摘要.
func (pc *ProgressiveCompressor) loadSummaries(ctx context.Context) error {
	if pc.summaryStore == nil || pc.sessionID == "" {
		return nil // 内存模式，不需要加载
	}

	summaries, err := pc.summaryStore.GetSummaries(ctx, pc.sessionID)
	if err != nil {
		return fmt.Errorf("failed to load summaries: %w", err)
	}
	pc.summaries = summaries
	return nil
}

// saveSummary 保存摘要到存储.
func (pc *ProgressiveCompressor) saveSummary(ctx context.Context, summary string) error {
	if pc.summaryStore == nil || pc.sessionID == "" {
		return nil // 内存模式，不需要保存
	}

	tokenCount := len(summary) / 4 // 估算
	return pc.summaryStore.SaveSummary(ctx, pc.sessionID, summary, tokenCount)
}

// replaceSummaries 替换所有摘要.
func (pc *ProgressiveCompressor) replaceSummaries(ctx context.Context, mergedSummary string) error {
	if pc.summaryStore == nil || pc.sessionID == "" {
		return nil // 内存模式，不需要保存
	}

	tokenCount := len(mergedSummary) / 4 // 估算
	return pc.summaryStore.ReplaceSummaries(ctx, pc.sessionID, mergedSummary, tokenCount)
}

// CompressionResult 压缩结果.
type CompressionResult struct {
	Messages             []*schema.Message // 压缩后的消息列表
	OriginalMessageCount int               // 原始消息数量
	FinalMessageCount    int               // 压缩后消息数量
	SummaryCount         int               // 当前摘要总数
	SummaryContent       string            // 新生成的摘要内容
	Compressed           bool              // 是否执行了压缩
	TokensSaved          int               // 估算节省的 token 数量
}

// Compress 执行渐进式压缩.
// 1. 加载已有摘要 (持久化模式)
// 2. 检查是否需要压缩
// 3. 分离系统消息和普通消息
// 4. 压缩旧消息为摘要
// 5. 保存摘要 (持久化模式)
// 6. 保留最近消息不变
func (pc *ProgressiveCompressor) Compress(ctx context.Context, messages []*schema.Message) (*CompressionResult, error) {
	// 1. 加载已有摘要 (持久化模式)
	if err := pc.loadSummaries(ctx); err != nil {
		// 加载失败，继续使用内存摘要
	}

	// 分离系统消息和普通消息
	systemMessages, regularMessages := pc.separateSystemMessages(messages)
	originalCount := len(messages)

	// 检查是否需要压缩
	if len(regularMessages) <= pc.cfg.CompressThreshold {
		return &CompressionResult{
			Messages:             messages,
			OriginalMessageCount: originalCount,
			FinalMessageCount:    originalCount,
			SummaryCount:         len(pc.summaries),
			Compressed:           false,
		}, nil
	}

	// 计算需要压缩的消息数量
	needsCompression := len(regularMessages) - pc.cfg.KeepRecentCount
	if needsCompression <= 0 {
		return &CompressionResult{
			Messages:             messages,
			OriginalMessageCount: originalCount,
			FinalMessageCount:    originalCount,
			SummaryCount:         len(pc.summaries),
			Compressed:           false,
		}, nil
	}

	// 分离旧消息和最近消息
	oldMessages := regularMessages[:needsCompression]
	recentMessages := regularMessages[needsCompression:]

	// 生成新摘要
	newSummary, err := pc.summarizeMessages(ctx, oldMessages)
	if err != nil {
		// 压缩失败，使用滑动窗口 fallback
		result := pc.buildResult(systemMessages, recentMessages)
		return &CompressionResult{
			Messages:             result,
			OriginalMessageCount: originalCount,
			FinalMessageCount:    len(result),
			SummaryCount:         len(pc.summaries),
			Compressed:           false,
		}, nil
	}

	// 检查是否需要合并摘要
	if len(pc.summaries) >= pc.cfg.MaxSummaries {
		// 合并所有摘要为一个
		pc.summaries = append(pc.summaries, newSummary)
		mergedSummary, err := pc.mergeSummaries(ctx)
		if err == nil {
			pc.summaries = []string{mergedSummary}
			newSummary = mergedSummary
			// 持久化：替换所有摘要
			_ = pc.replaceSummaries(ctx, mergedSummary)
		}
	} else {
		pc.summaries = append(pc.summaries, newSummary)
		// 持久化：保存新摘要
		_ = pc.saveSummary(ctx, newSummary)
	}

	// 构建结果
	result := pc.buildResultWithSummary(systemMessages, recentMessages)
	tokensSaved := pc.estimateSavedTokens(oldMessages)

	return &CompressionResult{
		Messages:             result,
		OriginalMessageCount: originalCount,
		FinalMessageCount:    len(result),
		SummaryCount:         len(pc.summaries),
		SummaryContent:       newSummary,
		Compressed:           true,
		TokensSaved:          tokensSaved,
	}, nil
}

// ShouldCompress 判断是否需要压缩.
func (pc *ProgressiveCompressor) ShouldCompress(messageCount int) bool {
	return messageCount >= pc.cfg.CompressThreshold
}

// GetSummaries 获取当前所有摘要.
func (pc *ProgressiveCompressor) GetSummaries() []string {
	return pc.summaries
}

// SetSummaries 设置摘要 (用于恢复会话状态).
func (pc *ProgressiveCompressor) SetSummaries(summaries []string) {
	pc.summaries = summaries
}

// ClearSummaries 清空摘要.
func (pc *ProgressiveCompressor) ClearSummaries() {
	pc.summaries = make([]string, 0, pc.cfg.MaxSummaries)
}

// separateSystemMessages 分离系统消息和普通消息.
func (pc *ProgressiveCompressor) separateSystemMessages(messages []*schema.Message) (system, regular []*schema.Message) {
	for _, msg := range messages {
		if msg.Role == schema.System {
			system = append(system, msg)
		} else {
			regular = append(regular, msg)
		}
	}
	return
}

// summarizeMessages 使用 LLM 生成摘要.
func (pc *ProgressiveCompressor) summarizeMessages(ctx context.Context, messages []*schema.Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	// 构建对话文本
	var conversation strings.Builder
	for _, msg := range messages {
		role := string(msg.Role)
		conversation.WriteString(fmt.Sprintf("[%s]: %s\n", role, msg.Content))
	}

	// 构建摘要提示词
	prompt := pc.buildSummaryPrompt()
	userPrompt := fmt.Sprintf("Please summarize the following conversation:\n\n%s", conversation.String())

	summaryMessages := []*schema.Message{
		schema.SystemMessage(prompt),
		schema.UserMessage(userPrompt),
	}

	resp, err := pc.cfg.ChatModel.Generate(ctx, summaryMessages)
	if err != nil {
		return "", fmt.Errorf("LLM summarization failed: %w", err)
	}

	return resp.Content, nil
}

// mergeSummaries 合并多个摘要为一个.
func (pc *ProgressiveCompressor) mergeSummaries(ctx context.Context) (string, error) {
	if len(pc.summaries) < 2 {
		if len(pc.summaries) == 1 {
			return pc.summaries[0], nil
		}
		return "", nil
	}

	// 构建合并文本
	summaryText := strings.Join(pc.summaries, "\n\n---\n\n")

	prompt := `You are a professional summarizer. Merge the following conversation summaries into one coherent summary.
Keep all important information, decisions, and context. Be concise but comprehensive.`

	userPrompt := fmt.Sprintf("Please merge these summaries:\n\n%s", summaryText)

	messages := []*schema.Message{
		schema.SystemMessage(prompt),
		schema.UserMessage(userPrompt),
	}

	resp, err := pc.cfg.ChatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM merge failed: %w", err)
	}

	return resp.Content, nil
}

// buildSummaryPrompt 构建摘要提示词.
func (pc *ProgressiveCompressor) buildSummaryPrompt() string {
	if pc.cfg.SummaryPrompt != "" {
		return pc.cfg.SummaryPrompt
	}

	return `You are a professional conversation summarizer. Please summarize the conversation following these guidelines:

1. Extract key information, decisions, and context from the conversation
2. Preserve important facts and details
3. Use clear and concise language
4. Keep the summary within 200-300 words

The summary should allow readers to quickly understand the main content while retaining enough information to continue the conversation.`
}

// buildResult 构建结果消息列表 (无摘要).
func (pc *ProgressiveCompressor) buildResult(systemMessages, recentMessages []*schema.Message) []*schema.Message {
	result := make([]*schema.Message, 0, len(systemMessages)+len(recentMessages))
	result = append(result, systemMessages...)
	result = append(result, recentMessages...)
	return result
}

// buildResultWithSummary 构建带摘要的结果消息列表.
func (pc *ProgressiveCompressor) buildResultWithSummary(systemMessages, recentMessages []*schema.Message) []*schema.Message {
	// 构建摘要消息
	summaryContent := pc.buildSummaryMessage()

	result := make([]*schema.Message, 0, len(systemMessages)+1+len(recentMessages))
	result = append(result, systemMessages...)

	if summaryContent != "" {
		result = append(result, schema.SystemMessage(fmt.Sprintf("Previous conversation summary:\n%s", summaryContent)))
	}

	result = append(result, recentMessages...)
	return result
}

// buildSummaryMessage 构建摘要消息内容.
func (pc *ProgressiveCompressor) buildSummaryMessage() string {
	if len(pc.summaries) == 0 {
		return ""
	}

	if len(pc.summaries) == 1 {
		return pc.summaries[0]
	}

	// 多层摘要，按顺序组合
	var sb strings.Builder
	for i, summary := range pc.summaries {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("[Summary %d]:\n%s", i+1, summary))
	}
	return sb.String()
}

// estimateSavedTokens 估算节省的 token 数量.
func (pc *ProgressiveCompressor) estimateSavedTokens(messages []*schema.Message) int {
	originalChars := 0
	for _, msg := range messages {
		originalChars += len(msg.Content)
		// 考虑 tool calls
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				originalChars += len(tc.Function.Name) + len(tc.Function.Arguments)
			}
		}
	}
	// 估算：4 字符 ≈ 1 token，摘要约为原始的 10-20%
	originalTokens := originalChars / 4
	summaryTokens := originalTokens / 5
	return originalTokens - summaryTokens
}

// EstimateTokens 估算消息的 token 数量.
func (pc *ProgressiveCompressor) EstimateTokens(messages []*schema.Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(string(msg.Role)) + len(msg.Content)
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				totalChars += len(tc.Function.Name) + len(tc.Function.Arguments)
			}
		}
	}
	return totalChars / 4
}
