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

// SessionResponse 会话响应
type SessionResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TenantID    uint64    `json:"tenant_id"`
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

// SearchKnowledgeRequest 知识搜索请求
type SearchKnowledgeRequest struct {
	Query            string   `json:"query" binding:"required"`
	KnowledgeBaseID  string   `json:"knowledge_base_id"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids"`
	KnowledgeIDs     []string `json:"knowledge_ids"`
}

// ExecuteRequest 问答执行请求（简化版）
type ExecuteRequest struct {
	Question string `json:"question" binding:"required"`
}

// ===== 扩展请求/响应类型 =====

// GetSessionRequest 获取会话请求
type GetSessionRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetSessionResponse 获取会话响应
type GetSessionResponse struct {
	Session *SessionResponse `json:"session"`
}

// ListSessionsRequest 会话列表请求
type ListSessionsRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ListSessionsResponse 会话列表响应
type ListSessionsResponse struct {
	Sessions []*SessionResponse `json:"sessions"`
	Total    int64              `json:"total"`
}

// CreateSessionResponse 创建会话响应
type CreateSessionResponse struct {
	Session *SessionResponse `json:"session"`
}

// UpdateSessionResponse 更新会话响应
type UpdateSessionResponse struct {
	Session *SessionResponse `json:"session"`
}

// DeleteSessionRequest 删除会话请求
type DeleteSessionRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteSessionResponse 删除会话响应
type DeleteSessionResponse struct {
	Success bool `json:"success"`
}
