// Package http 提供 HTTP 处理器.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/ashwinyue/eino-show/pkg/core"
)

// CreateModel 创建模型.
func (h *Handler) CreateModel(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.Model().Create)
}

// GetModel 获取模型详情.
func (h *Handler) GetModel(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.Model().Get)
}

// ListModels 获取模型列表.
func (h *Handler) ListModels(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.Model().List)
}

// UpdateModel 更新模型.
func (h *Handler) UpdateModel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model id is required"})
		return
	}

	var req v1.UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Model().Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteModel 删除模型.
func (h *Handler) DeleteModel(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.Model().Delete)
}

// SetDefaultModel 设置默认模型.
func (h *Handler) SetDefaultModel(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.Model().SetDefault)
}

// ListModelProviders 获取模型厂商列表
// GET /api/v1/models/providers
func (h *Handler) ListModelProviders(c *gin.Context) {
	modelType := c.Query("model_type")
	resp, err := h.biz.Model().ListProviders(c.Request.Context(), modelType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
