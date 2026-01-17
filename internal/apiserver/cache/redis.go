// Package cache 提供缓存存储实现.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ICache 缓存接口.
type ICache interface {
	// Set 设置缓存（带 TTL）
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get 获取缓存
	Get(ctx context.Context, key string, dest interface{}) error

	// Del 删除缓存
	Del(ctx context.Context, keys ...string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Expire 设置过期时间
	Expire(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// IsEnabled 检查缓存是否可用
	IsEnabled() bool

	// Client 返回底层 Redis 客户端（用于高级操作）
	Client() redis.UniversalClient
}

// redisCache Redis 缓存实现.
type redisCache struct {
	client redis.UniversalClient
}

// NewRedisCache 创建 Redis 缓存实例.
func NewRedisCache(client redis.UniversalClient) ICache {
	return &redisCache{client: client}
}

// 确保 redisCache 实现了 ICache 接口.
var _ ICache = (*redisCache)(nil)

// IsEnabled 检查 Redis 客户端是否可用.
func (c *redisCache) IsEnabled() bool {
	return c.client != nil
}

// Client 返回底层 Redis 客户端.
func (c *redisCache) Client() redis.UniversalClient {
	return c.client
}

// Set 设置缓存（带 TTL）.
// 如果 Redis 未配置，静默忽略（缓存是可选的优化）.
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.IsEnabled() {
		return nil // Redis 未配置，静默忽略
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Get 获取缓存.
// 如果 Redis 未配置，返回 redis.Nil 错误（未找到）.
func (c *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	if !c.IsEnabled() {
		return redis.Nil // Redis 未配置，视为缓存未命中
	}
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Del 删除缓存.
// 如果 Redis 未配置，静默忽略.
func (c *redisCache) Del(ctx context.Context, keys ...string) error {
	if !c.IsEnabled() {
		return nil // Redis 未配置，静默忽略
	}
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在.
// 如果 Redis 未配置，返回 false（不存在）.
func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	if !c.IsEnabled() {
		return false, nil // Redis 未配置，视为键不存在
	}
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 设置过期时间.
// 如果 Redis 未配置，返回 false.
func (c *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if !c.IsEnabled() {
		return false, nil // Redis 未配置
	}
	return c.client.Expire(ctx, key, ttl).Result()
}
