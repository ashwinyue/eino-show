// Package tool 提供 Eino 工具实现.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

// GrepChunks 关键词搜索工具.
type GrepChunks struct {
	store store.IStore
}

// NewGrepChunks 创建关键词搜索工具.
func NewGrepChunks(s store.IStore) *GrepChunks {
	return &GrepChunks{
		store: s,
	}
}

// 确保 GrepChunks 实现了 InvokableTool 接口.
var _ tool.InvokableTool = (*GrepChunks)(nil)

// Info 返回工具信息.
//
// IMPORTANT: The Desc field and parameter descriptions are sent to LLM.
// They MUST be in English for better model understanding.
// This description is copied from WeKnora to maintain compatibility.
func (t *GrepChunks) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "grep_chunks",
		Desc: `Unix-style text pattern matching tool for knowledge base chunks.

Searches for text patterns in chunk content using strict literal text matching (fixed-string search). This tool performs exact keyword lookup, not semantic search.

## Core Function
Performs exact, literal text pattern matching. Accepts a text pattern and returns chunks containing the pattern.

## CRITICAL – Keyword Extraction Rules
This tool MUST receive **short, high-value keywords** only.
**Do NOT use long phrases, sentences, or multi-word expressions.**

Provide only the **minimal core entities** extracted from user query, such as:
- Proper nouns
- Key concepts
- Domain terms
- Distinct entities that define the query

### Requirements
- Keywords should be **1–3 words maximum**
- Focus exclusively on **core entities**, not descriptions
- Avoid phrases, explanations, or anything that reduces match probability
- Preserve precision details embedded in the query (e.g., version numbers, build IDs) when they materially define the entity being matched

Long phrases dramatically reduce recall because chunks rarely contain identical wording.
Only short, atomic keywords ensure accurate matching and avoid unrelated retrieval.

## When to Use
- Extracting core entities from user input
- Exact keyword presence checks
- Fast preliminary filtering before semantic search
- Situations requiring deterministic text search`,
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(
			jsonschema.Reflect(grepChunksArgs{}),
		),
	}, nil
}

// grepChunksArgs 关键词搜索参数.
type grepChunksArgs struct {
	Query           string `json:"query" jsonschema:"description=REQUIRED: Short keyword or text pattern to search for (1-3 words recommended, exact literal matching),required"`
	KnowledgeBaseID string `json:"knowledge_base_id" jsonschema:"description=The ID of the knowledge base to search in,required"`
	CaseSensitive   bool   `json:"case_sensitive" jsonschema:"description=Whether to match case-sensitively, default is false"`
	TopK            int    `json:"top_k" jsonschema:"description=Maximum number of matching chunks to return, default is 10"`
}

// InvokableRun 执行工具.
func (t *GrepChunks) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var args grepChunksArgs
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
		args.TopK = 10
	}

	// 执行关键词搜索
	chunks, err := t.store.Chunk().KeywordSearch(ctx, args.KnowledgeBaseID, args.Query, args.TopK, args.CaseSensitive)
	if err != nil {
		return "", fmt.Errorf("failed to search chunks: %w", err)
	}

	// 在内存中过滤（如果数据库不支持）
	chunks = t.filterByKeyword(chunks, args.Query, args.CaseSensitive)
	if len(chunks) > args.TopK {
		chunks = chunks[:args.TopK]
	}

	// 格式化结果
	result := t.formatResults(chunks, args.Query)
	return result, nil
}

// filterByKeyword 在内存中过滤包含关键词的块.
func (t *GrepChunks) filterByKeyword(chunks []*model.ChunkM, keyword string, caseSensitive bool) []*model.ChunkM {
	var result []*model.ChunkM

	searchKeyword := keyword
	if !caseSensitive {
		searchKeyword = strings.ToLower(keyword)
	}

	for _, chunk := range chunks {
		content := chunk.Content
		if !caseSensitive {
			content = strings.ToLower(content)
		}

		if strings.Contains(content, searchKeyword) {
			result = append(result, chunk)
		}
	}

	return result
}

// formatResults 格式化搜索结果.
func (t *GrepChunks) formatResults(chunks []*model.ChunkM, query string) string {
	if len(chunks) == 0 {
		return fmt.Sprintf("No content found containing keyword \"%s\".", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d items containing keyword \"%s\":\n\n", len(chunks), query))
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("[%d] ", i+1))

		// Highlight keyword position
		content := chunk.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		sb.WriteString(content + "\n")

		// Optional: add source information
		if chunk.KnowledgeID != "" {
			sb.WriteString(fmt.Sprintf("    Source: Knowledge %s\n", chunk.KnowledgeID))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
