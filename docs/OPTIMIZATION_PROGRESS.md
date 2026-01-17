# Eino Agent 优化进度

> 参考项目:
> - AssistantAgent (Java) - 上下文工程、经验管理
> - Coze Studio (Go) - 模板系统、混合检索、对话优化
>
> 详细分析: [docs/reference/COZE_STUDIO_ANALYSIS.md](./reference/COZE_STUDIO_ANALYSIS.md)

## 已完成模块 ✅

| 模块 | 文件 | 说明 |
|------|------|------|
| **意图路由 (Intent Router)** | `internal/pkg/agent/router/intent_router.go` | 基于 compose.Graph 的多维意图识别 |
| **动态 Prompt (Dynamic Prompt)** | `internal/pkg/agent/router/dynamic_prompt.go` | 条件化 Prompt 片段动态组装 |
| **经验管理 (Experience)** | `internal/pkg/agent/router/experience.go` | 经验存储、检索、快速意图匹配 |
| **Redis 上下文存储** | `internal/pkg/llmcontext/redis_storage.go` | LLM Context 存储切换为 Redis (和 WeKnora 一致) |
| **自动学习 (Learning)** | `internal/pkg/agent/router/learning.go` | 从成功执行中自动提取经验 |
| **学习回调集成** | `internal/apiserver/biz/session/sse_adapter.go` | SSEAdapter 中收集工具调用信息并触发学习 |
| **工具拦截器 (Tool Interceptor)** | `internal/pkg/agent/router/tool_interceptor.go` | 工具调用拦截、经验提取、模式识别、指标收集 |
| **Redis Stream Manager** | `internal/pkg/stream/redis_manager.go` | SSE 事件流分布式存储 (和 WeKnora 一致) |
| **Web Search State** | `internal/pkg/stream/web_search_state.go` | Web 搜索临时状态缓存 (和 WeKnora 一致) |
| **EnhancedAgent** | `internal/pkg/agent/enhanced/agent.go` | 增强 Agent (集成意图路由/动态Prompt/经验) |
| **ADK Agent 包装器** | `internal/pkg/agent/enhanced/adk_wrapper.go` | ADK 多 Agent 支持 (子 Agent 转移) |
| **QA 集成** | `internal/apiserver/biz/session/qa.go` | 完整调用链集成 (enhanced + 多Agent 模式) |
| **依赖注入 Provider** | `internal/apiserver/biz/session/provider.go` | 组件初始化和配置管理 |
| **Biz 层集成** | `internal/apiserver/biz/biz.go` | 懒加载增强 Session (自动启用) |
| **Handler 层集成** | `internal/apiserver/handler/http/session.go` | SSE 流式响应对接 biz.Session().QA |
| **渐进式压缩策略** | `internal/pkg/llmcontext/compression.go` | ProgressiveCompressionStrategy 封装 |
| **混合检索** | `internal/apiserver/biz/knowledge/knowledge.go` | HybridSearch 基础实现 |
| **Handler 全量集成** | `internal/apiserver/handler/http/*.go` | 50+ 方法对接 biz 层 |
| **知识库统计** | `internal/apiserver/biz/knowledge/knowledge.go` | GetKBStats 实现 |
| **知识搜索** | `internal/apiserver/biz/session/session.go` | SearchKnowledge 关键词匹配 |
| **混合检索评分** | `internal/apiserver/biz/knowledge/knowledge.go` | HybridSearch TF 评分排序 |
| **Agent 执行** | `internal/apiserver/biz/agent/agent.go` | Execute SSE 流式响应 |
| **MCP 服务测试** | `internal/apiserver/biz/mcp/mcp.go` | Test/GetTools 实现 |
| **学习回调集成** | `internal/apiserver/biz/session/session.go` | ExperienceManager 自动学习 |

## 待办模块 📋

### P0: ChatTemplate 模板系统

**来源**: Coze Studio 最佳实践

**功能**: 使用 Eino ChatTemplate 替代字符串拼接

**收益**:
- 代码可维护性 ↑↑
- 支持 Jinja2/FString 模板语法
- 易于 A/B 测试不同提示词

**实现要点**:
```go
import "github.com/cloudwego/eino/components/model"

template, err := model.NewChatTemplate(ctx, &model.ChatTemplateConfig{
    Format:   model.FORMAT_JINJA2,  // 或 FORMAT_FSTRING
    Template: `
你是{{ role_type }}助手。

{% if knowledge_context %}
# 知识库内容
{{ knowledge_context }}
{% endif %}

{% if examples %}
# 示例
{% for ex in examples %}
Q: {{ ex.question }}
A: {{ ex.answer }}
{% endfor %}
{% endif %}
`,
})
```

**文件规划**:
- `internal/pkg/agent/prompts/template.go` - 模板管理器
- `internal/pkg/agent/prompts/templates/` - 模板文件目录

---

### P1: KV 内存变量管理

**来源**: Coze Studio 最佳实践

**功能**: 对话级别的键值存储，支持用户偏好记忆

**收益**:
- 对话连续性 ↑
- 实体提取缓存
- 用户偏好持久化

**实现要点**:
```go
type KVMemory struct {
    store kv.Store
}

type KVItem struct {
    Key           string
    Value         any
    Type          string  // string/number/json
    IsSystem      bool
    PromptInject  bool    // 是否注入到提示词
    TTL           time.Duration
}

// 与 Prompt 集成
func (m *KVMemory) BuildPromptVars() map[string]any {
    // 返回可注入到模板的变量
}
```

**文件规划**:
- `internal/pkg/agent/memory/kv_memory.go` - KV 内存核心
- `internal/pkg/agent/memory/store.go` - 存储抽象
- `internal/apiserver/store/memory.go` - 持久化存储

---

### P2: 混合检索 + RRF 重排序

**来源**: Coze Studio 最佳实践

**功能**: 向量检索 + 全文检索并行，RRF 算法融合

**收益**:
- 召回率 ↑
- 检索精度 ↑
- 支持关键词精确匹配

**实现要点**:
```go
// 并行检索
parallelNode := compose.NewParallel().
    AddLambda("vectorSearch", vectorRetrieveNode).
    AddLambda("fulltextSearch", fulltextRetrieveNode)

// RRF 重排序
func (r *rrfReranker) Rerank(ctx context.Context, req *Request) (*Response, error) {
    id2Score := make(map[string]float64)
    for _, resultList := range req.Data {
        for rank, result := range resultList {
            // RRF: 1 / (rank + k)
            score := 1.0 / (float64(rank) + float64(r.k))
            id2Score[result.ID] = max(id2Score[result.ID], score)
        }
    }
    return sortResults(id2Score), nil
}
```

**文件规划**:
- `internal/pkg/retriever/hybrid.go` - 混合检索
- `internal/pkg/retriever/rrf.go` - RRF 重排序

---

### P2: 思维链 (Chain of Thought)

**来源**: Coze Studio 最佳实践

**功能**: 推理过程独立存储和展示

**收益**:
- 透明度 ↑
- 可调试性 ↑
- 用户信任度 ↑

**实现要点**:
```go
// 1. 扩展消息结构
type Message struct {
    Role            string
    Content         string
    ReasoningContent string  // 推理内容
}

// 2. SSE 事件
sseEvent := map[string]any{
    "type": "agent_reasoning",
    "content": reasoningContent,
}

// 3. 推理模板
reasoningPrompt := `
请先分析问题，然后给出答案。

## 分析过程
{{ reasoning }}

## 最终答案
{{ final_answer }}
`
```

**文件规划**:
- `internal/pkg/agent/reasoning/reasoning.go` - 推理处理器
- 前端支持可折叠展示

---

### P3: 工作流作为工具

**来源**: Coze Studio 最佳实践

**功能**: 将工作流包装为 Agent 可调用的工具

**收益**:
- 功能扩展性 ↑
- Agent 能力组合

**实现要点**:
```go
type WorkflowTool struct {
    workflow *compose.Runnable[map[string]any, string]
}

func (w *WorkflowTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "workflow_" + w.name,
        Desc: "执行工作流",
    }, nil
}

func (w *WorkflowTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    input := parseArgs(args)
    result, err := w.workflow.Invoke(ctx, input)
    return result, err
}
```

---

### P10: Trigger 触发器模块

**状态**: 待实现

**功能**: 主动服务能力，支持定时/延迟/回调执行 Agent

**设计方案**:
- 复用 `pkg/watch` 模块 (已有 robfig/cron + 分布式锁)
- 参考 AssistantAgent `trigger` 模块设计

**调度模式**:
| 模式 | 说明 | 示例 |
|------|------|------|
| `CRON` | Cron 表达式定时执行 | `"0 9 * * *"` 每天9点 |
| `DELAY` | 延迟执行 | `"30m"` 30分钟后 |
| `CALLBACK` | 事件回调执行 | 外部事件触发 |

**核心结构**:
```go
type TriggerDefinition struct {
    ID            string
    Name          string
    ScheduleMode  ScheduleMode  // cron/delay/callback
    ScheduleValue string        // Cron表达式或延迟时间
    ExecuteFunc   string        // 要执行的 Agent 函数
    SessionID     string        // 关联的会话
    Status        TriggerStatus
    ExpireAt      *time.Time
}

type TriggerManager struct {
    cron     *cron.Cron
    triggers map[string]*TriggerDefinition
    executor AgentExecutor
    store    TriggerStore
}
```

**文件规划**:
- `internal/pkg/agent/trigger/trigger.go` - 核心定义
- `internal/pkg/agent/trigger/manager.go` - 触发器管理
- `internal/pkg/agent/trigger/executor.go` - Agent 执行器
- `internal/apiserver/store/trigger.go` - 持久化存储

---

### P11: OpenTelemetry + Jaeger 分布式追踪

**状态**: 待定 (按需实现)

**当前方案**: 已集成 **coze-loop** 用于 Agent/LLM 调用追踪

**Jaeger 用途**: 微服务分布式链路追踪 (HTTP 请求、数据库调用、Redis 等)

**参考**: WeKnora 使用 OTEL 标准
- 文件: `a-old/WeKnora/internal/tracing/init.go`
- 通过 `OTEL_EXPORTER_OTLP_ENDPOINT` 环境变量配置后端

**实现要点**:
```go
// 1. 添加依赖
// go.opentelemetry.io/otel
// go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc

// 2. 初始化 TracerProvider
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.AlwaysSample()),
    sdktrace.WithResource(res),
    sdktrace.WithSpanProcessor(bsp),
)
otel.SetTracerProvider(tp)

// 3. 配置环境变量
// OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
```

**Docker Compose 配置**:
```yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  ports:
    - "16686:16686"  # UI
    - "4317:4317"    # OTLP gRPC
```

**何时需要**:
- 单体应用 + coze-loop：**不需要**
- 微服务架构 + 全链路追踪：**需要**

---

## 工程化改进 🛠️

> 详细分析: [docs/reference/ENGINEERING_BEST_PRACTICES.md](./reference/ENGINEERING_BEST_PRACTICES.md)

### P0: 统一日志到 slog

**来源**: cagent 最佳实践

**功能**: 将现有的 pkg/log 和 pkg/logger 统一到 Go 标准库 slog

**收益**:
- 减少第三方依赖
- 统一日志接口
- 更好的性能

**实现要点**:
```go
import "log/slog"

func setupLogging(debug bool, logFile string) error {
    if !debug {
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

    opts := &slog.HandlerOptions{Level: slog.LevelDebug}
    slog.SetDefault(slog.New(slog.NewTextHandler(writer, opts)))
    return nil
}
```

**文件规划**:
- `pkg/log/slog.go` - slog 封装
- 逐步替换 `pkg/logger/` 使用

---

### P1: 补充单元测试

**来源**: WeKnora + cagent

**目标覆盖率**: 核心业务逻辑 70%+

**优先测试模块**:
| 模块 | 文件 | 优先级 |
|------|------|--------|
| 工具注册表 | `internal/pkg/agent/tool/registry.go` | 高 |
| 意图路由 | `internal/pkg/agent/router/intent_router.go` | 高 |
| 经验管理 | `internal/pkg/agent/router/experience.go` | 高 |
| Biz 层 | `internal/apiserver/biz/*/` | 中 |
| Store 层 | `internal/apiserver/store/*.go` | 中 |

---

### P1: 完善 OpenTelemetry 集成

**来源**: WeKnora

**功能**: 分布式追踪 + 性能监控

**实现要点**:
```go
// 1. 初始化 Tracer
tp, err := tracing.InitTracer("eino-show", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))

// 2. 追踪中间件
router.Use(middleware.TracingMiddleware("eino-show"))

// 3. Agent 追踪
ctx, span := tracer.Start(ctx, "agent.execute")
defer span.End()
```

**Docker Compose**:
```yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  ports:
    - "16686:16686"  # UI
    - "4317:4317"    # OTLP gRPC
```

---

### P2: 统一错误处理格式

**来源**: WeKnora

**功能**: 统一错误码 + 结构化错误

**实现要点**:
```go
type AppError struct {
    Code     ErrorCode `json:"code"`
    Message  string    `json:"message"`
    Details  any       `json:"details"`
    HTTPCode int       `json:"-"`
}

// 错误响应格式
{
    "success": false,
    "error": {
        "code": 2100,
        "message": "参数错误",
        "details": {}
    }
}
```

---

### P2: 健康检查端点

**来源**: WeKnora + cagent

**功能**: 检查依赖服务健康状态

**实现要点**:
```go
func (h *HealthChecker) Healthz(c *gin.Context) {
    status := map[string]any{
        "status": "ok",
        "database": h.checkDB(),
        "redis": h.checkRedis(),
    }
    c.JSON(statusCode, status)
}
```

---

## 不需要实现的模块

| 模块 | 原因 |
|------|------|
| **Code-as-Action** | 需要 GraalVM 沙箱，复杂度高 |
| **Reply Channel** | 多渠道回复，按需实现 |
| **后台异步任务** | WeKnora 使用 asynq，与 Trigger 不同用途 |

---

## 测试账号

| 邮箱 | 密码 |
|------|------|
| starry99c@163.com | qq123456 |

---

## 架构对比

### AssistantAgent (Java)
```
评估图 → Prompt Builder → Agent 执行 → 经验学习
                              ↓
                        触发器/回复渠道
```

### eino-show (Go)
```
意图路由 → 动态 Prompt → Agent 执行 → 自动学习
  (compose.Graph)              ↓
                         [待实现] Trigger
```
