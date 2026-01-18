// Package event 提供在 Agent 中发送自定义事件的示例.
package event

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ExampleAgentWithEvents 演示如何在 Agent 中发送自定义事件.
// 对齐 WeKnora 的事件类型: thinking, references, reflection, answer 等.
type ExampleAgentWithEvents struct {
	name  string
	tools []tool.BaseTool
}

// NewExampleAgentWithEvents 创建示例 Agent.
func NewExampleAgentWithEvents(name string, tools []tool.BaseTool) *ExampleAgentWithEvents {
	return &ExampleAgentWithEvents{
		name:  name,
		tools: tools,
	}
}

// Name 返回 Agent 名称.
func (a *ExampleAgentWithEvents) Name(ctx context.Context) string {
	return a.name
}

// Description 返回 Agent 描述.
func (a *ExampleAgentWithEvents) Description(ctx context.Context) string {
	return "An example agent that demonstrates custom event sending"
}

// Run 执行 Agent 逻辑并发送自定义事件.
func (a *ExampleAgentWithEvents) Run(ctx context.Context, input *adk.AgentInput, opts ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer gen.Close()

		// 创建事件发送器
		sender := NewSender(gen)

		// 示例 1: 发送思考事件 (对齐 WeKnora ResponseTypeThinking)
		sender.SendThinking("正在分析用户查询...", 1, 3)

		// ... Agent 逻辑 ...

		sender.SendThinking("正在检索知识库...", 2, 3)

		// 示例 2: 发送知识引用事件 (对齐 WeKnora ResponseTypeReferences)
		sender.SendReferences([]ReferenceChunk{
			{
				ID:          "chunk-1",
				Content:     "相关知识点 1",
				Score:       0.95,
				KnowledgeID: "kb-1",
			},
			{
				ID:          "chunk-2",
				Content:     "相关知识点 2",
				Score:       0.87,
				KnowledgeID: "kb-1",
			},
		})

		sender.SendThinking("正在生成回答...", 3, 3)

		// 示例 3: 发送最终答案 (对齐 WeKnora ResponseTypeAnswer)
		// 注意: 最终答案通常通过 MessageOutput 发送，但也可以用自定义事件
		gen.Send(&adk.AgentEvent{
			Output: &adk.AgentOutput{
				MessageOutput: &adk.MessageVariant{
					Message: schema.AssistantMessage("这是根据分析得出的答案。", nil),
				},
			},
		})

		// 示例 4: 发送反思事件 (对齐 WeKnora ResponseTypeReflection)
		sender.SendReflection("我确认这个答案准确回答了用户的问题。", 5)

		// 发送完成事件
		gen.Send(&adk.AgentEvent{
			Action: &adk.AgentAction{Exit: true},
		})
	}()

	return iter
}

// 使用示例:
//
// iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
// sender := event.NewSender(gen)
//
// // 发送 thinking 事件
// sender.SendThinking("正在分析...", 1, 3)
//
// // 发送简单 thinking
// sender.SendThinkingSimple("快速思考")
//
// // 发送 references 事件
// sender.SendReferences([]event.ReferenceChunk{
//     {ID: "1", Content: "知识片段", Score: 0.9},
// })
//
// // 发送 reflection 事件
// sender.SendReflection("答案评估", 5)
//
// // 发送自定义事件
// sender.SendCustom("custom_type", "内容", map[string]interface{}{
//     "key": "value",
// })
