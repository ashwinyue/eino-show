// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"
	"fmt"

	agentpkg "github.com/ashwinyue/eino-show/internal/pkg/agent"
	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
)

// SearchResult 搜索结果.
type SearchResult struct {
	ChunkID     string
	KnowledgeID string
	Content     string
	Score       float64
	SearchType  string
}

// SearchBiz 搜索业务接口.
type SearchBiz interface {
	// VectorSearch 向量搜索
	VectorSearch(ctx context.Context, kbID string, query string, topK int) ([]*SearchResult, error)

	// KeywordSearch 关键词搜索
	KeywordSearch(ctx context.Context, kbID string, query string, topK int, caseSensitive bool) ([]*SearchResult, error)
}

type searchBiz struct {
	store   store.IStore
	factory *agentpkg.Factory
}

// NewSearchBiz 创建 SearchBiz 实例.
func NewSearchBiz(store store.IStore, factory *agentpkg.Factory) SearchBiz {
	return &searchBiz{
		store:   store,
		factory: factory,
	}
}

// VectorSearch 向量搜索.
func (b *searchBiz) VectorSearch(ctx context.Context, kbID string, query string, topK int) ([]*SearchResult, error) {
	if b.factory == nil {
		return nil, fmt.Errorf("agent factory not available for embedding")
	}

	// TODO: 实现向量搜索
	// 1. 将 query 转换为向量
	// 2. 调用 store.Chunk().VectorSearch()

	// 临时实现：返回空结果
	return []*SearchResult{}, nil
}

// KeywordSearch 关键词搜索.
func (b *searchBiz) KeywordSearch(ctx context.Context, kbID string, query string, topK int, caseSensitive bool) ([]*SearchResult, error) {
	// 调用 Store 的关键词搜索
	chunks, err := b.store.Chunk().KeywordSearch(ctx, kbID, query, topK, caseSensitive)
	if err != nil {
		return nil, fmt.Errorf("keyword search failed: %w", err)
	}

	// 转换为 SearchResult
	results := make([]*SearchResult, 0, len(chunks))
	for _, chunk := range chunks {
		results = append(results, chunkToSearchResult(chunk, "keyword"))
	}

	return results, nil
}

// chunkToSearchResult 将 Chunk 转换为 SearchResult.
func chunkToSearchResult(chunk *model.ChunkM, searchType string) *SearchResult {
	return &SearchResult{
		ChunkID:     chunk.ID,
		KnowledgeID: chunk.KnowledgeID,
		Content:     chunk.Content,
		Score:       0.0, // TODO: 计算相关性分数
		SearchType:  searchType,
	}
}
