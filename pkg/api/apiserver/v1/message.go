// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== Message 请求/响应类型 =====

// LoadMessagesRequest 加载消息请求
type LoadMessagesRequest struct {
	SessionID  string     `uri:"session_id" binding:"required"`
	Limit      int        `form:"limit"`
	BeforeTime *time.Time `form:"before_time"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID            string                   `json:"id"`
	SessionID     string                   `json:"session_id"`
	Role          string                   `json:"role"`
	Content       string                   `json:"content"`
	KnowledgeRefs []map[string]interface{} `json:"knowledge_refs,omitempty"`
	AgentSteps    []map[string]interface{} `json:"agent_steps,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
}

// LoadMessagesResponse 加载消息响应
type LoadMessagesResponse struct {
	Messages []*MessageResponse `json:"messages"`
	HasMore  bool               `json:"has_more"`
}

// DeleteMessageRequest 删除消息请求
type DeleteMessageRequest struct {
	SessionID string `uri:"session_id" binding:"required"`
	ID        string `uri:"id" binding:"required"`
}
