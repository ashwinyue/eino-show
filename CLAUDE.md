# eino-show 项目指南

> 基于 miniblog-x 脚手架，集成 Eino ADK 实现 Agent 系统

## 项目架构

```
┌─────────────────────────────────────────────────────┐
│  Frontend (Vue 3 + TypeScript)                      │  ← frontend/
├─────────────────────────────────────────────────────┤
│  Handler (HTTP/gRPC)                                 │  ← internal/apiserver/handler/
├─────────────────────────────────────────────────────┤
│  Biz (业务逻辑)                                       │  ← internal/apiserver/biz/
│    ├─ agent/     ├─ session/  ├─ knowledge/         │
│    ├─ user/      ├─ tenant/   ├─ mcp/               │
│    └─ model/     (按模块组织)                        │
├─────────────────────────────────────────────────────┤
│  Store (数据访问)                                     │  ← internal/apiserver/store/
├─────────────────────────────────────────────────────┤
│  Model (GORM)                                        │  ← internal/apiserver/model/
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  内部基础设施 (internal/pkg/)                         │
│  ├─ agent/      Agent 抽象与实现                     │
│  ├─ mcp/        Model Context Protocol               │
│  ├─ document/   文档处理                             │
│  └─ retriever/  向量检索                             │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  公共包 (pkg/)                                       │
│  ├─ api/        API 定义 (Protobuf)                  │
│  ├─ server/     服务器抽象                           │
│  ├─ db/         数据库封装                           │
│  ├─ authn/      认证                                 │
│  ├─ authz/      授权                                 │
│  └─ ...         其他通用组件                         │
└─────────────────────────────────────────────────────┘
```

## 目录结构

```
eino-show/
├── cmd/                    # 命令行程序入口
│   ├── mb-apiserver/       # 主服务器
│   └── gen-gorm-model/     # GORM 模型生成器
│
├── internal/               # 内部代码（不对外暴露）
│   ├── apiserver/          # API 服务器核心实现
│   │   ├── handler/        # HTTP/gRPC 处理器
│   │   ├── biz/            # 业务逻辑（按模块组织）
│   │   ├── store/          # 数据访问层
│   │   ├── model/          # GORM 数据模型
│   │   └── pkg/            # 内部工具包
│   └── pkg/                # 跨服务通用包
│       ├── agent/          # Agent 抽象与实现
│       ├── mcp/            # MCP 协议实现
│       ├── document/       # 文档处理
│       └── retriever/      # 向量检索
│
├── pkg/                    # 公共包（可被外部引用）
│   ├── api/                # Protobuf API 定义
│   ├── server/             # 服务器抽象
│   ├── db/                 # 数据库封装
│   ├── authn/              # 认证
│   ├── authz/              # 授权
│   └── ...                 # 其他通用组件
│
├── frontend/               # Vue 3 前端项目
│   ├── src/
│   │   ├── api/            # API 接口
│   │   ├── components/     # 组件
│   │   ├── views/          # 页面视图
│   │   └── stores/         # 状态管理
│   └── package.json
│
├── api/                    # API 规范（OpenAPI）
├── configs/                # 配置文件
├── scripts/                # 脚本工具
├── docs/                   # 项目文档
└── a-old/                  # 旧代码参考（WeKnora、Eino 示例）
```

## 代码规范

### 分层架构

| 层 | 职责 | 目录 | 依赖 |
|---|------|------|------|
| Handler | 请求路由、参数校验、DTO转换 | `handler/http/`, `handler/grpc/` | → Biz |
| Biz | 业务逻辑、事务管理 | `biz/*模块*/` | → Store + Agent接口 |
| Store | 数据 CRUD | `store/` | → Model |
| Model | 数据模型定义 | `model/` | 无 |

### 模块化 Biz 层

业务层按领域模块组织，每个模块独立包：

```
internal/apiserver/biz/
├── biz.go              # IBiz 接口聚合
├── agent/              # Agent 业务逻辑
│   └── agent.go
├── session/            # Session 业务逻辑
│   ├── session.go
│   └── context.go
├── knowledge/          # 知识库业务逻辑
│   ├── kb.go
│   ├── document.go
│   └── search.go
├── user/               # 用户业务逻辑
├── tenant/             # 租户业务逻辑
├── mcp/                # MCP 业务逻辑
└── model/              # 模型业务逻辑
```

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| Model | `*M` 后缀 | `SessionM`, `UserM` |
| Store 接口 | `*Store` 后缀 | `SessionStore`, `UserStore` |
| Biz 接口 | `*Biz` 后缀 | `SessionBiz`, `AgentBiz` |
| 顶层接口 | `I` 前缀 | `IBiz`, `IStore` |
| 生成文件 | `.gen.go` 后缀 | `session.gen.go` |
| Wire 生成 | `_gen.go` 后缀 | `wire_gen.go` |
| 实现类型 | 小写 | `type sessionStore struct{}` |

### 导入顺序

```go
import (
    // 1. 标准库
    "context"
    "time"

    // 2. Eino 框架
    "github.com/cloudwego/eino/components/agent"
    "github.com/cloudwego/eino/components/tool"

    // 3. 项目内部
    "github.com/ashwinyue/eino-show/internal/apiserver/model"
    "github.com/ashwinyue/eino-show/internal/pkg/agent"

    // 4. 其他第三方库
    "github.com/gin-gonic/gin"
    "github.com/google/wire"
)
```

## 重构阶段

| Phase | 状态 | 说明 |
|-------|------|------|
| Phase 1 | ✅ | 基础设施 (Wire/Gin/GORM/日志) |
| Phase 2 | ✅ | 数据层: Session/Agent/Knowledge Model + Store |
| Phase 3 | ✅ | Eino Agent 实现 |
| Phase 4 | ✅ | 业务层: Biz 实现 |
| Phase 5 | ✅ | HTTP 层: Handler + SSE |
| Phase 6 | ⏳ | 前端集成与测试优化 |

详细计划: `docs/plan/重构计划.md`

## 关键约束

1. **接口兼容性** - 保持前端无需修改
   - API 路径不变: `/api/v1/sessions`, `/api/v1/custom-agents` 等
   - SSE 事件格式不变: `agent_thinking`, `agent_action`, `agent_observation`, `agent_complete`

2. **表结构兼容** - 支持平滑迁移
   - sessions, session_items, messages
   - custom_agents, knowledge_bases, knowledges, chunks
   - users, tenants

3. **Agent 解耦** - Biz 层依赖 `internal/pkg/agent/` 接口，不直接依赖 Eino

## 常用命令

### 开发环境

```bash
# 启动基础设施 (PostgreSQL + Redis)
make dev-start

# 启动后端应用（热重载）
make dev-app

# 启动前端开发服务器
make dev-frontend

# 查看服务状态
make dev-status

# 查看日志
make dev-logs

# 停止服务
make dev-stop
```

### 构建相关

```bash
# 编译后端
make build              # 输出到 bin/es-apiserver

# 快速编译检查
go build ./cmd/mb-apiserver/...
go build ./internal/apiserver/...
```

### 代码生成

```bash
# Wire 依赖注入
make wire

# Proto 代码生成
make gen.protoc

# GORM 模型生成
go run cmd/gen-gorm-model/gen_gorm_model.go
```

### 代码质量

```bash
make test               # 运行测试
make lint               # 代码检查
make fmt                # 格式化代码
```

## 编译规则

- **后端输出目录**: `bin/`
- **前端开发端口**: `http://localhost:5173`
- **API 服务端口**: `5555` (HTTP), `6666` (gRPC)

## 当前任务

参考 `docs/plan/重构计划.md` 中 Phase 6 任务清单。

## 开发参考

⚠️ **开发前必须参考 `a-old/` 目录中的 Eino 最佳实践**

| 参考目录 | 用途 |
|---------|------|
| `a-old/old/eino-examples/` | Eino 官方示例 (ReactAgent、Tool) |
| `a-old/old/eino-examples/adk/` | Eino ADK 示例 |
| `a-old/WeKnora/` | 旧项目实现 (业务逻辑、API 设计) |

详见: `CONVENTION.md` → 开发参考规范
