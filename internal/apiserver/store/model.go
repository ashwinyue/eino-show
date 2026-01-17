// Package store 提供 Model 存储.
package store

import (
	"context"

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

	// ListByType 获取指定类型的所有模型
	ListByType(ctx context.Context, modelType string) ([]*model.LLMModelM, error)

	// Update 更新模型
	Update(ctx context.Context, obj *model.LLMModelM) error

	// Delete 删除模型 (软删除)
	Delete(ctx context.Context, id string) error
}

// modelStore Model 存储实现.
type modelStore struct {
	store *datastore
}

// newModelStore 创建 ModelStore 实例.
func newModelStore(store *datastore) *modelStore {
	return &modelStore{store: store}
}

// 确保 modelStore 实现了 ModelStore 接口.
var _ ModelStore = (*modelStore)(nil)

// GetDefault 获取指定类型的默认模型.
func (s *modelStore) GetDefault(ctx context.Context, modelType string) (*model.LLMModelM, error) {
	var result model.LLMModelM
	err := s.store.DB(ctx).
		Where("type = ?", modelType).
		Where("is_default = ?", true).
		Where("status = ?", "active").
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByID 根据 ID 获取模型.
func (s *modelStore) GetByID(ctx context.Context, id string) (*model.LLMModelM, error) {
	var result model.LLMModelM
	err := s.store.DB(ctx).
		Where("id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List 获取模型列表.
func (s *modelStore) List(ctx context.Context, modelType string) ([]*model.LLMModelM, error) {
	var results []*model.LLMModelM

	query := s.store.DB(ctx).Where("status = ?", "active")
	if modelType != "" {
		query = query.Where("type = ?", modelType)
	}

	err := query.
		Order("is_default DESC").
		Order("created_at DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ListByType 获取指定类型的所有模型.
func (s *modelStore) ListByType(ctx context.Context, modelType string) ([]*model.LLMModelM, error) {
	var results []*model.LLMModelM
	err := s.store.DB(ctx).
		Where("type = ?", modelType).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Update 更新模型.
func (s *modelStore) Update(ctx context.Context, obj *model.LLMModelM) error {
	return s.store.DB(ctx).Save(obj).Error
}

// Delete 删除模型 (软删除).
func (s *modelStore) Delete(ctx context.Context, id string) error {
	return s.store.DB(ctx).Delete(&model.LLMModelM{}, "id = ?", id).Error
}
