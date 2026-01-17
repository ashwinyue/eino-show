// Package stream provides web search state management.
// Reference: WeKnora internal/application/service/web_search_state.go
package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// WebSearchState Web 搜索临时状态.
type WebSearchState struct {
	KBID         string          `json:"kbID"`         // 临时知识库 ID
	KnowledgeIDs []string        `json:"knowledgeIDs"` // 知识条目 ID 列表
	SeenURLs     map[string]bool `json:"seenURLs"`     // 已访问的 URL
	CreatedAt    time.Time       `json:"createdAt"`    // 创建时间
	UpdatedAt    time.Time       `json:"updatedAt"`    // 更新时间
}

// WebSearchStateService Web 搜索状态服务接口.
type WebSearchStateService interface {
	// GetState 获取会话的 Web 搜索状态
	GetState(ctx context.Context, sessionID string) (*WebSearchState, error)

	// SaveState 保存 Web 搜索状态
	SaveState(ctx context.Context, sessionID string, state *WebSearchState) error

	// DeleteState 删除 Web 搜索状态
	DeleteState(ctx context.Context, sessionID string) error

	// AddSeenURL 添加已访问 URL
	AddSeenURL(ctx context.Context, sessionID, url string) error

	// HasSeenURL 检查 URL 是否已访问
	HasSeenURL(ctx context.Context, sessionID, url string) (bool, error)

	// AddKnowledgeID 添加知识条目 ID
	AddKnowledgeID(ctx context.Context, sessionID, knowledgeID string) error
}

// RedisWebSearchStateService Redis 实现的 Web 搜索状态服务.
type RedisWebSearchStateService struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// WebSearchStateConfig 配置.
type WebSearchStateConfig struct {
	Client *redis.Client
	Prefix string
	TTL    time.Duration
}

// NewRedisWebSearchStateService 创建 Redis Web 搜索状态服务.
func NewRedisWebSearchStateService(cfg *WebSearchStateConfig) (*RedisWebSearchStateService, error) {
	if cfg == nil || cfg.Client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "websearch"
	}

	ttl := cfg.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return &RedisWebSearchStateService{
		client: cfg.Client,
		prefix: prefix,
		ttl:    ttl,
	}, nil
}

// buildKey 构建 Redis key.
func (s *RedisWebSearchStateService) buildKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", s.prefix, sessionID)
}

// GetState 获取会话的 Web 搜索状态.
func (s *RedisWebSearchStateService) GetState(ctx context.Context, sessionID string) (*WebSearchState, error) {
	key := s.buildKey(sessionID)

	raw, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 不存在
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	var state WebSearchState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// SaveState 保存 Web 搜索状态.
func (s *RedisWebSearchStateService) SaveState(ctx context.Context, sessionID string, state *WebSearchState) error {
	key := s.buildKey(sessionID)

	state.UpdatedAt = time.Now()
	if state.CreatedAt.IsZero() {
		state.CreatedAt = state.UpdatedAt
	}
	if state.SeenURLs == nil {
		state.SeenURLs = make(map[string]bool)
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return s.client.Set(ctx, key, data, s.ttl).Err()
}

// DeleteState 删除 Web 搜索状态.
func (s *RedisWebSearchStateService) DeleteState(ctx context.Context, sessionID string) error {
	key := s.buildKey(sessionID)
	return s.client.Del(ctx, key).Err()
}

// AddSeenURL 添加已访问 URL.
func (s *RedisWebSearchStateService) AddSeenURL(ctx context.Context, sessionID, url string) error {
	state, err := s.GetState(ctx, sessionID)
	if err != nil {
		return err
	}

	if state == nil {
		state = &WebSearchState{
			SeenURLs: make(map[string]bool),
		}
	}

	state.SeenURLs[url] = true
	return s.SaveState(ctx, sessionID, state)
}

// HasSeenURL 检查 URL 是否已访问.
func (s *RedisWebSearchStateService) HasSeenURL(ctx context.Context, sessionID, url string) (bool, error) {
	state, err := s.GetState(ctx, sessionID)
	if err != nil {
		return false, err
	}

	if state == nil || state.SeenURLs == nil {
		return false, nil
	}

	return state.SeenURLs[url], nil
}

// AddKnowledgeID 添加知识条目 ID.
func (s *RedisWebSearchStateService) AddKnowledgeID(ctx context.Context, sessionID, knowledgeID string) error {
	state, err := s.GetState(ctx, sessionID)
	if err != nil {
		return err
	}

	if state == nil {
		state = &WebSearchState{
			SeenURLs: make(map[string]bool),
		}
	}

	// 检查是否已存在
	for _, id := range state.KnowledgeIDs {
		if id == knowledgeID {
			return nil
		}
	}

	state.KnowledgeIDs = append(state.KnowledgeIDs, knowledgeID)
	return s.SaveState(ctx, sessionID, state)
}

// Ensure implementation
var _ WebSearchStateService = (*RedisWebSearchStateService)(nil)

// MemoryWebSearchStateService 内存实现 (用于测试).
type MemoryWebSearchStateService struct {
	states map[string]*WebSearchState
}

// NewMemoryWebSearchStateService 创建内存 Web 搜索状态服务.
func NewMemoryWebSearchStateService() *MemoryWebSearchStateService {
	return &MemoryWebSearchStateService{
		states: make(map[string]*WebSearchState),
	}
}

// GetState 获取状态.
func (s *MemoryWebSearchStateService) GetState(ctx context.Context, sessionID string) (*WebSearchState, error) {
	state, ok := s.states[sessionID]
	if !ok {
		return nil, nil
	}
	return state, nil
}

// SaveState 保存状态.
func (s *MemoryWebSearchStateService) SaveState(ctx context.Context, sessionID string, state *WebSearchState) error {
	state.UpdatedAt = time.Now()
	if state.CreatedAt.IsZero() {
		state.CreatedAt = state.UpdatedAt
	}
	s.states[sessionID] = state
	return nil
}

// DeleteState 删除状态.
func (s *MemoryWebSearchStateService) DeleteState(ctx context.Context, sessionID string) error {
	delete(s.states, sessionID)
	return nil
}

// AddSeenURL 添加已访问 URL.
func (s *MemoryWebSearchStateService) AddSeenURL(ctx context.Context, sessionID, url string) error {
	state, _ := s.GetState(ctx, sessionID)
	if state == nil {
		state = &WebSearchState{SeenURLs: make(map[string]bool)}
	}
	state.SeenURLs[url] = true
	return s.SaveState(ctx, sessionID, state)
}

// HasSeenURL 检查 URL 是否已访问.
func (s *MemoryWebSearchStateService) HasSeenURL(ctx context.Context, sessionID, url string) (bool, error) {
	state, _ := s.GetState(ctx, sessionID)
	if state == nil || state.SeenURLs == nil {
		return false, nil
	}
	return state.SeenURLs[url], nil
}

// AddKnowledgeID 添加知识条目 ID.
func (s *MemoryWebSearchStateService) AddKnowledgeID(ctx context.Context, sessionID, knowledgeID string) error {
	state, _ := s.GetState(ctx, sessionID)
	if state == nil {
		state = &WebSearchState{SeenURLs: make(map[string]bool)}
	}
	state.KnowledgeIDs = append(state.KnowledgeIDs, knowledgeID)
	return s.SaveState(ctx, sessionID, state)
}

var _ WebSearchStateService = (*MemoryWebSearchStateService)(nil)
