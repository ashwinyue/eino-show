// Package llmcontext provides Redis-based context storage for production environments.
// Reference: WeKnora llmcontext/redis_storage.go
package llmcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/cloudwego/eino/schema"
)

const (
	// DefaultKeyPrefix Redis key 前缀
	DefaultKeyPrefix = "llmcontext:"
	// DefaultTTL 默认过期时间 (24 小时)
	DefaultTTL = 24 * time.Hour
)

// RedisStorageConfig Redis 存储配置.
type RedisStorageConfig struct {
	// Client Redis 客户端
	Client *redis.Client

	// KeyPrefix key 前缀
	KeyPrefix string

	// TTL 过期时间
	TTL time.Duration
}

// redisStorage Redis 存储实现.
type redisStorage struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

// NewRedisStorage 创建 Redis 存储.
func NewRedisStorage(cfg *RedisStorageConfig) (ContextStorage, error) {
	if cfg == nil || cfg.Client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	keyPrefix := cfg.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = DefaultKeyPrefix
	}

	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	return &redisStorage{
		client:    cfg.Client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}, nil
}

// getKey 生成 Redis key.
func (s *redisStorage) getKey(sessionID string) string {
	return s.keyPrefix + sessionID
}

// Load 从 Redis 加载消息历史.
func (s *redisStorage) Load(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	key := s.getKey(sessionID)

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return []*schema.Message{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load from redis: %w", err)
	}

	var messages []*schema.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
	}

	return messages, nil
}

// Save 保存消息历史到 Redis.
func (s *redisStorage) Save(ctx context.Context, sessionID string, messages []*schema.Message) error {
	key := s.getKey(sessionID)

	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save to redis: %w", err)
	}

	return nil
}

// Delete 删除会话的消息历史.
func (s *redisStorage) Delete(ctx context.Context, sessionID string) error {
	key := s.getKey(sessionID)

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from redis: %w", err)
	}

	return nil
}

// Ensure redisStorage implements ContextStorage
var _ ContextStorage = (*redisStorage)(nil)

// RedisSummaryStore Redis-based summary storage for progressive compression.
type RedisSummaryStore struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

// RedisSummaryStoreConfig Redis 摘要存储配置.
type RedisSummaryStoreConfig struct {
	Client    *redis.Client
	KeyPrefix string
	TTL       time.Duration
}

// NewRedisSummaryStore 创建 Redis 摘要存储.
func NewRedisSummaryStore(cfg *RedisSummaryStoreConfig) (*RedisSummaryStore, error) {
	if cfg == nil || cfg.Client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	keyPrefix := cfg.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = "summary:"
	}

	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	return &RedisSummaryStore{
		client:    cfg.Client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}, nil
}

// summaryData 摘要数据结构.
type summaryData struct {
	Summaries  []string `json:"summaries"`
	TokenCount int      `json:"token_count"`
	UpdatedAt  int64    `json:"updated_at"`
}

// getKey 生成 Redis key.
func (s *RedisSummaryStore) getKey(sessionID string) string {
	return s.keyPrefix + sessionID
}

// SaveSummary 保存摘要到会话.
func (s *RedisSummaryStore) SaveSummary(ctx context.Context, sessionID string, summary string, tokenCount int) error {
	key := s.getKey(sessionID)

	// 获取现有摘要
	existing, err := s.loadData(ctx, key)
	if err != nil && err != redis.Nil {
		return err
	}

	if existing == nil {
		existing = &summaryData{Summaries: []string{}}
	}

	existing.Summaries = append(existing.Summaries, summary)
	existing.TokenCount += tokenCount
	existing.UpdatedAt = time.Now().Unix()

	return s.saveData(ctx, key, existing)
}

// GetSummaries 获取会话的所有摘要.
func (s *RedisSummaryStore) GetSummaries(ctx context.Context, sessionID string) ([]string, error) {
	key := s.getKey(sessionID)

	data, err := s.loadData(ctx, key)
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	return data.Summaries, nil
}

// ReplaceSummaries 替换所有摘要为一个合并后的摘要.
func (s *RedisSummaryStore) ReplaceSummaries(ctx context.Context, sessionID string, mergedSummary string, tokenCount int) error {
	key := s.getKey(sessionID)

	data := &summaryData{
		Summaries:  []string{mergedSummary},
		TokenCount: tokenCount,
		UpdatedAt:  time.Now().Unix(),
	}

	return s.saveData(ctx, key, data)
}

// ClearSummaries 清除会话的所有摘要.
func (s *RedisSummaryStore) ClearSummaries(ctx context.Context, sessionID string) error {
	key := s.getKey(sessionID)
	return s.client.Del(ctx, key).Err()
}

// loadData 从 Redis 加载摘要数据.
func (s *RedisSummaryStore) loadData(ctx context.Context, key string) (*summaryData, error) {
	bytes, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var data summaryData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary data: %w", err)
	}

	return &data, nil
}

// saveData 保存摘要数据到 Redis.
func (s *RedisSummaryStore) saveData(ctx context.Context, key string, data *summaryData) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal summary data: %w", err)
	}

	return s.client.Set(ctx, key, bytes, s.ttl).Err()
}
