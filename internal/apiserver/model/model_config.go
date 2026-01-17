// Package model 提供 Model 配置类型定义（对齐 WeKnora）.
package model

import (
	"database/sql/driver"
	"encoding/json"
)

// ModelType 模型类型
type ModelType string

const (
	ModelTypeEmbedding   ModelType = "Embedding"
	ModelTypeRerank      ModelType = "Rerank"
	ModelTypeKnowledgeQA ModelType = "KnowledgeQA"
	ModelTypeVLLM        ModelType = "VLLM"
)

// ModelSource 模型来源
type ModelSource string

const (
	ModelSourceLocal       ModelSource = "local"
	ModelSourceRemote      ModelSource = "remote"
	ModelSourceAliyun      ModelSource = "aliyun"
	ModelSourceZhipu       ModelSource = "zhipu"
	ModelSourceVolcengine  ModelSource = "volcengine"
	ModelSourceDeepseek    ModelSource = "deepseek"
	ModelSourceHunyuan     ModelSource = "hunyuan"
	ModelSourceMinimax     ModelSource = "minimax"
	ModelSourceOpenAI      ModelSource = "openai"
	ModelSourceGemini      ModelSource = "gemini"
	ModelSourceMimo        ModelSource = "mimo"
	ModelSourceSiliconFlow ModelSource = "siliconflow"
	ModelSourceJina        ModelSource = "jina"
	ModelSourceOpenRouter  ModelSource = "openrouter"
)

// EmbeddingParameters Embedding 模型参数
type EmbeddingParameters struct {
	Dimension            int `json:"dimension"`
	TruncatePromptTokens int `json:"truncate_prompt_tokens"`
}

// ModelParameters 模型参数
type ModelParameters struct {
	BaseURL             string              `json:"base_url"`
	APIKey              string              `json:"api_key"`
	InterfaceType       string              `json:"interface_type"`
	EmbeddingParameters EmbeddingParameters `json:"embedding_parameters"`
	ParameterSize       string              `json:"parameter_size"`
	Provider            string              `json:"provider"`
	ExtraConfig         map[string]string   `json:"extra_config"`
}

// Value 实现 driver.Valuer 接口
func (c ModelParameters) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ModelParameters) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// ProviderInfo 模型提供商信息
type ProviderInfo struct {
	Name        ModelSource          `json:"name"`
	DisplayName string               `json:"display_name"`
	Description string               `json:"description"`
	DefaultURLs map[ModelType]string `json:"default_urls"`
	ModelTypes  []ModelType          `json:"model_types"`
}

// GetProviders 获取所有支持的模型提供商
func GetProviders() []ProviderInfo {
	return []ProviderInfo{
		{
			Name:        ModelSourceOpenAI,
			DisplayName: "OpenAI",
			Description: "OpenAI GPT 系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://api.openai.com/v1",
				ModelTypeEmbedding:   "https://api.openai.com/v1",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding},
		},
		{
			Name:        ModelSourceAliyun,
			DisplayName: "阿里云 DashScope",
			Description: "阿里云通义千问系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://dashscope.aliyuncs.com/compatible-mode/v1",
				ModelTypeEmbedding:   "https://dashscope.aliyuncs.com/compatible-mode/v1",
				ModelTypeRerank:      "https://dashscope.aliyuncs.com/api/v1",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding, ModelTypeRerank},
		},
		{
			Name:        ModelSourceZhipu,
			DisplayName: "智谱 AI",
			Description: "智谱 GLM 系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://open.bigmodel.cn/api/paas/v4",
				ModelTypeEmbedding:   "https://open.bigmodel.cn/api/paas/v4",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding},
		},
		{
			Name:        ModelSourceVolcengine,
			DisplayName: "火山引擎",
			Description: "字节跳动豆包系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://ark.cn-beijing.volces.com/api/v3",
				ModelTypeEmbedding:   "https://ark.cn-beijing.volces.com/api/v3",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding},
		},
		{
			Name:        ModelSourceDeepseek,
			DisplayName: "DeepSeek",
			Description: "DeepSeek 系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://api.deepseek.com/v1",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA},
		},
		{
			Name:        ModelSourceHunyuan,
			DisplayName: "腾讯混元",
			Description: "腾讯混元系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://api.hunyuan.cloud.tencent.com/v1",
				ModelTypeEmbedding:   "https://api.hunyuan.cloud.tencent.com/v1",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding},
		},
		{
			Name:        ModelSourceMinimax,
			DisplayName: "MiniMax",
			Description: "MiniMax 系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://api.minimax.chat/v1",
				ModelTypeEmbedding:   "https://api.minimax.chat/v1",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding},
		},
		{
			Name:        ModelSourceSiliconFlow,
			DisplayName: "SiliconFlow",
			Description: "SiliconFlow 模型推理平台",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://api.siliconflow.cn/v1",
				ModelTypeEmbedding:   "https://api.siliconflow.cn/v1",
				ModelTypeRerank:      "https://api.siliconflow.cn/v1",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding, ModelTypeRerank},
		},
		{
			Name:        ModelSourceJina,
			DisplayName: "Jina AI",
			Description: "Jina Embedding 和 Rerank 模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeEmbedding: "https://api.jina.ai/v1",
				ModelTypeRerank:    "https://api.jina.ai/v1",
			},
			ModelTypes: []ModelType{ModelTypeEmbedding, ModelTypeRerank},
		},
		{
			Name:        ModelSourceOpenRouter,
			DisplayName: "OpenRouter",
			Description: "OpenRouter 多模型聚合平台",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://openrouter.ai/api/v1",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA},
		},
		{
			Name:        ModelSourceGemini,
			DisplayName: "Google Gemini",
			Description: "Google Gemini 系列模型",
			DefaultURLs: map[ModelType]string{
				ModelTypeKnowledgeQA: "https://generativelanguage.googleapis.com/v1beta",
				ModelTypeEmbedding:   "https://generativelanguage.googleapis.com/v1beta",
			},
			ModelTypes: []ModelType{ModelTypeKnowledgeQA, ModelTypeEmbedding},
		},
	}
}

// GetProvidersByModelType 根据模型类型获取支持的提供商
func GetProvidersByModelType(modelType ModelType) []ProviderInfo {
	var result []ProviderInfo
	for _, p := range GetProviders() {
		for _, mt := range p.ModelTypes {
			if mt == modelType {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

// ModelTypeToFrontend 将后端 ModelType 转换为前端兼容的字符串
func ModelTypeToFrontend(mt ModelType) string {
	switch mt {
	case ModelTypeKnowledgeQA:
		return "chat"
	case ModelTypeEmbedding:
		return "embedding"
	case ModelTypeRerank:
		return "rerank"
	case ModelTypeVLLM:
		return "vllm"
	default:
		return string(mt)
	}
}

// FrontendToModelType 将前端字符串转换为后端 ModelType
func FrontendToModelType(s string) ModelType {
	switch s {
	case "chat":
		return ModelTypeKnowledgeQA
	case "embedding":
		return ModelTypeEmbedding
	case "rerank":
		return ModelTypeRerank
	case "vllm":
		return ModelTypeVLLM
	default:
		return ModelType(s)
	}
}
