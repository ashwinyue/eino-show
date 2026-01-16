---
name: api-compatibility
description: API compatibility checker for eino-show refactoring. Use this skill to verify that new implementations maintain API compatibility with WeKnora (a-old/WeKnora/), ensuring the frontend works without changes. Check REST endpoints, request/response formats, and SSE event structures.
---

API 兼容性检查 Skill。确保重构后的 API 与 WeKnora 保持兼容，前端无需修改。

## 使用场景
实现新功能或修改现有功能时，验证 API 兼容性。

## 兼容性要求

### 1. REST API 接口

#### Session 相关
```
GET    /api/v1/sessions              → 会话列表
POST   /api/v1/sessions              → 创建会话
GET    /api/v1/sessions/:id          → 会话详情
PUT    /api/v1/sessions/:id          → 更新会话
DELETE /api/v1/sessions/:id          → 删除会话
POST   /api/v1/sessions/:id/qa       → 流式问答 (SSE)
```

#### Agent 相关
```
GET    /api/v1/custom-agents         → Agent 列表
POST   /api/v1/custom-agents         → 创建 Agent
GET    /api/v1/custom-agents/:id     → Agent 详情
PUT    /api/v1/custom-agents/:id     → 更新 Agent
DELETE /api/v1/custom-agents/:id     → 删除 Agent
```

#### Knowledge 相关
```
GET    /api/v1/knowledge-bases       → 知识库列表
POST   /api/v1/knowledge-bases       → 创建知识库
GET    /api/v1/knowledge-bases/:id   → 知识库详情
PUT    /api/v1/knowledge-bases/:id   → 更新知识库
DELETE /api/v1/knowledge-bases/:id   → 删除知识库
GET    /api/v1/knowledge-bases/:id/stats  → 知识库统计
POST   /api/v1/knowledge-bases/:id/documents  → 上传文档
POST   /api/v1/knowledge/search     → 知识搜索
```

### 2. SSE 事件格式

流式问答 (`/api/v1/sessions/:id/qa`) 必须返回以下事件格式：

```json
// 思考事件
{"event": "agent_thinking", "data": {"content": "正在思考..."}}

// 工具调用事件
{"event": "agent_action", "data": {"tool": "knowledge_search", "input": "..."}}

// 工具结果事件
{"event": "agent_observation", "data": {"tool": "knowledge_search", "output": "..."}}

// 完成事件
{"event": "agent_complete", "data": {"answer": "最终答案"}}

// 错误事件
{"event": "agent_error", "data": {"message": "错误信息"}}
```

### 3. 请求/响应格式

#### 创建会话请求
```json
{
  "title": "会话标题",
  "description": "会话描述",
  "agent_config": {
    "agent_id": "builtin-quick-answer",
    "mode": "rag",
    "knowledge_bases": ["kb-id-1"],
    "temperature": 0.7,
    "max_iterations": 10,
    "web_search_enabled": false
  },
  "context_config": {
    "max_messages": 20,
    "compression_threshold": 1000
  }
}
```

#### 流式问答请求
```json
{
  "query": "用户问题",
  "stream": true
}
```

## 检查步骤

### 1. 查找旧 API 定义
```bash
# 搜索路由定义
rg "POST.*sessions|GET.*sessions" a-old/WeKnora/

# 搜索 SSE 事件定义
rg "agent_thinking|agent_action|agent_complete" a-old/WeKnora/

# 查找请求/响应结构
rg "type.*Request|type.*Response" a-old/WeKnora/internal/
```

### 2. 对比新实现
```bash
# 检查新实现的路由
rg "POST.*sessions|GET.*sessions" internal/apiserver/handler/

# 检查 SSE 事件格式
rg "SSEvent|agent_" internal/apiserver/handler/http/session.go
```

### 3. 验证兼容性
- [ ] 路由路径完全一致
- [ ] 请求字段名和类型一致
- [ ] 响应字段名和类型一致
- [ ] SSE 事件类型名称一致
- [ ] SSE 事件数据结构一致

## 常见不兼容问题

| 问题 | 检查方法 | 修复方式 |
|------|----------|----------|
| 路由路径变化 | 对比 router 注册 | 保持路径一致 |
| 字段名大小写 | 检查 JSON tag | 保持 tag 一致 |
| 事件类型名称 | 检查 SSEvent 调用 | 使用相同的 event 名称 |
| 缺少必需字段 | 对比 Request 结构 | 补充缺失字段 |
| 响应结构变化 | 检查 Response 定义 | 保持结构一致 |

## 示例：验证 SSE 兼容性

```bash
# 1. 找到旧实现的事件定义
cat a-old/WeKnora/internal/chat/types.go | grep -A 5 "EventType"

# 2. 找到旧实现的 SSE 发送代码
cat a-old/WeKnora/internal/chat/service.go | grep -A 10 "SSEvent"

# 3. 对比新实现
cat internal/apiserver/handler/http/session.go | grep -A 10 "SSEvent"

# 4. 确保事件名称一致
# 旧: "agent_thinking", "agent_action", "agent_observation", "agent_complete"
# 新: 必须相同
```

## 前端依赖检查

前端代码依赖的 SSE 事件处理逻辑：
```javascript
// 参考前端代码
eventSource.addEventListener('agent_thinking', handler)
eventSource.addEventListener('agent_action', handler)
eventSource.addEventListener('agent_observation', handler)
eventSource.addEventListener('agent_complete', handler)
eventSource.addEventListener('agent_error', handler)
```

确保后端发送的事件名称与前端监听的事件名称**完全一致**。
