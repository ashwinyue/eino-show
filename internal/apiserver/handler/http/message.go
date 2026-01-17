// Package http 提供 HTTP 处理器.
// message.go - 消息相关 Handler（对齐 WeKnora /api/v1/messages）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoadMessages 加载会话消息
// GET /api/v1/messages/:session_id/load
// Query: limit, before_time
func (h *Handler) LoadMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// 获取会话消息
	messages, err := h.biz.Session().GetMessages(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
	})
}

// DeleteMessage 删除消息
// DELETE /api/v1/messages/:session_id/:id
func (h *Handler) DeleteMessage(c *gin.Context) {
	sessionID := c.Param("session_id")
	msgID := c.Param("id")
	if sessionID == "" || msgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id and id are required"})
		return
	}

	if err := h.biz.Session().DeleteMessage(c.Request.Context(), sessionID, msgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message deleted successfully",
	})
}
