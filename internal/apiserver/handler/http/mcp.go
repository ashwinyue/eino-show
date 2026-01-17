// Package http 提供 HTTP 处理器.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/ashwinyue/eino-show/pkg/core"
)

// ListMCPServices 获取 MCP 服务列表.
// @Summary      获取 MCP 服务列表
// @Description  获取当前租户的所有 MCP 服务
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Success      200  {object}  ListMCPServicesResponse  "MCP 服务列表"
// @Security     Bearer
// @Router       /api/v1/mcp-services [get]
func (h *Handler) ListMCPServices(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.MCP().List, nil)
}

// GetMCPService 获取 MCP 服务详情.
// @Summary      获取 MCP 服务详情
// @Description  根据 ID 获取 MCP 服务详细信息
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "MCP 服务 ID"
// @Success      200  {object}  GetMCPServiceResponse  "MCP 服务信息"
// @Failure      404  {object}  core.ErrorMessage  "MCP 服务不存在"
// @Security     Bearer
// @Router       /api/v1/mcp-services/{id} [get]
func (h *Handler) GetMCPService(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.MCP().Get, nil)
}

// CreateMCPService 创建 MCP 服务.
// @Summary      创建 MCP 服务
// @Description  创建新的 MCP 服务
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        request  body      CreateMCPServiceRequest  true  "创建 MCP 服务请求"
// @Success      200     {object}  CreateMCPServiceResponse  "MCP 服务信息"
// @Failure      400     {object}  core.ErrorMessage  "请求参数错误"
// @Security     Bearer
// @Router       /api/v1/mcp-services [post]
func (h *Handler) CreateMCPService(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.MCP().Create, nil)
}

// UpdateMCPService 更新 MCP 服务.
// @Summary      更新 MCP 服务
// @Description  更新 MCP 服务信息
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "MCP 服务 ID"
// @Param        request  body      UpdateMCPServiceRequest  true  "更新 MCP 服务请求"
// @Success      200  {object}  UpdateMCPServiceResponse  "更新后的 MCP 服务信息"
// @Failure      400  {object}  core.ErrorMessage  "请求参数错误"
// @Failure      404  {object}  core.ErrorMessage  "MCP 服务不存在"
// @Security     Bearer
// @Router       /api/v1/mcp-services/{id} [put]
func (h *Handler) UpdateMCPService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mcp service id is required"})
		return
	}

	var req v1.UpdateMCPServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.MCP().Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteMCPService 删除 MCP 服务.
// @Summary      删除 MCP 服务
// @Description  删除指定 MCP 服务
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "MCP 服务 ID"
// @Success      200  {object}  DeleteMCPServiceResponse  "删除成功"
// @Failure      404  {object}  core.ErrorMessage  "MCP 服务不存在"
// @Security     Bearer
// @Router       /api/v1/mcp-services/{id} [delete]
func (h *Handler) DeleteMCPService(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.MCP().Delete, nil)
}

// TestMCPService 测试 MCP 服务连接.
// @Summary      测试 MCP 服务连接
// @Description  测试指定 MCP 服务的连接状态
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "MCP 服务 ID"
// @Success      200  {object}  TestMCPServiceResponse  "测试结果"
// @Failure      404  {object}  core.ErrorMessage  "MCP 服务不存在"
// @Security     Bearer
// @Router       /api/v1/mcp-services/{id}/test [post]
func (h *Handler) TestMCPService(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.MCP().Test, nil)
}

// GetMCPServiceTools 获取 MCP 服务工具列表.
// @Summary      获取 MCP 服务工具列表
// @Description  获取指定 MCP 服务提供的工具列表
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "MCP 服务 ID"
// @Success      200  {object}  GetMCPServiceToolsResponse  "工具列表"
// @Failure      404  {object}  core.ErrorMessage  "MCP 服务不存在"
// @Security     Bearer
// @Router       /api/v1/mcp-services/{id}/tools [get]
func (h *Handler) GetMCPServiceTools(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.MCP().GetTools, nil)
}

// GetMCPServiceResources 获取 MCP 服务资源列表
// GET /api/v1/mcp-services/:id/resources
func (h *Handler) GetMCPServiceResources(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// 类型别名，用于 Swagger 文档生成
type (
	ListMCPServicesResponse    = GetMCPServiceResponse
	GetMCPServiceResponse      = any
	CreateMCPServiceRequest    = any
	CreateMCPServiceResponse   = any
	UpdateMCPServiceRequest    = any
	UpdateMCPServiceResponse   = any
	DeleteMCPServiceResponse   = any
	TestMCPServiceResponse     = any
	GetMCPServiceToolsResponse = any
)
