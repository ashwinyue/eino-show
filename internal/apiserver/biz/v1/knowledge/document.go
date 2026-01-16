// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	agentpkg "github.com/ashwinyue/eino-show/internal/pkg/agent"
	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/cloudwego/eino/components/embedding"
)

// DocumentBiz 文档业务接口.
type DocumentBiz interface {
	// UploadDocument 上传文档到知识库
	UploadDocument(ctx context.Context, kbID string, filename string, contentType string, reader io.Reader) (*v1.Knowledge, error)

	// ParseDocument 解析文档内容
	ParseDocument(ctx context.Context, knowledgeID string) error

	// DeleteDocument 删除文档
	DeleteDocument(ctx context.Context, knowledgeID string) error
}

type documentBiz struct {
	store   store.IStore
	factory *agentpkg.Factory
}

// NewDocumentBiz 创建 DocumentBiz 实例.
func NewDocumentBiz(store store.IStore, factory *agentpkg.Factory) DocumentBiz {
	return &documentBiz{
		store:   store,
		factory: factory,
	}
}

// UploadDocument 上传文档到知识库.
func (b *documentBiz) UploadDocument(ctx context.Context, kbID string, filename string, contentType string, reader io.Reader) (*v1.Knowledge, error) {
	// 验证知识库存在
	_, err := b.store.KnowledgeBase().Get(ctx, where.F("id", kbID))
	if err != nil {
		return nil, fmt.Errorf("knowledge base not found: %w", err)
	}

	// 读取内容
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// 确定文件类型
	fileType := b.detectFileType(filename, contentType)

	// 创建知识项
	knowledgeM := &model.KnowledgeM{
		TenantID:        int32(contextx.TenantID(ctx)),
		KnowledgeBaseID: kbID,
		Type:            "document",
		Title:           filename,
		FileName:        &filename,
		FileType:        &fileType,
		Source:          "upload",
		ParseStatus:     "pending",
		EnableStatus:    "enabled",
	}

	if err := b.store.Knowledge().Create(ctx, knowledgeM); err != nil {
		return nil, fmt.Errorf("failed to create knowledge: %w", err)
	}

	// 保存原始内容（可以存储到对象存储，这里简化为直接处理）
	_ = content

	// 异步解析文档
	go b.parseAndIndex(context.Background(), knowledgeM.ID, kbID, string(content))

	return modelKnowledgeToProto(knowledgeM), nil
}

// ParseDocument 解析文档内容.
func (b *documentBiz) ParseDocument(ctx context.Context, knowledgeID string) error {
	knowledgeM, err := b.store.Knowledge().Get(ctx, where.F("id", knowledgeID))
	if err != nil {
		return fmt.Errorf("knowledge not found: %w", err)
	}

	// 更新解析状态
	knowledgeM.ParseStatus = "parsing"
	if err := b.store.Knowledge().Update(ctx, knowledgeM); err != nil {
		return err
	}

	// TODO: 实现实际的解析逻辑
	knowledgeM.ParseStatus = "success"
	return b.store.Knowledge().Update(ctx, knowledgeM)
}

// DeleteDocument 删除文档.
func (b *documentBiz) DeleteDocument(ctx context.Context, knowledgeID string) error {
	// 获取知识项
	knowledgeM, err := b.store.Knowledge().Get(ctx, where.F("id", knowledgeID))
	if err != nil {
		return fmt.Errorf("knowledge not found: %w", err)
	}

	// 删除关联的分块
	chunks, err := b.store.Chunk().GetByKnowledgeID(ctx, knowledgeID)
	if err == nil {
		for _, chunk := range chunks {
			_ = b.store.Chunk().Delete(ctx, where.F("id", chunk.ID))
		}
	}

	// 删除知识项
	return b.store.Knowledge().Delete(ctx, where.F("id", knowledgeM.ID))
}

// detectFileType 检测文件类型.
func (b *documentBiz) detectFileType(filename, contentType string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "word"
	case ".txt", ".md":
		return "text"
	case ".html", ".htm":
		return "html"
	default:
		if strings.HasPrefix(contentType, "text/") {
			return "text"
		}
		return "unknown"
	}
}

// parseAndIndex 解析并索引文档内容.
func (b *documentBiz) parseAndIndex(ctx context.Context, knowledgeID, kbID, content string) {
	// 获取知识库配置
	kb, err := b.store.KnowledgeBase().Get(ctx, where.F("id", kbID))
	if err != nil {
		return
	}

	// 解析分块配置
	chunkingConfig := parseChunkingConfig(kb.ChunkingConfig)
	chunkSize := int(chunkingConfig.ChunkSize)
	if chunkSize <= 0 {
		chunkSize = 512
	}
	chunkOverlap := int(chunkingConfig.ChunkOverlap)
	if chunkOverlap <= 0 {
		chunkOverlap = 50
	}

	// 分块
	chunks := b.chunkContent(content, chunkSize, chunkOverlap)

	// 获取 Embedder
	var embedder embedding.Embedder
	if b.factory != nil {
		// TODO: 从工厂获取 Embedder
	}

	// 保存分块
	if len(chunks) > 0 && embedder != nil {
		// TODO: 生成向量并保存
		_ = embedder
	}

	// 更新知识项状态
	knowledgeM, _ := b.store.Knowledge().Get(ctx, where.F("id", knowledgeID))
	if knowledgeM != nil {
		knowledgeM.ParseStatus = "success"
		_ = b.store.Knowledge().Update(ctx, knowledgeM)
	}
}

// chunkContent 将内容分块.
func (b *documentBiz) chunkContent(content string, chunkSize, overlap int) []string {
	if len(content) <= chunkSize {
		return []string{content}
	}

	var chunks []string
	start := 0

	for start < len(content) {
		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}

		// 尝试在句号、换行符等位置分割
		if end < len(content) {
			if idx := strings.LastIndex(content[start:end], "\n\n"); idx > overlap {
				end = start + idx + 2
			} else if idx := strings.LastIndex(content[start:end], "\n"); idx > overlap {
				end = start + idx + 1
			} else if idx := strings.LastIndex(content[start:end], "。"); idx > overlap {
				end = start + idx + 3
			}
		}

		chunks = append(chunks, content[start:end])
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// parseChunkingConfig 解析分块配置.
func parseChunkingConfig(jsonStr string) *v1.KnowledgeChunkingConfig {
	// TODO: 实现 JSON 解析
	return &v1.KnowledgeChunkingConfig{
		ChunkSize:     512,
		ChunkOverlap:  50,
		SplitMarkers:  []string{"\n\n", "\n", "。"},
		KeepSeparator: true,
	}
}
