// Package retriever provides multi-query retrieval for RAG optimization.
package retriever

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MultiQueryRetrieverConfig multi-query retriever configuration.
type MultiQueryRetrieverConfig struct {
	// BaseRetriever the underlying retriever
	BaseRetriever Retriever

	// ChatModel LLM model for query generation
	ChatModel model.ChatModel

	// NumQueries number of alternative queries to generate (default: 3)
	NumQueries int

	// QueryPrompt custom prompt for generating queries
	QueryPrompt string

	// IncludeOriginal whether to include original query (default: true)
	IncludeOriginal bool
}

// MultiQueryRetriever generates multiple queries and combines results.
type MultiQueryRetriever struct {
	cfg *MultiQueryRetrieverConfig
}

// NewMultiQueryRetriever creates a new multi-query retriever.
func NewMultiQueryRetriever(cfg *MultiQueryRetrieverConfig) (*MultiQueryRetriever, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.BaseRetriever == nil {
		return nil, fmt.Errorf("base retriever is required")
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if cfg.NumQueries <= 0 {
		cfg.NumQueries = 3
	}
	if cfg.QueryPrompt == "" {
		cfg.QueryPrompt = defaultQueryPrompt
	}

	return &MultiQueryRetriever{cfg: cfg}, nil
}

const defaultQueryPrompt = `You are an AI assistant helping to generate alternative search queries.
Given the original question, generate {{num}} different versions of this question that could help retrieve relevant documents.
Each alternative should approach the question from a different angle or use different keywords.

Original question: {{question}}

Generate {{num}} alternative questions, one per line (no numbering or prefixes):`

// Retrieve executes multi-query retrieval.
func (r *MultiQueryRetriever) Retrieve(ctx context.Context, params *RetrieveParams) ([]*RetrieveResult, error) {
	// 1. Generate alternative queries
	queries, err := r.generateQueries(ctx, params.Query)
	if err != nil {
		// Fallback to single query
		return r.cfg.BaseRetriever.Retrieve(ctx, params)
	}

	// 2. Include original query if configured
	if r.cfg.IncludeOriginal {
		queries = append([]string{params.Query}, queries...)
	}

	// 3. Execute retrieval for each query
	allResults := make(map[string]*RetrieveResult)

	for _, query := range queries {
		queryParams := *params
		queryParams.Query = query

		results, err := r.cfg.BaseRetriever.Retrieve(ctx, &queryParams)
		if err != nil {
			continue
		}

		// Merge results, keeping highest score for duplicates
		for _, result := range results {
			if existing, ok := allResults[result.ChunkID]; ok {
				if result.Score > existing.Score {
					allResults[result.ChunkID] = result
				}
			} else {
				allResults[result.ChunkID] = result
			}
		}
	}

	// 4. Convert to slice and sort by score
	results := make([]*RetrieveResult, 0, len(allResults))
	for _, result := range allResults {
		results = append(results, result)
	}

	// Sort by score descending
	sortResultsByScore(results)

	// 5. Apply TopK limit
	if params.TopK > 0 && len(results) > params.TopK {
		results = results[:params.TopK]
	}

	return results, nil
}

// GetType returns the retriever type.
func (r *MultiQueryRetriever) GetType() RetrieverType {
	return RetrieverTypeMultiQuery
}

// generateQueries generates alternative queries using LLM.
func (r *MultiQueryRetriever) generateQueries(ctx context.Context, originalQuery string) ([]string, error) {
	// Build prompt
	prompt := strings.Replace(r.cfg.QueryPrompt, "{{question}}", originalQuery, -1)
	prompt = strings.Replace(prompt, "{{num}}", fmt.Sprintf("%d", r.cfg.NumQueries), -1)

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	resp, err := r.cfg.ChatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("query generation failed: %w", err)
	}

	// Parse response - each line is a query
	lines := strings.Split(resp.Content, "\n")
	queries := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Remove common prefixes like "1.", "- ", etc.
			line = trimQueryPrefix(line)
			if line != "" {
				queries = append(queries, line)
			}
		}
	}

	// Limit to configured number
	if len(queries) > r.cfg.NumQueries {
		queries = queries[:r.cfg.NumQueries]
	}

	return queries, nil
}

// trimQueryPrefix removes common prefixes from generated queries.
func trimQueryPrefix(s string) string {
	// Remove numbering like "1.", "2)", "1:"
	for i, c := range s {
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '.' || c == ')' || c == ':' {
			s = strings.TrimSpace(s[i+1:])
			break
		}
		break
	}

	// Remove bullet points
	s = strings.TrimPrefix(s, "- ")
	s = strings.TrimPrefix(s, "* ")
	s = strings.TrimPrefix(s, "• ")

	return strings.TrimSpace(s)
}

// sortResultsByScore sorts results by score in descending order.
func sortResultsByScore(results []*RetrieveResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// Ensure MultiQueryRetriever implements Retriever
var _ Retriever = (*MultiQueryRetriever)(nil)
