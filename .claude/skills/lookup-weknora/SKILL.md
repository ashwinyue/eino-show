---
name: lookup-weknora
description: Find WeKnora business implementation in a-old/WeKnora/ as refactoring reference. Use this skill when implementing features (Agent, Session, Knowledge) to understand how WeKnora handles them. Extract business logic (data models, API design, core flow), but replace WeKnora's custom LLM/Embedding/Tool wrappers with Eino components.
---

查找 WeKnora 业务实现 Skill。在 `a-old/WeKnora/` 目录中查找特定功能的实现，作为重构参考。

## 使用场景
当你需要实现某个功能（如 Agent、Session、Knowledge），想了解 WeKnora 中是如何实现的。

## 操作步骤

### 1. 定位功能模块
WeKnora 目录结构（在 `a-old/WeKnora/`）：
```
a-old/WeKnora/
├── internal/
│   ├── agent/          # Agent 服务
│   ├── chat/           # Chat 服务
│   ├── knowledge/      # 知识库服务
│   ├── document/       # 文档处理
│   ├── retrieval/      # 检索服务
│   ├── llm/            # LLM 调用
│   └── session/        # 会话管理
├── migrations/         # 数据库迁移
└── docreader/          # 文档解析
```

### 2. 搜索相关代码
使用 ripgrep 搜索关键词：
```bash
# 在 WeKnora 中搜索函数名
rg "func.*CreateAgent" a-old/WeKnora/

# 搜索结构体定义
rg "type Agent struct" a-old/WeKnora/

# 搜索 API 路由
rg "router.*POST" a-old/WeKnora/

# 搜索特定功能
rg "ToolCall|StreamChat" a-old/WeKnora/
```

### 3. 提取业务逻辑
阅读 WeKnora 代码时，关注：
- **业务流程**：核心逻辑是什么
- **数据模型**：如何组织数据（见 `migrations/`）
- **API 设计**：接口定义、请求/响应格式
- **SSE 事件格式**：前端依赖的事件结构

### 4. 忽略的实现细节
以下 WeKnora 的实现方式**不要**复制，用 Eino 替换：
| WeKnora | 替换为 |
|---------|--------|
| 自定义 LLM 封装 | Eino ChatModel |
| 自定义 Embedding | Eino Embedding |
| 自定义 Vector Store | Eino Retriever / PGVector |
| 自定义 Tool 框架 | Eino Tool 接口 |

## 示例

### 任务：实现 Agent 创建功能
```bash
# 1. 查找 WeKnora 中的 Agent 创建
rg "CreateAgent|CreateCustomAgent" a-old/WeKnora/

# 2. 找到相关文件
a-old/WeKnora/internal/agent/service.go
a-old/WeKnora/internal/agent/types.go
a-old/WeKnora/internal/agent/router.go

# 3. 阅读：了解需要哪些字段、验证逻辑、API 设计

# 4. 在 eino-show 中按四层架构重写
# model/agent.gen.go       → 数据模型
# store/agent.go           → 数据访问
# biz/v1/agent/agent.go    → 业务逻辑
# handler/http/agent.go    → HTTP 处理
```

### 任务：实现 SSE 流式问答
```bash
# 1. 查找 WeKnora 中的流式实现
rg "SSE|EventSource" a-old/WeKnora/

# 2. 找到事件格式定义
a-old/WeKnora/internal/chat/types.go

# 3. 保持事件格式兼容（确保前端无需修改）

# 4. 在 eino-show 中实现
# handler/http/session.go → StreamQA()
```

## eino-show 架构映射

WeKnora 的功能要映射到 eino-show 的四层架构：

| WeKnora | eino-show |
|---------|-----------|
| API 路由 | `internal/apiserver/handler/http/` |
| 业务逻辑 | `internal/apiserver/biz/v1/` |
| 数据访问 | `internal/apiserver/store/` |
| 数据模型 | `internal/apiserver/model/` |
| Agent 执行 | `internal/agent/` (Eino 实现) |

## 注意事项
- WeKnora 代码仅作为**功能参考**，不要直接复制
- 所有 AI 组件必须用 Eino 标准实现
- 遵循 miniblog-x 四层架构 (Handler → Biz → Store → Model)
- 保持 API 接口和 SSE 事件格式兼容
