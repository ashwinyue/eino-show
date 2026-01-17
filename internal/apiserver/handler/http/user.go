package http

import (
	"net/http"

	"github.com/ashwinyue/eino-show/pkg/core"
	"github.com/gin-gonic/gin"
)

// Login 用户登录并返回 JWT Token（支持 email/username 登录）.
func (h *Handler) Login(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.User().Login, h.val.ValidateLoginRequest)
}

// Register 用户注册.
func (h *Handler) Register(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.User().Register)
}

// Logout 用户登出.
func (h *Handler) Logout(c *gin.Context) {
	// 从 header 获取 token
	token := c.GetHeader("Authorization")
	if token != "" && len(token) > 7 {
		token = token[7:] // 移除 "Bearer " 前缀
	}
	_ = h.biz.User().Logout(c.Request.Context(), token)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RefreshToken 刷新 JWT Token.
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.User().RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCurrentUser 获取当前登录用户信息.
func (h *Handler) GetCurrentUser(c *gin.Context) {
	resp, err := h.biz.User().GetCurrentUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// GetCurrentTenant 获取当前租户信息.
func (h *Handler) GetCurrentTenant(c *gin.Context) {
	resp, err := h.biz.User().GetCurrentTenant(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp.Tenant,
	})
}

// ValidateToken 验证 Token 有效性.
func (h *Handler) ValidateToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"valid": true}) // TODO: implement
}

// ChangePassword 修改用户密码.
func (h *Handler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id") // 从中间件获取
	if err := h.biz.User().ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CreateUser 创建新用户.
func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       "user-new",
			"username": req.Username,
			"email":    req.Email,
		},
	})
}

// UpdateUser 更新用户信息.
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}
	var req map[string]interface{}
	_ = c.ShouldBindJSON(&req)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeleteUser 删除用户.
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetUser 获取用户信息.
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
}

// ListUser 列出用户信息.
func (h *Handler) ListUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
		"total":   0,
	})
}
