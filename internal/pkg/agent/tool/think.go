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

// Info 返回工具信息.
func (t *thinkTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	_ = ctx
	return &schema.ToolInfo{
		Name: ThinkToolName,
		Desc: `A detailed tool for dynamic and reflective problem-solving through thoughts.

This tool helps analyze problems through a flexible thinking process that can adapt and evolve.

Each thought can build on, question, or revise previous insights as understanding deepens.

## When to Use This Tool

- Breaking down complex problems into steps
- Planning and design with room for revision
- Analysis that might need course correction
- Problems where the full scope might not be clear initially
- Problems that require a multi-step solution
- Tasks that need to maintain context over multiple steps
- Situations where irrelevant information needs to be filtered out

## Key Features

- You can adjust total_thoughts up or down as you progress
- You can question or revise previous thoughts
- You can add more thoughts even after reaching what seemed like the end
- You can express uncertainty and explore alternative approaches
- Not every thought needs to build linearly - you can branch or backtrack
- Generates a solution hypothesis
- Verifies the hypothesis based on the Chain of Thought steps
- Repeats the process until satisfied
- Provides a correct answer

## Parameters Explained

- **thought**: Your current thinking step, which can include:
  * Regular analytical steps
  * Revisions of previous thoughts
  * Questions about previous decisions
  * Realizations about needing more analysis
  * Changes in approach
  * Hypothesis generation
  * Hypothesis verification
  
  **CRITICAL - User-Friendly Thinking**: Write your thoughts in natural, user-friendly language. NEVER mention tool names (like "grep_chunks", "knowledge_search", "web_search", etc.) in your thinking process. Instead, describe your actions in plain language:
  - BAD: "I'll use grep_chunks to search for keywords, then knowledge_search for semantic understanding"
  - GOOD: "I'll start by searching for key terms in the knowledge base, then explore related concepts"
  - BAD: "After grep_chunks returns results, I'll use knowledge_search"
  - GOOD: "After finding relevant documents, I'll search for semantically related content"
  
  Write thinking as if explaining your reasoning to a user, not documenting technical steps. Focus on WHAT you're trying to find and WHY, not HOW (which tools you'll use).

- **next_thought_needed**: True if you need more thinking, even if at what seemed like the end
- **thought_number**: Current number in sequence (can go beyond initial total if needed)
- **total_thoughts**: Current estimate of thoughts needed (can be adjusted up/down)
- **is_revision**: A boolean indicating if this thought revises previous thinking
- **revises_thought**: If is_revision is true, which thought number is being reconsidered
- **branch_from_thought**: If branching, which thought number is the branching point
- **branch_id**: Identifier for the current branch (if any)
- **needs_more_thoughts**: If reaching end but realizing more thoughts needed

## Best Practices

1. Start with an initial estimate of needed thoughts, but be ready to adjust
2. Feel free to question or revise previous thoughts
3. Don't hesitate to add more thoughts if needed, even at the "end"
4. Express uncertainty when present
5. Mark thoughts that revise previous thinking or branch into new paths
6. Ignore information that is irrelevant to the current step
7. Generate a solution hypothesis when appropriate
8. Verify the hypothesis based on the Chain of Thought steps
9. Repeat the process until satisfied with the solution
10. Provide a single, ideally correct answer as the final output
11. Only set next_thought_needed to false when truly done and a satisfactory answer is reached`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"thought": {
				Type:     "string",
				Desc:     "Your current thinking step. Write in natural, user-friendly language. NEVER mention tool names. Instead, describe actions in plain language. Focus on WHAT you're trying to find and WHY, not HOW.",
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
				Type: "boolean",
				Desc: "Whether this revises previous thinking",
			},
			"revises_thought": {
				Type: "integer",
				Desc: "Which thought is being reconsidered",
			},
			"branch_from_thought": {
				Type: "integer",
				Desc: "Branching point thought number",
			},
			"branch_id": {
				Type: "string",
				Desc: "Branch identifier",
			},
			"needs_more_thoughts": {
				Type: "boolean",
				Desc: "If more thoughts are needed",
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

	// total_thoughts 缺失时，用 thought_number 兜底
	if input.TotalThoughts == 0 && input.ThoughtNumber > 0 {
		input.TotalThoughts = input.ThoughtNumber
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
