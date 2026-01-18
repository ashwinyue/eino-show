// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== Session 请求/响应类型 =====

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	AgentID     string `json:"agent_id"`
}

// SessionResponse 会话响应（对齐 WeKnora）
type SessionResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TenantID    uint64    `json:"tenant_id"`
	AgentID     string    `json:"agent_id,omitempty"` // 关联的智能体 ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	AgentID     *string `json:"agent_id"`
}

// StopSessionRequest 停止会话请求
type StopSessionRequest struct {
	MessageID string `json:"message_id" binding:"required"`
}

// MentionedItemRequest 对齐 WeKnora 的 @提及项
type MentionedItemRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`    // "kb" or "file"
	KBType string `json:"kb_type"` // "document" or "faq"
}

// CreateKnowledgeQARequest 对齐 WeKnora 的问答请求
type CreateKnowledgeQARequest struct {
	Query            string                 `json:"query" binding:"required"`
	KnowledgeBaseIDs []string               `json:"knowledge_base_ids"`
	KnowledgeIDs     []string               `json:"knowledge_ids"`
	AgentEnabled     bool                   `json:"agent_enabled"`
	AgentID          string                 `json:"agent_id"`
	WebSearchEnabled bool                   `json:"web_search_enabled"`
	SummaryModelID   string                 `json:"summary_model_id"`
	MentionedItems   []MentionedItemRequest `json:"mentioned_items"`
	DisableTitle     bool                   `json:"disable_title"`
}

// GenerateTitleRequest 对齐 WeKnora 的标题生成请求
type GenerateTitleRequest struct {
	Messages []MessageRequest `json:"messages" binding:"required"`
}

// MessageRequest 消息请求
type MessageRequest struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ExecuteRequest 问答执行请求（简化版）
type ExecuteRequest struct {
	Question string `json:"question" binding:"required"`
}

// ===== 扩展请求/响应类型（对齐 WeKnora 格式）=====

// GetSessionRequest 获取会话请求
type GetSessionRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetSessionResponse 获取会话响应（对齐 WeKnora 格式）
type GetSessionResponse struct {
	Success bool             `json:"success"`
	Data    *SessionResponse `json:"data"`
}

// ListSessionsRequest 会话列表请求
type ListSessionsRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ListSessionsResponse 会话列表响应（对齐 WeKnora 格式）
type ListSessionsResponse struct {
	Success  bool               `json:"success"`
	Data     []*SessionResponse `json:"data"`
	Total    int64              `json:"total,omitempty"`
	Page     int64              `json:"page,omitempty"`
	PageSize int64              `json:"page_size,omitempty"`
}

// CreateSessionResponse 创建会话响应（对齐 WeKnora 格式）
type CreateSessionResponse struct {
	Success bool             `json:"success"`
	Data    *SessionResponse `json:"data"`
}

// UpdateSessionResponse 更新会话响应（对齐 WeKnora 格式）
type UpdateSessionResponse struct {
	Success bool             `json:"success"`
	Data    *SessionResponse `json:"data"`
}

// DeleteSessionRequest 删除会话请求
type DeleteSessionRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteSessionResponse 删除会话响应（对齐 WeKnora 格式）
type DeleteSessionResponse struct {
	Success bool `json:"success"`
}
