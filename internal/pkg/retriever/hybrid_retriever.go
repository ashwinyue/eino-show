// Package retriever 提供 RAG 检索能力，对齐 WeKnora 业务逻辑.
package retriever

import (
	"context"
	"fmt"
	"sort"
)

// FusionConfig 融合配置.
type FusionConfig struct {
	// VectorWeight 向量检索权重
	VectorWeight float64

	// KeywordsWeight 关键词检索权重
	KeywordsWeight float64

	// RRFK RRF (Reciprocal Rank Fusion) 参数
	RRFK int
}

// HybridRetrieverConfig 混合检索器配置.
type HybridRetrieverConfig struct {
	// VectorRetriever 向量检索器
	VectorRetriever Retriever

	// KeywordsRetriever 关键词检索器
	KeywordsRetriever Retriever

	// FusionConfig 融合配置
	FusionConfig *FusionConfig
}

// hybridRetriever 混合检索器实现.
type hybridRetriever struct {
	cfg *HybridRetrieverConfig
}

// NewHybridRetriever 创建混合检索器.
func NewHybridRetriever(cfg *HybridRetrieverConfig) (Retriever, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.VectorRetriever == nil && cfg.KeywordsRetriever == nil {
		return nil, fmt.Errorf("at least one retriever is required")
	}

	// 设置默认融合配置
	if cfg.FusionConfig == nil {
		cfg.FusionConfig = &FusionConfig{
			VectorWeight:    0.7,
			KeywordsWeight:  0.3,
			RRFK:            60,
		}
	}

	return &hybridRetriever{cfg: cfg}, nil
}

// Retrieve 执行混合检索.
func (r *hybridRetriever) Retrieve(ctx context.Context, params *RetrieveParams) ([]*RetrieveResult, error) {
	// 1. 并发执行向量检索和关键词检索
	vectorResults := make([]*RetrieveResult, 0)
	keywordsResults := make([]*RetrieveResult, 0)

	var vectorErr, keywordsErr error

	if r.cfg.VectorRetriever != nil {
		vectorParams := *params
		vectorParams.RetrieverType = RetrieverTypeVector
		vectorResults, vectorErr = r.cfg.VectorRetriever.Retrieve(ctx, &vectorParams)
	}

	if r.cfg.KeywordsRetriever != nil {
		keywordsParams := *params
		keywordsParams.RetrieverType = RetrieverTypeKeywords
		keywordsResults, keywordsErr = r.cfg.KeywordsRetriever.Retrieve(ctx, &keywordsParams)
	}

	// 2. 处理错误
	if vectorErr != nil && keywordsErr != nil {
		return nil, fmt.Errorf("both retrievers failed: vector=%v, keywords=%v", vectorErr, keywordsErr)
	}

	// 3. 融合结果
	results := r.fuseResults(vectorResults, keywordsResults)

	return results, nil
}

// GetType 获取检索器类型.
func (r *hybridRetriever) GetType() RetrieverType {
	return RetrieverTypeHybrid
}

// fuseResults 融合向量检索和关键词检索的结果.
func (r *hybridRetriever) fuseResults(vectorResults, keywordsResults []*RetrieveResult) []*RetrieveResult {
	// 使用加权评分融合
	resultMap := make(map[string]*FusedResult)

	// 处理向量检索结果
	for _, result := range vectorResults {
		fused := &FusedResult{
			ChunkID: result.ChunkID,
			Result:  result,
			Score:   result.Score * r.cfg.FusionConfig.VectorWeight,
		}
		resultMap[result.ChunkID] = fused
	}

	// 处理关键词检索结果
	for _, result := range keywordsResults {
		if existing, ok := resultMap[result.ChunkID]; ok {
			// 合并分数
			existing.Score += result.Score * r.cfg.FusionConfig.KeywordsWeight
		} else {
			fused := &FusedResult{
				ChunkID: result.ChunkID,
				Result:  result,
				Score:   result.Score * r.cfg.FusionConfig.KeywordsWeight,
			}
			resultMap[result.ChunkID] = fused
		}
	}

	// 转换为列表并排序
	results := make([]*RetrieveResult, 0, len(resultMap))
	for _, fused := range resultMap {
		fused.Result.Score = fused.Score
		results = append(results, fused.Result)
	}

	// 按分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// FusedResult 融合结果.
type FusedResult struct {
	ChunkID string
	Result  *RetrieveResult
	Score   float64
}
