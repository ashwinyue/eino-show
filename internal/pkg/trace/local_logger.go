// Package trace 提供本地追踪日志，格式与 coze-loop 统一
package trace

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/eino-show/internal/pkg/log"
)

// TraceLog 追踪日志结构（与 coze-loop tracespec 格式统一）
type TraceLog struct {
	TraceID   string         `json:"trace_id"`
	SpanID    string         `json:"span_id"`
	ParentID  string         `json:"parent_id,omitempty"`
	Name      string         `json:"name"`
	Component string         `json:"component"`
	Type      string         `json:"type"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time,omitempty"`
	Duration  time.Duration  `json:"duration_ms,omitempty"`
	Input     *TraceInput    `json:"input,omitempty"`
	Output    *TraceOutput   `json:"output,omitempty"`
	Error     string         `json:"error,omitempty"`
	Tags      map[string]any `json:"tags,omitempty"`
}

// TraceInput 输入数据（与 coze-loop ModelInput 格式统一）
type TraceInput struct {
	Messages []TraceMessage `json:"messages,omitempty"`
	Tools    []TraceTool    `json:"tools,omitempty"`
}

// TraceOutput 输出数据（与 coze-loop ModelOutput 格式统一）
type TraceOutput struct {
	Message      *TraceMessage `json:"message,omitempty"`
	FinishReason string        `json:"finish_reason,omitempty"`
	TokenUsage   *TokenUsage   `json:"token_usage,omitempty"`
}

// TraceMessage 消息格式（与 coze-loop ModelMessage 格式统一）
type TraceMessage struct {
	Role             string          `json:"role"`
	Content          string          `json:"content"`
	Name             string          `json:"name,omitempty"`
	ToolCalls        []TraceToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string          `json:"tool_call_id,omitempty"`
	ReasoningContent string          `json:"reasoning_content,omitempty"`
}

// TraceTool 工具定义
type TraceTool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// TraceToolCall 工具调用
type TraceToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// TokenUsage Token 使用量
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// spanInfo 存储 span 信息
type spanInfo struct {
	traceID   string
	spanID    string
	parentID  string
	name      string
	component string
	spanType  string
	startTime time.Time
	input     *TraceInput
}

type ctxKey struct{}

// LocalLoggerHandler 本地日志回调处理器
// 使用项目的 zap logger 实现统一的日志管理
type LocalLoggerHandler struct {
	mu      sync.Mutex
	spans   map[string]*spanInfo
	counter int64
	logger  log.Logger // zap logger 实例
	detail  bool       // 是否输出详细信息
}

// NewLocalLoggerHandler 创建本地日志回调处理器
// 使用项目的 zap 基础设施，输出到独立文件 logs/trace.log
func NewLocalLoggerHandler() *LocalLoggerHandler {
	// 从环境变量读取配置
	logPath := os.Getenv("TRACE_LOG_FILE")
	if logPath == "" {
		logPath = "logs/trace.log"
	}

	detail := os.Getenv("TRACE_LOG_DETAIL") == "true"

	// 创建独立的 trace logger（不影响主 logger）
	traceLogger := log.New(&log.Options{
		Level:             "info", // 追踪日志固定为 info 级别
		Format:            "json", // JSON 格式便于解析
		OutputPaths:       []string{logPath},
		DisableCaller:     true, // 追踪日志不需要调用位置
		DisableStacktrace: true, // 追踪日志不需要堆栈
	})

	return &LocalLoggerHandler{
		spans:  make(map[string]*spanInfo),
		logger: traceLogger,
		detail: detail,
	}
}

func (h *LocalLoggerHandler) generateID() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.counter++
	return time.Now().Format("20060102150405") + "-" + string(rune('0'+h.counter%10))
}

func (h *LocalLoggerHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if info == nil {
		return ctx
	}

	spanID := h.generateID()
	traceID := spanID // 简化处理，实际应从 context 获取

	si := &spanInfo{
		traceID:   traceID,
		spanID:    spanID,
		name:      info.Name,
		component: string(info.Component),
		spanType:  info.Type,
		startTime: time.Now(),
	}

	// 解析输入
	si.input = h.parseInput(input)

	h.mu.Lock()
	h.spans[spanID] = si
	h.mu.Unlock()

	// 输出开始日志
	traceLog := &TraceLog{
		TraceID:   si.traceID,
		SpanID:    si.spanID,
		Name:      si.name,
		Component: si.component,
		Type:      si.spanType,
		StartTime: si.startTime,
		Input:     si.input,
	}

	h.logTrace("TRACE_START", traceLog)

	return context.WithValue(ctx, ctxKey{}, spanID)
}

func (h *LocalLoggerHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if info == nil {
		return ctx
	}

	spanID, _ := ctx.Value(ctxKey{}).(string)

	h.mu.Lock()
	si := h.spans[spanID]
	delete(h.spans, spanID)
	h.mu.Unlock()

	if si == nil {
		return ctx
	}

	endTime := time.Now()
	traceLog := &TraceLog{
		TraceID:   si.traceID,
		SpanID:    si.spanID,
		Name:      si.name,
		Component: si.component,
		Type:      si.spanType,
		StartTime: si.startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(si.startTime) / time.Millisecond,
		Input:     si.input,
		Output:    h.parseOutput(output),
	}

	h.logTrace("TRACE_END", traceLog)

	return ctx
}

func (h *LocalLoggerHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	if info == nil {
		return ctx
	}

	spanID, _ := ctx.Value(ctxKey{}).(string)

	h.mu.Lock()
	si := h.spans[spanID]
	delete(h.spans, spanID)
	h.mu.Unlock()

	if si == nil {
		return ctx
	}

	endTime := time.Now()
	traceLog := &TraceLog{
		TraceID:   si.traceID,
		SpanID:    si.spanID,
		Name:      si.name,
		Component: si.component,
		Type:      si.spanType,
		StartTime: si.startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(si.startTime) / time.Millisecond,
		Input:     si.input,
		Error:     err.Error(),
	}

	h.logTrace("TRACE_ERROR", traceLog)

	return ctx
}

func (h *LocalLoggerHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	// 异步消费流式输入，避免阻塞
	go func() {
		defer input.Close()
		for {
			_, err := input.Recv()
			if err != nil {
				break
			}
		}
	}()
	return h.OnStart(ctx, info, nil)
}

func (h *LocalLoggerHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	// 异步消费流式输出，避免阻塞
	go func() {
		defer output.Close()
		for {
			_, err := output.Recv()
			if err != nil {
				break
			}
		}
	}()
	return h.OnEnd(ctx, info, nil)
}

func (h *LocalLoggerHandler) parseInput(input callbacks.CallbackInput) *TraceInput {
	if input == nil {
		return nil
	}

	ti := &TraceInput{}

	switch v := input.(type) {
	case *model.CallbackInput:
		if v != nil {
			ti.Messages = make([]TraceMessage, 0, len(v.Messages))
			for _, msg := range v.Messages {
				ti.Messages = append(ti.Messages, h.convertMessage(msg))
			}
			ti.Tools = make([]TraceTool, 0, len(v.Tools))
			for _, tool := range v.Tools {
				if tool != nil {
					ti.Tools = append(ti.Tools, TraceTool{
						Name:        tool.Name,
						Description: tool.Desc,
					})
				}
			}
		}
	}

	return ti
}

func (h *LocalLoggerHandler) parseOutput(output callbacks.CallbackOutput) *TraceOutput {
	if output == nil {
		return nil
	}

	to := &TraceOutput{}

	switch v := output.(type) {
	case *model.CallbackOutput:
		if v != nil && v.Message != nil {
			msg := h.convertMessage(v.Message)
			to.Message = &msg
			if v.Message.ResponseMeta != nil {
				to.FinishReason = v.Message.ResponseMeta.FinishReason
				if v.Message.ResponseMeta.Usage != nil {
					to.TokenUsage = &TokenUsage{
						PromptTokens:     v.Message.ResponseMeta.Usage.PromptTokens,
						CompletionTokens: v.Message.ResponseMeta.Usage.CompletionTokens,
						TotalTokens:      v.Message.ResponseMeta.Usage.TotalTokens,
					}
				}
			}
		}
	}

	return to
}

func (h *LocalLoggerHandler) convertMessage(msg *schema.Message) TraceMessage {
	if msg == nil {
		return TraceMessage{}
	}

	tm := TraceMessage{
		Role:             string(msg.Role),
		Content:          msg.Content,
		Name:             msg.Name,
		ToolCallID:       msg.ToolCallID,
		ReasoningContent: msg.ReasoningContent,
	}

	for _, tc := range msg.ToolCalls {
		tm.ToolCalls = append(tm.ToolCalls, TraceToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return tm
}

func (h *LocalLoggerHandler) logTrace(event string, trace *TraceLog) {
	// 使用 zap 的结构化日志
	// 简洁模式：只记录关键字段
	kvs := []any{
		"event", event,
		"trace_id", trace.TraceID,
		"span_id", trace.SpanID,
		"component", trace.Component,
		"type", trace.Type,
		"name", trace.Name,
		"start_time", trace.StartTime.Format(time.RFC3339Nano),
	}

	// 如果有结束时间，添加持续时长
	if !trace.EndTime.IsZero() {
		kvs = append(kvs, "duration_ms", trace.Duration)
		kvs = append(kvs, "end_time", trace.EndTime.Format(time.RFC3339Nano))
	}

	// 如果有错误，记录错误信息
	if trace.Error != "" {
		kvs = append(kvs, "error", trace.Error)
		h.logger.Errorw("[TRACE] "+event, kvs...)
		return
	}

	// 详细模式：记录完整的 input/output
	if h.detail {
		if trace.Input != nil {
			kvs = append(kvs, "input", trace.Input)
		}
		if trace.Output != nil {
			kvs = append(kvs, "output", trace.Output)
		}
	}

	h.logger.Infow("[TRACE] "+event, kvs...)
}

// 确保实现 callbacks.Handler 接口
var _ callbacks.Handler = (*LocalLoggerHandler)(nil)
