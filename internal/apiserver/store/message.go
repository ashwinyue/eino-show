package store

import (
	"context"
	"time"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	genericstore "github.com/ashwinyue/eino-show/pkg/store"
	"github.com/ashwinyue/eino-show/pkg/store/where"
	"gorm.io/gorm"
)

// MessageStore 定义了 message 模块在 store 层所实现的方法.
type MessageStore interface {
	Create(ctx context.Context, obj *model.MessageM) error
	// CreateWithSessionItem 在事务中同时创建 Message 和对应的 SessionItem（对齐 WeKnora）
	CreateWithSessionItem(ctx context.Context, obj *model.MessageM) error
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
	// GetBySessionIDBeforeTime 获取会话在指定时间之前的消息（用于分页加载）
	GetBySessionIDBeforeTime(ctx context.Context, sessionID string, beforeTime time.Time, limit int) ([]*model.MessageM, error)
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

// GetBySessionIDBeforeTime 获取会话在指定时间之前的消息（用于分页加载）.
func (s *messageStore) GetBySessionIDBeforeTime(ctx context.Context, sessionID string, beforeTime time.Time, limit int) ([]*model.MessageM, error) {
	var list []*model.MessageM
	err := s.store.DB(ctx).Where("session_id = ? AND created_at < ?", sessionID, beforeTime).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	// 反转顺序，使消息按时间正序排列（最早的在前）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

// CreateWithSessionItem 在事务中同时创建 Message 和对应的 SessionItem（对齐 WeKnora）.
// 这是创建消息的推荐方法，确保 message 和 session_item 表的一致性.
func (s *messageStore) CreateWithSessionItem(ctx context.Context, msg *model.MessageM) error {
	return s.store.core.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建 Message 记录
		if err := tx.Create(msg).Error; err != nil {
			return err
		}

		// 2. 创建对应的 SessionItem 记录
		// 获取当前最大的 sort_order
		var maxOrder int32
		if err := tx.Model(&model.SessionItemM{}).
			Where("session_id = ?", msg.SessionID).
			Select("COALESCE(MAX(sort_order), 0)").
			Scan(&maxOrder).Error; err != nil {
			return err
		}

		sessionItem := &model.SessionItemM{
			SessionID:  msg.SessionID,
			Type:       SessionItemTypeMessage,
			MessageID:  &msg.ID,
			SortOrder:  maxOrder + 1,
		}

		return tx.Create(sessionItem).Error
	})
}
