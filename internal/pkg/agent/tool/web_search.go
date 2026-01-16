// Package tool 提供 Eino 工具实现.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

// WebSearch 网络搜索工具.
type WebSearch struct {
	tool tool.InvokableTool
}

// NewWebSearch 创建网络搜索工具.
func NewWebSearch(ctx context.Context) (*WebSearch, error) {
	// 创建 DuckDuckGo 搜索工具
	searchTool, err := duckduckgo.NewTextSearchTool(ctx, &duckduckgo.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create duckduckgo tool: %w", err)
	}

	return &WebSearch{
		tool: searchTool,
	}, nil
}

// 确保 WebSearch 实现了 InvokableTool 接口.
var _ tool.InvokableTool = (*WebSearch)(nil)

// Info 返回工具信息.
//
// IMPORTANT: The Desc field and parameter descriptions are sent to LLM.
// They MUST be in English for better model understanding.
// This description is copied from WeKnora to maintain compatibility.
func (t *WebSearch) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search",
		Desc: `Search the web for current information and news. This tool searches the internet to find up-to-date information that may not be in the knowledge base.

## CRITICAL - KB First Rule
**ABSOLUTE RULE**: You MUST complete KB retrieval (grep_chunks AND knowledge_search) FIRST before using this tool.
- NEVER use web_search without first trying grep_chunks and knowledge_search
- ONLY use web_search if BOTH grep_chunks AND knowledge_search return insufficient/no results
- KB retrieval is MANDATORY - you CANNOT skip it

## Features
- Real-time web search: Search the internet for current information
- RAG compression: Automatically compresses and extracts relevant content from search results
- Session-scoped caching: Maintains temporary knowledge base for session to avoid re-indexing

## Usage

**Use when**:
- **ONLY after** completing grep_chunks AND knowledge_search
- KB retrieval returned insufficient or no results
- Need current or real-time information (news, events, recent updates)
- Information is not available in knowledge bases
- Need to verify or supplement information from knowledge bases
- Searching for recent developments or trends

**Parameters**:
- query (required): Search query string

**Returns**: Web search results with title, URL, snippet, and content`,
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(
			jsonschema.Reflect(webSearchArgs{}),
		),
	}, nil
}

// webSearchArgs 网络搜索参数.
type webSearchArgs struct {
	Query      string `json:"query" jsonschema:"description=Search query string,required"`
	MaxResults int    `json:"max_results" jsonschema:"description=Number of results to return, default is 5"`
}

// InvokableRun 执行工具.
func (t *WebSearch) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var args webSearchArgs
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// 验证参数
	if args.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	if args.MaxResults <= 0 {
		args.MaxResults = 5
	}

	// 执行网络搜索
	result, err := t.tool.InvokableRun(ctx, argumentsInJSON)
	if err != nil {
		return "", fmt.Errorf("failed to search web: %w", err)
	}

	// 格式化结果
	return t.formatResults(result, args.Query), nil
}

// formatResults 格式化搜索结果.
func (t *WebSearch) formatResults(result string, query string) string {
	// Try to parse DuckDuckGo response
	// If parsing fails, return raw result
	var searchResults []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	}

	if err := json.Unmarshal([]byte(result), &searchResults); err != nil {
		// Parse failed, return raw result
		return result
	}

	if len(searchResults) == 0 {
		return fmt.Sprintf("No results found for \"%s\".", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results for \"%s\":\n\n", len(searchResults), query))

	for i, r := range searchResults {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Title))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("    Snippet: %s\n", r.Snippet))
		}
		if r.URL != "" {
			sb.WriteString(fmt.Sprintf("    URL: %s\n", r.URL))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
