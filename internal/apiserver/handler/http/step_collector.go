// Package http 提供 HTTP 处理器.
package http

import (
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// StepCollector AgentStep 收集器，用于收集 agent 执行过程中的步骤.
type StepCollector struct {
	steps       []v1.AgentStep
	currentStep *v1.AgentStep
	iteration   int
}

// NewStepCollector 创建 StepCollector.
func NewStepCollector() *StepCollector {
	return &StepCollector{
		steps: make([]v1.AgentStep, 0),
	}
}

// ensureCurrentStep 确保有当前 step，没有则创建新的.
func (c *StepCollector) ensureCurrentStep() {
	if c.currentStep == nil {
		c.currentStep = &v1.AgentStep{
			Iteration: c.iteration,
			ToolCalls: make([]v1.ToolCall, 0),
		}
		c.iteration++
	}
}

// CollectThought 收集思考内容到当前 step.
func (c *StepCollector) CollectThought(thought string) {
	if thought == "" {
		return
	}
	c.ensureCurrentStep()
	if c.currentStep.Thought != "" {
		c.currentStep.Thought += "\n" + thought
	} else {
		c.currentStep.Thought = thought
	}
}

// CollectToolCall 收集工具调用到当前 step.
func (c *StepCollector) CollectToolCall(toolCallID, toolName, args string) {
	c.ensureCurrentStep()
	c.currentStep.ToolCalls = append(c.currentStep.ToolCalls, v1.ToolCall{
		ID:   toolCallID,
		Name: toolName,
		Args: args,
	})
}

// CollectToolResult 收集工具结果到对应的 tool_call.
func (c *StepCollector) CollectToolResult(toolCallID string, success bool, output, errMsg string, data map[string]interface{}) {
	if c.currentStep == nil {
		return
	}
	for i := range c.currentStep.ToolCalls {
		if c.currentStep.ToolCalls[i].ID == toolCallID {
			c.currentStep.ToolCalls[i].Result = &v1.ToolCallResult{
				Success: success,
				Output:  output,
				Error:   errMsg,
				Data:    data,
			}
			break
		}
	}
}

// FinalizeCurrentStep 完成当前正在构建的 step 并添加到 steps.
func (c *StepCollector) FinalizeCurrentStep() {
	if c.currentStep != nil && (c.currentStep.Thought != "" || len(c.currentStep.ToolCalls) > 0) {
		c.steps = append(c.steps, *c.currentStep)
		c.currentStep = nil
	}
}

// GetSteps 获取收集的所有 steps.
func (c *StepCollector) GetSteps() []v1.AgentStep {
	return c.steps
}

// Reset 重置收集器.
func (c *StepCollector) Reset() {
	c.steps = make([]v1.AgentStep, 0)
	c.currentStep = nil
	c.iteration = 0
}
