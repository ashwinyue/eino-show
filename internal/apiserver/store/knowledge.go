package store

import (
	"context"
	"fmt"

	genericstore "github.com/ashwinyue/eino-show/pkg/store"
	"github.com/ashwinyue/eino-show/pkg/store/where"

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
	Update(ctx context.Context, obj *model.ChunkM) error

	ChunkExpansion
}

// ChunkExpansion 定义了分块操作的附加方法.
// 对齐 WeKnora 的 ChunkRepository 扩展方法.
// nolint: iface
type ChunkExpansion interface {
	// GetByKnowledgeID 获取指定知识的所有分块
	GetByKnowledgeID(ctx context.Context, knowledgeID string) ([]*model.ChunkM, error)
	// GetByKnowledgeBaseID 获取指定知识库的所有分块
	GetByKnowledgeBaseID(ctx context.Context, kbID string) ([]*model.ChunkM, error)
	// GetByParentID 按父分块 ID 列出子分块（如图片 Chunk）
	GetByParentID(ctx context.Context, parentID string) ([]*model.ChunkM, error)
	// DeleteByKnowledgeID 删除指定知识的所有分块
	DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error
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

// Update 更新分块记录.
func (s *chunkStore) Update(ctx context.Context, obj *model.ChunkM) error {
	return s.store.DB(ctx).Save(obj).Error
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

// GetByParentID 按父分块 ID 列出子分块.
// 对齐 WeKnora 的 ListChunkByParentID.
func (s *chunkStore) GetByParentID(ctx context.Context, parentID string) ([]*model.ChunkM, error) {
	var list []*model.ChunkM
	err := s.store.DB(ctx).Where("parent_chunk_id = ?", parentID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

// DeleteByKnowledgeID 删除指定知识的所有分块.
func (s *chunkStore) DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error {
	return s.store.DB(ctx).Where("knowledge_id = ?", knowledgeID).Delete(new(model.ChunkM)).Error
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

// EmbeddingStore 定义了 embedding 模块在 store 层所实现的方法.
// 对齐 WeKnora 的 embeddings 表结构.
type EmbeddingStore interface {
	// Create 创建单条 embedding 记录
	Create(ctx context.Context, obj *model.EmbeddingM) error
	// CreateBatch 批量创建 embedding 记录
	CreateBatch(ctx context.Context, objs []*model.EmbeddingM) error
	// Delete 根据条件删除 embedding 记录
	Delete(ctx context.Context, opts *where.Options) error
	// Get 根据条件查询 embedding 记录
	Get(ctx context.Context, opts *where.Options) (*model.EmbeddingM, error)
	// List 返回 embedding 列表和总数
	List(ctx context.Context, opts *where.Options) (int64, []*model.EmbeddingM, error)

	EmbeddingExpansion
}

// EmbeddingExpansion 定义了 embedding 操作的附加方法.
// nolint: iface
type EmbeddingExpansion interface {
	// DeleteByKnowledgeID 删除指定知识的所有 embedding
	DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error
	// DeleteByChunkID 删除指定分块的所有 embedding
	DeleteByChunkID(ctx context.Context, chunkID string) error
	// DeleteByKnowledgeBaseID 删除指定知识库的所有 embedding
	DeleteByKnowledgeBaseID(ctx context.Context, kbID string) error
	// VectorSearch 向量搜索：根据查询向量返回最相似的 N 个 embedding
	// 使用 PGVector 的 HNSW 索引和 cosine 距离
	VectorSearch(ctx context.Context, kbID string, queryVector []float32, topK int) ([]*model.EmbeddingM, error)
	// GetByChunkID 获取指定分块的 embedding
	GetByChunkID(ctx context.Context, chunkID string) ([]*model.EmbeddingM, error)
}

// embeddingStore 是 EmbeddingStore 接口的实现.
type embeddingStore struct {
	store *datastore
}

// 确保 embeddingStore 实现了 EmbeddingStore 接口.
var _ EmbeddingStore = (*embeddingStore)(nil)

// newEmbeddingStore 创建 embeddingStore 的实例.
func newEmbeddingStore(store *datastore) *embeddingStore {
	return &embeddingStore{store: store}
}

// Create 创建 embedding 记录.
func (s *embeddingStore) Create(ctx context.Context, obj *model.EmbeddingM) error {
	return s.store.DB(ctx).Create(obj).Error
}

// CreateBatch 批量创建 embedding 记录.
func (s *embeddingStore) CreateBatch(ctx context.Context, objs []*model.EmbeddingM) error {
	if len(objs) == 0 {
		return nil
	}
	return s.store.DB(ctx).Create(&objs).Error
}

// Delete 根据条件删除 embedding 记录.
func (s *embeddingStore) Delete(ctx context.Context, opts *where.Options) error {
	return s.store.DB(ctx, opts).Delete(new(model.EmbeddingM)).Error
}

// Get 根据条件查询 embedding 记录.
func (s *embeddingStore) Get(ctx context.Context, opts *where.Options) (*model.EmbeddingM, error) {
	var obj model.EmbeddingM
	err := s.store.DB(ctx, opts).First(&obj).Error
	return &obj, err
}

// List 返回 embedding 列表和总数.
func (s *embeddingStore) List(ctx context.Context, opts *where.Options) (count int64, ret []*model.EmbeddingM, err error) {
	err = s.store.DB(ctx, opts).Order("id desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error
	return
}

// DeleteByKnowledgeID 删除指定知识的所有 embedding.
func (s *embeddingStore) DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error {
	return s.store.DB(ctx).Where("knowledge_id = ?", knowledgeID).Delete(new(model.EmbeddingM)).Error
}

// DeleteByChunkID 删除指定分块的所有 embedding.
func (s *embeddingStore) DeleteByChunkID(ctx context.Context, chunkID string) error {
	return s.store.DB(ctx).Where("chunk_id = ?", chunkID).Delete(new(model.EmbeddingM)).Error
}

// DeleteByKnowledgeBaseID 删除指定知识库的所有 embedding.
func (s *embeddingStore) DeleteByKnowledgeBaseID(ctx context.Context, kbID string) error {
	return s.store.DB(ctx).Where("knowledge_base_id = ?", kbID).Delete(new(model.EmbeddingM)).Error
}

// VectorSearch 向量搜索：根据查询向量返回最相似的 N 个 embedding.
// 使用 PGVector 的 HNSW 索引和 cosine 距离.
func (s *embeddingStore) VectorSearch(ctx context.Context, kbID string, queryVector []float32, topK int) ([]*model.EmbeddingM, error) {
	var list []*model.EmbeddingM

	if len(queryVector) == 0 {
		return list, nil
	}

	// 将 []float32 转换为 PGVector 格式的字符串 "[0.1,0.2,...]"
	embeddingStr := vectorToString(queryVector)

	// 根据向量维度确定使用的索引
	dimension := len(queryVector)

	// 构建查询：使用 cosine 距离 (1 - cosine_similarity)
	// 距离越小越相似
	query := s.store.DB(ctx).Where("dimension = ?", dimension)

	// 如果指定了知识库 ID，添加过滤条件
	if kbID != "" {
		query = query.Where("knowledge_base_id = ?", kbID)
	}

	// 使用 <=> 操作符计算 cosine 距离
	// PostgreSQL: embedding <=> '[...]'::halfvec(N)
	query = query.Order(fmt.Sprintf("embedding <=> '%s'::halfvec(%d)", embeddingStr, dimension))

	// 设置 topK
	if topK <= 0 {
		topK = 5
	}

	err := query.Limit(topK).Find(&list).Error
	return list, err
}

// VectorSearchWithScore 向量搜索并返回相似度分数.
func (s *embeddingStore) VectorSearchWithScore(ctx context.Context, kbID string, queryVector []float32, topK int) ([]*EmbeddingWithScore, error) {
	var results []*EmbeddingWithScore

	if len(queryVector) == 0 {
		return results, nil
	}

	embeddingStr := vectorToString(queryVector)
	dimension := len(queryVector)

	// 构建 SQL 查询，包含距离计算
	sql := `
		SELECT id, created_at, updated_at, source_id, source_type, chunk_id, 
		       knowledge_id, knowledge_base_id, tag_id, content, dimension, embedding,
		       (embedding <=> $1::halfvec($2)) as distance
		FROM embeddings
		WHERE dimension = $2
	`
	args := []interface{}{embeddingStr, dimension}

	if kbID != "" {
		sql += " AND knowledge_base_id = $3"
		args = append(args, kbID)
	}

	sql += fmt.Sprintf(" ORDER BY distance LIMIT %d", topK)

	err := s.store.DB(ctx).Raw(sql, args...).Scan(&results).Error
	return results, err
}

// EmbeddingWithScore embedding 结果带相似度分数.
type EmbeddingWithScore struct {
	model.EmbeddingM
	Distance float64 `gorm:"column:distance" json:"distance"`
}

// Score 返回相似度分数 (1 - distance).
func (e *EmbeddingWithScore) Score() float64 {
	return 1 - e.Distance
}

// GetByChunkID 获取指定分块的 embedding.
func (s *embeddingStore) GetByChunkID(ctx context.Context, chunkID string) ([]*model.EmbeddingM, error) {
	var list []*model.EmbeddingM
	err := s.store.DB(ctx).Where("chunk_id = ?", chunkID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}
