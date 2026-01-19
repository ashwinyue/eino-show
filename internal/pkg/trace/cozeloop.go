// Package trace 提供 Trace 追踪功能
// 集成 coze-loop 用于观测 Agent 执行过程
package trace

import (
	"context"
	"os"
	"time"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"

	"github.com/ashwinyue/eino-show/pkg/log"
)

// CloseFn 关闭函数类型
type CloseFn func(ctx context.Context)

// EndSpanFn 结束 Span 函数类型
type EndSpanFn func(ctx context.Context, output any)

// StartSpanFn 开始 Span 函数类型
type StartSpanFn func(ctx context.Context, name string, input any) (nCtx context.Context, endFn EndSpanFn)

// CozeLoopConfig coze-loop 配置
type CozeLoopConfig struct {
	WorkspaceID string // COZELOOP_WORKSPACE_ID
	APIToken    string // COZELOOP_API_TOKEN
}

// LoadCozeLoopConfigFromEnv 从环境变量加载配置
func LoadCozeLoopConfigFromEnv() *CozeLoopConfig {
	return &CozeLoopConfig{
		WorkspaceID: os.Getenv("COZELOOP_WORKSPACE_ID"),
		APIToken:    os.Getenv("COZELOOP_API_TOKEN"),
	}
}

// IsConfigured 检查是否已配置
func (c *CozeLoopConfig) IsConfigured() bool {
	return c.WorkspaceID != "" && c.APIToken != ""
}

// InitCozeLoop 初始化 coze-loop 追踪
// 返回关闭函数和创建 Span 的函数
// 如果未配置环境变量，返回空操作函数
func InitCozeLoop(ctx context.Context) (closeFn CloseFn, startSpanFn StartSpanFn) {
	cfg := LoadCozeLoopConfigFromEnv()
	return InitCozeLoopWithConfig(ctx, cfg)
}

// InitCozeLoopWithConfig 使用配置初始化 coze-loop
func InitCozeLoopWithConfig(_ context.Context, cfg *CozeLoopConfig) (closeFn CloseFn, startSpanFn StartSpanFn) {
	// 未配置时返回空操作
	if !cfg.IsConfigured() {
		log.Infow("coze-loop not configured, tracing disabled")
		return noopCloseFn, buildStartSpanFn(nil)
	}

	// 创建 coze-loop 客户端
	client, err := cozeloop.NewClient(
		cozeloop.WithWorkspaceID(cfg.WorkspaceID),
		cozeloop.WithAPIToken(cfg.APIToken),
	)
	if err != nil {
		log.Errorw(err, "failed to create coze-loop client")
		return noopCloseFn, buildStartSpanFn(nil)
	}

	// 注册全局回调处理器
	handler := ccb.NewLoopHandler(client)
	callbacks.AppendGlobalHandlers(handler)

	log.Infow("coze-loop tracing enabled", "workspace_id", cfg.WorkspaceID)

	return client.Close, buildStartSpanFn(client)
}

// noopCloseFn 空操作关闭函数
func noopCloseFn(_ context.Context) {}

// buildStartSpanFn 构建开始 Span 函数
func buildStartSpanFn(client cozeloop.Client) StartSpanFn {
	return func(ctx context.Context, name string, input any) (nCtx context.Context, endFn EndSpanFn) {
		if client == nil {
			return ctx, noopEndSpanFn
		}

		nCtx, span := client.StartSpan(ctx, name, "custom")
		span.SetInput(ctx, input)
		return nCtx, buildEndSpanFn(span)
	}
}

// noopEndSpanFn 空操作结束 Span 函数
func noopEndSpanFn(_ context.Context, _ any) {}

// buildEndSpanFn 构建结束 Span 函数
func buildEndSpanFn(span cozeloop.Span) EndSpanFn {
	return func(ctx context.Context, output any) {
		if span == nil {
			return
		}
		span.SetOutput(ctx, output)
		span.Finish(ctx)
	}
}

// Tracer 全局追踪器
type Tracer struct {
	client      cozeloop.Client
	startSpanFn StartSpanFn
	closeFn     CloseFn
}

// globalTracer 全局追踪器实例
var globalTracer *Tracer

// globalHandlers 全局回调处理器列表
var globalHandlers []callbacks.Handler

// Init 初始化全局追踪器
func Init(ctx context.Context) {
	closeFn, startSpanFn := InitCozeLoop(ctx)
	globalTracer = &Tracer{
		startSpanFn: startSpanFn,
		closeFn:     closeFn,
	}

	// 同时注册本地日志回调（输出到文件，格式与 coze-loop 统一）
	localLogger := NewLocalLoggerHandler()
	globalHandlers = append(globalHandlers, localLogger)
	callbacks.AppendGlobalHandlers(localLogger)

	// 输出日志配置信息
	logPath := os.Getenv("TRACE_LOG_FILE")
	if logPath == "" {
		logPath = "logs/trace.log"
	}
	log.Infow("local trace logger enabled",
		"log_file", logPath,
		"detail", os.Getenv("TRACE_LOG_DETAIL") == "true",
		"debug", os.Getenv("DEBUG") == "true",
	)
}

// GetHandlers 获取全局回调处理器
func GetHandlers() []callbacks.Handler {
	return globalHandlers
}

// Close 关闭全局追踪器
func Close(ctx context.Context) {
	if globalTracer != nil && globalTracer.closeFn != nil {
		globalTracer.closeFn(ctx)
	}
}

// StartSpan 开始一个 Span
func StartSpan(ctx context.Context, name string, input any) (context.Context, EndSpanFn) {
	if globalTracer == nil || globalTracer.startSpanFn == nil {
		return ctx, noopEndSpanFn
	}
	return globalTracer.startSpanFn(ctx, name, input)
}

// ===== 简化的日志函数（避免手写业务日志） =====

// LogChatStart 记录 Chat 开始
func LogChatStart(messageCount int) {
	log.Infow("[TRACE] CHAT_START", "messages", messageCount)
}

// LogChatEnd 记录 Chat 完成
func LogChatEnd(response string, duration time.Duration) {
	// 截断长响应
	if len(response) > 200 {
		response = response[:200] + "..."
	}
	log.Infow("[TRACE] CHAT_END", "duration_ms", duration.Milliseconds(), "response", response)
}

// LogChatError 记录 Chat 错误
func LogChatError(err error, duration time.Duration) {
	log.Errorw(err, "[TRACE] CHAT_ERROR", "duration_ms", duration.Milliseconds())
}
