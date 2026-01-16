---
name: weknora
description: Search WeKnora implementation in a-old/WeKnora/. Use this command to find business logic, API definitions, data models, and SSE event formats. Helps ensure API compatibility during refactoring.
---

搜索 WeKnora 业务实现代码。

## 使用方法

```
/weknora <搜索关键词>
```

## 示例

```
/weknora CreateAgent
/weknora SSE
/weknora Session
/weknora Knowledge
```

## 搜索范围

```
a-old/WeKnora/
├── internal/
│   ├── agent/          # Agent 服务
│   ├── chat/           # Chat 服务
│   ├── knowledge/      # 知识库服务
│   ├── session/        # 会话管理
│   └── llm/            # LLM 调用
├── migrations/         # 数据库迁移
└── docs/api/           # API 文档
```

## 常用搜索

| 查找什么 | 搜索关键词 |
|---------|-----------|
| API 路由 | `router.*POST`, `gin.*POST` |
| SSE 事件 | `agent_thinking`, `SSEvent` |
| 数据模型 | `type.*struct`, `gorm:` |
| 业务逻辑 | `func.*Create`, `func.*Get` |
| 请求/响应 | `Request`, `Response` |

## 兼容性注意

搜索时关注：
- API 路径格式
- 请求/响应字段名
- SSE 事件类型名称
- 数据模型字段定义
