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
		Success: true,
		Data:    toAgentResponse(agentM),
	}, nil
}

func (b *agentBiz) Get(ctx context.Context, req *v1.GetAgentRequest) (*v1.GetAgentResponse, error) {
	tenantID := contextx.TenantID(ctx)

	// 检查是否是内置 Agent（对齐 WeKnora）
	if model.IsBuiltinAgentID(req.Id) {
		return b.getBuiltinAgent(ctx, req.Id, tenantID)
	}

	agentM, err := b.store.CustomAgent().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetAgentResponse{
		Success: true,
		Data:    toAgentResponse(agentM),
	}, nil
}

// getBuiltinAgent 获取内置 Agent（对齐 WeKnora）
func (b *agentBiz) getBuiltinAgent(ctx context.Context, id string, tenantID uint64) (*v1.GetAgentResponse, error) {
	// 获取默认内置 Agent 信息
	var defaultInfo model.BuiltinAgentInfo
	for _, info := range model.GetBuiltinAgentInfos() {
		if info.ID == id {
			defaultInfo = info
			break
		}
	}
	if defaultInfo.ID == "" {
		return nil, ErrAgentNotFound
	}

	// 获取默认配置
	defaultConfig, _ := model.GetBuiltinAgentConfig(id)

	// 将默认配置转为 map
	configJSON, _ := json.Marshal(defaultConfig)
	var configMap map[string]interface{}
	json.Unmarshal(configJSON, &configMap)

	// 尝试从数据库获取自定义配置
	// 对于内置 agent，不使用 tenant_id 过滤（因为内置 agent 是全局共享的）
	dbAgent, err := b.store.CustomAgent().Get(ctx, where.F("id", id))
	if err == nil && dbAgent != nil && dbAgent.Config != "" {
		// 数据库有记录，只合并用户可自定义的字段
		var dbConfig map[string]interface{}
		json.Unmarshal([]byte(dbAgent.Config), &dbConfig)

		// 只允许用户自定义这些字段（其他使用默认值）
		userCustomizableFields := []string{
			"model_id", "rerank_model_id", "system_prompt", "temperature",
			"knowledge_bases", "kb_selection_mode", "mcp_services", "mcp_selection_mode",
		}
		for _, field := range userCustomizableFields {
			if val, ok := dbConfig[field]; ok && val != nil && val != "" {
				configMap[field] = val
			}
		}
	}

	return &v1.GetAgentResponse{
		Success: true,
		Data: &v1.AgentResponse{
			ID:          id,
			Name:        defaultInfo.Name,
			Description: defaultInfo.Description,
			Avatar:      defaultInfo.Avatar,
			Config:      configMap,
			TenantID:    tenantID,
			IsBuiltin:   true,
		},
	}, nil
}

func (b *agentBiz) List(ctx context.Context, req *v1.ListAgentsRequest) (*v1.ListAgentsResponse, error) {
	tenantID := contextx.TenantID(ctx)
	opts := where.NewWhere().F("tenant_id", tenantID)
	if req != nil && req.PageSize > 0 {
		opts.P(req.Page, req.PageSize)
	}

	_, list, err := b.store.CustomAgent().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	agents := make([]*v1.AgentResponse, len(list))
	for i, a := range list {
		agents[i] = toAgentResponse(a)
	}

	return &v1.ListAgentsResponse{
		Success: true,
		Data:    agents,
	}, nil
}

func (b *agentBiz) Update(ctx context.Context, id string, req *v1.UpdateAgentRequest) (*v1.UpdateAgentResponse, error) {
	tenantID := contextx.TenantID(ctx)

	// 检查是否是内置 Agent（对齐 WeKnora）
	if model.IsBuiltinAgentID(id) {
		return b.updateBuiltinAgent(ctx, id, req, tenantID)
	}

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
		Success: true,
		Data:    toAgentResponse(agentM),
	}, nil
}

// updateBuiltinAgent 更新内置 Agent 配置（对齐 WeKnora）
func (b *agentBiz) updateBuiltinAgent(ctx context.Context, id string, req *v1.UpdateAgentRequest, tenantID uint64) (*v1.UpdateAgentResponse, error) {
	// 获取内置 Agent 默认信息
	var defaultInfo model.BuiltinAgentInfo
	for _, info := range model.GetBuiltinAgentInfos() {
		if info.ID == id {
			defaultInfo = info
			break
		}
	}
	if defaultInfo.ID == "" {
		return nil, ErrAgentNotFound
	}

	now := time.Now()

	// 尝试从数据库获取已存在的自定义配置
	existingAgent, err := b.store.CustomAgent().Get(ctx, where.F("id", id).F("tenant_id", tenantID))
	if err == nil {
		// 存在记录，更新配置
		if req.Config != nil {
			configJSON, _ := json.Marshal(req.Config)
			existingAgent.Config = string(configJSON)
		}
		existingAgent.UpdatedAt = &now

		if err := b.store.CustomAgent().Update(ctx, existingAgent); err != nil {
			return nil, err
		}

		return &v1.UpdateAgentResponse{
			Success: true,
			Data:    toAgentResponse(existingAgent),
		}, nil
	}

	// 不存在记录，创建新记录保存自定义配置
	configJSON, _ := json.Marshal(req.Config)
	newAgent := &model.CustomAgentM{
		ID:          id,
		TenantID:    int32(tenantID),
		Name:        defaultInfo.Name,
		Description: &defaultInfo.Description,
		Avatar:      &defaultInfo.Avatar,
		Config:      string(configJSON),
		IsBuiltin:   true,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if err := b.store.CustomAgent().Create(ctx, newAgent); err != nil {
		return nil, err
	}

	return &v1.UpdateAgentResponse{
		Success: true,
		Data:    toAgentResponse(newAgent),
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
		Success: true,
		Data:    toAgentResponse(newAgent),
	}, nil
}

func (b *agentBiz) ListBuiltin(ctx context.Context) []*v1.BuiltinAgent {
	// 获取所有默认内置 Agent 信息
	defaultInfos := model.GetBuiltinAgentInfos()

	// 从数据库获取有自定义配置的内置 Agent
	dbAgents := make(map[string]*model.CustomAgentM)
	list, err := b.store.CustomAgent().GetBuiltinAgents(ctx)
	if err == nil {
		for _, a := range list {
			dbAgents[a.ID] = a
		}
	}

	// 返回所有默认内置 Agent，用数据库配置覆盖
	agents := make([]*v1.BuiltinAgent, len(defaultInfos))
	for i, info := range defaultInfos {
		name := info.Name
		description := info.Description
		avatar := info.Avatar

		// 如果数据库有记录，用数据库值覆盖（非空字段）
		if dbAgent, ok := dbAgents[info.ID]; ok {
			if dbAgent.Name != "" {
				name = dbAgent.Name
			}
			if dbAgent.Description != nil && *dbAgent.Description != "" {
				description = *dbAgent.Description
			}
			if dbAgent.Avatar != nil && *dbAgent.Avatar != "" {
				avatar = *dbAgent.Avatar
			}
		}

		agents[i] = &v1.BuiltinAgent{
			Id:          info.ID,
			Name:        name,
			Description: description,
			Avatar:      avatar,
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
			result[i] = v1.Placeholder{
				Name:        d.Name,
				Label:       d.Label,
				Description: d.Description,
			}
		}
		return result
	}

	return &v1.PlaceholdersResponse{
		Success: true,
		Data: &v1.PlaceholdersData{
			All:                 toV1Placeholders(allDefs),
			SystemPrompt:        toV1Placeholders(systemPromptDefs),
			ContextTemplate:     toV1Placeholders(contextTemplateDefs),
			AgentSystemPrompt:   toV1Placeholders(model.PlaceholdersByField(model.PromptFieldAgentSystemPrompt)),
			RewriteSystemPrompt: toV1Placeholders(model.PlaceholdersByField(model.PromptFieldRewriteSystemPrompt)),
			RewritePrompt:       toV1Placeholders(model.PlaceholdersByField(model.PromptFieldRewritePrompt)),
			FallbackPrompt:      toV1Placeholders(model.PlaceholdersByField(model.PromptFieldFallbackPrompt)),
		},
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
