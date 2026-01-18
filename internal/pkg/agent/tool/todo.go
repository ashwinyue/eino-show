// Package tool 提供 Todo 工具，用于任务管理和规划.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// TodoToolName Todo 工具名称.
const TodoToolName = "todo_write"

// NewTodoTool 创建 Todo 工具.
func NewTodoTool() tool.InvokableTool {
	return &todoTool{
		// 在内存中维护多个计划，用 map 存储每个计划的状态
		plans: make(map[string]*todoPlan),
	}
}

// todoPlan 任务计划.
type todoPlan struct {
	PlanID    string     `json:"plan_id"`
	Task      string     `json:"task"`
	Steps     []planStep `json:"steps"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// todoTool 实现 Todo 任务管理工具.
type todoTool struct {
	plans map[string]*todoPlan
}

// planStep 计划步骤.
type planStep struct {
	ID          string `json:"id" jsonschema_description:"步骤的唯一标识符，如 'step1' 或 'search_knowledge'. 建议3-8个有意义的步骤"`
	Description string `json:"description" jsonschema_description:"步骤的详细描述，说明需要做什么。应该是具体且可执行的，如'搜索知识库中的竞争对手产品信息'"`
	Status      string `json:"status" jsonschema_description:"步骤状态: pending(待处理), in_progress(进行中), completed(已完成). 一次只能有一个步骤为 in_progress"`
}

// todoInput Todo 输入参数.
type todoInput struct {
	// Task 要完成的任务描述
	Task string `json:"task" jsonschema_description:"要完成的复杂任务描述，如'分析竞争对手的产品功能', '对比三家云服务商的定价策略'"`
	// Steps 计划步骤列表（可选，如果不提供则由工具生成）
	// 每个步骤包含 id, description, status
	Steps []planStep `json:"steps,omitempty" jsonschema_description:"（可选）计划步骤列表。每个步骤需要包含 id, description, status 字段。如果不提供，工具会根据 task 自动生成计划"`
}

// Info 返回工具信息.
func (t *todoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	_ = ctx
	return &schema.ToolInfo{
		Name: TodoToolName,
		Desc: `Use this tool to create and manage a structured task list for retrieval and research tasks.

**CRITICAL - Focus on Retrieval and Research Tasks Only**:
- This tool is for tracking RETRIEVAL and RESEARCH tasks
- DO NOT include summary or synthesis tasks - those are handled by the thinking tool
- DO NOT call other tools within this tool - it is purely for tracking task state

## When to Use This Tool

1. **Complex multi-step tasks (3+ distinct steps)**
   - Examples: "分析竞争对手的产品功能", "对比三家云服务商的定价策略", "研究某个技术主题"

2. **User explicitly requests todo list**
   - Examples: "创建一个任务列表来分析...", "帮我规划一下..."

3. **Tasks requiring careful planning before execution**
   - Research tasks that need a structured approach
   - Investigations with multiple angles to explore

## When NOT to Use

1. **Single, straightforward task** - Just answer directly
2. **Pure conversational or informational task** - Use thinking tool instead
3. **Tasks that can be completed in one step** - No plan needed

## Task States

- **pending**: Not yet started
- **in_progress**: Currently working on (only ONE at a time!)
- **completed**: Finished successfully

## Best Practices

1. **One Task at a Time**: Only mark ONE task as in_progress at any time
2. **Sequential Progress**: Complete in_progress tasks before marking new tasks as in_progress
3. **Clear Descriptions**: Each step should be specific and actionable
4. **3-7 Steps**: Aim for 3-7 meaningful steps (too few = not structured, too many = overwhelming)
5. **Think First**: Use the thinking tool to plan before creating a todo list

## Parameters Explained

- **task**: The complex task or question to be addressed
- **steps**: (Optional) Array of plan steps. Each step needs: id (unique identifier), description (what to do), status (pending/in_progress/completed)
`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"task": {
				Type:     "string",
				Desc:     "The complex task or question to be planned. Examples: '分析竞争对手的产品功能', '对比三家云服务商的定价策略'",
				Required: true,
			},
			"steps": {
				Type:     "array",
				Desc:     "(Optional) Array of plan steps. Each step needs: id (unique identifier), description (what to do), status (pending/in_progress/completed). If not provided, the tool will generate a plan based on the task",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具.
func (t *todoTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var input todoInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		// 兼容 steps.id 为数字的情况
		var raw map[string]interface{}
		if err2 := json.Unmarshal([]byte(argumentsInJSON), &raw); err2 == nil {
			if stepsRaw, ok := raw["steps"].([]interface{}); ok {
				steps := make([]planStep, 0, len(stepsRaw))
				for _, item := range stepsRaw {
					m, ok := item.(map[string]interface{})
					if !ok {
						continue
					}
					step := planStep{}
					if idVal, ok := m["id"]; ok {
						switch v := idVal.(type) {
						case string:
							step.ID = v
						case float64:
							step.ID = fmt.Sprintf("step%d", int(v))
						case int:
							step.ID = fmt.Sprintf("step%d", v)
						}
					}
					if desc, ok := m["description"].(string); ok {
						step.Description = desc
					}
					if status, ok := m["status"].(string); ok {
						step.Status = status
					}
					steps = append(steps, step)
				}
				input.Steps = steps
			}
			if task, ok := raw["task"].(string); ok {
				input.Task = task
			}
		} else {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	if input.Task == "" {
		return "", fmt.Errorf("task cannot be empty")
	}

	// 如果没有提供步骤，生成默认计划
	if len(input.Steps) == 0 {
		return t.generateDefaultPlan(input.Task), nil
	}

	// 创建新计划
	planID := fmt.Sprintf("plan_%d", len(t.plans)+1)
	plan := &todoPlan{
		PlanID:    planID,
		Task:      input.Task,
		Steps:     input.Steps,
		CreatedAt: currentTimeStr(),
		UpdatedAt: currentTimeStr(),
	}

	// 验证并修正步骤
	correctedSteps, err := t.validateSteps(input.Steps)
	if err != nil {
		return "", fmt.Errorf("invalid steps: %w", err)
	}
	input.Steps = correctedSteps

	// 保存计划
	plan.Steps = correctedSteps
	t.plans[planID] = plan

	return t.formatPlanOutput(plan), nil
}

// validateSteps 验证并修正步骤定义.
// 返回修正后的步骤和可能的错误（严重错误时返回错误）.
func (t *todoTool) validateSteps(steps []planStep) ([]planStep, error) {
	if len(steps) == 0 {
		return steps, nil
	}

	validStatus := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
	}

	// 第一遍：验证状态值和必填字段
	for _, step := range steps {
		if !validStatus[step.Status] {
			return nil, fmt.Errorf("invalid status '%s' for step '%s'. Must be: pending, in_progress, or completed", step.Status, step.ID)
		}
		if step.ID == "" {
			return nil, fmt.Errorf("step ID cannot be empty")
		}
		if step.Description == "" {
			return nil, fmt.Errorf("step description cannot be empty for step %s", step.ID)
		}
	}

	// 第二遍：修正不合理的步骤状态
	// 规则1：最多只能有一个 in_progress
	// 规则2：前面的步骤必须为 completed 才能将后面的步骤设为 in_progress
	// 规则3：如果多个步骤是 in_progress，只保留第一个
	corrected := make([]planStep, len(steps))
	copy(corrected, steps)

	var firstInProgressIndex = -1

	for i := range corrected {
		if corrected[i].Status == "in_progress" && firstInProgressIndex < 0 {
			firstInProgressIndex = i
		}
	}

	for i := range corrected {
		// 如果当前步骤是 in_progress，但不是第一个 in_progress
		if corrected[i].Status == "in_progress" && i != firstInProgressIndex {
			// 将多余的 in_progress 改为 pending
			corrected[i].Status = "pending"
		}

		// 如果前面的步骤未完成，当前步骤不应该是 in_progress 或 completed
		if i > 0 && (corrected[i].Status == "in_progress" || corrected[i].Status == "completed") {
			if corrected[i-1].Status == "pending" {
				// 修正：将当前步骤改为 pending，除非第一个 in_progress 已经设置
				if i == firstInProgressIndex {
					// 这是第一个 in_progress，但前一步是 pending
					// 保持第一个 in_progress 状态不变，让用户自己决定
				} else {
					corrected[i].Status = "pending"
				}
			}
		}
	}

	return corrected, nil
}

// generateDefaultPlan 生成默认计划（当用户未提供步骤时）.
func (t *todoTool) generateDefaultPlan(task string) string {
	// 根据任务类型生成合理的计划
	steps := t.generateStepsFromTask(task)

	// 创建计划并保存
	planID := fmt.Sprintf("plan_%d", len(t.plans)+1)
	plan := &todoPlan{
		PlanID:    planID,
		Task:      task,
		Steps:     steps,
		CreatedAt: currentTimeStr(),
		UpdatedAt: currentTimeStr(),
	}
	t.plans[planID] = plan

	// 返回 JSON 格式，与 formatPlanOutput 保持一致
	return t.formatPlanOutput(plan)
}

// generateStepsFromTask 根据任务描述生成合理的步骤.
func (t *todoTool) generateStepsFromTask(task string) []planStep {
	// 简单的关键词匹配来生成计划
	taskLower := strings.ToLower(task)

	// 检查是否包含比较/分析类关键词
	if containsAny(taskLower, []string{"对比", "比较", "分析", "评估", "测试"}) {
		if containsAny(taskLower, []string{"产品", "服务", "方案", "技术", "功能", "模型", "API"}) {
			return []planStep{
				{ID: "step1", Description: "收集目标对象的信息", Status: "pending"},
				{ID: "step2", Description: "收集对比对象的信息", Status: "pending"},
				{ID: "step3", Description: "对比关键维度（功能、性能、价格等）", Status: "pending"},
				{ID: "step4", Description: "生成对比分析报告", Status: "pending"},
				{ID: "step5", Description: "得出结论和建议", Status: "pending"},
			}
		}
	}

	// 默认计划
	return []planStep{
		{ID: "step1", Description: "理解任务需求和范围", Status: "pending"},
		{ID: "step2", Description: "收集相关信息和数据", Status: "pending"},
		{ID: "step3", Description: "分析整理收集到的信息", Status: "pending"},
		{ID: "step4", Description: "生成总结或答案", Status: "pending"},
	}
}

// containsAny 检查字符串是否包含任意一个关键词.
func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

// formatPlanOutput 格式化计划输出（对齐 WeKnora，返回 JSON 格式）.
func (t *todoTool) formatPlanOutput(plan *todoPlan) string {
	if plan == nil {
		return `{"error": "No active plan"}`
	}

	// 构造 JSON 格式的输出，包含前端需要的所有字段
	result := map[string]interface{}{
		"display_type": "plan",
		"task":         plan.Task,
		"plan_id":      plan.PlanID,
		"steps":        plan.Steps,
		"total_steps":  len(plan.Steps),
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		// 回退到 Markdown 格式
		return t.formatPlanOutputMarkdown(plan)
	}

	return string(jsonBytes)
}

// formatPlanOutputMarkdown 格式化计划输出为 Markdown 格式（备用）.
func (t *todoTool) formatPlanOutputMarkdown(plan *todoPlan) string {
	if plan == nil {
		return "No active plan"
	}

	output := fmt.Sprintf("## 任务计划已创建\n\n**任务**: %s\n", plan.Task)
	output += fmt.Sprintf("**计划 ID**: %s\n\n", plan.PlanID)

	// 统计任务状态
	var pendingCount, inProgressCount, completedCount int
	for _, step := range plan.Steps {
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
	for i, step := range plan.Steps {
		output += t.formatPlanStep(i+1, step)
	}

	// 显示进度
	totalCount := len(plan.Steps)
	remainingCount := pendingCount + inProgressCount

	output += "\n### 任务进度\n"
	output += fmt.Sprintf("- 总计: %d 个任务\n", totalCount)
	output += fmt.Sprintf("- ✅ 已完成: %d 个\n", completedCount)
	output += fmt.Sprintf("- 🔄 进行中: %d 个\n", inProgressCount)
	output += fmt.Sprintf("- ⏳ 待处理: %d 个\n", pendingCount)

	if remainingCount > 0 {
		output += fmt.Sprintf("\n**还有 %d 个任务未完成！**\n", remainingCount)
		output += "完成后可进行总结或得出结论。\n"
	} else {
		output += "\n✅ **所有任务已完成！**\n"
		output += "可以综合所有发现生成最终答案。\n"
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

	return fmt.Sprintf("  %d. %s [%s] %s\n", index, step.ID, emoji, step.Description)
}

// formatStepsList 格式化步骤列表（用于默认计划）.
func (t *todoTool) formatStepsList(steps []planStep) string {
	var output strings.Builder
	for i, step := range steps {
		output.WriteString(t.formatPlanStep(i+1, step))
	}
	return output.String()
}

// GetCurrentPlan 获取当前计划.
func (t *todoTool) GetCurrentPlan() *todoPlan {
	if len(t.plans) == 0 {
		return nil
	}

	// 返回最新的计划
	var latestPlan *todoPlan
	var latestTime string

	for _, plan := range t.plans {
		if plan.CreatedAt > latestTime {
			latestPlan = plan
			latestTime = plan.CreatedAt
		}
	}

	return latestPlan
}

// GetPlanByID 根据 ID 获取特定计划.
func (t *todoTool) GetPlanByID(planID string) *todoPlan {
	return t.plans[planID]
}

// UpdateStepStatus 更新步骤状态.
func (t *todoTool) UpdateStepStatus(planID, stepID string, newStatus string) (string, error) {
	plan := t.plans[planID]
	if plan == nil {
		return "", fmt.Errorf("plan not found: %s", planID)
	}

	// 查找并更新步骤
	updated := false
	for i := range plan.Steps {
		if plan.Steps[i].ID == stepID {
			// 验证新状态
			validStatus := map[string]bool{
				"pending":     true,
				"in_progress": true,
				"completed":   true,
			}
			if !validStatus[newStatus] {
				return "", fmt.Errorf("invalid status '%s'. Must be: pending, in_progress, or completed", newStatus)
			}

			// 只允许有一个进行中的任务
			if newStatus == "in_progress" {
				for j, step := range plan.Steps {
					if j != i && step.Status == "in_progress" {
						return "", fmt.Errorf("cannot mark step %s as in_progress, because step %s is already in_progress", stepID, step.ID)
					}
				}
			}

			plan.Steps[i].Status = newStatus
			plan.UpdatedAt = currentTimeStr()
			updated = true
			break
		}
	}

	if !updated {
		return "", fmt.Errorf("step not found: %s", stepID)
	}

	return t.formatPlanOutput(plan), nil
}

// Clear 清除所有计划.
func (t *todoTool) Clear() string {
	t.plans = make(map[string]*todoPlan)
	return "所有计划已清除。"
}

// currentTimeStr 返回当前时间的字符串格式.
func currentTimeStr() string {
	return fmt.Sprintf("%d", getCurrentTimestamp())
}

// getCurrentTimestamp 返回当前 Unix 时间戳.
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
