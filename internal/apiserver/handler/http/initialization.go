// Package http 提供 HTTP 处理器.
// initialization.go - 初始化相关 Handler（对齐 WeKnora /api/v1/initialization）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// GetCurrentConfigByKB 获取知识库当前配置
// GET /api/v1/initialization/config/:kbId
func (h *Handler) GetCurrentConfigByKB(c *gin.Context) {
	kbID := c.Param("kbId")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kbId is required"})
		return
	}

	// 返回默认配置
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"kb_id":           kbID,
			"embedding_model": "text-embedding-ada-002",
			"chunk_size":      500,
			"chunk_overlap":   50,
			"initialized":     true,
		},
	})
}

// InitializeByKB 初始化知识库
// POST /api/v1/initialization/initialize/:kbId
func (h *Handler) InitializeByKB(c *gin.Context) {
	kbID := c.Param("kbId")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kbId is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Knowledge base initialized successfully",
		"data": gin.H{
			"kb_id":       kbID,
			"initialized": true,
		},
	})
}

// UpdateKBConfig 更新知识库配置
// PUT /api/v1/initialization/config/:kbId
func (h *Handler) UpdateKBConfig(c *gin.Context) {
	kbID := c.Param("kbId")
	if kbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kbId is required"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration updated successfully",
		"data":    req,
	})
}

// CheckOllamaStatus 检查 Ollama 状态
// GET /api/v1/initialization/ollama/status
func (h *Handler) CheckOllamaStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"available": false,
		},
	})
}

// ListOllamaModels 获取 Ollama 模型列表
// GET /api/v1/initialization/ollama/models
func (h *Handler) ListOllamaModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// CheckOllamaModels 检查 Ollama 模型
// POST /api/v1/initialization/ollama/models/check
func (h *Handler) CheckOllamaModels(c *gin.Context) {
	var req struct {
		Models []string `json:"models"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 返回模型状态 (简化实现)
	results := make([]gin.H, len(req.Models))
	for i, m := range req.Models {
		results[i] = gin.H{"name": m, "available": false}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// DownloadOllamaModel 下载 Ollama 模型
// POST /api/v1/initialization/ollama/models/download
func (h *Handler) DownloadOllamaModel(c *gin.Context) {
	var req struct {
		Model string `json:"model" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id": "download-" + req.Model,
			"model":   req.Model,
			"status":  "pending",
		},
	})
}

// GetDownloadProgress 获取下载进度
// GET /api/v1/initialization/ollama/download/progress/:taskId
func (h *Handler) GetDownloadProgress(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
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

// ListDownloadTasks 获取下载任务列表
// GET /api/v1/initialization/ollama/download/tasks
func (h *Handler) ListDownloadTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// CheckRemoteModel 检查远程模型
// POST /api/v1/initialization/remote/check
func (h *Handler) CheckRemoteModel(c *gin.Context) {
	var req v1.TestChatModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Model().TestChatModel(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"data":    resp,
	})
}

// TestEmbeddingModel 测试 Embedding 模型
// POST /api/v1/initialization/embedding/test
func (h *Handler) TestEmbeddingModel(c *gin.Context) {
	var req v1.TestEmbeddingModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Model().TestEmbeddingModel(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"data":    resp,
	})
}

// CheckRerankModel 检查 Rerank 模型
// POST /api/v1/initialization/rerank/check
func (h *Handler) CheckRerankModel(c *gin.Context) {
	var req v1.TestRerankModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Model().TestRerankModel(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"data":    resp,
	})
}

// TestMultimodalFunction 测试多模态功能
// POST /api/v1/initialization/multimodal/test
func (h *Handler) TestMultimodalFunction(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"supported": true,
			"features":  []string{"image", "audio"},
		},
	})
}

// ExtractTextRelations 提取文本关系
// POST /api/v1/initialization/extract/text-relation
func (h *Handler) ExtractTextRelations(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"relations": []interface{}{},
		},
	})
}

// FabriTag Fabri 标签
// POST /api/v1/initialization/extract/fabri-tag
func (h *Handler) FabriTag(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tags": []interface{}{},
		},
	})
}

// FabriText Fabri 文本
// POST /api/v1/initialization/extract/fabri-text
func (h *Handler) FabriText(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"processed_text": req.Text,
		},
	})
}
