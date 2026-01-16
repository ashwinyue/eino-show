// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package validation

import (
	"context"

	"github.com/ashwinyue/eino-show/internal/pkg/errno"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ValidateListAgents 校验获取 Agent 列表请求.
func (v *Validator) ValidateListAgents(ctx context.Context, rq *v1.ListAgentsRequest) error {
	_ = ctx
	_ = rq
	return nil
}

// ValidateGetAgent 校验获取 Agent 请求.
func (v *Validator) ValidateGetAgent(ctx context.Context, rq *v1.GetAgentRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("agent id cannot be empty")
	}
	return nil
}

// ValidateCreateAgent 校验创建 Agent 请求.
func (v *Validator) ValidateCreateAgent(ctx context.Context, rq *v1.CreateAgentRequest) error {
	if rq.Name == "" {
		return errno.ErrInvalidArgument.WithMessage("name cannot be empty")
	}
	if len(rq.Name) > 255 {
		return errno.ErrInvalidArgument.WithMessage("name must be less than 255 characters")
	}
	// Agent 现在通过 Config 中的 instruction 来设置系统提示词，而非直接字段
	return nil
}

// ValidateUpdateAgent 校验更新 Agent 请求.
func (v *Validator) ValidateUpdateAgent(ctx context.Context, rq *v1.UpdateAgentRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("agent id cannot be empty")
	}
	if rq.Name != nil && len(*rq.Name) > 255 {
		return errno.ErrInvalidArgument.WithMessage("name must be less than 255 characters")
	}
	return nil
}

// ValidateDeleteAgent 校验删除 Agent 请求.
func (v *Validator) ValidateDeleteAgent(ctx context.Context, rq *v1.DeleteAgentRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("agent id cannot be empty")
	}
	return nil
}
