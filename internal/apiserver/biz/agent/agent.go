package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// ErrAgentNotFound Agent 不存在
var ErrAgentNotFound = errors.New("agent not found")

// ErrCannotModifyBuiltin 不能修改内置 Agent
var ErrCannotModifyBuiltin = errors.New("cannot modify builtin agent")

// ErrCannotDeleteBuiltin 不能删除内置 Agent
var ErrCannotDeleteBuiltin = errors.New("cannot delete builtin agent")

type AgentBiz interface {
	Create(ctx context.Context, req *v1.CreateAgentRequest) (*v1.CreateAgentResponse, error)
	Get(ctx context.Context, req *v1.GetAgentRequest) (*v1.GetAgentResponse, error)
	List(ctx context.Context, req *v1.ListAgentsRequest) (*v1.ListAgentsResponse, error)
	Update(ctx context.Context, id string, req *v1.UpdateAgentRequest) (*v1.UpdateAgentResponse, error)
	Delete(ctx context.Context, req *v1.DeleteAgentRequest) (*v1.DeleteAgentResponse, error)
	Copy(ctx context.Context, id string) (*v1.CopyAgentResponse, error)
	ListBuiltin(ctx context.Context) []*v1.BuiltinAgent
	GetPlaceholders(ctx context.Context) *v1.PlaceholdersResponse
	GetConfig(ctx context.Context, id string) (*model.CustomAgentConfig, error)
	Execute(ctx context.Context, sessionID string, req *v1.ExecuteRequest) (io.ReadCloser, error)
}

type agentBiz struct {
	store store.IStore
}

func New(store store.IStore) AgentBiz {
	return &agentBiz{store: store}
}

func (b *agentBiz) Create(ctx context.Context, req *v1.CreateAgentRequest) (*v1.CreateAgentResponse, error) {
	tenantID := contextx.TenantID(ctx)
	now := time.Now()

	configJSON, _ := json.Marshal(req.Config)

	agentM := &model.CustomAgentM{
		ID:          uuid.New().String(),
		TenantID:    int32(tenantID),
		Name:        req.Name,
		Description: &req.Description,
		Avatar:      &req.Avatar,
		Config:      string(configJSON),
		IsBuiltin:   false,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if err := b.store.CustomAgent().Create(ctx, agentM); err != nil {
		return nil, err
	}

	return &v1.CreateAgentResponse{
		Agent: toAgentResponse(agentM),
	}, nil
}

func (b *agentBiz) Get(ctx context.Context, req *v1.GetAgentRequest) (*v1.GetAgentResponse, error) {
	agentM, err := b.store.CustomAgent().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetAgentResponse{
		Agent: toAgentResponse(agentM),
	}, nil
}

func (b *agentBiz) List(ctx context.Context, req *v1.ListAgentsRequest) (*v1.ListAgentsResponse, error) {
	tenantID := contextx.TenantID(ctx)
	opts := where.NewWhere().F("tenant_id", tenantID)
	if req != nil && req.PageSize > 0 {
		opts.P(req.Page, req.PageSize)
	}

	total, list, err := b.store.CustomAgent().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	agents := make([]*v1.AgentResponse, len(list))
	for i, a := range list {
		agents[i] = toAgentResponse(a)
	}

	return &v1.ListAgentsResponse{
		Agents: agents,
		Total:  total,
	}, nil
}

func (b *agentBiz) Update(ctx context.Context, id string, req *v1.UpdateAgentRequest) (*v1.UpdateAgentResponse, error) {
	agentM, err := b.store.CustomAgent().Get(ctx, where.F("id", id))
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		agentM.Name = *req.Name
	}
	if req.Description != nil {
		agentM.Description = req.Description
	}
	if req.Avatar != nil {
		agentM.Avatar = req.Avatar
	}
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		agentM.Config = string(configJSON)
	}
	now := time.Now()
	agentM.UpdatedAt = &now

	if err := b.store.CustomAgent().Update(ctx, agentM); err != nil {
		return nil, err
	}

	return &v1.UpdateAgentResponse{
		Agent: toAgentResponse(agentM),
	}, nil
}

func (b *agentBiz) Delete(ctx context.Context, req *v1.DeleteAgentRequest) (*v1.DeleteAgentResponse, error) {
	// 检查是否为内置 Agent
	agentM, err := b.store.CustomAgent().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}
	if agentM.IsBuiltin {
		return nil, ErrCannotDeleteBuiltin
	}

	if err := b.store.CustomAgent().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteAgentResponse{
		Success: true,
	}, nil
}

// Copy 复制 Agent
func (b *agentBiz) Copy(ctx context.Context, id string) (*v1.CopyAgentResponse, error) {
	tenantID := contextx.TenantID(ctx)

	// 获取源 Agent
	sourceAgent, err := b.store.CustomAgent().Get(ctx, where.F("id", id))
	if err != nil {
		return nil, ErrAgentNotFound
	}

	now := time.Now()
	newName := fmt.Sprintf("%s (副本)", sourceAgent.Name)

	// 创建新 Agent
	newAgent := &model.CustomAgentM{
		ID:          uuid.New().String(),
		TenantID:    int32(tenantID),
		Name:        newName,
		Description: sourceAgent.Description,
		Avatar:      sourceAgent.Avatar,
		Config:      sourceAgent.Config,
		IsBuiltin:   false, // 复制的 Agent 不是内置的
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if err := b.store.CustomAgent().Create(ctx, newAgent); err != nil {
		return nil, err
	}

	return &v1.CopyAgentResponse{
		Agent: toAgentResponse(newAgent),
	}, nil
}

func (b *agentBiz) ListBuiltin(ctx context.Context) []*v1.BuiltinAgent {
	// 从数据库获取内置 Agent
	list, err := b.store.CustomAgent().GetBuiltinAgents(ctx)
	if err != nil || len(list) == 0 {
		// 如果数据库没有，返回默认内置 Agent 信息
		infos := model.GetBuiltinAgentInfos()
		agents := make([]*v1.BuiltinAgent, len(infos))
		for i, info := range infos {
			agents[i] = &v1.BuiltinAgent{
				Id:          info.ID,
				Name:        info.Name,
				Description: info.Description,
				Avatar:      info.Avatar,
				Type:        "builtin",
			}
		}
		return agents
	}

	agents := make([]*v1.BuiltinAgent, len(list))
	for i, a := range list {
		agents[i] = &v1.BuiltinAgent{
			Id:          a.ID,
			Name:        a.Name,
			Description: ptrToString(a.Description),
			Avatar:      ptrToString(a.Avatar),
			Type:        "builtin",
		}
	}
	return agents
}

// GetPlaceholders 获取占位符定义
func (b *agentBiz) GetPlaceholders(ctx context.Context) *v1.PlaceholdersResponse {
	// 使用 model 层的占位符定义
	allDefs := model.AllPlaceholders()
	systemPromptDefs := model.PlaceholdersByField(model.PromptFieldSystemPrompt)
	contextTemplateDefs := model.PlaceholdersByField(model.PromptFieldContextTemplate)

	toV1Placeholders := func(defs []model.PlaceholderDef) []v1.Placeholder {
		result := make([]v1.Placeholder, len(defs))
		for i, d := range defs {
			result[i] = v1.Placeholder{Name: d.Name, Description: d.Description}
		}
		return result
	}

	return &v1.PlaceholdersResponse{
		All:             toV1Placeholders(allDefs),
		SystemPrompt:    toV1Placeholders(systemPromptDefs),
		ContextTemplate: toV1Placeholders(contextTemplateDefs),
	}
}

// GetConfig 获取 Agent 配置（强类型）
func (b *agentBiz) GetConfig(ctx context.Context, id string) (*model.CustomAgentConfig, error) {
	// 检查是否为内置 Agent ID
	if config, ok := model.GetBuiltinAgentConfig(id); ok {
		return &config, nil
	}

	// 从数据库获取
	agentM, err := b.store.CustomAgent().Get(ctx, where.F("id", id))
	if err != nil {
		return nil, ErrAgentNotFound
	}

	var config model.CustomAgentConfig
	if err := json.Unmarshal([]byte(agentM.Config), &config); err != nil {
		return nil, err
	}
	config.EnsureDefaults()
	return &config, nil
}

func (b *agentBiz) Execute(ctx context.Context, sessionID string, req *v1.ExecuteRequest) (io.ReadCloser, error) {
	// 创建 SSE 响应流
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		// 发送开始事件
		pw.Write([]byte("event: start\ndata: {\"session_id\":\"" + sessionID + "\"}\n\n"))

		// 模拟 Agent 执行
		response := "Agent 执行完成。问题: " + req.Question

		// 发送内容事件
		pw.Write([]byte("event: content\ndata: {\"content\":\"" + response + "\"}\n\n"))

		// 发送完成事件
		pw.Write([]byte("event: done\ndata: {\"status\":\"completed\"}\n\n"))
	}()

	return pr, nil
}

// toAgentResponse 将 model.CustomAgentM 转换为 v1.AgentResponse
func toAgentResponse(a *model.CustomAgentM) *v1.AgentResponse {
	var config map[string]interface{}
	json.Unmarshal([]byte(a.Config), &config)

	resp := &v1.AgentResponse{
		ID:        a.ID,
		Name:      a.Name,
		TenantID:  uint64(a.TenantID),
		IsBuiltin: a.IsBuiltin,
		Config:    config,
	}
	if a.Description != nil {
		resp.Description = *a.Description
	}
	if a.Avatar != nil {
		resp.Avatar = *a.Avatar
	}
	return resp
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
