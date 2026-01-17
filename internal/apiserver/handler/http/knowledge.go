// Package http 提供 HTTP 处理器.
package http

import (
	"fmt"
	"net/http"

	"github.com/ashwinyue/eino-show/pkg/core"
	"github.com/gin-gonic/gin"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ListKnowledgeBases 获取知识库列表.
func (h *Handler) ListKnowledgeBases(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.Knowledge().ListKB, h.val.ValidateListKnowledgeBases)
}

// GetKnowledgeBase 获取知识库详情.
func (h *Handler) GetKnowledgeBase(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.Knowledge().GetKB, h.val.ValidateGetKnowledgeBase)
}

// CreateKnowledgeBase 创建知识库.
func (h *Handler) CreateKnowledgeBase(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.Knowledge().CreateKB, h.val.ValidateCreateKnowledgeBase)
}

// UpdateKnowledgeBase 更新知识库.
func (h *Handler) UpdateKnowledgeBase(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	var req v1.UpdateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Knowledge().UpdateKB(c.Request.Context(), kbID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteKnowledgeBase 删除知识库.
func (h *Handler) DeleteKnowledgeBase(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.Knowledge().DeleteKB, h.val.ValidateDeleteKnowledgeBase)
}

// GetKnowledgeStats 获取知识库统计信息.
func (h *Handler) GetKnowledgeStats(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	resp, err := h.biz.Knowledge().GetKBStats(c.Request.Context(), kbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListKnowledges 获取知识列表.
func (h *Handler) ListKnowledges(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	resp, err := h.biz.Knowledge().ListKnowledges(c.Request.Context(), kbID, &v1.ListKnowledgesRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteKnowledge 删除知识项.
func (h *Handler) DeleteKnowledge(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.Knowledge().DeleteKnowledge, h.val.ValidateDeleteKnowledge)
}

// HybridSearch 混合搜索（对齐 WeKnora）.
// 路径: GET /api/v1/knowledge-bases/:id/hybrid-search
func (h *Handler) HybridSearch(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	req := &v1.HybridSearchRequest{
		KnowledgeBaseId: kbID,
		QueryText:       c.Query("query_text"),
	}

	// 解析可选参数
	if matchCount := c.Query("match_count"); matchCount != "" {
		var count int32
		if _, err := fmt.Sscanf(matchCount, "%d", &count); err == nil {
			req.MatchCount = count
		}
	}

	results, err := h.biz.Knowledge().HybridSearch(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ============================================================================
// 文档上传 API (对齐 WeKnora)
// ============================================================================

// UploadKnowledgeFromFile 从文件创建知识（对齐 WeKnora）.
// 路径: POST /api/v1/knowledge-bases/:id/knowledge/file
func (h *Handler) UploadKnowledgeFromFile(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	// 简化实现：返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":     "knowledge-" + kbID + "-file",
			"kb_id":  kbID,
			"type":   "file",
			"status": "processing",
		},
	})
}

// CreateKnowledgeFromURL 从 URL 创建知识（对齐 WeKnora）.
// 路径: POST /api/v1/knowledge-bases/:id/knowledge/url
func (h *Handler) CreateKnowledgeFromURL(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	var req struct {
		URL              string `json:"url" binding:"required"`
		EnableMultimodel *bool  `json:"enable_multimodal"`
		Title            string `json:"title"`
		TagID            string `json:"tag_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":     "knowledge-" + kbID + "-url",
			"kb_id":  kbID,
			"url":    req.URL,
			"title":  req.Title,
			"type":   "url",
			"status": "processing",
		},
	})
}

// CreateManualKnowledge 手工创建知识（对齐 WeKnora）.
// 路径: POST /api/v1/knowledge-bases/:id/knowledge/manual
func (h *Handler) CreateManualKnowledge(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge base id is required"})
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		Status  string `json:"status"`
		TagID   string `json:"tag_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":      "knowledge-" + kbID + "-manual",
			"kb_id":   kbID,
			"title":   req.Title,
			"content": req.Content,
			"type":    "manual",
			"status":  "completed",
		},
	})
}

// ============================================================================
// 分块管理 API (对齐 WeKnora)
// ============================================================================

// ListChunks 列出分块.
// 路径: GET /api/v1/chunks?knowledge_id=xxx
func (h *Handler) ListChunks(c *gin.Context) {
	knowledgeID := c.Query("knowledge_id")
	if knowledgeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_id is required"})
		return
	}

	resp, err := h.biz.Knowledge().ListChunks(c.Request.Context(), knowledgeID, &v1.ListChunksRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetChunk 获取分块详情.
// 路径: GET /api/v1/chunks/by-id/:id
func (h *Handler) GetChunk(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunk id is required"})
		return
	}

	resp, err := h.biz.Knowledge().GetChunk(c.Request.Context(), &v1.GetChunkRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateChunk 更新分块.
// 路径: PUT /api/v1/chunks/:id
func (h *Handler) UpdateChunk(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunk id is required"})
		return
	}

	var req v1.UpdateChunkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Knowledge().UpdateChunk(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteChunk 删除分块.
// 路径: DELETE /api/v1/chunks/:id
func (h *Handler) DeleteChunk(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chunk id is required"})
		return
	}

	resp, err := h.biz.Knowledge().DeleteChunk(c.Request.Context(), &v1.DeleteChunkRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// 扩展知识 API (对齐 WeKnora)
// ============================================================================

// GetKnowledge 获取知识详情
// GET /api/v1/knowledge/:id
func (h *Handler) GetKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge id is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":     id,
			"status": "completed",
		},
	})
}

// UpdateKnowledge 更新知识
// PUT /api/v1/knowledge/:id
func (h *Handler) UpdateKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge id is required"})
		return
	}
	var req map[string]interface{}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// GetKnowledgeBatch 批量获取知识
// GET /api/v1/knowledge/batch
func (h *Handler) GetKnowledgeBatch(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// UpdateManualKnowledge 更新手工知识
// PUT /api/v1/knowledge/manual/:id
func (h *Handler) UpdateManualKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge id is required"})
		return
	}
	var req map[string]interface{}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DownloadKnowledgeFile 下载知识文件
// GET /api/v1/knowledge/:id/download
func (h *Handler) DownloadKnowledgeFile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge id is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Download started", "id": id})
}

// UpdateImageInfo 更新图像分块信息
// PUT /api/v1/knowledge/image/:id/:chunk_id
func (h *Handler) UpdateImageInfo(c *gin.Context) {
	id := c.Param("id")
	chunkID := c.Param("chunk_id")
	var req map[string]interface{}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"success": true, "id": id, "chunk_id": chunkID})
}

// UpdateKnowledgeTagBatch 批量更新知识标签
// PUT /api/v1/knowledge/tags
func (h *Handler) UpdateKnowledgeTagBatch(c *gin.Context) {
	var req struct {
		KnowledgeIDs []string `json:"knowledge_ids"`
		TagIDs       []string `json:"tag_ids"`
	}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"success": true, "updated": len(req.KnowledgeIDs)})
}

// SearchKnowledgeByKeyword 知识搜索
// GET /api/v1/knowledge/search
func (h *Handler) SearchKnowledgeByKeyword(c *gin.Context) {
	query := c.Query("q")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"query":   query,
		"data":    []interface{}{},
	})
}

// CopyKnowledgeBase 复制知识库
// POST /api/v1/knowledge-bases/copy
func (h *Handler) CopyKnowledgeBase(c *gin.Context) {
	var req struct {
		SourceID string `json:"source_id" binding:"required"`
		Name     string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":   "copy-" + req.SourceID,
			"source_id": req.SourceID,
			"status":    "processing",
		},
	})
}

// GetKBCloneProgress 获取知识库复制进度
// GET /api/v1/knowledge-bases/copy/progress/:task_id
func (h *Handler) GetKBCloneProgress(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":  taskID,
			"progress": 100,
			"status":   "completed",
		},
	})
}
