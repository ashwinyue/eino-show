// Package retriever 提供 RAG 检索能力，对齐 WeKnora 业务逻辑.
package retriever

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
)

// VectorRetrieverConfig 向量检索器配置.
type VectorRetrieverConfig struct {
	// Embedder 向量化器
	Embedder embedding.Embedder

	// VectorStore 向量存储
	VectorStore VectorStore

	// DefaultTopK 默认返回结果数量
	DefaultTopK int

	// ScoreThreshold 相似度阈值
	ScoreThreshold float64
}

// vectorRetriever 向量检索器实现.
type vectorRetriever struct {
	cfg *VectorRetrieverConfig
}

// NewVectorRetriever 创建向量检索器.
func NewVectorRetriever(cfg *VectorRetrieverConfig) (Retriever, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.Embedder == nil {
		return nil, fmt.Errorf("embedder is required")
	}
	if cfg.VectorStore == nil {
		return nil, fmt.Errorf("vector store is required")
	}
	if cfg.DefaultTopK <= 0 {
		cfg.DefaultTopK = 5
	}

	return &vectorRetriever{cfg: cfg}, nil
}

// Retrieve 执行向量检索.
func (r *vectorRetriever) Retrieve(ctx context.Context, params *RetrieveParams) ([]*RetrieveResult, error) {
	// 1. 将查询文本转换为向量
	queryVector, err := r.embedQuery(ctx, params.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// 2. 设置 TopK
	topK := params.TopK
	if topK <= 0 {
		topK = r.cfg.DefaultTopK
	}

	// 3. 执行向量搜索
	results, err := r.cfg.VectorStore.Search(ctx, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 4. 应用分数阈值过滤
	if params.ScoreThreshold > 0 {
		results = r.filterByScore(results, params.ScoreThreshold)
	}

	// 5. 设置匹配类型
	for _, result := range results {
		result.MatchType = MatchTypeEmbedding
	}

	return results, nil
}

// GetType 获取检索器类型.
func (r *vectorRetriever) GetType() RetrieverType {
	return RetrieverTypeVector
}

// embedQuery 将查询文本转换为向量.
func (r *vectorRetriever) embedQuery(ctx context.Context, query string) ([]float32, error) {
	// 调用 Eino Embedder 接口
	vectors, err := r.cfg.Embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query failed: %w", err)
	}

	if len(vectors) != 1 {
		return nil, fmt.Errorf("invalid embedding result length: got %d, expected 1", len(vectors))
	}

	// 转换 []float64 -> []float32 (Eino Embedder 返回 float64)
	return float64ToFloat32(vectors[0]), nil
}

// float64ToFloat32 转换 float64 切片为 float32.
func float64ToFloat32(v []float64) []float32 {
	result := make([]float32, len(v))
	for i, f := range v {
		result[i] = float32(f)
	}
	return result
}

// filterByScore 根据分数阈值过滤结果.
func (r *vectorRetriever) filterByScore(results []*RetrieveResult, threshold float64) []*RetrieveResult {
	filtered := make([]*RetrieveResult, 0, len(results))
	for _, result := range results {
		if result.Score >= threshold {
			filtered = append(filtered, result)
		}
	}
	return filtered
}
