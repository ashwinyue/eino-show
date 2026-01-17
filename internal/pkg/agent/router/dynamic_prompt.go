// Package router provides dynamic prompt builder.
// Reference: AssistantAgent Prompt Builder
package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// PromptCondition 条件函数类型.
type PromptCondition func(ctx context.Context, evalResult *EvaluationResult) bool

// PromptFragment Prompt 片段.
type PromptFragment struct {
	// Name 片段名称
	Name string

	// Condition 条件函数 (返回 true 时注入)
	Condition PromptCondition

	// Content 静态内容
	Content string

	// Template 模板内容 (支持变量替换)
	Template string

	// Priority 优先级 (数字越小越靠前)
	Priority int

	// Role 消息角色 (默认 system)
	Role schema.RoleType
}

// EvaluationResult 评估结果 (来自意图路由).
type EvaluationResult struct {
	Intent       IntentType
	IsAmbiguous  bool
	HasTools     bool
	HasKnowledge bool
	Experiences  []Experience
	Knowledge    []string
	Tools        []string
	Variables    map[string]any
}

// DynamicPromptBuilder 动态 Prompt 构建器.
type DynamicPromptBuilder struct {
	basePrompt string
	fragments  []*PromptFragment
}

// NewDynamicPromptBuilder 创建动态 Prompt 构建器.
func NewDynamicPromptBuilder(basePrompt string) *DynamicPromptBuilder {
	return &DynamicPromptBuilder{
		basePrompt: basePrompt,
		fragments:  make([]*PromptFragment, 0),
	}
}

// AddFragment 添加条件片段.
func (b *DynamicPromptBuilder) AddFragment(fragment *PromptFragment) *DynamicPromptBuilder {
	if fragment.Role == "" {
		fragment.Role = schema.System
	}
	b.fragments = append(b.fragments, fragment)
	return b
}

// WithAmbiguousPrompt 添加模糊意图处理提示.
func (b *DynamicPromptBuilder) WithAmbiguousPrompt(content string) *DynamicPromptBuilder {
	return b.AddFragment(&PromptFragment{
		Name:      "ambiguous",
		Condition: func(ctx context.Context, r *EvaluationResult) bool { return r.IsAmbiguous },
		Content:   content,
		Priority:  10,
	})
}

// WithToolsPrompt 添加工具使用提示.
func (b *DynamicPromptBuilder) WithToolsPrompt(template string) *DynamicPromptBuilder {
	return b.AddFragment(&PromptFragment{
		Name:      "tools",
		Condition: func(ctx context.Context, r *EvaluationResult) bool { return r.HasTools },
		Template:  template,
		Priority:  20,
	})
}

// WithKnowledgePrompt 添加知识引用提示.
func (b *DynamicPromptBuilder) WithKnowledgePrompt(template string) *DynamicPromptBuilder {
	return b.AddFragment(&PromptFragment{
		Name:      "knowledge",
		Condition: func(ctx context.Context, r *EvaluationResult) bool { return r.HasKnowledge },
		Template:  template,
		Priority:  30,
	})
}

// WithExperiencePrompt 添加经验引用提示.
func (b *DynamicPromptBuilder) WithExperiencePrompt(template string) *DynamicPromptBuilder {
	return b.AddFragment(&PromptFragment{
		Name:      "experience",
		Condition: func(ctx context.Context, r *EvaluationResult) bool { return len(r.Experiences) > 0 },
		Template:  template,
		Priority:  40,
	})
}

// Build 使用 Eino ChatTemplate 构建最终 Prompt.
func (b *DynamicPromptBuilder) Build(ctx context.Context, evalResult *EvaluationResult) ([]*schema.Message, error) {
	// 构建模板变量
	variables := make(map[string]any)
	if evalResult.Variables != nil {
		for k, v := range evalResult.Variables {
			variables[k] = v
		}
	}

	// 添加评估结果作为变量
	variables["is_ambiguous"] = evalResult.IsAmbiguous
	variables["has_tools"] = evalResult.HasTools
	variables["has_knowledge"] = evalResult.HasKnowledge
	variables["tools"] = strings.Join(evalResult.Tools, ", ")
	variables["knowledge"] = strings.Join(evalResult.Knowledge, "\n")

	// 构建经验字符串
	var expStrs []string
	for _, exp := range evalResult.Experiences {
		expStrs = append(expStrs, fmt.Sprintf("- %s: %s", exp.Query, exp.Response))
	}
	variables["experiences"] = strings.Join(expStrs, "\n")

	// 使用 Eino 模板格式化
	templates := []schema.MessagesTemplate{
		schema.SystemMessage(b.basePrompt),
	}

	// 添加条件片段
	sortedFragments := b.sortFragments()
	for _, fragment := range sortedFragments {
		if fragment.Condition != nil && !fragment.Condition(ctx, evalResult) {
			continue
		}

		content := fragment.Content
		if fragment.Template != "" {
			content = fragment.Template
		}

		if content != "" {
			templates = append(templates, &schema.Message{
				Role:    fragment.Role,
				Content: content,
			})
		}
	}

	tpl := prompt.FromMessages(schema.FString, templates...)
	return tpl.Format(ctx, variables)
}

// sortFragments 按优先级排序.
func (b *DynamicPromptBuilder) sortFragments() []*PromptFragment {
	sorted := make([]*PromptFragment, len(b.fragments))
	copy(sorted, b.fragments)

	// 简单冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority > sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// 预定义的 Prompt 片段

// DefaultAmbiguousPrompt 默认模糊意图提示.
const DefaultAmbiguousPrompt = `The user's query seems ambiguous. Please ask clarifying questions before proceeding.
Be specific about what information you need to better understand the user's intent.`

// DefaultToolsPrompt 默认工具使用提示.
const DefaultToolsPrompt = `You have access to the following tools: {tools}

Use tools when necessary to complete the user's request. Always explain what you're doing before calling a tool.`

// DefaultKnowledgePrompt 默认知识引用提示.
const DefaultKnowledgePrompt = `Here is relevant knowledge from the knowledge base:

{knowledge}

Use this information to answer the user's question. Cite sources when appropriate.`

// DefaultExperiencePrompt 默认经验引用提示.
const DefaultExperiencePrompt = `Here are some relevant past experiences that may help:

{experiences}

Consider these experiences when formulating your response.`

// NewDefaultDynamicPromptBuilder 创建默认配置的构建器.
func NewDefaultDynamicPromptBuilder(basePrompt string) *DynamicPromptBuilder {
	return NewDynamicPromptBuilder(basePrompt).
		WithAmbiguousPrompt(DefaultAmbiguousPrompt).
		WithToolsPrompt(DefaultToolsPrompt).
		WithKnowledgePrompt(DefaultKnowledgePrompt).
		WithExperiencePrompt(DefaultExperiencePrompt)
}
