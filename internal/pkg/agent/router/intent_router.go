// Package router provides intent routing using Eino compose.Graph.
// Reference: AssistantAgent Evaluation Graph
package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// IntentType 意图类型.
type IntentType string

const (
	IntentChat     IntentType = "chat"     // 简单对话
	IntentTool     IntentType = "tool"     // 需要工具调用
	IntentRAG      IntentType = "rag"      // 知识检索
	IntentFastPath IntentType = "fastpath" // 快速意图 (规则匹配)
	IntentUnknown  IntentType = "unknown"  // 未知意图
)

// IntentInput 意图路由输入.
type IntentInput struct {
	Query     string            // 用户查询
	SessionID string            // 会话 ID
	History   []*schema.Message // 历史消息
	Metadata  map[string]any    // 额外元数据
}

// IntentOutput 意图路由输出.
type IntentOutput struct {
	Intent         IntentType     // 识别的意图
	RewrittenQuery string         // 改写后的查询
	Confidence     float64        // 置信度
	Knowledge      []string       // 检索到的知识
	MatchedTools   []string       // 匹配的工具
	FastResponse   string         // 快速响应 (FastPath)
	Metadata       map[string]any // 传递的元数据
}

// ClassifyResult 意图分类结果.
type ClassifyResult struct {
	Intent     IntentType
	Confidence float64
	Reason     string
}

// RouterConfig 路由器配置.
type RouterConfig struct {
	// ChatModel 用于意图分类的模型
	ChatModel model.ChatModel

	// FastIntentRules 快速意图规则
	FastIntentRules []FastIntentRule

	// KnowledgeSearcher 知识检索函数 (可选)
	KnowledgeSearcher func(ctx context.Context, query string) ([]string, error)

	// ToolMatcher 工具匹配函数 (可选)
	ToolMatcher func(ctx context.Context, query string) ([]string, error)

	// ClassifyPrompt 自定义分类提示词 (可选)
	ClassifyPrompt string
}

// FastIntentRule 快速意图规则.
type FastIntentRule struct {
	// Patterns 匹配模式 (前缀匹配)
	Patterns []string

	// Response 直接响应
	Response string

	// Intent 对应意图
	Intent IntentType
}

// IntentRouter 意图路由器.
type IntentRouter struct {
	cfg      *RouterConfig
	compiled compose.Runnable[*IntentInput, *IntentOutput]
}

// NewIntentRouter 创建意图路由器.
func NewIntentRouter(ctx context.Context, cfg *RouterConfig) (*IntentRouter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}

	router := &IntentRouter{cfg: cfg}

	// 构建图
	compiled, err := router.buildGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("build graph failed: %w", err)
	}

	router.compiled = compiled
	return router, nil
}

// Route 执行意图路由.
func (r *IntentRouter) Route(ctx context.Context, input *IntentInput) (*IntentOutput, error) {
	return r.compiled.Invoke(ctx, input)
}

// buildGraph 构建意图路由图.
func (r *IntentRouter) buildGraph(ctx context.Context) (compose.Runnable[*IntentInput, *IntentOutput], error) {
	g := compose.NewGraph[*IntentInput, *IntentOutput]()

	// 节点 1: 快速意图检查
	err := g.AddLambdaNode("fast_check", compose.InvokableLambda(r.fastIntentCheck))
	if err != nil {
		return nil, err
	}

	// 节点 2: 查询改写
	err = g.AddLambdaNode("rewrite", compose.InvokableLambda(r.rewriteQuery))
	if err != nil {
		return nil, err
	}

	// 节点 3: 意图分类
	err = g.AddLambdaNode("classify", compose.InvokableLambda(r.classifyIntent))
	if err != nil {
		return nil, err
	}

	// 节点 4: 知识检索
	err = g.AddLambdaNode("search_knowledge", compose.InvokableLambda(r.searchKnowledge))
	if err != nil {
		return nil, err
	}

	// 节点 5: 工具匹配
	err = g.AddLambdaNode("match_tools", compose.InvokableLambda(r.matchTools))
	if err != nil {
		return nil, err
	}

	// 节点 6: 聚合结果
	err = g.AddLambdaNode("aggregate", compose.InvokableLambda(r.aggregateResults))
	if err != nil {
		return nil, err
	}

	// 边: START -> fast_check
	err = g.AddEdge(compose.START, "fast_check")
	if err != nil {
		return nil, err
	}

	// 分支: fast_check -> (fast_path_end | rewrite)
	err = g.AddBranch("fast_check", compose.NewGraphBranch(
		func(ctx context.Context, result *fastCheckResult) (string, error) {
			if result.Matched {
				return "aggregate", nil // 快速响应直接到聚合
			}
			return "rewrite", nil
		},
		map[string]bool{"aggregate": true, "rewrite": true},
	))
	if err != nil {
		return nil, err
	}

	// 边: rewrite -> classify (并行: search_knowledge, match_tools)
	err = g.AddEdge("rewrite", "classify")
	if err != nil {
		return nil, err
	}

	err = g.AddEdge("rewrite", "search_knowledge")
	if err != nil {
		return nil, err
	}

	err = g.AddEdge("rewrite", "match_tools")
	if err != nil {
		return nil, err
	}

	// 边: 所有并行节点 -> aggregate
	err = g.AddEdge("classify", "aggregate")
	if err != nil {
		return nil, err
	}

	err = g.AddEdge("search_knowledge", "aggregate")
	if err != nil {
		return nil, err
	}

	err = g.AddEdge("match_tools", "aggregate")
	if err != nil {
		return nil, err
	}

	// 边: aggregate -> END
	err = g.AddEdge("aggregate", compose.END)
	if err != nil {
		return nil, err
	}

	return g.Compile(ctx)
}

// fastCheckResult 快速检查结果.
type fastCheckResult struct {
	Matched  bool
	Response string
	Intent   IntentType
	Input    *IntentInput
}

// fastIntentCheck 快速意图检查.
func (r *IntentRouter) fastIntentCheck(ctx context.Context, input *IntentInput) (*fastCheckResult, error) {
	for _, rule := range r.cfg.FastIntentRules {
		for _, pattern := range rule.Patterns {
			if strings.HasPrefix(input.Query, pattern) {
				return &fastCheckResult{
					Matched:  true,
					Response: rule.Response,
					Intent:   rule.Intent,
					Input:    input,
				}, nil
			}
		}
	}

	return &fastCheckResult{
		Matched: false,
		Input:   input,
	}, nil
}

// rewriteResult 改写结果.
type rewriteResult struct {
	Original  string
	Rewritten string
	Input     *IntentInput
}

// rewriteQuery 查询改写.
func (r *IntentRouter) rewriteQuery(ctx context.Context, result *fastCheckResult) (*rewriteResult, error) {
	// 简单实现：去除多余空格，规范化
	rewritten := strings.TrimSpace(result.Input.Query)
	rewritten = strings.Join(strings.Fields(rewritten), " ")

	return &rewriteResult{
		Original:  result.Input.Query,
		Rewritten: rewritten,
		Input:     result.Input,
	}, nil
}

// classifyResult 分类结果.
type classifyResult struct {
	Intent     IntentType
	Confidence float64
	Rewritten  string
	Input      *IntentInput
}

// classifyIntent 意图分类.
func (r *IntentRouter) classifyIntent(ctx context.Context, result *rewriteResult) (*classifyResult, error) {
	prompt := r.cfg.ClassifyPrompt
	if prompt == "" {
		prompt = defaultClassifyPrompt
	}

	messages := []*schema.Message{
		schema.SystemMessage(prompt),
		schema.UserMessage(fmt.Sprintf("Query: %s", result.Rewritten)),
	}

	resp, err := r.cfg.ChatModel.Generate(ctx, messages)
	if err != nil {
		// 分类失败，默认为 chat
		return &classifyResult{
			Intent:     IntentChat,
			Confidence: 0.5,
			Rewritten:  result.Rewritten,
			Input:      result.Input,
		}, nil
	}

	intent, confidence := parseClassifyResponse(resp.Content)

	return &classifyResult{
		Intent:     intent,
		Confidence: confidence,
		Rewritten:  result.Rewritten,
		Input:      result.Input,
	}, nil
}

// knowledgeResult 知识检索结果.
type knowledgeResult struct {
	Knowledge []string
	Input     *IntentInput
}

// searchKnowledge 知识检索.
func (r *IntentRouter) searchKnowledge(ctx context.Context, result *rewriteResult) (*knowledgeResult, error) {
	if r.cfg.KnowledgeSearcher == nil {
		return &knowledgeResult{Input: result.Input}, nil
	}

	knowledge, err := r.cfg.KnowledgeSearcher(ctx, result.Rewritten)
	if err != nil {
		return &knowledgeResult{Input: result.Input}, nil
	}

	return &knowledgeResult{
		Knowledge: knowledge,
		Input:     result.Input,
	}, nil
}

// toolsResult 工具匹配结果.
type toolsResult struct {
	Tools []string
	Input *IntentInput
}

// matchTools 工具匹配.
func (r *IntentRouter) matchTools(ctx context.Context, result *rewriteResult) (*toolsResult, error) {
	if r.cfg.ToolMatcher == nil {
		return &toolsResult{Input: result.Input}, nil
	}

	tools, err := r.cfg.ToolMatcher(ctx, result.Rewritten)
	if err != nil {
		return &toolsResult{Input: result.Input}, nil
	}

	return &toolsResult{
		Tools: tools,
		Input: result.Input,
	}, nil
}

// aggregateInput 聚合输入 (来自多个并行节点).
type aggregateInput struct {
	FastCheck *fastCheckResult
	Classify  *classifyResult
	Knowledge *knowledgeResult
	Tools     *toolsResult
}

// aggregateResults 聚合结果.
func (r *IntentRouter) aggregateResults(ctx context.Context, input any) (*IntentOutput, error) {
	output := &IntentOutput{
		Intent:   IntentUnknown,
		Metadata: make(map[string]any),
	}

	// 根据输入类型处理
	switch v := input.(type) {
	case *fastCheckResult:
		if v.Matched {
			output.Intent = v.Intent
			output.FastResponse = v.Response
			output.Confidence = 1.0
		}

	case *classifyResult:
		output.Intent = v.Intent
		output.Confidence = v.Confidence
		output.RewrittenQuery = v.Rewritten

	case *knowledgeResult:
		output.Knowledge = v.Knowledge

	case *toolsResult:
		output.MatchedTools = v.Tools
	}

	return output, nil
}

// parseClassifyResponse 解析分类响应.
func parseClassifyResponse(response string) (IntentType, float64) {
	response = strings.ToLower(strings.TrimSpace(response))

	switch {
	case strings.Contains(response, "tool"):
		return IntentTool, 0.9
	case strings.Contains(response, "rag"), strings.Contains(response, "search"), strings.Contains(response, "knowledge"):
		return IntentRAG, 0.9
	case strings.Contains(response, "chat"), strings.Contains(response, "conversation"):
		return IntentChat, 0.9
	default:
		return IntentChat, 0.6
	}
}

const defaultClassifyPrompt = `You are an intent classifier. Analyze the user query and classify it into one of the following intents:

1. "chat" - Simple conversation, greeting, or general questions that don't require external tools or knowledge search
2. "tool" - Queries that require calling external tools or APIs (e.g., weather, calculations, web search)
3. "rag" - Queries that require searching knowledge base or documents

Respond with ONLY the intent name (chat, tool, or rag).`
