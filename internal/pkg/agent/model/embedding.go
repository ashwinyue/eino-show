// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package model 提供 EmbeddingModel 工厂，支持多种向量模型提供商.
package model

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

// EmbeddingConfig 定义 Embedding 配置.
type EmbeddingConfig struct {
	// Provider 模型提供商: openai, azure, etc.
	Provider string `json:"provider"`

	// Model 模型名称
	Model string `json:"model"`

	// APIKey API 密钥
	APIKey string `json:"api_key"`

	// BaseURL API 基础 URL（可选）
	BaseURL string `json:"base_url"`

	// Dimensions 向量维度
	Dimensions int `json:"dimensions"`

	// Timeout 请求超时时间（秒）
	Timeout int `json:"timeout"`
}

// NewEmbedder 根据配置创建 Embedder.
func NewEmbedder(ctx context.Context, cfg *EmbeddingConfig) (embedding.Embedder, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	switch cfg.Provider {
	case "openai":
		return newOpenAIEmbedder(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.Provider)
	}
}

// newOpenAIEmbedder 创建 OpenAI Embedder.
func newOpenAIEmbedder(ctx context.Context, cfg *EmbeddingConfig) (*openai.Embedder, error) {
	config := &openai.EmbeddingConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	}

	// 设置可选参数（仅在非零值时设置）
	if cfg.Dimensions > 0 {
		config.Dimensions = &cfg.Dimensions
	}
	if cfg.Timeout > 0 {
		config.Timeout = time.Duration(cfg.Timeout) * time.Second
	}

	// 从环境变量获取默认值
	if config.APIKey == "" {
		config.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if config.Model == "" {
		config.Model = os.Getenv("EMBEDDING_MODEL")
		if config.Model == "" {
			config.Model = "text-embedding-3-small"
		}
	}
	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("EMBEDDING_BASE_URL")
	}

	return openai.NewEmbedder(ctx, config)
}

// DefaultEmbeddingConfig 返回默认 Embedding 配置.
func DefaultEmbeddingConfig() *EmbeddingConfig {
	cfg := &EmbeddingConfig{
		Provider: os.Getenv("EMBEDDING_PROVIDER"),
		Model:    os.Getenv("EMBEDDING_MODEL"),
		APIKey:   os.Getenv("EMBEDDING_API_KEY"),
		BaseURL:  os.Getenv("EMBEDDING_BASE_URL"),
	}

	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}

	return cfg
}

// NewDefaultEmbedder 使用默认配置创建 Embedder.
func NewDefaultEmbedder(ctx context.Context) (embedding.Embedder, error) {
	return NewEmbedder(ctx, DefaultEmbeddingConfig())
}
