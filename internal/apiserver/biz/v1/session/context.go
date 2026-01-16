// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package session 提供会话上下文管理.
package session

import (
	"encoding/json"
	"fmt"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/cloudwego/eino/schema"
)

// ContextManager 上下文管理器，负责管理对话历史和上下文压缩.
type ContextManager struct {
	session       *model.SessionM
	maxTokens     int
	contextConfig *v1.SessionContextConfig
}

// NewContextManager 创建上下文管理器.
func NewContextManager(session *model.SessionM) *ContextManager {
	maxMessages := 50 // 默认最大消息数
	cfg := parseContextConfig(session.ContextConfig)
	if cfg != nil && cfg.MaxMessages > 0 {
		maxMessages = int(cfg.MaxMessages)
	}

	return &ContextManager{
		session:       session,
		maxTokens:     maxMessages,
		contextConfig: cfg,
	}
}

// parseContextConfig 解析 JSON 字符串为 SessionContextConfig
func parseContextConfig(jsonStr *string) *v1.SessionContextConfig {
	if jsonStr == nil || *jsonStr == "" {
		return nil
	}
	var cfg v1.SessionContextConfig
	if err := json.Unmarshal([]byte(*jsonStr), &cfg); err != nil {
		return nil
	}
	return &cfg
}

// BuildMessages 构建用于 Agent 的消息列表.
// 包含系统提示词和历史对话消息.
func (m *ContextManager) BuildMessages(systemPrompt string, history []*schema.Message, userMessage string) []*schema.Message {
	messages := make([]*schema.Message, 0, len(history)+2)

	// 添加系统提示词
	if systemPrompt != "" {
		messages = append(messages, schema.SystemMessage(systemPrompt))
	}

	// 添加历史消息（压缩后）
	messages = append(messages, m.CompressHistory(history)...)

	// 添加当前用户消息
	messages = append(messages, schema.UserMessage(userMessage))

	return messages
}

// CompressHistory 压缩历史消息.
// 根据配置决定是否进行压缩和保留多少历史消息.
func (m *ContextManager) CompressHistory(history []*schema.Message) []*schema.Message {
	if len(history) == 0 {
		return history
	}

	// 如果启用了上下文压缩且超过阈值
	if m.shouldCompress() && len(history) > m.getCompressionThreshold() {
		return m.compress(history)
	}

	// 否则只保留最近的 N 条消息
	return m.truncate(history)
}

// shouldCompress 判断是否应该压缩.
func (m *ContextManager) shouldCompress() bool {
	return m.contextConfig != nil && m.contextConfig.EnableContextCompression
}

// getCompressionThreshold 获取压缩阈值.
func (m *ContextManager) getCompressionThreshold() int {
	if m.contextConfig != nil && m.contextConfig.CompressionThreshold > 0 {
		return int(m.contextConfig.CompressionThreshold)
	}
	return 20 // 默认阈值
}

// truncate 截断历史消息，保留最近的 N 条.
func (m *ContextManager) truncate(history []*schema.Message) []*schema.Message {
	maxSize := m.maxTokens
	if maxSize <= 0 {
		maxSize = 50
	}

	if len(history) <= maxSize {
		return history
	}

	// 保留最近的 maxSize 条消息
	return history[len(history)-maxSize:]
}

// compress 压缩历史消息.
// TODO: 实现实际的压缩逻辑（如使用 LLM 总结）.
func (m *ContextManager) compress(history []*schema.Message) []*schema.Message {
	// 简单实现：只保留最近的消息，并添加摘要
	threshold := m.getCompressionThreshold()

	// 保留系统消息
	var systemMessages []*schema.Message
	var otherMessages []*schema.Message
	for _, msg := range history {
		if msg.Role == schema.System {
			systemMessages = append(systemMessages, msg)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	// 保留最近的 threshold 条消息
	recent := otherMessages
	if len(otherMessages) > threshold {
		recent = otherMessages[len(otherMessages)-threshold:]

		// 添加压缩摘要
		summary := m.createSummary(otherMessages[:len(otherMessages)-threshold])
		if summary != "" {
			summaryMsg := schema.SystemMessage(
				"[对话摘要] " + summary + "\n以下是最近的对话：",
			)
			// 在系统消息后添加摘要
			result := make([]*schema.Message, 0, len(systemMessages)+1+len(recent))
			result = append(result, systemMessages...)
			result = append(result, summaryMsg)
			result = append(result, recent...)
			return result
		}
	}

	result := make([]*schema.Message, 0, len(systemMessages)+len(recent))
	result = append(result, systemMessages...)
	result = append(result, recent...)
	return result
}

// createSummary 创建历史消息的摘要.
// TODO: 使用 LLM 生成摘要.
func (m *ContextManager) createSummary(messages []*schema.Message) string {
	// 简单实现：返回消息数量
	return fmt.Sprintf("之前有 %d 条历史消息已被压缩", len(messages))
}

// GetMaxTokens 获取最大 token 数.
func (m *ContextManager) GetMaxTokens() int {
	return m.maxTokens
}

// SetMaxTokens 设置最大 token 数.
func (m *ContextManager) SetMaxTokens(maxTokens int) {
	m.maxTokens = maxTokens
}
