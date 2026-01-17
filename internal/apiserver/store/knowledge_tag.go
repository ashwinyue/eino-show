// Package store 提供知识标签存储.
package store

import (
	"context"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// KnowledgeTagStore 知识标签存储接口.
type KnowledgeTagStore interface {
	Create(ctx context.Context, obj *model.KnowledgeTagM) error
	Update(ctx context.Context, obj *model.KnowledgeTagM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.KnowledgeTagM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.KnowledgeTagM, error)

	KnowledgeTagExpansion
}

// KnowledgeTagExpansion 定义了知识标签操作的附加方法.
// nolint: iface
type KnowledgeTagExpansion interface {
	// ListByKnowledgeBase 获取指定知识库的标签列表
	ListByKnowledgeBase(ctx context.Context, tenantID int32, kbID string) ([]*model.KnowledgeTagM, error)
	// GetByName 根据名称获取标签
	GetByName(ctx context.Context, tenantID int32, kbID, name string) (*model.KnowledgeTagM, error)
}

// knowledgeTagStore 知识标签存储实现.
type knowledgeTagStore struct {
	store *datastore
}

// 确保 knowledgeTagStore 实现了 KnowledgeTagStore 接口.
var _ KnowledgeTagStore = (*knowledgeTagStore)(nil)

// newKnowledgeTagStore 创建 KnowledgeTagStore 实例.
func newKnowledgeTagStore(store *datastore) *knowledgeTagStore {
	return &knowledgeTagStore{store: store}
}

// Create 创建知识标签.
func (s *knowledgeTagStore) Create(ctx context.Context, obj *model.KnowledgeTagM) error {
	return s.store.DB(ctx).Create(obj).Error
}

// Update 更新知识标签.
func (s *knowledgeTagStore) Update(ctx context.Context, obj *model.KnowledgeTagM) error {
	return s.store.DB(ctx).Where("id = ?", obj.ID).Updates(obj).Error
}

// Delete 删除知识标签.
func (s *knowledgeTagStore) Delete(ctx context.Context, opts *where.Options) error {
	return s.store.DB(ctx, opts).Delete(new(model.KnowledgeTagM)).Error
}

// Get 获取知识标签.
func (s *knowledgeTagStore) Get(ctx context.Context, opts *where.Options) (*model.KnowledgeTagM, error) {
	var tag model.KnowledgeTagM
	err := s.store.DB(ctx, opts).First(&tag).Error
	return &tag, err
}

// List 获取知识标签列表.
func (s *knowledgeTagStore) List(ctx context.Context, opts *where.Options) (int64, []*model.KnowledgeTagM, error) {
	var count int64
	var list []*model.KnowledgeTagM
	err := s.store.DB(ctx, opts).Order("sort_order ASC, created_at DESC").Find(&list).Offset(-1).Limit(-1).Count(&count).Error
	return count, list, err
}

// ListByKnowledgeBase 获取指定知识库的标签列表.
func (s *knowledgeTagStore) ListByKnowledgeBase(ctx context.Context, tenantID int32, kbID string) ([]*model.KnowledgeTagM, error) {
	var list []*model.KnowledgeTagM
	err := s.store.DB(ctx).
		Where("tenant_id = ?", tenantID).
		Where("knowledge_base_id = ?", kbID).
		Order("sort_order ASC, created_at DESC").
		Find(&list).Error
	return list, err
}

// GetByName 根据名称获取标签.
func (s *knowledgeTagStore) GetByName(ctx context.Context, tenantID int32, kbID, name string) (*model.KnowledgeTagM, error) {
	var tag model.KnowledgeTagM
	err := s.store.DB(ctx).
		Where("tenant_id = ?", tenantID).
		Where("knowledge_base_id = ?", kbID).
		Where("name = ?", name).
		First(&tag).Error
	return &tag, err
}
