// Package session 提供 LLM 驱动的标题生成功能.
package session

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"

	chatagent "github.com/ashwinyue/eino-show/internal/pkg/agent/chat"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// 标题生成提示词（对齐 WeKnora）
const generateTitlePrompt = `You are a helpful assistant that generates concise conversation titles.

Based on the user's first message, generate a short, descriptive title for this conversation.

Rules:
1. The title should be 3-10 words
2. The title should capture the main topic or intent
3. Do NOT include quotes around the title
4. Do NOT include "Title:" prefix
5. Use the same language as the user's message
6. Be concise and specific

Just output the title directly, nothing else.`

// generateTitleWithLLM 使用 LLM 生成会话标题.
func (b *sessionBiz) generateTitleWithLLM(ctx context.Context, userMessage string) (string, error) {
	if b.qaExecutor == nil || b.qaExecutor.factory == nil {
		// 回退到简单截取
		return truncateTitle(userMessage, 50), nil
	}

	// 创建 Chat Agent
	chatCfg := &chatagent.Config{
		SystemPrompt: generateTitlePrompt,
	}

	chatAgent, err := b.qaExecutor.factory.CreateChatAgent(ctx, chatCfg)
	if err != nil {
		// LLM 创建失败，回退到简单截取
		return truncateTitle(userMessage, 50), nil
	}

	// 构建消息
	messages := []*schema.Message{
		schema.UserMessage(userMessage + " /no_think"),
	}

	// 生成标题
	msg, err := chatAgent.Generate(ctx, messages)
	if err != nil {
		// 生成失败，回退到简单截取
		return truncateTitle(userMessage, 50), nil
	}

	// 清理生成的标题
	title := cleanTitle(msg.Content)
	if title == "" {
		return truncateTitle(userMessage, 50), nil
	}

	return title, nil
}

// cleanTitle 清理 LLM 生成的标题.
func cleanTitle(raw string) string {
	title := strings.TrimSpace(raw)

	// 移除可能的 think 标签
	if idx := strings.Index(title, "</think>"); idx != -1 {
		title = strings.TrimSpace(title[idx+8:])
	}
	title = strings.TrimPrefix(title, "<think>")

	// 移除引号
	title = strings.Trim(title, "\"'`")

	// 移除常见前缀
	prefixes := []string{"Title:", "标题:", "title:", "TITLE:"}
	for _, p := range prefixes {
		title = strings.TrimPrefix(title, p)
	}
	title = strings.TrimSpace(title)

	// 限制长度
	if len([]rune(title)) > 100 {
		title = string([]rune(title)[:100])
	}

	return title
}

// truncateTitle 简单截取标题.
func truncateTitle(content string, maxLen int) string {
	// 移除换行
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	content = strings.TrimSpace(content)

	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}

// generateTitleAsync 异步生成标题并更新会话.
func (b *sessionBiz) generateTitleAsync(ctx context.Context, sessionID, userMessage string) {
	go func() {
		title, err := b.generateTitleWithLLM(ctx, userMessage)
		if err != nil || title == "" {
			return
		}

		// 更新会话标题
		sessionM, err := b.store.Session().Get(ctx, where.F("id", sessionID))
		if err != nil {
			return
		}

		// 只有在没有标题时才更新
		if sessionM.Title != nil && *sessionM.Title != "" {
			return
		}

		sessionM.Title = &title
		_ = b.store.Session().Update(ctx, sessionM)
	}()
}
