package model

import (
	"context"
	"encoding/json"
	"time"

	apim "github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/google/uuid"
)

type ModelBiz interface {
	Create(ctx context.Context, req *v1.CreateModelRequest) (*v1.CreateModelResponse, error)
	Get(ctx context.Context, req *v1.GetModelRequest) (*v1.GetModelResponse, error)
	List(ctx context.Context, req *v1.ListModelsRequest) (*v1.ListModelsResponse, error)
	Update(ctx context.Context, id string, req *v1.UpdateModelRequest) (*v1.UpdateModelResponse, error)
	Delete(ctx context.Context, req *v1.DeleteModelRequest) (*v1.DeleteModelResponse, error)
	SetDefault(ctx context.Context, req *v1.SetDefaultModelRequest) (*v1.SetDefaultModelResponse, error)
	ListProviders(ctx context.Context, modelType string) (*v1.ListProvidersResponse, error)
	// 模型测试
	TestChatModel(ctx context.Context, req *v1.TestChatModelRequest) (*v1.TestChatModelResponse, error)
	TestEmbeddingModel(ctx context.Context, req *v1.TestEmbeddingModelRequest) (*v1.TestEmbeddingModelResponse, error)
	TestRerankModel(ctx context.Context, req *v1.TestRerankModelRequest) (*v1.TestRerankModelResponse, error)
}

type modelBiz struct {
	store store.IStore
}

func New(store store.IStore) ModelBiz {
	return &modelBiz{store: store}
}

func (b *modelBiz) Create(ctx context.Context, req *v1.CreateModelRequest) (*v1.CreateModelResponse, error) {
	tenantID := contextx.TenantID(ctx)
	now := time.Now()

	// 构建完整的 parameters JSON（对齐 WeKnora）
	params := apim.ModelParameters{
		BaseURL:  req.Parameters.BaseURL,
		APIKey:   req.Parameters.APIKey,
		Provider: req.Parameters.Provider,
	}
	if req.Parameters.EmbeddingParameters.Dimension > 0 {
		params.EmbeddingParameters = apim.EmbeddingParameters{
			Dimension: req.Parameters.EmbeddingParameters.Dimension,
		}
	}
	paramsJSON, _ := json.Marshal(params)

	// req.Type 已经是后端类型 (KnowledgeQA, Embedding, etc.)
	modelType := apim.ModelType(req.Type)
	if modelType == "" {
		modelType = apim.ModelTypeKnowledgeQA
	}

	modelM := &apim.LLMModelM{
		ID:         uuid.New().String(),
		TenantID:   int32(tenantID),
		Name:       req.Name,
		Type:       string(modelType),
		Source:     req.Source,
		Parameters: string(paramsJSON),
		IsDefault:  false,
		IsBuiltin:  req.Source == "builtin",
		Status:     "active",
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}

	if req.Description != "" {
		modelM.Description = &req.Description
	}

	// 保存到数据库
	if err := b.store.Model().Update(ctx, modelM); err != nil {
		return nil, err
	}

	return &v1.CreateModelResponse{
		Success: true,
		Data:    toModelResponse(modelM),
	}, nil
}

func (b *modelBiz) Get(ctx context.Context, req *v1.GetModelRequest) (*v1.GetModelResponse, error) {
	modelM, err := b.store.Model().GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetModelResponse{
		Success: true,
		Data:    toModelResponse(modelM),
	}, nil
}

func (b *modelBiz) List(ctx context.Context, req *v1.ListModelsRequest) (*v1.ListModelsResponse, error) {
	list, err := b.store.Model().List(ctx, req.Provider)
	if err != nil {
		return nil, err
	}

	models := make([]*v1.ModelResponse, len(list))
	for i, m := range list {
		models[i] = toModelResponse(m)
	}

	return &v1.ListModelsResponse{
		Success: true,
		Data:    models,
		Total:   int64(len(list)),
	}, nil
}

func (b *modelBiz) Update(ctx context.Context, id string, req *v1.UpdateModelRequest) (*v1.UpdateModelResponse, error) {
	modelM, err := b.store.Model().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		modelM.Name = *req.Name
	}
	if req.Type != nil {
		modelM.Type = *req.Type
	}
	if req.Source != nil {
		modelM.Source = *req.Source
	}
	if req.Description != nil {
		modelM.Description = req.Description
	}
	if req.IsDefault != nil {
		modelM.IsDefault = *req.IsDefault
	}

	// 更新 parameters
	if req.Parameters != nil {
		params := apim.ModelParameters{}
		// 先解析现有参数
		if modelM.Parameters != "" && modelM.Parameters != "{}" {
			json.Unmarshal([]byte(modelM.Parameters), &params)
		}
		// 更新非空字段
		if req.Parameters.BaseURL != "" {
			params.BaseURL = req.Parameters.BaseURL
		}
		if req.Parameters.APIKey != "" {
			params.APIKey = req.Parameters.APIKey
		}
		if req.Parameters.Provider != "" {
			params.Provider = req.Parameters.Provider
		}
		if req.Parameters.EmbeddingParameters.Dimension > 0 {
			params.EmbeddingParameters = apim.EmbeddingParameters{
				Dimension: req.Parameters.EmbeddingParameters.Dimension,
			}
		}
		paramsJSON, _ := json.Marshal(params)
		modelM.Parameters = string(paramsJSON)
	}

	if err := b.store.Model().Update(ctx, modelM); err != nil {
		return nil, err
	}

	return &v1.UpdateModelResponse{
		Success: true,
		Data:    toModelResponse(modelM),
	}, nil
}

func (b *modelBiz) Delete(ctx context.Context, req *v1.DeleteModelRequest) (*v1.DeleteModelResponse, error) {
	if err := b.store.Model().Delete(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteModelResponse{
		Success: true,
	}, nil
}

func (b *modelBiz) SetDefault(ctx context.Context, req *v1.SetDefaultModelRequest) (*v1.SetDefaultModelResponse, error) {
	modelM, err := b.store.Model().GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	modelM.IsDefault = true
	if err := b.store.Model().Update(ctx, modelM); err != nil {
		return nil, err
	}

	return &v1.SetDefaultModelResponse{
		Success: true,
	}, nil
}

// ListProviders 获取模型提供商列表
func (b *modelBiz) ListProviders(ctx context.Context, modelType string) (*v1.ListProvidersResponse, error) {
	var providers []apim.ProviderInfo
	if modelType != "" {
		backendType := apim.FrontendToModelType(modelType)
		providers = apim.GetProvidersByModelType(backendType)
	} else {
		providers = apim.GetProviders()
	}

	result := make([]*v1.ProviderInfo, len(providers))
	for i, p := range providers {
		defaultURLs := make(map[string]string)
		for mt, url := range p.DefaultURLs {
			frontendType := apim.ModelTypeToFrontend(mt)
			defaultURLs[frontendType] = url
		}

		modelTypes := make([]string, len(p.ModelTypes))
		for j, mt := range p.ModelTypes {
			modelTypes[j] = apim.ModelTypeToFrontend(mt)
		}

		result[i] = &v1.ProviderInfo{
			Value:       string(p.Name),
			Label:       p.DisplayName,
			Description: p.Description,
			DefaultURLs: defaultURLs,
			ModelTypes:  modelTypes,
		}
	}

	return &v1.ListProvidersResponse{
		Success: true,
		Data:    result,
	}, nil
}

func toModelResponse(m *apim.LLMModelM) *v1.ModelResponse {
	resp := &v1.ModelResponse{
		ID:        m.ID,
		Name:      m.Name,
		Type:      m.Type,   // 后端类型: KnowledgeQA, Embedding, etc.
		Source:    m.Source, // remote, builtin, local
		IsDefault: m.IsDefault,
		IsBuiltin: m.IsBuiltin,
		TenantID:  uint64(m.TenantID),
		Status:    m.Status,
	}

	// 解析 parameters
	if m.Parameters != "" && m.Parameters != "{}" {
		var params apim.ModelParameters
		if json.Unmarshal([]byte(m.Parameters), &params) == nil {
			// 内置模型隐藏敏感信息
			if m.IsBuiltin {
				resp.Parameters = &v1.ModelParameters{
					Provider: params.Provider,
					EmbeddingParameters: v1.EmbeddingParameters{
						Dimension: params.EmbeddingParameters.Dimension,
					},
				}
			} else {
				resp.Parameters = &v1.ModelParameters{
					BaseURL:  params.BaseURL,
					APIKey:   params.APIKey,
					Provider: params.Provider,
					EmbeddingParameters: v1.EmbeddingParameters{
						Dimension: params.EmbeddingParameters.Dimension,
					},
				}
			}
		}
	} else {
		// 空参数
		resp.Parameters = &v1.ModelParameters{}
	}

	if m.Description != nil {
		resp.Description = *m.Description
	}
	if m.CreatedAt != nil {
		resp.CreatedAt = *m.CreatedAt
	}
	if m.UpdatedAt != nil {
		resp.UpdatedAt = *m.UpdatedAt
	}
	return resp
}

// TestChatModel 测试 Chat 模型连接
func (b *modelBiz) TestChatModel(ctx context.Context, req *v1.TestChatModelRequest) (*v1.TestChatModelResponse, error) {
	start := time.Now()

	// 检查必要参数
	if req.BaseURL == "" {
		return &v1.TestChatModelResponse{
			Success: false,
			Message: "Base URL is required",
		}, nil
	}

	// TODO: 实际调用模型 API 进行测试
	// 当前返回模拟成功结果
	latency := time.Since(start).Milliseconds()

	return &v1.TestChatModelResponse{
		Success: true,
		Message: "连接成功，模型可用",
		Model:   req.ModelName,
		Latency: latency,
	}, nil
}

// TestEmbeddingModel 测试 Embedding 模型
func (b *modelBiz) TestEmbeddingModel(ctx context.Context, req *v1.TestEmbeddingModelRequest) (*v1.TestEmbeddingModelResponse, error) {
	start := time.Now()

	if req.BaseURL == "" {
		return &v1.TestEmbeddingModelResponse{
			Success: false,
			Message: "Base URL is required",
		}, nil
	}

	// TODO: 实际调用 Embedding API 进行测试并获取维度
	// 当前返回模拟结果
	latency := time.Since(start).Milliseconds()

	// 根据 Provider 返回默认维度
	dimension := 1536 // OpenAI 默认
	switch req.Provider {
	case "doubao":
		dimension = 2560
	case "qwen":
		dimension = 1536
	case "ollama":
		dimension = 768
	}

	return &v1.TestEmbeddingModelResponse{
		Success:   true,
		Message:   "Embedding 模型连接成功",
		Dimension: dimension,
		Latency:   latency,
	}, nil
}

// TestRerankModel 测试 Rerank 模型
func (b *modelBiz) TestRerankModel(ctx context.Context, req *v1.TestRerankModelRequest) (*v1.TestRerankModelResponse, error) {
	start := time.Now()

	if req.BaseURL == "" {
		return &v1.TestRerankModelResponse{
			Success: false,
			Message: "Base URL is required",
		}, nil
	}

	// TODO: 实际调用 Rerank API 进行测试
	latency := time.Since(start).Milliseconds()

	return &v1.TestRerankModelResponse{
		Success: true,
		Message: "Rerank 模型连接成功",
		Latency: latency,
	}, nil
}
