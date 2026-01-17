// Package router provides tool call interception for experience extraction.
// Reference: AssistantAgent Tool Interceptor
package router

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ToolInterceptor 工具调用拦截器.
// 用于拦截工具调用，提取经验和模式.
type ToolInterceptor struct {
	cfg       *InterceptorConfig
	listeners []ToolCallListener
	mu        sync.RWMutex
}

// InterceptorConfig 拦截器配置.
type InterceptorConfig struct {
	// ExperienceManager 经验管理器
	ExperienceManager *ExperienceManager

	// LearningManager 学习管理器 (可选)
	LearningManager *LearningManager

	// EnableAutoLearn 是否启用自动学习
	EnableAutoLearn bool

	// MinDuration 最小执行时间 (过滤过快的调用)
	MinDuration time.Duration

	// MaxDuration 最大执行时间 (过滤超时调用)
	MaxDuration time.Duration

	// SuccessOnly 是否只学习成功的调用
	SuccessOnly bool
}

// ToolCallListener 工具调用监听器.
type ToolCallListener interface {
	// OnToolCall 工具调用前
	OnToolCall(ctx context.Context, call *ToolCallEvent) error
	// OnToolResult 工具调用后
	OnToolResult(ctx context.Context, result *ToolResultEvent) error
}

// ToolCallEvent 工具调用事件.
type ToolCallEvent struct {
	SessionID  string
	RequestID  string
	ToolName   string
	ToolCallID string
	Arguments  string
	Iteration  int
	Timestamp  time.Time
	Metadata   map[string]any
}

// ToolResultEvent 工具结果事件.
type ToolResultEvent struct {
	SessionID  string
	RequestID  string
	ToolName   string
	ToolCallID string
	Arguments  string
	Output     string
	Success    bool
	Error      string
	Duration   time.Duration
	Iteration  int
	Timestamp  time.Time
	Metadata   map[string]any
}

// NewToolInterceptor 创建工具拦截器.
func NewToolInterceptor(cfg *InterceptorConfig) *ToolInterceptor {
	if cfg == nil {
		cfg = &InterceptorConfig{}
	}

	// 默认配置
	if cfg.MinDuration == 0 {
		cfg.MinDuration = 10 * time.Millisecond
	}
	if cfg.MaxDuration == 0 {
		cfg.MaxDuration = 5 * time.Minute
	}

	return &ToolInterceptor{
		cfg:       cfg,
		listeners: make([]ToolCallListener, 0),
	}
}

// AddListener 添加监听器.
func (i *ToolInterceptor) AddListener(listener ToolCallListener) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.listeners = append(i.listeners, listener)
}

// RemoveListener 移除监听器.
func (i *ToolInterceptor) RemoveListener(listener ToolCallListener) {
	i.mu.Lock()
	defer i.mu.Unlock()
	for idx, l := range i.listeners {
		if l == listener {
			i.listeners = append(i.listeners[:idx], i.listeners[idx+1:]...)
			break
		}
	}
}

// OnToolCall 拦截工具调用.
func (i *ToolInterceptor) OnToolCall(ctx context.Context, call *ToolCallEvent) error {
	i.mu.RLock()
	listeners := make([]ToolCallListener, len(i.listeners))
	copy(listeners, i.listeners)
	i.mu.RUnlock()

	for _, l := range listeners {
		if err := l.OnToolCall(ctx, call); err != nil {
			// 记录错误但不中断
			continue
		}
	}
	return nil
}

// OnToolResult 拦截工具结果.
func (i *ToolInterceptor) OnToolResult(ctx context.Context, result *ToolResultEvent) error {
	i.mu.RLock()
	listeners := make([]ToolCallListener, len(i.listeners))
	copy(listeners, i.listeners)
	i.mu.RUnlock()

	for _, l := range listeners {
		if err := l.OnToolResult(ctx, result); err != nil {
			continue
		}
	}

	// 自动学习
	if i.cfg.EnableAutoLearn && i.shouldLearn(result) {
		go i.extractExperience(ctx, result)
	}

	return nil
}

// shouldLearn 判断是否应该学习.
func (i *ToolInterceptor) shouldLearn(result *ToolResultEvent) bool {
	// 只学习成功的
	if i.cfg.SuccessOnly && !result.Success {
		return false
	}

	// 执行时间过滤
	if result.Duration < i.cfg.MinDuration {
		return false
	}
	if result.Duration > i.cfg.MaxDuration {
		return false
	}

	return true
}

// extractExperience 从工具调用中提取经验.
func (i *ToolInterceptor) extractExperience(ctx context.Context, result *ToolResultEvent) {
	if i.cfg.ExperienceManager == nil {
		return
	}

	// 构建经验
	exp := &Experience{
		ID:        fmt.Sprintf("tool_%s_%d", result.ToolCallID, time.Now().UnixNano()),
		Type:      ExperienceTypeReact,
		Query:     result.Arguments,
		Response:  result.Output,
		ToolCalls: []string{result.ToolName},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]any{
			"tool_name":    result.ToolName,
			"tool_call_id": result.ToolCallID,
			"duration_ms":  result.Duration.Milliseconds(),
			"success":      result.Success,
			"session_id":   result.SessionID,
		},
	}

	// 保存经验
	_ = i.cfg.ExperienceManager.Learn(ctx, exp)
}

// ExperienceLearningListener 经验学习监听器.
type ExperienceLearningListener struct {
	experienceManager *ExperienceManager
	sessionTools      sync.Map // sessionID -> []ToolResultEvent
}

// NewExperienceLearningListener 创建经验学习监听器.
func NewExperienceLearningListener(em *ExperienceManager) *ExperienceLearningListener {
	return &ExperienceLearningListener{
		experienceManager: em,
	}
}

// OnToolCall 工具调用前 (记录开始).
func (l *ExperienceLearningListener) OnToolCall(ctx context.Context, call *ToolCallEvent) error {
	// 可以在这里记录调用开始时间等
	return nil
}

// OnToolResult 工具调用后 (收集结果).
func (l *ExperienceLearningListener) OnToolResult(ctx context.Context, result *ToolResultEvent) error {
	// 收集会话的工具调用结果
	key := result.SessionID
	if key == "" {
		return nil
	}

	var tools []ToolResultEvent
	if v, ok := l.sessionTools.Load(key); ok {
		tools = v.([]ToolResultEvent)
	}
	tools = append(tools, *result)
	l.sessionTools.Store(key, tools)

	return nil
}

// GetSessionToolCalls 获取会话的所有工具调用.
func (l *ExperienceLearningListener) GetSessionToolCalls(sessionID string) []ToolResultEvent {
	if v, ok := l.sessionTools.Load(sessionID); ok {
		return v.([]ToolResultEvent)
	}
	return nil
}

// ClearSession 清理会话数据.
func (l *ExperienceLearningListener) ClearSession(sessionID string) {
	l.sessionTools.Delete(sessionID)
}

// LearnFromSession 从会话的工具调用中学习.
func (l *ExperienceLearningListener) LearnFromSession(ctx context.Context, sessionID, query, answer string) error {
	tools := l.GetSessionToolCalls(sessionID)
	if len(tools) == 0 {
		return nil
	}

	// 检查是否全部成功
	allSuccess := true
	var toolNames []string
	for _, t := range tools {
		if !t.Success {
			allSuccess = false
			break
		}
		toolNames = append(toolNames, t.ToolName)
	}

	if !allSuccess {
		l.ClearSession(sessionID)
		return nil
	}

	// 构建经验
	exp := &Experience{
		ID:        fmt.Sprintf("session_%s_%d", sessionID, time.Now().UnixNano()),
		Type:      ExperienceTypeReact,
		Query:     query,
		Response:  answer,
		ToolCalls: toolNames,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]any{
			"session_id": sessionID,
			"tool_count": len(tools),
			"source":     "tool_interceptor",
		},
	}

	// 保存
	err := l.experienceManager.Learn(ctx, exp)

	// 清理
	l.ClearSession(sessionID)

	return err
}

// PatternExtractionListener 模式提取监听器.
// 用于识别重复的工具调用模式.
type PatternExtractionListener struct {
	patterns sync.Map // pattern_key -> count
}

// NewPatternExtractionListener 创建模式提取监听器.
func NewPatternExtractionListener() *PatternExtractionListener {
	return &PatternExtractionListener{}
}

// OnToolCall 工具调用前.
func (l *PatternExtractionListener) OnToolCall(ctx context.Context, call *ToolCallEvent) error {
	return nil
}

// OnToolResult 工具调用后 (分析模式).
func (l *PatternExtractionListener) OnToolResult(ctx context.Context, result *ToolResultEvent) error {
	if !result.Success {
		return nil
	}

	// 构建模式 key
	patternKey := fmt.Sprintf("%s:%s", result.ToolName, hashArguments(result.Arguments))

	// 计数
	if v, ok := l.patterns.Load(patternKey); ok {
		count := v.(int) + 1
		l.patterns.Store(patternKey, count)
	} else {
		l.patterns.Store(patternKey, 1)
	}

	return nil
}

// GetFrequentPatterns 获取高频模式.
func (l *PatternExtractionListener) GetFrequentPatterns(minCount int) map[string]int {
	result := make(map[string]int)
	l.patterns.Range(func(key, value interface{}) bool {
		count := value.(int)
		if count >= minCount {
			result[key.(string)] = count
		}
		return true
	})
	return result
}

// hashArguments 简单哈希参数 (用于模式识别).
func hashArguments(args string) string {
	// 简单实现：取前 32 字符
	if len(args) > 32 {
		return args[:32]
	}
	return args
}

// MetricsListener 指标监听器.
type MetricsListener struct {
	totalCalls    int64
	successCalls  int64
	failedCalls   int64
	totalDuration time.Duration
	mu            sync.Mutex
}

// NewMetricsListener 创建指标监听器.
func NewMetricsListener() *MetricsListener {
	return &MetricsListener{}
}

// OnToolCall 工具调用前.
func (l *MetricsListener) OnToolCall(ctx context.Context, call *ToolCallEvent) error {
	l.mu.Lock()
	l.totalCalls++
	l.mu.Unlock()
	return nil
}

// OnToolResult 工具调用后.
func (l *MetricsListener) OnToolResult(ctx context.Context, result *ToolResultEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.totalDuration += result.Duration
	if result.Success {
		l.successCalls++
	} else {
		l.failedCalls++
	}
	return nil
}

// GetMetrics 获取指标.
func (l *MetricsListener) GetMetrics() map[string]interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	avgDuration := time.Duration(0)
	completed := l.successCalls + l.failedCalls
	if completed > 0 {
		avgDuration = l.totalDuration / time.Duration(completed)
	}

	return map[string]interface{}{
		"total_calls":     l.totalCalls,
		"success_calls":   l.successCalls,
		"failed_calls":    l.failedCalls,
		"avg_duration_ms": avgDuration.Milliseconds(),
		"success_rate":    float64(l.successCalls) / float64(max(completed, 1)),
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
