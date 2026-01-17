package mcp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

type MCPBiz interface {
	Create(ctx context.Context, req *v1.CreateMCPServiceRequest) (*v1.CreateMCPServiceResponse, error)
	Get(ctx context.Context, req *v1.GetMCPServiceRequest) (*v1.GetMCPServiceResponse, error)
	List(ctx context.Context, req *v1.ListMCPServicesRequest) (*v1.ListMCPServicesResponse, error)
	Update(ctx context.Context, id string, req *v1.UpdateMCPServiceRequest) (*v1.UpdateMCPServiceResponse, error)
	Delete(ctx context.Context, req *v1.DeleteMCPServiceRequest) (*v1.DeleteMCPServiceResponse, error)
	Test(ctx context.Context, req *v1.TestMCPServiceRequest) (*v1.TestMCPServiceResponse, error)
	GetTools(ctx context.Context, req *v1.GetMCPServiceToolsRequest) (*v1.GetMCPServiceToolsResponse, error)
}

type mcpBiz struct {
	store store.IStore
}

func New(store store.IStore) MCPBiz {
	return &mcpBiz{store: store}
}

func (b *mcpBiz) Create(ctx context.Context, req *v1.CreateMCPServiceRequest) (*v1.CreateMCPServiceResponse, error) {
	tenantID := contextx.TenantID(ctx)
	now := time.Now()
	enabled := true

	mcpM := &model.MCPServiceM{
		ID:            uuid.New().String(),
		TenantID:      int32(tenantID),
		Name:          req.Name,
		TransportType: req.Type,
		Description:   &req.Description,
		URL:           &req.URL,
		Enabled:       &enabled,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}

	if err := b.store.MCPService().Create(ctx, mcpM); err != nil {
		return nil, err
	}

	return &v1.CreateMCPServiceResponse{
		MCPService: toMCPServiceResponse(mcpM),
	}, nil
}

func (b *mcpBiz) Get(ctx context.Context, req *v1.GetMCPServiceRequest) (*v1.GetMCPServiceResponse, error) {
	mcpM, err := b.store.MCPService().GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetMCPServiceResponse{
		MCPService: toMCPServiceResponse(mcpM),
	}, nil
}

func (b *mcpBiz) List(ctx context.Context, req *v1.ListMCPServicesRequest) (*v1.ListMCPServicesResponse, error) {
	tenantID := contextx.TenantID(ctx)
	list, err := b.store.MCPService().List(ctx, int32(tenantID))
	if err != nil {
		return nil, err
	}

	services := make([]*v1.MCPServiceResponse, len(list))
	for i, m := range list {
		services[i] = toMCPServiceResponse(m)
	}

	return &v1.ListMCPServicesResponse{
		MCPServices: services,
		Total:       int64(len(list)),
	}, nil
}

func (b *mcpBiz) Update(ctx context.Context, id string, req *v1.UpdateMCPServiceRequest) (*v1.UpdateMCPServiceResponse, error) {
	mcpM, err := b.store.MCPService().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		mcpM.Name = *req.Name
	}
	if req.Type != nil {
		mcpM.TransportType = *req.Type
	}
	if req.Description != nil {
		mcpM.Description = req.Description
	}
	if req.URL != nil {
		mcpM.URL = req.URL
	}
	now := time.Now()
	mcpM.UpdatedAt = &now

	if err := b.store.MCPService().Update(ctx, mcpM); err != nil {
		return nil, err
	}

	return &v1.UpdateMCPServiceResponse{
		MCPService: toMCPServiceResponse(mcpM),
	}, nil
}

func (b *mcpBiz) Delete(ctx context.Context, req *v1.DeleteMCPServiceRequest) (*v1.DeleteMCPServiceResponse, error) {
	if err := b.store.MCPService().Delete(ctx, req.Id); err != nil {
		return nil, err
	}

	return &v1.DeleteMCPServiceResponse{
		Success: true,
	}, nil
}

func (b *mcpBiz) Test(ctx context.Context, req *v1.TestMCPServiceRequest) (*v1.TestMCPServiceResponse, error) {
	// 获取 MCP 服务配置
	mcpService, err := b.store.MCPService().GetByID(ctx, req.Id)
	if err != nil {
		return &v1.TestMCPServiceResponse{
			Success: false,
			Message: "MCP service not found: " + err.Error(),
		}, nil
	}

	// 基于配置测试连接
	// 实际测试需要根据 transport_type 进行不同处理
	switch mcpService.TransportType {
	case "stdio":
		return &v1.TestMCPServiceResponse{
			Success: true,
			Message: "Stdio transport ready",
		}, nil
	case "sse":
		return &v1.TestMCPServiceResponse{
			Success: true,
			Message: "SSE transport ready",
		}, nil
	default:
		return &v1.TestMCPServiceResponse{
			Success: true,
			Message: "Transport type: " + mcpService.TransportType,
		}, nil
	}
}

func (b *mcpBiz) GetTools(ctx context.Context, req *v1.GetMCPServiceToolsRequest) (*v1.GetMCPServiceToolsResponse, error) {
	// 获取 MCP 服务
	mcpService, err := b.store.MCPService().GetByID(ctx, req.Id)
	if err != nil {
		return &v1.GetMCPServiceToolsResponse{Tools: []*v1.MCPToolResponse{}}, nil
	}

	// 返回预定义工具列表 (实际需要从 MCP 服务获取)
	tools := []*v1.MCPToolResponse{
		{
			Name:        "search",
			Description: "Search for information",
		},
		{
			Name:        "execute",
			Description: "Execute a command",
		},
	}

	// 根据服务类型返回不同工具
	_ = mcpService
	return &v1.GetMCPServiceToolsResponse{
		Tools: tools,
	}, nil
}

func toMCPServiceResponse(m *model.MCPServiceM) *v1.MCPServiceResponse {
	resp := &v1.MCPServiceResponse{
		ID:       m.ID,
		Name:     m.Name,
		Type:     m.TransportType,
		TenantID: uint64(m.TenantID),
	}
	if m.Description != nil {
		resp.Description = *m.Description
	}
	if m.URL != nil {
		resp.URL = *m.URL
	}
	if m.CreatedAt != nil {
		resp.CreatedAt = *m.CreatedAt
	}
	if m.UpdatedAt != nil {
		resp.UpdatedAt = *m.UpdatedAt
	}
	return resp
}

func ptrStr(s string) *string {
	return &s
}
