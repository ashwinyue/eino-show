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
func (t *GrepChunks) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "grep_chunks",
		Desc: "在知识库中搜索包含特定关键词的文档片段。适合精确查找特定术语或短语的场景。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(
			jsonschema.Reflect(grepChunksArgs{}),
		),
	}, nil
}

// grepChunksArgs 关键词搜索参数.
type grepChunksArgs struct {
	Query           string `json:"query" jsonschema:"description=要搜索的关键词或短语,required"`
	KnowledgeBaseID string `json:"knowledge_base_id" jsonschema:"description=要搜索的知识库ID,required"`
	CaseSensitive   bool   `json:"case_sensitive" jsonschema:"description=是否区分大小写，默认false"`
	TopK            int    `json:"top_k" jsonschema:"description=返回结果数量，默认10"`
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
		return fmt.Sprintf("未找到包含关键词 \"%s\" 的内容。", query)
	}

	result := fmt.Sprintf("找到 %d 条包含关键词 \"%s\" 的内容：\n\n", len(chunks), query)
	for i, chunk := range chunks {
		result += fmt.Sprintf("[%d] ", i+1)

		// 高亮关键词位置
		content := chunk.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		result += content + "\n"

		// 可选：添加来源信息
		if chunk.KnowledgeID != "" {
			result += fmt.Sprintf("    来源: 知识项 %s\n", chunk.KnowledgeID)
		}
		result += "\n"
	}

	return result
}
