# 基于 Eino ADK 的 WeKnora 重构设计文档

## 文档概述

本文档详细描述使用 CloudWeGo Eino ADK (Agent Development Kit) 重构 WeKnora 系统的设计方案，包括架构映射、改造点和实施路径。

---

## 1. 当前架构分析

### 1.1 核心组件

```
WeKnora 当前架构
├── AgentEngine (ReAct 循环引擎)
│   ├── BuildSystemPrompt - RAG 提示构建
│   ├── Execute - Think-Act-Observe 循环
│   └── streamThinkingToEventBus - 流式事件推送
│
├── ToolRegistry (工具注册表)
│   ├── knowledge_search - 语义搜索
│   ├── grep_chunks - 关键词搜索
│   ├── query_knowledge_graph - 知识图谱
│   ├── web_search - 网络搜索
│   ├── database_query - 数据库查询
│   └── data_analysis - 数据分析
│
├── AgentStreamHandler (流式事件处理器)
│   ├── handleThought - 思考事件
│   ├── handleToolCall - 工具调用事件
│   ├── handleToolResult - 工具结果事件
│   ├── handleReferences - 知识引用
│   └── handleFinalAnswer - 最终答案
│
├── RAG Pipeline (检索增强生成流水线)
│   ├── REWRITE_QUERY - 查询重写
│   ├── CHUNK_SEARCH_PARALLEL - 并行搜索
│   ├── CHUNK_RERANK - 重排序
│   ├── CHUNK_MERGE - 结果合并
│   └── CHAT_COMPLETION_STREAM - 流式完成
│
└── EventBus (自定义事件总线)
    ├── EventAgentThought
    ├── EventAgentToolCall
    ├── EventAgentToolResult
    ├── EventAgentReferences
    └── EventAgentFinalAnswer
```

### 1.2 关键接口

```go
// 当前 AgentEngine
type AgentEngine struct {
    config               *types.AgentConfig
    toolRegistry         *tools.ToolRegistry
    chatModel            chat.Chat
    eventBus             *event.EventBus
    knowledgeBasesInfo   []*KnowledgeBaseInfo
    contextManager       interfaces.ContextManager
    sessionID            string
}

// 当前工具接口
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx, arguments) (*ToolResult, error)
}

// 当前流式输出
func (e *AgentEngine) Execute(ctx context.Context) error {
    for iteration := 0; iteration < maxIterations; iteration++ {
        // Think: 调用 LLM
        response := e.chatModel.Stream(ctx, messages)

        // Act: 执行工具
        for _, toolCall := range response.ToolCalls {
            result := e.toolRegistry.ExecuteTool(ctx, toolCall)
            e.eventBus.Emit(ctx, EventAgentToolCall{...})
        }

        // Observe: 将结果添加到历史
        messages = append(messages, toolResultMessage)
    }
}
```

---

## 2. Eino ADK 架构

### 2.1 核心概念

```
Eino ADK 架构
├── Agent Interface
│   ├── Run(ctx, input) *AsyncIterator[*AgentEvent]
│   └── AgentEvent { Output, Action, Err }
│
├── 内置 Agent 类型
│   ├── ChatModelAgent - LLM 对话
│   ├── SequentialAgent - 顺序执行
│   ├── ParallelAgent - 并行执行
│   ├── PlanExecuteAgent - 计划执行
│   └── ReactAgent - ReAct 循环
│
├── Compose Graph
│   ├── AddNode - 添加节点
│   ├── AddEdge - 添加边
│   ├── Chain - 链式调用
│   └── Branch - 条件分支
│
└── 工具系统
    ├── BaseTool - 工具接口
    ├── InvokableRun - 同步执行
    ├── StreamableRun - 流式执行
    └── AgentTool - Agent 作为工具
```

### 2.2 Agent 定义模式

```go
// 1. ChatModelAgent - 简单对话
agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "qa_agent",
    Description: "Q&A Assistant",
    Instruction: systemPrompt,
    Model:       chatModel,
})

// 2. ReactAgent - 带工具的 Agent
rAgent, _ := react.NewAgent(ctx, &react.AgentConfig{
    ToolCallingModel: model,
    ToolsConfig: compose.ToolsNodeConfig{
        Tools: []tool.BaseTool{
            searchTool,
            databaseTool,
        },
    },
    MaxIterations: 10,
})

// 3. SequentialAgent - 顺序编排
seqAgent, _ := adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
    Name:      "workflow_agent",
    SubAgents: []adk.Agent{planner, executor, reviewer},
})

// 4. ParallelAgent - 并行执行
parAgent, _ := adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
    Name:      "multi_search",
    SubAgents: []adk.Agent{vectorSearch, keywordSearch, graphSearch},
})
```

### 2.3 Graph 编排模式

```go
// 状态驱动的 Graph
type RagState struct {
    Query          string
    RewrittenQuery string
    SearchResults  []SearchResult
    RerankedResults []SearchResult
    Context        string
    Answer         string
}

// 创建 Graph
g := compose.NewGraph[string, *RagState]()

// 添加节点
g.AddLambdaNode("rewrite_query", rewriteHandler)
g.AddLambdaNode("parallel_search", parallelSearchHandler)
g.AddLambdaNode("rerank", rerankHandler)
g.AddLambdaNode("generate_answer", generateAnswerHandler)

// 添加边
g.AddEdge(compose.START, "rewrite_query")
g.AddEdge("rewrite_query", "parallel_search")
g.AddEdge("parallel_search", "rerank")
g.AddEdge("rerank", "generate_answer")
g.AddEdge("generate_answer", compose.END)

// 作为 Agent 使用
agent, _ := adk.NewAgentFromGraph(ctx, g, &adk.AgentFromGraphConfig{
    Name: "rag_agent",
})
```

---

## 3. 架构映射设计

### 3.1 整体架构对比

| 当前组件 | Eino ADK 对应 | 改造方案 |
|---------|---------------|----------|
| AgentEngine (ReAct) | ReactAgent | 直接使用 ReactAgent，迁移工具系统 |
| ToolRegistry | ToolsNodeConfig | 转换为 Eino BaseTool 接口 |
| AgentStreamHandler | AsyncIterator 消费 | 迭代 AgentEvent 并转换为 SSE |
| EventBus (业务事件) | 保留 EventBus | adk 不支持多订阅者，保留现有 EventBus |
| RAG Pipeline | Compose Graph | 将流水线转换为 Graph 节点 |
| Session 管理 | runSession | 兼容现有 SessionRepository |

### 3.2 目标架构

```
┌─────────────────────────────────────────────────────────────────┐
│                          HTTP Layer                              │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              ChatHandler / StreamHandler                    │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                             │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    AgentService                             │ │
│  │  ├─ CreateAgentSession                                     │ │
│  │  ├─ StreamAgentRun                                        │ │
│  │  └─ Convert AgentEvent to SSE                             │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      ADK Agent Layer                             │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              Main Agent (SequentialAgent)                   │ │
│  │  ┌─────────────┬──────────────┬─────────────────────────┐  │ │
│  │  │  Planner    │   Executor   │      Reviewer           │  │ │
│  │  │  (Graph)    │  (ReactAgent)│       (ChatModelAgent)  │  │ │
│  │  └─────────────┴──────────────┴─────────────────────────┘  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                              │                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    RAG Graph (可选)                         │ │
│  │  rewrite → search → rerank → merge → build_context         │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Tool Layer                                │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────────┐  │
│  │ Knowledge│  Grep    │   KG     │ WebSearch│   Database   │  │
│  │  Search  │  Search  │  Query   │          │    Query     │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Custom EventBus (保留)                         │
│  用于：业务事件广播、多订阅者通知、监控埋点                        │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 详细改造方案

### 4.1 Agent 定义改造

#### 当前实现

```go
// internal/agent/engine.go
type AgentEngine struct {
    config               *types.AgentConfig
    toolRegistry         *tools.ToolRegistry
    chatModel            chat.Chat
    eventBus             *event.EventBus
    // ...
}

func (e *AgentEngine) Execute(ctx context.Context) error {
    // 自定义 ReAct 循环
    for i := 0; i < e.config.MaxIterations; i++ {
        response := e.chatModel.Stream(ctx, messages)
        // 工具调用、结果处理等
    }
}
```

#### 改造方案

```go
// internal/agent/react_agent.go
package agent

import (
    "context"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/flow/agent/react"
    adk "github.com/cloudwego/eino/adk"
)

// NewReactAgent 创建 ReAct 模式的 Agent
func NewReactAgent(ctx context.Context, config *ReactAgentConfig) (adk.Agent, error) {
    // 转换工具配置
    tools := convertTools(config.Tools)

    // 使用 React Agent
    rAgent, err := react.NewAgent(ctx, &react.AgentConfig{
        ToolCallingModel:     config.ChatModel,
        ToolsConfig:          compose.ToolsNodeConfig{Tools: tools},
        MaxIterations:        config.MaxIterations,
        ReflectionEnabled:    config.ReflectionEnabled,
    })

    return rAgent, nil
}

// 工具转换器
func convertTools(tools []ToolConfig) []tool.BaseTool {
    result := make([]tool.BaseTool, 0, len(tools))
    for _, t := range tools {
        result = append(result, &EinoToolAdapter{
            name:        t.Name,
            description: t.Description,
            executor:    t.Executor,
        })
    }
    return result
}

// 工具适配器
type EinoToolAdapter struct {
    name        string
    description string
    executor    ToolExecutor
}

func (a *EinoToolAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: a.name,
        Desc: a.description,
        ParamsOneOf: schema.NewParamsOneOfByParams(a.executor.Params()),
    }, nil
}

func (a *EinoToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    return a.executor.Execute(ctx, argumentsInJSON)
}
```

### 4.2 工具系统改造

#### 当前工具接口

```go
// internal/agent/tools/definitions.go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx, arguments) (*ToolResult, error)
}
```

#### 改造为 Eino BaseTool

```go
// internal/agent/tools/eino_tool.go
package tools

import (
    "context"
    "encoding/json"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/schema"
)

// EinoToolAdapter 将现有工具适配为 Eino BaseTool
type EinoToolAdapter struct {
    tool Tool
}

func NewEinoTool(tool Tool) tool.BaseTool {
    return &EinoToolAdapter{tool: tool}
}

func (a *EinoToolAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
    params := make(map[string]*schema.ParameterInfo)
    for k, v := range a.tool.Parameters() {
        params[k] = convertParamInfo(v)
    }

    return &schema.ToolInfo{
        Name:        a.tool.Name(),
        Desc:        a.tool.Description(),
        ParamsOneOf: schema.NewParamsOneOfByParams(params),
    }, nil
}

func (a *EinoToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    result, err := a.tool.Execute(ctx, argumentsInJSON)
    if err != nil {
        return "", err
    }
    return json.Marshal(result)
}

// 流式工具支持
func (a *EinoToolAdapter) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (
    *schema.StreamReader[string], error) {

    if streamer, ok := a.tool.(StreamingTool); ok {
        return streamer.StreamExecute(ctx, argumentsInJSON)
    }

    // 回退到同步执行
    result, err := a.InvokableRun(ctx, argumentsInJSON, opts...)
    return schema.StreamReaderFromString(result), err
}
```

### 4.3 RAG 流水线改造为 Graph

#### 当前流水线

```go
// 当前是硬编码的步骤序列
const rag_stream = PipelineStages{
    REWRITE_QUERY,
    CHUNK_SEARCH_PARALLEL,
    CHUNK_RERANK,
    CHUNK_MERGE,
    FILTER_TOP_K,
    INTO_CHAT_MESSAGE,
    CHAT_COMPLETION_STREAM,
}
```

#### 改造为 Compose Graph

```go
// internal/agent/graph/rag_graph.go
package graph

import (
    "context"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

// RagState RAG 流水线状态
type RagState struct {
    // 输入
    Query          string
    SessionID      string
    KnowledgeBases []string

    // 中间状态
    RewrittenQuery     string
    VectorResults      []SearchResult
    KeywordResults     []SearchResult
    GraphResults       []SearchResult
    MergedResults      []SearchResult
    RerankedResults    []SearchResult
    ContextMessage     string

    // 输出
    Answer *schema.Message
}

// NewRAGGraph 创建 RAG 流水线 Graph
func NewRAGGraph(ctx context.Context, services *RAGServices) (*compose.Graph[string, *RagState], error) {
    // 初始化状态函数
    genState := func(ctx context.Context) *RagState {
        return &RagState{}
    }

    g := compose.NewGraph[string, *RagState](
        compose.WithGenLocalState(genState),
    )

    // 1. 查询重写节点
    if services.RewriterEnabled {
        g.AddLambdaNode("rewrite_query", newRewriteHandler(services.Rewriter),
            compose.WithStatePreHandler(func(ctx context.Context, input string, state *RagState) (string, error) {
                state.Query = input
                return input, nil
            }),
            compose.WithStatePostHandler(func(ctx context.Context, output string, state *RagState) (*RagState, error) {
                state.RewrittenQuery = output
                return state, nil
            }),
        )
        g.AddEdge(compose.START, "rewrite_query")
    }

    // 2. 并行搜索节点
    searchNode := "parallel_search"
    g.AddLambdaNode(searchNode, newParallelSearchHandler(services.Searchers),
        compose.WithStatePreHandler(func(ctx context.Context, input string, state *RagState) (string, error) {
            // 使用重写后的查询或原查询
            if state.RewrittenQuery != "" {
                return state.RewrittenQuery, nil
            }
            return state.Query, nil
        }),
        compose.WithStatePostHandler(func(ctx context.Context, output SearchResults, state *RagState) (*RagState, error) {
            state.VectorResults = output.Vector
            state.KeywordResults = output.Keyword
            state.GraphResults = output.Graph
            return state, nil
        }),
    )

    // 3. 重排序节点
    rerankNode := "rerank"
    g.AddLambdaNode(rerankNode, newRerankHandler(services.Reranker),
        compose.WithStatePostHandler(func(ctx context.Context, output []SearchResult, state *RagState) (*RagState, error) {
            state.RerankedResults = output
            return state, nil
        }),
    )

    // 4. 构建上下文节点
    contextNode := "build_context"
    g.AddLambdaNode(contextNode, newBuildContextHandler(services.ContextBuilder))

    // 5. 生成答案节点
    answerNode := "generate_answer"
    g.AddLambdaNode(answerNode, newGenerateAnswerHandler(services.ChatModel),
        compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *RagState) (*RagState, error) {
            state.Answer = output
            return state, nil
        }),
    )

    // 连接边
    if services.RewriterEnabled {
        g.AddEdge("rewrite_query", searchNode)
    } else {
        g.AddEdge(compose.START, searchNode)
    }
    g.AddEdge(searchNode, rerankNode)
    g.AddEdge(rerankNode, contextNode)
    g.AddEdge(contextNode, answerNode)
    g.AddEdge(answerNode, compose.END)

    return g, nil
}

// RAGGraphAgent 将 Graph 包装为 Agent
func RAGGraphAgent(ctx context.Context, graph *compose.Graph[string, *RagState], name string) (adk.Agent, error) {
    return adk.NewAgentFromGraph(ctx, graph, &adk.AgentFromGraphConfig{
        Name:        name,
        Description: "RAG Enhanced Q&A Agent",
    })
}
```

### 4.4 流式输出改造

#### 当前实现

```go
// 当前通过 EventBus 推送事件
func (e *AgentEngine) streamThinkingToEventBus(ctx context.Context, stream *schema.StreamReader[*schema.Message]) error {
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        e.eventBus.Emit(ctx, event.Event{
            Type: event.EventAgentThought,
            Data: event.AgentThoughtData{
                Content: chunk.Content,
            },
        })
    }
    return nil
}
```

#### 改造方案

```go
// internal/handler/session/adk_stream_handler.go
package session

import (
    "context"
    "io"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/schema"
)

// ADKStreamHandler 处理 ADK Agent 的流式输出
type ADKStreamHandler struct {
    eventBus   *event.EventBus
    streamMgr  *StreamManager
    sessionID  string
}

// StreamAgentRun 执行 Agent 并流式处理输出
func (h *ADKStreamHandler) StreamAgentRun(ctx context.Context, agent adk.Agent, input *adk.AgentInput) error {
    // 1. 运行 Agent
    iter := agent.Run(ctx, input)

    // 2. 处理事件流
    for {
        agentEvent, ok := iter.Next()
        if !ok {
            break
        }

        // 3. 处理错误
        if agentEvent.Err != nil {
            h.handleError(ctx, agentEvent.Err)
            return agentEvent.Err
        }

        // 4. 处理 Action (转移、退出等)
        if agentEvent.Action != nil {
            h.handleAction(ctx, agentEvent.Action)
        }

        // 5. 处理输出消息
        if agentEvent.Output != nil && agentEvent.Output.MessageOutput != nil {
            if err := h.handleMessageOutput(ctx, agentEvent); err != nil {
                return err
            }
        }
    }

    return nil
}

// handleMessageOutput 处理消息输出
func (h *ADKStreamHandler) handleMessageOutput(ctx context.Context, agentEvent *adk.AgentEvent) error {
    output := agentEvent.Output.MessageOutput

    // 非流式消息
    if !output.IsStreaming {
        return h.sendNonStreamMessage(ctx, output.Message)
    }

    // 流式消息 - 分块发送
    stream := output.MessageStream
    defer stream.Close()

    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            h.eventBus.Emit(ctx, event.Event{
                Type: event.EventError,
                Data: event.ErrorData{Error: err},
            })
            return err
        }

        // 发送到 EventBus (保持现有前端兼容)
        h.eventBus.Emit(ctx, event.Event{
            Type: event.EventAgentThought,
            Data: event.AgentThoughtData{
                Content:   msg.Content,
                SessionID: h.sessionID,
            },
        })

        // 同时发送到 StreamManager (SSE)
        h.streamMgr.Send(h.sessionID, StreamEvent{
            Type:    StreamTypeThought,
            Content: msg.Content,
        })
    }

    return nil
}

// sendNonStreamMessage 发送非流式消息
func (h *ADKStreamHandler) sendNonStreamMessage(ctx context.Context, msg *schema.Message) error {
    // 处理工具调用
    if len(msg.ToolCalls) > 0 {
        for _, tc := range msg.ToolCalls {
            h.eventBus.Emit(ctx, event.Event{
                Type: event.EventAgentToolCall,
                Data: event.AgentToolCallData{
                    ToolName: tc.Function.Name,
                    Arguments: tc.Function.Arguments,
                },
            })
        }
        return nil
    }

    // 处理普通消息
    h.eventBus.Emit(ctx, event.Event{
        Type: event.EventAgentFinalAnswer,
        Data: event.AgentFinalAnswerData{
            Content: msg.Content,
        },
    })

    return nil
}
```

### 4.5 复杂 Agent 的 Graph 内部编排

#### 场景：带 RAG 和反思的复杂 Agent

```go
// internal/agent/graph/complex_agent.go
package graph

import (
    "context"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
    adk "github.com/cloudwego/eino/adk"
)

// ComplexAgentState 复杂 Agent 状态
type ComplexAgentState struct {
    // 输入
    UserQuery string
    SessionID string

    // RAG 阶段
    RagContext    string
    RagReferences []Reference

    // 思考阶段
    Thoughts      []Thought
    SelectedTools []string

    // 执行阶段
    ToolResults   []ToolResult

    // 反思阶段
    Reflection    string
    NeedsRetry    bool

    // 最终输出
    FinalAnswer   string
}

// NewComplexAgentGraph 创建复杂 Agent Graph
func NewComplexAgentGraph(ctx context.Context, services *ComplexAgentServices) (*compose.Graph[string, *ComplexAgentState], error) {
    g := compose.NewGraph[string, *ComplexAgentState]()

    // 1. RAG 检索节点 (条件执行)
    g.AddLambdaNode("rag_retrieval", newRAGRetrievalHandler(services.RAG))

    // 2. 思考节点
    g.AddChatModelNode("thinking", services.ChatModel,
        compose.WithStatePreHandler(func(ctx context.Context, input string, state *ComplexAgentState) ([]*schema.Message, error) {
            return buildThinkingMessages(state), nil
        }),
        compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *ComplexAgentState) (*ComplexAgentState, error) {
            state.Thoughts = append(state.Thoughts, parseThoughts(output))
            state.SelectedTools = extractToolCalls(output)
            return state, nil
        }),
    )

    // 3. 条件分支：是否有工具调用
    g.AddBranch("has_tools", func(ctx context.Context, input string, state *ComplexAgentState) (string, error) {
        if len(state.SelectedTools) > 0 {
            return "execute_tools"
        }
        return "generate_answer", nil
    })

    // 4. 工具执行节点
    g.AddToolsNode("execute_tools", compose.ToolsNodeConfig{
        Tools: services.Tools,
    },
        compose.WithStatePostHandler(func(ctx context.Context, output []tool.CallResult, state *ComplexAgentState) (*ComplexAgentState, error) {
            state.ToolResults = convertToolResults(output)
            return state, nil
        }),
    )

    // 5. 反思节点 (条件执行)
    g.AddLambdaNode("reflection", newReflectionHandler(services.ChatModel))

    // 6. 重试判断分支
    g.AddBranch("should_retry", func(ctx context.Context, input string, state *ComplexAgentState) (string, error) {
        if state.NeedsRetry && len(state.Thoughts) < state.MaxIterations {
            return "thinking"  // 回到思考
        }
        return "generate_answer", nil
    })

    // 7. 生成最终答案
    g.AddChatModelNode("generate_answer", services.ChatModel)

    // 连接边
    g.AddEdge(compose.START, "rag_retrieval")
    g.AddEdge("rag_retrieval", "thinking")
    g.AddEdge("thinking", "has_tools")

    g.AddBranchEdge("has_tools", "execute_tools")
    g.AddBranchEdge("has_tools", "generate_answer")

    g.AddEdge("execute_tools", "reflection")
    g.AddEdge("reflection", "should_retry")

    g.AddBranchEdge("should_retry", "thinking")
    g.AddBranchEdge("should_retry", "generate_answer")

    g.AddEdge("generate_answer", compose.END)

    return g, nil
}
```

---

## 5. 改造点清单

### 5.1 核心改造

| 模块 | 当前实现 | 目标实现 | 改造难度 |
|------|---------|---------|---------|
| Agent 引擎 | 自定义 AgentEngine | ReactAgent | 中 |
| 工具系统 | ToolRegistry + Tool | ToolsNodeConfig + BaseTool | 低 |
| RAG 流水线 | 硬编码 Pipeline | Compose Graph | 中 |
| 流式输出 | EventBus 事件 | AsyncIterator + EventBus | 中 |
| 多 Agent | 不支持 | Sequential/Parallel Agent | 低 |
| 计划执行 | 不支持 | PlanExecuteAgent | 高 |

### 5.2 保留组件

| 组件 | 原因 |
|------|------|
| EventBus | adk 不支持多订阅者，保留用于业务事件 |
| SessionRepository | 与现有数据模型兼容 |
| KnowledgeBase | 现有知识库实现 |
| ContextManager | 渐进式压缩逻辑 |

### 5.3 新增组件

| 组件 | 用途 |
|------|------|
| AgentFactory | 创建不同类型的 Agent |
| ToolAdapter | 工具接口适配器 |
| EventConverter | AgentEvent 转 EventBus 事件 |
| GraphRegistry | Graph 定义注册表 |

---

## 6. 迁移路径

### 6.1 Phase 1: 基础设施 (2-3 周)

```
Week 1-2: 核心适配
├── 引入 eino 依赖
├── 实现 ToolAdapter
├── 实现 EventConverter
└── 单元测试

Week 3: 基础 Agent
├── 实现 SimpleChatAgent (ChatModelAgent)
├── 实现 ReactAgent (带工具)
└── 流式输出适配
```

### 6.2 Phase 2: RAG 流水线 (2-3 周)

```
Week 4-5: Graph 编排
├── 实现 RAG Graph
├── 节点实现 (检索、重排序、合并)
└── 与现有检索服务集成

Week 6: 集成测试
├── 端到端测试
├── 性能对比
└── 修复问题
```

### 6.3 Phase 3: 高级特性 (3-4 周)

```
Week 7-8: 复杂 Agent
├── 实现 Sequential Agent
├── 实现 Parallel Agent
├── Plan-Execute 模式
└── 条件分支和循环

Week 9-10: 优化
├── 性能优化
├── 错误处理
├── 监控和日志
└── 文档完善
```

### 6.4 Phase 4: 上线验证 (2 周)

```
Week 11: 灰度发布
├── A/B 测试
├── 功能验证
└── 性能监控

Week 12: 全量上线
├── 监控告警
├── 问题修复
└── 回滚预案
```

---

## 7. 风险与挑战

### 7.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| eino 框架成熟度 | 中 | 充分测试，保留回滚方案 |
| 性能下降 | 中 | 基准测试对比，性能优化 |
| 调试复杂度增加 | 低 | 完善日志和追踪 |
| 迁移成本 | 高 | 分阶段迁移，双写验证 |

### 7.2 业务风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 功能缺失 | 高 | 功能对比清单，确保全覆盖 |
| 用户体验变化 | 中 | 保持 API 兼容性 |
| 数据一致性 | 中 | 复用现有 Repository |

---

## 8. 总结

### 8.1 改造收益

1. **标准化**：使用业界标准的 Agent 框架
2. **可扩展性**：更灵活的 Agent 编排能力
3. **可维护性**：清晰的组件边界和职责
4. **生态集成**：可利用 eino 生态的扩展

### 8.2 关键决策

1. **保留 EventBus**：adk 不满足多订阅者场景
2. **分阶段迁移**：降低风险，确保平滑过渡
3. **兼容性优先**：保持 API 和数据模型兼容

### 8.3 下一步

1. 完成 PoC 验证核心功能
2. 建立性能基准
3. 制定详细的迁移计划
