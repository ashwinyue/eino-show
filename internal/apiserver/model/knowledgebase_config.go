// Package model 提供知识库配置类型定义（对齐 WeKnora）.
package model

import (
	"database/sql/driver"
	"encoding/json"
)

// 知识库类型常量
const (
	KnowledgeBaseTypeDocument = "document"
	KnowledgeBaseTypeFAQ      = "faq"
)

// FAQIndexMode FAQ 索引模式
type FAQIndexMode string

const (
	// FAQIndexModeQuestionOnly 仅索引问题
	FAQIndexModeQuestionOnly FAQIndexMode = "question_only"
	// FAQIndexModeQuestionAnswer 索引问题和答案
	FAQIndexModeQuestionAnswer FAQIndexMode = "question_answer"
)

// FAQQuestionIndexMode FAQ 问题索引模式
type FAQQuestionIndexMode string

const (
	// FAQQuestionIndexModeCombined 合并索引
	FAQQuestionIndexModeCombined FAQQuestionIndexMode = "combined"
	// FAQQuestionIndexModeSeparate 分离索引
	FAQQuestionIndexModeSeparate FAQQuestionIndexMode = "separate"
)

// ChunkingConfig 文档分块配置
type ChunkingConfig struct {
	ChunkSize        int      `json:"chunk_size"`
	ChunkOverlap     int      `json:"chunk_overlap"`
	Separators       []string `json:"separators"`
	EnableMultimodal bool     `json:"enable_multimodal,omitempty"` // 已废弃，保留兼容
}

// Value 实现 driver.Valuer 接口
func (c ChunkingConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ChunkingConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// ImageProcessingConfig 图片处理配置
type ImageProcessingConfig struct {
	ModelID string `json:"model_id"`
}

// Value 实现 driver.Valuer 接口
func (c ImageProcessingConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ImageProcessingConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// VLMConfig 多模态配置
type VLMConfig struct {
	Enabled       bool   `json:"enabled"`
	ModelID       string `json:"model_id"`
	ModelName     string `json:"model_name"`     // 兼容老版本
	BaseURL       string `json:"base_url"`       // 兼容老版本
	APIKey        string `json:"api_key"`        // 兼容老版本
	InterfaceType string `json:"interface_type"` // "ollama" 或 "openai"
}

// Value 实现 driver.Valuer 接口
func (c VLMConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *VLMConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// IsEnabled 判断多模态是否启用
func (c VLMConfig) IsEnabled() bool {
	if c.Enabled && c.ModelID != "" {
		return true
	}
	if c.ModelName != "" && c.BaseURL != "" {
		return true
	}
	return false
}

// StorageConfig 存储配置
type StorageConfig struct {
	SecretID   string `json:"secret_id"`
	SecretKey  string `json:"secret_key"`
	Region     string `json:"region"`
	BucketName string `json:"bucket_name"`
	AppID      string `json:"app_id"`
	PathPrefix string `json:"path_prefix"`
	Provider   string `json:"provider"`
}

// Value 实现 driver.Valuer 接口
func (c StorageConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *StorageConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// ExtractConfig 知识图谱抽取配置
type ExtractConfig struct {
	Enabled   bool             `json:"enabled"`
	Text      string           `json:"text,omitempty"`
	Tags      []string         `json:"tags,omitempty"`
	Nodes     []*GraphNode     `json:"nodes,omitempty"`
	Relations []*GraphRelation `json:"relations,omitempty"`
}

// GraphNode 图谱节点
type GraphNode struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// GraphRelation 图谱关系
type GraphRelation struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Target      string `json:"target"`
	Description string `json:"description"`
}

// Value 实现 driver.Valuer 接口
func (c ExtractConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ExtractConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// FAQConfig FAQ 知识库配置
type FAQConfig struct {
	IndexMode         FAQIndexMode         `json:"index_mode"`
	QuestionIndexMode FAQQuestionIndexMode `json:"question_index_mode"`
}

// Value 实现 driver.Valuer 接口
func (c FAQConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *FAQConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// QuestionGenerationConfig 问题生成配置
type QuestionGenerationConfig struct {
	Enabled       bool `json:"enabled"`
	QuestionCount int  `json:"question_count"` // 每个 chunk 生成的问题数
}

// Value 实现 driver.Valuer 接口
func (c QuestionGenerationConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *QuestionGenerationConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// GetDefaultChunkingConfig 获取默认分块配置
func GetDefaultChunkingConfig() ChunkingConfig {
	return ChunkingConfig{
		ChunkSize:    500,
		ChunkOverlap: 50,
		Separators:   []string{"\n\n", "\n", "。", "！", "？", ".", "!", "?"},
	}
}

// GetDefaultFAQConfig 获取默认 FAQ 配置
func GetDefaultFAQConfig() FAQConfig {
	return FAQConfig{
		IndexMode:         FAQIndexModeQuestionAnswer,
		QuestionIndexMode: FAQQuestionIndexModeCombined,
	}
}

// GetDefaultQuestionGenerationConfig 获取默认问题生成配置
func GetDefaultQuestionGenerationConfig() QuestionGenerationConfig {
	return QuestionGenerationConfig{
		Enabled:       false,
		QuestionCount: 3,
	}
}
