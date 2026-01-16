// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package store

import (
	"context"
	"fmt"

	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// KnowledgeBaseStore 定义了 knowledge_base 模块在 store 层所实现的方法.
type KnowledgeBaseStore interface {
	Create(ctx context.Context, obj *model.KnowledgeBaseM) error
	Update(ctx context.Context, obj *model.KnowledgeBaseM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.KnowledgeBaseM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.KnowledgeBaseM, error)

	KnowledgeBaseExpansion
}

// KnowledgeBaseExpansion 定义了知识库操作的附加方法.
// nolint: iface
type KnowledgeBaseExpansion interface {
	// GetByTenantID 获取租户下的所有知识库
	GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.KnowledgeBaseM, error)
}

// knowledgeBaseStore 是 KnowledgeBaseStore 接口的实现.
type knowledgeBaseStore struct {
	store *datastore
	*genericstore.Store[model.KnowledgeBaseM]
}

// 确保 knowledgeBaseStore 实现了 KnowledgeBaseStore 接口.
var _ KnowledgeBaseStore = (*knowledgeBaseStore)(nil)

// newKnowledgeBaseStore 创建 knowledgeBaseStore 的实例.
func newKnowledgeBaseStore(store *datastore) *knowledgeBaseStore {
	return &knowledgeBaseStore{
		store: store,
		Store: genericstore.NewStore[model.KnowledgeBaseM](store, NewLogger()),
	}
}

// GetByTenantID 获取租户下的所有知识库.
func (s *knowledgeBaseStore) GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.KnowledgeBaseM, error) {
	var list []*model.KnowledgeBaseM
	err := s.store.DB(ctx).Where("tenant_id = ? AND (deleted_at IS NULL OR deleted_at > '0001-01-01')", tenantID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// KnowledgeStore 定义了 knowledge 模块在 store 层所实现的方法.
type KnowledgeStore interface {
	Create(ctx context.Context, obj *model.KnowledgeM) error
	Update(ctx context.Context, obj *model.KnowledgeM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.KnowledgeM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.KnowledgeM, error)

	KnowledgeExpansion
}

// KnowledgeExpansion 定义了知识操作的附加方法.
// nolint: iface
type KnowledgeExpansion interface {
	// GetByKnowledgeBaseID 获取指定知识库下的所有知识
	GetByKnowledgeBaseID(ctx context.Context, kbID string) ([]*model.KnowledgeM, error)
	// GetByTenantID 获取租户下的所有知识
	GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.KnowledgeM, error)
}

// knowledgeStore 是 KnowledgeStore 接口的实现.
type knowledgeStore struct {
	store *datastore
	*genericstore.Store[model.KnowledgeM]
}

// 确保 knowledgeStore 实现了 KnowledgeStore 接口.
var _ KnowledgeStore = (*knowledgeStore)(nil)

// newKnowledgeStore 创建 knowledgeStore 的实例.
func newKnowledgeStore(store *datastore) *knowledgeStore {
	return &knowledgeStore{
		store: store,
		Store: genericstore.NewStore[model.KnowledgeM](store, NewLogger()),
	}
}

// GetByKnowledgeBaseID 获取指定知识库下的所有知识.
func (s *knowledgeStore) GetByKnowledgeBaseID(ctx context.Context, kbID string) ([]*model.KnowledgeM, error) {
	var list []*model.KnowledgeM
	err := s.store.DB(ctx).Where("knowledge_base_id = ?", kbID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// GetByTenantID 获取租户下的所有知识.
func (s *knowledgeStore) GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.KnowledgeM, error) {
	var list []*model.KnowledgeM
	err := s.store.DB(ctx).Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// ChunkStore 定义了 chunk 模块在 store 层所实现的方法.
type ChunkStore interface {
	Create(ctx context.Context, obj *model.ChunkM) error
	CreateBatch(ctx context.Context, objs []*model.ChunkM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.ChunkM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.ChunkM, error)

	ChunkExpansion
}

// ChunkExpansion 定义了分块操作的附加方法.
// nolint: iface
type ChunkExpansion interface {
	// GetByKnowledgeID 获取指定知识的所有分块
	GetByKnowledgeID(ctx context.Context, knowledgeID string) ([]*model.ChunkM, error)
	// GetByKnowledgeBaseID 获取指定知识库的所有分块
	GetByKnowledgeBaseID(ctx context.Context, kbID string) ([]*model.ChunkM, error)
	// DeleteByKnowledgeID 删除指定知识的所有分块
	DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error
	// VectorSearch 向量搜索：根据查询向量返回最相似的 N 个分块
	VectorSearch(ctx context.Context, kbID string, embedding []float32, limit int) ([]*model.ChunkM, error)
	// KeywordSearch 关键词搜索：根据关键词返回包含该关键词的分块
	KeywordSearch(ctx context.Context, kbID string, keyword string, limit int, caseSensitive bool) ([]*model.ChunkM, error)
}

// chunkStore 是 ChunkStore 接口的实现.
type chunkStore struct {
	store *datastore
}

// 确保 chunkStore 实现了 ChunkStore 接口.
var _ ChunkStore = (*chunkStore)(nil)

// newChunkStore 创建 chunkStore 的实例.
func newChunkStore(store *datastore) *chunkStore {
	return &chunkStore{store: store}
}

// Create 创建分块记录.
func (s *chunkStore) Create(ctx context.Context, obj *model.ChunkM) error {
	return s.store.DB(ctx).Create(obj).Error
}

// CreateBatch 批量创建分块记录.
func (s *chunkStore) CreateBatch(ctx context.Context, objs []*model.ChunkM) error {
	if len(objs) == 0 {
		return nil
	}
	return s.store.DB(ctx).Create(&objs).Error
}

// Delete 根据条件删除分块记录.
func (s *chunkStore) Delete(ctx context.Context, opts *where.Options) error {
	return s.store.DB(ctx, opts).Delete(new(model.ChunkM)).Error
}

// Get 根据条件查询分块记录.
func (s *chunkStore) Get(ctx context.Context, opts *where.Options) (*model.ChunkM, error) {
	var obj model.ChunkM
	err := s.store.DB(ctx, opts).First(&obj).Error
	return &obj, err
}

// List 返回分块列表和总数.
func (s *chunkStore) List(ctx context.Context, opts *where.Options) (count int64, ret []*model.ChunkM, err error) {
	err = s.store.DB(ctx, opts).Order("id desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error
	return
}

// GetByKnowledgeID 获取指定知识的所有分块.
func (s *chunkStore) GetByKnowledgeID(ctx context.Context, knowledgeID string) ([]*model.ChunkM, error) {
	var list []*model.ChunkM
	err := s.store.DB(ctx).Where("knowledge_id = ?", knowledgeID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

// GetByKnowledgeBaseID 获取指定知识库的所有分块.
func (s *chunkStore) GetByKnowledgeBaseID(ctx context.Context, kbID string) ([]*model.ChunkM, error) {
	var list []*model.ChunkM
	err := s.store.DB(ctx).Where("knowledge_base_id = ?", kbID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

// DeleteByKnowledgeID 删除指定知识的所有分块.
func (s *chunkStore) DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error {
	return s.store.DB(ctx).Where("knowledge_id = ?", knowledgeID).Delete(new(model.ChunkM)).Error
}

// VectorSearch 向量搜索：根据查询向量返回最相似的 N 个分块.
// 使用 PGVector 的 cosine 距离计算相似度.
func (s *chunkStore) VectorSearch(ctx context.Context, kbID string, embedding []float32, limit int) ([]*model.ChunkM, error) {
	var list []*model.ChunkM

	// 将 []float32 转换为 PGVector 格式的字符串 "[0.1,0.2,...]"
	embeddingStr := vectorToString(embedding)

	// 使用 PGVector 的 cosine 距离计算相似度
	// 距离越小越相似，返回按距离排序的前 limit 条结果
	query := s.store.DB(ctx).
		Where("knowledge_base_id = ?", kbID).
		Order("embedding <=> ?")

	// 如果 limit <= 0，默认返回 5 条
	if limit <= 0 {
		limit = 5
	}

	err := query.Limit(limit).Find(&list, embeddingStr).Error
	return list, err
}

// vectorToString 将 []float32 向量转换为 PGVector 格式的字符串.
func vectorToString(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	result := "["
	for i, val := range v {
		if i > 0 {
			result += ","
		}
		result += floatToString(val)
	}
	result += "]"
	return result
}

// floatToString 将 float32 转换为字符串格式.
func floatToString(f float32) string {
	return fmt.Sprintf("%f", f)
}

// KeywordSearch 关键词搜索：根据关键词返回包含该关键词的分块.
func (s *chunkStore) KeywordSearch(ctx context.Context, kbID string, keyword string, limit int, caseSensitive bool) ([]*model.ChunkM, error) {
	var list []*model.ChunkM

	// 构建查询
	query := s.store.DB(ctx).Where("knowledge_base_id = ?", kbID)

	// 根据是否区分大小写使用不同的查询
	if caseSensitive {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	} else {
		// PostgreSQL 的 ILIKE 用于不区分大小写的搜索
		query = query.Where("content ILIKE ?", "%"+keyword+"%")
	}

	// 设置 limit
	if limit <= 0 {
		limit = 10
	}

	err := query.Order("id ASC").Limit(limit).Find(&list).Error
	return list, err
}
