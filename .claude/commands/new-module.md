---
name: new-module
description: Create a new business module following miniblog-x four-layer architecture. Generates Model → Store → Biz → Handler files with proper structure, interfaces, and Wire integration for eino-show project.
---

创建新业务模块骨架（四层架构）。

## 使用方法

```
/new-module <模块名>
```

## 示例

```
/new-module Session
/new-module Agent
/new-module Knowledge
```

## 生成的文件结构

```
internal/apiserver/
├── model/<module>.gen.go          # 数据模型
├── store/<module>.go              # Store 接口和实现
├── biz/v1/<module>/<module>.go    # Biz 接口和实现
└── handler/http/<module>.go       # HTTP 处理器
```

## 生成的代码模板

### Model 层
```go
const TableName<Module>M = "<table>"

type <Module>M struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `gorm:"not null;default:current_timestamp" json:"created_at"`
    UpdatedAt time.Time `gorm:"not null;default:current_timestamp" json:"updated_at"`
}

func (*<Module>M) TableName() string {
    return TableName<Module>M
}
```

### Store 层
```go
type <Module>Store interface {
    Create(ctx context.Context, obj *model.<Module>M) error
    Get(ctx context.Context, opts *where.Options) (*model.<Module>M, error)
    List(ctx context.Context, opts *where.Options) (int64, []*model.<Module>M, error)
    Update(ctx context.Context, obj *model.<Module>M) error
    Delete(ctx context.Context, opts *where.Options) error
}
```

### Biz 层
```go
type <Module>Biz interface {
    Create(ctx context.Context, req *Create<Module>Request) (*<Module>Response, error)
    Get(ctx context.Context, id string) (*<Module>Response, error)
    List(ctx context.Context, req *List<Module>Request) (*List<Module>Response, error)
    Update(ctx context.Context, id string, req *Update<Module>Request) (*<Module>Response, error)
    Delete(ctx context.Context, id string) error
}
```

### Handler 层
```go
func (h *Handler) Create<Module>(c *gin.Context) {
    core.HandleJSONRequest(c, h.biz.<Module>().Create, h.val.ValidateCreate<Module>Request)
}
```

## 后续步骤

1. 在 `store/store.go` 的 `IStore` 接口中添加 `<Module>()` 方法
2. 在 `biz/biz.go` 的 `IBiz` 接口中添加 `<Module>()` 方法
3. 在 `httpserver.go` 中注册路由
4. 运行 `wire` 生成依赖注入代码
