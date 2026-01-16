
// Package store 提供 Model 存储.
package store

import (
	"context"

	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// ModelStore 模型存储接口.
type ModelStore interface {
	// GetDefault 获取指定类型的默认模型
	GetDefault(ctx context.Context, modelType string) (*model.LLMModelM, error)

	// GetByID 根据 ID 获取模型
	GetByID(ctx context.Context, id string) (*model.LLMModelM, error)

	// List 获取模型列表
	List(ctx context.Context, modelType string) ([]*model.LLMModelM, error)
}

// modelStore Model 存储实现.
type modelStore struct {
	*genericstore.Store[model.LLMModelM]
}

// NewModelStore 创建 ModelStore 实例.
func NewModelStore(core store.IStore) ModelStore {
	return &modelStore{
		Store: genericstore.NewStore[model.LLMModelM](core, nil),
	}
}

// 确保 modelStore 实现了 ModelStore 接口.
var _ ModelStore = (*modelStore)(nil)

// GetDefault 获取指定类型的默认模型.
func (s *modelStore) GetDefault(ctx context.Context, modelType string) (*model.LLMModelM, error) {
	var result model.LLMModelM
	err := s.Store.Get(ctx,
		where.NewWhere().
			F("type", modelType).
			F("is_default", true).
			F("status", "active"),
	).Get(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByID 根据 ID 获取模型.
func (s *modelStore) GetByID(ctx context.Context, id string) (*model.LLMModelM, error) {
	var result model.LLMModelM
	err := s.Store.Get(ctx,
		where.NewWhere().F("id", id),
	).Get(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List 获取模型列表.
func (s *modelStore) List(ctx context.Context, modelType string) ([]*model.LLMModelM, error) {
	var results []*model.LLMModelM

	opts := where.NewWhere().F("status", "active")
	if modelType != "" {
		opts = opts.F("type", modelType)
	}

	_, results, err := s.Store.List(ctx, opts,
		genericstore.OrderBy("is_default").Desc(),
		genericstore.OrderBy("created_at").Desc(),
	)
	if err != nil {
		return nil, err
	}
	return results, nil
}
