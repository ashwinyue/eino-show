# 工程化最佳实践参考

> 分析来源: WeKnora、Coze Studio、Docker cagent
>
> 分析日期: 2026-01-17

## 目录

1. [错误处理](#1-错误处理)
2. [日志系统](#2-日志系统)
3. [配置管理](#3-配置管理)
4. [测试策略](#4-测试策略)
5. [API 设计](#5-api-设计)
6. [数据库实践](#6-数据库实践)
7. [缓存策略](#7-缓存策略)
8. [中间件设计](#8-中间件设计)
9. [构建系统](#9-构建系统)
10. [监控追踪](#10-监控追踪)

---

## 1. 错误处理

### 来源: WeKnora

**核心设计**: 统一错误码 + 结构化错误类型

```go
// internal/errors/errors.go

// 错误码按业务域划分
const (
    // 通用错误 (1000-1999)
    ErrBadRequest    ErrorCode = 1000
    ErrUnauthorized  ErrorCode = 1001
    ErrForbidden     ErrorCode = 1002
    ErrNotFound      ErrorCode = 1003

    // 租户相关 (2000-2999)
    ErrTenantNotFound ErrorCode = 2000

    // Agent 相关 (2100-2199)
    ErrAgentInvalidParams ErrorCode = 2100
)

// 结构化错误类型
type AppError struct {
    Code     ErrorCode `json:"code"`    // 业务错误码
    Message  string    `json:"message"` // 错误信息
    Details  any       `json:"details"` // 错误详情
    HTTPCode int       `json:"-"`       // HTTP状态码
    Err      error     `json:"-"`       // 原始错误
}

// 便捷构造函数
func NewBadRequestError(message string) *AppError {
    return &AppError{
        Code:     ErrBadRequest,
        Message:  message,
        HTTPCode: http.StatusBadRequest,
    }
}

// 链式添加详情
func (e *AppError) WithDetails(details any) *AppError {
    e.Details = details
    return e
}

// 错误判断
func IsAppError(err error) (*AppError, bool) {
    if appErr, ok := err.(*AppError); ok {
        return appErr, true
    }
    return nil, false
}
```

### 统一错误响应

```go
// HTTP 响应格式
{
    "success": false,
    "error": {
        "code": 2100,
        "message": "Agent 参数无效",
        "details": {
            "field": "temperature",
            "reason": "必须在 0-1 之间"
        }
    }
}
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **错误码分层** | 按业务域划分区间，便于维护 |
| **统一格式** | 所有错误返回相同 JSON 结构 |
| **链式处理** | 支持 `WithDetails()` 添加上下文 |
| **HTTP 映射** | 自动映射错误码到 HTTP 状态码 |
| **原始错误保留** | 保留 `Err` 字段用于日志 |

---

## 2. 日志系统

### 来源: WeKnora (Logrus) + cagent (slog)

### 2.1 结构化日志 (WeKnora)

```go
// internal/logger/logger.go

import "github.com/sirupsen/logrus"

// 日志级别
type LogLevel string
const (
    LevelDebug LogLevel = "debug"
    LevelInfo  LogLevel = "info"
    LevelWarn  LogLevel = "warn"
    LevelError LogLevel = "error"
    LevelFatal LogLevel = "fatal"
)

// 上下文日志
func WithRequestID(c context.Context, requestID string) context.Context {
    return WithField(c, "request_id", requestID)
}

func WithTenantID(c context.Context, tenantID string) context.Context {
    return WithField(c, "tenant_id", tenantID)
}

// 便捷函数（自动记录调用位置）
func Errorf(c context.Context, format string, args ...interface{}) {
    addCaller(GetLogger(c), 2).Errorf(format, args...)
}

func Infof(c context.Context, format string, args ...interface{}) {
    addCaller(GetLogger(c), 2).Infof(format, args...)
}
```

### 2.2 标准库日志 (cagent - 推荐)

```go
// cmd/root/root.go

import "log/slog"

func setupLogging(debug bool, logFile string) error {
    if !debug {
        // 生产环境：静默模式
        slog.SetDefault(slog.New(slog.DiscardHandler))
        return nil
    }

    var writer io.Writer = os.Stdout
    if logFile != "" {
        f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
        if err != nil {
            return err
        }
        writer = f
    }

    // 开发环境：文本格式 + 彩色
    opts := &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }

    slog.SetDefault(slog.New(slog.NewTextHandler(writer, opts)))
    return nil
}
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **结构化日志** | 使用 JSON 格式，便于查询分析 |
| **上下文传递** | 通过 Context 传递 request_id、tenant_id |
| **调用者信息** | 自动记录文件名、行号、函数名 |
| **日志级别** | 支持动态调整 |
| **标准库优先** | 新项目推荐使用 slog（Go 1.21+） |

### eino-show 当前状态

✅ 已有 `pkg/log/` 和 `pkg/logger/`，建议统一到标准库 `slog`

---

## 3. 配置管理

### 来源: WeKnora (Viper) + cagent (配置迁移)

### 3.1 Viper 配置 (WeKnora)

```go
// internal/config/config.go

import "github.com/spf13/viper"

type Config struct {
    Conversation  *ConversationConfig  `yaml:"conversation"`
    Server        *ServerConfig        `yaml:"server"`
    KnowledgeBase *KnowledgeBaseConfig `yaml:"knowledge_base"`
    Models        []ModelConfig        `yaml:"models"`
    Redis         RedisConfig          `yaml:"redis"`
}

func Load(path string) (*Config, error) {
    v := viper.New()

    // 设置配置文件
    v.SetConfigFile(path)
    v.SetConfigType("yaml")

    // 环境变量支持
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 读取配置
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }

    // 解析到结构体
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### 3.2 环境变量替换

```yaml
# config.yaml 支持环境变量引用
database:
  host: ${DB_HOST:localhost}  # 默认 localhost
  port: ${DB_PORT:5432}
  password: ${DB_PASSWORD}     # 必须设置

redis:
  addr: ${REDIS_ADDR:localhost:6379}
```

### 3.3 配置版本迁移 (cagent)

```go
// pkg/config/config.go

type Config struct {
    Version    string
    APIKey     string
    // ...
}

// 支持向后兼容的配置迁移
func Load(ctx context.Context, source Reader) (*latest.Config, error) {
    // 1. 读取原始配置
    raw, err := source.Read()
    if err != nil {
        return nil, err
    }

    // 2. 检测版本并迁移
    migrated, err := migrate.Latest(ctx, raw)
    if err != nil {
        return nil, err
    }

    // 3. 验证配置
    if err := migrated.Validate(); err != nil {
        return nil, err
    }

    return migrated, nil
}
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **多源配置** | 文件 + 环境变量，环境变量优先 |
| **类型安全** | 使用结构体 + MapStructure |
| **默认值** | 配置项支持默认值 |
| **配置验证** | 启动时验证配置完整性 |
| **版本迁移** | 支持配置文件版本升级 |

### eino-show 当前状态

✅ 已有 `configs/mb-apiserver.yaml` + Viper，实现较完善

---

## 4. 测试策略

### 来源: WeKnora + cagent

### 4.1 单元测试

```go
// pkg/client/client_test.go

func TestNewClient(t *testing.T) {
    tests := []struct {
        name    string
        addr    string
        wantErr bool
    }{
        {"valid address", "localhost:50051", false},
        {"invalid address", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client, err := NewClient(tt.addr)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if client != nil {
                defer client.Close()
            }
        })
    }
}
```

### 4.2 集成测试 (cagent VCR 模式)

```go
// e2e/e2e_test.go

func TestE2E(t *testing.T) {
    // 使用 VCR 记录和回放 HTTP 响应
    recorder := vcr.New(t, "fixtures")

    client := NewClient(vcrTransport(recorder))

    // 测试逻辑
    resp, err := client.Do(req)
    assert.NoError(t, err)
    assert.NotNil(t, resp)
}
```

### 4.3 测试覆盖率配置

```yaml
# .golangci.yml

linters:
  enable:
    - errcheck      # 检查未处理的错误
    - staticcheck   # 静态分析
    - unused        # 未使用检查
    - gosec         # 安全检查

tests:
  timeout: 30s
  # 测试时禁止使用 context.Background
  forbid-context-without-timeout: true
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **表驱动测试** | 使用测试表覆盖多种场景 |
| **子测试** | 使用 `t.Run()` 组织测试用例 |
| **资源清理** | 使用 `defer` 确保资源释放 |
| **Mock 外部依赖** | 集成测试使用 VCR 记录响应 |
| **覆盖率检查** | CI 中检查测试覆盖率 |

### eino-show 当前状态

⚠️ 测试覆盖较少，需要补充

---

## 5. API 设计

### 来源: WeKnora

### 5.1 RESTful 路由设计

```go
// routes.go

// RESTful 规范
router.POST   ("/api/v1/auth/register", handler.Register)
router.POST   ("/api/v1/auth/login", handler.Login)
router.GET    ("/api/v1/sessions", handler.ListSessions)
router.POST   ("/api/v1/sessions", handler.CreateSession)
router.GET    ("/api/v1/sessions/:id", handler.GetSession)
router.DELETE ("/api/v1/sessions/:id", handler.DeleteSession)
router.POST   ("/api/v1/sessions/:id/messages", handler.SendMessage)
router.GET    ("/api/v1/agents", handler.ListAgents)
router.POST   ("/api/v1/agents", handler.CreateAgent)
router.GET    ("/api/v1/agents/:id", handler.GetAgent)
router.PUT    ("/api/v1/agents/:id", handler.UpdateAgent)
router.DELETE ("/api/v1/agents/:id", handler.DeleteAgent)
```

### 5.2 统一响应格式

```go
// 成功响应
{
    "success": true,
    "data": {
        "id": "123",
        "name": "test"
    }
}

// 错误响应
{
    "success": false,
    "error": {
        "code": 2100,
        "message": "Agent 参数无效",
        "details": {}
    }
}

// 列表响应（带分页）
{
    "success": true,
    "data": {
        "items": [...],
        "total": 100,
        "page": 1,
        "page_size": 20
    }
}
```

### 5.3 参数验证

```go
import "github.com/go-playground/validator/v10"

type CreateAgentRequest struct {
    Name        string  `json:"name" binding:"required,min=1,max=100"`
    Description string  `json:"description" binding:"max=500"`
    Temperature float32 `json:"temperature" binding:"required,min=0,max=1"`
    MaxTokens   int     `json:"max_tokens" binding:"min=1,max=32000"`
}

func (h *Handler) CreateAgent(c *gin.Context) {
    var req CreateAgentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 自动返回验证错误
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error": gin.H{
                "code":    ErrBadRequest,
                "message": err.Error(),
            },
        })
        return
    }
    // ...
}
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **RESTful 规范** | 遵循 REST 设计原则 |
| **统一响应** | 成功/错误使用统一格式 |
| **参数验证** | 使用 validator 进行声明式验证 |
| **HTTP 状态码** | 正确使用状态码 |
| **API 版本** | 使用 `/api/v1/` 路径版本控制 |

### eino-show 当前状态

✅ 已遵循 RESTful 规范，实现良好

---

## 6. 数据库实践

### 来源: WeKnora

### 6.1 表设计规范

```sql
-- 通用字段规范
CREATE TABLE tenants (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_key VARCHAR(64) NOT NULL,

    -- JSON 字段存储动态配置
    retriever_engines JSON NOT NULL,
    agent_config JSON DEFAULT NULL,

    -- 状态字段
    status VARCHAR(50) DEFAULT 'active',

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,  -- 软删除

    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
);

-- 多租户索引
CREATE TABLE knowledge_bases (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,

    -- 租户+名称唯一索引
    UNIQUE INDEX idx_tenant_name (tenant_id, name),
    -- 租户查询索引
    INDEX idx_tenant_id (tenant_id),

    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

### 6.2 软删除模式

```go
// GORM 软删除支持
import "gorm.io/plugin/soft_delete"

type BaseModel struct {
    ID        uint64 `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt soft_delete.DeletedAt `gorm:"index"`
}

// 查询自动过滤已删除数据
db.Where("name = ?", "test").Find(&items)  // 自动加上 deleted_at IS NULL

// 包含已删除数据
db.Unscoped().Where("name = ?", "test").Find(&items)
```

### 6.3 事务处理

```go
// Store 层事务模式
func (s *store) CreateAgentWithTools(ctx context.Context, agent *Agent, tools []*Tool) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建 Agent
        if err := tx.Create(agent).Error; err != nil {
            return err
        }

        // 关联工具
        for _, tool := range tools {
            tool.AgentID = agent.ID
            if err := tx.Create(tool).Error; err != nil {
                return err  // 自动回滚
            }
        }

        return nil
    })
}
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **软删除** | 使用 deleted_at 字段 |
| **多租户** | tenant_id 字段 + 索引 |
| **JSON 字段** | 存储动态配置 |
| **时间戳** | 自动维护 created_at/updated_at |
| **事务边界** | 在 Store 层处理事务 |

### eino-show 当前状态

✅ 已实现 GORM 模型和软删除

---

## 7. 缓存策略

### 来源: WeKnora

### 7.1 缓存配置

```go
type CacheConfig struct {
    Type string `yaml:"type"` // "memory" 或 "redis"

    // Redis 配置
    Redis struct {
        Address  string        `yaml:"address"`
        Password string        `yaml:"password"`
        DB       int           `yaml:"db"`
        Prefix   string        `yaml:"prefix"`
        TTL      time.Duration `yaml:"ttl"`
    } `yaml:"redis"`

    // 内存缓存配置
    Memory struct {
        MaxSize int           `yaml:"max_size"`
        TTL     time.Duration `yaml:"ttl"`
    } `yaml:"memory"`
}
```

### 7.2 键命名规范

```go
const (
    // 会话上下文: session:{session_id}:context
    SessionContextKey = "session:%s:context"

    // 流式事件: stream:{session_id}:events
    StreamEventsKey = "stream:%s:events"

    // Web 搜索状态: search:temp:{search_id}
    SearchTempKey = "search:temp:%s"
)

func BuildSessionContextKey(sessionID string) string {
    return fmt.Sprintf("session:%s:context", sessionID)
}
```

### 7.3 缓存降级

```go
type CacheManager struct {
    redis  *redis.Client
    memory *sync.Map  // 降级到内存
}

func (m *CacheManager) Get(ctx context.Context, key string) (string, error) {
    // 优先从 Redis 获取
    if m.redis != nil {
        val, err := m.redis.Get(ctx, key).Result()
        if err == nil {
            return val, nil
        }
        // Redis 故障，降级到内存
        log.Warn("Redis unavailable, fallback to memory")
    }

    // 内存缓存降级
    if v, ok := m.memory.Load(key); ok {
        return v.(string), nil
    }

    return "", ErrCacheNotFound
}
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **多级缓存** | Redis + 内存降级 |
| **统一命名** | 使用前缀组织缓存键 |
| **合理 TTL** | 根据数据类型设置过期时间 |
| **缓存降级** | 缓存不可用时降级到数据源 |
| **分布式支持** | Redis 支持多实例部署 |

### eino-show 当前状态

✅ 已实现 Redis 分布式缓存 (`internal/pkg/llmcontext/`)

---

## 8. 中间件设计

### 来源: WeKnora

### 8.1 认证中间件

```go
// middleware/auth.go

func Auth(tenantService interfaces.TenantService,
          userService interfaces.UserService,
          cfg *config.Config) gin.HandlerFunc {

    return func(c *gin.Context) {
        // 1. 尝试 JWT Token 认证
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            token := strings.TrimPrefix(authHeader, "Bearer ")
            user, err := userService.ValidateToken(c.Request.Context(), token)
            if err == nil {
                c.Set("user", user)
                c.Set("auth_type", "jwt")
                c.Next()
                return
            }
        }

        // 2. 尝试 API Key 认证
        apiKey := c.GetHeader("X-API-Key")
        if apiKey != "" {
            tenant, err := tenantService.ValidateAPIKey(c.Request.Context(), apiKey)
            if err == nil {
                c.Set("tenant", tenant)
                c.Set("auth_type", "api_key")
                c.Next()
                return
            }
        }

        // 3. 认证失败
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "error": gin.H{
                "code":    ErrUnauthorized,
                "message": "未授权访问",
            },
        })
        c.Abort()
    }
}
```

### 8.2 错误处理中间件

```go
// middleware/error.go

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 处理过程中产生的错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            if appErr, ok := errors.IsAppError(err); ok {
                c.JSON(appErr.HTTPCode, gin.H{
                    "success": false,
                    "error": gin.H{
                        "code":    appErr.Code,
                        "message": appErr.Message,
                        "details": appErr.Details,
                    },
                })
                return
            }

            // 未知错误
            c.JSON(http.StatusInternalServerError, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    ErrInternal,
                    "message": "内部服务器错误",
                },
            })
        }
    }
}
```

### 8.3 日志中间件

```go
// middleware/logging.go

func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 生成请求 ID
        requestID := xid.New().String()
        c.Set("request_id", requestID)
        c.Request.Header.Set("X-Request-ID", requestID)

        // 记录请求
        logger.Infof(c.Request.Context(), "%s %s",
            c.Request.Method,
            c.Request.URL.Path,
        )

        // 记录响应
        c.Writer.Header().Set("X-Request-ID", requestID)

        c.Next()

        latency := time.Since(start)
        logger.Infof(c.Request.Context(), "completed in %dms", latency.Milliseconds())
    }
}
```

### 最佳实践

| 中间件 | 功能 |
|--------|------|
| **认证** | JWT + API Key 双模式 |
| **错误处理** | 统一错误响应格式 |
| **日志** | 请求 ID、耗时记录 |
| **限流** | 基于 IP 或用户的限流 |
| **CORS** | 跨域资源共享 |
| **恢复** | Panic 恢复，防止服务崩溃 |

### eino-show 当前状态

✅ 已实现认证、授权、日志等中间件 (`internal/pkg/middleware/`)

---

## 9. 构建系统

### 来源: cagent (Taskfile) + WeKnora (Makefile)

### 9.1 Taskfile (推荐 - 更现代)

```yaml
# Taskfile.yml

version: '3'

vars:
  BINARY_NAME: mb-apiserver
  BUILD_DIR: ./bin
  GIT_TAG:
    sh: git describe --tags --always --dirty
  LDFLAGS: '-X "github.com/ashwinyue/eino-show/pkg/version.Version={{.GIT_TAG}}"'

tasks:
  default:
    desc: 显示可用命令
    cmds:
      - task -l

  build:
    desc: 构建应用
    cmd: go build -ldflags "{{.LDFLAGS}}" -o {{.BUILD_DIR}}/{{.BINARY_NAME}}
    sources:
      - "**/*.go"
      - "go.mod"
      - "go.sum"
    generates:
      - "{{.BUILD_DIR}}/{{.BINARY_NAME}}"

  build:debug:
    desc: 构建调试版本
    cmd: go build -race -o {{.BUILD_DIR}}/{{.BINARY_NAME}}

  test:
    desc: 运行测试
    cmd: go test -v -race -coverprofile=coverage.out ./...

  test:coverage:
    desc: 测试覆盖率报告
    cmd: go tool cover -html=coverage.out -o coverage.html

  lint:
    desc: 代码检查
    cmd: golangci-lint run ./...

  proto:
    desc: 生成 Protobuf 代码
    cmd: buf generate

  wire:
    desc: 生成 Wire 依赖注入代码
    cmd: wire gen ./internal/apiserver/...

  clean:
    desc: 清理构建产物
    cmd: rm -rf {{.BUILD_DIR}}

  cross:
    desc: 跨平台构建
    cmd: docker buildx build --platform linux/amd64,linux/arm64
```

### 9.2 Makefile (兼容)

```makefile
# Makefile

.PHONY: build test lint clean

# 变量
BINARY_NAME=mb-apiserver
BUILD_DIR=./bin
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-X "github.com/ashwinyue/eino-show/pkg/version.Version=$(VERSION)"

# 默认目标
default:
	@echo "可用命令: build test lint clean"

# 构建
build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)

# 调试构建
build.debug:
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME)

# 测试
test:
	go test -v -race -coverprofile=coverage.out ./...

# 代码检查
lint:
	golangci-lint run ./...

# Proto 生成
proto:
	buf generate

# Wire 生成
wire:
	wire gen ./internal/apiserver/...

# 清理
clean:
	rm -rf $(BUILD_DIR)
```

### 9.3 多阶段 Dockerfile

```dockerfile
# Dockerfile

# 构建阶段
FROM golang:1.23-alpine AS builder

# 安装构建工具
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /build

# 下载依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags="-X 'github.com/ashwinyue/eino-show/pkg/version.Version=${VERSION}'" \
    -o /mb-apiserver ./cmd/mb-apiserver

# 运行阶段
FROM alpine:latest

# 安装 CA 证书
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 复制二进制文件
COPY --from=builder /mb-apiserver .

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5555/healthz || exit 1

# 暴露端口
EXPOSE 5555

# 运行
CMD ["./mb-apiserver"]
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **Taskfile** | 比 Makefile 更现代，支持依赖 |
| **版本注入** | 构建时注入 Git 版本信息 |
| **分层缓存** | Docker 构建利用 Go 模块缓存 |
| **健康检查** | Dockerfile 内置健康检查 |
| **跨平台构建** | 使用 buildx 支持多平台 |

### eino-show 当前状态

✅ 已有 Makefile，可考虑迁移到 Taskfile

---

## 10. 监控追踪

### 来源: WeKnora (OpenTelemetry) + Coze Studio (分布式追踪)

### 10.1 OpenTelemetry 初始化

```go
// internal/tracing/init.go

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(serviceName, endpoint string) (*sdktrace.TracerProvider, error) {
    // 资源定义
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String(version.Version),
        ),
    )
    if err != nil {
        return nil, err
    }

    // OTLP 导出器（支持 Jaeger、Zipkin）
    var exporter sdktrace.SpanExporter
    if endpoint != "" {
        client := otlptracegrpc.NewClient(
            otlptracegrpc.WithEndpoint(endpoint),
            otlptracegrpc.WithInsecure(),
        )
        exporter, err = otlptrace.New(context.Background(), client)
    } else {
        // 回退到标准输出
        exporter, err = stdouttrace.New()
    }
    if err != nil {
        return nil, err
    }

    // TracerProvider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

### 10.2 追踪中间件

```go
// middleware/tracing.go

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)

    return func(c *gin.Context) {
        spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
        ctx, span := tracer.Start(
            c.Request.Context(),
            spanName,
            trace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.url", c.Request.URL.String()),
                attribute.String("http.remote_addr", c.Request.RemoteAddr),
            ),
        )
        defer span.End()

        // 记录用户代理
        ua := c.Request.UserAgent()
        if ua != "" {
            span.SetAttributes(attribute.String("http.user_agent", ua))
        }

        // 替换请求上下文
        c.Request = c.Request.WithContext(ctx)

        c.Next()

        // 记录响应状态
        span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
    }
}
```

### 10.3 健康检查

```go
// handler/http/healthz.go

type HealthChecker struct {
    db    *gorm.DB
    redis *redis.Client
}

func (h *HealthChecker) Healthz(c *gin.Context) {
    status := map[string]any{
        "status": "ok",
        "time":   time.Now().Unix(),
    }

    // 数据库检查
    if h.db != nil {
        sqlDB, err := h.db.DB()
        if err != nil {
            status["database"] = "error: " + err.Error()
        } else if err := sqlDB.Ping(); err != nil {
            status["database"] = "error: " + err.Error()
        } else {
            status["database"] = "ok"
        }
    }

    // Redis 检查
    if h.redis != nil {
        if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
            status["redis"] = "error: " + err.Error()
        } else {
            status["redis"] = "ok"
        }
    }

    // 判断整体健康状态
    isHealthy := true
    for _, v := range status {
        if s, ok := v.(string); ok && strings.HasPrefix(s, "error") {
            isHealthy = false
            break
        }
    }

    if isHealthy {
        c.JSON(http.StatusOK, status)
    } else {
        c.JSON(http.StatusServiceUnavailable, status)
    }
}
```

### 最佳实践

| 实践 | 说明 |
|------|------|
| **OpenTelemetry** | 使用标准协议，兼容多种后端 |
| **环境变量配置** | 通过 OTEL_EXPORTER_OTLP_ENDPOINT 配置 |
| **采样策略** | 生产环境使用概率采样 |
| **健康检查** | 检查所有依赖服务 |
| **Docker Compose** | 集成 Jaeger 用于开发调试 |

### Docker Compose 配置

```yaml
# docker-compose.yml

services:
  app:
    image: mb-apiserver:latest
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
    depends_on:
      - jaeger

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
```

### eino-show 当前状态

✅ 已有 `pkg/otel/` 目录，需完善集成

---

## 总结对比表

| 工程化领域 | eino-show 当前状态 | 参考项目建议 |
|-----------|------------------|-------------|
| **错误处理** | 基础实现 | 参考 WeKnora 的统一错误码 |
| **日志系统** | pkg/log + logger | 建议统一到 slog |
| **配置管理** | Viper + YAML | ✅ 已完善 |
| **测试策略** | 覆盖较少 | 需补充单元测试 |
| **API 设计** | RESTful 规范 | ✅ 已完善 |
| **数据库** | GORM + 软删除 | ✅ 已完善 |
| **缓存策略** | Redis 分布式 | ✅ 已完善 |
| **中间件** | 认证/授权/日志 | ✅ 已完善 |
| **构建系统** | Makefile | 可考虑 Taskfile |
| **监控追踪** | pkg/otel 存在 | 需完善集成 |

**优先改进建议**:

1. **P0**: 统一日志到 slog
2. **P1**: 补充单元测试覆盖
3. **P1**: 完善 OpenTelemetry 集成
4. **P2**: 统一错误处理格式
5. **P2**: 添加健康检查端点
