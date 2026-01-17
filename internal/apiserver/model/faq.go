package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameFAQEntryM = "faq_entries"

// FAQEntryM FAQ 条目模型
// FAQ 条目存储在独立表中，与 Chunk 关联
type FAQEntryM struct {
	ID                int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID          int32          `gorm:"column:tenant_id;not null" json:"tenant_id"`
	KnowledgeBaseID   string         `gorm:"column:knowledge_base_id;not null" json:"knowledge_base_id"`
	KnowledgeID       string         `gorm:"column:knowledge_id" json:"knowledge_id"`
	ChunkID           string         `gorm:"column:chunk_id" json:"chunk_id"`
	TagID             *int64         `gorm:"column:tag_id" json:"tag_id"`
	StandardQuestion  string         `gorm:"column:standard_question;not null" json:"standard_question"`
	SimilarQuestions  string         `gorm:"column:similar_questions;type:text" json:"similar_questions"`
	NegativeQuestions string         `gorm:"column:negative_questions;type:text" json:"negative_questions"`
	Answers           string         `gorm:"column:answers;type:text;not null" json:"answers"`
	IsEnabled         bool           `gorm:"column:is_enabled;not null;default:true" json:"is_enabled"`
	IsRecommended     bool           `gorm:"column:is_recommended;not null;default:false" json:"is_recommended"`
	IndexMode         string         `gorm:"column:index_mode;not null;default:hybrid" json:"index_mode"`
	CreatedAt         *time.Time     `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         *time.Time     `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName FAQEntryM's table name
func (*FAQEntryM) TableName() string {
	return TableNameFAQEntryM
}

// FAQChunkMetadata FAQ 条目在 Chunk.Metadata 中的结构
type FAQChunkMetadata struct {
	StandardQuestion  string   `json:"standard_question"`
	SimilarQuestions  []string `json:"similar_questions,omitempty"`
	NegativeQuestions []string `json:"negative_questions,omitempty"`
	Answers           []string `json:"answers,omitempty"`
	Version           int      `json:"version,omitempty"`
}
