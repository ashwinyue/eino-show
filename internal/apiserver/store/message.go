package store

import (
	"context"

	genericstore "github.com/ashwinyue/eino-show/pkg/store"
	"github.com/ashwinyue/eino-show/pkg/store/where"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// MessageStore 定义了 message 模块在 store 层所实现的方法.
type MessageStore interface {
	Create(ctx context.Context, obj *model.MessageM) error
	Update(ctx context.Context, obj *model.MessageM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.MessageM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.MessageM, error)

	MessageExpansion
}

// MessageExpansion 定义了消息操作的附加方法.
// nolint: iface
type MessageExpansion interface {
	// GetBySessionID 获取会话的所有消息（按创建时间排序）
	GetBySessionID(ctx context.Context, sessionID string) ([]*model.MessageM, error)
	// GetRecentBySessionID 获取会话最近的 N 条消息
	GetRecentBySessionID(ctx context.Context, sessionID string, limit int) ([]*model.MessageM, error)
}

// messageStore 是 MessageStore 接口的实现.
type messageStore struct {
	store *datastore
	*genericstore.Store[model.MessageM]
}

// 确保 messageStore 实现了 MessageStore 接口.
var _ MessageStore = (*messageStore)(nil)

// newMessageStore 创建 messageStore 的实例.
func newMessageStore(store *datastore) *messageStore {
	return &messageStore{
		store: store,
		Store: genericstore.NewStore[model.MessageM](store, NewLogger()),
	}
}

// GetBySessionID 获取会话的所有消息.
func (s *messageStore) GetBySessionID(ctx context.Context, sessionID string) ([]*model.MessageM, error) {
	var list []*model.MessageM
	err := s.store.DB(ctx).Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&list).Error
	return list, err
}

// GetRecentBySessionID 获取会话最近的 N 条消息.
func (s *messageStore) GetRecentBySessionID(ctx context.Context, sessionID string, limit int) ([]*model.MessageM, error) {
	var list []*model.MessageM
	err := s.store.DB(ctx).Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	// 反转顺序，使消息按时间正序排列
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}
