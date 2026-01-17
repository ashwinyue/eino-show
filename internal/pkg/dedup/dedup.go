// Package dedup 提供请求去重和合并功能.
// 基于 singleflight 实现，防止相同请求并发时重复调用 LLM.
package dedup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Deduplicator 请求去重器.
type Deduplicator struct {
	group  singleflight.Group
	cache  sync.Map // 短期缓存，防止短时间内重复请求
	ttl    time.Duration
	config *Config
}

// Config 去重器配置.
type Config struct {
	// TTL 缓存有效期（相同请求在此时间内直接返回缓存）
	TTL time.Duration
	// KeyFunc 自定义 key 生成函数
	KeyFunc func(ctx context.Context, query string, extra ...string) string
	// EnableCache 是否启用短期缓存
	EnableCache bool
}

// DefaultConfig 默认配置.
func DefaultConfig() *Config {
	return &Config{
		TTL:         30 * time.Second,
		KeyFunc:     DefaultKeyFunc,
		EnableCache: true,
	}
}

// DefaultKeyFunc 默认 key 生成函数（SHA256 哈希）.
func DefaultKeyFunc(ctx context.Context, query string, extra ...string) string {
	h := sha256.New()
	h.Write([]byte(query))
	for _, e := range extra {
		h.Write([]byte(e))
	}
	return hex.EncodeToString(h.Sum(nil))[:16] // 取前 16 位
}

// cacheEntry 缓存条目.
type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// New 创建去重器.
func New(cfg *Config) *Deduplicator {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = DefaultKeyFunc
	}
	if cfg.TTL == 0 {
		cfg.TTL = 30 * time.Second
	}

	d := &Deduplicator{
		ttl:    cfg.TTL,
		config: cfg,
	}

	// 启动缓存清理协程
	if cfg.EnableCache {
		go d.cleanupLoop()
	}

	return d
}

// Do 执行去重操作.
// 相同 key 的并发请求只会执行一次 fn，其他请求等待结果.
func (d *Deduplicator) Do(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	// 1. 检查短期缓存
	if d.config.EnableCache {
		if entry, ok := d.cache.Load(key); ok {
			e := entry.(*cacheEntry)
			if time.Now().Before(e.expiresAt) {
				return e.value, nil, true // shared=true 表示来自缓存
			}
			d.cache.Delete(key)
		}
	}

	// 2. 使用 singleflight 去重
	result, err, shared := d.group.Do(key, func() (interface{}, error) {
		return fn()
	})

	// 3. 缓存结果
	if err == nil && d.config.EnableCache {
		d.cache.Store(key, &cacheEntry{
			value:     result,
			expiresAt: time.Now().Add(d.ttl),
		})
	}

	return result, err, shared
}

// DoWithTimeout 带超时的去重操作.
func (d *Deduplicator) DoWithTimeout(ctx context.Context, key string, timeout time.Duration, fn func() (interface{}, error)) (interface{}, error, bool) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultCh := make(chan struct {
		val    interface{}
		err    error
		shared bool
	}, 1)

	go func() {
		val, err, shared := d.Do(ctx, key, fn)
		resultCh <- struct {
			val    interface{}
			err    error
			shared bool
		}{val, err, shared}
	}()

	select {
	case result := <-resultCh:
		return result.val, result.err, result.shared
	case <-ctx.Done():
		return nil, ctx.Err(), false
	}
}

// GenerateKey 生成去重 key.
func (d *Deduplicator) GenerateKey(ctx context.Context, query string, extra ...string) string {
	return d.config.KeyFunc(ctx, query, extra...)
}

// Forget 主动移除正在执行的请求（用于取消）.
func (d *Deduplicator) Forget(key string) {
	d.group.Forget(key)
	d.cache.Delete(key)
}

// cleanupLoop 定期清理过期缓存.
func (d *Deduplicator) cleanupLoop() {
	ticker := time.NewTicker(d.ttl)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		d.cache.Range(func(key, value interface{}) bool {
			entry := value.(*cacheEntry)
			if now.After(entry.expiresAt) {
				d.cache.Delete(key)
			}
			return true
		})
	}
}

// Stats 统计信息.
type Stats struct {
	CacheSize int
}

// GetStats 获取统计信息.
func (d *Deduplicator) GetStats() *Stats {
	size := 0
	d.cache.Range(func(_, _ interface{}) bool {
		size++
		return true
	})
	return &Stats{CacheSize: size}
}

// ==================== QA 专用去重器 ====================

// QADeduplicator QA 请求去重器.
type QADeduplicator struct {
	*Deduplicator
}

// NewQADeduplicator 创建 QA 去重器.
func NewQADeduplicator() *QADeduplicator {
	cfg := &Config{
		TTL:         10 * time.Second, // QA 缓存时间较短
		EnableCache: true,
		KeyFunc: func(ctx context.Context, query string, extra ...string) string {
			// QA key = sessionID + query hash
			h := sha256.New()
			h.Write([]byte(query))
			for _, e := range extra {
				h.Write([]byte(e))
			}
			return hex.EncodeToString(h.Sum(nil))[:16]
		},
	}
	return &QADeduplicator{
		Deduplicator: New(cfg),
	}
}

// DeduplicateQA 去重 QA 请求.
// sessionID: 会话 ID
// query: 用户问题
// fn: 实际执行函数
func (d *QADeduplicator) DeduplicateQA(ctx context.Context, sessionID, query string, fn func() (interface{}, error)) (interface{}, error, bool) {
	key := d.GenerateKey(ctx, query, sessionID)
	return d.Do(ctx, key, fn)
}
