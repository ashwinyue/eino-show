// Package retriever provides reranker for RAG optimization.
package retriever

import (
	"context"
	"sort"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
)

// RerankerConfig reranker configuration.
type RerankerConfig struct {
	// ScoreFieldKey specifies the key in metadata that stores the document score.
	// If nil, uses Document's Score() method.
	ScoreFieldKey *string

	// TopK limits the number of results after reranking (0 = no limit)
	TopK int
}

// Reranker reranks documents based on their scores.
// Implements the "primacy and recency effect" optimization for LLM context.
// Higher scored documents are placed at both beginning and end of the array.
type Reranker struct {
	cfg         *RerankerConfig
	scoreGetter func(doc *schema.Document) float64
}

// NewReranker creates a new reranker.
func NewReranker(cfg *RerankerConfig) *Reranker {
	if cfg == nil {
		cfg = &RerankerConfig{}
	}

	var getter func(doc *schema.Document) float64
	if cfg.ScoreFieldKey == nil {
		getter = func(doc *schema.Document) float64 {
			return doc.Score()
		}
	} else {
		key := *cfg.ScoreFieldKey
		getter = func(doc *schema.Document) float64 {
			if doc.MetaData == nil {
				return 0
			}
			v, ok := doc.MetaData[key]
			if !ok {
				return 0
			}
			vv, okk := v.(float64)
			if !okk {
				return 0
			}
			return vv
		}
	}

	return &Reranker{
		cfg:         cfg,
		scoreGetter: getter,
	}
}

// Rerank reorders documents for optimal LLM context processing.
// Documents with higher scores are placed at beginning and end (primacy/recency effect).
func (r *Reranker) Rerank(ctx context.Context, docs []*schema.Document) []*schema.Document {
	if len(docs) == 0 {
		return docs
	}

	// Copy and sort by score descending
	copied := make([]*schema.Document, len(docs))
	copy(copied, docs)

	sort.Slice(copied, func(i, j int) bool {
		return r.scoreGetter(copied[i]) > r.scoreGetter(copied[j])
	})

	// Apply TopK limit if configured
	if r.cfg.TopK > 0 && len(copied) > r.cfg.TopK {
		copied = copied[:r.cfg.TopK]
	}

	// Reorder: high scores at beginning and end (primacy and recency effect)
	ret := make([]*schema.Document, len(copied))
	for i, d := range copied {
		if i%2 == 0 {
			ret[i/2] = d
		} else {
			ret[len(ret)-1-i/2] = d
		}
	}

	return ret
}

// RerankResults reranks RetrieveResult list.
func (r *Reranker) RerankResults(ctx context.Context, results []*RetrieveResult) []*RetrieveResult {
	if len(results) == 0 {
		return results
	}

	// Sort by score descending
	copied := make([]*RetrieveResult, len(results))
	copy(copied, results)

	sort.Slice(copied, func(i, j int) bool {
		return copied[i].Score > copied[j].Score
	})

	// Apply TopK limit if configured
	if r.cfg.TopK > 0 && len(copied) > r.cfg.TopK {
		copied = copied[:r.cfg.TopK]
	}

	// Reorder: high scores at beginning and end
	ret := make([]*RetrieveResult, len(copied))
	for i, d := range copied {
		if i%2 == 0 {
			ret[i/2] = d
		} else {
			ret[len(ret)-1-i/2] = d
		}
	}

	return ret
}

// Transform implements Eino document.Transformer interface.
func (r *Reranker) Transform(ctx context.Context, docs []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	return r.Rerank(ctx, docs), nil
}

// GetType returns the component type.
func (r *Reranker) GetType() string {
	return "ScoreReranker"
}

// Ensure Reranker implements document.Transformer
var _ document.Transformer = (*Reranker)(nil)
