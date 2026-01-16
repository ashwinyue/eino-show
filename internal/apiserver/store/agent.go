package store

import (
	"context"

	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// CustomAgentStore 定义了 custom_agent 模块在 store 层所实现的方法.
type CustomAgentStore interface {
	Create(ctx context.Context, obj *model.CustomAgentM) error
	Update(ctx context.Context, obj *model.CustomAgentM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.CustomAgentM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.CustomAgentM, error)

	CustomAgentExpansion
}

// CustomAgentExpansion 定义了 Agent 操作的附加方法.
// nolint: iface
type CustomAgentExpansion interface {
	// GetByTenantID 获取租户下的所有 Agent
	GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.CustomAgentM, error)
	// GetBuiltinAgents 获取所有内置 Agent
	GetBuiltinAgents(ctx context.Context) ([]*model.CustomAgentM, error)
}

// customAgentStore 是 CustomAgentStore 接口的实现.
type customAgentStore struct {
	store *datastore
	*genericstore.Store[model.CustomAgentM]
}

// 确保 customAgentStore 实现了 CustomAgentStore 接口.
var _ CustomAgentStore = (*customAgentStore)(nil)

// newCustomAgentStore 创建 customAgentStore 的实例.
func newCustomAgentStore(store *datastore) *customAgentStore {
	return &customAgentStore{
		store: store,
		Store: genericstore.NewStore[model.CustomAgentM](store, NewLogger()),
	}
}

// GetByTenantID 获取租户下的所有 Agent.
func (s *customAgentStore) GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.CustomAgentM, error) {
	var list []*model.CustomAgentM
	err := s.store.DB(ctx).Where("tenant_id = ? AND (deleted_at IS NULL OR deleted_at > '0001-01-01')", tenantID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// GetBuiltinAgents 获取所有内置 Agent.
func (s *customAgentStore) GetBuiltinAgents(ctx context.Context) ([]*model.CustomAgentM, error) {
	var list []*model.CustomAgentM
	err := s.store.DB(ctx).Where("is_builtin = ? AND (deleted_at IS NULL OR deleted_at > '0001-01-01')", true).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}
