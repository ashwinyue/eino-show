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
func (t *KnowledgeSearch) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "knowledge_search",
		Desc: "在知识库中搜索相关信息。输入应包含查询问题和知识库ID。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(
			jsonschema.Reflect(knowledgeSearchArgs{}),
		),
	}, nil
}

// knowledgeSearchArgs 知识搜索参数.
type knowledgeSearchArgs struct {
	Query           string `json:"query" jsonschema:"description=搜索查询问题,required"`
	KnowledgeBaseID string `json:"knowledge_base_id" jsonschema:"description=要搜索的知识库ID,required"`
	TopK            int    `json:"top_k" jsonschema:"description=返回结果数量，默认5"`
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
		return "未找到相关内容。"
	}

	result := fmt.Sprintf("找到 %d 条相关内容：\n\n", len(chunks))
	for i, chunk := range chunks {
		result += fmt.Sprintf("[%d] %s\n", i+1, chunk.Content)
		// 可选：添加更多信息
		result += "\n"
	}
	return result
}
