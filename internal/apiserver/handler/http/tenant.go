// Package http 提供 HTTP 处理器.
// tenant.go - 租户相关 Handler（对齐 WeKnora /api/v1/tenants）
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ListAllTenants 获取所有租户
// GET /api/v1/tenants/all
func (h *Handler) ListAllTenants(c *gin.Context) {
	resp, err := h.biz.Tenant().List(c.Request.Context(), &v1.ListTenantsRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SearchTenants 搜索租户
// GET /api/v1/tenants/search
func (h *Handler) SearchTenants(c *gin.Context) {
	var req v1.SearchTenantsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Tenant().Search(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateTenant 创建租户
// POST /api/v1/tenants
func (h *Handler) CreateTenant(c *gin.Context) {
	var req v1.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Tenant().Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetTenant 获取租户详情
// GET /api/v1/tenants/:id
func (h *Handler) GetTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	resp, err := h.biz.Tenant().Get(c.Request.Context(), &v1.GetTenantRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateTenant 更新租户
// PUT /api/v1/tenants/:id
func (h *Handler) UpdateTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	var req v1.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id

	resp, err := h.biz.Tenant().Update(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteTenant 删除租户
// DELETE /api/v1/tenants/:id
func (h *Handler) DeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	resp, err := h.biz.Tenant().Delete(c.Request.Context(), &v1.DeleteTenantRequest{ID: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListTenants 获取租户列表
// GET /api/v1/tenants
func (h *Handler) ListTenants(c *gin.Context) {
	var req v1.ListTenantsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Tenant().List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetTenantKV 获取租户 KV 配置
// GET /api/v1/tenants/kv/:key
func (h *Handler) GetTenantKV(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	tenantID := contextx.TenantID(c.Request.Context())
	resp, err := h.biz.Tenant().GetKV(c.Request.Context(), tenantID, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateTenantKV 更新租户 KV 配置
// PUT /api/v1/tenants/kv/:key
func (h *Handler) UpdateTenantKV(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	var req v1.UpdateTenantKVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Key = key

	tenantID := contextx.TenantID(c.Request.Context())
	resp, err := h.biz.Tenant().UpdateKV(c.Request.Context(), tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
