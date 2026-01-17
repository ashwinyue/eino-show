// Package store provides summary storage implementation for progressive compression.
package store

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/pkg/retriever"
)

const (
	// SessionItemTypeSummary 摘要类型
	SessionItemTypeSummary = "summary"
)

// SummaryStore 定义了摘要存储接口.
type SummaryStore interface {
	retriever.SummaryStore
}

// summaryStore 实现 SummaryStore 接口.
type summaryStore struct {
	store *datastore
}

// newSummaryStore 创建 summaryStore 实例.
func newSummaryStore(store *datastore) *summaryStore {
	return &summaryStore{
		store: store,
	}
}

// Ensure summaryStore implements SummaryStore
var _ SummaryStore = (*summaryStore)(nil)

// SaveSummary 保存摘要到会话.
func (s *summaryStore) SaveSummary(ctx context.Context, sessionID string, summary string, tokenCount int) error {
	// 获取当前最大 sort_order
	var maxOrder int32
	err := s.store.DB(ctx).WithContext(ctx).
		Model(&model.SessionItemM{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error
	if err != nil {
		return err
	}

	tc := int32(tokenCount)
	item := &model.SessionItemM{
		ID:         uuid.New().String(),
		SessionID:  sessionID,
		Type:       SessionItemTypeSummary,
		Summary:    &summary,
		SortOrder:  maxOrder + 1,
		TokenCount: &tc,
	}

	return s.store.DB(ctx).WithContext(ctx).Create(item).Error
}

// GetSummaries 获取会话的所有摘要.
func (s *summaryStore) GetSummaries(ctx context.Context, sessionID string) ([]string, error) {
	var items []*model.SessionItemM
	err := s.store.DB(ctx).WithContext(ctx).
		Where("session_id = ? AND type = ?", sessionID, SessionItemTypeSummary).
		Order("sort_order ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	summaries := make([]string, 0, len(items))
	for _, item := range items {
		if item.Summary != nil {
			summaries = append(summaries, *item.Summary)
		}
	}

	return summaries, nil
}

// ReplaceSummaries 替换所有摘要为一个合并后的摘要.
func (s *summaryStore) ReplaceSummaries(ctx context.Context, sessionID string, mergedSummary string, tokenCount int) error {
	return s.store.DB(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除所有旧摘要
		if err := tx.Where("session_id = ? AND type = ?", sessionID, SessionItemTypeSummary).
			Delete(&model.SessionItemM{}).Error; err != nil {
			return err
		}

		// 2. 获取当前最大 sort_order
		var maxOrder int32
		if err := tx.Model(&model.SessionItemM{}).
			Where("session_id = ?", sessionID).
			Select("COALESCE(MAX(sort_order), 0)").
			Scan(&maxOrder).Error; err != nil {
			return err
		}

		// 3. 创建新的合并摘要
		tc := int32(tokenCount)
		item := &model.SessionItemM{
			ID:         uuid.New().String(),
			SessionID:  sessionID,
			Type:       SessionItemTypeSummary,
			Summary:    &mergedSummary,
			SortOrder:  maxOrder + 1,
			TokenCount: &tc,
		}

		return tx.Create(item).Error
	})
}

// ClearSummaries 清除会话的所有摘要.
func (s *summaryStore) ClearSummaries(ctx context.Context, sessionID string) error {
	return s.store.DB(ctx).WithContext(ctx).
		Where("session_id = ? AND type = ?", sessionID, SessionItemTypeSummary).
		Delete(&model.SessionItemM{}).Error
}
