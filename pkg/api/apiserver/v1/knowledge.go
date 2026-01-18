// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== Knowledge Base 请求/响应类型 =====

// ChunkingConfig 分块配置（对齐 WeKnora）
type ChunkingConfig struct {
	ChunkSize     int      `json:"chunk_size"`
	ChunkOverlap  int      `json:"chunk_overlap"`
	Separators    []string `json:"separators"`
	EnableMultimodal bool   `json:"enable_multimodal,omitempty"` // 兼容老版本
}

// ImageProcessingConfig 图像处理配置（对齐 WeKnora）
type ImageProcessingConfig struct {
	ModelID string `json:"model_id"`
}

// VLMConfig 多模态配置（对齐 WeKnora）
type VLMConfig struct {
	Enabled       bool   `json:"enabled"`
	ModelID       string `json:"model_id"`
	ModelName     string `json:"model_name,omitempty"`     // 兼容老版本
	BaseURL      string `json:"base_url,omitempty"`       // 兼容老版本
	APIKey        string `json:"api_key,omitempty"`        // 兼容老版本
	InterfaceType string `json:"interface_type,omitempty"` // 兼容老版本: "ollama" or "openai"
}

// IsEnabled 判断多模态是否启用（兼容新老版本配置）
func (c VLMConfig) IsEnabled() bool {
	if c.Enabled && c.ModelID != "" {
		return true
	}
	// 兼容老版本配置
	if c.ModelName != "" && c.BaseURL != "" {
		return true
	}
	return false
}

// StorageConfig 存储配置（对齐 WeKnora）
type StorageConfig struct {
	SecretID   string `json:"secret_id"`
	SecretKey  string `json:"secret_key"`
	Region     string `json:"region"`
	BucketName string `json:"bucket_name"`
	AppID      string `json:"app_id"`
	PathPrefix string `json:"path_prefix"`
	Provider   string `json:"provider"`
}

// GraphNode 图谱节点（对齐 WeKnora）
type GraphNode struct {
	Name string `json:"name"`
}

// GraphRelation 图谱关系（对齐 WeKnora）
type GraphRelation struct {
	Node1 string `json:"node1"`
	Node2 string `json:"node2"`
	Type  string `json:"type"`
}

// ExtractConfig 抽取配置（对齐 WeKnora）
type ExtractConfig struct {
	Enabled   bool           `json:"enabled"`
	Text      string         `json:"text,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
	Nodes     []*GraphNode   `json:"nodes,omitempty"`
	Relations []*GraphRelation `json:"relations,omitempty"`
}

// FAQConfig FAQ 知识库配置（对齐 WeKnora）
type FAQConfig struct {
	IndexMode         string `json:"index_mode"`          // "question_only" or "question_answer"
	QuestionIndexMode string `json:"question_index_mode"` // "combined" or "separate"
}

// QuestionGenerationConfig 问题生成配置（对齐 WeKnora）
type QuestionGenerationConfig struct {
	Enabled       bool `json:"enabled"`
	QuestionCount int  `json:"question_count"` // 默认3，最大10
}

// KnowledgeBaseConfig 知识库配置（对齐 WeKnora）
type KnowledgeBaseConfig struct {
	ChunkingConfig         ChunkingConfig             `json:"chunking_config"`
	ImageProcessingConfig  ImageProcessingConfig      `json:"image_processing_config"`
}

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	Type        string                `json:"type"` // "document" or "faq"
	Config      *KnowledgeBaseConfig  `json:"config"`
}

// UpdateKnowledgeBaseRequest 更新知识库请求（对齐 WeKnora）
type UpdateKnowledgeBaseRequest struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	Config      *KnowledgeBaseConfig  `json:"config" binding:"required"`
}

// KnowledgeBase 知识库（对齐 WeKnora）
type KnowledgeBase struct {
	ID                string                       `json:"id"`
	Name              string                       `json:"name"`
	Description       string                       `json:"description"`
	Type              string                       `json:"type"`
	TenantID          uint64                       `json:"tenant_id"`
	IsTemporary       bool                         `json:"is_temporary"`
	ChunkingConfig    ChunkingConfig               `json:"chunking_config"`
	ImageProcessingConfig ImageProcessingConfig    `json:"image_processing_config"`
	EmbeddingModelID  string                       `json:"embedding_model_id"`
	SummaryModelID    string                       `json:"summary_model_id"`
	VLMConfig         VLMConfig                    `json:"vlm_config"`
	StorageConfig     StorageConfig                `json:"cos_config"`
	ExtractConfig     *ExtractConfig               `json:"extract_config"`
	FAQConfig         *FAQConfig                   `json:"faq_config"`
	QuestionGenerationConfig *QuestionGenerationConfig `json:"question_generation_config"`
	KnowledgeCount    int64                        `json:"knowledge_count"`
	ChunkCount        int64                        `json:"chunk_count"`
	IsProcessing      bool                         `json:"is_processing"`
	ProcessingCount   int64                        `json:"processing_count"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

// KnowledgeBaseResponse 知识库响应（简化版，用于列表展示）
type KnowledgeBaseResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	TenantID    uint64    `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IsMultimodalEnabled 判断多模态是否启用（兼容新老版本配置）
func (kb *KnowledgeBase) IsMultimodalEnabled() bool {
	if kb == nil {
		return false
	}
	if kb.VLMConfig.IsEnabled() {
		return true
	}
	// 兼容老版本：chunking_config 中的 enable_multimodal 字段
	if kb.ChunkingConfig.EnableMultimodal {
		return true
	}
	return false
}

// EnsureDefaults 确保类型与配置具备默认值
func (kb *KnowledgeBase) EnsureDefaults() {
	if kb == nil {
		return
	}
	if kb.Type == "" {
		kb.Type = "document"
	}
	if kb.Type == "faq" && kb.FAQConfig == nil {
		kb.FAQConfig = &FAQConfig{
			IndexMode:         "question_answer",
			QuestionIndexMode: "combined",
		}
	}
}

// ListKnowledgeBasesRequest 知识库列表请求
type ListKnowledgeBasesRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ListKnowledgeBasesResponse 知识库列表响应
type ListKnowledgeBasesResponse struct {
	Success bool                   `json:"success"`
	Data    []*KnowledgeBaseResponse `json:"data"`
	Total   int64                  `json:"total"`
}

// GetKnowledgeBaseRequest 获取知识库请求
type GetKnowledgeBaseRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetKnowledgeBaseResponse 获取知识库响应
type GetKnowledgeBaseResponse struct {
	Success bool                   `json:"success"`
	Data    *KnowledgeBaseResponse `json:"data"`
}

// CreateKnowledgeBaseResponse 创建知识库响应
type CreateKnowledgeBaseResponse struct {
	Success bool                   `json:"success"`
	Data    *KnowledgeBaseResponse `json:"data"`
}

// UpdateKnowledgeBaseResponse 更新知识库响应
type UpdateKnowledgeBaseResponse struct {
	Success bool                   `json:"success"`
	Data    *KnowledgeBaseResponse `json:"data"`
}

// DeleteKnowledgeBaseRequest 删除知识库请求
type DeleteKnowledgeBaseRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteKnowledgeBaseResponse 删除知识库响应
type DeleteKnowledgeBaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
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

// ===== Knowledge 请求/响应类型 =====

// ManualKnowledgePayload 手工知识内容（对齐 WeKnora）
type ManualKnowledgePayload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"` // "draft" or "publish"
	TagID   string `json:"tag_id"`
}

// IsDraft 判断是否为草稿
func (p ManualKnowledgePayload) IsDraft() bool {
	return p.Status == "" || p.Status == "draft"
}

// Knowledge 知识实体（对齐 WeKnora）
type Knowledge struct {
	ID                  string    `json:"id"`
	TenantID            uint64    `json:"tenant_id"`
	KnowledgeBaseID     string    `json:"knowledge_base_id"` // 对齐 WeKnora 字段名
	TagID               string    `json:"tag_id"`
	Type                string    `json:"type"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Source              string    `json:"source"`
	ParseStatus         string    `json:"parse_status"`    // 对齐 WeKnora 字段名
	SummaryStatus       string    `json:"summary_status"`
	EnableStatus        string    `json:"enable_status"`
	EmbeddingModelID    string    `json:"embedding_model_id"`
	FileName            string    `json:"file_name"`
	FileType            string    `json:"file_type"`
	FileSize            int64     `json:"file_size"`
	FileHash            string    `json:"file_hash"`
	FilePath            string    `json:"file_path"`
	StorageSize         int64     `json:"storage_size"`
	Metadata            map[string]string `json:"metadata"`
	ProcessedAt         *time.Time `json:"processed_at"`
	ErrorMessage        string    `json:"error_message"`
	KnowledgeBaseName   string    `json:"knowledge_base_name"`
	ChunkCount          int       `json:"chunk_count"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// KnowledgeResponse 知识响应（简化版）
type KnowledgeResponse struct {
	ID         string    `json:"id"`
	KBID       string    `json:"kb_id"`
	Title      string    `json:"title"`
	Type       string    `json:"type"`
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
	Success bool              `json:"success"`
	Data    []*KnowledgeResponse `json:"data"`
	Total   int64              `json:"total"`
}

// GetKnowledgeRequest 获取知识请求
type GetKnowledgeRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetKnowledgeResponse 获取知识响应
type GetKnowledgeResponse struct {
	Success bool     `json:"success"`
	Data    *Knowledge `json:"data"`
}

// UpdateKnowledgeRequest 更新知识请求
type UpdateKnowledgeRequest struct {
	Id string `uri:"id" binding:"required"`
	*Knowledge
}

// UpdateKnowledgeResponse 更新知识响应
type UpdateKnowledgeResponse struct {
	Success bool     `json:"success"`
	Data    *Knowledge `json:"data"`
}

// DeleteKnowledgeRequest 删除知识请求
type DeleteKnowledgeRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteKnowledgeResponse 删除知识响应
type DeleteKnowledgeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// GetKnowledgeBatchRequest 批量获取知识请求
type GetKnowledgeBatchRequest struct {
	IDs []string `form:"ids" binding:"required"`
}

// GetKnowledgeBatchResponse 批量获取知识响应
type GetKnowledgeBatchResponse struct {
	Success bool       `json:"success"`
	Data    []*Knowledge `json:"data"`
}

// UpdateManualKnowledgeRequest 更新手工知识请求
type UpdateManualKnowledgeRequest struct {
	Id string `uri:"id" binding:"required"`
	*ManualKnowledgePayload
}

// UpdateManualKnowledgeResponse 更新手工知识响应
type UpdateManualKnowledgeResponse struct {
	Success bool     `json:"success"`
	Data    *Knowledge `json:"data"`
}

// DownloadKnowledgeFileRequest 下载知识文件请求
type DownloadKnowledgeFileRequest struct {
	Id string `uri:"id" binding:"required"`
}

// SearchKnowledgeRequest 搜索知识请求（对齐 WeKnora）
type SearchKnowledgeRequest struct {
	Query            string   `json:"query" binding:"required"`
	KnowledgeBaseID  string   `json:"knowledge_base_id"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids"`
	KnowledgeIDs     []string `json:"knowledge_ids"`
	// 搜索列表用字段
	Keyword   string `form:"keyword"`
	Offset    int    `form:"offset"`
	Limit     int    `form:"limit"`
	FileTypes string `form:"file_types"` // 逗号分隔
}

// SearchKnowledgeResponse 搜索知识响应
type SearchKnowledgeResponse struct {
	Success bool        `json:"success"`
	Data    []*Knowledge `json:"data"`
	HasMore bool         `json:"has_more"`
}

// UpdateKnowledgeTagBatchRequest 批量更新知识标签请求
type UpdateKnowledgeTagBatchRequest struct {
	Updates map[string]*string `json:"updates" binding:"required,min=1"`
}

// UpdateKnowledgeTagBatchResponse 批量更新知识标签响应
type UpdateKnowledgeTagBatchResponse struct {
	Success bool `json:"success"`
}

// UpdateImageInfoRequest 更新图像信息请求
type UpdateImageInfoRequest struct {
	Id         string `uri:"id" binding:"required"`
	ChunkID    string `uri:"chunk_id" binding:"required"`
	ImageInfo  string `json:"image_info" binding:"required"`
}

// UpdateImageInfoResponse 更新图像信息响应
type UpdateImageInfoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ===== 混合搜索类型 =====

// SearchParams 搜索参数（对齐 WeKnora）
type SearchParams struct {
	QueryText       string  `json:"query_text" binding:"required"`
	MatchCount      int32   `json:"match_count"`
	MatchThreshold  float64 `json:"match_threshold"`
}

// HybridSearchRequest 混合搜索请求（对齐 WeKnora）
type HybridSearchRequest struct {
	KnowledgeBaseId string      `uri:"id" binding:"required"`
	QueryText       string     `form:"query_text" binding:"required"`
	MatchCount      int32      `form:"match_count"`
	MatchThreshold  float64    `form:"match_threshold"`
}

// HybridSearchResponse 混合搜索响应（对齐 WeKnora）
type HybridSearchResponse struct {
	Success bool               `json:"success"`
	Data    []*SearchResultItem `json:"data"`
	Total   int64               `json:"total"`
}

// SearchResultItem 搜索结果项（对齐 WeKnora）
type SearchResultItem struct {
	ID          string  `json:"id"`
	KnowledgeID string  `json:"knowledge_id"`
	Content     string  `json:"content"`
	Score       float64 `json:"score"`
	ChunkIndex  int     `json:"chunk_index"`
}

// ===== Chunk 请求/响应类型 =====

// Chunk 分块（对齐 WeKnora）
type Chunk struct {
	ID             string        `json:"id"`
	KnowledgeID    string        `json:"knowledge_id"`
	Content        string        `json:"content"`
	ChunkIndex     int           `json:"chunk_index"`
	Embedding      []float32     `json:"embedding,omitempty"`
	GeneratedQuestions []string  `json:"generated_questions,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// ChunkResponse 分块响应（简化版）
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
	Success bool               `json:"success"`
	Data    []*ChunkResponse   `json:"data"`
	Total   int64              `json:"total"`
}

// GetChunkRequest 获取分块请求
type GetChunkRequest struct {
	Id string `uri:"id" binding:"required"`
}

// GetChunkResponse 获取分块响应
type GetChunkResponse struct {
	Success bool   `json:"success"`
	Data    *Chunk `json:"data"`
}

// UpdateChunkRequest 更新分块请求
type UpdateChunkRequest struct {
	Id      string `uri:"id" binding:"required"`
	Content string `json:"content"`
}

// UpdateChunkResponse 更新分块响应
type UpdateChunkResponse struct {
	Success bool   `json:"success"`
	Data    *Chunk `json:"data"`
	Message string `json:"message,omitempty"`
}

// DeleteChunkRequest 删除分块请求
type DeleteChunkRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteChunkResponse 删除分块响应
type DeleteChunkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// DeleteGeneratedQuestionRequest 删除生成问题请求
type DeleteGeneratedQuestionRequest struct {
	Id string `uri:"id" binding:"required"`
}

// DeleteGeneratedQuestionResponse 删除生成问题响应
type DeleteGeneratedQuestionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ===== Tag 请求/响应类型 =====

// Tag 标签（对齐 WeKnora）
type Tag struct {
	ID        string    `json:"id"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagResponse 标签响应（简化版）
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
	TagID string  `uri:"tag_id" binding:"required"`
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// ListTagsResponse 标签列表响应
type ListTagsResponse struct {
	Success bool         `json:"success"`
	Data    []*TagResponse `json:"data"`
	Total   int64         `json:"total"`
}

// ===== 知识库复制相关 =====

// CopyKnowledgeBaseRequest 复制知识库请求（对齐 WeKnora）
type CopyKnowledgeBaseRequest struct {
	TaskID   string `json:"task_id"`
	SourceID string `json:"source_id" binding:"required"`
	TargetID string `json:"target_id"`
	Name     string `json:"name"`
}

// CopyKnowledgeBaseResponse 复制知识库响应（对齐 WeKnora）
type CopyKnowledgeBaseResponse struct {
	Success bool   `json:"success"`
	Data    struct {
		TaskID   string `json:"task_id"`
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
		Message  string `json:"message"`
	} `json:"data"`
}

// GetKBCloneProgressRequest 获取复制进度请求
type GetKBCloneProgressRequest struct {
	TaskID string `uri:"task_id" binding:"required"`
}

// KBCloneProgress 知识库复制进度（对齐 WeKnora）
type KBCloneProgress struct {
	TaskID       string  `json:"task_id"`
	SourceID     string  `json:"source_id"`
	TargetID     string  `json:"target_id"`
	Status       string  `json:"status"`       // pending, processing, completed, failed
	Progress     int     `json:"progress"`     // 0-100
	Message      string  `json:"message"`
	TotalChunks  int     `json:"total_chunks"`
	CopiedChunks int     `json:"copied_chunks"`
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
}

// GetKBCloneProgressResponse 获取复制进度响应
type GetKBCloneProgressResponse struct {
	Success bool           `json:"success"`
	Data    *KBCloneProgress `json:"data"`
}
