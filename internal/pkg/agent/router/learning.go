// Package router provides automatic learning from successful executions.
// Reference: AssistantAgent Learning Module
package router

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ExecutionResult 执行结果.
type ExecutionResult struct {
	SessionID      string           // 会话 ID
	OriginalQuery  string           // 原始查询
	RewrittenQuery string           // 改写后的查询
	Intent         IntentType       // 识别的意图
	ToolCalls      []ToolCallRecord // 工具调用记录
	FinalAnswer    string           // 最终答案
	Success        bool             // 是否成功
	Duration       time.Duration    // 执行耗时
	UserFeedback   *UserFeedback    // 用户反馈 (可选)
	Metadata       map[string]any   // 元数据
}

// ToolCallRecord 工具调用记录.
type ToolCallRecord struct {
	ToolName string        // 工具名称
	Input    string        // 输入参数
	Output   string        // 输出结果
	Success  bool          // 是否成功
	Duration time.Duration // 耗时
}

// UserFeedback 用户反馈.
type UserFeedback struct {
	Helpful   bool   // 是否有帮助
	Rating    int    // 评分 (1-5)
	Comment   string // 评论
	Timestamp time.Time
}

// LearningConfig 学习配置.
type LearningConfig struct {
	// ExperienceManager 经验管理器
	ExperienceManager *ExperienceManager

	// ChatModel LLM 模型 (用于提取经验摘要)
	ChatModel model.ChatModel

	// MinSuccessDuration 最小成功执行时间 (过滤过快的简单任务)
	MinSuccessDuration time.Duration

	// MaxSuccessDuration 最大成功执行时间 (过滤超时任务)
	MaxSuccessDuration time.Duration

	// MinToolCalls 最少工具调用数 (过滤无工具调用的简单对话)
	MinToolCalls int

	// AutoLearnEnabled 是否启用自动学习
	AutoLearnEnabled bool

	// ExtractSummary 是否使用 LLM 提取摘要
	ExtractSummary bool
}

// LearningManager 学习管理器.
type LearningManager struct {
	cfg *LearningConfig
}

// NewLearningManager 创建学习管理器.
func NewLearningManager(cfg *LearningConfig) (*LearningManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.ExperienceManager == nil {
		return nil, fmt.Errorf("experience manager is required")
	}

	// 默认配置
	if cfg.MinSuccessDuration == 0 {
		cfg.MinSuccessDuration = 500 * time.Millisecond
	}
	if cfg.MaxSuccessDuration == 0 {
		cfg.MaxSuccessDuration = 5 * time.Minute
	}
	if cfg.MinToolCalls == 0 {
		cfg.MinToolCalls = 1
	}

	return &LearningManager{cfg: cfg}, nil
}

// LearnFromExecution 从执行结果中学习.
func (m *LearningManager) LearnFromExecution(ctx context.Context, result *ExecutionResult) error {
	if !m.cfg.AutoLearnEnabled {
		return nil
	}

	// 1. 检查是否值得学习
	if !m.shouldLearn(result) {
		return nil
	}

	// 2. 判断经验类型
	expType := m.determineExperienceType(result)

	// 3. 提取经验
	exp, err := m.extractExperience(ctx, result, expType)
	if err != nil {
		return fmt.Errorf("extract experience failed: %w", err)
	}

	// 4. 检查是否已存在相似经验
	existing, err := m.cfg.ExperienceManager.Recall(ctx, result.OriginalQuery)
	if err == nil && len(existing) > 0 {
		for _, e := range existing {
			if e.Score > 0.95 {
				// 非常相似的经验已存在，更新使用次数
				return m.cfg.ExperienceManager.RecordUsage(ctx, e.ID, result.Success)
			}
		}
	}

	// 5. 保存新经验
	return m.cfg.ExperienceManager.Learn(ctx, exp)
}

// LearnFromFeedback 从用户反馈中学习.
func (m *LearningManager) LearnFromFeedback(ctx context.Context, result *ExecutionResult, feedback *UserFeedback) error {
	result.UserFeedback = feedback

	// 正面反馈：强化学习
	if feedback.Helpful && feedback.Rating >= 4 {
		return m.LearnFromExecution(ctx, result)
	}

	// 负面反馈：记录失败
	if !feedback.Helpful || feedback.Rating <= 2 {
		// 查找相关经验并降低其成功率
		existing, err := m.cfg.ExperienceManager.Recall(ctx, result.OriginalQuery)
		if err == nil && len(existing) > 0 {
			for _, e := range existing {
				if e.Score > 0.9 {
					_ = m.cfg.ExperienceManager.RecordUsage(ctx, e.ID, false)
				}
			}
		}
	}

	return nil
}

// shouldLearn 判断是否值得学习.
func (m *LearningManager) shouldLearn(result *ExecutionResult) bool {
	// 必须成功
	if !result.Success {
		return false
	}

	// 执行时间过滤
	if result.Duration < m.cfg.MinSuccessDuration {
		return false // 太快，可能是简单对话
	}
	if result.Duration > m.cfg.MaxSuccessDuration {
		return false // 太慢，可能有问题
	}

	// 工具调用过滤
	if len(result.ToolCalls) < m.cfg.MinToolCalls {
		return false // 没有工具调用，简单对话不学习
	}

	// 检查工具调用是否都成功
	for _, tc := range result.ToolCalls {
		if !tc.Success {
			return false // 有工具调用失败
		}
	}

	// 有用户正面反馈优先学习
	if result.UserFeedback != nil && result.UserFeedback.Helpful {
		return true
	}

	return true
}

// determineExperienceType 判断经验类型.
func (m *LearningManager) determineExperienceType(result *ExecutionResult) ExperienceType {
	// 根据意图和工具调用判断
	switch result.Intent {
	case IntentTool:
		return ExperienceTypeReact
	case IntentRAG:
		return ExperienceTypeKnowledge
	default:
		if len(result.ToolCalls) > 0 {
			return ExperienceTypeReact
		}
		return ExperienceTypeKnowledge
	}
}

// extractExperience 提取经验.
func (m *LearningManager) extractExperience(ctx context.Context, result *ExecutionResult, expType ExperienceType) (*Experience, error) {
	exp := &Experience{
		ID:        generateExperienceID(result),
		Type:      expType,
		Query:     result.OriginalQuery,
		Response:  result.FinalAnswer,
		ToolCalls: m.extractToolCallNames(result.ToolCalls),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  result.Metadata,
	}

	// 使用 LLM 提取摘要 (可选)
	if m.cfg.ExtractSummary && m.cfg.ChatModel != nil {
		summary, err := m.extractSummaryWithLLM(ctx, result)
		if err == nil && summary != "" {
			exp.Response = summary
		}
	}

	// 检查是否可以作为快速意图
	if m.canBeFastIntent(result) {
		exp.Type = ExperienceTypeFastIntent
		exp.FastIntentConfig = &FastIntentConfig{
			Patterns:       m.extractPatterns(result.OriginalQuery),
			MatchType:      "prefix",
			DirectResponse: result.FinalAnswer,
		}
	}

	return exp, nil
}

// extractToolCallNames 提取工具调用名称.
func (m *LearningManager) extractToolCallNames(calls []ToolCallRecord) []string {
	names := make([]string, 0, len(calls))
	for _, c := range calls {
		names = append(names, c.ToolName)
	}
	return names
}

// extractSummaryWithLLM 使用 LLM 提取摘要.
func (m *LearningManager) extractSummaryWithLLM(ctx context.Context, result *ExecutionResult) (string, error) {
	prompt := fmt.Sprintf(`请从以下对话中提取关键信息作为经验摘要：

用户查询: %s
执行的工具: %s
最终答案: %s

请用简洁的一段话总结这次成功执行的关键步骤和结果，以便后续类似问题可以参考。`,
		result.OriginalQuery,
		strings.Join(m.extractToolCallNames(result.ToolCalls), ", "),
		result.FinalAnswer)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个经验总结助手，负责从成功的执行中提取可复用的经验。"),
		schema.UserMessage(prompt),
	}

	resp, err := m.cfg.ChatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// canBeFastIntent 判断是否可以作为快速意图.
func (m *LearningManager) canBeFastIntent(result *ExecutionResult) bool {
	// 条件：
	// 1. 查询较短 (< 20 字符)
	// 2. 只有一个工具调用
	// 3. 执行很快 (< 2秒)
	// 4. 有正面用户反馈

	if len(result.OriginalQuery) > 20 {
		return false
	}

	if len(result.ToolCalls) != 1 {
		return false
	}

	if result.Duration > 2*time.Second {
		return false
	}

	if result.UserFeedback != nil && result.UserFeedback.Rating >= 5 {
		return true
	}

	return false
}

// extractPatterns 提取匹配模式.
func (m *LearningManager) extractPatterns(query string) []string {
	// 简单实现：取前缀
	patterns := []string{}

	// 完整查询
	patterns = append(patterns, query)

	// 前缀 (如果足够长)
	if len(query) > 5 {
		patterns = append(patterns, query[:len(query)/2])
	}

	return patterns
}

// generateExperienceID 生成经验 ID.
func generateExperienceID(result *ExecutionResult) string {
	data := fmt.Sprintf("%s_%s_%d", result.SessionID, result.OriginalQuery, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return "exp_" + hex.EncodeToString(hash[:8])
}

// LearningCallback 学习回调 (用于集成到 Agent 执行流程).
type LearningCallback struct {
	manager   *LearningManager
	startTime time.Time
	result    *ExecutionResult
}

// NewLearningCallback 创建学习回调.
func NewLearningCallback(manager *LearningManager, sessionID, query string) *LearningCallback {
	return &LearningCallback{
		manager:   manager,
		startTime: time.Now(),
		result: &ExecutionResult{
			SessionID:     sessionID,
			OriginalQuery: query,
			ToolCalls:     make([]ToolCallRecord, 0),
			Metadata:      make(map[string]any),
		},
	}
}

// RecordToolCall 记录工具调用.
func (c *LearningCallback) RecordToolCall(toolName, input, output string, success bool, duration time.Duration) {
	c.result.ToolCalls = append(c.result.ToolCalls, ToolCallRecord{
		ToolName: toolName,
		Input:    input,
		Output:   output,
		Success:  success,
		Duration: duration,
	})
}

// SetIntent 设置意图.
func (c *LearningCallback) SetIntent(intent IntentType) {
	c.result.Intent = intent
}

// Complete 完成执行并触发学习.
func (c *LearningCallback) Complete(ctx context.Context, finalAnswer string, success bool) error {
	c.result.FinalAnswer = finalAnswer
	c.result.Success = success
	c.result.Duration = time.Since(c.startTime)

	return c.manager.LearnFromExecution(ctx, c.result)
}

// GetResult 获取执行结果.
func (c *LearningCallback) GetResult() *ExecutionResult {
	return c.result
}
