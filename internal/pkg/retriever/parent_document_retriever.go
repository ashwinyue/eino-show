// Package retriever provides parent document retrieval for RAG optimization.
package retriever

import (
	"context"
	"fmt"
)

// ParentDocumentRetrieverConfig parent document retriever configuration.
type ParentDocumentRetrieverConfig struct {
	// BaseRetriever the underlying retriever for chunk search
	BaseRetriever Retriever

	// ParentStore store for retrieving parent documents
	ParentStore ParentDocumentStore

	// ChildChunkSize size of child chunks for indexing
	ChildChunkSize int

	// ParentChunkSize size of parent chunks to return
	ParentChunkSize int

	// IncludeChildren whether to include matched child chunks in metadata
	IncludeChildren bool
}

// ParentDocumentStore interface for parent document storage.
type ParentDocumentStore interface {
	// GetParentByChunkID retrieves parent document by child chunk ID
	GetParentByChunkID(ctx context.Context, chunkID string) (*ParentDocument, error)

	// GetParentByKnowledgeID retrieves parent document by knowledge ID
	GetParentByKnowledgeID(ctx context.Context, knowledgeID string) (*ParentDocument, error)

	// GetChunksByParentID retrieves all child chunks by parent ID
	GetChunksByParentID(ctx context.Context, parentID string) ([]*RetrieveResult, error)
}

// ParentDocument represents a parent document with its chunks.
type ParentDocument struct {
	// ID parent document ID
	ID string

	// KnowledgeID knowledge ID
	KnowledgeID string

	// Content full parent content
	Content string

	// ChildChunks child chunk IDs
	ChildChunks []string

	// Metadata additional metadata
	Metadata map[string]any
}

// ParentDocumentRetriever retrieves parent documents based on child chunk matches.
type ParentDocumentRetriever struct {
	cfg *ParentDocumentRetrieverConfig
}

// NewParentDocumentRetriever creates a new parent document retriever.
func NewParentDocumentRetriever(cfg *ParentDocumentRetrieverConfig) (*ParentDocumentRetriever, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.BaseRetriever == nil {
		return nil, fmt.Errorf("base retriever is required")
	}
	if cfg.ParentStore == nil {
		return nil, fmt.Errorf("parent store is required")
	}
	if cfg.ChildChunkSize <= 0 {
		cfg.ChildChunkSize = 256
	}
	if cfg.ParentChunkSize <= 0 {
		cfg.ParentChunkSize = 1024
	}

	return &ParentDocumentRetriever{cfg: cfg}, nil
}

// Retrieve executes parent document retrieval.
func (r *ParentDocumentRetriever) Retrieve(ctx context.Context, params *RetrieveParams) ([]*RetrieveResult, error) {
	// 1. Search for matching child chunks
	childResults, err := r.cfg.BaseRetriever.Retrieve(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("child chunk retrieval failed: %w", err)
	}

	if len(childResults) == 0 {
		return childResults, nil
	}

	// 2. Get unique parent documents
	parentMap := make(map[string]*parentWithScore)

	for _, child := range childResults {
		parent, err := r.cfg.ParentStore.GetParentByChunkID(ctx, child.ChunkID)
		if err != nil {
			continue
		}

		if existing, ok := parentMap[parent.ID]; ok {
			// Aggregate scores and track matched children
			existing.maxScore = max(existing.maxScore, child.Score)
			existing.totalScore += child.Score
			existing.matchCount++
			if r.cfg.IncludeChildren {
				existing.matchedChildren = append(existing.matchedChildren, child)
			}
		} else {
			pws := &parentWithScore{
				parent:     parent,
				maxScore:   child.Score,
				totalScore: child.Score,
				matchCount: 1,
			}
			if r.cfg.IncludeChildren {
				pws.matchedChildren = []*RetrieveResult{child}
			}
			parentMap[parent.ID] = pws
		}
	}

	// 3. Convert to results and sort
	results := make([]*RetrieveResult, 0, len(parentMap))

	for _, pws := range parentMap {
		result := &RetrieveResult{
			ChunkID:     pws.parent.ID,
			KnowledgeID: pws.parent.KnowledgeID,
			Content:     pws.parent.Content,
			Score:       pws.maxScore, // Use max score from children
			MatchType:   MatchTypeParentChunk,
			Metadata:    pws.parent.Metadata,
		}

		// Add matched children info to metadata
		if r.cfg.IncludeChildren && len(pws.matchedChildren) > 0 {
			if result.Metadata == nil {
				result.Metadata = make(map[string]any)
			}
			result.Metadata["matched_children_count"] = pws.matchCount
			result.Metadata["total_score"] = pws.totalScore
		}

		results = append(results, result)
	}

	// Sort by score descending
	sortResultsByScore(results)

	// 4. Apply TopK limit
	if params.TopK > 0 && len(results) > params.TopK {
		results = results[:params.TopK]
	}

	return results, nil
}

// GetType returns the retriever type.
func (r *ParentDocumentRetriever) GetType() RetrieverType {
	return "parent_document"
}

// parentWithScore tracks parent document with aggregated scores.
type parentWithScore struct {
	parent          *ParentDocument
	maxScore        float64
	totalScore      float64
	matchCount      int
	matchedChildren []*RetrieveResult
}

// max returns the maximum of two float64 values.
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Ensure ParentDocumentRetriever implements Retriever
var _ Retriever = (*ParentDocumentRetriever)(nil)

// InMemoryParentStore simple in-memory implementation of ParentDocumentStore.
type InMemoryParentStore struct {
	// parents maps parent ID to ParentDocument
	parents map[string]*ParentDocument

	// chunkToParent maps chunk ID to parent ID
	chunkToParent map[string]string

	// knowledgeToParent maps knowledge ID to parent ID
	knowledgeToParent map[string]string
}

// NewInMemoryParentStore creates an in-memory parent store.
func NewInMemoryParentStore() *InMemoryParentStore {
	return &InMemoryParentStore{
		parents:           make(map[string]*ParentDocument),
		chunkToParent:     make(map[string]string),
		knowledgeToParent: make(map[string]string),
	}
}

// AddParent adds a parent document to the store.
func (s *InMemoryParentStore) AddParent(parent *ParentDocument) {
	s.parents[parent.ID] = parent
	s.knowledgeToParent[parent.KnowledgeID] = parent.ID

	for _, chunkID := range parent.ChildChunks {
		s.chunkToParent[chunkID] = parent.ID
	}
}

// GetParentByChunkID retrieves parent by chunk ID.
func (s *InMemoryParentStore) GetParentByChunkID(ctx context.Context, chunkID string) (*ParentDocument, error) {
	parentID, ok := s.chunkToParent[chunkID]
	if !ok {
		return nil, fmt.Errorf("parent not found for chunk: %s", chunkID)
	}

	parent, ok := s.parents[parentID]
	if !ok {
		return nil, fmt.Errorf("parent not found: %s", parentID)
	}

	return parent, nil
}

// GetParentByKnowledgeID retrieves parent by knowledge ID.
func (s *InMemoryParentStore) GetParentByKnowledgeID(ctx context.Context, knowledgeID string) (*ParentDocument, error) {
	parentID, ok := s.knowledgeToParent[knowledgeID]
	if !ok {
		return nil, fmt.Errorf("parent not found for knowledge: %s", knowledgeID)
	}

	parent, ok := s.parents[parentID]
	if !ok {
		return nil, fmt.Errorf("parent not found: %s", parentID)
	}

	return parent, nil
}

// GetChunksByParentID retrieves chunks by parent ID.
func (s *InMemoryParentStore) GetChunksByParentID(ctx context.Context, parentID string) ([]*RetrieveResult, error) {
	parent, ok := s.parents[parentID]
	if !ok {
		return nil, fmt.Errorf("parent not found: %s", parentID)
	}

	results := make([]*RetrieveResult, len(parent.ChildChunks))
	for i, chunkID := range parent.ChildChunks {
		results[i] = &RetrieveResult{
			ChunkID:     chunkID,
			KnowledgeID: parent.KnowledgeID,
		}
	}

	return results, nil
}

// Ensure InMemoryParentStore implements ParentDocumentStore
var _ ParentDocumentStore = (*InMemoryParentStore)(nil)
