package http

import (
	genericoptions "github.com/ashwinyue/eino-show/pkg/options"

	"github.com/ashwinyue/eino-show/internal/apiserver/biz"
	"github.com/ashwinyue/eino-show/internal/apiserver/pkg/validation"
)

// Handler 处理博客模块的请求.
type Handler struct {
	biz             biz.IBiz
	val             *validation.Validator
	webSearchConfig *genericoptions.WebSearchOptions
}

// NewHandler 创建新的 Handler 实例.
func NewHandler(biz biz.IBiz, val *validation.Validator, webSearchConfig *genericoptions.WebSearchOptions) *Handler {
	return &Handler{
		biz:             biz,
		val:             val,
		webSearchConfig: webSearchConfig,
	}
}
