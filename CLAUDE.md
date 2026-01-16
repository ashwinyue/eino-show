# eino-show 项目指南

> 基于 miniblog-x 脚手架，集成 Eino ADK 实现 Agent 系统

## 项目架构

```
┌─────────────────────────────────────────────────────┐
│  Handler (HTTP/gRPC)                                 │  ← internal/apiserver/handler/
├─────────────────────────────────────────────────────┤
│  Biz (业务逻辑)                                       │  ← internal/apiserver/biz/
├─────────────────────────────────────────────────────┤
│  Store (数据访问)                                     │  ← internal/apiserver/store/
├─────────────────────────────────────────────────────┤
│  Model (GORM)                                        │  ← internal/apiserver/model/
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  Agent 接口 (解耦 Eino)                               │  ← internal/pkg/agent/
├─────────────────────────────────────────────────────┤
│  Agent 实现 (依赖 Eino)                               │  ← internal/agent/
└─────────────────────────────────────────────────────┘
```

## 代码规范

### 依赖注入
- 使用 Wire 进行依赖注入
- 构造函数命名：`New{TypeName}`
- ProviderSet 声明在各自包中，统一在 `wire.go` 聚合

### 分层规则
| 层 | 职责 | 依赖 |
|---|------|------|
| Handler | 请求路由、参数校验、DTO转换 | → Biz |
| Biz | 业务逻辑、事务管理 | → Store + Agent接口 |
| Store | 数据 CRUD | → Model |
| Model | 数据模型定义 | 无 |

### 命名规范
- Model 后缀: `*M` (如 `SessionM`)
- 接口: `I{名}` (如 `IBiz`, `IStore`)
- 生成文件: `*.gen.go`
- Wire 生成: `wire_gen.go`

### 导入顺序
```go
import (
    // 标准库
    "context"

    // 项目内部
    "github.com/ashwinyue/eino-show/internal/..."

    // 第三方库
    "github.com/gin-gonic/gin"
    "github.com/google/wire"
)
```

## 重构阶段

| Phase | 状态 | 说明 |
|-------|------|------|
| Phase 1 | ✅ | 基础设施 (Wire/Gin/GORM/日志) |
| Phase 2 | 🚧 | 数据层: Session/Agent/Knowledge Model + Store |
| Phase 3 | ⏳ | Agent 抽象层: internal/pkg/agent/ |
| Phase 4 | ⏳ | Eino Agent 实现: internal/agent/ |
| Phase 5 | ⏳ | 业务层: Biz 实现 |
| Phase 6 | ⏳ | HTTP 层: Handler + SSE |
| Phase 7 | ⏳ | 测试与优化 |

详细计划: `docs/plan/重构计划.md`

## 关键约束

1. **接口兼容性** - 保持前端无需修改
   - API 路径不变: `/api/v1/sessions`, `/api/v1/custom-agents` 等
   - SSE 事件格式不变: `agent_thinking`, `agent_action`, `agent_observation`, `agent_complete`

2. **表结构兼容** - 支持平滑迁移
   - sessions, session_items, messages
   - custom_agents, knowledge_bases, knowledges, chunks
   - users, tenants

3. **Agent 解耦** - Biz 层依赖 `internal/pkg/agent/Agent` 接口，不直接依赖 Eino

## 常用命令

```bash
make build BINS=mb-apiserver   # 编译
make test                       # 迋试
make lint                       # 代码检查
wire ./internal/apiserver       # 生成依赖注入代码
```

## 当前任务

参考 `docs/plan/重构计划.md` 中 Phase 2 任务清单。

## 开发参考

⚠️ **开发前必须参考 `a-old/` 目录中的 Eino 最佳实践**

| 参考目录 | 用途 |
|---------|------|
| `a-old/old/eino-examples/` | Eino 官方示例 (ReactAgent、Tool) |
| `a-old/old/eino-examples/adk/` | Eino ADK 示例 |
| `a-old/WeKnora/` | 旧项目实现 (业务逻辑、API 设计) |

详见: `CONVENTION.md` → 开发参考规范
