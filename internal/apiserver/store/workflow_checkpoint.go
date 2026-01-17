// Package store 提供 Workflow 检查点存储.
package store

import (
	"context"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// WorkflowCheckpointStore Workflow 检查点存储接口.
type WorkflowCheckpointStore interface {
	Create(ctx context.Context, obj *model.WorkflowCheckpointM) error
	Update(ctx context.Context, obj *model.WorkflowCheckpointM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.WorkflowCheckpointM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.WorkflowCheckpointM, error)

	WorkflowCheckpointExpansion
}

// WorkflowCheckpointExpansion 定义了 Workflow 检查点操作的附加方法.
// nolint: iface
type WorkflowCheckpointExpansion interface {
	// GetBySessionID 获取会话的最新检查点
	GetBySessionID(ctx context.Context, sessionID string) (*model.WorkflowCheckpointM, error)
	// GetPendingBySession 获取会话待处理的检查点
	GetPendingBySession(ctx context.Context, sessionID string) (*model.WorkflowCheckpointM, error)
	// UpdateStatus 更新检查点状态
	UpdateStatus(ctx context.Context, id string, status string) error
	// DeleteBySessionID 删除会话的所有检查点
	DeleteBySessionID(ctx context.Context, sessionID string) error
}

// workflowCheckpointStore Workflow 检查点存储实现.
type workflowCheckpointStore struct {
	store *datastore
}

// 确保 workflowCheckpointStore 实现了 WorkflowCheckpointStore 接口.
var _ WorkflowCheckpointStore = (*workflowCheckpointStore)(nil)

// newWorkflowCheckpointStore 创建 WorkflowCheckpointStore 实例.
func newWorkflowCheckpointStore(store *datastore) *workflowCheckpointStore {
	return &workflowCheckpointStore{store: store}
}

// Create 创建检查点.
func (s *workflowCheckpointStore) Create(ctx context.Context, obj *model.WorkflowCheckpointM) error {
	return s.store.DB(ctx).Create(obj).Error
}

// Update 更新检查点.
func (s *workflowCheckpointStore) Update(ctx context.Context, obj *model.WorkflowCheckpointM) error {
	return s.store.DB(ctx).Where("id = ?", obj.ID).Updates(obj).Error
}

// Delete 删除检查点.
func (s *workflowCheckpointStore) Delete(ctx context.Context, opts *where.Options) error {
	return s.store.DB(ctx, opts).Delete(new(model.WorkflowCheckpointM)).Error
}

// Get 获取检查点.
func (s *workflowCheckpointStore) Get(ctx context.Context, opts *where.Options) (*model.WorkflowCheckpointM, error) {
	var cp model.WorkflowCheckpointM
	err := s.store.DB(ctx, opts).First(&cp).Error
	return &cp, err
}

// List 获取检查点列表.
func (s *workflowCheckpointStore) List(ctx context.Context, opts *where.Options) (int64, []*model.WorkflowCheckpointM, error) {
	var count int64
	var list []*model.WorkflowCheckpointM
	err := s.store.DB(ctx, opts).Order("created_at DESC").Find(&list).Offset(-1).Limit(-1).Count(&count).Error
	return count, list, err
}

// GetBySessionID 获取会话的最新检查点.
func (s *workflowCheckpointStore) GetBySessionID(ctx context.Context, sessionID string) (*model.WorkflowCheckpointM, error) {
	var cp model.WorkflowCheckpointM
	err := s.store.DB(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		First(&cp).Error
	return &cp, err
}

// GetPendingBySession 获取会话待处理的检查点.
func (s *workflowCheckpointStore) GetPendingBySession(ctx context.Context, sessionID string) (*model.WorkflowCheckpointM, error) {
	var cp model.WorkflowCheckpointM
	err := s.store.DB(ctx).
		Where("session_id = ?", sessionID).
		Where("status = ?", "pending").
		Order("created_at DESC").
		First(&cp).Error
	return &cp, err
}

// UpdateStatus 更新检查点状态.
func (s *workflowCheckpointStore) UpdateStatus(ctx context.Context, id string, status string) error {
	return s.store.DB(ctx).
		Model(&model.WorkflowCheckpointM{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// DeleteBySessionID 删除会话的所有检查点.
func (s *workflowCheckpointStore) DeleteBySessionID(ctx context.Context, sessionID string) error {
	return s.store.DB(ctx).
		Where("session_id = ?", sessionID).
		Delete(new(model.WorkflowCheckpointM)).Error
}
