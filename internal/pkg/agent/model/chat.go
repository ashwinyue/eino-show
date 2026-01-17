// Package model 提供 ChatModel 工厂，支持多种 LLM 提供商.
package model

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// Config 定义 ChatModel 配置.
type Config struct {
	// Provider 模型提供商: openai, ark, dashscope, azure, etc.
	Provider string `json:"provider"`

	// Model 模型名称
	Model string `json:"model"`

	// APIKey API 密钥
	APIKey string `json:"api_key"`

	// BaseURL API 基础 URL（可选，用于自定义端点）
	BaseURL string `json:"base_url"`

	// Temperature 温度参数 (0.0-2.0)
	Temperature float64 `json:"temperature"`

	// TopP Top-P 采样参数
	TopP float64 `json:"top_p"`

	// MaxTokens 最大生成 token 数
	MaxTokens int `json:"max_tokens"`

	// Timeout 请求超时时间（秒）
	Timeout int `json:"timeout"`
}

// DashScope 专用常量
const (
	DashScopeProvider     = "dashscope"
	DashScopeBaseURL      = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	DashScopeDefaultModel = "qwen-turbo"
)

// NewChatModel 根据配置创建 ChatModel.
func NewChatModel(ctx context.Context, cfg *Config) (model.ChatModel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	switch cfg.Provider {
	case "openai":
		return newOpenAIChatModel(ctx, cfg)
	case "ark":
		return newArkChatModel(ctx, cfg)
	case DashScopeProvider:
		return newDashScopeChatModel(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// newOpenAIChatModel 创建 OpenAI ChatModel.
func newOpenAIChatModel(ctx context.Context, cfg *Config) (*openai.ChatModel, error) {
	config := &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	}

	// 设置可选参数（仅在非零值时设置）
	if cfg.Temperature > 0 {
		t := float32(cfg.Temperature)
		config.Temperature = &t
	}
	if cfg.TopP > 0 {
		p := float32(cfg.TopP)
		config.TopP = &p
	}
	if cfg.MaxTokens > 0 {
		config.MaxTokens = &cfg.MaxTokens
	}
	// Timeout 暂不设置，使用默认值

	// 从环境变量获取默认值
	if config.APIKey == "" {
		config.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if config.Model == "" {
		config.Model = os.Getenv("OPENAI_MODEL")
		if config.Model == "" {
			config.Model = "gpt-4o-mini"
		}
	}
	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("OPENAI_BASE_URL")
	}

	return openai.NewChatModel(ctx, config)
}

// newArkChatModel 创建火山引擎 Ark ChatModel.
func newArkChatModel(ctx context.Context, cfg *Config) (*ark.ChatModel, error) {
	config := &ark.ChatModelConfig{
		Model:  cfg.Model,
		APIKey: cfg.APIKey,
	}

	// 设置可选参数（仅在非零值时设置）
	if cfg.Temperature > 0 {
		t := float32(cfg.Temperature)
		config.Temperature = &t
	}
	if cfg.TopP > 0 {
		p := float32(cfg.TopP)
		config.TopP = &p
	}
	if cfg.MaxTokens > 0 {
		config.MaxTokens = &cfg.MaxTokens
	}
	// Timeout 暂不设置，使用默认值

	// 从环境变量获取默认值
	if config.APIKey == "" {
		config.APIKey = os.Getenv("ARK_API_KEY")
	}
	if config.Model == "" {
		config.Model = os.Getenv("ARK_CHAT_MODEL")
		if config.Model == "" {
			config.Model = "ep-20250514090203-hmwcg"
		}
	}

	return ark.NewChatModel(ctx, config)
}

// newDashScopeChatModel 创建 DashScope ChatModel (阿里云).
// DashScope 使用 OpenAI 兼容 API，但需要特定的 BaseURL.
func newDashScopeChatModel(ctx context.Context, cfg *Config) (*openai.ChatModel, error) {
	config := &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	}

	// DashScope 默认使用兼容模式的 API 端点
	if config.BaseURL == "" {
		config.BaseURL = DashScopeBaseURL
	}

	// 设置默认模型名称
	if config.Model == "" {
		config.Model = DashScopeDefaultModel
	}

	// 设置可选参数（仅在非零值时设置）
	if cfg.Temperature > 0 {
		t := float32(cfg.Temperature)
		config.Temperature = &t
	}
	if cfg.TopP > 0 {
		p := float32(cfg.TopP)
		config.TopP = &p
	}
	if cfg.MaxTokens > 0 {
		config.MaxTokens = &cfg.MaxTokens
	}
	// Timeout 暂不设置，使用默认值

	// 从环境变量获取 APIKey
	if config.APIKey == "" {
		config.APIKey = os.Getenv("DASHSCOPE_API_KEY")
	}

	return openai.NewChatModel(ctx, config)
}

// DefaultConfig 返回默认配置（从环境变量读取）.
func DefaultConfig() *Config {
	cfg := &Config{
		Provider: os.Getenv("LLM_PROVIDER"),
		Model:    os.Getenv("LLM_MODEL"),
		APIKey:   os.Getenv("LLM_API_KEY"),
		BaseURL:  os.Getenv("LLM_BASE_URL"),
	}

	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}

	return cfg
}

// NewDefaultChatModel 使用默认配置创建 ChatModel.
func NewDefaultChatModel(ctx context.Context) (model.ChatModel, error) {
	return NewChatModel(ctx, DefaultConfig())
}

// ModelParameters 数据库存储的模型参数.
type ModelParameters struct {
	APIKey      string            `json:"api_key"`
	BaseURL     string            `json:"base_url"`
	ExtraConfig map[string]string `json:"extra_config"`
}

// NewChatModelFromDB 从数据库模型配置创建 ChatModel.
func NewChatModelFromDB(ctx context.Context, modelName, modelSource string, parametersJSON string) (model.ChatModel, error) {
	// 解析参数
	var params ModelParameters
	if parametersJSON != "" && parametersJSON != "{}" {
		if err := json.Unmarshal([]byte(parametersJSON), &params); err != nil {
			return nil, fmt.Errorf("failed to parse model parameters: %w", err)
		}
	}

	// 根据来源选择 provider
	var provider string
	var baseURL string

	switch modelSource {
	case "aliyun", "dashscope", "DashScope", "Aliyun", "MODEL_SOURCE_ZHIPU", "zhipu":
		provider = DashScopeProvider
		baseURL = DashScopeBaseURL
	case "openai", "OpenAI":
		provider = "openai"
	case "ark", "Ark":
		provider = "ark"
	case "deepseek", "DeepSeek", "MODEL_SOURCE_DEEPSEEK":
		provider = "openai" // DeepSeek 使用 OpenAI 兼容接口
	default:
		return nil, fmt.Errorf("unsupported model source: %s", modelSource)
	}

	// 使用数据库配置覆盖环境变量
	cfg := &Config{
		Provider: provider,
		Model:    modelName,
		APIKey:   params.APIKey,
		BaseURL:  baseURL,
	}

	// 如果数据库中有自定义 BaseURL，使用它
	if params.BaseURL != "" {
		cfg.BaseURL = params.BaseURL
	}

	return NewChatModel(ctx, cfg)
}
