// Package http 提供 SSE 事件发送逻辑.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SSESender SSE 事件发送器.
type SSESender struct{}

// NewSSESender 创建 SSE 发送器.
func NewSSESender() *SSESender {
	return &SSESender{}
}

// SendSSEEvent 发送 SSE 事件.
func (s *SSESender) Send(c *gin.Context, event SSEEvent) error {
	c.SSEvent("message", event)
	return nil
}

// SendError 发送错误事件.
func (s *SSESender) SendError(c *gin.Context, errMsg string) {
	s.Send(c, SSEEvent{
		ResponseType: SSEEventError,
		Error:       errMsg,
		SessionID:   "", // 从上下文获取
		ID:          "",
	})
}

// SendStart 发送开始事件（对齐 WeKnora agent_query 事件格式）.
func (s *SSESender) SendStart(c *gin.Context, sessionID, messageID string, flusher http.Flusher) {
	// assistant_message_id 在顶层，session_id 在 data 中
	s.Send(c, SSEEvent{
		ResponseType:        SSEEventQuery,
		Content:             "",
		ID:                  messageID,
		AssistantMessageID:  messageID, // 顶层字段，前端直接读取
		Data: map[string]interface{}{
			"session_id": sessionID, // data 中也包含 session_id
		},
	})
	flusher.Flush()
}

// SendComplete 发送完成事件.
func (s *SSESender) SendComplete(c *gin.Context, sessionID, messageID string) {
	s.Send(c, SSEEvent{
		ResponseType: SSEEventComplete,
		SessionID:   sessionID,
		ID:          messageID,
	})
}

// Flush 安全地刷新 SSE 流.
func (s *SSESender) Flush(flusher http.Flusher) {
	defer func() {
		_ = recover()
	}()
	if flusher != nil {
		flusher.Flush()
	}
}
