// Package tool 提供 Think 工具，用于序列化思考.
package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ThinkToolName Think 工具名称.
const ThinkToolName = "think"

// NewThinkTool 创建 Think 工具.
func NewThinkTool() tool.InvokableTool {
	return &thinkTool{}
}

// thinkTool 序列化思考工具.
type thinkTool struct {
	thoughtHistory []thinkInput
	branches       map[string][]thinkInput
}

// thinkInput Think 输入参数.
type thinkInput struct {
	Thought           string  `json:"thought"`
	NextThoughtNeeded  bool   `json:"next_thought_needed"`
	ThoughtNumber      int    `json:"thought_number"`
	TotalThoughts      int    `json:"total_thoughts"`
	IsRevision         bool   `json:"is_revision,omitempty"`
	RevisesThought     *int   `json:"revises_thought,omitempty"`
	BranchFromThought  *int   `json:"branch_from_thought,omitempty"`
	BranchID           string `json:"branch_id,omitempty"`
	NeedsMoreThoughts  bool   `json:"needs_more_thoughts,omitempty"`
}

// Info 返回工具信息.
func (t *thinkTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	_ = ctx
	return &schema.ToolInfo{
		Name: ThinkToolName,
		Desc: `A detailed tool for dynamic and reflective problem-solving through thoughts.

This tool helps analyze problems through a flexible thinking process that can adapt and evolve.

## When to Use
- Breaking down complex problems into steps
- Planning and design with room for revision
- Analysis that might need course correction
- Problems that require multi-step solution
- Tasks that need to maintain context over multiple steps

## Key Features
- You can adjust total_thoughts up or down as you progress
- You can question or revise previous thoughts
- You can add more thoughts even after reaching what seemed like the end
- You can express uncertainty and explore alternative approaches
- Not every thought needs to build linearly - you can branch or backtrack

## Parameters
- thought: Your current thinking step
- next_thought_needed: Whether another thought step is needed
- thought_number: Current thought number (>= 1)
- total_thoughts: Estimated total thoughts needed (>= 1)
- is_revision: Whether this revises previous thinking
- revises_thought: Which thought is being reconsidered
- branch_from_thought: Branching point thought number
- branch_id: Branch identifier
- needs_more_thoughts: If more thoughts are needed`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"thought": {
				Type:     "string",
				Desc:     "Your current thinking step",
				Required: true,
			},
			"next_thought_needed": {
				Type:     "boolean",
				Desc:     "Whether another thought step is needed",
				Required: true,
			},
			"thought_number": {
				Type:     "integer",
				Desc:     "Current thought number (>= 1)",
				Required: true,
			},
			"total_thoughts": {
				Type:     "integer",
				Desc:     "Estimated total thoughts needed (>= 1)",
				Required: true,
			},
			"is_revision": {
				Type:     "boolean",
				Desc:     "Whether this revises previous thinking",
			},
			"revises_thought": {
				Type:     "integer",
				Desc:     "Which thought is being reconsidered",
			},
			"branch_from_thought": {
				Type:     "integer",
				Desc:     "Branching point thought number",
			},
			"branch_id": {
				Type:     "string",
				Desc:     "Branch identifier",
			},
			"needs_more_thoughts": {
				Type:     "boolean",
				Desc:     "If more thoughts are needed",
			},
		}),
	}, nil
}

// InvokableRun 执行工具.
func (t *thinkTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var input thinkInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 验证输入
	if err := t.validate(input); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// 调整 totalThoughts 如果 thoughtNumber 超过它
	if input.ThoughtNumber > input.TotalThoughts {
		input.TotalThoughts = input.ThoughtNumber
	}

	// 添加到思考历史
	t.thoughtHistory = append(t.thoughtHistory, input)

	// 处理分支
	if input.BranchFromThought != nil && input.BranchID != "" {
		if t.branches == nil {
			t.branches = make(map[string][]thinkInput)
		}
		if t.branches[input.BranchID] == nil {
			t.branches[input.BranchID] = make([]thinkInput, 0)
		}
		t.branches[input.BranchID] = append(t.branches[input.BranchID], input)
	}

	// 判断是否未完成
	incomplete := input.NextThoughtNeeded || input.NeedsMoreThoughts ||
		input.ThoughtNumber < input.TotalThoughts

	// 构建响应
	output := fmt.Sprintf("Thought %d/%d: %s", input.ThoughtNumber, input.TotalThoughts, input.Thought)
	if incomplete {
		output += " - More thoughts needed"
	}

	return output, nil
}

// validate 验证输入.
func (t *thinkTool) validate(input thinkInput) error {
	if input.Thought == "" {
		return fmt.Errorf("thought must be non-empty")
	}
	if input.ThoughtNumber < 1 {
		return fmt.Errorf("thought_number must be >= 1")
	}
	if input.TotalThoughts < 1 {
		return fmt.Errorf("total_thoughts must be >= 1")
	}
	return nil
}

// GetThoughtHistory 获取思考历史.
func (t *thinkTool) GetThoughtHistory() []thinkInput {
	return t.thoughtHistory
}

// Clear 清除思考历史.
func (t *thinkTool) Clear() {
	t.thoughtHistory = nil
	t.branches = nil
}
