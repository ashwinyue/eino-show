# eino-show 项目规约

## 开发流程

### 首次启动

```bash
# 1. 创建环境配置
cp .env.example .env

# 2. 启动基础设施（PostgreSQL + Redis）
make dev-start

# 3. 等待数据库初始化完成（约30秒）

# 4. 在新终端启动后端应用
make dev-app

# 5. 在另一个新终端启动前端
make dev-frontend
```

### 日常开发

```bash
# 启动开发环境
make dev-start    # 启动 PostgreSQL + Redis [+ 可选服务]

# 启动后端应用（Air 热重载）
make dev-app      # 检测 Go 代码修改后自动重启

# 启动前端开发服务器
make dev-frontend # 运行在 http://localhost:5173

# 查看服务状态
make dev-status

# 查看容器日志
make dev-logs

# 停止服务
make dev-stop
```

### 可选服务 Profile

```bash
# 启动 MinIO 对象存储
./scripts/dev.sh start --minio

# 启动 Qdrant 向量数据库
./scripts/dev.sh start --qdrant

# 启动 Neo4j 图数据库
./scripts/dev.sh start --neo4j

# 启动 Jaeger 链路追踪
./scripts/dev.sh start --jaeger

# 启动所有可选服务
./scripts/dev.sh start --full
```

### 代码修改后

```bash
# 后端：Air 自动检测并重新编译

# 手动编译检查
make build              # 输出到 bin/es-apiserver

# 前端：Vite 自动热更新

# 前端手动构建
cd frontend && npm run build
```

### Proto 文件修改

```bash
# 修改 .proto 文件后，重新生成代码
make gen.protoc

# 验证编译
go build ./pkg/api/apiserver/...
```

### 数据库连接信息

| 配置项 | 默认值 |
|--------|--------|
| 地址 | 127.0.0.1:5432 |
| 用户 | einoshow |
| 密码 | einoshow1234 |
| 数据库 | einoshow |

### 服务端口

| 服务 | 端口 |
|------|------|
| 后端 HTTP API | 5555 |
| 后端 gRPC | 6666 |
| 前端开发服务器 | 5173 |
| PostgreSQL | 5432 |
| Redis | 6379 |
| MinIO Console | 9001 |
| Qdrant Dashboard | 6333 |
| Neo4j Browser | 7474 |
| Jaeger UI | 16686 |

---

## 构建规范

### 后端构建

```bash
# 输出统一到 bin/ 目录
make build              # 输出到 bin/es-apiserver

# 清理
make clean              # 删除 bin/ 目录下的可执行文件
```

### 前端构建

```bash
# 开发模式
cd frontend && npm run dev

# 生产构建
cd frontend && npm run build

# 类型检查
cd frontend && npm run type-check

# 预览构建结果
cd frontend && npm run preview
```

---

## 目录结构规范

### 包命名规范

| 目录 | 包名 | 说明 |
|------|------|------|
| `internal/apiserver/model` | `model` | 数据模型 |
| `internal/apiserver/store` | `store` | 数据访问层 |
| `internal/apiserver/biz/agent` | `agent` | Agent 业务逻辑 |
| `internal/apiserver/biz/session` | `session` | Session 业务逻辑 |
| `internal/apiserver/biz/knowledge` | `knowledge` | Knowledge 业务逻辑 |
| `internal/apiserver/biz/user` | `user` | User 业务逻辑 |
| `internal/apiserver/biz/tenant` | `tenant` | Tenant 业务逻辑 |
| `internal/apiserver/biz/mcp` | `mcp` | MCP 业务逻辑 |
| `internal/apiserver/biz/model` | `model` | Model 业务逻辑 |
| `internal/apiserver/handler/http` | `http` | HTTP 处理器 |
| `internal/apiserver/handler/grpc` | `grpc` | gRPC 处理器 |
| `internal/apiserver/pkg/...` | 按目录名 | apiserver 内部公共包 |
| `internal/pkg/...` | 按目录名 | 跨服务的通用工具包 |
| `pkg/...` | 按目录名 | 对外暴露的公共包 |

**目录结构原则**：
- **单个微服务模块** 放在 `internal/{module}/`，如 `internal/apiserver/`
- **跨服务的业务工具包** 放在 `internal/pkg/{package}/`，如 `internal/pkg/agent/`、`internal/pkg/mcp/`
- **通用工具函数** 放在 `pkg/{package}/`，如 `pkg/server/`、`pkg/db/`

---

## 类型命名规范

### 后端命名

| 后缀/前缀 | 用途 | 示例 |
|----------|------|------|
| `M` | GORM 模型 | `SessionM`, `UserM` |
| `Store` | Store 接口 | `SessionStore`, `UserStore` |
| `Biz` | Biz 接口 | `SessionBiz`, `AgentBiz` |
| `Handler` | Handler 结构体 | `Handler` |
| `I` 前缀 | 顶层接口 | `IBiz`, `IStore` |
| `.gen.go` | 生成文件 | `session.gen.go` |
| `_gen.go` | Wire 生成 | `wire_gen.go` |

### 小写实现类型

```go
// 接口
type SessionStore interface { ... }

// 实现
type sessionStore struct { ... }
```

---

## Model 层规范

### 代码生成原则

**Model 层必须通过 `cmd/gen-gorm-model` 工具生成，禁止手动编写或修改 `.gen.go` 文件。**

### 生成工具使用

```bash
# 1. 确保 PostgreSQL 数据库已启动并包含目标表
make dev-status

# 2. 运行代码生成工具
go run cmd/gen-gorm-model/gen_gorm_model.go

# 3. 指定数据库参数（如需要）
go run cmd/gen-gorm-model/gen_gorm_model.go \
  --db-type postgresql \
  --addr "127.0.0.1:5432" \
  --username "einoshow" \
  --password "einoshow1234" \
  --db "einoshow"
```

### 支持的表配置

| 表名 | 模型名 | 说明 |
|------|--------|------|
| `users` | `UserM` | 用户 |
| `tenants` | `TenantM` | 租户 |
| `custom_agents` | `CustomAgentM` | 自定义 Agent |
| `mcp_services` | `MCPServiceM` | MCP 服务 |
| `sessions` | `SessionM` | 会话 |
| `session_items` | `SessionItemM` | 会话项 |
| `messages` | `MessageM` | 消息 |
| `knowledge_bases` | `KnowledgeBaseM` | 知识库 |
| `knowledges` | `KnowledgeM` | 知识项 |
| `chunks` | `ChunkM` | 知识分块 |
| `knowledge_tags` | `KnowledgeTagM` | 知识标签 |
| `models` | `LLMModelM` | LLM 模型配置 |

---

## Store 层规范

### 接口定义模板

```go
// SessionStore 定义了 session 模块在 store 层所实现的方法.
type SessionStore interface {
    Create(ctx context.Context, obj *model.SessionM) error
    Update(ctx context.Context, obj *model.SessionM) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*model.SessionM, error)
    List(ctx context.Context, opts *where.Options) (int64, []*model.SessionM, error)

    SessionExpansion  // 扩展方法
}

// SessionExpansion 定义了会话操作的附加方法.
type SessionExpansion interface {
    GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.SessionM, error)
}
```

---

## Biz 层规范

### 模块化结构

```
internal/apiserver/biz/
├── biz.go              # IBiz 接口聚合
├── agent/              # Agent 业务逻辑
│   └── agent.go
├── session/            # Session 业务逻辑
│   ├── session.go
│   └── context.go
├── knowledge/          # 知识库业务逻辑
│   ├── knowledge.go
│   ├── document.go
│   └── search.go
├── user/               # 用户业务逻辑
├── tenant/             # 租户业务逻辑
├── mcp/                # MCP 业务逻辑
└── model/              # 模型业务逻辑
```

### 接口定义模板

```go
// SessionBiz 定义了会话业务逻辑接口.
type SessionBiz interface {
    Create(ctx context.Context, req *CreateSessionRequest) (*SessionResponse, error)
    Get(ctx context.Context, id string) (*SessionResponse, error)
}

type biz struct {
    store store.IStore
    agentFactory agent.Factory
}

var _ SessionBiz = (*biz)(nil)

func New(store store.IStore, factory agent.Factory) SessionBiz {
    return &biz{store: store, agentFactory: factory}
}
```

---

## Handler 层规范

### HTTP Handler

```go
// CreateSession 创建会话.
func (h *Handler) CreateSession(c *gin.Context) {
    core.HandleJSONRequest(c, h.biz.Session().Create, h.val.ValidateCreateSessionRequest)
}

// StreamQA 流式问答.
func (h *Handler) StreamQA(c *gin.Context) {
    // 自定义 SSE 处理
}
```

### 路由注册

在 `internal/apiserver/httpserver.go` 中注册：
```go
func (s *httpServer) register() {
    v1 := s.engine.Group("/api/v1")
    {
        sessions := v1.Group("/sessions")
        {
            sessions.POST("", h.CreateSession)
            sessions.GET("", h.ListSessions)
            sessions.GET("/:id", h.GetSession)
            sessions.POST("/:id/qa", h.StreamQA)
        }
    }
}
```

---

## Wire 依赖注入规范

### ProviderSet 定义

```go
// store/store.go
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*datastore)))

// store/session.go
var ProviderSet = wire.NewSet(NewSessionStore)
```

### Wire 聚合

在 `internal/apiserver/wire.go` 中聚合：
```go
//go:generate wire

var ProviderSet = wire.NewSet(
    store.ProviderSet,
    biz.ProviderSet,
    http.ProviderSet,
)
```

---

## Agent 开发规范

### 架构原则

**ADR-001: 直接使用 Eino 框架**
- ✅ 直接使用 Eino 的 `Agent` 和 `Tool` 接口
- ❌ 不创建额外的抽象层
- ✅ 仅在业务层（Biz）做必要的适配
- ✅ SSE 事件格式在 HTTP 层处理

### 目录结构

```
internal/pkg/agent/
├── react/              # ReAct Agent 实现
│   └── agent.go        # 实现 eino.Agent
├── chat/               # Chat Agent 实现
│   └── agent.go
├── tool/               # 工具实现
│   ├── registry.go     # 工具注册
│   ├── knowledge_search.go
│   ├── grep_chunks.go
│   ├── web_search.go
│   ├── web_fetch.go
│   ├── mcp.go
│   ├── think.go
│   └── todo.go
├── model/              # 模型封装
│   ├── chat.go         # ChatModel 工厂
│   └── embedding.go    # EmbeddingModel 工厂
├── factory.go          # Agent 工厂
└── chat.go             # ChatAgent
```

### 内置工具

| 工具 | 功能 | 文件 |
|------|------|------|
| `knowledge_search` | 语义/向量搜索知识库 | `tool/knowledge_search.go` |
| `grep_chunks` | 关键词搜索分块 | `tool/grep_chunks.go` |
| `web_search` | 网络搜索 | `tool/web_search.go` |
| `web_fetch` | 网页内容获取 | `tool/web_fetch.go` |
| `mcp` | MCP 工具调用 | `tool/mcp.go` |
| `think` | 思考工具 | `tool/think.go` |
| `todo` | 待办事项管理 | `tool/todo.go` |

### 提示词语言规范

**核心原则**: 所有传给 LLM 的内容（Agent Name、Tool Name、Description、System Prompt）**必须使用英文**。

| 用途 | 语言 | 示例 | 说明 |
|------|------|------|------|
| **传给 LLM** | 英文 | `knowledge_search` | Tool.Info() 返回的 Name、Desc |
| **前端显示** | 中文 | `语义搜索` | API 返回的 label、description |

---

## MCP 开发规范

### 目录结构

```
internal/pkg/mcp/
├── client.go           # MCP 客户端接口
├── http_client.go      # HTTP 传输实现
├── stdio_client.go     # Stdio 传输实现
└── types.go            # MCP 类型定义
```

### Biz 层集成

```
internal/apiserver/biz/mcp/
└── mcp.go              # MCP 服务管理业务逻辑
```

---

## 前端开发规范

### 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建**: Vite
- **组件库**: TDesign Vue Next
- **状态管理**: Pinia
- **路由**: Vue Router
- **国际化**: Vue i18n
- **HTTP 客户端**: Axios
- **Markdown**: Marked + DOMPurify
- **代码高亮**: Highlight.js

### 目录结构

```
frontend/
├── src/
│   ├── api/            # API 接口定义
│   ├── assets/         # 静态资源
│   ├── components/     # 公共组件
│   ├── hooks/          # 组合式函数
│   ├── i18n/           # 国际化配置
│   ├── router/         # 路由配置
│   ├── stores/         # Pinia 状态管理
│   ├── types/          # TypeScript 类型定义
│   ├── utils/          # 工具函数
│   ├── views/          # 页面视图
│   ├── App.vue         # 根组件
│   └── main.ts         # 应用入口
├── public/             # 公共静态资源
├── index.html          # HTML 模板
├── vite.config.ts      # Vite 配置
├── tsconfig.json       # TypeScript 配置
└── package.json        # 依赖配置
```

### 组件命名规范

- **公共组件**: PascalCase（如 `AgentAvatar.vue`）
- **页面组件**: PascalCase（如 `SessionList.vue`）
- **工具函数**: camelCase（如 `formatDate.ts`）
- **类型定义**: PascalCase（如 `Session.ts`）
- ** API 接口**: camelCase（如 `sessionApi.ts`）

### 状态管理规范

```
stores/
├── index.ts            # Pinia 实例
├── modules/
│   ├── user.ts         # 用户状态
│   ├── session.ts      # 会话状态
│   ├── agent.ts        # Agent 状态
│   └── knowledge.ts    # 知识库状态
```

### API 接口规范

```typescript
// src/api/session.ts
import request from './request'

export const sessionApi = {
  list: (params?: ListParams) => request.get<Session[]>('/api/v1/sessions', { params }),
  get: (id: string) => request.get<Session>(`/api/v1/sessions/${id}`),
  create: (data: CreateSessionRequest) => request.post<Session>('/api/v1/sessions', data),
  update: (id: string, data: UpdateSessionRequest) => request.put<Session>(`/api/v1/sessions/${id}`, data),
  delete: (id: string) => request.delete(`/api/v1/sessions/${id}`),
}
```

---

## 导入顺序规范

### Go

```go
import (
    // 1. 标准库
    "context"
    "time"

    // 2. Eino 框架
    "github.com/cloudwego/eino/components/agent"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"

    // 3. 项目内部
    "github.com/ashwinyue/eino-show/internal/apiserver/model"
    "github.com/ashwinyue/eino-show/internal/pkg/agent"

    // 4. 其他第三方库
    "github.com/gin-gonic/gin"
    "github.com/google/wire"
    "gorm.io/gorm"
)
```

### TypeScript

```typescript
// 1. Vue 相关
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

// 2. 项目内部
import { sessionApi } from '@/api/session'
import { useSessionStore } from '@/stores/session'

// 3. 第三方库
import axios from 'axios'
import { MessagePlugin } from 'tdesign-vue-next'
```

---

## 新模块开发 Checklist

### 后端模块（以 Example 为例）

**数据层**:
- [ ] 创建数据库表
- [ ] 在 `cmd/gen-gorm-model/gen_gorm_model.go` 添加生成配置
- [ ] 运行生成 Model
- [ ] `store/example.go` - 定义 ExampleStore 接口和实现
- [ ] `store/store.go` - 在 IStore 中添加 Example() 方法

**业务层**:
- [ ] `biz/example/example.go` - 定义 ExampleBiz 接口和实现
- [ ] `biz/biz.go` - 在 IBiz 中添加 Example() 方法

**HTTP 层**:
- [ ] `handler/http/example.go` - 定义 HTTP 处理函数
- [ ] `httpserver.go` - 注册路由
- [ ] `pkg/validation/example.go` - 请求验证

**Wire**:
- [ ] 添加 ProviderSet

### 前端模块

**API 层**:
- [ ] `src/api/example.ts` - API 接口定义

**类型定义**:
- [ ] `src/types/example.ts` - TypeScript 类型

**状态管理**:
- [ ] `src/stores/modules/example.ts` - Pinia store

**视图组件**:
- [ ] `src/views/example/` - 页面组件

**路由配置**:
- [ ] `src/router/index.ts` - 添加路由

---

## 注释规范

### 包注释

```go
// Package store 提供数据访问层实现，封装了与数据库的交互逻辑.
package store
```

### 函数注释

```go
// Create 创建一条会话记录.
// ctx: 请求上下文
// obj: 要创建的会话对象
// 返回: 错误信息
func (s *sessionStore) Create(ctx context.Context, obj *model.SessionM) error {
    return s.DB(ctx).Create(obj).Error
}
```

### 接口注释

```go
// SessionStore 定义了会话模块在 store 层所实现的方法.
// 提供会话的 CRUD 操作以及按租户查询等扩展功能.
type SessionStore interface {
    // ...
}
```

---

## 错误处理规范

### 返回错误

```go
import "github.com/onexstack/onexstack/pkg/errors"

if err != nil {
    return errors.WithCode(err, codes.ErrDatabase, "failed to create session")
}
```

### 日志记录

```go
import "go.uber.org/zap"

if err != nil {
    log.L(ctx).Error("failed to create session", zap.Error(err))
    return err
}
```

---

## 开发参考规范

### 必须参考的实现

开发过程中，**必须**参考 `a-old/` 目录中的 Eino 最佳实践实现：

| 参考目录 | 说明 | 关键内容 |
|---------|------|----------|
| `a-old/old/eino-examples/` | Eino 官方示例 | ReactAgent、Tool、数据流模式 |
| `a-old/old/eino-examples/adk/` | Eino ADK 示例 | Agent 开发最佳实践 |
| `a-old/old/eino-examples/compose/graph/` | Graph 编排模式 | tool_call_agent、react_with_interrupt |
| `a-old/WeKnora/` | 旧项目实现 | 业务逻辑、数据模型、API 设计 |

### 开发前检查清单

1. **搜索相关示例**
   ```bash
   # 查找 Agent 相关示例
   find a-old/old/eino-examples -name "*.go" | xargs grep -l "Agent"

   # 查找 Tool 相关示例
   find a-old/old/eino-examples -name "*.go" | xargs grep -l "Tool"
   ```

2. **阅读示例代码**
   - 理解数据流向
   - 复用错误处理模式
   - 参考接口设计

3. **对比旧实现**
   - 查看 `a-old/WeKnora/` 中对应功能
   - 确保接口兼容性
   - 迁移核心业务逻辑

### 禁止事项

- ❌ 盲目实现，不参考官方示例
- ❌ 直接复制旧代码，不理解逻辑
- ❌ 忽略 Eino 最佳实践模式
