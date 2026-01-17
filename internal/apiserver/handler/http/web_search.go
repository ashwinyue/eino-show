// Package http 提供 HTTP 处理器.
// web_search.go - 网络搜索相关 Handler（对齐 WeKnora /api/v1/web-search）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetWebSearchProviders 获取网络搜索提供商
// GET /api/v1/web-search/providers
func (h *Handler) GetWebSearchProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}
