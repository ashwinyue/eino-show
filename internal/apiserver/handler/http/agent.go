// Package http 提供 HTTP 处理器.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ===== Agent 请求/响应类型 =====

// CreateAgentRequest 创建 Agent 请求
type CreateAgentRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Avatar      string                 `json:"avatar"`
	Config      map[string]interface{} `json:"config"`
}

// UpdateAgentRequest 更新 Agent 请求
type UpdateAgentRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Avatar      *string                `json:"avatar"`
	Config      map[string]interface{} `json:"config"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// DeleteResponse 删除响应
type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ListAgents 获取 Agent 列表（包括内置和自定义）.
// @Summary      获取 Agent 列表
// @Description  获取当前租户的所有 Agent（包括内置和自定义）
// @Tags         agents
// @Accept       json
// @Produce      json
// @Success      200  {object}  ListAgentsResponse  "Agent 列表"
// @Security     Bearer
// @Router       /api/v1/custom-agents [get]
func (h *Handler) ListAgents(c *gin.Context) {
	resp, err := h.biz.Agent().List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetAgent 获取 Agent 详情.
// @Summary      获取 Agent 详情
// @Description  根据 ID 获取 Agent 详细信息
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Agent ID"
// @Success      200  {object}  GetAgentResponse  "Agent 信息"
// @Failure      404  {object}  core.ErrorMessage  "Agent 不存在"
// @Security     Bearer
// @Router       /api/v1/custom-agents/{id} [get]
func (h *Handler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}

	resp, err := h.biz.Agent().Get(c.Request.Context(), &v1.GetAgentRequest{Id: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateAgent 创建自定义 Agent.
// @Summary      创建 Agent
// @Description  创建新的自定义 Agent
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        request  body      CreateAgentRequest  true  "创建 Agent 请求"
// @Success      200     {object}  CreateAgentResponse  "Agent 信息"
// @Failure      400     {object}  core.ErrorMessage  "请求参数错误"
// @Security     Bearer
// @Router       /api/v1/custom-agents [post]
func (h *Handler) CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Agent().Create(c.Request.Context(), &v1.CreateAgentRequest{
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		Config:      req.Config,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateAgent 更新 Agent.
// @Summary      更新 Agent
// @Description  更新 Agent 信息
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Agent ID"
// @Param        request  body      UpdateAgentRequest  true  "更新 Agent 请求"
// @Success      200  {object}  UpdateAgentResponse  "更新后的 Agent 信息"
// @Failure      400  {object}  core.ErrorMessage  "请求参数错误"
// @Failure      404  {object}  core.ErrorMessage  "Agent 不存在"
// @Security     Bearer
// @Router       /api/v1/custom-agents/{id} [put]
func (h *Handler) UpdateAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Agent().Update(c.Request.Context(), id, &v1.UpdateAgentRequest{
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		Config:      req.Config,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteAgent 删除 Agent.
// @Summary      删除 Agent
// @Description  删除指定 Agent
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Agent ID"
// @Success      200  {object}  core.DeleteResponse  "删除成功"
// @Failure      404  {object}  core.ErrorMessage  "Agent 不存在"
// @Security     Bearer
// @Router       /api/v1/custom-agents/{id} [delete]
func (h *Handler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}

	if _, err := h.biz.Agent().Delete(c.Request.Context(), &v1.DeleteAgentRequest{Id: id}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Agent deleted successfully"})
}

// ListBuiltinAgents 获取内置 Agent 列表.
// @Summary      获取内置 Agent 列表
// @Description  获取系统内置的 Agent 列表
// @Tags         agents
// @Accept       json
// @Produce      json
// @Success      200  {object}  ListBuiltinAgentsResponse  "内置 Agent 列表"
// @Security     Bearer
// @Router       /api/v1/agents/builtin [get]
func (h *Handler) ListBuiltinAgents(c *gin.Context) {
	builtinAgents := h.biz.Agent().ListBuiltin(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    builtinAgents,
	})
}

// ListAllAgents 获取所有 Agent（内置+自定义组合，对齐 WeKnora 格式）.
// @Summary      获取所有 Agent
// @Description  获取所有 Agent，包括内置 Agent 和自定义 Agent
// @Tags         agents
// @Accept       json
// @Produce      json
// @Success      200  {object}  "Agent 列表"
// @Security     Bearer
// @Router       /api/v1/agents [get]
func (h *Handler) ListAllAgents(c *gin.Context) {
	// 获取内置 Agent（已包含数据库自定义配置）
	builtinAgents := h.biz.Agent().ListBuiltin(c.Request.Context())

	// 获取自定义 Agent（只取非内置的）
	customAgentsResp, _ := h.biz.Agent().List(c.Request.Context(), nil)

	// 组合所有 Agent
	resultAgents := []v1.AgentResponse{}

	// 1. 添加所有内置 Agent（包含默认配置）
	for _, agent := range builtinAgents {
		// 获取内置 Agent 的详细信息（包含配置）
		agentResp, err := h.biz.Agent().Get(c.Request.Context(), &v1.GetAgentRequest{Id: agent.Id})
		if err == nil && agentResp != nil && agentResp.Data != nil {
			resultAgents = append(resultAgents, *agentResp.Data)
		} else {
			// 回退：只返回基本信息
			resultAgents = append(resultAgents, v1.AgentResponse{
				ID:          agent.Id,
				Name:        agent.Name,
				Description: agent.Description,
				Avatar:      agent.Avatar,
				IsBuiltin:   true,
			})
		}
	}

	// 2. 添加非内置的自定义 Agent
	if customAgentsResp != nil {
		for _, agent := range customAgentsResp.Data {
			if !agent.IsBuiltin {
				resultAgents = append(resultAgents, v1.AgentResponse{
					ID:          agent.ID,
					Name:        agent.Name,
					Description: agent.Description,
					Avatar:      agent.Avatar,
					Config:      agent.Config,
					TenantID:    agent.TenantID,
					IsBuiltin:   false,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resultAgents,
	})
}

// ============================================================================
// 扩展 Agent API (对齐 WeKnora)
// ============================================================================

// GetPlaceholders 获取 Agent 占位符定义
// GET /api/v1/agents/placeholders
func (h *Handler) GetPlaceholders(c *gin.Context) {
	resp := h.biz.Agent().GetPlaceholders(c.Request.Context())
	c.JSON(http.StatusOK, resp)
}

// CopyAgent 复制 Agent
// POST /api/v1/agents/:id/copy
func (h *Handler) CopyAgent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}

	resp, err := h.biz.Agent().Copy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
