// Package tools 提供 Agent 可用的工具集合.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/eino-show/pkg/log"
)

type thinkingEventKey struct{}

type ThinkingEvent struct {
	Thought           string
	ThoughtNumber     int
	TotalThoughts     int
	NextThoughtNeeded bool
}

type ThinkingEmitter func(ThinkingEvent)

func WithThinkingEmitter(ctx context.Context, emitter ThinkingEmitter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, thinkingEventKey{}, emitter)
}

func GetThinkingEmitter(ctx context.Context) (ThinkingEmitter, bool) {
	if ctx == nil {
		return nil, false
	}
	v := ctx.Value(thinkingEventKey{})
	if v == nil {
		return nil, false
	}
	emitter, ok := v.(ThinkingEmitter)
	return emitter, ok
}

// ToolThinking 是 thinking 工具的名称.
const ToolThinking = "thinking"

// SequentialThinkingInput 定义 thinking 工具的输入参数.
type SequentialThinkingInput struct {
	Thought           string `json:"thought"`
	NextThoughtNeeded bool   `json:"next_thought_needed"`
	ThoughtNumber     int    `json:"thought_number"`
	TotalThoughts     int    `json:"total_thoughts"`
	IsRevision        bool   `json:"is_revision,omitempty"`
	RevisesThought    *int   `json:"revises_thought,omitempty"`
	BranchFromThought *int   `json:"branch_from_thought,omitempty"`
	BranchID          string `json:"branch_id,omitempty"`
	NeedsMoreThoughts bool   `json:"needs_more_thoughts,omitempty"`
}

// SequentialThinkingTool 是一个动态反思问题解决工具.
type SequentialThinkingTool struct {
	thoughtHistory []SequentialThinkingInput
	branches       map[string][]SequentialThinkingInput
}

// NewSequentialThinkingTool 创建一个新的 thinking 工具实例.
func NewSequentialThinkingTool() *SequentialThinkingTool {
	return &SequentialThinkingTool{
		thoughtHistory: make([]SequentialThinkingInput, 0),
		branches:       make(map[string][]SequentialThinkingInput),
	}
}

// Info 返回工具信息.
func (t *SequentialThinkingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolThinking,
		Desc: `A detailed tool for dynamic and reflective problem-solving through thoughts.

This tool helps analyze problems through a flexible thinking process that can adapt and evolve.

## When to Use This Tool
- Breaking down complex problems into steps
- Planning and design with room for revision
- Analysis that might need course correction
- Problems where the full scope might not be clear initially

## Parameters
- thought: Your current thinking step in natural, user-friendly language
- next_thought_needed: Whether another thought step is needed
- thought_number: Current thought number (1, 2, 3...)
- total_thoughts: Estimated total thoughts needed`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"thought": {
				Type:     schema.String,
				Desc:     "Your current thinking step. Write in natural, user-friendly language.",
				Required: true,
			},
			"next_thought_needed": {
				Type:     schema.Boolean,
				Desc:     "Whether another thought step is needed",
				Required: true,
			},
			"thought_number": {
				Type:     schema.Integer,
				Desc:     "Current thought number",
				Required: true,
			},
			"total_thoughts": {
				Type:     schema.Integer,
				Desc:     "Estimated total thoughts needed",
				Required: true,
			},
			"is_revision": {
				Type: schema.Boolean,
				Desc: "Whether this revises previous thinking",
			},
			"revises_thought": {
				Type: schema.Integer,
				Desc: "Which thought is being reconsidered",
			},
			"branch_from_thought": {
				Type: schema.Integer,
				Desc: "Branching point thought number",
			},
			"branch_id": {
				Type: schema.String,
				Desc: "Branch identifier",
			},
			"needs_more_thoughts": {
				Type: schema.Boolean,
				Desc: "If more thoughts are needed",
			},
		}),
	}, nil
}

// InvokableRun 执行 thinking 工具.
func (t *SequentialThinkingTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	log.Infow("SequentialThinkingTool executing", "args", argumentsInJSON)

	var input SequentialThinkingInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("failed to parse args: %w", err)
	}

	// 兼容模型偶发的 total_thorts 拼写错误
	if input.TotalThoughts == 0 {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err == nil {
			if v, ok := raw["total_thorts"]; ok {
				if num, ok := v.(float64); ok {
					input.TotalThoughts = int(num)
				}
			}
		}
	}

	// total_thoughts 缺失或小于1时，用 thought_number 兜底
	if input.ThoughtNumber < 1 {
		input.ThoughtNumber = 1
	}
	if input.TotalThoughts < 1 {
		input.TotalThoughts = input.ThoughtNumber
	}

	if err := t.validate(input); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// 调整 totalThoughts
	if input.ThoughtNumber > input.TotalThoughts {
		input.TotalThoughts = input.ThoughtNumber
	}

	// 添加到历史
	t.thoughtHistory = append(t.thoughtHistory, input)

	// 处理分支
	if input.BranchFromThought != nil && input.BranchID != "" {
		if t.branches[input.BranchID] == nil {
			t.branches[input.BranchID] = make([]SequentialThinkingInput, 0)
		}
		t.branches[input.BranchID] = append(t.branches[input.BranchID], input)
	}

	incomplete := input.NextThoughtNeeded || input.NeedsMoreThoughts ||
		input.ThoughtNumber < input.TotalThoughts

	// 返回结果
	result := map[string]interface{}{
		"thought_number":         input.ThoughtNumber,
		"total_thoughts":         input.TotalThoughts,
		"next_thought_needed":    input.NextThoughtNeeded,
		"thought_history_length": len(t.thoughtHistory),
		"thought":                input.Thought,
		"incomplete_steps":       incomplete,
	}

	if emitter, ok := GetThinkingEmitter(ctx); ok && emitter != nil {
		emitter(ThinkingEvent{
			Thought:           input.Thought,
			ThoughtNumber:     input.ThoughtNumber,
			TotalThoughts:     input.TotalThoughts,
			NextThoughtNeeded: input.NextThoughtNeeded,
		})
	}

	resultJSON, _ := json.Marshal(result)

	log.Infow("SequentialThinkingTool completed",
		"thought_number", input.ThoughtNumber,
		"total_thoughts", input.TotalThoughts)

	if incomplete {
		return string(resultJSON), nil
	}
	return string(resultJSON), nil
}

// validate 验证输入.
func (t *SequentialThinkingTool) validate(data SequentialThinkingInput) error {
	if data.Thought == "" {
		return fmt.Errorf("invalid thought: must be a non-empty string")
	}
	if data.ThoughtNumber < 1 {
		return fmt.Errorf("invalid thoughtNumber: must be >= 1")
	}
	if data.TotalThoughts < 1 {
		return fmt.Errorf("invalid totalThoughts: must be >= 1")
	}
	return nil
}

// GetThoughtHistory 返回思考历史.
func (t *SequentialThinkingTool) GetThoughtHistory() []SequentialThinkingInput {
	return t.thoughtHistory
}

// GetLastThought 返回最后一个思考.
func (t *SequentialThinkingTool) GetLastThought() *SequentialThinkingInput {
	if len(t.thoughtHistory) == 0 {
		return nil
	}
	return &t.thoughtHistory[len(t.thoughtHistory)-1]
}
