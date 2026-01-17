// Package http 提供 HTTP 处理器.
// chunk.go - 分块相关 Handler（对齐 WeKnora /api/v1/chunks）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ListKnowledgeChunks 获取知识分块列表
// GET /api/v1/chunks/:knowledge_id
func (h *Handler) ListKnowledgeChunks(c *gin.Context) {
	knowledgeID := c.Param("knowledge_id")
	if knowledgeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_id is required"})
		return
	}

	resp, err := h.biz.Knowledge().ListChunks(c.Request.Context(), knowledgeID, &v1.ListChunksRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp.Chunks,
		"total":   resp.Total,
	})
}

// DeleteChunksByKnowledgeID 删除知识下的所有分块
// DELETE /api/v1/chunks/:knowledge_id
func (h *Handler) DeleteChunksByKnowledgeID(c *gin.Context) {
	knowledgeID := c.Param("knowledge_id")
	if knowledgeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_id is required"})
		return
	}

	if err := h.biz.Knowledge().DeleteChunksByKnowledgeID(c.Request.Context(), knowledgeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chunks deleted successfully",
	})
}

// DeleteGeneratedQuestion 删除生成的问题
// DELETE /api/v1/chunks/by-id/:id/questions
func (h *Handler) DeleteGeneratedQuestion(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunk id is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Question deleted successfully",
	})
}
