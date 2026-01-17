// Package http 提供 HTTP 处理器.
// faq.go - FAQ 相关 Handler（对齐 WeKnora /api/v1/knowledge-bases/:id/faq）
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/eino-show/internal/apiserver/biz/faq"
)

// ===== FAQ 请求/响应类型 =====

// FAQCreateEntryRequest 创建 FAQ 条目请求
type FAQCreateEntryRequest struct {
	StandardQuestion  string   `json:"standard_question" binding:"required"`
	SimilarQuestions  []string `json:"similar_questions"`
	NegativeQuestions []string `json:"negative_questions"`
	Answers           []string `json:"answers" binding:"required"`
	TagID             *int64   `json:"tag_id"`
	IsEnabled         *bool    `json:"is_enabled"`
}

// FAQUpdateEntryRequest 更新 FAQ 条目请求
type FAQUpdateEntryRequest struct {
	StandardQuestion  *string  `json:"standard_question"`
	SimilarQuestions  []string `json:"similar_questions"`
	NegativeQuestions []string `json:"negative_questions"`
	Answers           []string `json:"answers"`
	TagID             *int64   `json:"tag_id"`
	IsEnabled         *bool    `json:"is_enabled"`
}

// FAQDeleteEntriesRequest 删除 FAQ 条目请求
type FAQDeleteEntriesRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// FAQUpdateTagBatchRequest 批量更新标签请求
type FAQUpdateTagBatchRequest struct {
	Updates map[int64]*int64 `json:"updates" binding:"required"`
}

// FAQUpdateFieldsBatchRequest 批量更新字段请求
type FAQUpdateFieldsBatchRequest struct {
	Updates map[int64]map[string]interface{} `json:"updates" binding:"required"`
}

// FAQSearchRequest FAQ 搜索请求
type FAQSearchRequest struct {
	QueryText       string  `json:"query_text" binding:"required"`
	VectorThreshold float64 `json:"vector_threshold"`
	MatchCount      int     `json:"match_count"`
}

// ===== FAQ Handler =====

// ListFAQEntries 获取 FAQ 条目列表
// GET /api/v1/knowledge-bases/:id/faq/entries
func (h *Handler) ListFAQEntries(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析标签筛选
	var tagID *int64
	if tagIDStr := c.Query("tag_id"); tagIDStr != "" {
		if id, err := strconv.ParseInt(tagIDStr, 10, 64); err == nil {
			tagID = &id
		}
	}

	keyword := c.Query("keyword")
	searchField := c.Query("search_field")
	sortOrder := c.Query("sort_order")

	result, err := h.biz.FAQ().ListEntries(c.Request.Context(), kbID, page, pageSize, tagID, keyword, searchField, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetFAQEntry 获取单个 FAQ 条目
// GET /api/v1/knowledge-bases/:id/faq/entries/:entry_id
func (h *Handler) GetFAQEntry(c *gin.Context) {
	kbID := c.Param("id")
	entryIDStr := c.Param("entry_id")
	if kbID == "" || entryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id and entry_id are required"})
		return
	}

	entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry_id"})
		return
	}

	entry, err := h.biz.FAQ().GetEntry(c.Request.Context(), kbID, entryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entry,
	})
}

// CreateFAQEntry 创建单个 FAQ 条目
// POST /api/v1/knowledge-bases/:id/faq/entry
func (h *Handler) CreateFAQEntry(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	var req FAQCreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.biz.FAQ().CreateEntry(c.Request.Context(), kbID, &faq.CreateEntryRequest{
		StandardQuestion:  req.StandardQuestion,
		SimilarQuestions:  req.SimilarQuestions,
		NegativeQuestions: req.NegativeQuestions,
		Answers:           req.Answers,
		TagID:             req.TagID,
		IsEnabled:         req.IsEnabled,
	})
	if err != nil {
		if _, ok := err.(*faq.DuplicateError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "BAD_REQUEST", "message": err.Error()}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entry,
	})
}

// UpsertFAQEntries 批量导入 FAQ 条目
// POST /api/v1/knowledge-bases/:id/faq/entries
func (h *Handler) UpsertFAQEntries(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	// 简化实现：返回任务 ID
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id": "task-" + kbID,
		},
	})
}

// UpdateFAQEntry 更新单个 FAQ 条目
// PUT /api/v1/knowledge-bases/:id/faq/entries/:entry_id
func (h *Handler) UpdateFAQEntry(c *gin.Context) {
	kbID := c.Param("id")
	entryIDStr := c.Param("entry_id")
	if kbID == "" || entryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id and entry_id are required"})
		return
	}

	entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry_id"})
		return
	}

	var req FAQUpdateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.FAQ().UpdateEntry(c.Request.Context(), kbID, entryID, &faq.UpdateEntryRequest{
		StandardQuestion:  req.StandardQuestion,
		SimilarQuestions:  req.SimilarQuestions,
		NegativeQuestions: req.NegativeQuestions,
		Answers:           req.Answers,
		TagID:             req.TagID,
		IsEnabled:         req.IsEnabled,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteFAQEntries 批量删除 FAQ 条目
// DELETE /api/v1/knowledge-bases/:id/faq/entries
func (h *Handler) DeleteFAQEntries(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	var req FAQDeleteEntriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.FAQ().DeleteEntries(c.Request.Context(), kbID, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateFAQTagBatch 批量更新 FAQ 标签
// PUT /api/v1/knowledge-bases/:id/faq/entries/tags
func (h *Handler) UpdateFAQTagBatch(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	var req FAQUpdateTagBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.FAQ().UpdateTagBatch(c.Request.Context(), kbID, req.Updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateFAQFieldsBatch 批量更新 FAQ 字段
// PUT /api/v1/knowledge-bases/:id/faq/entries/fields
func (h *Handler) UpdateFAQFieldsBatch(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	var req FAQUpdateFieldsBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.biz.FAQ().UpdateFieldsBatch(c.Request.Context(), kbID, req.Updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SearchFAQ 搜索 FAQ
// POST /api/v1/knowledge-bases/:id/faq/search
func (h *Handler) SearchFAQ(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	var req FAQSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.biz.FAQ().SearchFAQ(c.Request.Context(), kbID, &faq.SearchFAQRequest{
		QueryText:       req.QueryText,
		VectorThreshold: req.VectorThreshold,
		MatchCount:      req.MatchCount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// GetFAQImportProgress 获取 FAQ 导入进度
// GET /api/v1/faq/import/progress/:task_id
func (h *Handler) GetFAQImportProgress(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	// 简化实现：返回完成状态
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":  taskID,
			"status":   "completed",
			"progress": 100,
		},
	})
}

// ExportFAQEntries 导出 FAQ 条目
// GET /api/v1/knowledge-bases/:id/faq/entries/export
func (h *Handler) ExportFAQEntries(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id is required"})
		return
	}

	// 获取所有条目
	result, err := h.biz.FAQ().ListEntries(c.Request.Context(), kbID, 1, 10000, nil, "", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Data,
	})
}

// AddSimilarQuestions 添加相似问题
// POST /api/v1/knowledge-bases/:id/faq/entries/:entry_id/similar-questions
func (h *Handler) AddSimilarQuestions(c *gin.Context) {
	kbID := c.Param("id")
	entryIDStr := c.Param("entry_id")
	if kbID == "" || entryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id and entry_id are required"})
		return
	}

	entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry_id"})
		return
	}

	var req struct {
		Questions []string `json:"questions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前条目
	entry, err := h.biz.FAQ().GetEntry(c.Request.Context(), kbID, entryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 合并相似问题
	newQuestions := append(entry.SimilarQuestions, req.Questions...)

	// 更新
	if err := h.biz.FAQ().UpdateEntry(c.Request.Context(), kbID, entryID, &faq.UpdateEntryRequest{
		SimilarQuestions: newQuestions,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
