// Package tool 提供 Todo 工具，用于任务管理.
package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// TodoToolName Todo 工具名称.
const TodoToolName = "todo"

// NewTodoTool 创建 Todo 工具.
func NewTodoTool() tool.InvokableTool {
	return &todoTool{}
}

// todoTool 任务管理工具.
type todoTool struct {
	currentPlan *todoPlan
}

// todoPlan 任务计划.
type todoPlan struct {
	Task  string      `json:"task"`
	Steps []planStep  `json:"steps"`
}

// planStep 计划步骤.
type planStep struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"` // pending, in_progress, completed
}

// todoInput Todo 输入参数.
type todoInput struct {
	Task  string     `json:"task"`
	Steps []planStep `json:"steps"`
}

// Info 返回工具信息.
func (t *todoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	_ = ctx
	return &schema.ToolInfo{
		Name: TodoToolName,
		Desc: `Use this tool to create and manage a structured task list for retrieval and research tasks.

**CRITICAL - Focus on Retrieval Tasks Only**:
- This tool is for tracking RETRIEVAL and RESEARCH tasks
- DO NOT include summary or synthesis tasks - those are handled by the thinking tool

## When to Use
1. Complex multi-step tasks (3+ distinct steps)
2. Non-trivial tasks requiring careful planning
3. User explicitly requests todo list
4. User provides multiple tasks
5. Mark task as in_progress BEFORE starting work
6. Mark task as completed after finishing

## When NOT to Use
1. Single, straightforward task
2. Trivial task with no organizational benefit
3. Purely conversational or informational task

## Task States
- pending: Not yet started
- in_progress: Currently working on (limit to ONE at a time)
- completed: Finished successfully

## Parameters
- task: The complex task or question
- steps: Array of plan steps with id, description, status`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"task": {
				Type:     "string",
				Desc:     "The complex task or question",
				Required: true,
			},
			"steps": {
				Type:     "array",
				Desc:     "Array of plan steps with status tracking",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具.
func (t *todoTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var input todoInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if input.Task == "" {
		input.Task = "未提供任务描述"
	}

	// 更新当前计划
	t.currentPlan = &todoPlan{
		Task:  input.Task,
		Steps: input.Steps,
	}

	return t.generatePlanOutput(), nil
}

// generatePlanOutput 生成计划输出.
func (t *todoTool) generatePlanOutput() string {
	if t.currentPlan == nil {
		return "No active plan"
	}

	output := "## 计划已创建\n\n"
	output += fmt.Sprintf("**任务**: %s\n\n", t.currentPlan.Task)

	if len(t.currentPlan.Steps) == 0 {
		output += "注意：未提供具体步骤。建议创建3-7个检索任务。\n"
		return output
	}

	// 统计任务状态
	var pendingCount, inProgressCount, completedCount int
	for _, step := range t.currentPlan.Steps {
		switch step.Status {
		case "pending":
			pendingCount++
		case "in_progress":
			inProgressCount++
		case "completed":
			completedCount++
		}
	}

	// 显示所有步骤
	output += "**计划步骤**:\n\n"
	for i, step := range t.currentPlan.Steps {
		output += t.formatPlanStep(i+1, step)
	}

	// 显示进度
	totalCount := len(t.currentPlan.Steps)
	remainingCount := pendingCount + inProgressCount

	output += "\n### 任务进度\n"
	output += fmt.Sprintf("- 总计: %d 个任务\n", totalCount)
	output += fmt.Sprintf("- ✅ 已完成: %d 个\n", completedCount)
	output += fmt.Sprintf("- 🔄 进行中: %d 个\n", inProgressCount)
	output += fmt.Sprintf("- ⏳ 待处理: %d 个\n", pendingCount)

	if remainingCount > 0 {
		output += fmt.Sprintf("\n**还有 %d 个任务未完成！**\n", remainingCount)
		output += "完成所有任务后才能总结或得出结论。\n"
	} else {
		output += "\n✅ **所有任务已完成！**\n"
		output += "现在可以综合所有发现生成最终答案。\n"
	}

	return output
}

// formatPlanStep 格式化计划步骤.
func (t *todoTool) formatPlanStep(index int, step planStep) string {
	statusEmoji := map[string]string{
		"pending":     "⏳",
		"in_progress": "🔄",
		"completed":   "✅",
	}

	emoji := statusEmoji[step.Status]
	if emoji == "" {
		emoji = "⏳"
	}

	return fmt.Sprintf("  %d. %s [%s] %s\n", index, emoji, step.Status, step.Description)
}

// GetCurrentPlan 获取当前计划.
func (t *todoTool) GetCurrentPlan() *todoPlan {
	return t.currentPlan
}

// UpdateStepStatus 更新步骤状态.
func (t *todoTool) UpdateStepStatus(stepID string, newStatus string) error {
	if t.currentPlan == nil {
		return fmt.Errorf("no active plan")
	}

	for i := range t.currentPlan.Steps {
		if t.currentPlan.Steps[i].ID == stepID {
			t.currentPlan.Steps[i].Status = newStatus
			return nil
		}
	}

	return fmt.Errorf("step not found: %s", stepID)
}

// Clear 清除当前计划.
func (t *todoTool) Clear() {
	t.currentPlan = nil
}
