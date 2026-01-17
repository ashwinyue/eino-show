// Package document 提供文档处理能力，集成 eino-ext 组件.
package document

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
)

// PipelineConfig 文档处理管道配置.
type PipelineConfig struct {
	// SplitterConfig 分块器配置
	SplitterConfig *SplitterConfig

	// ParserConfig 解析器配置
	ParserConfig *ParserConfig
}

// DefaultPipelineConfig 默认管道配置.
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		SplitterConfig: DefaultSplitterConfig(),
		ParserConfig:   DefaultParserConfig(),
	}
}

// pipeline 文档处理管道实现.
type pipeline struct {
	splitter Splitter
	parser   Parser
	cfg      *PipelineConfig
}

// NewPipeline 创建文档处理管道.
func NewPipeline(cfg *PipelineConfig) Pipeline {
	if cfg == nil {
		cfg = DefaultPipelineConfig()
	}

	return &pipeline{
		splitter: NewRecursiveSplitter(cfg.SplitterConfig),
		parser:   NewExtParser(cfg.ParserConfig),
		cfg:      cfg,
	}
}

// Process 处理文档：加载 -> 解析 -> 分块.
func (p *pipeline) Process(ctx context.Context, source SourceType, uri string, reader io.Reader) ([]*Chunk, error) {
	// 1. 解析文档
	doc, err := p.parser.Parse(ctx, reader)
	if err != nil {
		return nil, fmt.Errorf("parse document failed: %w", err)
	}

	// 设置来源信息
	doc.Source = source
	doc.URI = uri
	if doc.Metadata == nil {
		doc.Metadata = make(map[string]any)
	}
	doc.Metadata["source"] = source
	doc.Metadata["uri"] = uri

	// 2. 分块处理
	chunks, err := p.splitter.Split(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("split document failed: %w", err)
	}

	return chunks, nil
}

// ProcessString 处理文本内容.
func (p *pipeline) ProcessString(ctx context.Context, content string) ([]*Chunk, error) {
	doc := &Document{
		Content:  content,
		Source:   SourceTypeRaw,
		Metadata: make(map[string]any),
	}

	chunks, err := p.splitter.Split(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("split document failed: %w", err)
	}

	return chunks, nil
}

// ProcessWithParser 使用自定义解析器处理文档.
func (p *pipeline) ProcessWithParser(ctx context.Context, parser Parser, source SourceType, uri string, reader io.Reader) ([]*Chunk, error) {
	// 使用自定义解析器
	doc, err := parser.Parse(ctx, reader)
	if err != nil {
		return nil, fmt.Errorf("parse document failed: %w", err)
	}

	doc.Source = source
	doc.URI = uri
	if doc.Metadata == nil {
		doc.Metadata = make(map[string]any)
	}
	doc.Metadata["source"] = source
	doc.Metadata["uri"] = uri

	chunks, err := p.splitter.Split(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("split document failed: %w", err)
	}

	return chunks, nil
}

// GetSplitter 获取分块器.
func (p *pipeline) GetSplitter() Splitter {
	return p.splitter
}

// GetParser 获取解析器.
func (p *pipeline) GetParser() Parser {
	return p.parser
}

// Ensure pipeline implements Pipeline
var _ Pipeline = (*pipeline)(nil)

// ToEinoTransformer 转换为 Eino Transformer 接口.
func ToEinoTransformer(s Splitter) document.Transformer {
	if et, ok := s.(document.Transformer); ok {
		return et
	}
	return nil
}

// ProcessEinoDocuments 处理 Eino Document 格式.
func ProcessEinoDocuments(ctx context.Context, splitter Splitter, docs []*schema.Document) ([]*schema.Document, error) {
	transformer := ToEinoTransformer(splitter)
	if transformer == nil {
		return nil, fmt.Errorf("splitter does not implement eino document.Transformer")
	}

	return transformer.Transform(ctx, docs)
}
