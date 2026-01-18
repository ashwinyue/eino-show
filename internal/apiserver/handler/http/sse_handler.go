// Package http 提供基于 ADK 的 SSE 处理器.
// 对齐 Eino 官方实现: a-old/old/eino-examples/adk/intro/http-sse-service/main.go
package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/eino-show/internal/pkg/sse"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// SSEEventType 类型别名，指向 sse 包.
type SSEEventType = sse.EventType

// 事件类型常量别名.
const (
	SSEEventQuery      = sse.EventTypeQuery
	SSEEventAnswer     = sse.EventTypeAnswer
	SSEEventThinking   = sse.EventTypeThinking
	SSEEventToolCall   = sse.EventTypeToolCall
	SSEEventToolResult = sse.EventTypeToolResult
	SSEEventReferences = sse.EventTypeReferences
	SSEEventReflection = sse.EventTypeReflection
	SSEEventAction     = sse.EventTypeAction
	SSEEventComplete   = sse.EventTypeComplete
	SSEEventError      = sse.EventTypeError
)

// ThinkingEvent 类型别名.
type ThinkingEvent = sse.ThinkingEvent

// ReferencesEvent 类型别名.
type ReferencesEvent = sse.ReferencesEvent

// ReferenceChunk 类型别名.
type ReferenceChunk = sse.ReferenceChunk

// ReflectionEvent 类型别名.
type ReflectionEvent = sse.ReflectionEvent

// SSEEvent SSE 事件结构（保持兼容，使用 schema.ToolCall）.
type SSEEvent struct {
	ResponseType       SSEEventType           `json:"response_type"`
	ID                 string                 `json:"id"`
	Content            string                 `json:"content,omitempty"`
	Done               bool                   `json:"done,omitempty"`
	AgentName          string                 `json:"agent_name,omitempty"`
	RunPath            string                 `json:"run_path,omitempty"`
	ToolCalls          []schema.ToolCall      `json:"tool_calls,omitempty"`
	ActionType         string                 `json:"action_type,omitempty"`
	Error              string                 `json:"error,omitempty"`
	SessionID          string                 `json:"session_id,omitempty"`
	Data               map[string]interface{} `json:"data,omitempty"`
	AssistantMessageID string                 `json:"assistant_message_id,omitempty"`
}

// ADKSSEHandler 基于 ADK 的 SSE 处理器.
type ADKSSEHandler struct {
	runner          *adk.Runner
	sessionID       string
	messageID       string
	assistantMsgID  string // 助手消息 ID（用于保存到数据库）
	messages        []adk.Message
	toolCallsBuffer map[int][]*schema.Message
	contentBuffer   strings.Builder                                                                                         // 累积回答内容
	stepCollector   *StepCollector                                                                                          // Agent steps 收集器
	saveCtx         context.Context                                                                                         // 用于保存消息的上下文
	updateMessageFn func(ctx context.Context, messageID, content string, agentSteps []v1.AgentStep, isCompleted bool) error // 保存消息的回调函数
	titleChan       chan string                                                                                             // 标题生成 channel（对齐 WeKnora）
}

// NewADKSSEHandler 创建 ADK SSE 处理器.
func NewADKSSEHandler(runner *adk.Runner, sessionID, messageID string, messages []adk.Message) *ADKSSEHandler {
	return &ADKSSEHandler{
		runner:          runner,
		sessionID:       sessionID,
		messageID:       messageID,
		messages:        messages,
		toolCallsBuffer: make(map[int][]*schema.Message),
	}
}

// NewADKSSEHandlerWithSave 创建带保存功能的 ADK SSE 处理器.
func NewADKSSEHandlerWithSave(runner *adk.Runner, sessionID, assistantMsgID string, messages []adk.Message, updateFn func(ctx context.Context, messageID, content string, agentSteps []v1.AgentStep, isCompleted bool) error, ctx context.Context) *ADKSSEHandler {
	return &ADKSSEHandler{
		runner:          runner,
		sessionID:       sessionID,
		messageID:       assistantMsgID, // 使用 assistantMsgID 作为 messageID
		assistantMsgID:  assistantMsgID,
		messages:        messages,
		toolCallsBuffer: make(map[int][]*schema.Message),
		stepCollector:   NewStepCollector(),
		saveCtx:         ctx,
		updateMessageFn: updateFn,
	}
}

// SetTitleChannel 设置标题生成 channel（对齐 WeKnora）.
func (h *ADKSSEHandler) SetTitleChannel(ch chan string) {
	h.titleChan = ch
}

// HandleStream 处理 ADK 流式响应并发送 SSE 事件.
func (h *ADKSSEHandler) HandleStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.sendSSEError(c, "streaming not supported")
		return
	}

	ctx := c.Request.Context()

	// 发送开始事件
	h.sendStart(c, flusher)

	// 使用 Runner.Run 获取事件迭代器
	iter := h.runner.Run(ctx, h.messages)

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if err := h.processAgentEvent(c, event); err != nil {
			h.sendSSEError(c, err.Error())
			return
		}

		safeFlush(flusher)
	}

	h.flushToolCalls(c)
	h.sendComplete(c)
	safeFlush(flusher)
}

// processAgentEvent 处理单个 Agent 事件.
func (h *ADKSSEHandler) processAgentEvent(c *gin.Context, event *adk.AgentEvent) error {
	// 处理错误事件
	if event.Err != nil {
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventError,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			Error:        event.Err.Error(),
			SessionID:    h.sessionID,
			ID:           h.messageID,
		})
	}

	// 处理消息输出
	if event.Output != nil && event.Output.MessageOutput != nil {
		return h.handleMessageOutput(c, event)
	}

	// 处理自定义输出（Thinking、References、Reflection 等）
	if event.Output != nil && event.Output.CustomizedOutput != nil {
		return h.handleCustomizedOutput(c, event)
	}

	// 处理 Action 事件
	if event.Action != nil {
		return h.handleAction(c, event)
	}

	return nil
}

// handleMessageOutput 处理消息输出.
func (h *ADKSSEHandler) handleMessageOutput(c *gin.Context, event *adk.AgentEvent) error {
	msgOutput := event.Output.MessageOutput
	if msgOutput == nil {
		return nil
	}

	if msg := msgOutput.Message; msg != nil {
		return h.handleMessage(c, event, msg)
	}

	if stream := msgOutput.MessageStream; stream != nil {
		return h.handleStreamingMessage(c, event, stream)
	}

	return nil
}

// handleMessage 处理单个消息.
func (h *ADKSSEHandler) handleMessage(c *gin.Context, event *adk.AgentEvent, msg *schema.Message) error {
	// 检测 think 工具调用（在 tool_call 阶段）
	if h.isThinkToolCall(msg) {
		return h.handleThinkToolCall(c, event, msg)
	}

	// 检测 todo_write 工具调用（在 tool_call 阶段），像 think 一样改写内容
	if h.isTodoWriteToolCall(msg) {
		return h.handleTodoWriteToolCall(c, event, msg)
	}

	eventType := SSEEventAnswer

	// 累积 assistant 回答内容（用于保存到数据库）
	if msg.Role == schema.Assistant && msg.Content != "" && h.assistantMsgID != "" {
		h.contentBuffer.WriteString(msg.Content)
	}

	if msg.Role == schema.Tool {
		// 对于 todo_write 工具的结果，解析输出并提取步骤信息
		if msg.ToolName == "todo_write" {
			return h.handleTodoWriteResult(c, event, msg)
		}
		// 对于 thinking 工具的结果，转换为 thinking 事件显示（对齐 WeKnora）
		if msg.ToolName == "thinking" || msg.ToolName == "think" {
			// 解析 thinking 工具的 JSON 结果
			var thinkingResult struct {
				Thought           string `json:"thought"`
				ThoughtNumber     int    `json:"thought_number"`
				TotalThoughts     int    `json:"total_thoughts"`
				NextThoughtNeeded bool   `json:"next_thought_needed"`
				IncompleteSteps   bool   `json:"incomplete_steps"`
			}
			if err := json.Unmarshal([]byte(msg.Content), &thinkingResult); err == nil {
				// 计算对齐 WeKnora 的字段
				done := !thinkingResult.NextThoughtNeeded && thinkingResult.ThoughtNumber >= thinkingResult.TotalThoughts

				// 构建 data 字段（对齐 WeKnora StreamEvent.Data）
				data := map[string]interface{}{
					"iteration":           thinkingResult.ThoughtNumber, // 对齐 WeKnora Iteration
					"done":                done,                         // 对齐 WeKnora Done
					"thought_number":      thinkingResult.ThoughtNumber,
					"total_thoughts":      thinkingResult.TotalThoughts,
					"next_thought_needed": thinkingResult.NextThoughtNeeded,
				}

				// 如果完成，添加 duration_ms（对齐 WeKnora）
				if done {
					data["duration_ms"] = 0 // 可选：计算实际耗时
				}

				// 收集思考内容用于持久化
				if h.stepCollector != nil {
					h.stepCollector.CollectThought(thinkingResult.Thought)
				}

				// 发送 thinking 事件（对齐 WeKnora ResponseTypeThinking）
				return h.sendSSEEvent(c, SSEEvent{
					ResponseType: SSEEventThinking,
					AgentName:    event.AgentName,
					RunPath:      formatRunPath(event.RunPath),
					Content:      thinkingResult.Thought, // 对齐 WeKnora StreamEvent.Content
					Done:         done,                   // 对齐 WeKnora StreamEvent.Done
					SessionID:    h.sessionID,
					ID:           h.messageID,
					Data:         data, // 对齐 WeKnora StreamEvent.Data
				})
			}
		}
		eventType = SSEEventToolResult
	}

	sseEvent := SSEEvent{
		ResponseType: eventType,
		AgentName:    event.AgentName,
		RunPath:      formatRunPath(event.RunPath),
		Content:      msg.Content,
		SessionID:    h.sessionID,
		ID:           h.messageID,
	}

	if len(msg.ToolCalls) > 0 {
		sseEvent.ResponseType = SSEEventToolCall
		sseEvent.ToolCalls = msg.ToolCalls
	}

	// 对于 tool_result，添加工具调用信息到 data 字段（对齐前端期待）
	if msg.Role == schema.Tool {
		sseEvent.Data = map[string]interface{}{
			"tool_call_id": msg.ToolCallID,
			"tool_name":    msg.ToolName,
			"success":      true,
			"output":       msg.Content,
		}
	}

	return h.sendSSEEvent(c, sseEvent)
}

// handleStreamingMessage 处理流式消息.
func (h *ADKSSEHandler) handleStreamingMessage(c *gin.Context, event *adk.AgentEvent, stream *schema.StreamReader[*schema.Message]) error {
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return h.sendSSEEvent(c, SSEEvent{
				ResponseType: SSEEventError,
				AgentName:    event.AgentName,
				RunPath:      formatRunPath(event.RunPath),
				Error:        fmt.Sprintf("stream error: %v", err),
			})
		}

		// 先聚合 ToolCalls（不立即发送，等完整后再发送）
		if len(chunk.ToolCalls) > 0 {
			for _, tc := range chunk.ToolCalls {
				if tc.Index != nil {
					h.toolCallsBuffer[*tc.Index] = append(h.toolCallsBuffer[*tc.Index], &schema.Message{
						Role: chunk.Role,
						ToolCalls: []schema.ToolCall{
							{
								ID:    tc.ID,
								Type:  tc.Type,
								Index: tc.Index,
								Function: schema.FunctionCall{
									Name:      tc.Function.Name,
									Arguments: tc.Function.Arguments,
								},
							},
						},
					})
				}
			}
		}

		if chunk.Content != "" {
			// 累积回答内容（用于保存到数据库）
			roleStr := string(chunk.Role)

			// 使用字符串比较，兼容 Eino 的 RoleType
			if h.assistantMsgID != "" && (roleStr == "assistant" || roleStr == "Assistant" || string(chunk.Role) == "assistant") {
				h.contentBuffer.WriteString(chunk.Content)
			}

			eventType := SSEEventAnswer

			// 对于 todo_write 工具的结果，解析输出为 PlanData 格式（对齐 WeKnora）
			if chunk.Role == schema.Tool && (chunk.ToolName == "todo_write") {
				// 解析 todo_write 工具的 JSON 输出
				var todoResult struct {
					PlanID string `json:"plan_id"`
					Task   string `json:"task"`
					Steps  []struct {
						ID          string `json:"id"`
						Description string `json:"description"`
						Status      string `json:"status"`
					} `json:"steps"`
				}
				if err := json.Unmarshal([]byte(chunk.Content), &todoResult); err == nil {
					// 先发送 tool_call 事件，确保前端能正确关联 tool_result
					h.sendToolCallEvent(c, event, chunk.ToolCallID, chunk.ToolName, chunk.Content)

					// 生成 Markdown 文本作为 Content（对齐 WeKnora）
					content := h.formatPlanAsMarkdown(todoResult.Task, todoResult.Steps)

					// 构建 data 字段，直接包含所有计划相关字段（对齐 WeKnora）
					dataFields := map[string]interface{}{
						"tool_call_id": chunk.ToolCallID,
						"tool_name":    chunk.ToolName,
						"success":      true,
						"output":       content, // 对齐 WeKnora：output 字段
						"display_type": "plan",  // 直接在 data 中（对齐 WeKnora）
						"task":         todoResult.Task,
						"steps":        todoResult.Steps,
						"total_steps":  len(todoResult.Steps),
						"plan_created": true, // 对齐 WeKnora
					}
					if todoResult.PlanID != "" {
						dataFields["plan_id"] = todoResult.PlanID
					}

					// 发送 tool_result 事件（对齐 WeKnora）
					if err := h.sendSSEEvent(c, SSEEvent{
						ResponseType: SSEEventToolResult,
						AgentName:    event.AgentName,
						RunPath:      formatRunPath(event.RunPath),
						Content:      content, // 对齐 WeKnora：包含 Markdown 文本
						SessionID:    h.sessionID,
						ID:           h.messageID,
						Data:         dataFields,
					}); err != nil {
						return err
					}
					// 已经发送了带 tool_data 的事件，跳过后续处理
					continue
				}
			}

			// 对于 web_search 工具的结果，解析输出为 WebSearchResultsData 格式（对齐 WeKnora）
			if chunk.Role == schema.Tool && chunk.ToolName == "web_search" {
				// 先发送 tool_call 事件
				h.sendToolCallEvent(c, event, chunk.ToolCallID, chunk.ToolName, chunk.Content)

				// 尝试解析搜索结果
				results, query := h.parseWebSearchResults(chunk.Content)

				// 构建 data 字段
				dataFields := map[string]interface{}{
					"tool_call_id": chunk.ToolCallID,
					"tool_name":    chunk.ToolName,
					"success":      true,
					"output":       chunk.Content,
					"display_type": "web_search_results",
					"query":        query,
					"results":      results,
					"count":        len(results),
				}

				// 发送 tool_result 事件
				if err := h.sendSSEEvent(c, SSEEvent{
					ResponseType: SSEEventToolResult,
					AgentName:    event.AgentName,
					RunPath:      formatRunPath(event.RunPath),
					Content:      chunk.Content,
					SessionID:    h.sessionID,
					ID:           h.messageID,
					Data:         dataFields,
				}); err != nil {
					return err
				}
				continue
			}

			// 对于 thinking 工具的结果，转换为 thinking 事件显示（对齐 WeKnora）
			if chunk.Role == schema.Tool && (chunk.ToolName == "thinking" || chunk.ToolName == "think") {
				// 先 flush ToolCalls，确保 tool_call 在 thinking 之前
				h.flushToolCalls(c)

				// 解析 thinking 工具的 JSON 结果
				var thinkingResult struct {
					Thought           string `json:"thought"`
					ThoughtNumber     int    `json:"thought_number"`
					TotalThoughts     int    `json:"total_thoughts"`
					NextThoughtNeeded bool   `json:"next_thought_needed"`
					IncompleteSteps   bool   `json:"incomplete_steps"`
				}
				if err := json.Unmarshal([]byte(chunk.Content), &thinkingResult); err == nil {
					// 计算对齐 WeKnora 的字段
					done := !thinkingResult.NextThoughtNeeded && thinkingResult.ThoughtNumber >= thinkingResult.TotalThoughts

					// 构建 data 字段（对齐 WeKnora StreamEvent.Data）
					data := map[string]interface{}{
						"iteration":           thinkingResult.ThoughtNumber, // 对齐 WeKnora Iteration
						"done":                done,                         // 对齐 WeKnora Done
						"thought_number":      thinkingResult.ThoughtNumber,
						"total_thoughts":      thinkingResult.TotalThoughts,
						"next_thought_needed": thinkingResult.NextThoughtNeeded,
					}

					// 如果完成，添加 duration_ms（对齐 WeKnora）
					if done {
						data["duration_ms"] = 0 // 可选：计算实际耗时
					}

					// 发送 thinking 事件（对齐 WeKnora ResponseTypeThinking）
					if err := h.sendSSEEvent(c, SSEEvent{
						ResponseType: SSEEventThinking,
						AgentName:    event.AgentName,
						RunPath:      formatRunPath(event.RunPath),
						Content:      thinkingResult.Thought, // 对齐 WeKnora StreamEvent.Content
						Done:         done,                   // 对齐 WeKnora StreamEvent.Done
						SessionID:    h.sessionID,
						ID:           h.messageID,
						Data:         data, // 对齐 WeKnora StreamEvent.Data
					}); err != nil {
						return err
					}
					// 已经发送了 thinking 事件，跳过后续 tool_result 处理
					continue
				}
			}

			if chunk.Role == schema.Tool {
				// 其他 tool 结果，先 flush ToolCalls
				h.flushToolCalls(c)
				eventType = SSEEventToolResult
			}

			sseEvent := SSEEvent{
				ResponseType: eventType,
				AgentName:    event.AgentName,
				RunPath:      formatRunPath(event.RunPath),
				Content:      chunk.Content,
				SessionID:    h.sessionID,
				ID:           h.messageID,
			}

			// 对于 tool_result，添加工具调用信息到 data 字段（对齐前端期待）
			if chunk.Role == schema.Tool {
				sseEvent.Data = map[string]interface{}{
					"tool_call_id": chunk.ToolCallID,
					"tool_name":    chunk.ToolName,
					"success":      true,
					"output":       chunk.Content,
				}
			}

			if err := h.sendSSEEvent(c, sseEvent); err != nil {
				return err
			}
		}
	}

	return nil
}

// handleCustomizedOutput 处理自定义输出.
func (h *ADKSSEHandler) handleCustomizedOutput(c *gin.Context, event *adk.AgentEvent) error {
	customOutput := event.Output.CustomizedOutput

	switch v := customOutput.(type) {
	case ThinkingEvent:
		return h.sendSSEThinkingEvent(c, event, &v)

	case *ThinkingEvent:
		return h.sendSSEThinkingEvent(c, event, v)

	case ReferencesEvent:
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventReferences,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			SessionID:    h.sessionID,
			ID:           h.messageID,
			Data: map[string]interface{}{
				"chunks": v.Chunks,
			},
		})

	case *ReferencesEvent:
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventReferences,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			SessionID:    h.sessionID,
			ID:           h.messageID,
			Data: map[string]interface{}{
				"chunks": v.Chunks,
			},
		})

	case ReflectionEvent:
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventReflection,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			SessionID:    h.sessionID,
			ID:           h.messageID,
			Data: map[string]interface{}{
				"reflection": v.Reflection,
				"score":      v.Score,
			},
		})

	case *ReflectionEvent:
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventReflection,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			SessionID:    h.sessionID,
			ID:           h.messageID,
			Data: map[string]interface{}{
				"reflection": v.Reflection,
				"score":      v.Score,
			},
		})

	case map[string]interface{}:
		if typeVal, ok := v["type"].(string); ok {
			return h.sendSSEEvent(c, SSEEvent{
				ResponseType: SSEEventType(typeVal),
				AgentName:    event.AgentName,
				RunPath:      formatRunPath(event.RunPath),
				Content:      toString(v["content"]),
				SessionID:    h.sessionID,
				ID:           h.messageID,
				Data:         v,
			})
		}

	case string:
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventThinking,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			Content:      v,
			SessionID:    h.sessionID,
			ID:           h.messageID,
		})

	default:
		// Unknown CustomizedOutput type, ignore
	}

	return nil
}

// toString 安全地将任意值转换为字符串.
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", v)
}

// sendToolCallEvent 发送 tool_call 事件，确保 tool_result 之前有对应的 tool_call.
func (h *ADKSSEHandler) sendToolCallEvent(c *gin.Context, event *adk.AgentEvent, toolCallID, toolName, arguments string) {
	// 收集 tool_call 用于持久化
	if h.stepCollector != nil {
		h.stepCollector.CollectToolCall(toolCallID, toolName, arguments)
	}

	_ = h.sendSSEEvent(c, SSEEvent{
		ResponseType: SSEEventToolCall,
		AgentName:    event.AgentName,
		RunPath:      formatRunPath(event.RunPath),
		SessionID:    h.sessionID,
		ID:           h.messageID,
		ToolCalls: []schema.ToolCall{{
			ID:   toolCallID,
			Type: "function",
			Function: schema.FunctionCall{
				Name:      toolName,
				Arguments: arguments,
			},
		}},
		Data: map[string]interface{}{
			"tool_call_id": toolCallID,
			"tool_name":    toolName,
			"arguments":    arguments,
		},
	})
}

// flushToolCalls 刷新聚合的 ToolCalls（只发送完整的 ToolCall）.
func (h *ADKSSEHandler) flushToolCalls(c *gin.Context) {
	if len(h.toolCallsBuffer) == 0 {
		return
	}

	for index, msgs := range h.toolCallsBuffer {
		concatenatedMsg, err := schema.ConcatMessages(msgs)
		if err != nil {
			continue
		}

		if len(concatenatedMsg.ToolCalls) > 0 {
			// 检查 ToolCall 是否完整（arguments 应该是完整的 JSON）
			for _, tc := range concatenatedMsg.ToolCalls {
				if h.isToolCallComplete(tc) {
					// 收集 tool_call 用于持久化
					if h.stepCollector != nil {
						h.stepCollector.CollectToolCall(tc.ID, tc.Function.Name, tc.Function.Arguments)
					}
					_ = h.sendSSEEvent(c, SSEEvent{
						ResponseType: SSEEventToolCall,
						ToolCalls:    []schema.ToolCall{tc},
						SessionID:    h.sessionID,
						ID:           h.messageID,
					})
				} else {
					// ToolCall 不完整，保留在 buffer 中等待后续数据
				}
			}
		}

		// 清空已发送的 ToolCalls buffer
		delete(h.toolCallsBuffer, index)
	}
}

// forceFlushToolCalls 强制刷新所有 ToolCalls（包括不完整的）.
// 用于处理 todo_write 等工具，确保 tool_call 事件在 tool_result 之前被发送.
func (h *ADKSSEHandler) forceFlushToolCalls(c *gin.Context) {
	if len(h.toolCallsBuffer) == 0 {
		return
	}

	for index, msgs := range h.toolCallsBuffer {
		concatenatedMsg, err := schema.ConcatMessages(msgs)
		if err != nil {
			// 即使合并失败，也尝试发送原始消息
			for _, msg := range msgs {
				if len(msg.ToolCalls) > 0 {
					for _, tc := range msg.ToolCalls {
						// 收集 tool_call 用于持久化
						if h.stepCollector != nil {
							h.stepCollector.CollectToolCall(tc.ID, tc.Function.Name, tc.Function.Arguments)
						}
						_ = h.sendSSEEvent(c, SSEEvent{
							ResponseType: SSEEventToolCall,
							ToolCalls:    []schema.ToolCall{tc},
							SessionID:    h.sessionID,
							ID:           h.messageID,
						})
					}
				}
			}
		} else if len(concatenatedMsg.ToolCalls) > 0 {
			for _, tc := range concatenatedMsg.ToolCalls {
				// 收集 tool_call 用于持久化
				if h.stepCollector != nil {
					h.stepCollector.CollectToolCall(tc.ID, tc.Function.Name, tc.Function.Arguments)
				}
				// 强制发送，不管是否完整
				_ = h.sendSSEEvent(c, SSEEvent{
					ResponseType: SSEEventToolCall,
					ToolCalls:    []schema.ToolCall{tc},
					SessionID:    h.sessionID,
					ID:           h.messageID,
				})
			}
		}

		// 清空 buffer
		delete(h.toolCallsBuffer, index)
	}
}

// isToolCallComplete 检查 ToolCall 是否完整.
func (h *ADKSSEHandler) isToolCallComplete(tc schema.ToolCall) bool {
	// 检查 ID 是否存在
	if tc.ID == "" {
		return false
	}

	// 检查 Function.Name 是否存在
	if tc.Function.Name == "" {
		return false
	}

	// 检查 Arguments 是否是完整的 JSON
	args := tc.Function.Arguments
	if args == "" {
		return false
	}

	// 尝试解析 JSON，检查是否完整
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(args), &js); err != nil {
		return false
	}

	return true
}

// handleAction 处理 Agent Action 事件.
func (h *ADKSSEHandler) handleAction(c *gin.Context, event *adk.AgentEvent) error {
	action := event.Action

	if action.TransferToAgent != nil {
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventAction,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			ActionType:   "transfer",
			Content:      fmt.Sprintf("Transfer to agent: %s", action.TransferToAgent.DestAgentName),
			SessionID:    h.sessionID,
			ID:           h.messageID,
		})
	}

	if action.Interrupted != nil {
		for _, ic := range action.Interrupted.InterruptContexts {
			content := fmt.Sprintf("%v", ic.Info)
			if stringer, ok := ic.Info.(fmt.Stringer); ok {
				content = stringer.String()
			}

			_ = h.sendSSEEvent(c, SSEEvent{
				ResponseType: SSEEventAction,
				AgentName:    event.AgentName,
				RunPath:      formatRunPath(event.RunPath),
				ActionType:   "interrupted",
				Content:      content,
				SessionID:    h.sessionID,
				ID:           h.messageID,
			})
		}
	}

	if action.Exit {
		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventAction,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			ActionType:   "exit",
			Content:      "Agent execution completed",
			SessionID:    h.sessionID,
			ID:           h.messageID,
		})
	}

	return nil
}

// sendSSEEvent 发送 SSE 事件.
func (h *ADKSSEHandler) sendSSEEvent(c *gin.Context, event SSEEvent) error {
	c.SSEvent("message", event)
	return nil
}

// sendSSEError 发送错误事件.
func (h *ADKSSEHandler) sendSSEError(c *gin.Context, errMsg string) {
	h.sendSSEEvent(c, SSEEvent{
		ResponseType: SSEEventError,
		Error:        errMsg,
		SessionID:    h.sessionID,
		ID:           h.messageID,
	})
}

// sendStart 发送开始事件（仅用于标识请求开始，不显示内容）.
func (h *ADKSSEHandler) sendStart(c *gin.Context, flusher http.Flusher) {
	// 对齐 WeKnora agent_query 事件格式
	// assistant_message_id 在顶层，session_id 在 data 中
	h.sendSSEEvent(c, SSEEvent{
		ResponseType:       SSEEventQuery,
		Content:            "",
		ID:                 h.messageID,
		AssistantMessageID: h.messageID, // 顶层字段，前端直接读取
		Data: map[string]interface{}{
			"session_id": h.sessionID, // data 中也包含 session_id
		},
	})
	safeFlush(flusher)
}

// sendComplete 发送完成事件并保存消息到数据库（对齐 WeKnora completeAssistantMessage）.
func (h *ADKSSEHandler) sendComplete(c *gin.Context) {
	// 在 stop 事件前发送 session_title 事件（对齐 WeKnora）
	if h.titleChan != nil {
		select {
		case title := <-h.titleChan:
			if title != "" {
				h.sendSSEEvent(c, SSEEvent{
					ResponseType: "session_title",
					SessionID:    h.sessionID,
					ID:           h.messageID,
					Content:      title,
					Data: map[string]interface{}{
						"session_id": h.sessionID,
						"title":      title,
					},
				})
			}
		default:
			// 标题还没生成完，不等待
		}
	}

	h.sendSSEEvent(c, SSEEvent{
		ResponseType: SSEEventComplete,
		SessionID:    h.sessionID,
		ID:           h.messageID,
	})

	// 如果有配置保存回调，保存累积的消息内容
	if h.updateMessageFn != nil && h.assistantMsgID != "" {
		finalContent := h.contentBuffer.String()
		// 完成当前 step（如果有）
		if h.stepCollector != nil {
			h.stepCollector.FinalizeCurrentStep()
		}
		var agentSteps []v1.AgentStep
		if h.stepCollector != nil {
			agentSteps = h.stepCollector.GetSteps()
		}
		// 异步保存，避免阻塞 SSE
		go func() {
			_ = h.updateMessageFn(h.saveCtx, h.assistantMsgID, finalContent, agentSteps, true)
		}()
	}
}

// safeFlush 安全地刷新 SSE 流.
func safeFlush(flusher http.Flusher) {
	defer func() {
		_ = recover()
	}()
	if flusher != nil {
		flusher.Flush()
	}
}

// formatRunPath 格式化 RunPath 为字符串.
func formatRunPath(runPath []adk.RunStep) string {
	if len(runPath) == 0 {
		return ""
	}
	result := ""
	for i, step := range runPath {
		if i > 0 {
			result += " -> "
		}
		result += step.String()
	}
	return result
}

// isThinkToolCall 检测消息是否包含 think 工具调用.
func (h *ADKSSEHandler) isThinkToolCall(msg *schema.Message) bool {
	for _, tc := range msg.ToolCalls {
		if tc.Function.Name == "think" {
			return true
		}
	}
	return false
}

// isTodoWriteToolCall 检测消息是否包含 todo_write 工具调用.
func (h *ADKSSEHandler) isTodoWriteToolCall(msg *schema.Message) bool {
	for _, tc := range msg.ToolCalls {
		if tc.Function.Name == "todo_write" {
			return true
		}
	}
	return false
}

// handleTodoWriteResult 处理 todo_write 工具的结果，解析步骤为 PlanData 格式（对齐 WeKnora）.
func (h *ADKSSEHandler) handleTodoWriteResult(c *gin.Context, event *adk.AgentEvent, msg *schema.Message) error {
	// 先发送 tool_call 事件，确保前端能正确关联 tool_result
	h.sendToolCallEvent(c, event, msg.ToolCallID, msg.ToolName, msg.Content)

	// 尝试解析 JSON 格式（如果工具输出被修改为 JSON）
	var jsonResult struct {
		PlanID string `json:"plan_id"`
		Task   string `json:"task"`
		Steps  []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Status      string `json:"status"`
		} `json:"steps"`
	}

	var task string
	var steps []struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	var planID string

	// 首先尝试 JSON 解析
	if err := json.Unmarshal([]byte(msg.Content), &jsonResult); err == nil {
		task = jsonResult.Task
		steps = jsonResult.Steps
		planID = jsonResult.PlanID
	} else {
		// JSON 解析失败，尝试解析 Markdown 格式
		task, steps, planID = parseTodoWriteMarkdown(msg.Content)
	}

	// 如果成功提取到任务和步骤，发送 PlanData 格式（对齐 WeKnora）
	// 注意：display_type/task/steps 直接在 data 中，不嵌套在 tool_data 中
	if task != "" && len(steps) > 0 {
		// 生成 Markdown 文本作为 Content（对齐 WeKnora）
		content := h.formatPlanAsMarkdown(task, steps)

		// 构建 data 字段，直接包含所有计划相关字段
		dataFields := map[string]interface{}{
			"tool_call_id": msg.ToolCallID,
			"tool_name":    msg.ToolName,
			"success":      true,
			"output":       content, // 对齐 WeKnora：output 字段
			"display_type": "plan",  // 直接在 data 中（对齐 WeKnora）
			"task":         task,
			"steps":        steps,
			"total_steps":  len(steps),
			"plan_created": true, // 对齐 WeKnora
		}
		if planID != "" {
			dataFields["plan_id"] = planID
		}

		return h.sendSSEEvent(c, SSEEvent{
			ResponseType: SSEEventToolResult,
			AgentName:    event.AgentName,
			RunPath:      formatRunPath(event.RunPath),
			Content:      content, // 对齐 WeKnora：包含 Markdown 文本
			SessionID:    h.sessionID,
			ID:           h.messageID,
			Data:         dataFields,
		})
	}

	// 解析失败时使用默认处理
	return h.sendSSEEvent(c, SSEEvent{
		ResponseType: SSEEventToolResult,
		AgentName:    event.AgentName,
		RunPath:      formatRunPath(event.RunPath),
		Content:      msg.Content,
		SessionID:    h.sessionID,
		ID:           h.messageID,
		Data: map[string]interface{}{
			"tool_call_id": msg.ToolCallID,
			"tool_name":    msg.ToolName,
			"success":      true,
			"output":       msg.Content,
		},
	})
}

// parseTodoWriteMarkdown 从 Markdown 格式的 todo_write 输出中提取任务和步骤.
func parseTodoWriteMarkdown(content string) (task string, steps []struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}, planID string) {
	// 提取任务: **任务**: xxx
	lines := splitLines(content)
	for _, line := range lines {
		if strings.HasPrefix(line, "**任务**:") || strings.HasPrefix(line, "**任务**：") {
			task = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "**任务**:"), "**任务**："))
			break
		}
	}

	// 提取 plan_id: **计划 ID**: xxx
	for _, line := range lines {
		if strings.HasPrefix(line, "**计划 ID**:") || strings.HasPrefix(line, "**计划 ID**：") {
			planID = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "**计划 ID**:"), "**计划 ID**："))
			break
		}
	}

	// 提取步骤: 格式如 "  1. step1 [⏳] description"
	statusMap := map[string]string{
		"⏳": "pending",
		"🔄": "in_progress",
		"✅": "completed",
	}

	for _, line := range lines {
		// 匹配步骤行: "  1. step1 [⏳] description"
		if matches := regexp.MustCompile(`^\s*\d+\.\s+(\S+)\s+\[([⏳🔄✅])\]\s+(.+)$`).FindStringSubmatch(line); len(matches) == 4 {
			step := struct {
				ID          string `json:"id"`
				Description string `json:"description"`
				Status      string `json:"status"`
			}{
				ID:          matches[1],
				Description: strings.TrimSpace(matches[3]),
				Status:      statusMap[matches[2]],
			}
			if step.Status == "" {
				step.Status = "pending"
			}
			steps = append(steps, step)
		}
	}

	return task, steps, planID
}

// splitLines 分割字符串为行.
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// handleThinkToolCall 处理 think 工具调用，发送 thinking 事件（对齐 WeKnora）.
// WeKnora 事件流程: Agent → EventBus(EventAgentThought) → StreamManager → SSE(thinking)
func (h *ADKSSEHandler) handleThinkToolCall(c *gin.Context, event *adk.AgentEvent, msg *schema.Message) error {
	for _, tc := range msg.ToolCalls {
		if tc.Function.Name == "think" {
			// 解析 think 工具参数
			var thinkArgs struct {
				Thought           string `json:"thought"`
				ThoughtNumber     int    `json:"thought_number"`
				TotalThoughts     int    `json:"total_thoughts"`
				NextThoughtNeeded bool   `json:"next_thought_needed"`
			}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &thinkArgs); err == nil {
				// 计算对齐 WeKnora 的字段
				// Done = !next_thought_needed && thought_number >= total_thoughts
				done := !thinkArgs.NextThoughtNeeded && thinkArgs.ThoughtNumber >= thinkArgs.TotalThoughts

				// 构建 data 字段（对齐 WeKnora StreamEvent.Data）
				data := map[string]interface{}{
					// WeKnora 核心字段（对齐 AgentThoughtData）
					"iteration": thinkArgs.ThoughtNumber, // 对齐 Iteration
					"done":      done,                    // 对齐 Done
					// 兼容字段（从 think 工具参数）
					"thought_number":      thinkArgs.ThoughtNumber,
					"total_thoughts":      thinkArgs.TotalThoughts,
					"next_thought_needed": thinkArgs.NextThoughtNeeded,
				}

				// 如果完成，添加 duration_ms（对齐 WeKnora）
				if done {
					data["duration_ms"] = 0 // 可选：计算实际耗时
				}

				// 发送 thinking 事件（对齐 WeKnora ResponseTypeThinking）
				return h.sendSSEEvent(c, SSEEvent{
					ResponseType: SSEEventThinking,
					AgentName:    event.AgentName,
					RunPath:      formatRunPath(event.RunPath),
					Content:      thinkArgs.Thought, // 对齐 WeKnora StreamEvent.Content
					Done:         done,              // 对齐 WeKnora StreamEvent.Done
					SessionID:    h.sessionID,
					ID:           h.messageID,
					Data:         data, // 对齐 WeKnora StreamEvent.Data（包含元数据）
				})
			}
		}
	}
	// 解析失败时回退到普通处理
	return h.sendSSEEvent(c, SSEEvent{
		ResponseType: SSEEventToolCall,
		AgentName:    event.AgentName,
		RunPath:      formatRunPath(event.RunPath),
		ToolCalls:    msg.ToolCalls,
		SessionID:    h.sessionID,
		ID:           h.messageID,
	})
}

// handleTodoWriteToolCall 处理 todo_write 工具调用（对齐 WeKnora）.
// 注意：这个方法在 tool_call 阶段被调用，不应该拦截发送 tool_result 事件。
// 正确的流程是：先发送 tool_call 事件，然后在 tool_result 阶段发送带 display_type 的 tool_result 事件。
// 所以这个方法现在只返回 false，让正常的 tool_call 事件发送出去。
func (h *ADKSSEHandler) handleTodoWriteToolCall(c *gin.Context, event *adk.AgentEvent, msg *schema.Message) error {
	// 不在 tool_call 阶段拦截，让正常流程处理
	// 实际的处理在 handleStreamingMessage 和 handleTodoWriteResult 中
	return h.sendSSEEvent(c, SSEEvent{
		ResponseType: SSEEventToolCall,
		AgentName:    event.AgentName,
		RunPath:      formatRunPath(event.RunPath),
		ToolCalls:    msg.ToolCalls,
		SessionID:    h.sessionID,
		ID:           h.messageID,
	})
}

// formatPlanAsMarkdown 将计划格式化为 Markdown 文本（对齐 WeKnora 输出格式）.
func (h *ADKSSEHandler) formatPlanAsMarkdown(task string, steps []struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}) string {
	statusEmoji := map[string]string{
		"pending":     "⏳",
		"in_progress": "🔄",
		"completed":   "✅",
	}

	var sb strings.Builder
	sb.WriteString("## 已创建任务计划\n\n")
	fmt.Fprintf(&sb, "**任务**: %s\n\n", task)
	sb.WriteString("**计划步骤**:\n\n")

	for i, step := range steps {
		emoji := statusEmoji[step.Status]
		if emoji == "" {
			emoji = "⏳"
		}
		fmt.Fprintf(&sb, "  %d. %s [%s] %s\n", i+1, step.ID, emoji, step.Description)
	}

	return sb.String()
}

// GenerateID 生成消息 ID.
func GenerateID(sessionID string) string {
	return fmt.Sprintf("%s-%d", sessionID, 0)
}

// sendSSEThinkingEvent 发送 thinking 事件（对齐 WeKnora handleThought）.
// WeKnora 事件流程: Agent → EventBus(EventAgentThought) → StreamManager → SSE(thinking)
func (h *ADKSSEHandler) sendSSEThinkingEvent(c *gin.Context, event *adk.AgentEvent, thinking *ThinkingEvent) error {
	// 构建 data 字段（对齐 WeKnora StreamEvent.Data）
	data := map[string]interface{}{
		// WeKnora 核心字段
		"iteration": thinking.Iteration,
		"done":      thinking.Done,
		// 兼容字段（从 think 工具参数）
		"thought_number":      thinking.ThoughtNumber,
		"total_thoughts":      thinking.TotalThoughts,
		"next_thought_needed": thinking.NextThoughtNeeded,
	}

	// 如果完成，添加 duration_ms（对齐 WeKnora）
	if thinking.Done {
		data["duration_ms"] = 0 // 可选：计算实际耗时
	}

	// 使用 Content 字段（对齐 WeKnora AgentThoughtData.Content）
	content := thinking.Content
	if content == "" {
		// 兼容旧代码：如果没有 Content，使用空字符串
		content = ""
	}

	return h.sendSSEEvent(c, SSEEvent{
		ResponseType: SSEEventThinking, // 对齐 WeKnora ResponseTypeThinking
		AgentName:    event.AgentName,
		RunPath:      formatRunPath(event.RunPath),
		Content:      content,       // 对齐 WeKnora StreamEvent.Content
		Done:         thinking.Done, // 对齐 WeKnora StreamEvent.Done
		SessionID:    h.sessionID,
		ID:           h.messageID,
		Data:         data, // 对齐 WeKnora StreamEvent.Data（包含元数据）
	})
}

// parseWebSearchResults 解析 web_search 工具的输出结果.
func (h *ADKSSEHandler) parseWebSearchResults(content string) ([]map[string]interface{}, string) {
	// 尝试解析 JSON 格式的搜索结果
	var searchResults []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	}

	if err := json.Unmarshal([]byte(content), &searchResults); err == nil {
		results := make([]map[string]interface{}, 0, len(searchResults))
		for i, r := range searchResults {
			results = append(results, map[string]interface{}{
				"result_index": i + 1,
				"title":        r.Title,
				"url":          r.URL,
				"snippet":      r.Snippet,
			})
		}
		return results, ""
	}

	// 尝试解析带 query 字段的格式
	var resultWithQuery struct {
		Query   string `json:"query"`
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Snippet string `json:"snippet"`
		} `json:"results"`
	}

	if err := json.Unmarshal([]byte(content), &resultWithQuery); err == nil {
		results := make([]map[string]interface{}, 0, len(resultWithQuery.Results))
		for i, r := range resultWithQuery.Results {
			results = append(results, map[string]interface{}{
				"result_index": i + 1,
				"title":        r.Title,
				"url":          r.URL,
				"snippet":      r.Snippet,
			})
		}
		return results, resultWithQuery.Query
	}

	// 解析失败，返回空结果
	return nil, ""
}
