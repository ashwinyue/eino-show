# WeKnora RAG 技术实现指南

> 基于项目代码分析的 RAG（检索增强生成）技术实现详解

## 目录

1. [分片策略](#1-分片策略)
2. [TopK 限制处理](#2-topk-限制处理)
3. [跨向量聚合查询](#3-跨向量聚合查询)
4. [来源追溯](#4-来源追溯)
5. [文档更新](#5-文档更新)
6. [图片和表格处理](#6-图片和表格处理)
7. [RAG 幻觉解决](#7-rag-幻觉解决)
8. [召回率提升](#8-召回率提升)
9. [查询改写](#9-查询改写)
10. [后处理机制](#10-后处理机制)
11. [ReRank 使用策略](#11-rerank-使用策略)

---

## 1. 分片策略

### 1.1 核心配置

分片配置定义在 `internal/types/knowledgebase.go:96-106`：

```go
type ChunkingConfig struct {
    ChunkSize     int      `yaml:"chunk_size"    json:"chunk_size"`       // 分片大小
    ChunkOverlap  int      `yaml:"chunk_overlap" json:"chunk_overlap"`    // 分片重叠
    Separators    []string `yaml:"separators"    json:"separators"`       // 分隔符
}
```

### 1.2 分片实现（Python DocReader）

分片核心实现在 `docreader/splitter/splitter.py`，采用递归分片策略：

```python
class TextSplitter:
    def __init__(
        self,
        chunk_size: int = 512,           # 默认分片大小
        chunk_overlap: int = 100,        # 默认重叠大小
        separators: List[str] = ["\n", "。", " "],  # 分隔符优先级
        protected_regex: List[str] = [...] # 保护模式
    ):
```

#### 关键特性

1. **保护模式（Protected Regex）**
   - 数学公式: `r"\$\$[\s\S]*?\$\$"`
   - 图片: `r"!\[.*?\]\(.*?\)"`
   - 链接: `r"\[.*?\]\(.*?\)"`
   - 表格: 完整的 Markdown 表格结构
   - 代码块: `r"```(?:\w+)[\r\n]+[^\r\n]*"`

2. **递归分片流程**
   ```
   原始文本 → 按分隔符分割 → 提取保护内容 → 合并保护内容 → 按大小合并 → 添加重叠
   ```

3. **标题跟踪（Header Tracking）**
   - 自动跟踪文档标题层级
   - 将标题前置到分片中以保持上下文

### 1.3 Chunk 类型

`internal/types/chunk.go` 定义了多种 Chunk 类型：

```go
const (
    ChunkTypeText             = "text"              // 普通文本
    ChunkTypeImageOCR         = "image_ocr"         // 图片OCR
    ChunkTypeImageCaption     = "image_caption"     // 图片描述
    ChunkTypeSummary          = "summary"           // 摘要
    ChunkTypeEntity           = "entity"            // 实体
    ChunkTypeRelationship     = "relationship"      // 关系
    ChunkTypeFAQ              = "faq"               // FAQ
    ChunkTypeWebSearch        = "web_search"        // 网络搜索
    ChunkTypeTableSummary     = "table_summary"     // 表格摘要
    ChunkTypeTableColumn      = "table_column"      // 表格列描述
)
```

### 1.4 Chunk 数据结构

```go
type Chunk struct {
    ID               string     // Chunk UUID
    KnowledgeID      string     // 所属知识ID
    KnowledgeBaseID  string     // 所属知识库ID
    Content          string     // 内容文本
    ChunkType        ChunkType  // Chunk类型
    ImageInfo        string     // 图片信息(JSON)
    Metadata         JSON       // 元数据
    ContentHash      string     // 内容哈希
    StartAt          int        // 起始位置
    EndAt            int        // 结束位置
    ParentChunkID    string     // 父Chunk ID
    NextChunkID      string     // 下一个Chunk ID
    RelationChunks   JSON       // 关联Chunk
}
```

---

## 2. TopK 限制处理

### 2.1 配置参数

`internal/config/config.go:42-48` 定义了 TopK 相关配置：

```go
type ConversationConfig struct {
    EmbeddingTopK   int     `yaml:"embedding_top_k"`   // 向量检索TopK
    RerankTopK      int     `yaml:"rerank_top_k"`      // 重排序后TopK
    VectorThreshold float64 `yaml:"vector_threshold"`  // 向量相似度阈值
}
```

### 2.2 分页支持

`internal/types/search.go` 提供了分页结构：

```go
type Pagination struct {
    Page     int `form:"page" json:"page"`
    PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"`
}
```

### 2.3 处理策略

1. **多阶段 TopK**
   - 初始向量检索：`EmbeddingTopK`（通常较大，如 20-50）
   - 重排序后：`RerankTopK`（通常较小，如 5-10）
   - 最终返回：再次筛选

2. **批量分页处理**
   - Elasticsearch/Qdrant 使用分页批量处理大量数据
   - 默认批次大小：40 条/批

---

## 3. 跨向量聚合查询

### 3.1 复合检索引擎

`internal/application/service/retriever/composite.go` 实现了并发聚合查询：

```go
type CompositeRetrieveEngine struct {
    // 支持多个检索引擎并发查询
}

func (c *CompositeRetrieveEngine) Retrieve(ctx context.Context,
    retrieveParams []types.RetrieveParams,
) ([]*types.RetrieveResult, error) {
    return concurrentRetrieve(ctx, retrieveParams,
        func(ctx context.Context, param types.RetrieveParams, ...) error {
            // 并发查询
            mu.Lock()
            *results = append(*results, result...)
            mu.Unlock()
            return nil
        },
    )
}
```

### 3.2 多知识库查询

支持三种查询模式：

```go
// 1. 按知识库查询
KnowledgeBaseIDs []string  // 多个知识库ID

// 2. 按文档查询
KnowledgeIDs []string     // 多个文档ID

// 3. 组合查询
// 同时提供 KnowledgeBaseIDs 和 KnowledgeIDs
```

### 3.3 聚合策略

1. **线程安全聚合**：使用互斥锁保护结果合并
2. **去重处理**：通过 MMR 算法进行结果去重
3. **分数归一化**：不同来源的结果分数归一化后合并

---

## 4. 来源追溯

### 4.1 多级来源跟踪

```
知识库 (KnowledgeBaseID)
    └── 文档 (KnowledgeID)
            └── 分片 (ChunkID)
                    └── 位置 (StartAt, EndAt)
```

### 4.2 元数据设计

```go
type Chunk struct {
    KnowledgeID     string `json:"knowledge_id"`      // 文档ID
    KnowledgeBaseID string `json:"knowledge_base_id"` // 知识库ID
    TagID           string `json:"tag_id"`            // 标签ID
    ImageInfo       string `json:"image_info"`        // 图片信息
    Metadata        JSON   `json:"metadata"`          // 扩展元数据
}

type SearchResult struct {
    KnowledgeID     string            `json:"knowledge_id"`
    KnowledgeTitle  string            `json:"knowledge_title"`
    KnowledgeSource string            `json:"knowledge_source"`
    Metadata        map[string]string `json:"metadata"`
}
```

### 4.3 增强内容追溯

在重排序过程中，会合并来源信息：

```go
func getEnrichedPassage(ctx context.Context, result *types.SearchResult) string {
    // 原始内容
    combinedText := result.Content

    // 添加图片描述
    for _, img := range imageInfos {
        if img.Caption != "" {
            enrichments = append(enrichments, fmt.Sprintf("图片描述: %s", img.Caption))
        }
        if img.OCRText != "" {
            enrichments = append(enrichments, fmt.Sprintf("图片文本: %s", img.OCRText))
        }
    }

    // 添加相关问题
    if len(questionStrings) > 0 {
        enrichments = append(enrichments, fmt.Sprintf("相关问题: %s", ...))
    }
}
```

---

## 5. 文档更新

### 5.1 批量索引

`internal/application/service/retriever/keywords_vector_hybrid_indexer.go`：

```go
func (v *KeywordsVectorHybridRetrieveEngineService) BatchIndex(
    ctx context.Context,
    embedder embedding.Embedder,
    indexInfoList []*types.IndexInfo,
    retrieverTypes []types.RetrieverType,
) error {
    // 1. 批量生成嵌入
    embeddings, err := embedder.BatchEmbedWithPool(ctx, embedder, contentList)

    // 2. 分批保存（batchSize=40）
    for i, indexChunk := range utils.ChunkSlice(indexInfoList, batchSize) {
        params := make(map[string]any)
        embeddingMap := make(map[string][]float32)
        for j, indexInfo := range indexChunk {
            embeddingMap[indexInfo.SourceID] = embeddings[i*batchSize+j]
        }
        params["embedding"] = embeddingMap
        err = v.indexRepository.BatchSave(ctx, indexChunk, params)
    }
}
```

### 5.2 更新策略

1. **内容哈希校验**：通过 `ContentHash` 检测变更
2. **增量更新**：仅更新变更的 Chunk
3. **批量处理**：支持批量更新操作
4. **多向量库同步**：支持 Elasticsearch、Qdrant 等多种向量库

### 5.3 删除操作

```go
// 按知识库删除
DeleteChunksByKnowledgeID(ctx, knowledgeID string)

// 按知识列表删除
DeleteByKnowledgeList(ctx, ids []string)

// 按SourceID删除向量索引
DeleteBySourceIDList(ctx, sourceIDs []string, dimensions int, kbType string)
```

---

## 6. 图片和表格处理

### 6.1 图片处理流程

#### 图片信息结构

```go
type ImageInfo struct {
    URL         string // 图片URL（COS）
    Caption     string // 图片描述
    OCRText     string // OCR提取的文本
    OriginalURL string // 原始图片URL
    Start       int    // 在文本中的开始位置
    End         int    // 在文本中的结束位置
}
```

#### 处理步骤

1. **提取图片**：从文档中识别并提取图片
2. **VLM 描述**：使用视觉语言模型生成图片描述
3. **OCR 识别**：提取图片中的文字内容
4. **索引策略**：
   - `ChunkTypeImageCaption`：图片描述单独索引
   - `ChunkTypeImageOCR`：OCR 内容单独索引
   - 与文本合并索引

### 6.2 表格处理

#### 表格 Chunk 类型

```go
ChunkTypeTableSummary = "table_summary"  // 表格摘要
ChunkTypeTableColumn  = "table_column"   // 表格列描述
```

#### 表格摘要生成

`internal/application/service/extract.go`：

```go
func (s *DataTableSummaryService) generateTableDescription(
    ctx context.Context,
    chatModel chat.Chat,
    tableName, schemaDesc, sampleDesc string,
) (string, error) {
    // 生成200-300字的表格描述
    prompt := fmt.Sprintf(tableDescriptionPromptTemplate,
        tableName, schemaDesc, sampleDesc)
    // ...
}
```

### 6.3 多模态配置

```go
type ImageProcessingConfig struct {
    ModelID string `yaml:"model_id" json:"model_id"`
}

type VLMConfig struct {
    Enabled bool   `yaml:"enabled" json:"enabled"`
    ModelID string `yaml:"model_id" json:"model_id"`
}
```

---

## 7. RAG 幻觉解决

### 7.1 多重验证机制

`internal/application/service/chat_pipline/rerank.go` 实现了复合评分：

```go
func compositeScore(sr *types.SearchResult, modelScore, baseScore float64) float64 {
    // 来源权重
    sourceWeight := 1.0
    switch strings.ToLower(sr.KnowledgeSource) {
    case "web_search":
        sourceWeight = 0.95  // Web搜索结果权重略低
    }

    // 位置先验
    positionPrior := 1.0
    if sr.StartAt >= 0 {
        positionPrior += ClampFloat(1.0-float64(sr.StartAt)/float64(sr.EndAt+1), -0.05, 0.05)
    }

    // 复合分数：60%模型分 + 30%基础分 + 10%来源权重
    composite := 0.6*modelScore + 0.3*baseScore + 0.1*sourceWeight
    composite *= positionPrior

    return ClampFloat(composite, 0, 1)
}
```

### 7.2 阈值过滤

```go
func (p *PluginRerank) rerank(...) {
    for _, result := range rerankResp {
        th := chatManage.RerankThreshold
        matchType := candidates[result.Index].MatchType

        if matchType == types.MatchTypeHistory {
            th = math.Max(th-0.1, 0.5)  // 历史匹配降低阈值
        }

        if result.RelevanceScore > th {
            rankFilter = append(rankFilter, result)
        }
    }
}
```

### 7.3 阈值退降策略

```go
// 如果无结果且阈值较高，自动降低阈值
if len(rerankResp) == 0 && originalThreshold > 0.3 {
    degradedThreshold := math.Max(originalThreshold * 0.7, 0.3)
    rerankResp = p.rerank(ctx, chatManage, rerankModel, query, passages, candidates)
}
```

### 7.4 引用增强

- 将图片描述和 OCR 文本与原始内容结合
- 添加相关问题作为补充上下文
- 保持来源链路完整，便于验证

---

## 8. 召回率提升

### 8.1 查询改写

`internal/application/service/chat_pipline/rewrite.go`：

```go
func (p *PluginRewrite) OnEvent(...) {
    // 1. 获取历史对话
    history, err := p.messageService.GetRecentMessagesBySession(
        ctx, chatManage.SessionID, 20)

    // 2. 构建对话历史
    conversationText := formatConversationHistory(historyList)

    // 3. 调用LLM进行查询改写
    response, err := rewriteModel.Chat(ctx, []chat.Message{
        {Role: "system", Content: systemContent},
        {Role: "user", Content: userContent},
    }, &chat.ChatOptions{
        Temperature:         0.3,
        MaxCompletionTokens: 50,
    })

    if response.Content != "" {
        chatManage.RewriteQuery = response.Content
    }
}
```

### 8.2 混合检索

支持三种检索类型：

```go
const (
    KeywordsRetrieverType  = "keywords"  // 关键词检索
    VectorRetrieverType    = "vector"    // 向量检索
    WebSearchRetrieverType = "websearch" // 网络搜索
)
```

### 8.3 查询扩展

```go
func expandQueries(ctx context.Context, chatManage *types.ChatManage) []string {
    // 1. 移除停用词
    keywords := extractKeywords(query)

    // 2. 提取短语
    phrases := extractPhrases(query)

    // 3. 按分割符切分
    segments := splitByDelimiters(query)

    // 4. 移除疑问词
    cleaned := removeQuestionWords(query)

    return combineAll(keywords, phrases, segments, cleaned)
}
```

### 8.4 问题生成

```go
type QuestionGenerationConfig struct {
    Enabled       bool `yaml:"enabled" json:"enabled"`
    QuestionCount int  `yaml:"question_count" json:"question_count"` // 每个chunk生成的问题数
}
```

为每个 Chunk 生成相关问题，单独索引以提升召回。

---

## 9. 查询改写

### 9.1 改写触发条件

```yaml
# internal/config/config.yaml
conversation:
  enable_rewrite: true         # 启用查询改写
  enable_query_expansion: true # 启用查询扩展
```

### 9.2 改写策略

1. **历史上下文改写**：利用对话历史补全查询中的指代
2. **本地查询扩展**：基于关键词提取的本地扩展
3. **简化查询**：移除冗余词汇，聚焦核心意图

### 9.3 改写提示词模板

```yaml
conversation:
  rewrite_prompt_system: "..."
  rewrite_prompt_user: "..."
  simplify_query_prompt: "..."
```

---

## 10. 后处理机制

### 10.1 分数归一化

`internal/searchutil/normalize.go`：

```go
func NormalizeKeywordScores[T](
    results []T,
    isKeyword func(T) bool,
    getScore func(T) float64,
    setScore func(T, float64),
    callbacks KeywordScoreCallbacks,
) {
    // 使用百分位数进行鲁棒归一化
    p5Idx := len(scores) * 5 / 100
    p95Idx := len(scores) * 95 / 100
    normalizeMin := scores[p5Idx]
    normalizeMax := scores[p95Idx]
    // ...
}
```

### 10.2 MMR 去重

`internal/application/service/chat_pipline/rerank.go`：

```go
func applyMMR(
    ctx context.Context,
    results []*types.SearchResult,
    chatManage *types.ChatManage,
    k int,
    lambda float64,  // 相关性权重
) []*types.SearchResult {
    // 预计算所有token集合
    allTokenSets := make([]map[string]struct{}, len(results))
    for i, r := range results {
        allTokenSets[i] = TokenizeSimple(getEnrichedPassage(ctx, r))
    }

    // MMR选择算法
    for len(selected) < k && len(selectedIndices) < len(results) {
        for i, r := range results {
            relevance := r.Score
            redundancy := 0.0

            // 计算与已选结果的冗余度
            for _, selTokens := range selectedTokenSets {
                sim := Jaccard(allTokenSets[i], selTokens)
                if sim > redundancy {
                    redundancy = sim
                }
            }

            mmr := lambda*relevance - (1.0-lambda)*redundancy
            // 选择MMR最高的
        }
    }
}
```

### 10.3 结果转换

```go
// Web搜索结果到标准搜索结果的转换
func ConvertWebSearchResult(webResult *types.WebSearchResult) *types.SearchResult {
    return &types.SearchResult{
        Content:         webResult.Snippet,
        KnowledgeTitle:  webResult.Title,
        KnowledgeSource: "web_search",
        Metadata: map[string]string{
            "url":   webResult.URL,
            "date":  webResult.Date,
        },
    }
}
```

---

## 11. ReRank 使用策略

### 11.1 配置参数

```yaml
conversation:
  enable_rerank: true
  rerank_top_k: 5
  rerank_threshold: 0.5
```

### 11.2 使用场景

#### 智能绕过

```go
for _, result := range chatManage.SearchResult {
    if result.MatchType == types.MatchTypeDirectLoad {
        directLoadResults = append(directLoadResults, result)
        continue  // 跳过重排序
    }
}
```

#### FAQ 优先

```go
if sr.ChunkType == string(types.ChunkTypeFAQ) {
    sr.Score = math.Min(sr.Score*chatManage.FAQScoreBoost, 1.0)
}
```

### 11.3 不能随便用的原因

1. **成本问题**：ReRank 模型调用通常按 token 计费
2. **延迟问题**：增加了额外的网络请求和处理时间
3. **效果不保证**：对某些场景（如精确匹配）可能反而降低效果
4. **配置复杂**：阈值、TopK 等参数需要根据场景调优

### 11.4 推荐策略

| 场景 | 推荐策略 | 原因 |
|------|---------|------|
| FAQ 检索 | 推荐使用 | 问题-答案匹配效果好 |
| 语义搜索 | 推荐使用 | 向量检索后重排序提升精度 |
| 关键词搜索 | 不推荐 | 精确匹配不需要重排序 |
| DirectLoad | 不使用 | 小文档直接加载，绕过重排序 |
| 实时性要求高 | 谨慎使用 | 增加延迟 |

---

## 总结

WeKnora 的 RAG 实现展现了以下工程化特点：

1. **模块化设计**：检索器、重排序器、后处理器分层清晰
2. **多模态支持**：文本、图片、表格、OCR 统一处理
3. **鲁棒性**：阈值退降、多级容错、渐进式降级
4. **可扩展性**：支持多种向量数据库和模型
5. **精细控制**：详细的元数据、分数归一化、去重策略

---

*文档生成时间：2025-01-14*
*基于代码版本：main分支*
