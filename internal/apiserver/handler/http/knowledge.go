// Package http 提供 HTTP 处理器.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ListKnowledgeBases 获取知识库列表.
func (h *Handler) ListKnowledgeBases(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.KnowledgeV1().ListKnowledgeBases, h.val.ValidateListKnowledgeBases)
}

// GetKnowledgeBase 获取知识库详情.
func (h *Handler) GetKnowledgeBase(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.KnowledgeV1().GetKnowledgeBase, h.val.ValidateGetKnowledgeBase)
}

// CreateKnowledgeBase 创建知识库.
func (h *Handler) CreateKnowledgeBase(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.KnowledgeV1().CreateKnowledgeBase, h.val.ValidateCreateKnowledgeBase)
}

// UpdateKnowledgeBase 更新知识库.
func (h *Handler) UpdateKnowledgeBase(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.KnowledgeV1().UpdateKnowledgeBase, h.val.ValidateUpdateKnowledgeBase)
}

// DeleteKnowledgeBase 删除知识库.
func (h *Handler) DeleteKnowledgeBase(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.KnowledgeV1().DeleteKnowledgeBase, h.val.ValidateDeleteKnowledgeBase)
}

// GetKnowledgeStats 获取知识库统计信息.
func (h *Handler) GetKnowledgeStats(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.KnowledgeV1().GetKnowledgeStats, h.val.ValidateGetKnowledgeStats)
}

// ListKnowledges 获取知识列表.
func (h *Handler) ListKnowledges(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	req := &v1.ListKnowledgesRequest{KbId: kbID}
	knowledges, err := h.biz.KnowledgeV1().ListKnowledges(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"knowledge": knowledges.Knowledge, "total": knowledges.Total})
}

// DeleteKnowledge 删除知识项.
func (h *Handler) DeleteKnowledge(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.KnowledgeV1().DeleteKnowledge, h.val.ValidateDeleteKnowledge)
}
