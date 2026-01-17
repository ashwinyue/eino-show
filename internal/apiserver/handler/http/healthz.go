package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/eino-show/internal/pkg/log"
)

// HealthzResponse 健康检查响应
type HealthzResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Healthz 服务健康检查.
func (h *Handler) Healthz(c *gin.Context) {
	log.W(c.Request.Context()).Infow("Healthz handler is called", "method", "Healthz", "status", "healthy")
	c.JSON(http.StatusOK, HealthzResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.DateTime),
	})
}
