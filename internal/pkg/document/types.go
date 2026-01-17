// Package document 提供文档处理能力，集成 eino-ext 组件.
package document

import (
	"context"
	"io"
)

// SourceType 文档来源类型.
type SourceType string

const (
	SourceTypeFile SourceType = "file" // 本地文件
	SourceTypeURL  SourceType = "url"  // 远程 URL
	SourceTypeRaw  SourceType = "raw"  // 原始内容
)

// ChunkType 分块类型.
type ChunkType string

const (
	ChunkTypeText       ChunkType = "text"        // 文本分块
	ChunkTypeImageOCR   ChunkType = "image_ocr"   // 图片 OCR
	ChunkTypeSummary    ChunkType = "summary"     // 摘要
	ChunkTypeHeading    ChunkType = "heading"     // 标题
	ChunkTypeTable      ChunkType = "table"       // 表格
)

// Document 文档.
type Document struct {
	// ID 文档唯一标识
	ID string

	// Content 文档内容
	Content string

	// Metadata 元数据
	Metadata map[string]any

	// Source 来源类型
	Source SourceType

	// URI 来源 URI (文件路径、URL 等)
	URI string
}

// Chunk 文档分块.
type Chunk struct {
	// Content 分块内容
	Content string

	// Index 分块索引
	Index int

	// Type 分块类型
	Type ChunkType

	// StartAt 在原文档中的起始位置
	StartAt int

	// EndAt 在原文档中的结束位置
	EndAt int

	// Metadata 元数据
	Metadata map[string]any
}

// Parser 文档解析器接口.
type Parser interface {
	// Parse 解析文档内容，返回文本
	Parse(ctx context.Context, reader io.Reader) (*Document, error)
}

// Splitter 文档分块器接口.
type Splitter interface {
	// Split 将文档分割成多个块
	Split(ctx context.Context, doc *Document) ([]*Chunk, error)
}

// Pipeline 文档处理管道.
type Pipeline interface {
	// Process 处理文档：加载 -> 解析 -> 分块
	Process(ctx context.Context, source SourceType, uri string, reader io.Reader) ([]*Chunk, error)

	// ProcessString 处理文本内容
	ProcessString(ctx context.Context, content string) ([]*Chunk, error)
}

// Embedder 文档向量化接口.
type Embedder interface {
	// Embed 将文本转换为向量
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
