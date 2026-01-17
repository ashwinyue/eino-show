// Package http 提供 HTTP 处理器.
// system.go - 系统相关 Handler（对齐 WeKnora /api/v1/system）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSystemInfo 获取系统信息
// GET /api/v1/system/info
func (h *Handler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"version": "1.0.0",
		},
	})
}

// ListMinioBuckets 获取 MinIO 存储桶列表
// GET /api/v1/system/minio/buckets
func (h *Handler) ListMinioBuckets(c *gin.Context) {
	// TODO: 实现 MinIO 存储桶列表
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}
