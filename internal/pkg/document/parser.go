// Package document 提供文档处理能力，集成 eino-ext 组件.
package document

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
)

// ParserConfig 解析器配置.
type ParserConfig struct {
	// EnablePDF 是否启用 PDF 解析
	EnablePDF bool

	// EnableWord 是否启用 Word 解析
	EnableWord bool

	// EnableHTML 是否启用 HTML 解析
	EnableHTML bool

	// FallbackToText 是否在无法解析时回退到纯文本
	FallbackToText bool
}

// DefaultParserConfig 默认解析器配置.
func DefaultParserConfig() *ParserConfig {
	return &ParserConfig{
		EnablePDF:       false,
		EnableWord:      false,
		EnableHTML:      false,
		FallbackToText:  true,
	}
}

// extParser 扩展解析器实现.
type extParser struct {
	cfg *ParserConfig
}

// NewExtParser 创建扩展解析器.
func NewExtParser(cfg *ParserConfig) Parser {
	if cfg == nil {
		cfg = DefaultParserConfig()
	}
	return &extParser{cfg: cfg}
}

// Parse 解析文档内容.
func (p *extParser) Parse(ctx context.Context, reader io.Reader) (*Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read content failed: %w", err)
	}

	// 根据文件类型选择解析器
	// 这里简化处理，直接返回文本内容
	// 实际应用中可以根据文件扩展名或 MIME 类型选择不同的解析器
	return &Document{
		Content:  string(content),
		Metadata: make(map[string]any),
	}, nil
}

// DetectFileType 检测文件类型.
func DetectFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "docx"
	case ".txt":
		return "text"
	case ".md", ".markdown":
		return "markdown"
	case ".html", ".htm":
		return "html"
	case ".json":
		return "json"
	case ".csv":
		return "csv"
	case ".xlsx", ".xls":
		return "xlsx"
	default:
		return "unknown"
	}
}

// ParseByExtension 根据文件扩展名解析文档.
func ParseByExtension(ctx context.Context, filename string, reader io.Reader) (*Document, error) {
	fileType := DetectFileType(filename)

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read content failed: %w", err)
	}

	doc := &Document{
		Content: string(content),
		Metadata: map[string]any{
			"file_name": filename,
			"file_type": fileType,
		},
	}

	return doc, nil
}

// Ensure extParser implements Parser
var _ Parser = (*extParser)(nil)

// ToEinoParser 转换为 Eino Parser 接口.
func ToEinoParser(p Parser) parser.Parser {
	return &einoParserWrapper{parser: p}
}

// einoParserWrapper Eino Parser 接口包装器.
type einoParserWrapper struct {
	parser Parser
}

// Parse 实现 Eino parser.Parser 接口.
func (w *einoParserWrapper) Parse(ctx context.Context, reader io.Reader, opts ...parser.Option) ([]*schema.Document, error) {
	doc, err := w.parser.Parse(ctx, reader)
	if err != nil {
		return nil, err
	}

	return []*schema.Document{
		{
			ID:       doc.ID,
			Content:  doc.Content,
			MetaData: doc.Metadata,
		},
	}, nil
}

// textParser 纯文本解析器.
type textParser struct{}

// NewTextParser 创建纯文本解析器.
func NewTextParser() Parser {
	return &textParser{}
}

// Parse 解析纯文本.
func (p *textParser) Parse(ctx context.Context, reader io.Reader) (*Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read text failed: %w", err)
	}

	return &Document{
		Content:  string(content),
		Metadata: make(map[string]any),
	}, nil
}

// Ensure textParser implements Parser
var _ Parser = (*textParser)(nil)

// NewEinoTextParser 创建 Eino 纯文本解析器.
func NewEinoTextParser() parser.Parser {
	return &einoParserWrapper{parser: NewTextParser()}
}
