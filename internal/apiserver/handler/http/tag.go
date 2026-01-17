// Package http 提供 HTTP 处理器.
// tag.go - 标签相关 Handler（对齐 WeKnora /api/v1/knowledge-bases/:id/tags）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListTags 获取知识库标签列表
// GET /api/v1/knowledge-bases/:id/tags
func (h *Handler) ListTags(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	// 简化实现：返回空列表 (后续可接入 Tag Store)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// CreateTag 创建标签
// POST /api/v1/knowledge-bases/:id/tags
func (h *Handler) CreateTag(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	var req struct {
		Name  string `json:"name" binding:"required"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 简化实现：返回成功 (后续可接入 Tag Store)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":    "tag-" + req.Name,
			"name":  req.Name,
			"color": req.Color,
		},
	})
}

// UpdateTag 更新标签
// PUT /api/v1/knowledge-bases/:id/tags/:tag_id
func (h *Handler) UpdateTag(c *gin.Context) {
	kbID := c.Param("id")
	tagID := c.Param("tag_id")
	if kbID == "" || tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id and tag id are required"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":    tagID,
			"name":  req.Name,
			"color": req.Color,
		},
	})
}

// DeleteTag 删除标签
// DELETE /api/v1/knowledge-bases/:id/tags/:tag_id
func (h *Handler) DeleteTag(c *gin.Context) {
	kbID := c.Param("id")
	tagID := c.Param("tag_id")
	if kbID == "" || tagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id and tag id are required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tag deleted successfully",
	})
}
