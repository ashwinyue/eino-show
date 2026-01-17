// Package llmcontext 提供 LLM 上下文管理，对齐 WeKnora 实现.
package llmcontext

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// ContextStorage 上下文存储接口.
type ContextStorage interface {
	// Load 加载会话的消息历史
	Load(ctx context.Context, sessionID string) ([]*schema.Message, error)

	// Save 保存会话的消息历史
	Save(ctx context.Context, sessionID string, messages []*schema.Message) error

	// Delete 删除会话的消息历史
	Delete(ctx context.Context, sessionID string) error
}
