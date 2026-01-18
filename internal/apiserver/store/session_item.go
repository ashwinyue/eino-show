// Package store 提供 session_items 表的存储层实现.
package store

import (
	"context"

	genericstore "github.com/ashwinyue/eino-show/pkg/store"
	"github.com/ashwinyue/eino-show/pkg/store/where"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// SessionItemType 定义会话项的类型（对齐 WeKnora）.
// 使用 string 类型以与 summary.go 中的常量保持兼容
type SessionItemType = string

const (
	// SessionItemTypeMessage 普通消息类型（关联 messages 表）
	SessionItemTypeMessage SessionItemType = "message"
)

// SessionItemStore 定义了 session_item 模块在 store 层所实现的方法.
type SessionItemStore interface {
	Create(ctx context.Context, obj *model.SessionItemM) error
	Update(ctx context.Context, obj *model.SessionItemM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.SessionItemM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.SessionItemM, error)

	SessionItemExpansion
}

// SessionItemExpansion 定义了 session_item 操作的附加方法.
type SessionItemExpansion interface {
	// GetMaxSortOrder 获取会话中最大的 sort_order
	GetMaxSortOrder(ctx context.Context, sessionID string) (int32, error)
	// GetBySessionID 获取会话的所有 session_items
	GetBySessionID(ctx context.Context, sessionID string) ([]*model.SessionItemM, error)
}

// sessionItemStore 是 SessionItemStore 接口的实现.
type sessionItemStore struct {
	store *datastore
	*genericstore.Store[model.SessionItemM]
}

// 确保 sessionItemStore 实现了 SessionItemStore 接口.
var _ SessionItemStore = (*sessionItemStore)(nil)

// newSessionItemStore 创建 sessionItemStore 的实例.
func newSessionItemStore(store *datastore) *sessionItemStore {
	return &sessionItemStore{
		store: store,
		Store: genericstore.NewStore[model.SessionItemM](store, NewLogger()),
	}
}

// GetMaxSortOrder 获取会话中最大的 sort_order.
func (s *sessionItemStore) GetMaxSortOrder(ctx context.Context, sessionID string) (int32, error) {
	var maxOrder int32
	err := s.store.DB(ctx).Model(&model.SessionItemM{}).
		Where("session_id = ?", sessionID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder, err
}

// GetBySessionID 获取会话的所有 session_items（按 sort_order 排序）.
func (s *sessionItemStore) GetBySessionID(ctx context.Context, sessionID string) ([]*model.SessionItemM, error) {
	var list []*model.SessionItemM
	err := s.store.DB(ctx).Where("session_id = ?", sessionID).
		Order("sort_order ASC").
		Find(&list).Error
	return list, err
}
