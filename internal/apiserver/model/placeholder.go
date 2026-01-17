// Package model 提供占位符系统定义（对齐 WeKnora）.
package model

// PromptField 提示词字段类型
type PromptField string

const (
	PromptFieldSystemPrompt        PromptField = "system_prompt"
	PromptFieldAgentSystemPrompt   PromptField = "agent_system_prompt"
	PromptFieldContextTemplate     PromptField = "context_template"
	PromptFieldRewriteSystemPrompt PromptField = "rewrite_system_prompt"
	PromptFieldRewritePrompt       PromptField = "rewrite_prompt"
	PromptFieldFallbackPrompt      PromptField = "fallback_prompt"
)

// PlaceholderDef 占位符定义
type PlaceholderDef struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Example     string        `json:"example,omitempty"`
	Fields      []PromptField `json:"fields"` // 适用的字段
}

// 所有占位符定义
var allPlaceholders = []PlaceholderDef{
	{
		Name:        "{{query}}",
		Description: "用户的原始问题",
		Example:     "什么是人工智能？",
		Fields:      []PromptField{PromptFieldContextTemplate, PromptFieldRewritePrompt},
	},
	{
		Name:        "{{contexts}}",
		Description: "从知识库检索到的相关内容",
		Example:     "[1] AI是人工智能的缩写...\n[2] 机器学习是AI的子领域...",
		Fields:      []PromptField{PromptFieldContextTemplate},
	},
	{
		Name:        "{{history}}",
		Description: "历史对话记录",
		Example:     "用户: 你好\n助手: 你好，有什么可以帮助你的？",
		Fields:      []PromptField{PromptFieldContextTemplate, PromptFieldRewritePrompt},
	},
	{
		Name:        "{{current_time}}",
		Description: "当前时间",
		Example:     "2024-01-15 14:30:00",
		Fields:      []PromptField{PromptFieldSystemPrompt, PromptFieldAgentSystemPrompt},
	},
	{
		Name:        "{{global_context}}",
		Description: "全局上下文信息",
		Example:     "公司名称: ABC科技\n产品: 智能助手",
		Fields:      []PromptField{PromptFieldSystemPrompt, PromptFieldAgentSystemPrompt, PromptFieldContextTemplate},
	},
	{
		Name:        "{{rewritten_query}}",
		Description: "重写后的查询",
		Example:     "人工智能的定义是什么？",
		Fields:      []PromptField{PromptFieldContextTemplate},
	},
	{
		Name:        "{{knowledge_base_name}}",
		Description: "知识库名称",
		Example:     "产品文档库",
		Fields:      []PromptField{PromptFieldSystemPrompt, PromptFieldAgentSystemPrompt},
	},
	{
		Name:        "{{session_id}}",
		Description: "会话 ID",
		Example:     "sess-abc123",
		Fields:      []PromptField{PromptFieldSystemPrompt, PromptFieldAgentSystemPrompt},
	},
	{
		Name:        "{{user_id}}",
		Description: "用户 ID",
		Example:     "user-xyz789",
		Fields:      []PromptField{PromptFieldSystemPrompt, PromptFieldAgentSystemPrompt},
	},
	{
		Name:        "{{tools}}",
		Description: "可用工具列表",
		Example:     "knowledge_search, web_search, calculator",
		Fields:      []PromptField{PromptFieldAgentSystemPrompt},
	},
}

// AllPlaceholders 获取所有占位符
func AllPlaceholders() []PlaceholderDef {
	return allPlaceholders
}

// PlaceholdersByField 根据字段获取占位符
func PlaceholdersByField(field PromptField) []PlaceholderDef {
	var result []PlaceholderDef
	for _, p := range allPlaceholders {
		for _, f := range p.Fields {
			if f == field {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

// GetPlaceholderNames 获取所有占位符名称
func GetPlaceholderNames() []string {
	names := make([]string, len(allPlaceholders))
	for i, p := range allPlaceholders {
		names[i] = p.Name
	}
	return names
}

// IsValidPlaceholder 检查是否为有效占位符
func IsValidPlaceholder(name string) bool {
	for _, p := range allPlaceholders {
		if p.Name == name {
			return true
		}
	}
	return false
}
