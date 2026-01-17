// Package http 提供基于 ADK 的 SSE 处理器.
package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/eino-show/internal/pkg/agent/enhanced"
)

// SSEEventType SSE 事件类型.
type SSEEventType string

const (
	SSEEventMessage   SSEEventType = "message"
	SSEEventThinking  SSEEventType = "thinking"
	SSEEventToolCall  SSEEventType = "tool_call"
	SSEEventToolReply SSEEventType = "tool_reply"
	SSEEventError     SSEEventType = "error"
	SSEEventComplete  SSEEventType = "complete"
)

// SSEEvent SSE 事件结构.
type SSEEvent struct {
	ID        string       `json:"id"`
	Type      SSEEventType `json:"response_type"`
	Content   string       `json:"content,omitempty"`
	Done      bool         `json:"done,omitempty"`
	SessionID string       `json:"session_id,omitempty"`
	MessageID string       `json:"assistant_message_id,omitempty"`
	Data      any          `json:"data,omitempty"`
}

// ADKSSEHandler 基于 ADK 的 SSE 处理器.
type ADKSSEHandler struct {
	agent     *enhanced.ADKAgent
	sessionID string
	messageID string
}

// NewADKSSEHandler 创建 ADK SSE 处理器.
func NewADKSSEHandler(agent *enhanced.ADKAgent, sessionID, messageID string) *ADKSSEHandler {
	return &ADKSSEHandler{
		agent:     agent,
		sessionID: sessionID,
		messageID: messageID,
	}
}

// HandleStream 处理 ADK 流式响应并发送 SSE 事件.
func (h *ADKSSEHandler) HandleStream(c *gin.Context, messages []adk.Message) {
	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.sendError(c, "streaming not supported")
		return
	}

	ctx := c.Request.Context()

	// 创建 ADK 输入
	input := &adk.AgentInput{
		Messages: messages,
	}

	// 执行 Agent 并获取迭代器
	iterator := h.agent.Run(ctx, input)

	// 处理流式事件
	h.processEvents(ctx, c, flusher, iterator)

	// 发送完成事件
	h.sendComplete(c, flusher)
}

// processEvents 处理 ADK 事件流.
func (h *ADKSSEHandler) processEvents(ctx context.Context, c *gin.Context, flusher http.Flusher, iterator *adk.AsyncIterator[*adk.AgentEvent]) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			event, ok := iterator.Next()
			if !ok {
				return
			}

			// 处理错误
			if event.Err != nil {
				h.sendError(c, event.Err.Error())
				flusher.Flush()
				return
			}

			// 根据事件类型处理
			sseEvent := h.convertEvent(event)
			if sseEvent != nil {
				h.sendSSEEvent(c, sseEvent)
				flusher.Flush()
			}
		}
	}
}

// convertEvent 转换 ADK 事件为 SSE 事件.
func (h *ADKSSEHandler) convertEvent(event *adk.AgentEvent) *SSEEvent {
	// 处理 Output
	if event.Output != nil && event.Output.MessageOutput != nil {
		msgVariant := event.Output.MessageOutput

		// 处理流式消息
		if msgVariant.IsStreaming && msgVariant.MessageStream != nil {
			msg, err := msgVariant.MessageStream.Recv()
			if err != nil {
				return nil
			}
			return h.convertMessage(msg)
		}

		// 处理非流式消息
		if msgVariant.Message != nil {
			return h.convertMessage(msgVariant.Message)
		}
	}

	return nil
}

// convertMessage 转换消息为 SSE 事件.
func (h *ADKSSEHandler) convertMessage(msg *schema.Message) *SSEEvent {
	sseEvent := &SSEEvent{
		ID:        h.messageID,
		SessionID: h.sessionID,
		MessageID: h.messageID,
	}

	// 检查是否是工具调用
	if len(msg.ToolCalls) > 0 {
		sseEvent.Type = SSEEventToolCall
		sseEvent.Data = map[string]any{
			"tool_calls": msg.ToolCalls,
		}
		return sseEvent
	}

	// 检查消息角色
	switch msg.Role {
	case schema.Tool:
		sseEvent.Type = SSEEventToolReply
		sseEvent.Content = msg.Content
		sseEvent.Data = map[string]any{
			"tool_call_id": msg.ToolCallID,
			"tool_name":    msg.Name,
		}
	case schema.Assistant:
		sseEvent.Type = SSEEventMessage
		sseEvent.Content = msg.Content
	default:
		sseEvent.Type = SSEEventMessage
		sseEvent.Content = msg.Content
	}

	return sseEvent
}

// sendSSEEvent 发送 SSE 事件.
func (h *ADKSSEHandler) sendSSEEvent(c *gin.Context, event *SSEEvent) {
	data, _ := json.Marshal(event)
	c.Writer.Write([]byte("event: message\n"))
	c.Writer.Write([]byte("data: "))
	c.Writer.Write(data)
	c.Writer.Write([]byte("\n\n"))
}

// sendError 发送错误事件.
func (h *ADKSSEHandler) sendError(c *gin.Context, errMsg string) {
	event := &SSEEvent{
		ID:        h.messageID,
		Type:      SSEEventError,
		Content:   errMsg,
		Done:      true,
		SessionID: h.sessionID,
		MessageID: h.messageID,
	}
	h.sendSSEEvent(c, event)
}

// sendComplete 发送完成事件.
func (h *ADKSSEHandler) sendComplete(c *gin.Context, flusher http.Flusher) {
	event := &SSEEvent{
		ID:        h.messageID,
		Type:      SSEEventComplete,
		Done:      true,
		SessionID: h.sessionID,
		MessageID: h.messageID,
	}
	h.sendSSEEvent(c, event)
	flusher.Flush()
}
