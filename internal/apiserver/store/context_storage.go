// Package store provides database-backed context storage for LLM context management.
package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/pkg/llmcontext"
)

// DBContextStorage 数据库上下文存储.
type DBContextStorage struct {
	store *datastore
}

// NewDBContextStorage 创建数据库上下文存储.
func NewDBContextStorage(store *datastore) llmcontext.ContextStorage {
	return &DBContextStorage{store: store}
}

// newDBContextStorage 内部创建方法.
func newDBContextStorage(store *datastore) *DBContextStorage {
	return &DBContextStorage{store: store}
}

// Ensure DBContextStorage implements ContextStorage
var _ llmcontext.ContextStorage = (*DBContextStorage)(nil)

// Load 从数据库加载会话的消息历史.
func (s *DBContextStorage) Load(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	var messageMs []*model.MessageM
	err := s.store.DB(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messageMs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}

	messages := make([]*schema.Message, 0, len(messageMs))
	for _, m := range messageMs {
		msg := convertToSchemaMessage(m)
		if msg != nil {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// Save 保存会话的消息历史到数据库.
// 注意：这会替换所有现有消息（用于压缩后的更新）
func (s *DBContextStorage) Save(ctx context.Context, sessionID string, messages []*schema.Message) error {
	return s.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除现有消息 (软删除)
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.MessageM{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing messages: %w", err)
		}

		// 2. 插入新消息
		for i, msg := range messages {
			messageM := convertFromSchemaMessage(sessionID, msg, i)
			if messageM == nil {
				continue
			}
			if err := tx.Create(messageM).Error; err != nil {
				return fmt.Errorf("failed to save message: %w", err)
			}
		}

		return nil
	})
}

// Delete 删除会话的消息历史.
func (s *DBContextStorage) Delete(ctx context.Context, sessionID string) error {
	err := s.store.DB(ctx).
		Where("session_id = ?", sessionID).
		Delete(&model.MessageM{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	return nil
}

// convertToSchemaMessage 将数据库消息转换为 schema.Message.
func convertToSchemaMessage(m *model.MessageM) *schema.Message {
	var role schema.RoleType
	switch m.Role {
	case "user":
		role = schema.User
	case "assistant":
		role = schema.Assistant
	case "system":
		role = schema.System
	case "tool":
		role = schema.Tool
	default:
		role = schema.User
	}

	msg := &schema.Message{
		Role:    role,
		Content: m.Content,
	}

	// 解析 agent_steps 中的 tool_calls (如果有)
	if m.AgentSteps != nil && *m.AgentSteps != "" {
		var steps struct {
			ToolCalls []schema.ToolCall `json:"tool_calls,omitempty"`
		}
		if json.Unmarshal([]byte(*m.AgentSteps), &steps) == nil && len(steps.ToolCalls) > 0 {
			msg.ToolCalls = steps.ToolCalls
		}
	}

	return msg
}

// convertFromSchemaMessage 将 schema.Message 转换为数据库消息.
func convertFromSchemaMessage(sessionID string, msg *schema.Message, order int) *model.MessageM {
	var role string
	switch msg.Role {
	case schema.User:
		role = "user"
	case schema.Assistant:
		role = "assistant"
	case schema.System:
		role = "system"
	case schema.Tool:
		role = "tool"
	default:
		role = "user"
	}

	messageM := &model.MessageM{
		SessionID:   sessionID,
		RequestID:   uuid.New().String(),
		Role:        role,
		Content:     msg.Content,
		IsCompleted: true,
	}

	// 保存 tool_calls 到 agent_steps
	if len(msg.ToolCalls) > 0 {
		steps := struct {
			ToolCalls []schema.ToolCall `json:"tool_calls,omitempty"`
		}{
			ToolCalls: msg.ToolCalls,
		}
		if data, err := json.Marshal(steps); err == nil {
			stepsStr := string(data)
			messageM.AgentSteps = &stepsStr
		}
	}

	return messageM
}
