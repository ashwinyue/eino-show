// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== Knowledge Base 请求/响应类型 =====

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"` // "document" or "faq"
}

// UpdateKnowledgeBaseRequest 更新知识库请求
type UpdateKnowledgeBaseRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// KnowledgeBaseResponse 知识库响应
type KnowledgeBaseResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	TenantID    uint64    `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListKnowledgeBasesResponse 知识库列表响应
type ListKnowledgeBasesResponse struct {
	KnowledgeBases []*KnowledgeBaseResponse `json:"knowledge_bases"`
	Total          int64                    `json:"total"`
}

// ===== Knowledge 请求/响应类型 =====

// KnowledgeResponse 知识响应
type KnowledgeResponse struct {
	ID         string    `json:"id"`
	KbID       string    `json:"kb_id"`
	Title      string    `json:"title"`
	Type       string    `json:"type"` // "file", "url", "manual"
	Status     string    `json:"status"`
	FileSize   int64     `json:"file_size"`
	ChunkCount int       `json:"chunk_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ListKnowledgesRequest 知识列表请求
type ListKnowledgesRequest struct {
	KbId string `uri:"id" binding:"required"`
}

// ListKnowledgesResponse 知识列表响应
type ListKnowledgesResponse struct {
	Knowledge []*KnowledgeResponse `json:"knowledge"`
	Total     int64                `json:"total"`
}

// HybridSearchRequest 混合搜索请求
type HybridSearchRequest struct {
	KnowledgeBaseId string `uri:"id" binding:"required"`
	QueryText       string `form:"query_text" binding:"required"`
	MatchCount      int32  `form:"match_count"`
}

// HybridSearchResponse 混合搜索响应
type HybridSearchResponse struct {
	Results []*SearchResultItem `json:"results"`
	Total   int64               `json:"total"`
}

// SearchResultItem 搜索结果项
type SearchResultItem struct {
	ID          string  `json:"id"`
	KnowledgeID string  `json:"knowledge_id"`
	Content     string  `json:"content"`
	Score       float64 `json:"score"`
}

// ===== Chunk 请求/响应类型 =====

// ChunkResponse 分块响应
type ChunkResponse struct {
	ID          string    `json:"id"`
	KnowledgeID string    `json:"knowledge_id"`
	Content     string    `json:"content"`
	Index       int       `json:"index"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListChunksRequest 分块列表请求
type ListChunksRequest struct {
	KnowledgeId string `form:"knowledge_id" binding:"required"`
}

// ListChunksResponse 分块列表响应
type ListChunksResponse struct {
	Chunks []*ChunkResponse `json:"chunks"`
	Total  int64            `json:"total"`
}

// GetChunkRequest 获取分块请求
type GetChunkRequest struct {
	Id string `uri:"id" binding:"required"`
}

// UpdateChunkRequest 更新分块请求
type UpdateChunkRequest struct {
	Id      string `uri:"id" binding:"required"`
	Content string `json:"content"`
}

// DeleteChunkRequest 删除分块请求
type DeleteChunkRequest struct {
	Id string `uri:"id" binding:"required"`
}

// ===== Tag 请求/响应类型 =====

// TagResponse 标签响应
type TagResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// ListTagsResponse 标签列表响应
type ListTagsResponse struct {
	Tags  []*TagResponse `json:"tags"`
	Total int64          `json:"total"`
}

// ===== 扩展请求/响应类型 =====

// ListKnowledgeBasesRequest 知识库列表请求
type ListKnowledgeBasesRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// GetKnowledgeBaseRequest 获取知识库请求
type GetKnowledgeBaseRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetKnowledgeBaseResponse 获取知识库响应
type GetKnowledgeBaseResponse struct {
	KnowledgeBase *KnowledgeBaseResponse `json:"knowledge_base"`
}

// CreateKnowledgeBaseResponse 创建知识库响应
type CreateKnowledgeBaseResponse struct {
	KnowledgeBase *KnowledgeBaseResponse `json:"knowledge_base"`
}

// UpdateKnowledgeBaseResponse 更新知识库响应
type UpdateKnowledgeBaseResponse struct {
	KnowledgeBase *KnowledgeBaseResponse `json:"knowledge_base"`
}

// DeleteKnowledgeBaseRequest 删除知识库请求
type DeleteKnowledgeBaseRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteKnowledgeBaseResponse 删除知识库响应
type DeleteKnowledgeBaseResponse struct {
	Success bool `json:"success"`
}

// GetKnowledgeStatsRequest 获取知识库统计请求
type GetKnowledgeStatsRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetKnowledgeStatsResponse 获取知识库统计响应
type GetKnowledgeStatsResponse struct {
	KnowledgeCount int64 `json:"knowledge_count"`
	ChunkCount     int64 `json:"chunk_count"`
	TotalSize      int64 `json:"total_size"`
}

// DeleteKnowledgeRequest 删除知识请求
type DeleteKnowledgeRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteKnowledgeResponse 删除知识响应
type DeleteKnowledgeResponse struct {
	Success bool `json:"success"`
}

// Knowledge 知识实体
type Knowledge struct {
	ID         string    `json:"id"`
	KbID       string    `json:"kb_id"`
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	FileSize   int64     `json:"file_size"`
	ChunkCount int       `json:"chunk_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetChunkResponse 获取分块响应
type GetChunkResponse struct {
	Chunk *ChunkResponse `json:"chunk"`
}

// UpdateChunkResponse 更新分块响应
type UpdateChunkResponse struct {
	Chunk *ChunkResponse `json:"chunk"`
}

// DeleteChunkResponse 删除分块响应
type DeleteChunkResponse struct {
	Success bool `json:"success"`
}

// KnowledgeStatsResponse 知识库统计响应（兼容别名）
type KnowledgeStatsResponse = GetKnowledgeStatsResponse

// KnowledgeChunkingConfig 知识分块配置
type KnowledgeChunkingConfig struct {
	ChunkSize     int      `json:"chunk_size"`
	ChunkOverlap  int      `json:"chunk_overlap"`
	Separator     string   `json:"separator"`
	SplitMarkers  []string `json:"split_markers"`
	KeepSeparator bool     `json:"keep_separator"`
}

// KnowledgeImageConfig 知识图片配置
type KnowledgeImageConfig struct {
	EnableOCR        bool   `json:"enable_ocr"`
	EnableMultimodal bool   `json:"enable_multimodal"`
	ModelId          string `json:"model_id"`
}
