// Package llmcontext 提供内存存储实现.
package llmcontext

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// memoryStorage 内存存储实现.
type memoryStorage struct {
	mu   sync.RWMutex
	data map[string][]*schema.Message
}

// NewMemoryStorage 创建内存存储.
func NewMemoryStorage() ContextStorage {
	return &memoryStorage{
		data: make(map[string][]*schema.Message),
	}
}

func (s *memoryStorage) Load(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if messages, ok := s.data[sessionID]; ok {
		// 返回副本，避免外部修改
		result := make([]*schema.Message, len(messages))
		copy(result, messages)
		return result, nil
	}
	return []*schema.Message{}, nil
}

func (s *memoryStorage) Save(ctx context.Context, sessionID string, messages []*schema.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 保存副本
	result := make([]*schema.Message, len(messages))
	copy(result, messages)
	s.data[sessionID] = result
	return nil
}

func (s *memoryStorage) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, sessionID)
	return nil
}
