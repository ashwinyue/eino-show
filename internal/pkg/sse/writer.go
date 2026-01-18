package sse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Writer SSE 写入器接口.
type Writer interface {
	// Send 发送 SSE 事件
	Send(event Event) error
	// Flush 刷新缓冲区
	Flush()
	// SetHeaders 设置 SSE 响应头
	SetHeaders()
}

// GinWriter 基于 Gin 的 SSE 写入器.
type GinWriter struct {
	c       *gin.Context
	flusher http.Flusher
}

// NewGinWriter 创建 Gin SSE 写入器.
func NewGinWriter(c *gin.Context) *GinWriter {
	flusher, _ := c.Writer.(http.Flusher)
	return &GinWriter{
		c:       c,
		flusher: flusher,
	}
}

// SetHeaders 设置 SSE 响应头.
func (w *GinWriter) SetHeaders() {
	w.c.Header("Content-Type", "text/event-stream")
	w.c.Header("Cache-Control", "no-cache")
	w.c.Header("Connection", "keep-alive")
	w.c.Header("X-Accel-Buffering", "no")
}

// Send 发送 SSE 事件.
func (w *GinWriter) Send(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	// 写入 SSE 格式
	_, err = fmt.Fprintf(w.c.Writer, "event:message\ndata:%s\n\n", data)
	if err != nil {
		return fmt.Errorf("write event failed: %w", err)
	}

	w.Flush()
	return nil
}

// Flush 刷新缓冲区.
func (w *GinWriter) Flush() {
	defer func() {
		_ = recover()
	}()
	if w.flusher != nil {
		w.flusher.Flush()
	}
}

// SendError 发送错误事件.
func (w *GinWriter) SendError(message string) error {
	return w.Send(Event{
		Type:    EventTypeError,
		Content: message,
	})
}

// SendComplete 发送完成事件.
func (w *GinWriter) SendComplete(sessionID, messageID string) error {
	return w.Send(Event{
		Type:      EventTypeComplete,
		SessionID: sessionID,
		ID:        messageID,
	})
}
