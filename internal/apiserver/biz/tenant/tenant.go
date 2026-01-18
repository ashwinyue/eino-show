// Package tenant 提供租户业务逻辑（对齐 WeKnora）.
package tenant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// TenantBiz 租户业务接口.
type TenantBiz interface {
	// Create 创建租户
	Create(ctx context.Context, req *v1.CreateTenantRequest) (*v1.CreateTenantResponse, error)
	// Get 获取租户
	Get(ctx context.Context, req *v1.GetTenantRequest) (*v1.GetTenantResponse, error)
	// List 获取租户列表
	List(ctx context.Context, req *v1.ListTenantsRequest) (*v1.ListTenantsResponse, error)
	// Update 更新租户
	Update(ctx context.Context, req *v1.UpdateTenantRequest) (*v1.UpdateTenantResponse, error)
	// Delete 删除租户
	Delete(ctx context.Context, req *v1.DeleteTenantRequest) (*v1.DeleteTenantResponse, error)
	// Search 搜索租户
	Search(ctx context.Context, req *v1.SearchTenantsRequest) (*v1.SearchTenantsResponse, error)
	// GetKV 获取 KV 配置
	GetKV(ctx context.Context, tenantID uint64, key string) (*v1.TenantKVResponse, error)
	// UpdateKV 更新 KV 配置
	UpdateKV(ctx context.Context, tenantID uint64, req *v1.UpdateTenantKVRequest) (*v1.UpdateTenantKVResponse, error)
}

// tenantBiz 租户业务实现.
type tenantBiz struct {
	store store.IStore
}

var _ TenantBiz = (*tenantBiz)(nil)

// New 创建租户业务实例.
func New(store store.IStore) TenantBiz {
	return &tenantBiz{store: store}
}

// Create 创建租户（对齐 WeKnora）.
func (b *tenantBiz) Create(ctx context.Context, req *v1.CreateTenantRequest) (*v1.CreateTenantResponse, error) {
	if req.Name == "" {
		return &v1.CreateTenantResponse{
			Success: false,
			Message: "tenant name cannot be empty",
		}, nil
	}

	now := time.Now()
	status := "active"
	storageQuota := req.StorageQuota
	if storageQuota == 0 {
		storageQuota = 10737418240 // 10GB default
	}

	tenantM := &model.TenantM{
		Name:         req.Name,
		Description:  &req.Description,
		Business:     req.Business,
		APIKey:       generateAPIKey(),
		Status:       &status,
		StorageQuota: storageQuota,
		StorageUsed:  0,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if err := b.store.Tenant().Create(ctx, tenantM); err != nil {
		return &v1.CreateTenantResponse{
			Success: false,
			Message: "failed to create tenant: " + err.Error(),
		}, nil
	}

	return &v1.CreateTenantResponse{
		Success: true,
		Message: "Tenant created successfully",
		Tenant:  toTenantFull(tenantM),
	}, nil
}

// Get 获取租户（对齐 WeKnora）.
func (b *tenantBiz) Get(ctx context.Context, req *v1.GetTenantRequest) (*v1.GetTenantResponse, error) {
	if req.ID == 0 {
		return &v1.GetTenantResponse{
			Success: false,
			Message: "tenant ID cannot be 0",
		}, nil
	}

	tenantM, err := b.store.Tenant().GetByID(ctx, req.ID)
	if err != nil {
		return &v1.GetTenantResponse{
			Success: false,
			Message: "tenant not found",
		}, nil
	}

	return &v1.GetTenantResponse{
		Success: true,
		Tenant:  toTenantFull(tenantM),
	}, nil
}

// List 获取租户列表（对齐 WeKnora）.
func (b *tenantBiz) List(ctx context.Context, req *v1.ListTenantsRequest) (*v1.ListTenantsResponse, error) {
	opts := where.NewWhere()
	if req.Status != "" {
		opts.F("status", req.Status)
	}
	if req.Name != "" {
		opts.F("name", req.Name)
	}
	if req.PageSize > 0 {
		opts.P(req.Page, req.PageSize)
	}

	total, list, err := b.store.Tenant().List(ctx, opts)
	if err != nil {
		return &v1.ListTenantsResponse{
			Success: false,
			Message: "failed to list tenants: " + err.Error(),
		}, nil
	}

	tenants := make([]*v1.TenantFull, len(list))
	for i, t := range list {
		tenants[i] = toTenantFull(t)
	}

	return &v1.ListTenantsResponse{
		Success: true,
		Tenants: tenants,
		Total:   total,
	}, nil
}

// Update 更新租户（对齐 WeKnora）.
func (b *tenantBiz) Update(ctx context.Context, req *v1.UpdateTenantRequest) (*v1.UpdateTenantResponse, error) {
	if req.ID == 0 {
		return &v1.UpdateTenantResponse{
			Success: false,
			Message: "tenant ID cannot be 0",
		}, nil
	}

	tenantM, err := b.store.Tenant().GetByID(ctx, req.ID)
	if err != nil {
		return &v1.UpdateTenantResponse{
			Success: false,
			Message: "tenant not found",
		}, nil
	}

	if req.Name != nil {
		tenantM.Name = *req.Name
	}
	if req.Description != nil {
		tenantM.Description = req.Description
	}
	if req.Business != nil {
		tenantM.Business = *req.Business
	}
	if req.Status != nil {
		tenantM.Status = req.Status
	}
	if req.StorageQuota != nil {
		tenantM.StorageQuota = *req.StorageQuota
	}

	now := time.Now()
	tenantM.UpdatedAt = &now

	if err := b.store.Tenant().Update(ctx, tenantM); err != nil {
		return &v1.UpdateTenantResponse{
			Success: false,
			Message: "failed to update tenant: " + err.Error(),
		}, nil
	}

	return &v1.UpdateTenantResponse{
		Success: true,
		Message: "Tenant updated successfully",
		Tenant:  toTenantFull(tenantM),
	}, nil
}

// Delete 删除租户（对齐 WeKnora）.
func (b *tenantBiz) Delete(ctx context.Context, req *v1.DeleteTenantRequest) (*v1.DeleteTenantResponse, error) {
	if req.ID == 0 {
		return &v1.DeleteTenantResponse{
			Success: false,
			Message: "tenant ID cannot be 0",
		}, nil
	}

	if err := b.store.Tenant().Delete(ctx, where.F("id", req.ID)); err != nil {
		return &v1.DeleteTenantResponse{
			Success: false,
			Message: "failed to delete tenant: " + err.Error(),
		}, nil
	}

	return &v1.DeleteTenantResponse{
		Success: true,
		Message: "Tenant deleted successfully",
	}, nil
}

// Search 搜索租户（对齐 WeKnora）.
func (b *tenantBiz) Search(ctx context.Context, req *v1.SearchTenantsRequest) (*v1.SearchTenantsResponse, error) {
	total, list, err := b.store.Tenant().Search(ctx, req.Query, req.Page, req.PageSize)
	if err != nil {
		return &v1.SearchTenantsResponse{
			Success: false,
			Message: "failed to search tenants: " + err.Error(),
		}, nil
	}

	tenants := make([]*v1.TenantFull, len(list))
	for i, t := range list {
		tenants[i] = toTenantFull(t)
	}

	return &v1.SearchTenantsResponse{
		Success: true,
		Tenants: tenants,
		Total:   total,
	}, nil
}

// 支持的 KV 配置键（对齐 WeKnora，使用连字符）
const (
	KVKeyAgentConfig        = "agent-config"
	KVKeyWebSearchConfig    = "web-search-config"
	KVKeyConversationConfig = "conversation-config"
	KVKeyContextConfig      = "context-config"
	KVKeyPromptTemplates    = "prompt-templates"
)

// GetKV 获取 KV 配置.
func (b *tenantBiz) GetKV(ctx context.Context, tenantID uint64, key string) (*v1.TenantKVResponse, error) {
	if tenantID == 0 {
		return nil, errors.New("tenant ID cannot be 0")
	}

	// 获取租户信息
	tenantM, err := b.store.Tenant().GetByID(ctx, tenantID)
	if err != nil {
		return &v1.TenantKVResponse{
			Success: false,
			Message: "tenant not found",
		}, nil
	}

	var value interface{}
	switch key {
	case KVKeyAgentConfig:
		if tenantM.AgentConfig != nil && *tenantM.AgentConfig != "" {
			var config model.AgentConfig
			if json.Unmarshal([]byte(*tenantM.AgentConfig), &config) == nil {
				value = config
			}
		}
	case KVKeyWebSearchConfig:
		if tenantM.WebSearchConfig != nil && *tenantM.WebSearchConfig != "" {
			var config model.WebSearchConfig
			if json.Unmarshal([]byte(*tenantM.WebSearchConfig), &config) == nil {
				value = config
			}
		}
	case KVKeyConversationConfig:
		if tenantM.ConversationConfig != nil && *tenantM.ConversationConfig != "" {
			var config model.ConversationConfig
			if json.Unmarshal([]byte(*tenantM.ConversationConfig), &config) == nil {
				value = config
			}
		} else {
			// 返回默认配置
			value = model.GetDefaultConversationConfig()
		}
	case KVKeyContextConfig:
		if tenantM.ContextConfig != nil && *tenantM.ContextConfig != "" {
			var config model.ContextConfig
			if json.Unmarshal([]byte(*tenantM.ContextConfig), &config) == nil {
				value = config
			}
		}
	case KVKeyPromptTemplates:
		// 返回默认提示词模板（和 WeKnora 一致，从配置文件读取）
		value = model.GetDefaultPromptTemplates()
	default:
		return &v1.TenantKVResponse{
			Success: false,
			Message: "unsupported key: " + key,
		}, nil
	}

	return &v1.TenantKVResponse{
		Success: true,
		Data:    value,
	}, nil
}

// UpdateKV 更新 KV 配置.
func (b *tenantBiz) UpdateKV(ctx context.Context, tenantID uint64, req *v1.UpdateTenantKVRequest) (*v1.UpdateTenantKVResponse, error) {
	if tenantID == 0 {
		return nil, errors.New("tenant ID cannot be 0")
	}

	// 获取租户信息
	tenantM, err := b.store.Tenant().GetByID(ctx, tenantID)
	if err != nil {
		return &v1.UpdateTenantKVResponse{
			Success: false,
			Message: "tenant not found",
		}, nil
	}

	// 序列化新值
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		return &v1.UpdateTenantKVResponse{
			Success: false,
			Message: "invalid value format",
		}, nil
	}
	valueStr := string(valueJSON)

	// 根据 key 更新对应字段
	switch req.Key {
	case KVKeyAgentConfig:
		tenantM.AgentConfig = &valueStr
	case KVKeyWebSearchConfig:
		tenantM.WebSearchConfig = &valueStr
	case KVKeyConversationConfig:
		tenantM.ConversationConfig = &valueStr
	case KVKeyContextConfig:
		tenantM.ContextConfig = &valueStr
	default:
		return &v1.UpdateTenantKVResponse{
			Success: false,
			Message: "unsupported key: " + req.Key,
		}, nil
	}

	now := time.Now()
	tenantM.UpdatedAt = &now

	if err := b.store.Tenant().Update(ctx, tenantM); err != nil {
		return &v1.UpdateTenantKVResponse{
			Success: false,
			Message: "failed to update: " + err.Error(),
		}, nil
	}

	// 返回更新后的配置数据（对齐 WeKnora）
	return &v1.UpdateTenantKVResponse{
		Success: true,
		Message: "Configuration updated successfully",
		Data:    req.Value,
	}, nil
}

// toTenantFull 将 model.TenantM 转换为 v1.TenantFull.
func toTenantFull(t *model.TenantM) *v1.TenantFull {
	tenant := &v1.TenantFull{
		ID:           uint64(t.ID),
		Name:         t.Name,
		APIKey:       t.APIKey,
		Business:     t.Business,
		StorageQuota: t.StorageQuota,
		StorageUsed:  t.StorageUsed,
	}
	if t.Description != nil {
		tenant.Description = *t.Description
	}
	if t.Status != nil {
		tenant.Status = *t.Status
	}
	if t.CreatedAt != nil {
		tenant.CreatedAt = *t.CreatedAt
	}
	if t.UpdatedAt != nil {
		tenant.UpdatedAt = *t.UpdatedAt
	}

	// 解析 RetrieverEngines
	if t.RetrieverEngines != "" && t.RetrieverEngines != "[]" {
		var engines model.RetrieverEngines
		if json.Unmarshal([]byte(t.RetrieverEngines), &engines) == nil {
			tenant.RetrieverEngines = &v1.RetrieverEngines{
				Engines: make([]v1.RetrieverEngineParams, len(engines.Engines)),
			}
			for i, e := range engines.Engines {
				tenant.RetrieverEngines.Engines[i] = v1.RetrieverEngineParams{
					RetrieverType:       string(e.RetrieverType),
					RetrieverEngineType: string(e.RetrieverEngineType),
				}
			}
		}
	}

	// 解析 AgentConfig
	if t.AgentConfig != nil && *t.AgentConfig != "" {
		var agentConfig model.AgentConfig
		if json.Unmarshal([]byte(*t.AgentConfig), &agentConfig) == nil {
			tenant.AgentConfig = &v1.AgentConfig{
				MaxIterations:         agentConfig.MaxIterations,
				ReflectionEnabled:     agentConfig.ReflectionEnabled,
				AllowedTools:          agentConfig.AllowedTools,
				Temperature:           agentConfig.Temperature,
				SystemPrompt:          agentConfig.SystemPrompt,
				UseCustomSystemPrompt: agentConfig.UseCustomSystemPrompt,
				WebSearchEnabled:      agentConfig.WebSearchEnabled,
				WebSearchMaxResults:   agentConfig.WebSearchMaxResults,
				MultiTurnEnabled:      agentConfig.MultiTurnEnabled,
				HistoryTurns:          agentConfig.HistoryTurns,
			}
		}
	}

	// 解析 WebSearchConfig
	if t.WebSearchConfig != nil && *t.WebSearchConfig != "" {
		var webConfig model.WebSearchConfig
		if json.Unmarshal([]byte(*t.WebSearchConfig), &webConfig) == nil {
			tenant.WebSearchConfig = &v1.WebSearchConfig{
				Provider:           webConfig.Provider,
				APIKey:             webConfig.APIKey,
				MaxResults:         webConfig.MaxResults,
				IncludeDate:        webConfig.IncludeDate,
				CompressionMethod:  webConfig.CompressionMethod,
				Blacklist:          webConfig.Blacklist,
				EmbeddingModelID:   webConfig.EmbeddingModelID,
				EmbeddingDimension: webConfig.EmbeddingDimension,
				RerankModelID:      webConfig.RerankModelID,
				DocumentFragments:  webConfig.DocumentFragments,
			}
		}
	}

	// 解析 ConversationConfig
	if t.ConversationConfig != nil && *t.ConversationConfig != "" {
		var convConfig model.ConversationConfig
		if json.Unmarshal([]byte(*t.ConversationConfig), &convConfig) == nil {
			tenant.ConversationConfig = &v1.ConversationConfig{
				Prompt:               convConfig.Prompt,
				ContextTemplate:      convConfig.ContextTemplate,
				Temperature:          convConfig.Temperature,
				MaxCompletionTokens:  convConfig.MaxCompletionTokens,
				MaxRounds:            convConfig.MaxRounds,
				EmbeddingTopK:        convConfig.EmbeddingTopK,
				KeywordThreshold:     convConfig.KeywordThreshold,
				VectorThreshold:      convConfig.VectorThreshold,
				RerankTopK:           convConfig.RerankTopK,
				RerankThreshold:      convConfig.RerankThreshold,
				EnableRewrite:        convConfig.EnableRewrite,
				EnableQueryExpansion: convConfig.EnableQueryExpansion,
				FallbackStrategy:     convConfig.FallbackStrategy,
				FallbackResponse:     convConfig.FallbackResponse,
				FallbackPrompt:       convConfig.FallbackPrompt,
			}
		}
	}

	// 解析 ContextConfig
	if t.ContextConfig != nil && *t.ContextConfig != "" {
		var ctxConfig model.ContextConfig
		if json.Unmarshal([]byte(*t.ContextConfig), &ctxConfig) == nil {
			tenant.ContextConfig = &v1.ContextConfig{
				GlobalContext:        ctxConfig.GlobalContext,
				EnableForAllSessions: ctxConfig.EnableForAllSessions,
			}
		}
	}

	return tenant
}

// generateAPIKey 生成 API Key.
func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
