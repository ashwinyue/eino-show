// Package gin 提供 Gin 框架的限流中间件.
package gin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/ratelimit"
)

// RateLimitConfig 限流配置.
type RateLimitConfig struct {
	// Rate 每秒允许的请求数
	Rate int
	// Burst 突发容量（仅用于令牌桶）
	Burst int
	// KeyFunc 获取限流 key 的函数（用于用户级限流）
	KeyFunc func(c *gin.Context) string
	// ExcludePaths 排除的路径（不限流）
	ExcludePaths []string
	// ErrorHandler 自定义错误处理
	ErrorHandler func(c *gin.Context)
}

// DefaultRateLimitConfig 默认配置.
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Rate:  100, // 每秒 100 请求
		Burst: 10,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
		ExcludePaths: []string{"/healthz", "/metrics"},
		ErrorHandler: defaultErrorHandler,
	}
}

func defaultErrorHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"code":    http.StatusTooManyRequests,
		"message": "rate limit exceeded, please try again later",
	})
}

// ==================== 全局限流（基于 uber/ratelimit 漏桶算法）====================

// GlobalRateLimiter 全局限流器.
type GlobalRateLimiter struct {
	limiter ratelimit.Limiter
	config  *RateLimitConfig
}

// NewGlobalRateLimiter 创建全局限流器.
func NewGlobalRateLimiter(rate int) *GlobalRateLimiter {
	return &GlobalRateLimiter{
		limiter: ratelimit.New(rate),
		config:  DefaultRateLimitConfig(),
	}
}

// Middleware 返回 Gin 中间件.
func (l *GlobalRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查排除路径
		for _, path := range l.config.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 漏桶限流（阻塞式，保证平滑）
		l.limiter.Take()
		c.Next()
	}
}

// ==================== 用户级限流（基于 Redis 滑动窗口）====================

// RedisRateLimiter 基于 Redis 的分布式限流器.
type RedisRateLimiter struct {
	client redis.UniversalClient
	config *RateLimitConfig
	prefix string
}

// NewRedisRateLimiter 创建 Redis 限流器.
func NewRedisRateLimiter(client redis.UniversalClient, config *RateLimitConfig) *RedisRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	return &RedisRateLimiter{
		client: client,
		config: config,
		prefix: "ratelimit:",
	}
}

// Middleware 返回 Gin 中间件.
func (l *RedisRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查排除路径
		for _, path := range l.config.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 获取限流 key
		key := l.prefix + l.config.KeyFunc(c)

		// 检查是否超限
		allowed, err := l.isAllowed(c.Request.Context(), key)
		if err != nil {
			// Redis 错误时放行（降级策略）
			c.Next()
			return
		}

		if !allowed {
			l.config.ErrorHandler(c)
			return
		}

		c.Next()
	}
}

// isAllowed 检查是否允许请求（滑动窗口算法）.
func (l *RedisRateLimiter) isAllowed(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixNano()
	windowSize := time.Second.Nanoseconds()

	pipe := l.client.Pipeline()

	// 移除窗口外的请求记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-windowSize))

	// 获取当前窗口内的请求数
	countCmd := pipe.ZCard(ctx, key)

	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// 设置过期时间（窗口大小 + 1秒）
	pipe.Expire(ctx, key, 2*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := countCmd.Val()
	return count < int64(l.config.Rate), nil
}

// ==================== 接口级限流（不同接口不同限流策略）====================

// EndpointRateLimiter 接口级限流器.
type EndpointRateLimiter struct {
	limiters map[string]ratelimit.Limiter
	mu       sync.RWMutex
	defaults *RateLimitConfig
}

// EndpointLimit 接口限流配置.
type EndpointLimit struct {
	Path   string // 接口路径（支持通配符 *）
	Method string // HTTP 方法（空表示所有方法）
	Rate   int    // 每秒请求数
}

// NewEndpointRateLimiter 创建接口级限流器.
func NewEndpointRateLimiter(limits []EndpointLimit, defaultRate int) *EndpointRateLimiter {
	l := &EndpointRateLimiter{
		limiters: make(map[string]ratelimit.Limiter),
		defaults: DefaultRateLimitConfig(),
	}
	l.defaults.Rate = defaultRate

	// 初始化各接口的限流器
	for _, limit := range limits {
		key := l.makeKey(limit.Method, limit.Path)
		l.limiters[key] = ratelimit.New(limit.Rate)
	}

	return l
}

func (l *EndpointRateLimiter) makeKey(method, path string) string {
	if method == "" {
		return "*:" + path
	}
	return method + ":" + path
}

// Middleware 返回 Gin 中间件.
func (l *EndpointRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查排除路径
		for _, path := range l.defaults.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 查找对应的限流器
		limiter := l.getLimiter(c.Request.Method, c.Request.URL.Path)
		if limiter != nil {
			limiter.Take()
		}

		c.Next()
	}
}

func (l *EndpointRateLimiter) getLimiter(method, path string) ratelimit.Limiter {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 精确匹配
	key := l.makeKey(method, path)
	if limiter, ok := l.limiters[key]; ok {
		return limiter
	}

	// 方法通配符匹配
	key = l.makeKey("", path)
	if limiter, ok := l.limiters[key]; ok {
		return limiter
	}

	return nil
}

// ==================== 组合限流（多层限流）====================

// CombinedRateLimiter 组合限流器（全局 + 用户级）.
type CombinedRateLimiter struct {
	global   *GlobalRateLimiter
	perUser  *RedisRateLimiter
	endpoint *EndpointRateLimiter
}

// CombinedConfig 组合限流配置.
type CombinedConfig struct {
	// GlobalRate 全局每秒请求数（0 表示不限制）
	GlobalRate int
	// UserRate 每用户每秒请求数（0 表示不限制）
	UserRate int
	// RedisClient Redis 客户端（用户级限流需要）
	RedisClient redis.UniversalClient
	// EndpointLimits 接口级限流配置
	EndpointLimits []EndpointLimit
	// ExcludePaths 排除的路径
	ExcludePaths []string
}

// NewCombinedRateLimiter 创建组合限流器.
func NewCombinedRateLimiter(cfg *CombinedConfig) *CombinedRateLimiter {
	c := &CombinedRateLimiter{}

	// 全局限流
	if cfg.GlobalRate > 0 {
		c.global = NewGlobalRateLimiter(cfg.GlobalRate)
		if len(cfg.ExcludePaths) > 0 {
			c.global.config.ExcludePaths = cfg.ExcludePaths
		}
	}

	// 用户级限流
	if cfg.UserRate > 0 && cfg.RedisClient != nil {
		userConfig := DefaultRateLimitConfig()
		userConfig.Rate = cfg.UserRate
		if len(cfg.ExcludePaths) > 0 {
			userConfig.ExcludePaths = cfg.ExcludePaths
		}
		c.perUser = NewRedisRateLimiter(cfg.RedisClient, userConfig)
	}

	// 接口级限流
	if len(cfg.EndpointLimits) > 0 {
		c.endpoint = NewEndpointRateLimiter(cfg.EndpointLimits, cfg.GlobalRate)
		if len(cfg.ExcludePaths) > 0 {
			c.endpoint.defaults.ExcludePaths = cfg.ExcludePaths
		}
	}

	return c
}

// Middleware 返回 Gin 中间件.
func (c *CombinedRateLimiter) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 全局限流
		if c.global != nil {
			c.global.limiter.Take()
		}

		// 2. 接口级限流
		if c.endpoint != nil {
			limiter := c.endpoint.getLimiter(ctx.Request.Method, ctx.Request.URL.Path)
			if limiter != nil {
				limiter.Take()
			}
		}

		// 3. 用户级限流（基于 Redis）
		if c.perUser != nil {
			key := c.perUser.prefix + c.perUser.config.KeyFunc(ctx)
			allowed, err := c.perUser.isAllowed(ctx.Request.Context(), key)
			if err == nil && !allowed {
				c.perUser.config.ErrorHandler(ctx)
				return
			}
		}

		ctx.Next()
	}
}
