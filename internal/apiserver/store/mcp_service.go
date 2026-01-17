// Package store 提供 MCP 服务存储.
package store

import (
	"context"
	"encoding/json"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// MCPServiceStore MCP 服务存储接口.
type MCPServiceStore interface {
	// Create 创建 MCP 服务
	Create(ctx context.Context, service *model.MCPServiceM) error

	// Update 更新 MCP 服务
	Update(ctx context.Context, service *model.MCPServiceM) error

	// Delete 删除 MCP 服务
	Delete(ctx context.Context, id string) error

	// GetByID 根据 ID 获取 MCP 服务
	GetByID(ctx context.Context, id string) (*model.MCPServiceM, error)

	// List 获取 MCP 服务列表
	List(ctx context.Context, tenantID int32) ([]*model.MCPServiceM, error)

	// ListEnabled 获取启用的 MCP 服务列表
	ListEnabled(ctx context.Context, tenantID int32) ([]*model.MCPServiceM, error)

	// GetByTenantAndName 根据租户和名称获取 MCP 服务
	GetByTenantAndName(ctx context.Context, tenantID int32, name string) (*model.MCPServiceM, error)
}

// mcpServiceStore MCP 服务存储实现.
type mcpServiceStore struct {
	store *datastore
}

// newMCPServiceStore 创建 MCPServiceStore 实例.
func newMCPServiceStore(store *datastore) *mcpServiceStore {
	return &mcpServiceStore{store: store}
}

// 确保 mcpServiceStore 实现了 MCPServiceStore 接口.
var _ MCPServiceStore = (*mcpServiceStore)(nil)

// Create 创建 MCP 服务.
func (s *mcpServiceStore) Create(ctx context.Context, service *model.MCPServiceM) error {
	// 调用 BeforeCreate 钩子
	if err := model.BeforeCreateMCPService(service); err != nil {
		return err
	}
	return s.store.DB(ctx).Create(service).Error
}

// Update 更新 MCP 服务.
func (s *mcpServiceStore) Update(ctx context.Context, service *model.MCPServiceM) error {
	return s.store.DB(ctx).
		Where("id = ?", service.ID).
		Updates(service).Error
}

// Delete 删除 MCP 服务.
func (s *mcpServiceStore) Delete(ctx context.Context, id string) error {
	return s.store.DB(ctx).
		Where("id = ?", id).
		Delete(&model.MCPServiceM{}).Error
}

// GetByID 根据 ID 获取 MCP 服务.
func (s *mcpServiceStore) GetByID(ctx context.Context, id string) (*model.MCPServiceM, error) {
	var result model.MCPServiceM
	err := s.store.DB(ctx).
		Where("id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List 获取 MCP 服务列表.
func (s *mcpServiceStore) List(ctx context.Context, tenantID int32) ([]*model.MCPServiceM, error) {
	var results []*model.MCPServiceM
	query := s.store.DB(ctx)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.
		Order("created_at DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ListEnabled 获取启用的 MCP 服务列表.
func (s *mcpServiceStore) ListEnabled(ctx context.Context, tenantID int32) ([]*model.MCPServiceM, error) {
	var results []*model.MCPServiceM
	query := s.store.DB(ctx).Where("enabled = ?", true)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.
		Order("created_at DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetByTenantAndName 根据租户和名称获取 MCP 服务.
func (s *mcpServiceStore) GetByTenantAndName(ctx context.Context, tenantID int32, name string) (*model.MCPServiceM, error) {
	var result model.MCPServiceM
	err := s.store.DB(ctx).
		Where("tenant_id = ?", tenantID).
		Where("name = ?", name).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ParseHeaders 解析 Headers JSON 字符串为 MCPHeaders.
func ParseMCPHeaders(headersJSON *string) model.MCPHeaders {
	if headersJSON == nil || *headersJSON == "" {
		return make(model.MCPHeaders)
	}
	var headers model.MCPHeaders
	json.Unmarshal([]byte(*headersJSON), &headers)
	return headers
}

// ParseAuthConfig 解析 AuthConfig JSON 字符串为 MCPAuthConfig.
func ParseMCPAuthConfig(configJSON *string) *model.MCPAuthConfig {
	if configJSON == nil || *configJSON == "" {
		return nil
	}
	var config model.MCPAuthConfig
	if err := json.Unmarshal([]byte(*configJSON), &config); err == nil {
		return &config
	}
	return nil
}

// ParseAdvancedConfig 解析 AdvancedConfig JSON 字符串为 MCPAdvancedConfig.
func ParseMCPAdvancedConfig(configJSON *string) *model.MCPAdvancedConfig {
	if configJSON == nil || *configJSON == "" {
		return model.GetDefaultAdvancedConfig()
	}
	var config model.MCPAdvancedConfig
	if err := json.Unmarshal([]byte(*configJSON), &config); err == nil {
		return &config
	}
	return model.GetDefaultAdvancedConfig()
}

// ParseStdioConfig 解析 StdioConfig JSON 字符串为 MCPStdioConfig.
func ParseMCPStdioConfig(configJSON *string) *model.MCPStdioConfig {
	if configJSON == nil || *configJSON == "" {
		return nil
	}
	var config model.MCPStdioConfig
	if err := json.Unmarshal([]byte(*configJSON), &config); err == nil {
		return &config
	}
	return nil
}

// ParseEnvVars 解析 EnvVars JSON 字符串为 MCPEnvVars.
func ParseMCPEnvVars(envVarsJSON *string) model.MCPEnvVars {
	if envVarsJSON == nil || *envVarsJSON == "" {
		return make(model.MCPEnvVars)
	}
	var envVars model.MCPEnvVars
	json.Unmarshal([]byte(*envVarsJSON), &envVars)
	return envVars
}
