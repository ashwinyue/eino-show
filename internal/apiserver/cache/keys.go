// Package cache 提供缓存键定义和辅助函数.
package cache

import (
	"context"
	"fmt"
	"time"
)

// 缓存键前缀
const (
	PrefixAgent     = "agent"     // Agent 配置缓存
	PrefixSession   = "session"   // 会话缓存
	PrefixEmbedding = "embedding" // 向量缓存
	PrefixModel     = "model"     // 模型缓存
)

// 默认 TTL
const (
	DefaultTTL   = 30 * time.Minute
	ShortTTL     = 5 * time.Minute
	LongTTL      = 24 * time.Hour
	EmbeddingTTL = 7 * 24 * time.Hour // 向量缓存 7 天
)

// AgentKey 生成 Agent 缓存键.
func AgentKey(agentID string) string {
	return fmt.Sprintf("%s:%s", PrefixAgent, agentID)
}

// SessionKey 生成会话缓存键.
func SessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", PrefixSession, sessionID)
}

// EmbeddingKey 生成向量缓存键.
func EmbeddingKey(key string) string {
	return fmt.Sprintf("%s:%s", PrefixEmbedding, key)
}

// ModelKey 生成模型缓存键.
func ModelKey(modelID string) string {
	return fmt.Sprintf("%s:%s", PrefixModel, modelID)
}

// CacheHelper 缓存辅助结构体，提供便捷的缓存操作.
type CacheHelper struct {
	cache ICache
}

// NewCacheHelper 创建缓存辅助实例.
func NewCacheHelper(cache ICache) *CacheHelper {
	return &CacheHelper{cache: cache}
}

// SetAgent 缓存 Agent 配置.
func (h *CacheHelper) SetAgent(ctx context.Context, agentID string, agent interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = DefaultTTL
	}
	return h.cache.Set(ctx, AgentKey(agentID), agent, ttl)
}

// GetAgent 获取缓存的 Agent 配置.
func (h *CacheHelper) GetAgent(ctx context.Context, agentID string, dest interface{}) error {
	return h.cache.Get(ctx, AgentKey(agentID), dest)
}

// SetSession 缓存会话数据.
func (h *CacheHelper) SetSession(ctx context.Context, sessionID string, data interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = DefaultTTL
	}
	return h.cache.Set(ctx, SessionKey(sessionID), data, ttl)
}

// GetSession 获取缓存的会话数据.
func (h *CacheHelper) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	return h.cache.Get(ctx, SessionKey(sessionID), dest)
}

// SetEmbedding 缓存向量嵌入结果.
func (h *CacheHelper) SetEmbedding(ctx context.Context, key string, vector []float32) error {
	return h.cache.Set(ctx, EmbeddingKey(key), vector, EmbeddingTTL)
}

// GetEmbedding 获取缓存的向量嵌入结果.
func (h *CacheHelper) GetEmbedding(ctx context.Context, key string, dest *[]float32) error {
	return h.cache.Get(ctx, EmbeddingKey(key), dest)
}

// DelAgent 删除 Agent 缓存.
func (h *CacheHelper) DelAgent(ctx context.Context, agentID string) error {
	return h.cache.Del(ctx, AgentKey(agentID))
}

// DelSession 删除会话缓存.
func (h *CacheHelper) DelSession(ctx context.Context, sessionID string) error {
	return h.cache.Del(ctx, SessionKey(sessionID))
}
