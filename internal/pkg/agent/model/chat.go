// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package model 提供 ChatModel 工厂，支持多种 LLM 提供商.
package model

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// Config 定义 ChatModel 配置.
type Config struct {
	// Provider 模型提供商: openai, ark, azure, etc.
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
	if cfg.Timeout > 0 {
		config.Timeout = cfg.timeoutDuration()
	}

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
	if cfg.Timeout > 0 {
		t := cfg.timeoutDuration()
		config.Timeout = &t
	}

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

// timeoutDuration 将秒转换为 time.Duration.
func (c *Config) timeoutDuration() time.Duration {
	return time.Duration(c.Timeout) * time.Second
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
