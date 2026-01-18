// Package http 提供 HTTP 处理器.
// web_search.go - 网络搜索相关 Handler（对齐 WeKnora /api/v1/web-search）
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebSearchProviderInfo Web 搜索提供商信息（API 响应格式）
type WebSearchProviderInfo struct {
	ID             string `json:"id"`                // 提供商ID
	Name           string `json:"name"`              // 提供商名称
	Free           bool   `json:"free"`              // 是否免费
	RequiresAPIKey bool   `json:"requires_api_key"`  // 是否需要API密钥
	Description    string `json:"description"`       // 描述
	APIURL         string `json:"api_url,omitempty"` // API地址（可选）
}

// GetWebSearchProviders 获取网络搜索提供商
// GET /api/v1/web-search/providers
func (h *Handler) GetWebSearchProviders(c *gin.Context) {
	if h.webSearchConfig == nil || len(h.webSearchConfig.Providers) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []WebSearchProviderInfo{},
		})
		return
	}

	// 将配置转换为 API 响应格式
	providers := make([]WebSearchProviderInfo, 0, len(h.webSearchConfig.Providers))
	for _, provider := range h.webSearchConfig.Providers {
		providers = append(providers, WebSearchProviderInfo{
			ID:             provider.ID,
			Name:           provider.Name,
			Free:           provider.Free,
			RequiresAPIKey: provider.RequiresAPIKey,
			Description:    provider.Description,
			APIURL:         provider.APIURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    providers,
	})
}
