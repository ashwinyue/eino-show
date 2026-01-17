// Package chat 提供 Chat Agent 封装，基于 Eino ChatModel.
package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/eino-show/internal/pkg/trace"
)

// Config Chat Agent 配置.
type Config struct {
	// ChatModel 对话模型
	ChatModel model.ChatModel

	// SystemPrompt 系统提示词（可选）
	SystemPrompt string

	// Temperature 温度参数（可选）
	Temperature *float32

	// TopP Top-P 采样参数（可选）
	TopP *float32

	// MaxTokens 最大生成 token 数（可选）
	MaxTokens *int
}

// Agent Chat Agent，提供纯对话功能.
type Agent struct {
	cfg          *Config
	chatModel    model.ChatModel
	systemPrompt string
}

// NewAgent 创建 Chat Agent.
func NewAgent(ctx context.Context, cfg *Config) (*Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}

	return &Agent{
		cfg:          cfg,
		chatModel:    cfg.ChatModel,
		systemPrompt: cfg.SystemPrompt,
	}, nil
}

// NewSimpleAgent 创建简单的 Chat Agent（使用默认配置）.
func NewSimpleAgent(ctx context.Context, chatModel model.ChatModel) (*Agent, error) {
	return NewAgent(ctx, &Config{
		ChatModel: chatModel,
	})
}

// Generate 生成回复（非流式）.
func (a *Agent) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 如果有系统提示词，添加到消息列表
	messages = a.prependSystemPrompt(messages)

	// 应用配置选项
	opts = a.applyConfigOptions(opts)

	// 记录追踪日志 - 开始
	startTime := time.Now()
	trace.LogChatStart(len(messages))

	// 调用 ChatModel 生成回复
	msg, err := a.chatModel.Generate(ctx, messages, opts...)
	if err != nil {
		// 记录追踪日志 - 错误
		trace.LogChatError(err, time.Since(startTime))
		return nil, err
	}

	// 记录追踪日志 - 完成
	trace.LogChatEnd(msg.Content, time.Since(startTime))

	return msg, nil
}

// StreamGenerate 生成回复（流式）.
func (a *Agent) StreamGenerate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 如果有系统提示词，添加到消息列表
	messages = a.prependSystemPrompt(messages)

	// 应用配置选项
	opts = a.applyConfigOptions(opts)

	// 调用 ChatModel 生成流式回复
	return a.chatModel.Stream(ctx, messages, opts...)
}

// Chat 简化的对话接口.
func (a *Agent) Chat(ctx context.Context, userMessage string, opts ...model.Option) (*schema.Message, error) {
	messages := []*schema.Message{
		schema.UserMessage(userMessage),
	}
	return a.Generate(ctx, messages, opts...)
}

// prependSystemPrompt 预置系统提示词.
func (a *Agent) prependSystemPrompt(messages []*schema.Message) []*schema.Message {
	if a.systemPrompt == "" {
		return messages
	}

	// 检查是否已有系统消息
	for _, msg := range messages {
		if msg.Role == schema.System {
			return messages
		}
	}

	// 在开头添加系统消息
	result := make([]*schema.Message, 0, len(messages)+1)
	result = append(result, schema.SystemMessage(a.systemPrompt))
	result = append(result, messages...)
	return result
}

// applyConfigOptions 应用配置选项.
func (a *Agent) applyConfigOptions(opts []model.Option) []model.Option {
	if a.cfg.Temperature != nil {
		opts = append(opts, model.WithTemperature(*a.cfg.Temperature))
	}
	if a.cfg.TopP != nil {
		opts = append(opts, model.WithTopP(*a.cfg.TopP))
	}
	if a.cfg.MaxTokens != nil {
		opts = append(opts, model.WithMaxTokens(*a.cfg.MaxTokens))
	}
	return opts
}
