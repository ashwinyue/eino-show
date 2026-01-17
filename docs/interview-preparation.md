# 面试准备：元宝AIGC后台开发岗位

## 岗位要求概述

### 岗位职责
1. 负责元宝AIGC应用（如文本生成、图像生成、音视频对话）的后台系统架构设计与开发，支撑高并发、低延迟的AI应用服务
2. 基于LLM大模型能力场景，推进建设AIGC平台研发落地，支持插件接入&管理、数据飞轮pipeline等平台能力
3. 负责解决平台和服务在高并发场景下的性能优化，问题定位&解决，保障服务SLA

### 岗位要求
1. 计算机专业本科及以上学历
2. 2年及以上后台开发经验，熟悉系统运维的基础知识
3. 精通GO开发语言，熟悉常见rpc框架
4. 熟悉微服务架构、数据库（MySQL/Redis/MongoDB）、消息队列（Kafka/RabbitMQ）及容器化技术（Docker/K8s）
5. 有流程编排引擎（如Airflow）、**LLM应用开发框架（如eino）** 开发经验优先
6. 有责任心，能积极主动推进项目进展

---

## 项目与岗位匹配度分析

### ✅ 已有亮点（面试重点讲）

| 岗位要求 | 项目对应实现 | 文件位置 |
|---------|-------------|---------|
| **eino 框架** | 深度使用 `cloudwego/eino`，实现了 Agent 工厂模式 | `internal/pkg/agent/factory.go` |
| **AI Agent 任务链编排** | Sequential/Loop/Parallel 三种工作流模式 | `internal/pkg/agent/workflow/workflow.go` |
| **Go + gRPC** | 完整的 gRPC + HTTP 双协议服务 | `internal/apiserver/grpcserver.go`, `httpserver.go` |
| **Redis** | 缓存、Session 存储、流式传输 | `internal/apiserver/cache/redis.go` |
| **PostgreSQL + pgvector** | 向量数据库支持 RAG | `docker-compose.yml` |
| **Docker/K8s** | K8s 部署配置 | `deployments/` |
| **多 Agent 协作** | Supervisor 模式、子 Agent 配置 | `internal/pkg/agent/supervisor/` |
| **工具管理** | 工具注册表、动态工具加载 | `internal/pkg/agent/tool/` |
| **MCP 协议** | MCP Server 集成 | `internal/pkg/mcp/` |

### 核心技术栈
```
Go 1.24 + Gin + gRPC + GORM + PostgreSQL + Redis + Docker/K8s
cloudwego/eino + OpenTelemetry + Casbin + JWT
```

---

## ⚠️ 建议补充的功能

### 1. 消息队列实际应用（高优先级）

**现状：** 项目有 `pkg/options/kafka_options.go` 但缺少实际业务场景

**建议添加：**
- [ ] 异步任务队列：文档解析、向量化任务
- [ ] Agent 执行结果异步回调
- [ ] 消息驱动的事件总线

**实现思路：**
```go
// internal/pkg/queue/kafka_producer.go
type TaskProducer struct {
    writer *kafka.Writer
}

func (p *TaskProducer) PublishDocumentTask(ctx context.Context, task *DocumentTask) error
func (p *TaskProducer) PublishAgentTask(ctx context.Context, task *AgentTask) error
```

---

### 2. 高并发/限流/熔断（高优先级）

**现状：** 缺少明显的限流、熔断、降级机制

**建议添加：**
- [ ] 接口限流中间件（令牌桶/漏桶）
- [ ] LLM 调用熔断降级
- [ ] 请求并发池控制
- [ ] 缓存预热机制

**实现思路：**
```go
// internal/pkg/middleware/gin/ratelimit.go
func RateLimitMiddleware(rate int, burst int) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(rate), burst)
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
            return
        }
        c.Next()
    }
}

// internal/pkg/circuitbreaker/breaker.go
type LLMCircuitBreaker struct {
    breaker *gobreaker.CircuitBreaker
}
```

---

### 3. 数据飞轮 Pipeline（中优先级）

**现状：** 缺少数据收集和反馈闭环

**建议添加：**
- [ ] 用户反馈收集（点赞/点踩）
- [ ] 对话质量评估指标
- [ ] 自动生成训练数据的 Pipeline

**数据库表设计：**
```sql
CREATE TABLE feedback (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(64),
    message_id VARCHAR(64),
    feedback_type VARCHAR(20), -- 'like', 'dislike', 'report'
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE quality_metrics (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(64),
    latency_ms INT,
    token_count INT,
    relevance_score FLOAT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

### 4. 可观测性增强（中优先级）

**现状：** 有 OpenTelemetry 但可以更突出

**建议添加：**
- [ ] Agent 执行链路追踪
- [ ] LLM Token 用量监控
- [ ] 响应延迟 P99 指标
- [ ] Grafana Dashboard 配置

**实现思路：**
```go
// internal/pkg/metrics/llm_metrics.go
var (
    LLMRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "llm_request_duration_seconds",
            Help:    "LLM request duration in seconds",
            Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
        },
        []string{"model", "status"},
    )
    
    LLMTokenUsage = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_token_usage_total",
            Help: "Total LLM token usage",
        },
        []string{"model", "type"}, // type: input/output
    )
)
```

---

### 5. 单元测试覆盖（低优先级）

**现状：** 主业务代码测试较少

**建议添加：**
- [ ] Agent 工厂核心逻辑测试
- [ ] Handler 层集成测试
- [ ] Benchmark 性能测试

---

## 📋 快速行动计划

### 优先级排序
1. **P0 - 限流熔断** - 1天内可完成，展示高并发处理能力
2. **P0 - Kafka 异步任务** - 1-2天，展示消息队列实战
3. **P1 - 用户反馈收集** - 0.5天，展示数据飞轮概念
4. **P1 - LLM 监控指标** - 0.5天，展示可观测性
5. **P2 - 单元测试** - 按需添加

### 面试话术准备

**关于 eino 框架：**
> "我在项目中深度使用了字节的 eino 框架，实现了 Agent 工厂模式，支持 ReAct、Chat、Supervisor 等多种 Agent 类型，以及 Sequential、Loop、Parallel 三种工作流编排模式。"

**关于高并发：**
> "项目实现了接口限流（令牌桶）、LLM调用熔断降级、Redis缓存预热等机制，保障高并发场景下的服务稳定性。"

**关于数据飞轮：**
> "我们建立了用户反馈收集机制，通过对话质量评估指标形成数据闭环，为模型优化提供数据支撑。"

---

---

## 高并发低延迟实现分析（对话功能）

### ✅ 已实现的优化

#### 1. SSE 流式响应（低延迟核心）
```go
// internal/apiserver/biz/session/sse_adapter.go
type SSEAdapter struct {
    ch chan []byte  // 带缓冲通道，100 容量
}
```
- **首 Token 延迟低**：用户无需等待完整回答
- **通道缓冲**：防止写阻塞
- **实时推送**：LLM 生成即推送

#### 2. 异步执行 + Context Cancel
```go
// internal/apiserver/biz/session/qa.go
go func() {  // 异步 goroutine 执行
    defer close(reader.done)
    // Agent 执行逻辑
}()
```
- **请求立即返回**：后台异步处理
- **支持中途取消**：用户可随时中断
- **资源及时释放**：`context.WithCancel`

#### 3. Redis 缓存层
```go
// internal/apiserver/cache/keys.go
SetSession()    // Session 数据缓存
SetEmbedding()  // 向量缓存（避免重复调用 Embedding API）
SetAgent()      // Agent 配置缓存
```

#### 4. 分布式支持（已具备基础）

| 组件 | 文件位置 | 说明 |
|-----|---------|-----|
| **分布式锁** | `pkg/distlock/redis.go` | Redis 分布式锁，带自动续期 |
| **服务注册** | `pkg/server/kratos_server.go` | 支持 Consul/Etcd 注册中心 |
| **K8s 多副本** | `deployments/mb-apiserver-deployment.yaml` | replicas: 2，支持水平扩展 |
| **Pod 反亲和** | 同上 | 多副本分散到不同节点 |

### ✅ 已实现：接口限流中间件

#### 限流中间件（已集成）
```go
// internal/pkg/middleware/gin/ratelimit.go
// 支持三层限流：全局 + 用户级 + 接口级

// 1. 全局限流（漏桶算法）
GlobalRateLimiter  // 基于 uber/ratelimit，平滑限流

// 2. 用户级限流（Redis 滑动窗口）
RedisRateLimiter   // 分布式限流，按 IP/用户 ID 限流

// 3. 接口级限流（差异化策略）
EndpointRateLimiter // 不同接口不同 QPS 上限

// 4. 组合限流器
CombinedRateLimiter // 多层限流叠加
```

**当前配置**（`internal/apiserver/httpserver.go`）：
```go
rateLimitCfg := &mw.CombinedConfig{
    GlobalRate: 1000,  // 全局 1000 QPS
    UserRate:   100,   // 每用户 100 QPS
    EndpointLimits: []mw.EndpointLimit{
        {Path: "/api/v1/sessions/*/qa", Rate: 50},      // 问答：50 QPS
        {Path: "/api/v1/agent-chat/*", Rate: 30},       // Agent：30 QPS
        {Path: "/api/v1/knowledge-bases/*/file", Rate: 10}, // 上传：10 QPS
    },
}
```

### ✅ 已实现：请求去重/合并

#### 去重组件（已集成）
```go
// internal/pkg/dedup/dedup.go
// 基于 singleflight 实现，防止相同请求并发时重复调用 LLM

// 核心功能：
// 1. 并发请求合并 - 相同 query 只执行一次 LLM 调用
// 2. 短期缓存 - 10秒内相同问题直接返回缓存
// 3. 自动清理 - 定期清理过期缓存

type QADeduplicator struct {
    group singleflight.Group  // 并发请求合并
    cache sync.Map            // 短期结果缓存
}

// 使用示例
result, err, shared := deduplicator.DeduplicateQA(ctx, sessionID, query, fn)
// shared=true 表示结果来自缓存或其他并发请求
```

**效果**：高并发下相同问题只调用一次 LLM，节省成本、降低延迟。

---

### ⚠️ 建议补充的优化

#### 1. LLM 调用熔断降级（高优先级）
```go
// 建议新增：internal/pkg/circuitbreaker/llm_breaker.go
import "github.com/sony/gobreaker"

type LLMCircuitBreaker struct {
    breaker *gobreaker.CircuitBreaker
}

func (b *LLMCircuitBreaker) Call(ctx context.Context, fn func() error) error {
    _, err := b.breaker.Execute(func() (interface{}, error) {
        return nil, fn()
    })
    return err
}
```

#### 3. 连接池优化
```go
// LLM API 连接复用
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
```

#### 4. 请求去重/合并
```go
// 相同问题短时间内去重
import "golang.org/x/sync/singleflight"

var group singleflight.Group

func (s *service) Query(ctx context.Context, query string) (string, error) {
    v, err, _ := group.Do(query, func() (interface{}, error) {
        return s.doQuery(ctx, query)
    })
    return v.(string), err
}
```

---

## 分布式架构方案

### 是否需要分布式？

**答案：看场景，但项目已具备基础**

| 场景 | 是否需要 | 说明 |
|-----|---------|-----|
| 单机 1000 QPS 以下 | ❌ | 单实例够用 |
| 高可用要求 | ✅ | 至少 2 副本 |
| 10000+ 并发 | ✅ | 需要水平扩展 |
| 多机房部署 | ✅ | 需要服务发现 |

### 当前项目的分布式能力

```
┌─────────────────────────────────────────────────────────────┐
│                        K8s Ingress                          │
│                      (负载均衡入口)                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
         ┌────────────┴────────────┐
         ▼                         ▼
┌─────────────────┐       ┌─────────────────┐
│  mb-apiserver   │       │  mb-apiserver   │
│    (Pod 1)      │       │    (Pod 2)      │
└────────┬────────┘       └────────┬────────┘
         │                         │
         └────────────┬────────────┘
                      │
    ┌─────────────────┼─────────────────┐
    ▼                 ▼                 ▼
┌────────┐      ┌──────────┐      ┌──────────┐
│ Redis  │      │PostgreSQL│      │ LLM API  │
│(缓存/锁)│      │ (pgvector)│      │(OpenAI等)│
└────────┘      └──────────┘      └──────────┘
```

### 关键分布式组件

1. **Redis**
   - Session 共享存储
   - 分布式锁（防止并发冲突）
   - Embedding 缓存

2. **服务注册发现**（可选）
   - Consul/Etcd 支持
   - Kratos 框架集成

3. **无状态设计**
   - Session 存储在 Redis/DB
   - 任意 Pod 可处理任意请求

---

## 面试话术：高并发低延迟

**问：你的项目如何实现高并发低延迟？**

> "对话功能的高并发低延迟主要通过以下几点实现：
> 
> 1. **SSE 流式响应**：LLM 生成内容实时推送，首 Token 延迟从几秒降到毫秒级
> 2. **异步执行**：请求立即返回，后台 goroutine 处理 LLM 调用
> 3. **Redis 缓存**：Session、Embedding 向量都做了缓存，减少重复计算
> 4. **分布式锁**：使用 Redis 实现，带自动续期，防止并发冲突
> 5. **K8s 水平扩展**：多副本部署，Pod 反亲和分散到不同节点
> 
> 后续计划补充接口限流和 LLM 调用熔断，进一步提升稳定性。"

---

## 相关文件索引

- Agent 工厂：`internal/pkg/agent/factory.go`
- 工作流编排：`internal/pkg/agent/workflow/`
- HTTP 路由：`internal/apiserver/httpserver.go`
- gRPC 服务：`internal/apiserver/grpcserver.go`
- Redis 缓存：`internal/apiserver/cache/redis.go`
- 分布式锁：`pkg/distlock/redis.go`
- 服务注册：`pkg/server/kratos_server.go`
- K8s 部署：`deployments/mb-apiserver-deployment.yaml`
- SSE 适配器：`internal/apiserver/biz/session/sse_adapter.go`
- QA 执行器：`internal/apiserver/biz/session/qa.go`
