// Package document 提供文档处理能力，集成 eino-ext 组件.
package document

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
)

// SplitterConfig 分块器配置.
type SplitterConfig struct {
	// ChunkSize 分块大小（字符数）
	ChunkSize int

	// ChunkOverlap 分块重叠大小
	ChunkOverlap int

	// Separators 分隔符列表，按优先级尝试
	Separators []string

	// KeepSeparator 是否保留分隔符
	KeepSeparator bool

	// LenFunc 计算文本长度的函数
	LenFunc func(string) int
}

// DefaultSplitterConfig 默认分块器配置.
func DefaultSplitterConfig() *SplitterConfig {
	return &SplitterConfig{
		ChunkSize:      512,
		ChunkOverlap:   50,
		Separators:     []string{"\n\n", "\n", "。", ".", "!", "?", "；", ";", "，", ","},
		KeepSeparator:  true,
		LenFunc:        func(s string) int { return len(s) },
	}
}

// recursiveSplitter 递归分块器实现.
type recursiveSplitter struct {
	cfg *SplitterConfig
}

// NewRecursiveSplitter 创建递归分块器.
func NewRecursiveSplitter(cfg *SplitterConfig) Splitter {
	if cfg == nil {
		cfg = DefaultSplitterConfig()
	}
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 512
	}
	if cfg.ChunkOverlap < 0 {
		cfg.ChunkOverlap = 0
	}
	if len(cfg.Separators) == 0 {
		cfg.Separators = []string{"\n\n", "\n", "。", ".", "!", "?"}
	}
	if cfg.LenFunc == nil {
		cfg.LenFunc = func(s string) int { return len(s) }
	}

	return &recursiveSplitter{cfg: cfg}
}

// Split 将文档分割成多个块.
func (s *recursiveSplitter) Split(ctx context.Context, doc *Document) ([]*Chunk, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	content := doc.Content
	if s.cfg.LenFunc(content) <= s.cfg.ChunkSize {
		return []*Chunk{{
			Content:  content,
			Index:    0,
			Type:     ChunkTypeText,
			StartAt:  0,
			EndAt:    s.cfg.LenFunc(content),
			Metadata: copyMetadata(doc.Metadata),
		}}, nil
	}

	// 使用递归分割算法
	chunks := s.splitText(ctx, content, s.cfg.Separators)

	// 转换为 Chunk 对象
	result := make([]*Chunk, len(chunks))
	pos := 0
	for i, text := range chunks {
		chunkLen := s.cfg.LenFunc(text)
		result[i] = &Chunk{
			Content:  text,
			Index:    i,
			Type:     ChunkTypeText,
			StartAt:  pos,
			EndAt:    pos + chunkLen,
			Metadata: copyMetadata(doc.Metadata),
		}
		pos = pos + chunkLen - s.cfg.ChunkOverlap
		if pos < 0 {
			pos = 0
		}
	}

	return result, nil
}

// splitText 递归分割文本.
func (s *recursiveSplitter) splitText(ctx context.Context, text string, separators []string) []string {
	finalChunks := make([]string, 0)

	// 找到合适的分隔符
	separator := separators[len(separators)-1]
	var newSeparators []string
	for i, sep := range separators {
		if sep == "" || strings.Contains(text, sep) {
			separator = sep
			newSeparators = separators[i+1:]
			break
		}
	}

	// 使用分隔符分割
	splits := s.split(text, separator)
	goodSplits := make([]string, 0)

	// 合并较小的分割，递归处理较大的文本
	for _, split := range splits {
		if s.cfg.LenFunc(split) < s.cfg.ChunkSize {
			goodSplits = append(goodSplits, split)
			continue
		}

		// 先合并已有的分割
		if len(goodSplits) > 0 {
			merged := s.mergeSplits(goodSplits, separator)
			finalChunks = append(finalChunks, merged...)
			goodSplits = make([]string, 0)
		}

		// 递归处理较大的文本
		if len(newSeparators) == 0 {
			finalChunks = append(finalChunks, split)
		} else {
			otherChunks := s.splitText(ctx, split, newSeparators)
			finalChunks = append(finalChunks, otherChunks...)
		}
	}

	// 合并剩余的分割
	if len(goodSplits) > 0 {
		merged := s.mergeSplits(goodSplits, separator)
		finalChunks = append(finalChunks, merged...)
	}

	return finalChunks
}

// split 使用分隔符分割文本.
func (s *recursiveSplitter) split(text string, separator string) []string {
	if separator == "" {
		return []string{text}
	}

	if s.cfg.KeepSeparator {
		// 保留分隔符在末尾
		return strings.SplitAfter(text, separator)
	}
	return strings.Split(text, separator)
}

// mergeSplits 合并较小的分割为接近 ChunkSize 的块.
func (s *recursiveSplitter) mergeSplits(splits []string, separator string) []string {
	result := make([]string, 0)
	currentParts := make([]string, 0)
	totalLen := 0

	for _, split := range splits {
		splitLen := s.cfg.LenFunc(split)

		// 计算添加此分割后的总长度
		newTotalLen := totalLen + splitLen
		if len(currentParts) > 0 && !s.cfg.KeepSeparator {
			newTotalLen += s.cfg.LenFunc(separator)
		}

		// 如果超出块大小且有已有内容，先保存当前块
		if newTotalLen > s.cfg.ChunkSize && len(currentParts) > 0 {
			result = append(result, s.joinDocs(currentParts, separator))

			// 移除开头的部分以实现重叠
			for s.shouldPop(totalLen, splitLen, s.cfg.LenFunc(separator), len(currentParts)) {
				totalLen -= s.cfg.LenFunc(currentParts[0])
				if len(currentParts) > 1 && !s.cfg.KeepSeparator {
					totalLen -= s.cfg.LenFunc(separator)
				}
				currentParts = currentParts[1:]
			}
		}

		currentParts = append(currentParts, split)
		totalLen += splitLen
		if len(currentParts) > 1 && !s.cfg.KeepSeparator {
			totalLen += s.cfg.LenFunc(separator)
		}
	}

	// 添加最后一个块
	if len(currentParts) > 0 {
		result = append(result, s.joinDocs(currentParts, separator))
	}

	return result
}

// shouldPop 判断是否需要移除开头部分以实现重叠.
func (s *recursiveSplitter) shouldPop(total, splitLen, sepLen, currentPartsLen int) bool {
	docsNeededToAddSep := 2
	if currentPartsLen < docsNeededToAddSep {
		sepLen = 0
	}

	if !s.cfg.KeepSeparator {
		return currentPartsLen > 0 && (total > s.cfg.ChunkOverlap ||
			(total+splitLen+sepLen > s.cfg.ChunkSize && total > 0))
	}
	return currentPartsLen > 0 && (total > s.cfg.ChunkOverlap ||
		(total+splitLen > s.cfg.ChunkSize && total > 0))
}

// joinDocs 连接文档部分.
func (s *recursiveSplitter) joinDocs(docs []string, separator string) string {
	if s.cfg.KeepSeparator {
		return strings.TrimSpace(strings.Join(docs, ""))
	}
	return strings.TrimSpace(strings.Join(docs, separator))
}

// copyMetadata 复制元数据.
func copyMetadata(meta map[string]any) map[string]any {
	if meta == nil {
		return nil
	}
	result := make(map[string]any, len(meta))
	for k, v := range meta {
		result[k] = v
	}
	return result
}

// ToEinoDocuments 转换为 Eino Document 格式.
func ToEinoDocuments(chunks []*Chunk) []*schema.Document {
	docs := make([]*schema.Document, len(chunks))
	for i, c := range chunks {
		docs[i] = &schema.Document{
			ID:       fmt.Sprintf("chunk_%d", c.Index),
			Content:  c.Content,
			MetaData: copyMetadata(c.Metadata),
		}
	}
	return docs
}

// FromEinoDocuments 从 Eino Document 格式转换.
func FromEinoDocuments(docs []*schema.Document, source SourceType, uri string) []*Document {
	result := make([]*Document, len(docs))
	for i, d := range docs {
		result[i] = &Document{
			ID:       d.ID,
			Content:  d.Content,
			Metadata: d.MetaData,
			Source:   source,
			URI:      uri,
		}
	}
	return result
}

// Ensure recursiveSplitter implements Splitter
var _ Splitter = (*recursiveSplitter)(nil)

// Ensure recursiveSplitter implements document.Transformer
var _ document.Transformer = (*recursiveSplitter)(nil)

// Transform 实现 Eino document.Transformer 接口.
func (s *recursiveSplitter) Transform(ctx context.Context, docs []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	ret := make([]*schema.Document, 0, len(docs))
	for _, doc := range docs {
		d := &Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Metadata: doc.MetaData,
		}

		chunks, err := s.Split(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("split document failed: %w", err)
		}

		for _, c := range chunks {
			ret = append(ret, &schema.Document{
				ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, c.Index),
				Content:  c.Content,
				MetaData: copyMetadataWithIndex(doc.MetaData, c.Index, c.StartAt, c.EndAt),
			})
		}
	}
	return ret, nil
}

// GetType 返回组件类型.
func (s *recursiveSplitter) GetType() string {
	return "RecursiveSplitter"
}

// copyMetadataWithIndex 复制元数据并添加分块信息.
func copyMetadataWithIndex(meta map[string]any, index, startAt, endAt int) map[string]any {
	result := copyMetadata(meta)
	if result == nil {
		result = make(map[string]any)
	}
	result["chunk_index"] = index
	meta["start_at"] = startAt
	meta["end_at"] = endAt
	return result
}
