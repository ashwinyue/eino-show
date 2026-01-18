// Package sse 提供 SSE（Server-Sent Events）协议封装.
package sse

// EventType SSE 事件类型（对齐 WeKnora）.
type EventType string

const (
	// EventTypeQuery 查询开始事件（对齐 WeKnora 使用 agent_query）
	EventTypeQuery EventType = "agent_query"
	// EventTypeAnswer Agent 最终答案（流式）
	EventTypeAnswer EventType = "answer"
	// EventTypeThinking Agent 思考过程
	EventTypeThinking EventType = "thinking"
	// EventTypeToolCall 工具调用
	EventTypeToolCall EventType = "tool_call"
	// EventTypeToolResult 工具执行结果
	EventTypeToolResult EventType = "tool_result"
	// EventTypeReferences 知识引用
	EventTypeReferences EventType = "references"
	// EventTypeReflection 自我反思
	EventTypeReflection EventType = "reflection"
	// EventTypeAction Agent 动作（转移、中断、退出）
	EventTypeAction EventType = "action"
	// EventTypeComplete 完成事件（对齐 WeKnora 使用 stop）
	EventTypeComplete EventType = "stop"
	// EventTypeError 错误事件
	EventTypeError EventType = "error"
)

// Event SSE 事件结构（完全对齐 WeKnora）.
type Event struct {
	// 事件类型
	Type EventType `json:"response_type"`
	// 消息 ID
	ID string `json:"id"`
	// 内容
	Content string `json:"content,omitempty"`
	// 是否完成
	Done bool `json:"done,omitempty"`
	// Agent 名称
	AgentName string `json:"agent_name,omitempty"`
	// 运行路径
	RunPath string `json:"run_path,omitempty"`
	// 工具调用
	ToolCalls interface{} `json:"tool_calls,omitempty"`
	// 动作类型
	ActionType string `json:"action_type,omitempty"`
	// 错误信息
	Error string `json:"error,omitempty"`
	// 会话 ID
	SessionID string `json:"session_id,omitempty"`
	// 额外数据
	Data map[string]interface{} `json:"data,omitempty"`
	// 助手消息 ID（agent_query 专用）
	AssistantMessageID string `json:"assistant_message_id,omitempty"`
}

// ThinkingEvent 思考事件数据（对齐 WeKnora AgentThoughtData）.
type ThinkingEvent struct {
	Content   string `json:"content"`   // 思考内容
	Iteration int    `json:"iteration"` // 当前迭代次数
	Done      bool   `json:"done"`      // 是否完成思考

	// 兼容字段
	ThoughtNumber     int  `json:"thought_number,omitempty"`
	TotalThoughts     int  `json:"total_thoughts,omitempty"`
	NextThoughtNeeded bool `json:"next_thought_needed,omitempty"`
}

// ReferencesEvent 知识引用事件数据.
type ReferencesEvent struct {
	Chunks []ReferenceChunk `json:"chunks"`
}

// ReferenceChunk 知识分块引用.
type ReferenceChunk struct {
	ID          string  `json:"id"`
	Content     string  `json:"content"`
	Score       float64 `json:"score"`
	KnowledgeID string  `json:"knowledge_id,omitempty"`
}

// ReflectionEvent 反思事件数据.
type ReflectionEvent struct {
	Reflection string `json:"reflection"`
	Score      int    `json:"score,omitempty"`
}
