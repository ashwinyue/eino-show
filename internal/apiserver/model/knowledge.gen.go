// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package model

import (
	"time"
)

const TableNameKnowledgeBaseM = "knowledge_bases"

// KnowledgeBaseM 知识库模型
type KnowledgeBaseM struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	Name      string    `gorm:"column:name;not null;type:varchar(255);comment:知识库名称" json:"name"`
	Description string  `gorm:"column:description;type:text;comment:知识库描述" json:"description"`
	TenantID  uint64    `gorm:"column:tenant_id;not null;index:idx_tenant;comment:租户ID" json:"tenant_id"`
	// ChunkingConfig 文档分块配置（JSON 格式）
	ChunkingConfig *KnowledgeChunkingConfig `gorm:"column:chunking_config;type:jsonb;not null;comment:分块配置" json:"chunking_config"`
	// ImageProcessingConfig 图像处理配置（JSON 格式）
	ImageProcessingConfig *KnowledgeImageConfig `gorm:"column:image_processing_config;type:jsonb;not null;comment:图像处理配置" json:"image_processing_config,omitempty"`
	// EmbeddingModelID 向量模型 ID
	EmbeddingModelID string `gorm:"column:embedding_model_id;not null;type:varchar(64);comment:向量模型ID" json:"embedding_model_id"`
	// SummaryModelID 摘要模型 ID
	SummaryModelID string `gorm:"column:summary_model_id;not null;type:varchar(64);comment:摘要模型ID" json:"summary_model_id"`
	// RerankModelID 重排序模型 ID
	RerankModelID string `gorm:"column:rerank_model_id;not null;type:varchar(64);comment:重排序模型ID" json:"rerank_model_id"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:current_timestamp;comment:更新时间" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index:idx_deleted;comment:删除时间" json:"deleted_at,omitempty"`
}

// TableName KnowledgeBaseM's table name
func (*KnowledgeBaseM) TableName() string {
	return TableNameKnowledgeBaseM
}

// KnowledgeChunkingConfig 文档分块配置
type KnowledgeChunkingConfig struct {
	ChunkSize      int      `json:"chunk_size,omitempty"`      // 分块大小
	ChunkOverlap   int      `json:"chunk_overlap,omitempty"`    // 重叠大小
	SplitMarkers   []string `json:"split_markers,omitempty"`    // 分隔符
	KeepSeparator  bool     `json:"keep_separator,omitempty"`   // 是否保留分隔符
}

// KnowledgeImageConfig 图像处理配置
type KnowledgeImageConfig struct {
	EnableMultimodal bool    `json:"enable_multimodal,omitempty"` // 是否启用多模态
	ModelID           string  `json:"model_id,omitempty"`            // VLM 模型 ID
}

const TableNameKnowledgeM = "knowledges"

// KnowledgeM 知识项模型（文档）
type KnowledgeM struct {
	ID               string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID         uint64    `gorm:"column:tenant_id;not null;index:idx_tenant;comment:租户ID" json:"tenant_id"`
	KnowledgeBaseID  string    `gorm:"column:knowledge_base_id;not null;index:idx_kb;comment:知识库ID" json:"knowledge_base_id"`
	Type             string    `gorm:"column:type;not null;type:varchar(50);comment:知识类型" json:"type"`
	Title            string    `gorm:"column:title;not null;type:varchar(255);comment:标题" json:"title"`
	Description      string    `gorm:"column:description;type:text;comment:描述" json:"description"`
	Source           string    `gorm:"column:source;not null;type:varchar(128);comment:来源" json:"source"`
	ParseStatus      string    `gorm:"column:parse_status;not null;type:varchar(50);default:unprocessed;comment:解析状态" json:"parse_status"`
	EnableStatus     string    `gorm:"column:enable_status;not null;type:varchar(50);default:enabled;comment:启用状态" json:"enable_status"`
	EmbeddingModelID string    `gorm:"column:embedding_model_id;type:varchar(64);comment:向量模型ID" json:"embedding_model_id,omitempty"`
	FileName         string    `gorm:"column:file_name;type:varchar(255);comment:文件名" json:"file_name,omitempty"`
	FileType         string    `gorm:"column:file_type;type:varchar(50);comment:文件类型" json:"file_type,omitempty"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;not null;default:current_timestamp;comment:更新时间" json:"updated_at"`
}

// TableName KnowledgeM's table name
func (*KnowledgeM) TableName() string {
	return TableNameKnowledgeM
}

const TableNameChunkM = "chunks"

// ChunkM 知识分块模型（向量存储）
type ChunkM struct {
	ID               string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	TenantID         uint64    `gorm:"column:tenant_id;not null;index:idx_tenant;comment:租户ID" json:"tenant_id"`
	KnowledgeID      string    `gorm:"column:knowledge_id;not null;index:idx_knowledge;comment:知识ID" json:"knowledge_id"`
	KnowledgeBaseID  string    `gorm:"column:knowledge_base_id;not null;index:idx_kb;comment:知识库ID" json:"knowledge_base_id"`
	Content          string    `gorm:"column:content;type:text;not null;comment:分块内容" json:"content"`
	Embedding        []float32 `gorm:"column:embedding;type:vector(1536);comment:向量嵌入" json:"embedding,omitempty"`
	// 根据实际向量维度调整，PostgreSQL with pgvector uses vector(n) type
	EmbeddingModelID string    `gorm:"column:embedding_model_id;type:varchar(64);comment:向量模型ID" json:"embedding_model_id,omitempty"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;default:current_timestamp;comment:创建时间" json:"created_at"`
}

// TableName ChunkM's table name
func (*ChunkM) TableName() string {
	return TableNameChunkM
}
