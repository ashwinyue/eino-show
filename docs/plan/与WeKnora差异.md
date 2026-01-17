# eino-show 与 WeKnora 后端差异分析

## 对齐目标

确保后端 API 与 WeKnora 完全兼容，前端无需修改。

---

## P0: 核心差异（必须对齐）

### 1. Session API 路由差异

| 功能 | WeKnora 路由 | eino-show 路由 | 状态 |
|------|-------------|---------------|------|
| 问答 | `POST /knowledge-chat/:session_id` | `POST /sessions/:id/qa` | 需添加兼容路由 |
| Agent问答 | `POST /agent-chat/:session_id` | 缺失 | 需添加 |
| 生成标题 | `POST /sessions/:session_id/generate_title` | 内部异步 | 需添加接口 |
| 停止生成 | `POST /sessions/:session_id/stop` | ✅ 已实现 | - |
| 继续流 | `GET /sessions/continue-stream/:session_id` | 缺失 | 需添加 |
| 知识搜索 | `POST /knowledge-search` | 缺失 | 需添加 |

### 2. Agent API 路由差异

| 功能 | WeKnora 路由 | eino-show 路由 | 状态 |
|------|-------------|---------------|------|
| Agent列表 | `GET /agents` | `GET /custom-agents` | 需添加 /agents 兼容 |
| 获取占位符 | `GET /agents/placeholders` | 缺失 | 需添加 |
| 复制Agent | `POST /agents/:id/copy` | 缺失 | 需添加 |

### 3. Message API（缺失）

| 功能 | WeKnora 路由 | 状态 |
|------|-------------|------|
| 加载消息 | `GET /messages/:session_id/load` | 需添加 |
| 删除消息 | `DELETE /messages/:session_id/:id` | 需添加 |

### 4. Knowledge API 差异

| 功能 | WeKnora 路由 | eino-show 路由 | 状态 |
|------|-------------|---------------|------|
| 获取知识 | `GET /knowledge/:id` | 缺失 | 需添加 |
| 更新知识 | `PUT /knowledge/:id` | 缺失 | 需添加 |
| 批量获取 | `GET /knowledge/batch` | 缺失 | 需添加 |
| 下载文件 | `GET /knowledge/:id/download` | 缺失 | 需添加 |
| 搜索知识 | `GET /knowledge/search` | 缺失 | 需添加 |
| 更新手工知识 | `PUT /knowledge/manual/:id` | 缺失 | 需添加 |
| 批量更新标签 | `PUT /knowledge/tags` | 缺失 | 需添加 |

### 5. Chunk API 差异

| 功能 | WeKnora 路由 | eino-show 路由 | 状态 |
|------|-------------|---------------|------|
| 列表 | `GET /chunks/:knowledge_id` | `GET /chunks?knowledge_id=xxx` | 需添加路径参数版本 |
| 删除所有 | `DELETE /chunks/:knowledge_id` | 缺失 | 需添加 |
| 删除单个 | `DELETE /chunks/:knowledge_id/:id` | `DELETE /chunks/:id` | 需添加路径参数版本 |
| 更新 | `PUT /chunks/:knowledge_id/:id` | `PUT /chunks/:id` | 需添加路径参数版本 |

### 6. Model API 差异

| 功能 | WeKnora 路由 | eino-show 路由 | 状态 |
|------|-------------|---------------|------|
| 厂商列表 | `GET /models/providers` | 缺失 | 需添加 |

### 7. MCP API 差异

| 功能 | WeKnora 路由 | eino-show 路由 | 状态 |
|------|-------------|---------------|------|
| 获取资源 | `GET /mcp-services/:id/resources` | 缺失 | 需添加 |

---

## P1: 次要差异

### 8. Knowledge Base API 差异

| 功能 | WeKnora 路由 | 状态 |
|------|-------------|------|
| 复制知识库 | `POST /knowledge-bases/copy` | 需添加 |
| 复制进度 | `GET /knowledge-bases/copy/progress/:task_id` | 需添加 |

### 9. Web Search API（缺失）

| 功能 | WeKnora 路由 | 状态 |
|------|-------------|------|
| 搜索引擎列表 | `GET /web-search/providers` | 需添加 |

---

## P2: 可选功能

### 10. Tag API（全新模块）

- `GET /knowledge-bases/:id/tags`
- `POST /knowledge-bases/:id/tags`
- `PUT /knowledge-bases/:id/tags/:tag_id`
- `DELETE /knowledge-bases/:id/tags/:tag_id`

### 11. FAQ API（全新模块）

- `/knowledge-bases/:id/faq/entries` 下的 CRUD
- FAQ 搜索、导入导出等

### 12. Tenant API（全新模块）

- 租户管理 CRUD
- KV 配置管理

### 13. Initialization API（全新模块）

- Ollama 模型管理
- 远程 API 检查
- 配置管理

### 14. System API（全新模块）

- 系统信息
- MinIO buckets

### 15. Evaluation API（全新模块）

- 评估功能

---

## 请求/响应格式对齐

### Session 请求

WeKnora `CreateKnowledgeQARequest`:
```json
{
  "query": "string",
  "knowledge_base_ids": ["string"],
  "knowledge_ids": ["string"],
  "agent_enabled": true,
  "agent_id": "string",
  "web_search_enabled": true,
  "summary_model_id": "string",
  "mentioned_items": [{"id": "", "name": "", "type": "", "kb_type": ""}],
  "disable_title": false
}
```

eino-show 当前 `ExecuteRequest`:
```json
{
  "question": "string"
}
```

**需要扩展 ExecuteRequest 以支持更多字段**

### SSE 事件类型

WeKnora 事件类型:
- `agent_query` - 查询开始
- `answer` - 回答内容
- `references` - 引用
- `thinking` - 思考过程
- `tool_call` - 工具调用
- `tool_result` - 工具结果
- `reflection` - 反思
- `session_title` - 标题生成
- `complete` - 完成
- `error` - 错误
- `stop` - 停止

eino-show 当前事件类型:
- `agent_query`
- `answer`
- `complete`
- `error`

**需要补充其他事件类型**

---

## 实施计划

### Phase 1: 核心路由对齐
1. 添加 `/knowledge-chat/:session_id` 路由
2. 添加 `/agent-chat/:session_id` 路由
3. 添加 `/sessions/:session_id/generate_title` 路由
4. 添加 Message API
5. 添加 `/agents` 路由

### Phase 2: 请求格式对齐
1. 扩展 ExecuteRequest 字段
2. 添加 MentionedItems 支持
3. 添加更多 SSE 事件类型

### Phase 3: 补充功能
1. Knowledge 扩展 API
2. Chunk 路径参数版本
3. Model providers API

### Phase 4: 可选模块
1. Tag/FAQ/Tenant 等模块按需添加

---

## 当前实现状态

### ✅ Handler 已定义 + 路由已注册

| 路由 | 状态 |
|------|------|
| `POST /knowledge-chat/:session_id` | ✅ |
| `POST /agent-chat/:session_id` | ✅ |
| `POST /knowledge-search` | ✅ |
| `POST /sessions/:id/generate_title` | ✅ |
| `GET /sessions/continue-stream/:id` | ✅ |

### ⚠️ Handler 已定义，路由待注册

| 分类 | 路由 |
|------|------|
| Message | `/messages/:session_id/load`, `/messages/:session_id/:id` |
| Agent | `/agents`, `/agents/placeholders`, `/agents/:id/copy` |
| Tenant | `/tenants/*` 全部 |
| Tag | `/knowledge-bases/:id/tags/*` 全部 |
| System | `/system/*` 全部 |
| Evaluation | `/evaluation/*` 全部 |
| Initialization | `/initialization/*` 全部 |
| WebSearch | `/web-search/providers` |
| Knowledge | `/knowledge/:id`, `/knowledge/batch`, `/knowledge/manual/:id` 等 |
| Chunk | `/chunks/:knowledge_id` 路径参数版本 |
| Model | `/models/providers` |
| MCP | `/mcp-services/:id/resources` |
| KB | `/knowledge-bases/copy`, `/knowledge-bases/copy/progress/:task_id` |

### ❌ 功能待实现（当前返回占位符）

| 分类 | 说明 |
|------|------|
| MessageBiz | 需要 Store 层支持 |
| SSE 事件 | 仅 3 种，需扩展到 12 种 |
| ExecuteRequest | 需扩展字段 |
| Agent 配置 | 需扩展字段 |

---

## 更新日志

- 2026-01-16: 更新实现状态，Handler 分文件完成
- 2026-01-16: 初始版本，完成差异分析
