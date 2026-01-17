// Package retriever 提供 RAG 检索能力，对齐 WeKnora 业务逻辑.
package retriever

import (
	"context"
	"fmt"

	"github.com/ashwinyue/eino-show/pkg/store/where"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

// PGVectorConfig PostgreSQL pgvector 配置.
type PGVectorConfig struct {
	// Embedder 向量化器
	Embedder embedding.Embedder

	// VectorDimension 向量维度
	VectorDimension int

	// DistanceMetric 距离度量: cosine, euclidean, inner_product
	DistanceMetric string
}

// pgVectorStore PostgreSQL pgvector 向量存储实现.
type pgVectorStore struct {
	cfg   *PGVectorConfig
	store VectorStoreGetter
}

// VectorStoreGetter 获取 Store 的接口（避免循环依赖）.
type VectorStoreGetter interface {
	GetEmbeddingStore() EmbeddingStoreGetter
}

// EmbeddingStoreGetter Embedding Store 接口.
type EmbeddingStoreGetter interface {
	CreateBatch(ctx context.Context, objs interface{}) error
	Delete(ctx context.Context, opts *where.Options) error
	VectorSearch(ctx context.Context, kbID string, queryVector []float32, topK int) (interface{}, error)
}

// NewPGVectorStore 创建 pgvector 向量存储.
func NewPGVectorStore(cfg *PGVectorConfig, store VectorStoreGetter) VectorStore {
	return &pgVectorStore{
		cfg:   cfg,
		store: store,
	}
}

// Index 索引文档.
func (s *pgVectorStore) Index(ctx context.Context, docs []*schema.Document) error {
	if len(docs) == 0 {
		return nil
	}

	// 生成向量
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.Content
	}

	vectors, err := s.embedStrings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to embed texts: %w", err)
	}

	// TODO: 将向量存储到数据库
	// 当前模型定义中没有向量字段，需要添加
	_ = vectors

	return nil
}

// Delete 删除文档.
func (s *pgVectorStore) Delete(ctx context.Context, ids []string) error {
	// TODO: 实现删除逻辑
	return nil
}

// Search 向量搜索.
func (s *pgVectorStore) Search(ctx context.Context, query []float32, topK int) ([]*RetrieveResult, error) {
	return s.SearchWithKB(ctx, "", query, topK)
}

// SearchWithKB 在指定知识库中向量搜索.
func (s *pgVectorStore) SearchWithKB(ctx context.Context, kbID string, query []float32, topK int) ([]*RetrieveResult, error) {
	// 调用 EmbeddingStore 的向量搜索
	embeddings, err := s.store.GetEmbeddingStore().VectorSearch(ctx, kbID, query, topK)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 转换为检索结果
	return convertEmbeddingsToResults(embeddings), nil
}

// convertEmbeddingsToResults 将 embedding 结果转换为检索结果.
func convertEmbeddingsToResults(embeddings interface{}) []*RetrieveResult {
	if embeddings == nil {
		return []*RetrieveResult{}
	}

	// 尝试类型断言并转换
	switch v := embeddings.(type) {
	case []*EmbeddingResult:
		results := make([]*RetrieveResult, len(v))
		for i, e := range v {
			results[i] = &RetrieveResult{
				ChunkID:     e.ChunkID,
				KnowledgeID: e.KnowledgeID,
				Content:     e.Content,
				Score:       e.Score,
				MatchType:   MatchTypeEmbedding,
			}
		}
		return results
	default:
		return []*RetrieveResult{}
	}
}

// EmbeddingResult embedding 搜索结果.
type EmbeddingResult struct {
	ChunkID     string
	KnowledgeID string
	Content     string
	Score       float64
}

// embedStrings 批量向量化文本.
func (s *pgVectorStore) embedStrings(ctx context.Context, texts []string) ([][]float32, error) {
	if s.cfg.Embedder == nil {
		return nil, fmt.Errorf("embedder not configured")
	}

	// 调用 Eino Embedder 接口
	vectors, err := s.cfg.Embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed strings failed: %w", err)
	}

	// 转换 [][]float64 -> [][]float32
	results := make([][]float32, len(vectors))
	for i, v := range vectors {
		results[i] = float64SliceToFloat32(v)
	}

	return results, nil
}

// float64SliceToFloat32 转换 float64 切片为 float32.
func float64SliceToFloat32(v []float64) []float32 {
	result := make([]float32, len(v))
	for i, f := range v {
		result[i] = float32(f)
	}
	return result
}
