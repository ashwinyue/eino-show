---
name: eino
description: Search Eino examples in a-old/old/eino-examples/. Use this command to find reference implementations for ChatModel, Agent, Tool, Retriever, and other Eino components. Searches through official examples to find best practices and usage patterns.
---

搜索 Eino 官方示例代码。

## 使用方法

```
/eino <搜索关键词>
```

## 示例

```
/eino ReactAgent
/eino Tool
/eino StreamRun
/eino ChatModel
```

## 搜索范围

```
a-old/old/eino-examples/
├── compose/graph/          # Graph 编排
│   ├── tool_call_agent/    # Tool 调用 Agent
│   ├── react_with_interrupt/
│   └── state/
├── quickstart/
│   ├── eino_assistant/     # 助手示例
│   └── chat/
└── adk/                    # ADK 示例
    ├── helloworld/
    ├── multiagent/
    └── human-in-the-loop/
```

## 常用搜索

| 查找什么 | 搜索关键词 |
|---------|-----------|
| Agent 创建 | `adk.NewChatModelAgent` |
| 工具实现 | `InvokableTool` |
| 流式输出 | `StreamRun` |
| ChatModel | `NewChatModel` |
| Tool 调用 | `tool_call_agent` |
