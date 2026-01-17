// Package llmcontext provides model context window limits.
// Reference: WeKnora llmcontext/model_limits.go
package llmcontext

import (
	"sync"
)

const (
	// DefaultMaxToolTokens Tool 结果的最大 token 预算
	DefaultMaxToolTokens = 40000
	// DefaultSafetyMargin 安全边距
	DefaultSafetyMargin = 1024
)

// ModelContextLimits 定义各模型的上下文窗口限制.
var ModelContextLimits = map[string]int64{
	// OpenAI Models
	"gpt-4o":                 128000,
	"gpt-4o-mini":            128000,
	"gpt-4-turbo":            128000,
	"gpt-4-turbo-2024-04-09": 128000,
	"gpt-4-0125-preview":     128000,
	"gpt-4-1106-preview":     128000,
	"gpt-3.5-turbo":          16385,
	"gpt-3.5-turbo-0125":     16385,
	"gpt-3.5-turbo-1106":     16385,
	"gpt-4":                  8192,
	"gpt-4-32k":              32768,

	// Anthropic Claude Models
	"claude-3-5-sonnet-20241022": 200000,
	"claude-3-5-sonnet-20240620": 200000,
	"claude-3-5-sonnet":          200000,
	"claude-3-5-haiku-20241022":  200000,
	"claude-3-5-haiku":           200000,
	"claude-3-opus-20240229":     200000,
	"claude-3-opus":              200000,
	"claude-3-sonnet-20240229":   200000,
	"claude-3-sonnet":            200000,
	"claude-3-haiku-20240307":    200000,
	"claude-3-haiku":             200000,
	"claude-2.1":                 200000,
	"claude-2.0":                 200000,
	"claude-instant-1.2":         100000,

	// 阿里云 Qwen Models
	"qwen-max":   30000,
	"qwen-plus":  30000,
	"qwen-turbo": 8000,
	"qwen-long":  1000000,

	// Google Gemini Models
	"gemini-1.5-pro":   1000000,
	"gemini-1.5-flash": 1000000,
	"gemini-1.0-pro":   30000,
	"gemini-pro":       30000,

	// Meta Llama Models
	"llama-3.1-405b": 128000,
	"llama-3.1-70b":  128000,
	"llama-3.1-8b":   128000,
	"llama-3-70b":    8192,
	"llama-3-8b":     8192,

	// Mistral Models
	"mistral-large-2402":  128000,
	"mistral-large-2312":  32000,
	"mistral-medium-2312": 32000,
	"mistral-small-2402":  32000,
	"mistral-7b":          32000,
	"mixtral-8x7b":        32000,

	// DeepSeek Models
	"deepseek-chat":  128000,
	"deepseek-coder": 128000,

	// Moonshot Models (月之暗面 Kimi)
	"moonshot-v1-128k": 128000,
	"moonshot-v1-32k":  32000,
	"moonshot-v1-8k":   8192,

	// 百度文心一言
	"ernie-bot-4":     128000,
	"ernie-bot":       128000,
	"ernie-bot-turbo": 128000,

	// 腾讯混元
	"hunyuan-lite":     256000,
	"hunyuan-standard": 256000,
	"hunyuan-pro":      256000,
	"hunyuan-turbo":    256000,

	// 智谱 GLM
	"glm-4-plus":  128000,
	"glm-4-air":   128000,
	"glm-4-flash": 128000,
	"glm-3-turbo": 128000,

	// 字节豆包
	"doubao-pro-128k":  128000,
	"doubao-pro-32k":   32000,
	"doubao-lite-128k": 128000,
	"doubao-lite-32k":  32000,

	// 默认值
	"_default": 128000,
}

// mu protects overrides during concurrent access
var mu sync.RWMutex

// ModelLimitConfig 模型限制配置.
type ModelLimitConfig struct {
	SafetyMargin   int64            // 安全边距
	MaxToolTokens  int              // Tool 结果最大 token
	ModelOverrides map[string]int64 // 模型覆盖配置
}

// ModelLimitManager 管理模型上下文窗口限制.
type ModelLimitManager struct {
	staticLimits         map[string]int64
	overrides            map[string]int64
	safetyMargin         int64
	defaultMaxToolTokens int
}

// NewModelLimitManager 创建模型限制管理器.
func NewModelLimitManager(cfg *ModelLimitConfig) *ModelLimitManager {
	if cfg == nil {
		cfg = &ModelLimitConfig{
			SafetyMargin:  DefaultSafetyMargin,
			MaxToolTokens: DefaultMaxToolTokens,
		}
	}

	mgr := &ModelLimitManager{
		staticLimits:         ModelContextLimits,
		overrides:            make(map[string]int64),
		safetyMargin:         cfg.SafetyMargin,
		defaultMaxToolTokens: cfg.MaxToolTokens,
	}

	// 应用配置中的模型覆盖
	for model, limit := range cfg.ModelOverrides {
		mgr.SetOverride(model, limit)
	}

	return mgr
}

// GetModelLimit 获取指定模型的上下文窗口限制.
// 优先级: 用户覆盖 > 预定义限制 > 模糊匹配 > 默认值
func (m *ModelLimitManager) GetModelLimit(model string) int64 {
	mu.RLock()
	defer mu.RUnlock()

	// 首先检查用户自定义覆盖
	if limit, ok := m.overrides[model]; ok && limit > 0 {
		return limit
	}

	// 检查预定义限制
	if limit, ok := m.staticLimits[model]; ok && limit > 0 {
		return limit
	}

	// 尝试模糊匹配
	return m.fuzzyMatchLimit(model)
}

// fuzzyMatchLimit 对模型名称进行模糊匹配.
// 例如: "gpt-4o-20240508" 可以匹配 "gpt-4o"
func (m *ModelLimitManager) fuzzyMatchLimit(model string) int64 {
	var bestPrefix string
	var bestLimit int64

	for prefix, limit := range m.staticLimits {
		if prefix == "_default" {
			continue
		}
		if len(model) > len(prefix) && model[:len(prefix)] == prefix {
			if len(prefix) > len(bestPrefix) {
				bestPrefix = prefix
				bestLimit = limit
			}
		}
	}

	if bestPrefix != "" {
		return bestLimit
	}

	return m.staticLimits["_default"]
}

// SetOverride 设置模型限制的覆盖值.
func (m *ModelLimitManager) SetOverride(model string, limit int64) {
	mu.Lock()
	defer mu.Unlock()
	m.overrides[model] = limit
}

// RemoveOverride 移除模型限制的覆盖值.
func (m *ModelLimitManager) RemoveOverride(model string) {
	mu.Lock()
	defer mu.Unlock()
	delete(m.overrides, model)
}

// CalculateRemainingTokens 计算剩余可用的 token 数量.
func (m *ModelLimitManager) CalculateRemainingTokens(model string, usedTokens, additionalTokens int64) int64 {
	maxTokens := m.GetModelLimit(model)
	remaining := maxTokens - usedTokens - m.safetyMargin - additionalTokens
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// CalculateUtilization 计算 token 使用率.
func (m *ModelLimitManager) CalculateUtilization(model string, usedTokens int64) float64 {
	maxTokens := m.GetModelLimit(model)
	if maxTokens <= 0 {
		return 0
	}
	return float64(usedTokens) / float64(maxTokens)
}

// GetMaxToolTokens 获取 Tool 结果的最大 token 预算.
func (m *ModelLimitManager) GetMaxToolTokens() int {
	return m.defaultMaxToolTokens
}

// GetSafetyMargin 获取安全边距.
func (m *ModelLimitManager) GetSafetyMargin() int64 {
	return m.safetyMargin
}

// GetSupportedModels 返回所有支持的模型列表.
func (m *ModelLimitManager) GetSupportedModels() []string {
	models := make([]string, 0, len(m.staticLimits))
	for model := range m.staticLimits {
		if model != "_default" {
			models = append(models, model)
		}
	}
	return models
}

// 全局默认管理器
var (
	defaultManager     *ModelLimitManager
	defaultManagerOnce sync.Once
)

// GetDefaultManager 获取全局默认的模型限制管理器.
func GetDefaultManager() *ModelLimitManager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewModelLimitManager(nil)
	})
	return defaultManager
}

// SetDefaultManager 设置全局默认的模型限制管理器.
func SetDefaultManager(manager *ModelLimitManager) {
	defaultManager = manager
}

// 便捷函数

// GetModelLimitByName 获取指定模型的上下文窗口限制 (便捷函数).
func GetModelLimitByName(model string) int64 {
	return GetDefaultManager().GetModelLimit(model)
}

// CalculateRemainingTokensByModel 计算剩余可用 token (便捷函数).
func CalculateRemainingTokensByModel(model string, usedTokens, additionalTokens int64) int64 {
	return GetDefaultManager().CalculateRemainingTokens(model, usedTokens, additionalTokens)
}

// CalculateUtilizationByModel 计算 token 使用率 (便捷函数).
func CalculateUtilizationByModel(model string, usedTokens int64) float64 {
	return GetDefaultManager().CalculateUtilization(model, usedTokens)
}
