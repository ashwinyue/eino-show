// Package event 提供 Agent 事件发送辅助工具.
// 对齐 WeKnora 事件类型，使用 Eino ADK 的 CustomizedOutput 扩展点.
package event

import (
	"github.com/cloudwego/eino/adk"
)

// ThinkingEvent 思考事件数据.
type ThinkingEvent struct {
	Thought           string `json:"thought"`
	ThoughtNumber     int    `json:"thought_number,omitempty"`
	TotalThoughts     int    `json:"total_thoughts,omitempty"`
	NextThoughtNeeded  bool   `json:"next_thought_needed,omitempty"`
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
	Score       int    `json:"score,omitempty"` // 1-5 分
}

// Sender 事件发送器，用于在 Agent 中发送自定义事件.
type Sender struct {
	gen *adk.AsyncGenerator[*adk.AgentEvent]
}

// NewSender 创建事件发送器.
func NewSender(gen *adk.AsyncGenerator[*adk.AgentEvent]) *Sender {
	return &Sender{gen: gen}
}

// SendThinking 发送思考事件.
func (s *Sender) SendThinking(thought string, step int, total int) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: ThinkingEvent{
				Thought:          thought,
				ThoughtNumber:    step,
				TotalThoughts:    total,
				NextThoughtNeeded: step < total,
			},
		},
	})
}

// SendThinkingSimple 发送简单思考事件（只有内容）.
func (s *Sender) SendThinkingSimple(thought string) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: thought,
		},
	})
}

// SendReferences 发送知识引用事件.
func (s *Sender) SendReferences(chunks []ReferenceChunk) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: ReferencesEvent{
				Chunks: chunks,
			},
		},
	})
}

// SendReflection 发送反思事件.
func (s *Sender) SendReflection(reflection string, score int) {
	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: ReflectionEvent{
				Reflection: reflection,
				Score:       score,
			},
		},
	})
}

// SendCustom 发送自定义事件.
func (s *Sender) SendCustom(eventType string, content string, data map[string]interface{}) {
	customData := make(map[string]interface{})
	for k, v := range data {
		customData[k] = v
	}
	customData["type"] = eventType
	if content != "" {
		customData["content"] = content
	}

	s.gen.Send(&adk.AgentEvent{
		Output: &adk.AgentOutput{
			CustomizedOutput: customData,
		},
	})
}
