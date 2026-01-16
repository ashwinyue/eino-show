---
name: miniblog-patterns
description: miniblog-x four-layer architecture patterns. Use this skill when implementing business features following Handler → Biz → Store → Model layers. Understand data flow, dependency rules, and how each layer works. Reference existing code in internal/apiserver/ for patterns: handler/http/, biz/v1/, store/, model/.
---

miniblog-x 四层架构模式 Skill。理解和使用 miniblog-x 脚手架的分层架构。

## 架构概览

```
┌─────────────────────────────────────────────────────┐
│  Handler (HTTP/gRPC)                                 │  ← 请求路由、参数校验、DTO转换
├─────────────────────────────────────────────────────┤
│  Biz (业务逻辑)                                       │  ← 业务逻辑、事务管理、调用 Agent
├─────────────────────────────────────────────────────┤
│  Store (数据访问)                                     │  ← 数据 CRUD、数据库操作
├─────────────────────────────────────────────────────┤
│  Model (GORM)                                        │  ← 数据模型定义
└─────────────────────────────────────────────────────┘
```

## 依赖规则

| 层 | 可依赖 | 禁止依赖 |
|---|--------|----------|
| Handler | Biz | Store, Model, Agent(Eino) |
| Biz | Store, Agent接口 | Model, Eino |
| Store | Model | Biz, Handler |
| Model | 无 | 无 |

## 各层模式

### 1. Handler 层

**职责**：请求路由、参数校验、调用 Biz

```go
// internal/apiserver/handler/http/session.go

// CreateSession 创建会话.
func (h *Handler) CreateSession(c *gin.Context) {
    core.HandleJSONRequest(c, h.biz.Session().Create, h.val.ValidateCreateSessionRequest)
}

// StreamQA 流式问答.
func (h *Handler) StreamQA(c *gin.Context) {
    sessionID := c.Param("id")
    var req QASRequest
    c.ShouldBindJSON(&req)

    eventCh, err := h.biz.Session().StreamQA(c.Request.Context(), sessionID, req.Query)
    if err != nil {
        restful.Error(c, restful.ErrInternalServer, err)
        return
    }

    // SSE 流式输出
    c.Writer.Header().Set("Content-Type", "text/event-stream")
    for event := range eventCh {
        c.SSEvent(event.Type, event.Data)
    }
}
```

### 2. Biz 层

**职责**：业务逻辑、事务管理、调用 Agent 接口

```go
// internal/apiserver/biz/v1/session/session.go

type SessionBiz interface {
    Create(ctx context.Context, req *CreateSessionRequest) (*SessionResponse, error)
    StreamQA(ctx context.Context, sessionID, query string) (<-chan *Event, error)
}

type biz struct {
    store       store.IStore
    agentFactory agent.Factory  // 返回 agent.Agent 接口，不依赖 Eino
}

func (b *biz) Create(ctx context.Context, req *CreateSessionRequest) (*SessionResponse, error) {
    // 1. 参数校验
    // 2. 业务逻辑处理
    // 3. 调用 Store
    sessionM := &model.SessionM{
        ID:        uuid.New().String(),
        Title:     req.Title,
        TenantID:  req.TenantID,
    }
    if err := b.store.Session().Create(ctx, sessionM); err != nil {
        return nil, err
    }

    // 4. DTO 转换
    return toSessionResponse(sessionM), nil
}

func (b *biz) StreamQA(ctx context.Context, sessionID, query string) (<-chan *Event, error) {
    // 调用 Agent 接口（不直接依赖 Eino）
    agent := b.agentFactory.Create(ctx, agentConfig)
    return agent.StreamRun(ctx, &agent.Input{Query: query})
}
```

### 3. Store 层

**职责**：数据 CRUD、数据库操作

```go
// internal/apiserver/store/session.go

type SessionStore interface {
    Create(ctx context.Context, obj *model.SessionM) error
    Update(ctx context.Context, obj *model.SessionM) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*model.SessionM, error)
    List(ctx context.Context, opts *where.Options) (int64, []*model.SessionM, error)

    SessionExpansion  // 扩展方法
}

type SessionExpansion interface {
    GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.SessionM, error)
}

type sessionStore struct {
    *genericstore.Store[model.SessionM]
}

func newSessionStore(store *datastore) *sessionStore {
    return &sessionStore{
        Store: genericstore.NewStore[model.SessionM](store, NewLogger()),
    }
}

func (s *sessionStore) GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.SessionM, error) {
    var list []*model.SessionM
    err := s.DB(ctx).Where("tenant_id = ?", tenantID).Find(&list).Error
    return list, err
}
```

### 4. Model 层

**职责**：数据模型定义

```go
// internal/apiserver/model/session.gen.go

const TableNameSessionM = "sessions"

type SessionM struct {
    ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
    Title     string    `gorm:"type:varchar(255);not null" json:"title"`
    TenantID  uint64    `gorm:"not null;index:idx_tenant" json:"tenant_id"`
    CreatedAt time.Time `gorm:"not null;default:current_timestamp" json:"created_at"`
    UpdatedAt time.Time `gorm:"not null;default:current_timestamp" json:"updated_at"`
}

func (*SessionM) TableName() string {
    return TableNameSessionM
}
```

## 新模块开发流程

添加新业务模块时，按以下顺序创建：

```
1. model/xxx.gen.go        → 定义数据模型
2. store/xxx.go            → 定义 Store 接口和实现
3. store/store.go          → 在 IStore 中添加 Xxx() 方法
4. biz/v1/xxx/xxx.go       → 定义 Biz 接口和实现
5. biz/biz.go              → 在 IBiz 中添加 Xxx() 方法
6. handler/http/xxx.go     → 定义 HTTP 处理函数
7. httpserver.go           → 注册路由
```

## Wire 依赖注入

```go
// internal/apiserver/wire.go
func InitializeWebServer(*Config) (server.Server, error) {
    wire.Build(
        wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
        wire.Struct(new(ServerConfig), "*"),
        wire.NewSet(store.ProviderSet, biz.ProviderSet),
        ProvideDB,
        validation.ProviderSet,
    )
    return nil, nil
}
```

## 参考示例

| 功能 | 参考文件 |
|------|----------|
| 用户 CRUD | `handler/http/user.go`, `biz/v1/user/`, `store/user.go` |
| 博客 CRUD | `handler/http/post.go`, `biz/v1/post/`, `store/post.go` |
| 健康检查 | `handler/http/healthz.go` |

## 关键原则

1. **单向依赖**：Handler → Biz → Store → Model
2. **接口解耦**：Biz 层通过接口依赖 Agent，不直接依赖 Eino
3. **DTO 转换**：Handler 层负责 Request/Model 转换
4. **事务管理**：在 Biz 层通过 `store.TX()` 管理
