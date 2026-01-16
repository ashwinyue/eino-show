// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package gin

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/token"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	"github.com/ashwinyue/eino-show/internal/pkg/errno"
	"github.com/ashwinyue/eino-show/internal/pkg/log"
)

const (
	// defaultTenantID 是默认租户 ID，当用户未分配租户时使用.
	defaultTenantID uint64 = 1
)

// UserRetriever 用于根据用户名获取用户的接口.
type UserRetriever interface {
	// GetUser 根据用户ID获取用户信息
	GetUser(ctx context.Context, userID string) (*model.UserM, error)
}

// AuthnMiddleware 是一个认证中间件，用于从 gin.Context 中提取 token 并验证 token 是否合法.
func AuthnMiddleware(retriever UserRetriever) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 JWT Token
		userID, err := token.ParseRequest(c)
		if err != nil {
			core.WriteResponse(c, nil, errno.ErrTokenInvalid.WithMessage(err.Error()))
			c.Abort()
			return
		}

		log.Debugw("Token parsing successful", "userID", userID)

		user, err := retriever.GetUser(c, userID)
		if err != nil {
			core.WriteResponse(c, nil, errno.ErrUnauthenticated.WithMessage(err.Error()))
			c.Abort()
			return
		}

		// 获取租户 ID，如果用户未分配租户则使用默认租户
		tenantID := defaultTenantID
		if user.TenantID != nil {
			tenantID = uint64(*user.TenantID)
		}

		ctx := contextx.WithUserID(c.Request.Context(), user.ID)
		ctx = contextx.WithUsername(ctx, user.Username)
		ctx = contextx.WithTenantID(ctx, tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
