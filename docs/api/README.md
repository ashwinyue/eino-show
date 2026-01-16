# eino-show API 文档

> 参考自 WeKnora API，用于接口兼容性验证

## 概述

本文档描述 WeKnora 的 API 接口规范。eino-show 重构时需保持这些接口的兼容性，确保前端无需修改。

## 核心接口（重构重点）

### 会话相关
```
POST   /api/v1/sessions              创建会话
GET    /api/v1/sessions              会话列表
GET    /api/v1/sessions/:id          会话详情
PUT    /api/v1/sessions/:id          更新会话
DELETE /api/v1/sessions/:id          删除会话
POST   /api/v1/sessions/:id/qa       流式问答 (SSE)
```

### Agent 相关
```
GET    /api/v1/custom-agents         Agent 列表
POST   /api/v1/custom-agents         创建 Agent
GET    /api/v1/custom-agents/:id     Agent 详情
PUT    /api/v1/custom-agents/:id     更新 Agent
DELETE /api/v1/custom-agents/:id     删除 Agent
```

### 知识库相关
```
GET    /api/v1/knowledge-bases       知识库列表
POST   /api/v1/knowledge-bases       创建知识库
GET    /api/v1/knowledge-bases/:id   知识库详情
PUT    /api/v1/knowledge-bases/:id   更新知识库
DELETE /api/v1/knowledge-bases/:id   删除知识库
GET    /api/v1/knowledge-bases/:id/stats  知识库统计
POST   /api/v1/knowledge-bases/:id/documents  上传文档
POST   /api/v1/knowledge/search     知识搜索
```

## SSE 事件格式

流式问答接口 (`/api/v1/sessions/:id/qa`) 返回以下事件类型：

| 事件类型 | 说明 |
|---------|------|
| `agent_thinking` | 思考中 |
| `agent_action` | 工具调用 |
| `agent_observation` | 工具结果 |
| `agent_complete` | 完成 |
| `agent_error` | 错误 |

## 详细文档

- [chat.md](./chat.md) - 聊天功能
- [session.md](./session.md) - 会话管理
- [knowledge-base.md](./knowledge-base.md) - 知识库管理
- [knowledge.md](./knowledge.md) - 知识管理
- [knowledge-search.md](./knowledge-search.md) - 知识搜索
- [model.md](./model.md) - 模型管理
