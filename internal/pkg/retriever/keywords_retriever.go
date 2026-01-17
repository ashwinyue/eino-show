// Package retriever 提供 RAG 检索能力，对齐 WeKnora 业务逻辑.
package retriever

import (
	"context"
	"fmt"
)

// KeywordsRetrieverConfig 关键词检索器配置.
type KeywordsRetrieverConfig struct {
	// StoreGetter Store 获取器
	StoreGetter KeywordsStoreGetter

	// DefaultTopK 默认返回结果数量
	DefaultTopK int

	// CaseSensitive 是否区分大小写
	CaseSensitive bool
}

// KeywordsStoreGetter 关键词 Store 接口.
type KeywordsStoreGetter interface {
	GetChunkStore() KeywordsChunkStore
}

// KeywordsChunkStore 关键词 Chunk Store 接口.
type KeywordsChunkStore interface {
	KeywordSearch(ctx context.Context, kbID string, keyword string, limit int, caseSensitive bool) (interface{}, error)
}

// keywordsRetriever 关键词检索器实现.
type keywordsRetriever struct {
	cfg *KeywordsRetrieverConfig
}

// NewKeywordsRetriever 创建关键词检索器.
func NewKeywordsRetriever(cfg *KeywordsRetrieverConfig) (Retriever, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.StoreGetter == nil {
		return nil, fmt.Errorf("store getter is required")
	}
	if cfg.DefaultTopK <= 0 {
		cfg.DefaultTopK = 10
	}

	return &keywordsRetriever{cfg: cfg}, nil
}

// Retrieve 执行关键词检索.
func (r *keywordsRetriever) Retrieve(ctx context.Context, params *RetrieveParams) ([]*RetrieveResult, error) {
	// 1. 设置 TopK
	topK := params.TopK
	if topK <= 0 {
		topK = r.cfg.DefaultTopK
	}

	// 2. 设置是否区分大小写
	caseSensitive := r.cfg.CaseSensitive

	// 3. 执行关键词搜索
	chunks, err := r.cfg.StoreGetter.GetChunkStore().KeywordSearch(
		ctx,
		params.KnowledgeBaseID,
		params.Query,
		topK,
		caseSensitive,
	)
	if err != nil {
		return nil, fmt.Errorf("keyword search failed: %w", err)
	}

	// 4. 转换为检索结果
	results := r.convertToResults(chunks)

	// 5. 设置匹配类型
	for _, result := range results {
		result.MatchType = MatchTypeKeywords
	}

	return results, nil
}

// GetType 获取检索器类型.
func (r *keywordsRetriever) GetType() RetrieverType {
	return RetrieverTypeKeywords
}

// convertToResults 转换 Chunk 列表为检索结果.
func (r *keywordsRetriever) convertToResults(chunks interface{}) []*RetrieveResult {
	if chunks == nil {
		return []*RetrieveResult{}
	}

	// 尝试类型断言并转换
	switch v := chunks.(type) {
	case []*KeywordChunkResult:
		results := make([]*RetrieveResult, len(v))
		for i, c := range v {
			results[i] = &RetrieveResult{
				ChunkID:     c.ChunkID,
				KnowledgeID: c.KnowledgeID,
				Content:     c.Content,
				Score:       c.Score,
				MatchType:   MatchTypeKeywords,
				ChunkIndex:  c.ChunkIndex,
			}
		}
		return results
	default:
		return []*RetrieveResult{}
	}
}

// KeywordChunkResult 关键词搜索结果.
type KeywordChunkResult struct {
	ChunkID     string
	KnowledgeID string
	Content     string
	Score       float64
	ChunkIndex  int
}
