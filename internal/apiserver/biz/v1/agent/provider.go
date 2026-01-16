// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"
	"io"

	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	agentpkg "github.com/ashwinyue/eino-show/internal/pkg/agent"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/google/wire"
)

// ProviderSet Wire Provider 集合.
var ProviderSet = wire.NewSet(
	NewAgentFactory,
	ProvideAgentBizWithFactory,
)

// NewAgentFactory 创建 Agent 工厂.
func NewAgentFactory(store store.IStore) (*agentpkg.Factory, error) {
	cfg := CreateAgentFactoryConfig()
	return agentpkg.NewFactory(context.Background(), cfg)
}

// ProvideAgentBizWithFactory 提供带 Agent 工厂的 AgentBiz.
// 这是一个桥接函数，用于将 Agent 工厂注入到 AgentBiz 中.
type agentBizWithFactory struct {
	*agentBiz
	factory *agentpkg.Factory
}

func (a *agentBizWithFactory) Execute(ctx context.Context, req *v1.ExecuteRequest) (io.ReadCloser, error) {
	return executeWithFactory(ctx, a.factory, a.store, req)
}

// ProvideAgentBizWithFactory 创建带工厂的 AgentBiz.
func ProvideAgentBizWithFactory(factory *agentpkg.Factory, store store.IStore) AgentBiz {
	base := &agentBiz{store: store}
	return &agentBizWithFactory{
		agentBiz: base,
		factory:  factory,
	}
}
