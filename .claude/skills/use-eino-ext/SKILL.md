---
name: use-eino-ext
description: Quick reference for Eino-Ext extension components in a-old/old/eino-ext/. Use this skill when implementing AI features with ChatModel, Embedding, Retriever, Tool, Parser, Splitter. Directly use eino-ext components: openai.NewChatModel, dashscope.NewEmbedder, es8.NewRetriever, duckduckgo.NewTextSearchTool, pdf.NewPDFParser, recursive.NewSplitter. Initialize in internal/agent/model/, no wrappers, no factories.
---

使用 Eino-Ext 扩展组件 Skill。快速查找和使用 eino-ext 提供的扩展组件（OpenAI、DashScope、ES8 等）。

## 使用场景
需要使用具体的 AI 组件实现时，直接使用 eino-ext，**不自己封装**。

## 可用组件

### 1. ChatModel (对话模型)
```go
import "github.com/cloudwego/eino-ext/components/model/openai"

// 直接使用
chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
    APIKey:      apiKey,
    BaseURL:     baseURL,
    Model:       modelName,
    Temperature: &temperature,
})
```

**参考路径**：`a-old/old/eino-ext/components/model/openai/`

### 2. Embedding (向量化)
```go
import "github.com/cloudwego/eino-ext/components/embedding/dashscope"

// 直接使用
embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
    APIKey: apiKey,
    Model:  "text-embedding-v3",
})
```

**参考路径**：`a-old/old/eino-ext/components/embedding/dashscope/`

### 3. Retriever (检索器)
```go
import "github.com/cloudwego/eino-ext/components/retriever/es8"
import "github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"

// 直接使用
retriever, err := es8.NewRetriever(ctx, &es8.RetrieverConfig{
    Client:     esClient,
    Index:      indexName,
    TopK:       10,
    SearchMode: search_mode.SearchModeDenseVectorSimilarity(...),
    Embedding:  embedder,
})
```

**参考路径**：`a-old/old/eino-ext/components/retriever/es8/`

### 4. Tool (工具)
```go
import duckduckgov2 "github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"

// 直接使用
searchTool, err := duckduckgov2.NewTextSearchTool(ctx, &duckduckgov2.Config{
    ToolName:   "web_search",
    ToolDesc:   "Search the web",
    MaxResults: 10,
})
```

**参考路径**：`a-old/old/eino-ext/components/tool/duckduckgo/v2/`

### 5. Parser (文档解析)
```go
import "github.com/cloudwego/eino-ext/components/document/parser/pdf"
import "github.com/cloudwego/eino-ext/components/document/parser/docx"

// PDF 解析
pdfParser, err := pdf.NewPDFParser(ctx, &pdf.Config{ToPages: false})

// DOCX 解析
docxParser, err := docx.NewDocxParser(ctx, &docx.Config{
    IncludeHeaders: true,
    IncludeTables:  true,
})
```

**参考路径**：`a-old/old/eino-ext/components/document/parser/`

### 6. Splitter (文档分块)
```go
import "github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"

// 直接使用
splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
    ChunkSize:   512,
    OverlapSize: 50,
    Separators:  []string{"\n\n", "\n", ". "},
})
```

**参考路径**：`a-old/old/eino-ext/components/document/transformer/splitter/recursive/`

## 在 eino-show 中的使用方式

### 初始化位置
所有 Eino 组件的初始化放在 `internal/agent/model/`：

```go
// internal/agent/model/chat.go

func newChatModel(ctx context.Context, cfg *Config) (model.ChatModel, error) {
    return openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey: cfg.LLM.APIKey,
        Model:  cfg.LLM.Model,
    })
}

// internal/agent/model/embedding.go

func newEmbedder(ctx context.Context, cfg *Config) (embedding.Embedder, error) {
    return dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
        APIKey: cfg.Embedding.APIKey,
        Model:  cfg.Embedding.Model,
    })
}
```

### 传递方式
通过 Wire 依赖注入到 Biz 层：

```go
// internal/agent/factory.go
type Factory interface {
    CreateChatAgent(ctx context.Context, config *AgentConfig) (agent.Agent, error)
}

// internal/agent/react/agent.go
type ReactAgent struct {
    chatModel model.ChatModel
    tools     []einotool.BaseTool
}
```

## 架构映射

| 组件 | 位置 | 说明 |
|------|------|------|
| ChatModel 初始化 | `internal/agent/model/chat.go` | LLM 模型工厂 |
| Embedding 初始化 | `internal/agent/model/embedding.go` | 向量模型工厂 |
| Tool 封装 | `internal/agent/tool/` | 业务工具适配 |
| Agent 实现 | `internal/agent/react/`, `internal/agent/chat/` | Eino Agent |
| Agent 接口 | `internal/pkg/agent/` | 与 Biz 解耦 |

## 禁止事项
❌ **不要**在 Biz 层直接依赖 Eino
❌ **不要**创建额外的工厂封装
❌ **不要**修改 eino-ext 组件
✅ **直接**在 `internal/agent/` 中使用 eino-ext
✅ **通过** `internal/pkg/agent/` 接口与 Biz 层交互
