package store

import (
	"context"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// FAQStore 定义了 FAQ 模块在 store 层所实现的方法.
type FAQStore interface {
	Create(ctx context.Context, obj *model.FAQEntryM) error
	CreateBatch(ctx context.Context, objs []*model.FAQEntryM) error
	Update(ctx context.Context, obj *model.FAQEntryM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.FAQEntryM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.FAQEntryM, error)

	FAQExpansion
}

// FAQExpansion 定义了 FAQ 操作的附加方法.
type FAQExpansion interface {
	// ListByKnowledgeBaseID 获取知识库下的所有 FAQ 条目
	ListByKnowledgeBaseID(ctx context.Context, kbID string, page, pageSize int, tagID *int64, keyword, searchField, sortOrder string) (int64, []*model.FAQEntryM, error)
	// DeleteByKnowledgeBaseID 删除知识库下的所有 FAQ 条目
	DeleteByKnowledgeBaseID(ctx context.Context, kbID string) error
	// GetByStandardQuestion 根据标准问题查找 FAQ
	GetByStandardQuestion(ctx context.Context, kbID, question string) (*model.FAQEntryM, error)
	// UpdateTagBatch 批量更新标签
	UpdateTagBatch(ctx context.Context, updates map[int64]*int64) error
	// UpdateFieldsBatch 批量更新字段
	UpdateFieldsBatch(ctx context.Context, updates map[int64]map[string]interface{}) error
}

// faqStore 是 FAQStore 接口的实现.
type faqStore struct {
	store *datastore
}

// 确保 faqStore 实现了 FAQStore 接口.
var _ FAQStore = (*faqStore)(nil)

// newFAQStore 创建 faqStore 的实例.
func newFAQStore(store *datastore) *faqStore {
	return &faqStore{store: store}
}

// Create 创建 FAQ 条目.
func (s *faqStore) Create(ctx context.Context, obj *model.FAQEntryM) error {
	return s.store.DB(ctx).Create(obj).Error
}

// CreateBatch 批量创建 FAQ 条目.
func (s *faqStore) CreateBatch(ctx context.Context, objs []*model.FAQEntryM) error {
	if len(objs) == 0 {
		return nil
	}
	return s.store.DB(ctx).CreateInBatches(objs, 100).Error
}

// Update 更新 FAQ 条目.
func (s *faqStore) Update(ctx context.Context, obj *model.FAQEntryM) error {
	return s.store.DB(ctx).Save(obj).Error
}

// Delete 删除 FAQ 条目.
func (s *faqStore) Delete(ctx context.Context, opts *where.Options) error {
	return s.store.DB(ctx, opts).Delete(&model.FAQEntryM{}).Error
}

// Get 获取单个 FAQ 条目.
func (s *faqStore) Get(ctx context.Context, opts *where.Options) (*model.FAQEntryM, error) {
	var obj model.FAQEntryM
	if err := s.store.DB(ctx, opts).First(&obj).Error; err != nil {
		return nil, err
	}
	return &obj, nil
}

// List 获取 FAQ 条目列表.
func (s *faqStore) List(ctx context.Context, opts *where.Options) (int64, []*model.FAQEntryM, error) {
	var total int64
	var list []*model.FAQEntryM

	db := s.store.DB(ctx, opts)
	if err := db.Model(&model.FAQEntryM{}).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	if err := db.Find(&list).Error; err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

// ListByKnowledgeBaseID 获取知识库下的所有 FAQ 条目.
func (s *faqStore) ListByKnowledgeBaseID(ctx context.Context, kbID string, page, pageSize int, tagID *int64, keyword, searchField, sortOrder string) (int64, []*model.FAQEntryM, error) {
	var total int64
	var list []*model.FAQEntryM

	db := s.store.DB(ctx).Model(&model.FAQEntryM{}).Where("knowledge_base_id = ?", kbID)

	// 标签筛选
	if tagID != nil {
		if *tagID == 0 {
			db = db.Where("tag_id IS NULL")
		} else {
			db = db.Where("tag_id = ?", *tagID)
		}
	}

	// 关键词搜索
	if keyword != "" {
		switch searchField {
		case "standard_question":
			db = db.Where("standard_question ILIKE ?", "%"+keyword+"%")
		case "similar_questions":
			db = db.Where("similar_questions ILIKE ?", "%"+keyword+"%")
		case "answers":
			db = db.Where("answers ILIKE ?", "%"+keyword+"%")
		default:
			// 搜索全部字段
			db = db.Where("standard_question ILIKE ? OR similar_questions ILIKE ? OR answers ILIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	}

	// 计数
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	// 排序
	if sortOrder == "asc" {
		db = db.Order("updated_at ASC")
	} else {
		db = db.Order("updated_at DESC")
	}

	// 分页
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		db = db.Offset(offset).Limit(pageSize)
	}

	if err := db.Find(&list).Error; err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

// DeleteByKnowledgeBaseID 删除知识库下的所有 FAQ 条目.
func (s *faqStore) DeleteByKnowledgeBaseID(ctx context.Context, kbID string) error {
	return s.store.DB(ctx).Where("knowledge_base_id = ?", kbID).Delete(&model.FAQEntryM{}).Error
}

// GetByStandardQuestion 根据标准问题查找 FAQ.
func (s *faqStore) GetByStandardQuestion(ctx context.Context, kbID, question string) (*model.FAQEntryM, error) {
	var obj model.FAQEntryM
	err := s.store.DB(ctx).Where("knowledge_base_id = ? AND standard_question = ?", kbID, question).First(&obj).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

// UpdateTagBatch 批量更新标签.
func (s *faqStore) UpdateTagBatch(ctx context.Context, updates map[int64]*int64) error {
	for id, tagID := range updates {
		if err := s.store.DB(ctx).Model(&model.FAQEntryM{}).Where("id = ?", id).Update("tag_id", tagID).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpdateFieldsBatch 批量更新字段.
func (s *faqStore) UpdateFieldsBatch(ctx context.Context, updates map[int64]map[string]interface{}) error {
	for id, fields := range updates {
		if err := s.store.DB(ctx).Model(&model.FAQEntryM{}).Where("id = ?", id).Updates(fields).Error; err != nil {
			return err
		}
	}
	return nil
}
