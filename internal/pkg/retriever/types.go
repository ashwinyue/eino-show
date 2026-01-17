// Package retriever 提供 RAG 检索能力，对齐 WeKnora 业务逻辑.
package retriever

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// RetrieverType 检索器类型.
type RetrieverType string

const (
	RetrieverTypeVector    RetrieverType = "vector"    // 向量检索
	RetrieverTypeKeywords  RetrieverType = "keywords"  // 关键词检索
	RetrieverTypeHybrid    RetrieverType = "hybrid"    // 混合检索
	RetrieverTypeMultiQuery RetrieverType = "multi_query" // 多查询检索
)

// MatchType 匹配类型.
type MatchType string

const (
	MatchTypeEmbedding      MatchType = "embedding"       // 向量匹配
	MatchTypeKeywords       MatchType = "keywords"        // 关键词匹配
	MatchTypeNearByChunk    MatchType = "nearby_chunk"    // 附近块匹配
	MatchTypeHistory        MatchType = "history"         // 历史记录匹配
	MatchTypeParentChunk    MatchType = "parent_chunk"    // 父块匹配
	MatchTypeRelationChunk  MatchType = "relation_chunk"  // 关系块匹配
)

// RetrieveParams 检索参数.
type RetrieveParams struct {
	// Query 查询文本
	Query string

	// KnowledgeBaseID 知识库ID
	KnowledgeBaseID string

	// RetrieverType 检索器类型
	RetrieverType RetrieverType

	// TopK 返回结果数量
	TopK int

	// ScoreThreshold 相似度阈值
	ScoreThreshold float64

	// Filters 过滤条件
	Filters map[string]any

	// IncludeMetadata 是否包含元数据
	IncludeMetadata bool

	// ChunkTypeFilter 分块类型过滤
	ChunkTypeFilter []string
}

// RetrieveResult 检索结果.
type RetrieveResult struct {
	// ChunkID 分块ID
	ChunkID string

	// KnowledgeID 知识ID
	KnowledgeID string

	// Content 分块内容
	Content string

	// Score 相似度分数
	Score float64

	// MatchType 匹配类型
	MatchType MatchType

	// Metadata 元数据
	Metadata map[string]any

	// ChunkIndex 分块索引
	ChunkIndex int

	// StartAt 在原文中的起始位置
	StartAt int

	// EndAt 在原文中的结束位置
	EndAt int
}

// ToDocument 转换为 Eino Document.
func (r *RetrieveResult) ToDocument() *schema.Document {
	doc := &schema.Document{
		ID:      r.ChunkID,
		Content: r.Content,
		MetaData: map[string]any{
			"knowledge_id": r.KnowledgeID,
			"score":        r.Score,
			"match_type":   r.MatchType,
			"chunk_index":  r.ChunkIndex,
			"start_at":     r.StartAt,
			"end_at":       r.EndAt,
		},
	}

	// 合并元数据
	for k, v := range r.Metadata {
		doc.MetaData[k] = v
	}

	return doc.WithScore(r.Score)
}

// ToDocumentWithMetadata 转换为 Eino Document，可选择是否包含元数据.
func (r *RetrieveResult) ToDocumentWithMetadata(includeMetadata bool) *schema.Document {
	doc := &schema.Document{
		ID:      r.ChunkID,
		Content: r.Content,
		MetaData: map[string]any{
			"knowledge_id": r.KnowledgeID,
			"score":        r.Score,
			"match_type":   r.MatchType,
			"chunk_index":  r.ChunkIndex,
			"start_at":     r.StartAt,
			"end_at":       r.EndAt,
		},
	}

	if includeMetadata {
		for k, v := range r.Metadata {
			doc.MetaData[k] = v
		}
	}

	return doc.WithScore(r.Score)
}

// Retriever 检索器接口.
type Retriever interface {
	// Retrieve 执行检索
	Retrieve(ctx context.Context, params *RetrieveParams) ([]*RetrieveResult, error)

	// GetType 获取检索器类型
	GetType() RetrieverType
}

// VectorStore 向量存储接口.
type VectorStore interface {
	// Index 索引文档
	Index(ctx context.Context, docs []*schema.Document) error

	// Delete 删除文档
	Delete(ctx context.Context, ids []string) error

	// Search 向量搜索
	Search(ctx context.Context, query []float32, topK int) ([]*RetrieveResult, error)
}
