package faq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// FAQBiz FAQ 业务接口
type FAQBiz interface {
	ListEntries(ctx context.Context, kbID string, page, pageSize int, tagID *int64, keyword, searchField, sortOrder string) (*ListEntriesResponse, error)
	CreateEntry(ctx context.Context, kbID string, req *CreateEntryRequest) (*FAQEntryResponse, error)
	GetEntry(ctx context.Context, kbID string, entryID int64) (*FAQEntryResponse, error)
	UpdateEntry(ctx context.Context, kbID string, entryID int64, req *UpdateEntryRequest) error
	DeleteEntries(ctx context.Context, kbID string, ids []int64) error
	UpdateTagBatch(ctx context.Context, kbID string, updates map[int64]*int64) error
	UpdateFieldsBatch(ctx context.Context, kbID string, updates map[int64]map[string]interface{}) error
	SearchFAQ(ctx context.Context, kbID string, req *SearchFAQRequest) ([]*FAQEntryResponse, error)
}

// CreateEntryRequest 创建 FAQ 条目请求
type CreateEntryRequest struct {
	StandardQuestion  string   `json:"standard_question" binding:"required"`
	SimilarQuestions  []string `json:"similar_questions"`
	NegativeQuestions []string `json:"negative_questions"`
	Answers           []string `json:"answers" binding:"required"`
	TagID             *int64   `json:"tag_id"`
	IsEnabled         *bool    `json:"is_enabled"`
}

// UpdateEntryRequest 更新 FAQ 条目请求
type UpdateEntryRequest struct {
	StandardQuestion  *string  `json:"standard_question"`
	SimilarQuestions  []string `json:"similar_questions"`
	NegativeQuestions []string `json:"negative_questions"`
	Answers           []string `json:"answers"`
	TagID             *int64   `json:"tag_id"`
	IsEnabled         *bool    `json:"is_enabled"`
}

// SearchFAQRequest 搜索 FAQ 请求
type SearchFAQRequest struct {
	QueryText       string  `json:"query_text" binding:"required"`
	VectorThreshold float64 `json:"vector_threshold"`
	MatchCount      int     `json:"match_count"`
}

// ListEntriesResponse 列表响应
type ListEntriesResponse struct {
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Data     []*FAQEntryResponse `json:"data"`
}

// FAQEntryResponse FAQ 条目响应
type FAQEntryResponse struct {
	ID                int64      `json:"id"`
	ChunkID           string     `json:"chunk_id"`
	KnowledgeID       string     `json:"knowledge_id"`
	KnowledgeBaseID   string     `json:"knowledge_base_id"`
	TagID             *int64     `json:"tag_id"`
	IsEnabled         bool       `json:"is_enabled"`
	IsRecommended     bool       `json:"is_recommended"`
	StandardQuestion  string     `json:"standard_question"`
	SimilarQuestions  []string   `json:"similar_questions"`
	NegativeQuestions []string   `json:"negative_questions"`
	Answers           []string   `json:"answers"`
	IndexMode         string     `json:"index_mode"`
	ChunkType         string     `json:"chunk_type"`
	Score             float64    `json:"score,omitempty"`
	MatchType         string     `json:"match_type,omitempty"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
}

type faqBiz struct {
	store store.IStore
}

// New 创建 FAQ Biz 实例
func New(store store.IStore) FAQBiz {
	return &faqBiz{store: store}
}

func (b *faqBiz) ListEntries(ctx context.Context, kbID string, page, pageSize int, tagID *int64, keyword, searchField, sortOrder string) (*ListEntriesResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	total, list, err := b.store.FAQ().ListByKnowledgeBaseID(ctx, kbID, page, pageSize, tagID, keyword, searchField, sortOrder)
	if err != nil {
		return nil, err
	}

	data := make([]*FAQEntryResponse, len(list))
	for i, entry := range list {
		data[i] = toFAQEntryResponse(entry)
	}

	return &ListEntriesResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

func (b *faqBiz) CreateEntry(ctx context.Context, kbID string, req *CreateEntryRequest) (*FAQEntryResponse, error) {
	tenantID := contextx.TenantID(ctx)

	// 检查标准问题是否重复
	existing, _ := b.store.FAQ().GetByStandardQuestion(ctx, kbID, req.StandardQuestion)
	if existing != nil {
		return nil, &DuplicateError{Message: "标准问与已有FAQ重复"}
	}

	now := time.Now()
	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	entry := &model.FAQEntryM{
		TenantID:          int32(tenantID),
		KnowledgeBaseID:   kbID,
		TagID:             req.TagID,
		StandardQuestion:  req.StandardQuestion,
		SimilarQuestions:  toJSON(req.SimilarQuestions),
		NegativeQuestions: toJSON(req.NegativeQuestions),
		Answers:           toJSON(req.Answers),
		IsEnabled:         isEnabled,
		IndexMode:         "hybrid",
		CreatedAt:         &now,
		UpdatedAt:         &now,
	}

	if err := b.store.FAQ().Create(ctx, entry); err != nil {
		return nil, err
	}

	return toFAQEntryResponse(entry), nil
}

func (b *faqBiz) GetEntry(ctx context.Context, kbID string, entryID int64) (*FAQEntryResponse, error) {
	entry, err := b.store.FAQ().Get(ctx, where.NewWhere().F("id", entryID).F("knowledge_base_id", kbID))
	if err != nil {
		return nil, err
	}
	return toFAQEntryResponse(entry), nil
}

func (b *faqBiz) UpdateEntry(ctx context.Context, kbID string, entryID int64, req *UpdateEntryRequest) error {
	entry, err := b.store.FAQ().Get(ctx, where.NewWhere().F("id", entryID).F("knowledge_base_id", kbID))
	if err != nil {
		return err
	}

	if req.StandardQuestion != nil {
		entry.StandardQuestion = *req.StandardQuestion
	}
	if req.SimilarQuestions != nil {
		entry.SimilarQuestions = toJSON(req.SimilarQuestions)
	}
	if req.NegativeQuestions != nil {
		entry.NegativeQuestions = toJSON(req.NegativeQuestions)
	}
	if req.Answers != nil {
		entry.Answers = toJSON(req.Answers)
	}
	if req.TagID != nil {
		entry.TagID = req.TagID
	}
	if req.IsEnabled != nil {
		entry.IsEnabled = *req.IsEnabled
	}

	now := time.Now()
	entry.UpdatedAt = &now

	return b.store.FAQ().Update(ctx, entry)
}

func (b *faqBiz) DeleteEntries(ctx context.Context, kbID string, ids []int64) error {
	for _, id := range ids {
		if err := b.store.FAQ().Delete(ctx, where.NewWhere().F("id", id).F("knowledge_base_id", kbID)); err != nil {
			return err
		}
	}
	return nil
}

func (b *faqBiz) UpdateTagBatch(ctx context.Context, kbID string, updates map[int64]*int64) error {
	return b.store.FAQ().UpdateTagBatch(ctx, updates)
}

func (b *faqBiz) UpdateFieldsBatch(ctx context.Context, kbID string, updates map[int64]map[string]interface{}) error {
	return b.store.FAQ().UpdateFieldsBatch(ctx, updates)
}

func (b *faqBiz) SearchFAQ(ctx context.Context, kbID string, req *SearchFAQRequest) ([]*FAQEntryResponse, error) {
	matchCount := req.MatchCount
	if matchCount <= 0 {
		matchCount = 10
	}

	// 简单关键词搜索
	_, list, err := b.store.FAQ().ListByKnowledgeBaseID(ctx, kbID, 1, matchCount, nil, req.QueryText, "", "")
	if err != nil {
		return nil, err
	}

	results := make([]*FAQEntryResponse, len(list))
	for i, entry := range list {
		resp := toFAQEntryResponse(entry)
		resp.Score = 1.0
		resp.MatchType = "keyword"
		results[i] = resp
	}

	return results, nil
}

// DuplicateError 重复错误
type DuplicateError struct {
	Message string
}

func (e *DuplicateError) Error() string {
	return e.Message
}

func toFAQEntryResponse(entry *model.FAQEntryM) *FAQEntryResponse {
	return &FAQEntryResponse{
		ID:                entry.ID,
		ChunkID:           entry.ChunkID,
		KnowledgeID:       entry.KnowledgeID,
		KnowledgeBaseID:   entry.KnowledgeBaseID,
		TagID:             entry.TagID,
		IsEnabled:         entry.IsEnabled,
		IsRecommended:     entry.IsRecommended,
		StandardQuestion:  entry.StandardQuestion,
		SimilarQuestions:  fromJSON(entry.SimilarQuestions),
		NegativeQuestions: fromJSON(entry.NegativeQuestions),
		Answers:           fromJSON(entry.Answers),
		IndexMode:         entry.IndexMode,
		ChunkType:         "faq",
		CreatedAt:         entry.CreatedAt,
		UpdatedAt:         entry.UpdatedAt,
	}
}

func toJSON(arr []string) string {
	if arr == nil {
		return "[]"
	}
	data, _ := json.Marshal(arr)
	return string(data)
}

func fromJSON(s string) []string {
	if s == "" || s == "[]" {
		return []string{}
	}
	var arr []string
	_ = json.Unmarshal([]byte(s), &arr)
	return arr
}
