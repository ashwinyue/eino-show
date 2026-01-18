package sse

import (
	"fmt"
	"strings"
)

// FormatPlanAsMarkdown 将计划步骤格式化为 Markdown（对齐 WeKnora）.
func FormatPlanAsMarkdown(task string, steps []PlanStep) string {
	var sb strings.Builder
	sb.WriteString("## 📋 任务计划\n\n")
	if task != "" {
		sb.WriteString(fmt.Sprintf("**任务**: %s\n\n", task))
	}

	for i, step := range steps {
		statusIcon := getStatusIcon(step.Status)
		sb.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, statusIcon, step.Description))
	}

	return sb.String()
}

// PlanStep 计划步骤.
type PlanStep struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// getStatusIcon 获取状态图标.
func getStatusIcon(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "in_progress":
		return "🚀"
	case "pending":
		return "📋"
	default:
		return "📋"
	}
}

// FormatRunPath 格式化运行路径.
func FormatRunPath(runPath []string) string {
	if len(runPath) == 0 {
		return ""
	}
	return strings.Join(runPath, " -> ")
}
