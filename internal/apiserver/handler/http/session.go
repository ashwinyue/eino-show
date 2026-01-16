// Package http 提供 HTTP 处理器.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// CreateSession 创建会话.
func (h *Handler) CreateSession(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.SessionV1().Create, h.val.ValidateCreateSession)
}

// GetSession 获取会话详情.
func (h *Handler) GetSession(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.SessionV1().Get, h.val.ValidateGetSession)
}

// ListSessions 获取会话列表.
func (h *Handler) ListSessions(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.SessionV1().List, h.val.ValidateListSessions)
}

// UpdateSession 更新会话.
func (h *Handler) UpdateSession(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.SessionV1().Update, h.val.ValidateUpdateSession)
}

// DeleteSession 删除会话.
func (h *Handler) DeleteSession(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.SessionV1().Delete, h.val.ValidateDeleteSession)
}

// QA 问答（流式）.
func (h *Handler) QA(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id is required"})
		return
	}

	// 使用 proto 定义的 ExecuteRequest 类型
	var req v1.ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 调用 Biz 层获取流式响应
	reader, err := h.biz.SessionV1().QA(c.Request.Context(), sessionID, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// 流式写入响应
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			c.Writer.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
}
