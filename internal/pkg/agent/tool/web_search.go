// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

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
func (t *WebSearch) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search",
		Desc: "在互联网上搜索最新信息。适合查找实时新闻、技术文档、时事动态等内容。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(
			jsonschema.Reflect(webSearchArgs{}),
		),
	}, nil
}

// webSearchArgs 网络搜索参数.
type webSearchArgs struct {
	Query      string `json:"query" jsonschema:"description=搜索查询问题,required"`
	MaxResults int    `json:"max_results" jsonschema:"description=返回结果数量，默认5"`
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
	// 尝试解析 DuckDuckGo 的返回结果
	// 如果解析失败，直接返回原始结果
	var searchResults []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	}

	if err := json.Unmarshal([]byte(result), &searchResults); err != nil {
		// 解析失败，返回原始结果
		return result
	}

	if len(searchResults) == 0 {
		return fmt.Sprintf("未找到关于 \"%s\" 的相关结果。", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("找到 %d 条关于 \"%s\" 的搜索结果：\n\n", len(searchResults), query))

	for i, r := range searchResults {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, r.Title))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("    摘要: %s\n", r.Snippet))
		}
		if r.URL != "" {
			sb.WriteString(fmt.Sprintf("    链接: %s\n", r.URL))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
