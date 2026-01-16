// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package store

import (
	"context"

	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// SessionStore 定义了 session 模块在 store 层所实现的方法.
type SessionStore interface {
	Create(ctx context.Context, obj *model.SessionM) error
	Update(ctx context.Context, obj *model.SessionM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.SessionM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.SessionM, error)

	SessionExpansion
}

// SessionExpansion 定义了会话操作的附加方法.
// nolint: iface
type SessionExpansion interface {
	// GetByTenantID 获取租户下的所有会话
	GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.SessionM, error)
	// GetByAgentID 获取使用指定 Agent 的会话列表
	GetByAgentID(ctx context.Context, agentID string) ([]*model.SessionM, error)
}

// sessionStore 是 SessionStore 接口的实现.
type sessionStore struct {
	store *datastore
	*genericstore.Store[model.SessionM]
}

// 确保 sessionStore 实现了 SessionStore 接口.
var _ SessionStore = (*sessionStore)(nil)

// newSessionStore 创建 sessionStore 的实例.
func newSessionStore(store *datastore) *sessionStore {
	return &sessionStore{
		store: store,
		Store: genericstore.NewStore[model.SessionM](store, NewLogger()),
	}
}

// GetByTenantID 获取租户下的所有会话.
func (s *sessionStore) GetByTenantID(ctx context.Context, tenantID uint64) ([]*model.SessionM, error) {
	var list []*model.SessionM
	err := s.store.DB(ctx).Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// GetByAgentID 获取使用指定 Agent 的会话列表.
func (s *sessionStore) GetByAgentID(ctx context.Context, agentID string) ([]*model.SessionM, error) {
	var list []*model.SessionM
	err := s.store.DB(ctx).Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}
