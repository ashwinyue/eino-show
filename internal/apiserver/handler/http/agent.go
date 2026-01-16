// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package http 提供 HTTP 处理器.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
)

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
	core.HandleQueryRequest(c, h.biz.AgentV1().List, h.val.ValidateListAgents)
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
	core.HandleUriRequest(c, h.biz.AgentV1().Get, h.val.ValidateGetAgent)
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
	core.HandleJSONRequest(c, h.biz.AgentV1().Create, h.val.ValidateCreateAgent)
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
	core.HandleJSONRequest(c, h.biz.AgentV1().Update, h.val.ValidateUpdateAgent)
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
	core.HandleUriRequest(c, h.biz.AgentV1().Delete, h.val.ValidateDeleteAgent)
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
	builtinAgents := h.biz.AgentV1().ListBuiltin(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"agents": builtinAgents,
	})
}
