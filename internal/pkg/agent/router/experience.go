// Package router provides experience storage and retrieval.
// Reference: AssistantAgent Experience Module
package router

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

// ExperienceType 经验类型.
type ExperienceType string

const (
	ExperienceTypeCode       ExperienceType = "code"        // 代码生成经验
	ExperienceTypeReact      ExperienceType = "react"       // ReAct 决策经验
	ExperienceTypeKnowledge  ExperienceType = "knowledge"   // 常识经验
	ExperienceTypeFastIntent ExperienceType = "fast_intent" // 快速意图经验
)

// Experience 经验记录.
type Experience struct {
	ID          string         // 经验 ID
	Type        ExperienceType // 经验类型
	Query       string         // 原始查询
	Response    string         // 成功响应
	ToolCalls   []string       // 工具调用记录
	Code        string         // 生成的代码 (Code 类型)
	Embedding   []float32      // 查询的向量表示
	Score       float64        // 匹配分数
	UsageCount  int            // 使用次数
	SuccessRate float64        // 成功率
	CreatedAt   time.Time      // 创建时间
	UpdatedAt   time.Time      // 更新时间
	Metadata    map[string]any // 元数据

	// FastIntent 快速意图配置 (仅 fast_intent 类型)
	FastIntentConfig *FastIntentConfig
}

// FastIntentConfig 快速意图配置.
type FastIntentConfig struct {
	// Patterns 匹配模式
	Patterns []string

	// MatchType 匹配类型: prefix, contains, regex
	MatchType string

	// DirectResponse 直接响应
	DirectResponse string

	// ToolCall 直接工具调用
	ToolCall *schema.ToolCall
}

// ExperienceStore 经验存储接口.
type ExperienceStore interface {
	// Save 保存经验
	Save(ctx context.Context, exp *Experience) error

	// Get 获取经验
	Get(ctx context.Context, id string) (*Experience, error)

	// Search 搜索相似经验
	Search(ctx context.Context, query string, topK int) ([]*Experience, error)

	// SearchByEmbedding 通过向量搜索
	SearchByEmbedding(ctx context.Context, embedding []float32, topK int) ([]*Experience, error)

	// Delete 删除经验
	Delete(ctx context.Context, id string) error

	// Update 更新经验
	Update(ctx context.Context, exp *Experience) error

	// List 列出所有经验
	List(ctx context.Context, expType ExperienceType, limit int) ([]*Experience, error)
}

// ExperienceManagerConfig 经验管理器配置.
type ExperienceManagerConfig struct {
	// Store 经验存储
	Store ExperienceStore

	// Embedder 向量化模型 (可选，用于语义搜索)
	Embedder embedding.Embedder

	// SimilarityThreshold 相似度阈值 (默认 0.8)
	SimilarityThreshold float64

	// MaxExperiences 最大返回经验数 (默认 5)
	MaxExperiences int
}

// ExperienceManager 经验管理器.
type ExperienceManager struct {
	cfg *ExperienceManagerConfig
}

// NewExperienceManager 创建经验管理器.
func NewExperienceManager(cfg *ExperienceManagerConfig) (*ExperienceManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.Store == nil {
		cfg.Store = NewMemoryExperienceStore()
	}
	if cfg.SimilarityThreshold <= 0 {
		cfg.SimilarityThreshold = 0.8
	}
	if cfg.MaxExperiences <= 0 {
		cfg.MaxExperiences = 5
	}

	return &ExperienceManager{cfg: cfg}, nil
}

// Learn 从执行结果中学习经验.
func (m *ExperienceManager) Learn(ctx context.Context, exp *Experience) error {
	// 生成向量 (如果有 Embedder)
	if m.cfg.Embedder != nil && len(exp.Embedding) == 0 {
		vectors, err := m.cfg.Embedder.EmbedStrings(ctx, []string{exp.Query})
		if err == nil && len(vectors) > 0 {
			exp.Embedding = float64ToFloat32(vectors[0])
		}
	}

	exp.CreatedAt = time.Now()
	exp.UpdatedAt = time.Now()
	exp.UsageCount = 0
	exp.SuccessRate = 1.0

	return m.cfg.Store.Save(ctx, exp)
}

// Recall 检索相关经验.
func (m *ExperienceManager) Recall(ctx context.Context, query string) ([]*Experience, error) {
	// 1. 尝试向量搜索
	if m.cfg.Embedder != nil {
		vectors, err := m.cfg.Embedder.EmbedStrings(ctx, []string{query})
		if err == nil && len(vectors) > 0 {
			experiences, err := m.cfg.Store.SearchByEmbedding(ctx, float64ToFloat32(vectors[0]), m.cfg.MaxExperiences)
			if err == nil && len(experiences) > 0 {
				// 过滤低相似度
				filtered := make([]*Experience, 0)
				for _, exp := range experiences {
					if exp.Score >= m.cfg.SimilarityThreshold {
						filtered = append(filtered, exp)
					}
				}
				if len(filtered) > 0 {
					return filtered, nil
				}
			}
		}
	}

	// 2. 回退到文本搜索
	return m.cfg.Store.Search(ctx, query, m.cfg.MaxExperiences)
}

// RecordUsage 记录经验使用.
func (m *ExperienceManager) RecordUsage(ctx context.Context, expID string, success bool) error {
	exp, err := m.cfg.Store.Get(ctx, expID)
	if err != nil {
		return err
	}

	exp.UsageCount++
	if success {
		exp.SuccessRate = (exp.SuccessRate*float64(exp.UsageCount-1) + 1.0) / float64(exp.UsageCount)
	} else {
		exp.SuccessRate = (exp.SuccessRate * float64(exp.UsageCount-1)) / float64(exp.UsageCount)
	}
	exp.UpdatedAt = time.Now()

	return m.cfg.Store.Update(ctx, exp)
}

// CheckFastIntent 检查快速意图.
func (m *ExperienceManager) CheckFastIntent(ctx context.Context, query string) (*Experience, bool) {
	experiences, err := m.cfg.Store.List(ctx, ExperienceTypeFastIntent, 100)
	if err != nil {
		return nil, false
	}

	for _, exp := range experiences {
		if exp.FastIntentConfig == nil {
			continue
		}

		cfg := exp.FastIntentConfig
		for _, pattern := range cfg.Patterns {
			matched := false
			switch cfg.MatchType {
			case "prefix":
				matched = strings.HasPrefix(query, pattern)
			case "contains":
				matched = strings.Contains(query, pattern)
			default:
				matched = strings.HasPrefix(query, pattern)
			}

			if matched {
				return exp, true
			}
		}
	}

	return nil, false
}

// MemoryExperienceStore 内存经验存储.
type MemoryExperienceStore struct {
	mu          sync.RWMutex
	experiences map[string]*Experience
	counter     int
}

// NewMemoryExperienceStore 创建内存经验存储.
func NewMemoryExperienceStore() *MemoryExperienceStore {
	return &MemoryExperienceStore{
		experiences: make(map[string]*Experience),
	}
}

func (s *MemoryExperienceStore) Save(ctx context.Context, exp *Experience) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if exp.ID == "" {
		s.counter++
		exp.ID = fmt.Sprintf("exp_%d", s.counter)
	}
	s.experiences[exp.ID] = exp
	return nil
}

func (s *MemoryExperienceStore) Get(ctx context.Context, id string) (*Experience, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, ok := s.experiences[id]
	if !ok {
		return nil, fmt.Errorf("experience not found: %s", id)
	}
	return exp, nil
}

func (s *MemoryExperienceStore) Search(ctx context.Context, query string, topK int) ([]*Experience, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Experience
	queryLower := strings.ToLower(query)

	for _, exp := range s.experiences {
		if strings.Contains(strings.ToLower(exp.Query), queryLower) ||
			strings.Contains(strings.ToLower(exp.Response), queryLower) {
			results = append(results, exp)
			if len(results) >= topK {
				break
			}
		}
	}

	return results, nil
}

func (s *MemoryExperienceStore) SearchByEmbedding(ctx context.Context, emb []float32, topK int) ([]*Experience, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type scored struct {
		exp   *Experience
		score float64
	}

	var scoredExps []scored
	for _, exp := range s.experiences {
		if len(exp.Embedding) == 0 {
			continue
		}
		score := cosineSimilarity(emb, exp.Embedding)
		scoredExps = append(scoredExps, scored{exp: exp, score: score})
	}

	// 排序
	for i := 0; i < len(scoredExps)-1; i++ {
		for j := i + 1; j < len(scoredExps); j++ {
			if scoredExps[j].score > scoredExps[i].score {
				scoredExps[i], scoredExps[j] = scoredExps[j], scoredExps[i]
			}
		}
	}

	// 返回 topK
	results := make([]*Experience, 0, topK)
	for i := 0; i < len(scoredExps) && i < topK; i++ {
		exp := scoredExps[i].exp
		exp.Score = scoredExps[i].score
		results = append(results, exp)
	}

	return results, nil
}

func (s *MemoryExperienceStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.experiences, id)
	return nil
}

func (s *MemoryExperienceStore) Update(ctx context.Context, exp *Experience) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.experiences[exp.ID] = exp
	return nil
}

func (s *MemoryExperienceStore) List(ctx context.Context, expType ExperienceType, limit int) ([]*Experience, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Experience
	for _, exp := range s.experiences {
		if expType == "" || exp.Type == expType {
			results = append(results, exp)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// cosineSimilarity 计算余弦相似度.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// float64ToFloat32 converts []float64 to []float32.
func float64ToFloat32(input []float64) []float32 {
	result := make([]float32, len(input))
	for i, v := range input {
		result[i] = float32(v)
	}
	return result
}
