// Package tool 提供 Eino 工具实现.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

// KnowledgeSearch 知识库搜索工具.
type KnowledgeSearch struct {
	store    store.IStore
	embedder embedding.Embedder
}

// NewKnowledgeSearch 创建知识库搜索工具.
func NewKnowledgeSearch(s store.IStore, embedder embedding.Embedder) *KnowledgeSearch {
	return &KnowledgeSearch{
		store:    s,
		embedder: embedder,
	}
}

// 确保 KnowledgeSearch 实现了 InvokableTool 接口.
var _ tool.InvokableTool = (*KnowledgeSearch)(nil)

// Info 返回工具信息.
//
// IMPORTANT: The Desc field and parameter descriptions are sent to LLM.
// They MUST be in English for better model understanding.
// This description is copied from WeKnora to maintain compatibility.
func (t *KnowledgeSearch) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "knowledge_search",
		Desc: `Semantic/vector search tool for retrieving knowledge by meaning, intent, and conceptual relevance.

This tool uses embeddings to understand the user's query and find semantically similar content across knowledge base chunks.

## Purpose
Designed for high-level understanding tasks, such as:
- conceptual explanations
- topic overviews
- reasoning-based information needs
- contextual or intent-driven retrieval
- queries that cannot be answered with literal keyword matching

The tool searches by MEANING rather than exact text. It identifies chunks that are conceptually relevant even when the wording differs.

## What the Tool Does NOT Do
- Does NOT perform exact keyword matching
- Does NOT search for specific named entities
- Should NOT be used for literal lookup tasks
- Should NOT receive long raw text or user messages as queries
- Should NOT be used to locate specific strings or error codes

For literal/keyword/entity search, use grep_chunks tool instead.

## Required Input Behavior
"query" must be a **short, well-formed semantic question or conceptual statement** that clearly expresses the meaning you are trying to retrieve.

The query should represent a **concept, idea, topic, explanation, or intent**, such as:
- abstract topics
- definitions
- mechanisms
- best practices
- comparisons
- how/why questions

Avoid:
- keyword lists
- raw text from user messages
- full paragraphs
- unprocessed input`,
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(
			jsonschema.Reflect(knowledgeSearchArgs{}),
		),
	}, nil
}

// knowledgeSearchArgs 知识搜索参数.
type knowledgeSearchArgs struct {
	Query           string `json:"query" jsonschema:"description=REQUIRED: A semantic question or conceptual statement expressing the meaning to retrieve (e.g., 'What is RAG?', 'How does vector search work?'),required"`
	KnowledgeBaseID string `json:"knowledge_base_id" jsonschema:"description=The ID of the knowledge base to search in,required"`
	TopK            int    `json:"top_k" jsonschema:"description=Number of results to return, default is 5"`
}

// InvokableRun 执行工具.
func (t *KnowledgeSearch) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var args struct {
		Query           string `json:"query"`
		KnowledgeBaseID string `json:"knowledge_base_id"`
		TopK            int    `json:"top_k"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 验证参数
	if args.Query == "" {
		return "", fmt.Errorf("query is required")
	}
	if args.KnowledgeBaseID == "" {
		return "", fmt.Errorf("knowledge_base_id is required")
	}
	if args.TopK <= 0 {
		args.TopK = 5
	}

	// 将查询转换为向量
	embeddings, err := t.embedder.EmbedStrings(ctx, []string{args.Query})
	if err != nil {
		return "", fmt.Errorf("failed to embed query: %w", err)
	}
	if len(embeddings) == 0 {
		return "", fmt.Errorf("empty embedding result")
	}

	// 将 []float64 转换为 []float32（PGVector 使用 float32）
	embedding32 := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		embedding32[i] = float32(v)
	}

	// 执行向量搜索
	chunks, err := t.store.Chunk().VectorSearch(ctx, args.KnowledgeBaseID, embedding32, args.TopK)
	if err != nil {
		return "", fmt.Errorf("failed to search knowledge base: %w", err)
	}

	// 格式化结果
	result := t.formatResults(chunks)
	return result, nil
}

// formatResults 格式化搜索结果.
func (t *KnowledgeSearch) formatResults(chunks []*model.ChunkM) string {
	if len(chunks) == 0 {
		return "No relevant content found in the knowledge base."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d relevant items:\n\n", len(chunks)))
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, chunk.Content))
		// Optional: add more information
		sb.WriteString("\n")
	}
	return sb.String()
}
