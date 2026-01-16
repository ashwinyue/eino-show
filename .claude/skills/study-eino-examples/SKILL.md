---
name: study-eino-examples
description: Learn Eino best practices from official examples in a-old/old/eino-examples/. Use this skill when implementing Eino components (ChatModel, Agent, Tool, Retriever) to understand correct usage patterns. Reference compose/graph/tool_call_agent, compose/graph/react_with_interrupt, quickstart/eino_assistant. Key patterns: adk.NewChatModelAgent, NewChatModel, tool.InvokableTool.
---

参考 Eino 官方示例 Skill。查找和学习 eino-examples 中的最佳实践，确保代码风格符合 Eino 标准。

## 使用场景
当你需要实现某个 Eino 组件（ChatModel、Agent、Tool、Retriever），想了解正确的使用方式。

## 操作步骤

### 1. 定位相关示例
eino-examples 目录结构（在 `a-old/old/eino-examples/`）：
```
a-old/old/eino-examples/
├── compose/
│   └── graph/
│       ├── tool_call_agent/      # Tool 调用 Agent
│       ├── react_with_interrupt/ # ReAct with 中断
│       └── state/                # 状态管理
├── quickstart/
│   ├── eino_assistant/           # 助手示例
│   │   └── eino/tool/            # Tool 实现
│   └── chat/                     # Chat 示例
└── adk/
    ├── helloworld/               # ADK 快速开始
    ├── multiagent/               # 多 Agent
    └── human-in-the-loop/        # 人机协作
```

### 2. 搜索示例代码
```bash
# 查找 Agent 示例
rg "adk.NewChatModelAgent" a-old/old/eino-examples/

# 查找 ChatModel 使用
rg "NewChatModel" a-old/old/eino-examples/

# 查找 Tool 实现
rg "tool.InvokableTool" a-old/old/eino-examples/

# 查找流式输出
rg "StreamRun" a-old/old/eino-examples/
```

### 3. 学习代码模式
关注以下模式：

#### ReactAgent 创建模式
```go
// 参考 a-old/old/eino-examples/compose/graph/tool_call_agent/
agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:          "react_agent",
    Description:   "Agent that can use tools",
    Instruction:   systemPrompt,
    Model:         chatModel,
    MaxIterations: 10,
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []einotool.BaseTool{tool1, tool2},
        },
    },
})

// 流式执行
for chunk := range agent.StreamRun(ctx, input) {
    // 处理事件
}
```

#### Tool 实现模式
```go
// 直接实现 tool.InvokableTool 接口
type MyTool struct{}

func (t *MyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "my_tool",
        Desc: "Tool description",
        ParamsOneOf: schema.NewParamsOneOfByParams(
            &ToolInput{},
        ),
    }, nil
}

func (t *MyTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 解析参数
    var input ToolInput
    json.Unmarshal([]byte(argumentsInJSON), &input)
    // 实现逻辑
    return result, nil
}
```

### 4. 应用到 eino-show
在 `internal/agent/` 中按相同模式实现：
- 直接使用 eino 类型
- 通过 `internal/pkg/agent/` 接口与 Biz 层解耦
- 初始化函数使用 `newXxx()` 命名

## 示例

### 任务：创建带工具的 Agent
```bash
# 1. 查找示例
ls a-old/old/eino-examples/compose/graph/tool_call_agent/

# 2. 阅读示例代码
cat a-old/old/eino-examples/compose/graph/tool_call_agent/main.go

# 3. 在 eino-show 中实现
# internal/agent/react/agent.go
# internal/agent/tool/knowledge_search.go
```

## 核心原则
1. **参考官方示例**：确保实现符合 Eino 最佳实践
2. **接口解耦**：Biz 层依赖 `internal/pkg/agent/Agent` 接口，不直接依赖 Eino
3. **适配器模式**：`internal/agent/` 中实现 Eino 适配层
