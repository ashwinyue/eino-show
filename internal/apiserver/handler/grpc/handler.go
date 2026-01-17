// Package grpc 提供 gRPC 处理器.
package grpc

import (
	"github.com/ashwinyue/eino-show/internal/apiserver/biz"
	apiv1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// Handler 负责处理博客模块的请求.
type Handler struct {
	apiv1.UnimplementedMiniBlogServer

	biz biz.IBiz
}

// NewHandler 创建一个新的 Handler 实例.
func NewHandler(biz biz.IBiz) *Handler {
	return &Handler{
		biz: biz,
	}
}
