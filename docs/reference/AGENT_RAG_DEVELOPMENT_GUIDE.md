# WeKnora Agent 与 RAG 开发指南

> 本文档深入解析 WeKnora 项目中 Agent 和 RAG 的实现原理，帮助开发者理解企业级 AI 应用的开发模式。

## 目录

1. [项目概览](#1-项目概览)
2. [Agent 引擎实现](#2-agent-引擎实现)
3. [RAG 检索系统](#3-rag-检索系统)
4. [工具系统架构](#4-工具系统架构)
5. [向量嵌入与检索](#5-向量嵌入与检索)
6. [关键设计模式](#6-关键设计模式)
7. [最佳实践](#7-最佳实践)

---

## 1. 项目概览

### 1.1 技术栈

```
┌─────────────────────────────────────────────────────────────┐
│                      WeKnora 技术架构                          │
├─────────────────────────────────────────────────────────────┤
│  前端: Vue 3 + TypeScript + TDesign + Vite                    │
│  后端: Go + Gin + GORM + Uber Dig                            │
│  数据库: PostgreSQL + Redis                                   │
│  向量库: Elasticsearch / Qdrant / ParadeDB                    │
│  图数据库: Neo4j (可选)                                        │
│  LLM: OpenAI / Ollama / 阿里云通义千问                         │
├─────────────────────────────────────────────────────────────┤
│  核心: ReAct Agent + Progressive RAG + 混合检索                │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 核心概念

- **ReAct Agent**: 推理-行动循环的智能体模式
- **Progressive RAG**: 渐进式检索增强生成
- **混合检索**: 向量检索 + 关键词检索的融合
- **Deep Reading**: 深度阅读机制，强制获取完整内容

---

## 2. Agent 引擎实现

### 2.1 AgentEngine 核心结构

**文件位置**: `internal/agent/engine.go`

```go
type AgentEngine struct {
    config               *types.AgentConfig      // 配置
    toolRegistry         *tools.ToolRegistry     // 工具注册表
    chatModel            chat.Chat               // LLM 模型
    eventBus             *event.EventBus         // 事件总线
    knowledgeBasesInfo   []*KnowledgeBaseInfo    // 知识库信息
    selectedDocs         []*SelectedDocumentInfo // 用户选中文档
    contextManager       interfaces.ContextManager // 上下文管理
    sessionID            string                  // 会话ID
    systemPromptTemplate string                  // 系统提示词模板
}
```

### 2.2 ReAct 循环执行流程

```
┌──────────────────────────────────────────────────────────────┐
│                    ReAct 循环 (executeLoop)                    │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐   │
│  │  Think  │───►│   Act   │───►│ Observe │───►│ Reflect │   │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘   │
│       │              │              │              │          │
│       ▼              ▼              ▼              ▼          │
│   调用LLM        执行工具        收集结果        可选反思      │
│   流式输出        工具调用        加入消息        质量评估      │
│                  结果事件        上下文                      │
│                                                               │
│  ◄──────────────────── 迭代直到完成 ──────────────────────►  │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### 2.3 核心方法解析

#### 2.3.1 Execute - 入口方法

```go
// internal/agent/engine.go:76-154
func (e *AgentEngine) Execute(
    ctx context.Context,
    sessionID, messageID, query string,
    llmContext []chat.Message,
) (*types.AgentState, error)
```

**执行步骤**:
1. 初始化 Agent 状态
2. 构建系统提示词（Progressive RAG 模板）
3. 构建消息历史
4. 获取工具定义
5. 进入主循环

#### 2.3.2 executeLoop - 主循环

```go
// internal/agent/engine.go:159-493
func (e *AgentEngine) executeLoop(...)
```

**每轮迭代**:
1. **Think**: 调用 LLM 进行思考，流式输出思考过程
2. **检查完成**: 如果 `finish_reason=stop` 且无工具调用，结束循环
3. **Act**: 执行所有工具调用，收集结果
4. **Reflect**: （可选）对工具结果进行反思
5. **Observe**: 将工具结果加入消息历史

### 2.4 事件驱动流式输出

WeKnora 使用事件总线实现流式输出：

```go
// 事件类型
event.EventAgentThought      // 思考过程
event.EventAgentToolCall     // 工具调用
event.EventAgentToolResult   // 工具结果
event.EventAgentReflection   // 反思过程
event.EventAgentFinalAnswer  // 最终答案
event.EventAgentComplete     // 完成事件
```

### 2.5 系统提示词工程

**文件位置**: `internal/agent/prompts.go`

WeKnora 采用 **Progressive RAG** 提示词模式：

```go
var ProgressiveRAGSystemPrompt = `### Role
You are WeKnora, an intelligent retrieval assistant powered by Progressive Agentic RAG.

### Critical Constraints (ABSOLUTE RULES)
1. NO Internal Knowledge
2. Mandatory Deep Read - 获取完整内容后才能回答
3. KB First, Web Second
4. Strict Plan Adherence

### Workflow: The "Reconnaissance-Plan-Execute" Cycle
Phase 1: Preliminary Reconnaissance (Mandatory Initial Step)
Phase 2: Strategic Decision & Planning
Phase 3: Disciplined Execution & Deep Reflection (The Loop)
Phase 4: Final Synthesis
...`
```

**关键特性**:
- 支持占位符替换：`{{knowledge_bases}}`, `{{web_search_status}}`, `{{current_time}}`
- 动态适配知识库列表
- 强制 Deep Reading 机制

---

## 3. RAG 检索系统

### 3.1 Progressive RAG 流程

```
┌─────────────────────────────────────────────────────────────┐
│                   Progressive RAG 流程                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Phase 1: Preliminary Reconnaissance                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  1. grep_chunks (关键词) → 获取候选文档               │   │
│  │  2. knowledge_search (语义) → 扩展相关内容             │   │
│  │  3. list_knowledge_chunks → DEEP READ 完整内容       │   │
│  │  4. 评估信息完整性                                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                          │                                  │
│                          ▼                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  信息充足？                                          │   │
│  │     YES → 直接生成答案                               │   │
│  │     NO  → 进入 Phase 2                              │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Phase 2: Strategic Planning                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  使用 todo_write 创建结构化计划                       │   │
│  │  - 拆分子任务                                        │   │
│  │  - 定义检索目标                                      │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Phase 3: Execution & Reflection (Loop)                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  对每个任务:                                         │   │
│  │  1. 执行检索                                         │   │
│  │  2. DEEP READ 完整内容                               │   │
│  │  3. 深度反思评估                                     │   │
│  │  4. 标记完成                                         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Phase 4: Final Synthesis                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  综合所有检索结果，生成最终答案                        │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 混合检索实现

**文件位置**: `internal/agent/tools/knowledge_search.go`

#### 3.2.1 并发检索

```go
// internal/agent/tools/knowledge_search.go:433-489
func (t *KnowledgeSearchTool) concurrentSearchByTargets(...) {
    // 对每个查询、每个搜索目标并发执行混合检索
    for _, query := range queries {
        for _, target := range searchTargets {
            // 并发执行 HybridSearch
            kbResults, err := t.knowledgeBaseService.HybridSearch(...)
        }
    }
}
```

#### 3.2.2 RRF (Reciprocal Rank Fusion) 融合

混合检索使用 RRF 算法融合向量和关键词检索结果：

```
RRF Score = Σ (1 / (k + rank_i))

其中:
- k = 60 (常数)
- rank_i = 结果在某个检索列表中的排名
```

#### 3.2.3 重排序 (Rerank)

支持两种重排序方式：

```go
// 1. LLM-based 重排序（优先）
func (t *KnowledgeSearchTool) rerankWithLLM(...) {
    // 使用 LLM 对每个结果打分 (0.0-1.0)
    // 批量处理，每批 15 个结果
}

// 2. 专用 Rerank 模型（备选）
func (t *KnowledgeSearchTool) rerankWithModel(...) {
    // 使用 rerank model API
}
```

#### 3.2.4 MMR 去重

```go
// internal/agent/tools/knowledge_search.go:1317-1391
func (t *KnowledgeSearchTool) applyMMR(...) {
    // MMR = λ * relevance - (1-λ) * redundancy
    // λ = 0.7 (平衡相关性和多样性)
}
```

### 3.3 工具链设计

| 工具 | 用途 | 调用时机 |
|------|------|----------|
| `grep_chunks` | 关键词精确匹配 | 初步侦查，定位核心实体 |
| `knowledge_search` | 语义向量检索 | 扩展相关内容 |
| `list_knowledge_chunks` | 读取完整内容 | Deep Reading 必须步骤 |
| `query_knowledge_graph` | 知识图谱查询 | 关系探索 |
| `web_search` | 网络搜索 | KB 不足时补充 |
| `web_fetch` | 网页抓取 | 获取具体页面内容 |

---

## 4. 工具系统架构

### 4.1 工具接口定义

```go
// internal/types/interfaces.go
type Tool interface {
    Name() string
    Description() string
    Parameters() json.RawMessage
    Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
```

### 4.2 工具注册表

**文件位置**: `internal/agent/tools/registry.go`

```go
type ToolRegistry struct {
    tools map[string]types.Tool
}

// 注册工具
func (r *ToolRegistry) RegisterTool(tool types.Tool)

// 执行工具
func (r *ToolRegistry) ExecuteTool(
    ctx context.Context,
    name string,
    args json.RawMessage,
) (*types.ToolResult, error)
```

### 4.3 工具定义示例

```go
var grepChunksTool = BaseTool{
    name: ToolGrepChunks,
    description: `Unix-style text pattern matching tool for knowledge base chunks.

## CRITICAL – Keyword Extraction Rules
This tool MUST receive **short, high-value keywords** only.
...
`,
    schema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "pattern": {"type": "array", ...},
            "knowledge_base_ids": {"type": "array", ...},
            "max_results": {"type": "integer", ...}
        },
        "required": ["pattern"]
    }`),
}
```

### 4.4 工具执行流程

```
┌─────────────────────────────────────────────────────────────┐
│                      工具执行流程                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. LLM 决定调用工具                                         │
│     ├── 生成工具名称                                         │
│     └── 生成参数 JSON                                        │
│                                                             │
│  2. AgentEngine 解析工具调用                                 │
│     ├── 解析工具名称                                         │
│     └── 解析参数                                             │
│                                                             │
│  3. ToolRegistry.ExecuteTool                                │
│     ├── 查找工具实例                                         │
│     └── 调用 Execute 方法                                    │
│                                                             │
│  4. 工具执行                                                │
│     ├── 参数验证                                             │
│     ├── 业务逻辑执行                                         │
│     └── 返回 ToolResult                                     │
│                                                             │
│  5. 结果处理                                                │
│     ├── 发送事件到 EventBus                                  │
│     └── 将结果加入消息历史                                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. 向量嵌入与检索

### 5.1 嵌入模型接口

**文件位置**: `internal/models/embedding/embedder.go`

```go
type Embedder interface {
    // 单文本嵌入
    Embed(ctx context.Context, text string) ([]float32, error)

    // 批量嵌入
    BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)

    // 获取模型信息
    GetModelName() string
    GetDimensions() int
    GetModelID() string
}
```

### 5.2 支持的嵌入模型

| 类型 | 提供商 | 示例模型 |
|------|--------|----------|
| Local | Ollama | nomic-embed-text, mxbai-embed-large |
| Remote | OpenAI | text-embedding-3-small/large |
| Remote | 阿里云 | text-embedding-v1/v2/v3 |
| Remote | Jina AI | jina-embeddings-v2/v3 |
| Remote | 火山引擎 | multimodal-embedding-* |

### 5.3 批量嵌入优化

```go
// internal/models/embedding/batch.go
type EmbedderPooler interface {
    BatchEmbedWithPool(ctx context.Context, model Embedder, texts []string) ([][]float32, error)
}
```

**优化策略**:
- 分批处理大量文本
- 连接池复用
- 并发控制

### 5.4 向量存储后端

| 后端 | 特点 | 适用场景 |
|------|------|----------|
| PostgreSQL + pgvector | 关系数据库集成 | 中小规模，事务要求高 |
| Elasticsearch | 全文+向量混合检索 | 需要复杂查询 |
| Qdrant | 专用向量数据库 | 大规模向量检索 |
| ParadeDB | PostgreSQL 扩展 | 需要关系+向量 |

---

## 6. 关键设计模式

### 6.1 依赖注入 (Uber Dig)

```go
// internal/container/container.go
container := dig.New()

// 注册服务
container.Provide(func(...) *AgentEngine { ... })

// 解析依赖
container.Invoke(func(engine *AgentEngine) {
    // 使用 engine
})
```

### 6.2 事件驱动架构

```go
// internal/event/event.go
type EventBus struct {
    subscribers map[string][]EventHandler
}

// 发送事件
eventBus.Emit(ctx, event.Event{
    Type: event.EventAgentThought,
    Data: event.AgentThoughtData{...},
})

// 订阅事件
eventBus.Subscribe(event.EventAgentThought, handler)
```

### 6.3 接口隔离

```go
// internal/types/interfaces.go
type KnowledgeBaseService interface {
    CreateKnowledgeBase(...)
    GetKnowledgeBaseByID(...)
    HybridSearch(...)  // 混合检索
    ...
}
```

### 6.4 CQRS 模式

查询和命令分离：
- Repository 层处理数据访问
- Service 层处理业务逻辑
- Handler 层处理 API 请求

---

## 7. 最佳实践

### 7.1 RAG 开发建议

1. **Deep Reading 强制执行**
   - 搜索返回 ID 后，必须调用 `list_knowledge_chunks` 获取完整内容
   - 不能仅凭搜索片段回答

2. **渐进式检索**
   - 先用 `grep_chunks` 定位核心实体
   - 再用 `knowledge_search` 扩展语义相关内容
   - 最后用 `list_knowledge_chunks` 深度阅读

3. **答案溯源**
   - 每个事实陈述必须标注来源
   - 使用 inline citation: `<kb doc="..." chunk_id="..." />`

### 7.2 Agent 开发建议

1. **工具定义规范**
   - description 必须详细说明用途和限制
   - parameters 使用 JSON Schema 定义
   - 返回结构化数据

2. **提示词工程**
   - 使用 Progressive RAG 模板
   - 明确工具使用场景
   - 强制执行 Deep Reading

3. **流式输出**
   - 使用 EventBus 发送事件
   - 支持思考、工具调用、结果分离
   - 生成唯一事件 ID 用于聚合

### 7.3 性能优化

1. **并发检索**
   - 多个查询/知识库并发执行
   - 使用 sync.WaitGroup 协调

2. **批量嵌入**
   - 批量处理文本
   - 使用连接池复用

3. **结果去重**
   - 多级去重：ID 去重、内容签名去重
   - MMR 算法减少冗余

---

## 8. 项目文件导航

| 模块 | 文件路径 | 说明 |
|------|----------|------|
| Agent 引擎 | `internal/agent/engine.go` | ReAct 循环核心实现 |
| 系统提示词 | `internal/agent/prompts.go` | Progressive RAG 提示词 |
| 工具注册 | `internal/agent/tools/registry.go` | 工具注册表 |
| 知识搜索 | `internal/agent/tools/knowledge_search.go` | 语义检索工具 |
| 关键词搜索 | `internal/agent/tools/grep_chunks.go` | 关键词匹配工具 |
| 块读取 | `internal/agent/tools/list_knowledge_chunks.go` | Deep Reading 工具 |
| 嵌入接口 | `internal/models/embedding/embedder.go` | 嵌入模型接口 |
| 嵌入实现 | `internal/models/embedding/*.go` | 各提供商嵌入实现 |
| 知识库服务 | `internal/application/service/knowledgebase.go` | 知识库业务逻辑 |

---

## 9. 学习路线图

```
┌─────────────────────────────────────────────────────────────┐
│                      学习路线图                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  第一阶段：基础理解                                          │
│  ├── 阅读 engine.go，理解 ReAct 循环                         │
│  ├── 阅读 prompts.go，理解提示词设计                         │
│  └── 阅读 registry.go，理解工具系统                          │
│                                                             │
│  第二阶段：检索机制                                          │
│  ├── 阅读 knowledge_search.go，理解混合检索                   │
│  ├── 阅读 grep_chunks.go，理解关键词匹配                     │
│  └── 阅读 list_knowledge_chunks.go，理解 Deep Reading       │
│                                                             │
│  第三阶段：向量系统                                          │
│  ├── 阅读 embedder.go，理解嵌入接口                          │
│  ├── 阅读 openai.go / ollama.go，理解具体实现                │
│  └── 理解向量存储后端集成                                    │
│                                                             │
│  第四阶段：高级特性                                          │
│  ├── 理解 RRF 融合算法                                       │
│  ├── 理解 MMR 去重算法                                       │
│  └── 理解 LLM Rerank 实现                                   │
│                                                             │
│  第五阶段：实战应用                                          │
│  ├── 添加自定义工具                                          │
│  ├── 修改提示词模板                                          │
│  └── 优化检索流程                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

**文档版本**: 1.0
**最后更新**: 2025-01-09
**维护者**: WeKnora Team
