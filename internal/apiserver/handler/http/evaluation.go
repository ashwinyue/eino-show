// Package http 提供 HTTP 处理器.
// evaluation.go - 评估相关 Handler（对齐 WeKnora /api/v1/evaluation）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Evaluation 执行评估
// POST /api/v1/evaluation/
func (h *Handler) Evaluation(c *gin.Context) {
	var req struct {
		SessionID string   `json:"session_id"`
		Questions []string `json:"questions"`
		Metrics   []string `json:"metrics"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"evaluation_id": "eval-" + req.SessionID,
			"status":        "processing",
		},
	})
}

// GetEvaluationResult 获取评估结果
// GET /api/v1/evaluation/
func (h *Handler) GetEvaluationResult(c *gin.Context) {
	evalID := c.Query("evaluation_id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"evaluation_id": evalID,
			"status":        "completed",
			"results":       []interface{}{},
		},
	})
}
