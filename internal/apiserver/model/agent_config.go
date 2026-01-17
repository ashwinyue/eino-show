// Package model 提供 Agent 配置类型定义（对齐 WeKnora）.
package model

import (
	"database/sql/driver"
	"encoding/json"
)

// BuiltinAgentID 内置 Agent ID 常量
const (
	// BuiltinQuickAnswerID 快速问答 (RAG) Agent
	BuiltinQuickAnswerID = "builtin-quick-answer"
	// BuiltinSmartReasoningID 智能推理 (ReAct) Agent
	BuiltinSmartReasoningID = "builtin-smart-reasoning"
	// BuiltinDataAnalystID 数据分析师 Agent
	BuiltinDataAnalystID = "builtin-data-analyst"
)

// AgentMode Agent 运行模式
const (
	// AgentModeQuickAnswer RAG 快速问答模式
	AgentModeQuickAnswer = "quick-answer"
	// AgentModeSmartReasoning ReAct 智能推理模式
	AgentModeSmartReasoning = "smart-reasoning"
)

// CustomAgentConfig 自定义 Agent 配置（对齐 WeKnora）
type CustomAgentConfig struct {
	// ===== 基础设置 =====
	// AgentMode: "quick-answer" for RAG mode, "smart-reasoning" for ReAct agent mode
	AgentMode string `json:"agent_mode"`
	// SystemPrompt 系统提示词
	SystemPrompt string `json:"system_prompt"`
	// ContextTemplate 上下文模板（普通模式）
	ContextTemplate string `json:"context_template"`

	// ===== 模型设置 =====
	// ModelID 对话模型 ID
	ModelID string `json:"model_id"`
	// RerankModelID 重排模型 ID
	RerankModelID string `json:"rerank_model_id"`
	// Temperature LLM 温度 (0-1)
	Temperature float64 `json:"temperature"`
	// MaxCompletionTokens 最大生成 token 数
	MaxCompletionTokens int `json:"max_completion_tokens"`

	// ===== Agent 模式设置 =====
	// MaxIterations ReAct 循环最大迭代次数
	MaxIterations int `json:"max_iterations"`
	// AllowedTools 允许的工具列表
	AllowedTools []string `json:"allowed_tools"`
	// ReflectionEnabled 是否启用反思
	ReflectionEnabled bool `json:"reflection_enabled"`
	// MCPSelectionMode MCP 服务选择模式: "all", "selected", "none"
	MCPSelectionMode string `json:"mcp_selection_mode"`
	// MCPServices 选中的 MCP 服务 ID 列表
	MCPServices []string `json:"mcp_services"`

	// ===== 知识库设置 =====
	// KBSelectionMode 知识库选择模式: "all", "selected", "none"
	KBSelectionMode string `json:"kb_selection_mode"`
	// KnowledgeBases 关联的知识库 ID 列表
	KnowledgeBases []string `json:"knowledge_bases"`

	// ===== 文件类型限制 =====
	// SupportedFileTypes 支持的文件类型
	SupportedFileTypes []string `json:"supported_file_types"`

	// ===== FAQ 策略 =====
	// FAQPriorityEnabled 是否启用 FAQ 优先策略
	FAQPriorityEnabled bool `json:"faq_priority_enabled"`
	// FAQDirectAnswerThreshold FAQ 直接回答阈值
	FAQDirectAnswerThreshold float64 `json:"faq_direct_answer_threshold"`
	// FAQScoreBoost FAQ 分数提升倍数
	FAQScoreBoost float64 `json:"faq_score_boost"`

	// ===== 网络搜索设置 =====
	// WebSearchEnabled 是否启用网络搜索
	WebSearchEnabled bool `json:"web_search_enabled"`
	// WebSearchMaxResults 网络搜索最大结果数
	WebSearchMaxResults int `json:"web_search_max_results"`

	// ===== 多轮对话设置 =====
	// MultiTurnEnabled 是否启用多轮对话
	MultiTurnEnabled bool `json:"multi_turn_enabled"`
	// HistoryTurns 历史对话轮数
	HistoryTurns int `json:"history_turns"`

	// ===== 检索策略设置 =====
	// EmbeddingTopK 向量检索 TopK
	EmbeddingTopK int `json:"embedding_top_k"`
	// KeywordThreshold 关键词检索阈值
	KeywordThreshold float64 `json:"keyword_threshold"`
	// VectorThreshold 向量检索阈值
	VectorThreshold float64 `json:"vector_threshold"`
	// RerankTopK 重排 TopK
	RerankTopK int `json:"rerank_top_k"`
	// RerankThreshold 重排阈值
	RerankThreshold float64 `json:"rerank_threshold"`

	// ===== 高级设置 =====
	// EnableQueryExpansion 是否启用查询扩展
	EnableQueryExpansion bool `json:"enable_query_expansion"`
	// EnableRewrite 是否启用查询重写
	EnableRewrite bool `json:"enable_rewrite"`
	// RewritePromptSystem 重写系统提示词
	RewritePromptSystem string `json:"rewrite_prompt_system"`
	// RewritePromptUser 重写用户提示词模板
	RewritePromptUser string `json:"rewrite_prompt_user"`
	// FallbackStrategy 兜底策略: "fixed" or "model"
	FallbackStrategy string `json:"fallback_strategy"`
	// FallbackResponse 固定兜底回复
	FallbackResponse string `json:"fallback_response"`
	// FallbackPrompt 兜底提示词
	FallbackPrompt string `json:"fallback_prompt"`

	// ===== 多 Agent 协作 =====
	// SubAgents 子 Agent ID 列表
	SubAgents []string `json:"sub_agents"`
}

// Value 实现 driver.Valuer 接口
func (c CustomAgentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *CustomAgentConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// EnsureDefaults 设置默认值
func (c *CustomAgentConfig) EnsureDefaults() {
	if c.Temperature == 0 {
		c.Temperature = 0.7
	}
	if c.MaxIterations == 0 {
		c.MaxIterations = 10
	}
	if c.WebSearchMaxResults == 0 {
		c.WebSearchMaxResults = 5
	}
	if c.HistoryTurns == 0 {
		c.HistoryTurns = 5
	}
	if c.EmbeddingTopK == 0 {
		c.EmbeddingTopK = 10
	}
	if c.KeywordThreshold == 0 {
		c.KeywordThreshold = 0.3
	}
	if c.VectorThreshold == 0 {
		c.VectorThreshold = 0.5
	}
	if c.RerankTopK == 0 {
		c.RerankTopK = 5
	}
	if c.RerankThreshold == 0 {
		c.RerankThreshold = 0.5
	}
	if c.FallbackStrategy == "" {
		c.FallbackStrategy = "model"
	}
	if c.MaxCompletionTokens == 0 {
		c.MaxCompletionTokens = 2048
	}
	// Agent 模式强制启用多轮对话
	if c.AgentMode == AgentModeSmartReasoning {
		c.MultiTurnEnabled = true
	}
}

// IsAgentMode 是否为 Agent 模式
func (c *CustomAgentConfig) IsAgentMode() bool {
	return c.AgentMode == AgentModeSmartReasoning
}

// GetBuiltinQuickAnswerConfig 获取内置快速问答 Agent 配置
func GetBuiltinQuickAnswerConfig() CustomAgentConfig {
	return CustomAgentConfig{
		AgentMode:    AgentModeQuickAnswer,
		SystemPrompt: "",
		ContextTemplate: `请根据以下参考资料回答用户问题。

参考资料：
{{contexts}}

用户问题：{{query}}`,
		Temperature:              0.7,
		MaxCompletionTokens:      2048,
		WebSearchEnabled:         true,
		WebSearchMaxResults:      5,
		MultiTurnEnabled:         true,
		HistoryTurns:             5,
		KBSelectionMode:          "all",
		FAQPriorityEnabled:       true,
		FAQDirectAnswerThreshold: 0.9,
		FAQScoreBoost:            1.2,
		EmbeddingTopK:            10,
		KeywordThreshold:         0.3,
		VectorThreshold:          0.5,
		RerankTopK:               10,
		RerankThreshold:          0.3,
		EnableQueryExpansion:     true,
		EnableRewrite:            true,
		FallbackStrategy:         "model",
	}
}

// GetBuiltinSmartReasoningConfig 获取内置智能推理 Agent 配置
func GetBuiltinSmartReasoningConfig() CustomAgentConfig {
	return CustomAgentConfig{
		AgentMode:           AgentModeSmartReasoning,
		SystemPrompt:        "",
		Temperature:         0.7,
		MaxCompletionTokens: 2048,
		MaxIterations:       50,
		KBSelectionMode:     "all",
		AllowedTools: []string{
			"thinking", "todo_write", "knowledge_search",
			"grep_chunks", "list_knowledge_chunks",
			"query_knowledge_graph", "get_document_info",
		},
		WebSearchEnabled:         true,
		WebSearchMaxResults:      5,
		ReflectionEnabled:        false,
		MultiTurnEnabled:         true,
		HistoryTurns:             5,
		FAQPriorityEnabled:       true,
		FAQDirectAnswerThreshold: 0.9,
		FAQScoreBoost:            1.2,
		EmbeddingTopK:            10,
		KeywordThreshold:         0.3,
		VectorThreshold:          0.5,
		RerankTopK:               10,
		RerankThreshold:          0.3,
	}
}

// GetBuiltinDataAnalystConfig 获取内置数据分析师 Agent 配置
func GetBuiltinDataAnalystConfig() CustomAgentConfig {
	return CustomAgentConfig{
		AgentMode: AgentModeSmartReasoning,
		SystemPrompt: `### Role
You are Data Analyst, an intelligent data analysis assistant powered by DuckDB. You specialize in analyzing structured data from CSV and Excel files using SQL queries.

### Mission
Help users explore, analyze, and derive insights from their tabular data through intelligent SQL query generation and execution.

### Critical Constraints
1. **Schema First:** ALWAYS call data_schema before writing any SQL query to understand the table structure.
2. **Read-Only:** Only SELECT queries allowed. INSERT, UPDATE, DELETE, CREATE, DROP are forbidden.
3. **Iterative Refinement:** If a query fails, analyze the error and refine your approach.

### Workflow
1. **Understand:** Call data_schema to get table name, columns, types, and row count.
2. **Plan:** For complex questions, use todo_write to break into sub-queries.
3. **Query:** Call data_analysis with the knowledge_id and SQL query.
4. **Analyze:** Interpret results and provide insights.

Current Time: {{current_time}}
`,
		Temperature:         0.3,
		MaxCompletionTokens: 4096,
		MaxIterations:       30,
		KBSelectionMode:     "all",
		SupportedFileTypes:  []string{"csv", "xlsx"},
		AllowedTools: []string{
			"thinking", "todo_write", "data_schema", "data_analysis",
		},
		WebSearchEnabled:    false,
		WebSearchMaxResults: 0,
		ReflectionEnabled:   true,
		MultiTurnEnabled:    true,
		HistoryTurns:        10,
		EmbeddingTopK:       5,
		KeywordThreshold:    0.3,
		VectorThreshold:     0.5,
		RerankTopK:          5,
		RerankThreshold:     0.3,
	}
}

// BuiltinAgentRegistry 内置 Agent 注册表
var BuiltinAgentRegistry = map[string]func() CustomAgentConfig{
	BuiltinQuickAnswerID:    GetBuiltinQuickAnswerConfig,
	BuiltinSmartReasoningID: GetBuiltinSmartReasoningConfig,
	BuiltinDataAnalystID:    GetBuiltinDataAnalystConfig,
}

// BuiltinAgentInfo 内置 Agent 信息
type BuiltinAgentInfo struct {
	ID          string
	Name        string
	Description string
	Avatar      string
}

// GetBuiltinAgentInfos 获取所有内置 Agent 信息
func GetBuiltinAgentInfos() []BuiltinAgentInfo {
	return []BuiltinAgentInfo{
		{
			ID:          BuiltinQuickAnswerID,
			Name:        "快速问答",
			Description: "基于知识库的 RAG 问答，快速准确地回答问题",
			Avatar:      "💬",
		},
		{
			ID:          BuiltinSmartReasoningID,
			Name:        "智能推理",
			Description: "ReAct 推理框架，支持多步思考和工具调用",
			Avatar:      "🧠",
		},
		{
			ID:          BuiltinDataAnalystID,
			Name:        "数据分析师",
			Description: "专业数据分析智能体，支持 CSV/Excel 文件的 SQL 查询与统计分析",
			Avatar:      "📊",
		},
	}
}

// IsBuiltinAgentID 检查是否为内置 Agent ID
func IsBuiltinAgentID(id string) bool {
	_, exists := BuiltinAgentRegistry[id]
	return exists
}

// GetBuiltinAgentConfig 根据 ID 获取内置 Agent 配置
func GetBuiltinAgentConfig(id string) (CustomAgentConfig, bool) {
	if factory, exists := BuiltinAgentRegistry[id]; exists {
		return factory(), true
	}
	return CustomAgentConfig{}, false
}
