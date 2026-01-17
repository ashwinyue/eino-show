// Package retriever provides context compression for RAG optimization.
package retriever

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ContextCompressorConfig context compressor configuration.
type ContextCompressorConfig struct {
	// ChatModel LLM model for compression
	ChatModel model.ChatModel

	// MaxTokens maximum tokens for compressed context (approximate)
	MaxTokens int

	// CompressionPrompt custom compression prompt template
	CompressionPrompt string
}

// ContextCompressor compresses retrieved documents to fit LLM context window.
type ContextCompressor struct {
	cfg *ContextCompressorConfig
}

// NewContextCompressor creates a new context compressor.
func NewContextCompressor(cfg *ContextCompressorConfig) (*ContextCompressor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 2000
	}
	if cfg.CompressionPrompt == "" {
		cfg.CompressionPrompt = defaultCompressionPrompt
	}

	return &ContextCompressor{cfg: cfg}, nil
}

const defaultCompressionPrompt = `Given the following document and question, extract only the parts that are relevant to answering the question. Remove any irrelevant information while preserving key facts and context.

Question: {{question}}

Document:
{{document}}

Extract the relevant information concisely:`

// Compress compresses documents based on the query.
func (c *ContextCompressor) Compress(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) == 0 {
		return docs, nil
	}

	result := make([]*schema.Document, 0, len(docs))

	for _, doc := range docs {
		compressed, err := c.compressDocument(ctx, query, doc)
		if err != nil {
			// On error, keep original document
			result = append(result, doc)
			continue
		}
		result = append(result, compressed)
	}

	return result, nil
}

// CompressResults compresses RetrieveResult list.
func (c *ContextCompressor) CompressResults(ctx context.Context, query string, results []*RetrieveResult) ([]*RetrieveResult, error) {
	if len(results) == 0 {
		return results, nil
	}

	compressed := make([]*RetrieveResult, len(results))

	for i, r := range results {
		compressedContent, err := c.compressContent(ctx, query, r.Content)
		if err != nil {
			// On error, keep original content
			compressed[i] = r
			continue
		}

		// Create new result with compressed content
		newResult := *r
		newResult.Content = compressedContent
		compressed[i] = &newResult
	}

	return compressed, nil
}

// compressDocument compresses a single document.
func (c *ContextCompressor) compressDocument(ctx context.Context, query string, doc *schema.Document) (*schema.Document, error) {
	compressedContent, err := c.compressContent(ctx, query, doc.Content)
	if err != nil {
		return nil, err
	}

	// Create new document with compressed content
	newDoc := &schema.Document{
		ID:       doc.ID,
		Content:  compressedContent,
		MetaData: doc.MetaData,
	}

	return newDoc, nil
}

// compressContent compresses content using LLM.
func (c *ContextCompressor) compressContent(ctx context.Context, query, content string) (string, error) {
	// Build prompt
	prompt := strings.Replace(c.cfg.CompressionPrompt, "{{question}}", query, 1)
	prompt = strings.Replace(prompt, "{{document}}", content, 1)

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	resp, err := c.cfg.ChatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM compression failed: %w", err)
	}

	return resp.Content, nil
}

// CompressToMaxTokens compresses and truncates to fit token limit.
func (c *ContextCompressor) CompressToMaxTokens(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	compressed, err := c.Compress(ctx, query, docs)
	if err != nil {
		return nil, err
	}

	// Estimate tokens and truncate if needed
	result := make([]*schema.Document, 0, len(compressed))
	totalChars := 0
	maxChars := c.cfg.MaxTokens * 4 // Rough estimate: 1 token ≈ 4 chars

	for _, doc := range compressed {
		docChars := len(doc.Content)
		if totalChars+docChars > maxChars {
			// Truncate this document
			remaining := maxChars - totalChars
			if remaining > 100 {
				truncated := &schema.Document{
					ID:       doc.ID,
					Content:  doc.Content[:remaining] + "...",
					MetaData: doc.MetaData,
				}
				result = append(result, truncated)
			}
			break
		}
		result = append(result, doc)
		totalChars += docChars
	}

	return result, nil
}
