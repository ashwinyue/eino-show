// Package stream provides Redis-based stream management for SSE events.
// Reference: WeKnora internal/stream/redis_manager.go
package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamEvent represents a single event in the stream.
type StreamEvent struct {
	ID        string                 `json:"id"`             // Unique event ID
	Type      string                 `json:"type"`           // Event type (thinking, tool_call, tool_result, references, complete, etc.)
	Content   string                 `json:"content"`        // Event content (chunk for streaming events)
	Done      bool                   `json:"done"`           // Whether this event is done
	Timestamp time.Time              `json:"timestamp"`      // When this event occurred
	Data      map[string]interface{} `json:"data,omitempty"` // Additional event data
}

// StreamManager stream manager interface - minimal append-only design.
type StreamManager interface {
	// AppendEvent appends a single event to the stream
	AppendEvent(ctx context.Context, sessionID, messageID string, event StreamEvent) error

	// GetEvents gets events starting from offset
	// Returns: events slice, next offset for subsequent reads, error
	GetEvents(ctx context.Context, sessionID, messageID string, fromOffset int) ([]StreamEvent, int, error)

	// Close closes the connection
	Close() error
}

// RedisStreamManager implements StreamManager using Redis Lists for append-only event streaming.
type RedisStreamManager struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

// RedisStreamConfig Redis Stream Manager 配置.
type RedisStreamConfig struct {
	Client *redis.Client
	Prefix string
	TTL    time.Duration
}

// NewRedisStreamManager creates a new Redis-based stream manager.
func NewRedisStreamManager(cfg *RedisStreamConfig) (*RedisStreamManager, error) {
	if cfg == nil || cfg.Client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	// Verify connection
	_, err := cfg.Client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	ttl := cfg.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour // Default TTL: 24 hours
	}

	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "stream:events" // Default prefix
	}

	return &RedisStreamManager{
		client: cfg.Client,
		ttl:    ttl,
		prefix: prefix,
	}, nil
}

// NewRedisStreamManagerWithAddr creates a new Redis-based stream manager with address.
func NewRedisStreamManagerWithAddr(addr, password string, db int, prefix string, ttl time.Duration) (*RedisStreamManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return NewRedisStreamManager(&RedisStreamConfig{
		Client: client,
		Prefix: prefix,
		TTL:    ttl,
	})
}

// buildKey builds the Redis key for event list.
func (r *RedisStreamManager) buildKey(sessionID, messageID string) string {
	return fmt.Sprintf("%s:%s:%s", r.prefix, sessionID, messageID)
}

// AppendEvent appends a single event to the stream using Redis RPush.
func (r *RedisStreamManager) AppendEvent(
	ctx context.Context,
	sessionID, messageID string,
	event StreamEvent,
) error {
	key := r.buildKey(sessionID, messageID)

	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Append to Redis list with RPush (O(1) operation)
	if err := r.client.RPush(ctx, key, eventJSON).Err(); err != nil {
		return fmt.Errorf("failed to append event to Redis: %w", err)
	}

	// Set/refresh TTL on the key
	if err := r.client.Expire(ctx, key, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	return nil
}

// GetEvents gets events starting from offset using Redis LRange.
// Returns: events slice, next offset, error
func (r *RedisStreamManager) GetEvents(
	ctx context.Context,
	sessionID, messageID string,
	fromOffset int,
) ([]StreamEvent, int, error) {
	key := r.buildKey(sessionID, messageID)

	// Get all events from offset to end using LRange
	results, err := r.client.LRange(ctx, key, int64(fromOffset), -1).Result()
	if err != nil {
		if err == redis.Nil {
			return []StreamEvent{}, fromOffset, nil
		}
		return nil, fromOffset, fmt.Errorf("failed to get events from Redis: %w", err)
	}

	// No new events
	if len(results) == 0 {
		return []StreamEvent{}, fromOffset, nil
	}

	// Unmarshal events
	events := make([]StreamEvent, 0, len(results))
	for _, result := range results {
		var event StreamEvent
		if err := json.Unmarshal([]byte(result), &event); err != nil {
			continue
		}
		events = append(events, event)
	}

	// Calculate next offset
	nextOffset := fromOffset + len(results)

	return events, nextOffset, nil
}

// DeleteStream deletes a stream.
func (r *RedisStreamManager) DeleteStream(ctx context.Context, sessionID, messageID string) error {
	key := r.buildKey(sessionID, messageID)
	return r.client.Del(ctx, key).Err()
}

// GetStreamLength gets the length of a stream.
func (r *RedisStreamManager) GetStreamLength(ctx context.Context, sessionID, messageID string) (int64, error) {
	key := r.buildKey(sessionID, messageID)
	return r.client.LLen(ctx, key).Result()
}

// Close closes the Redis connection.
func (r *RedisStreamManager) Close() error {
	return r.client.Close()
}

// Ensure RedisStreamManager implements StreamManager interface
var _ StreamManager = (*RedisStreamManager)(nil)

// MemoryStreamManager implements StreamManager using in-memory storage (for testing/single node).
type MemoryStreamManager struct {
	streams map[string][]StreamEvent
}

// NewMemoryStreamManager creates a new in-memory stream manager.
func NewMemoryStreamManager() *MemoryStreamManager {
	return &MemoryStreamManager{
		streams: make(map[string][]StreamEvent),
	}
}

// buildKey builds the key for the stream.
func (m *MemoryStreamManager) buildKey(sessionID, messageID string) string {
	return fmt.Sprintf("%s:%s", sessionID, messageID)
}

// AppendEvent appends a single event to the stream.
func (m *MemoryStreamManager) AppendEvent(
	ctx context.Context,
	sessionID, messageID string,
	event StreamEvent,
) error {
	key := m.buildKey(sessionID, messageID)
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	m.streams[key] = append(m.streams[key], event)
	return nil
}

// GetEvents gets events starting from offset.
func (m *MemoryStreamManager) GetEvents(
	ctx context.Context,
	sessionID, messageID string,
	fromOffset int,
) ([]StreamEvent, int, error) {
	key := m.buildKey(sessionID, messageID)
	events, ok := m.streams[key]
	if !ok || fromOffset >= len(events) {
		return []StreamEvent{}, fromOffset, nil
	}
	return events[fromOffset:], len(events), nil
}

// Close closes the stream manager.
func (m *MemoryStreamManager) Close() error {
	m.streams = make(map[string][]StreamEvent)
	return nil
}

var _ StreamManager = (*MemoryStreamManager)(nil)
