# WeKnora 与 cagent 对比分析与优化建议

## 文档概述

本文档基于对 WeKnora 主项目和 old/cagent 项目的深度对比分析，提出具体的优化建议和实施路径。

**分析日期**: 2025-01-14
**分析范围**: 架构设计、核心模块、工程实践

---

## 一、项目概述对比

### 1.1 WeKnora 项目

**定位**: 企业级 LLM 文档理解与检索框架

**核心特性**:
- 多租户知识库管理
- RAG 检索增强生成
- 15+ LLM 提供商支持
- Vue 3 前端 + Go 后端
- WebSocket 流式对话

**技术栈**:
| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24+, Gin, GORM, Uber Dig |
| 数据库 | PostgreSQL, Redis, Elasticsearch/Qdrant |
| 前端 | Vue 3, TypeScript, TDesign, Vite |
| 部署 | Docker, Kubernetes, Helm |

### 1.2 cagent 项目

**定位**: 多智能体运行时系统

**核心特性**:
- 多智能体协作与移交
- MCP 协议原生支持
- 丰富的工具生态
- 交互式 TUI 界面
- 版本化配置管理

**技术栈**:
| 层级 | 技术 |
|------|------|
| 后端 | Go 1.25+, Cobra CLI, ConnectRPC |
| 前端 | Bubble Tea, Lipgloss, Glamour |
| 通信 | Protocol Buffers, gRPC |
| 集成 | MCP 协议, 多模型提供商 |

---

## 二、核心架构对比

### 2.1 Agent 引擎

#### WeKnora Agent 设计

```go
// internal/agent/engine.go
type AgentEngine struct {
    config               *types.AgentConfig
    toolRegistry         *tools.ToolRegistry
    chatModel            chat.Chat
    eventBus             *event.EventBus
    knowledgeBasesInfo   []*KnowledgeBaseInfo
    selectedDocs         []*SelectedDocumentInfo
    contextManager       interfaces.ContextManager
    sessionID            string
}
```

**优势**:
- 清晰的 ReAct 执行流程（Think-Act-Observe）
- 事件驱动架构，可观测性强
- 专门的 SequentialThinkingTool 支持思考链
- 内置上下文管理器

**可改进点**:
- 缺少多智能体协作机制
- 无智能体间移交功能
- 工具集管理相对简单

#### cagent Agent 设计

```go
// old/cagent/pkg/agent/agent.go
type Agent struct {
    name               string
    description        string
    instruction        string
    toolsets           []*StartableToolSet
    models             []provider.Provider
    subAgents          []*Agent        // 子智能体
    handoffs           []*Agent        // 移交目标
    parents            []*Agent        // 父智能体
    maxIterations      int
}
```

**优势**:
- 完整的多智能体架构
- 支持智能体层次关系
- 专门的移交工具（handoff）
- 运行时模型覆盖

**可借鉴设计**:

| 特性 | 建议优先级 | 实施难度 |
|------|-----------|---------|
| 子智能体支持 | 高 | 中 |
| 移交机制 | 高 | 中 |
| 工具集生命周期管理 | 中 | 低 |

#### 优化建议: 增强 Agent 引擎

```go
// 建议新增: 智能体网络
type AgentNetwork struct {
    agents     map[string]*Agent
    handoffRules []HandoffRule
    currentAgent  string
}

type HandoffRule struct {
    SourceAgent  string
    TargetAgent  string
    Condition    HandoffCondition
    AutoTrigger  bool
}

type HandoffCondition struct {
    KeywordMatch  []string
    IntentMatch   []string
    ConfidenceThreshold float64
}

// 建议新增: 智能体移交工具
type HandoffTool struct {
    BaseTool
    network *AgentNetwork
}

func (h *HandoffTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
    var params struct {
        TargetAgent string `json:"target_agent"`
        Reason      string `json:"reason"`
        Context     map[string]interface{} `json:"context"`
    }
    // 解析参数并执行移交
}
```

### 2.2 工具系统

#### WeKnora 工具系统

```go
// internal/agent/tools/registry.go
type ToolRegistry struct {
    tools map[string]types.Tool
}

type Tool interface {
    Name() string
    Description() string
    Schema() json.RawMessage
    Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
```

**特点**:
- 简洁的单一接口设计
- 集中式注册管理
- 基础的工具执行

#### cagent 工具系统

```go
// old/cagent/pkg/tools/tools.go
type ToolSet interface {
    Tools(ctx context.Context) ([]Tool, error)
    Instructions() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    SetElicitationHandler(handler ElicitationHandler)
}

type Tool struct {
    Name         string
    Category     string
    Description  string
    Parameters   any
    Annotations  ToolAnnotations
    OutputSchema any
    Handler      ToolHandler
}
```

**特点**:
- ToolSet 双层抽象
- 生命周期管理（Start/Stop）
- MCP 协议原生支持
- 工具确认机制（Elicitation）

#### 对比总结

| 特性 | WeKnora | cagent | 建议 |
|------|---------|--------|------|
| 核心抽象 | 单一 Tool 接口 | Tool + ToolSet | 引入 ToolSet |
| 生命周期 | 无明确管理 | Start/Stop | 添加生命周期 |
| MCP 支持 | 包装器模式 | 原生支持 | 考虑深度集成 |
| 工具分类 | 无 | Category 标签 | 添加分类系统 |
| 元数据 | 基础 | 丰富的 Annotations | 扩展元数据 |

#### 优化建议: 增强 ToolSet 抽象

```go
// 建议新增: 工具集接口
type ToolSet interface {
    // 获取工具集名称
    Name() string

    // 获取所有工具
    Tools(ctx context.Context) ([]types.Tool, error)

    // 获取工具集说明（用于 Agent prompt）
    Instructions() string

    // 生命周期管理
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    IsStarted() bool

    // 可选: 依赖管理
    Dependencies() []string
}

// 建议新增: 基础工具集实现
type BaseToolSet struct {
    name     string
    started  bool
    mu       sync.RWMutex
}

func (b *BaseToolSet) Name() string { return b.name }
func (b *BaseToolSet) Start(ctx context.Context) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.started = true
    return nil
}
func (b *BaseToolSet) Stop(ctx context.Context) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.started = false
    return nil
}
func (b *BaseToolSet) IsStarted() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    return b.started
}

// 建议新增: 工具注册中心增强
type EnhancedToolRegistry struct {
    tools     map[string]types.Tool
    toolSets  map[string]ToolSet
    started   map[string]bool
    mu        sync.RWMutex
}

func (r *EnhancedToolRegistry) RegisterToolSet(ts ToolSet) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.toolSets[ts.Name()] = ts
    return nil
}

func (r *EnhancedToolRegistry) StartAll(ctx context.Context) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, ts := range r.toolSets {
        if !ts.IsStarted() {
            if err := ts.Start(ctx); err != nil {
                return fmt.Errorf("failed to start toolset %s: %w", ts.Name(), err)
            }
            r.started[ts.Name()] = true
        }
    }
    return nil
}

func (r *EnhancedToolRegistry) GetAllTools(ctx context.Context) ([]types.Tool, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var allTools []types.Tool

    // 添加单独注册的工具
    for _, tool := range r.tools {
        allTools = append(allTools, tool)
    }

    // 添加工具集中的工具
    for _, ts := range r.toolSets {
        if ts.IsStarted() {
            tools, err := ts.Tools(ctx)
            if err != nil {
                continue // 记录警告但继续
            }
            allTools = append(allTools, tools...)
        }
    }

    return allTools, nil
}
```

### 2.3 配置管理

#### WeKnora 配置

```go
type Config struct {
    Conversation    *ConversationConfig
    Server          *ServerConfig
    KnowledgeBase   *KnowledgeBaseConfig
    Tenant          *TenantConfig
    Models          []ModelConfig
    VectorDatabase  *VectorDatabaseConfig
    // ...
}
```

**特点**:
- 简洁的单一结构
- Viper 加载
- 环境变量支持

**不足**:
- 无版本标识
- 无配置迁移机制
- 验证较弱

#### cagent 配置

```go
// 多版本支持
func Parsers() map[string]func([]byte) (any, error) {
    return map[string]func([]byte) (any, error){
        v0.Version: func(d []byte) (any, error) { return v0.Parse(d) },
        v1.Version: func(d []byte) (any, error) { return v1.Parse(d) },
        v2.Version: func(d []byte) (any, error) { return v2.Parse(d) },
        v3.Version: func(d []byte) (any, error) { return v3.Parse(d) },
        latest.Version: func(d []byte) (any, error) { return latest.Parse(d) },
    }
}

// 自动迁移
func migrateToLatestConfig(c any, raw []byte) (latest.Config, error) {
    var err error
    for _, upgrade := range Upgrades() {
        c, err = upgrade(c, raw)
        if err != nil {
            return latest.Config{}, err
        }
    }
    return c.(latest.Config), nil
}
```

**特点**:
- 版本化配置（v0-v4）
- 自动迁移机制
- 模型别名解析
- 完整的验证

#### 优化建议: 版本化配置

```go
// 1. 添加版本标识
type Config struct {
    Version string `yaml:"version" json:"version" validate:"required,semver"`

    // 现有配置项...
    Conversation    *ConversationConfig    `yaml:"conversation"`
    Server          *ServerConfig          `yaml:"server"`
    // ...
}

// 2. 配置验证接口
type ConfigValidator interface {
    Validate() error
    GetValidationRules() []ValidationRule
}

func (c *Config) Validate() error {
    var errs []error

    // 版本检查
    if c.Version == "" {
        errs = append(errs, errors.New("config version is required"))
    }

    // 必需字段检查
    if c.Server == nil || c.Server.Port == 0 {
        errs = append(errs, errors.New("server.port is required"))
    }

    // 端口范围验证
    if c.Server != nil && (c.Server.Port < 1 || c.Server.Port > 65535) {
        errs = append(errs, errors.New("server.port must be between 1 and 65535"))
    }

    // 模型配置验证
    for i, model := range c.Models {
        if model.Type == "" {
            errs = append(errs, fmt.Errorf("models[%d].type is required", i))
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("configuration validation failed: %v", errs)
    }
    return nil
}

// 3. 配置迁移支持
type ConfigMigrator interface {
    MigrateFrom(version string, raw []byte) (*Config, error)
    SupportedVersions() []string
}

type ConfigMigratorV1 struct{}

func (m *ConfigMigratorV1) MigrateFrom(version string, raw []byte) (*Config, error) {
    // 解析旧版本配置
    var oldConfig struct {
        Server struct {
            Port int `yaml:"port"`
        } `yaml:"server"`
        // 旧版本字段...
    }

    if err := yaml.Unmarshal(raw, &oldConfig); err != nil {
        return nil, err
    }

    // 转换为新版本
    newConfig := &Config{
        Version: "2.0",
        Server: &ServerConfig{
            Port: oldConfig.Server.Port,
        },
        // 其他字段映射...
    }

    return newConfig, nil
}

// 4. 增强的配置加载
func LoadConfigWithMigration(configPath string) (*Config, error) {
    // 读取配置文件
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }

    // 解析版本
    var versionStruct struct {
        Version string `yaml:"version"`
    }
    if err := yaml.Unmarshal(data, &versionStruct); err != nil {
        return nil, err
    }

    // 如果是旧版本，进行迁移
    migrator := &ConfigMigratorV1{}
    if versionStruct.Version != "" && versionStruct.Version != "2.0" {
        config, err := migrator.MigrateFrom(versionStruct.Version, data)
        if err != nil {
            return nil, fmt.Errorf("config migration failed: %w", err)
        }
        return config, nil
    }

    // 解析当前版本配置
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    // 验证配置
    if err := config.Validate(); err != nil {
        return nil, err
    }

    return &config, nil
}

// 5. 推荐的配置文件格式
/*
version: "2.0"

# 元数据
metadata:
  app_name: "weknora"
  environment: "production"
  config_updated_at: "2025-01-14T00:00:00Z"

# 服务配置
services:
  server:
    port: 8080
    host: "0.0.0.0"
    read_timeout: 30s
    write_timeout: 30s

  database:
    url: "${DATABASE_URL}"
    max_connections: 100
    min_connections: 10

# 功能开关
features:
  rag:
    enabled: true
    strategies: ["vector", "keyword"]
    fusion: "rrf"

  web_search:
    enabled: true
    providers: ["duckduckgo", "brave"]

# 模型配置
models:
  defaults:
    chat:
      provider: "openai"
      model: "gpt-4"
      temperature: 0.7
    embedding:
      provider: "openai"
      model: "text-embedding-3-small"

  aliases:
    gpt4: "openai:gpt-4-turbo"
    claude: "anthropic:claude-3-opus-20240229"

# 工具配置
tools:
  sets:
    filesystem:
      enabled: true
      permissions: "read-only"

    web:
      enabled: true
      timeout: 30s
*/
```

---

## 三、功能模块对比

### 3.1 RAG 系统

#### WeKnora RAG

**现状**:
- 基础的向量检索
- 单一向量存储后端
- 简单的相似度匹配

#### cagent RAG

```go
// 策略接口
type Strategy interface {
    Name() string
    Initialize(ctx context.Context, docPaths []string) error
    Query(ctx context.Context, query string, limit int) ([]Result, error)
    Close() error
}

// 内置策略
var strategies = map[string]Strategy{
    "bm25":           &BM25{},
    "semantic":       &SemanticEmbeddings{},
    "chunked":        &ChunkedEmbeddings{},
}

// 融合算法
type Fusion interface {
    Fuse(results map[string][]Result) ([]Result, error)
}

var fusionAlgorithms = map[string]Fusion{
    "rrf":          &ReciprocalRankFusion{K: 60},
    "weighted":     &WeightedFusion{},
    "max_score":    &MaxScoreFusion{},
}
```

**特点**:
- 多策略检索
- 结果融合（RRF、加权、最大分数）
- 重排序支持
- 事件驱动进度跟踪

#### 优化建议: 增强 RAG 系统

```go
// 1. 检索策略接口
type RetrievalStrategy interface {
    Name() string
    Type() StrategyType
    Initialize(ctx context.Context, kbID string) error
    Query(ctx context.Context, query string, opts QueryOptions) ([]SearchResult, error)
    Close() error
}

type StrategyType string

const (
    StrategyVector   StrategyType = "vector"
    StrategyKeyword  StrategyType = "keyword"
    StrategyHybrid   StrategyType = "hybrid"
    StrategyGraph    StrategyType = "graph"
)

type QueryOptions struct {
    Limit          int
    MinScore       float64
    FilterTags     []string
    ReturnSource   bool
}

// 2. 混合检索器
type HybridRetriever struct {
    strategies []RetrievalStrategy
    fusion     FusionStrategy
    reranker   Reranker
}

func (h *HybridRetriever) Query(ctx context.Context, query string, opts QueryOptions) ([]SearchResult, error) {
    // 并行执行多个策略
    resultCh := make(chan map[string][]SearchResult, len(h.strategies))
    errCh := make(chan error, len(h.strategies))

    for _, strategy := range h.strategies {
        go func(s RetrievalStrategy) {
            results, err := s.Query(ctx, query, opts)
            if err != nil {
                errCh <- err
                return
            }
            resultCh <- map[string][]SearchResult{s.Name(): results}
        }(strategy)
    }

    // 收集结果
    allResults := make(map[string][]SearchResult)
    for i := 0; i < len(h.strategies); i++ {
        select {
        case results := <-resultCh:
            for k, v := range results {
                allResults[k] = v
            }
        case err := <-errCh:
            // 记录但继续
        }
    }

    // 融合结果
    fused, err := h.fusion.Fuse(allResults)
    if err != nil {
        return nil, err
    }

    // 重排序
    if h.reranker != nil {
        fused, err = h.reranker.Rerank(ctx, query, fused)
        if err != nil {
            return nil, err
        }
    }

    return fused, nil
}

// 3. RRF 融合实现
type ReciprocalRankFusion struct {
    K int // RRF 参数，默认 60
}

func (rrf *ReciprocalRankFusion) Fuse(results map[string][]SearchResult) ([]SearchResult, error) {
    k := rrf.K
    if k == 0 {
        k = 60
    }

    scores := make(map[string float64)
    docMap := make(map[string]*SearchResult)

    // 计算 RRF 分数
    for strategy, strategyResults := range results {
        for rank, result := range strategyResults {
            docID := result.DocumentID
            if _, exists := scores[docID]; !exists {
                scores[docID] = 0
                docMap[docID] = &result
            }
            scores[docID] += 1.0 / float64(k+rank+1)
        }
    }

    // 排序
    var sorted []SearchResult
    for docID, score := range scores {
        result := *docMap[docID]
        result.Score = score
        sorted = append(sorted, result)
    }

    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Score > sorted[j].Score
    })

    return sorted, nil
}

// 4. 重排序接口
type Reranker interface {
    Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error)
}

type CrossEncoderReranker struct {
    model *models.RerankModel
}

func (c *CrossEncoderReranker) Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error) {
    // 使用交叉编码器重排序
    for i, result := range results {
        score, err := c.model.Score(ctx, query, result.Content)
        if err != nil {
            continue
        }
        results[i].RerankScore = score
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].RerankScore > results[j].RerankScore
    })

    return results, nil
}

// 5. 事件驱动的 RAG 进度
type RAGEvent struct {
    Type      RAGEventType
    Strategy  string
    Message   string
    Progress  *RAGProgress
    Error     error
    Timestamp time.Time
}

type RAGEventType string

const (
    RAGEventIndexingStarted  RAGEventType = "indexing_started"
    RAGEventIndexingProgress RAGEventType = "indexing_progress"
    RAGEventIndexingComplete RAGEventType = "indexing_complete"
    RAGEventQueryStarted     RAGEventType = "query_started"
    RAGEventQueryComplete    RAGEventType = "query_complete"
)

type RAGProgress struct {
    Current int
    Total   int
    Percent float64
}
```

### 3.2 TUI 系统

#### cagent TUI 设计

```go
// 基于 Bubble Tea
type Component interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() string
}

// 消息组件
type MessageComponent struct {
    role     string
    content  string
    tools    []ToolCall
    status   MessageStatus
    expanded bool
}

// 工具执行组件
type ToolExecutionComponent struct {
    tool      *ToolCall
    output    string
    progress  float64
    showDetails bool
}
```

**特点**:
- 响应式组件设计
- 丰富的交互功能
- 工具执行可视化
- 状态持久化

#### 优化建议: WeKnora CLI 增强

```go
// 1. 交互式消息组件
type InteractiveMessage struct {
    Role      string
    Content   string
    ToolCalls []ToolCallInfo
    Status    ExecutionStatus
    Timestamp time.Time
}

type ExecutionStatus string

const (
    StatusPending    ExecutionStatus = "pending"
    StatusRunning    ExecutionStatus = "running"
    StatusComplete   ExecutionStatus = "complete"
    StatusFailed     ExecutionStatus = "failed"
)

type ToolCallInfo struct {
    ToolName string
    Args     map[string]interface{}
    Result   string
    Error    string
    Duration time.Duration
}

// 2. 渲染器
type MessageRenderer interface {
    Render(msg *InteractiveMessage) string
    RenderToolCall(call *ToolCallInfo) string
    RenderProgress(current, total int) string
}

type ColorMessageRenderer struct {
    useColors bool
    width     int
}

func (r *ColorMessageRenderer) Render(msg *InteractiveMessage) string {
    var sb strings.Builder

    // 角色
    role := msg.Role
    if r.useColors {
        switch role {
        case "user":
            role = color.Cyan.Sprintf("[User]")
        case "assistant":
            role = color.Green.Sprintf("[Assistant]")
        case "system":
            role = color.Yellow.Sprintf("[System]")
        }
    }

    sb.WriteString(role + " ")
    sb.WriteString(msg.Content)
    sb.WriteString("\n")

    // 工具调用
    for _, call := range msg.ToolCalls {
        sb.WriteString(r.RenderToolCall(&call))
    }

    return sb.String()
}

// 3. 进度条
type ProgressBar struct {
    total   int
    current int
    width   int
}

func (p *ProgressBar) Render() string {
    if p.total == 0 {
        return ""
    }

    percent := float64(p.current) / float64(p.total)
    filled := int(float64(p.width) * percent)

    var sb strings.Builder
    sb.WriteString("[")
    for i := 0; i < p.width; i++ {
        if i < filled {
            sb.WriteString("█")
        } else {
            sb.WriteString("░")
        }
    }
    sb.WriteString(fmt.Sprintf("] %d%%", int(percent*100)))

    return sb.String()
}

// 4. 状态管理
type TUIState struct {
    Messages         []*InteractiveMessage
    CurrentMessage   int
    ShowToolDetails  bool
    AutoScroll       bool
    SessionID        string
}

type StateManager struct {
    state  *TUIState
    history *TUIHistory
}

func (sm *StateManager) Save() error {
    // 持久化状态
    return sm.history.Save(sm.state)
}

func (sm *StateManager) Load() (*TUIState, error) {
    return sm.history.Load()
}
```

### 3.3 事件系统

#### cagent 事件系统

```go
// 统一事件格式
type Event struct {
    Type        EventType
    Strategy    string
    Message     string
    Progress    *Progress
    Error       error
    TotalTokens int64
    Cost        float64
}

// 事件流
type EventStream <-chan Event
```

#### 优化建议: 统一事件总线

```go
// 1. 事件定义
type Event struct {
    ID        string
    Type      string
    Source    string
    Timestamp time.Time
    Data      interface{}
    Error     error
    Metadata  map[string]interface{}
}

// 2. 事件总线
type EventBus struct {
    subscribers map[string][]EventHandler
    history     *EventHistory
    mutex       sync.RWMutex
    middleware  []EventMiddleware
}

type EventHandler func(ctx context.Context, event Event) error
type EventMiddleware func(next EventHandler) EventHandler

type EventHistory struct {
    events    []Event
    maxSize   int
    mutex     sync.RWMutex
}

func NewEventBus(historySize int) *EventBus {
    return &EventBus{
        subscribers: make(map[string][]EventHandler),
        history:     &EventHistory{maxSize: historySize},
        middleware:  make([]EventMiddleware, 0),
    }
}

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) func() {
    eb.mutex.Lock()
    defer eb.mutex.Unlock()

    eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)

    // 返回取消订阅函数
    return func() {
        eb.mutex.Lock()
        defer eb.mutex.Unlock()

        handlers := eb.subscribers[eventType]
        for i, h := range handlers {
            // 使用反射比较函数
            if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
                eb.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
                break
            }
        }
    }
}

func (eb *EventBus) Publish(ctx context.Context, event Event) error {
    // 设置时间戳和 ID
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now()
    }
    if event.ID == "" {
        event.ID = generateEventID()
    }

    // 记录到历史
    eb.history.Add(event)

    // 获取订阅者
    eb.mutex.RLock()
    handlers := eb.subscribers[event.Type]
    eb.mutex.RUnlock()

    // 应用中间件
    baseHandler := func(h EventHandler) EventHandler {
        return h
    }

    for _, mw := range eb.middleware {
        baseHandler = mw(baseHandler)
    }

    // 通知订阅者
    var errs []error
    for _, handler := range handlers {
        wrappedHandler := baseHandler(handler)
        if err := wrappedHandler(ctx, event); err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("event handling errors: %v", errs)
    }

    return nil
}

// 3. 事件类型定义
const (
    // Agent 事件
    EventAgentThinking    = "agent.thinking"
    EventAgentActing      = "agent.acting"
    EventAgentObserving   = "agent.observing"
    EventAgentComplete    = "agent.complete"
    EventAgentError       = "agent.error"

    // 工具事件
    EventToolExecuting    = "tool.executing"
    EventToolCompleted    = "tool.completed"
    EventToolFailed       = "tool.failed"

    // RAG 事件
    EventRAGIndexing      = "rag.indexing"
    EventRAGQuerying      = "rag.querying"
    EventRAGResult        = "rag.result"

    // 会话事件
    EventSessionStarted   = "session.started"
    EventSessionMessage   = "session.message"
    EventSessionEnded     = "session.ended"
)

// 4. 事件中间件
func LoggingMiddleware(logger logger.Logger) EventMiddleware {
    return func(next EventHandler) EventHandler {
        return func(ctx context.Context, event Event) error {
            logger.Info("Event received",
                "type", event.Type,
                "source", event.Source,
                "id", event.ID,
            )
            return next(ctx, event)
        }
    }
}

func RecoveryMiddleware() EventMiddleware {
    return func(next EventHandler) EventHandler {
        return func(ctx context.Context, event Event) error {
            defer func() {
                if r := recover(); r != nil {
                    logger.Error("Event handler panic",
                        "event", event.Type,
                        "panic", r,
                    )
                }
            }()
            return next(ctx, event)
        }
    }
}
```

### 3.4 内存管理

#### cagent 内存管理

```go
type Database interface {
    AddMemory(ctx context.Context, memory UserMemory) error
    GetMemories(ctx context.Context) ([]UserMemory, error)
    DeleteMemory(ctx context.Context, memory UserMemory) error
}
```

#### 优化建议: 统一内存管理

```go
// 1. 内存条目定义
type MemoryEntry struct {
    ID        string
    SessionID string
    Type      MemoryType
    Content   string
    Metadata  map[string]interface{}
    Timestamp time.Time
    TTL       time.Duration
}

type MemoryType string

const (
    MemoryTypeConversation MemoryType = "conversation"
    MemoryTypeToolResult   MemoryType = "tool_result"
    MemoryTypeUserInput    MemoryType = "user_input"
    MemoryTypeSystem       MemoryType = "system"
)

// 2. 内存管理器接口
type MemoryManager interface {
    // 基本操作
    Add(ctx context.Context, entry MemoryEntry) error
    Get(ctx context.Context, id string) (*MemoryEntry, error)
    Delete(ctx context.Context, id string) error

    // 查询操作
    List(ctx context.Context, opts ListOptions) ([]MemoryEntry, error)
    Search(ctx context.Context, query string, opts SearchOptions) ([]MemoryEntry, error)

    // 会话操作
    GetSessionMemories(ctx context.Context, sessionID string) ([]MemoryEntry, error)
    ClearSession(ctx context.Context, sessionID string) error

    // 维护操作
    Cleanup(ctx context.Context, opts CleanupOptions) (int, error)
}

// 3. 内存存储实现
type SQLiteMemoryStore struct {
    db *sql.DB
}

func NewSQLiteMemoryStore(dbPath string) (*SQLiteMemoryStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // 创建表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS memories (
            id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL,
            type TEXT NOT NULL,
            content TEXT NOT NULL,
            metadata TEXT,
            timestamp INTEGER NOT NULL,
            ttl INTEGER
        );

        CREATE INDEX IF NOT EXISTS idx_session_id ON memories(session_id);
        CREATE INDEX IF NOT EXISTS idx_type ON memories(type);
        CREATE INDEX IF NOT EXISTS idx_timestamp ON memories(timestamp);
    `)
    if err != nil {
        return nil, err
    }

    return &SQLiteMemoryStore{db: db}, nil
}

func (s *SQLiteMemoryStore) Add(ctx context.Context, entry MemoryEntry) error {
    metadata, err := json.Marshal(entry.Metadata)
    if err != nil {
        return err
    }

    _, err = s.db.ExecContext(ctx,
        `INSERT INTO memories (id, session_id, type, content, metadata, timestamp, ttl)
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
        entry.ID,
        entry.SessionID,
        string(entry.Type),
        entry.Content,
        string(metadata),
        entry.Timestamp.Unix(),
        int64(entry.TTL.Seconds()),
    )

    return err
}

func (s *SQLiteMemoryStore) GetSessionMemories(ctx context.Context, sessionID string) ([]MemoryEntry, error) {
    rows, err := s.db.QueryContext(ctx,
        `SELECT id, session_id, type, content, metadata, timestamp, ttl
         FROM memories
         WHERE session_id = ?
         ORDER BY timestamp ASC`,
        sessionID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entries []MemoryEntry
    for rows.Next() {
        var entry MemoryEntry
        var metadataJSON string
        var ttl int64

        err := rows.Scan(
            &entry.ID,
            &entry.SessionID,
            &entry.Type,
            &entry.Content,
            &metadataJSON,
            &entry.Timestamp,
            &ttl,
        )
        if err != nil {
            return nil, err
        }

        json.Unmarshal([]byte(metadataJSON), &entry.Metadata)
        entry.TTL = time.Duration(ttl) * time.Second

        entries = append(entries, entry)
    }

    return entries, nil
}

func (s *SQLiteMemoryStore) Cleanup(ctx context.Context, opts CleanupOptions) (int, error) {
    var query string
    var args []interface{}

    if opts.ExpireBefore.IsZero() {
        query = `DELETE FROM memories WHERE ttl > 0 AND timestamp + ttl < ?`
        args = []interface{}{time.Now().Unix()}
    } else {
        query = `DELETE FROM memories WHERE timestamp < ?`
        args = []interface{}{opts.ExpireBefore.Unix()}
    }

    result, err := s.db.ExecContext(ctx, query, args...)
    if err != nil {
        return 0, err
    }

    count, _ := result.RowsAffected()
    return int(count), nil
}
```

---

## 四、实施路径

### 4.1 优先级矩阵

| 优化项 | 优先级 | 复杂度 | 预估工期 | 依赖 |
|--------|--------|--------|----------|------|
| ToolSet 抽象 | 高 | 低 | 1周 | 无 |
| 事件总线 | 高 | 中 | 2周 | 无 |
| RAG 多策略 | 高 | 中 | 3周 | 事件总线 |
| 配置版本化 | 中 | 低 | 1周 | 无 |
| 智能体协作 | 高 | 中 | 2周 | ToolSet |
| 内存管理统一 | 中 | 低 | 1周 | 无 |
| TUI 增强 | 低 | 高 | 4周 | 事件总线 |
| MCP 深度集成 | 中 | 高 | 3周 | ToolSet |

### 4.2 分阶段实施

#### 第一阶段 (1-2月): 核心架构增强

**目标**: 建立可扩展的基础架构

1. **ToolSet 抽象** (1周)
   - 定义 ToolSet 接口
   - 实现 BaseToolSet
   - 重构现有工具

2. **事件总线** (2周)
   - 实现 EventBus
   - 定义事件类型
   - 迁移现有日志

3. **配置版本化** (1周)
   - 添加版本标识
   - 实现配置验证
   - 准备迁移机制

**交付物**:
- 增强的工具系统
- 统一事件总线
- 版本化配置

#### 第二阶段 (2-3月): 功能增强

**目标**: 增强核心功能

1. **RAG 多策略** (3周)
   - 实现策略接口
   - 添加融合算法
   - 集成重排序

2. **智能体协作** (2周)
   - 实现智能体网络
   - 添加移交工具
   - 配置协作规则

3. **内存管理统一** (1周)
   - 统一内存接口
   - 实现 SQLite 存储
   - 添加清理机制

**交付物**:
- 混合 RAG 检索
- 多智能体协作
- 统一内存管理

#### 第三阶段 (3-4月): 体验优化

**目标**: 提升用户体验

1. **TUI 增强** (4周)
   - 实现交互组件
   - 添加进度显示
   - 状态持久化

2. **MCP 深度集成** (3周)
   - ToolSet 适配 MCP
   - 实现生命周期管理
   - 支持 Elicitation

**交付物**:
- 交互式 CLI
- MCP 工具集支持

### 4.3 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 破坏现有功能 | 高 | 中 | 充分测试，渐进式迁移 |
| 性能下降 | 中 | 低 | 性能基准测试 |
| 学习曲线 | 中 | 中 | 文档和示例 |
| 兼容性问题 | 高 | 低 | 保留旧接口 |

---

## 五、总结

### 5.1 核心建议

1. **保持 WeKnora 架构优势**
   - 清晰的分层设计
   - 事件驱动架构
   - 完善的上下文管理

2. **借鉴 cagent 设计精华**
   - ToolSet 抽象
   - 多智能体协作
   - 版本化配置
   - RAG 多策略融合

3. **渐进式演进**
   - 不破坏现有功能
   - 保持向后兼容
   - 充分测试验证

### 5.2 关键收益

| 收益 | 描述 |
|------|------|
| 可扩展性 | ToolSet 和策略模式支持灵活扩展 |
| 可维护性 | 清晰的接口和职责划分 |
| 协作能力 | 多智能体支持复杂任务分解 |
| 检索效果 | 多策略融合提升 RAG 质量 |
| 用户体验 | 交互式界面和实时反馈 |

### 5.3 后续工作

1. 建立性能基准测试
2. 完善单元测试覆盖
3. 编写迁移指南
4. 更新 API 文档

---

**文档版本**: v1.0
**最后更新**: 2025-01-14
