// Package enhanced provides an enhanced agent with intent routing, dynamic prompt, and experience management.
// Reference: Eino flow/agent/react/react.go best practices
package enhanced

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/ashwinyue/eino-show/internal/pkg/agent/router"
)

// State 增强 Agent 的状态.
type State struct {
	Messages   []*schema.Message
	Intent     router.IntentType
	Experience *router.Experience
	Prompt     string
}

func init() {
	schema.RegisterName[*State]("_eino_enhanced_agent_state")
}

// AgentConfig 增强 Agent 配置.
type AgentConfig struct {
	// ChatModel 对话模型
	ChatModel model.ChatModel

	// ToolCallingModel 工具调用模型
	ToolCallingModel model.ToolCallingChatModel

	// ToolsConfig 工具配置
	ToolsConfig compose.ToolsNodeConfig

	// IntentRouter 意图路由器 (可选)
	IntentRouter *router.IntentRouter

	// DynamicPromptBuilder 动态 Prompt 构建器 (可选)
	DynamicPromptBuilder *router.DynamicPromptBuilder

	// ExperienceManager 经验管理器 (可选)
	ExperienceManager *router.ExperienceManager

	// SystemPrompt 系统提示 (如果没有 DynamicPromptBuilder)
	SystemPrompt string

	// MaxStep 最大步数
	MaxStep int

	// GraphName 图名称
	GraphName string
}

// Node keys
const (
	nodeKeyIntentRouter  = "intent_router"
	nodeKeyExperience    = "experience"
	nodeKeyPromptBuilder = "prompt_builder"
	nodeKeyModel         = "chat_model"
	nodeKeyTools         = "tools"
	nodeKeyFastIntent    = "fast_intent"
)

// Agent 增强 Agent.
type Agent struct {
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
	graph    *compose.Graph[[]*schema.Message, *schema.Message]
	config   *AgentConfig
}

// NewAgent 创建增强 Agent.
func NewAgent(ctx context.Context, config *AgentConfig) (*Agent, error) {
	if config.MaxStep <= 0 {
		config.MaxStep = 12
	}

	graphName := config.GraphName
	if graphName == "" {
		graphName = "EnhancedAgent"
	}

	// 创建 Graph
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *State {
			return &State{Messages: make([]*schema.Message, 0, config.MaxStep+1)}
		}),
	)

	// 添加节点
	if err := addNodes(ctx, graph, config); err != nil {
		return nil, fmt.Errorf("failed to add nodes: %w", err)
	}

	// 添加边和分支
	if err := addEdgesAndBranches(graph, config); err != nil {
		return nil, fmt.Errorf("failed to add edges: %w", err)
	}

	// 编译
	compileOpts := []compose.GraphCompileOption{
		compose.WithMaxRunSteps(config.MaxStep),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
		compose.WithGraphName(graphName),
	}

	runnable, err := graph.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	return &Agent{
		runnable: runnable,
		graph:    graph,
		config:   config,
	}, nil
}

// addNodes 添加节点.
func addNodes(ctx context.Context, graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	if err := addIntentRouterNode(graph, config); err != nil {
		return err
	}
	if err := addExperienceNode(graph, config); err != nil {
		return err
	}
	if err := addPromptBuilderNode(graph, config); err != nil {
		return err
	}
	if err := addFastIntentNode(graph, config); err != nil {
		return err
	}
	if err := addModelNode(graph, config); err != nil {
		return err
	}
	if err := addToolsNode(ctx, graph, config); err != nil {
		return err
	}
	return nil
}

// extractUserQuery 从消息列表中提取用户查询.
func extractUserQuery(messages []*schema.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == schema.User {
			return messages[i].Content
		}
	}
	return ""
}

// addIntentRouterNode 添加意图路由节点.
func addIntentRouterNode(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	if config.IntentRouter == nil {
		return nil
	}

	lambda := func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		query := extractUserQuery(input)
		routerInput := &router.IntentInput{Query: query, History: input}

		result, err := config.IntentRouter.Route(ctx, routerInput)
		if err != nil {
			return input, nil
		}

		_ = compose.ProcessState(ctx, func(ctx context.Context, state *State) error {
			state.Intent = result.Intent
			state.Messages = input
			return nil
		})
		return input, nil
	}

	return graph.AddLambdaNode(nodeKeyIntentRouter, compose.InvokableLambda(lambda))
}

// addExperienceNode 添加经验检索节点.
func addExperienceNode(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	if config.ExperienceManager == nil {
		return nil
	}

	lambda := func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		query := extractUserQuery(input)
		experiences, err := config.ExperienceManager.Recall(ctx, query)
		if err == nil && len(experiences) > 0 {
			_ = compose.ProcessState(ctx, func(ctx context.Context, state *State) error {
				state.Experience = experiences[0]
				return nil
			})
		}
		return input, nil
	}

	return graph.AddLambdaNode(nodeKeyExperience, compose.InvokableLambda(lambda))
}

// addPromptBuilderNode 添加动态 Prompt 构建节点.
func addPromptBuilderNode(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	if config.DynamicPromptBuilder == nil {
		return nil
	}

	lambda := func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		var evalResult *router.EvaluationResult

		_ = compose.ProcessState(ctx, func(ctx context.Context, state *State) error {
			evalResult = &router.EvaluationResult{Intent: state.Intent}
			if state.Experience != nil {
				evalResult.Experiences = []router.Experience{*state.Experience}
			}
			return nil
		})

		promptMessages, err := config.DynamicPromptBuilder.Build(ctx, evalResult)
		if err != nil && config.SystemPrompt != "" {
			promptMessages = []*schema.Message{schema.SystemMessage(config.SystemPrompt)}
		}

		if len(promptMessages) > 0 {
			result := make([]*schema.Message, 0, len(promptMessages)+len(input))
			return append(append(result, promptMessages...), input...), nil
		}
		return input, nil
	}

	return graph.AddLambdaNode(nodeKeyPromptBuilder, compose.InvokableLambda(lambda))
}

// addFastIntentNode 添加快速意图响应节点.
func addFastIntentNode(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	if config.ExperienceManager == nil {
		return nil
	}

	lambda := func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
		var exp *router.Experience
		_ = compose.ProcessState(ctx, func(ctx context.Context, state *State) error {
			exp = state.Experience
			return nil
		})

		if exp != nil && exp.Type == router.ExperienceTypeFastIntent && exp.FastIntentConfig != nil {
			return schema.AssistantMessage(exp.FastIntentConfig.DirectResponse, nil), nil
		}
		return nil, fmt.Errorf("not a fast intent")
	}

	return graph.AddLambdaNode(nodeKeyFastIntent, compose.InvokableLambda(lambda))
}

// addModelNode 添加模型节点.
func addModelNode(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	modelPreHandle := func(ctx context.Context, input []*schema.Message, state *State) ([]*schema.Message, error) {
		state.Messages = append(state.Messages, input...)
		return state.Messages, nil
	}

	if config.ToolCallingModel != nil {
		return graph.AddChatModelNode(nodeKeyModel, config.ToolCallingModel, compose.WithStatePreHandler(modelPreHandle))
	}
	if config.ChatModel != nil {
		return graph.AddChatModelNode(nodeKeyModel, config.ChatModel, compose.WithStatePreHandler(modelPreHandle))
	}
	return fmt.Errorf("either ChatModel or ToolCallingModel is required")
}

// addToolsNode 添加工具节点.
func addToolsNode(ctx context.Context, graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	if len(config.ToolsConfig.Tools) == 0 {
		return nil
	}

	toolsNode, err := compose.NewToolNode(ctx, &config.ToolsConfig)
	if err != nil {
		return err
	}

	toolsPreHandle := func(ctx context.Context, input *schema.Message, state *State) (*schema.Message, error) {
		state.Messages = append(state.Messages, input)
		return input, nil
	}

	return graph.AddToolsNode(nodeKeyTools, toolsNode, compose.WithStatePreHandler(toolsPreHandle))
}

// addEdgesAndBranches 添加边和分支.
// 流程: START → IntentRouter → Experience → PromptBuilder → Model → [Tools → Model]* → END
func addEdgesAndBranches(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	prevNode := addPreModelEdges(graph, config)
	if err := addExperienceBranch(graph, config, prevNode); err != nil {
		return err
	}
	if err := addPromptToModelEdge(graph, config, prevNode); err != nil {
		return err
	}
	return addModelBranch(graph, config)
}

// addPreModelEdges 添加模型前的边.
func addPreModelEdges(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) string {
	if config.IntentRouter != nil {
		_ = graph.AddEdge(compose.START, nodeKeyIntentRouter)
		return nodeKeyIntentRouter
	}
	return ""
}

// addExperienceBranch 添加经验节点及其分支.
func addExperienceBranch(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig, prevNode string) error {
	if config.ExperienceManager == nil {
		return nil
	}

	// 连接到经验节点
	if prevNode == "" {
		if err := graph.AddEdge(compose.START, nodeKeyExperience); err != nil {
			return err
		}
	} else {
		if err := graph.AddEdge(prevNode, nodeKeyExperience); err != nil {
			return err
		}
	}

	// 快速意图分支条件
	branchCondition := func(ctx context.Context, input []*schema.Message) (string, error) {
		var exp *router.Experience
		_ = compose.ProcessState(ctx, func(ctx context.Context, state *State) error {
			exp = state.Experience
			return nil
		})

		if exp != nil && exp.Type == router.ExperienceTypeFastIntent && exp.FastIntentConfig != nil {
			return nodeKeyFastIntent, nil
		}
		if config.DynamicPromptBuilder != nil {
			return nodeKeyPromptBuilder, nil
		}
		return nodeKeyModel, nil
	}

	endNodes := map[string]bool{nodeKeyFastIntent: true, nodeKeyModel: true}
	if config.DynamicPromptBuilder != nil {
		endNodes[nodeKeyPromptBuilder] = true
	}

	if err := graph.AddBranch(nodeKeyExperience, compose.NewGraphBranch(branchCondition, endNodes)); err != nil {
		return err
	}

	return graph.AddEdge(nodeKeyFastIntent, compose.END)
}

// addPromptToModelEdge 添加 Prompt 到 Model 的边.
func addPromptToModelEdge(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig, prevNode string) error {
	if config.DynamicPromptBuilder != nil {
		if prevNode != "" && prevNode != nodeKeyExperience && config.ExperienceManager == nil {
			if err := graph.AddEdge(prevNode, nodeKeyPromptBuilder); err != nil {
				return err
			}
		}
		return graph.AddEdge(nodeKeyPromptBuilder, nodeKeyModel)
	}

	if prevNode == "" {
		return graph.AddEdge(compose.START, nodeKeyModel)
	}
	if config.ExperienceManager == nil {
		return graph.AddEdge(prevNode, nodeKeyModel)
	}
	return nil
}

// addModelBranch 添加模型后的分支.
func addModelBranch(graph *compose.Graph[[]*schema.Message, *schema.Message], config *AgentConfig) error {
	if len(config.ToolsConfig.Tools) == 0 {
		return graph.AddEdge(nodeKeyModel, compose.END)
	}

	// 有工具调用 → Tools, 无 → END
	branchCondition := func(ctx context.Context, msg *schema.Message) (string, error) {
		if len(msg.ToolCalls) > 0 {
			return nodeKeyTools, nil
		}
		return compose.END, nil
	}

	if err := graph.AddBranch(nodeKeyModel, compose.NewGraphBranch(branchCondition, map[string]bool{
		nodeKeyTools: true,
		compose.END:  true,
	})); err != nil {
		return err
	}

	return graph.AddEdge(nodeKeyTools, nodeKeyModel)
}

// Generate 生成响应.
func (a *Agent) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return a.runnable.Invoke(ctx, messages)
}

// Stream 流式生成.
func (a *Agent) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return a.runnable.Stream(ctx, messages)
}

// GetGraph 获取内部图 (用于调试).
func (a *Agent) GetGraph() *compose.Graph[[]*schema.Message, *schema.Message] {
	return a.graph
}
