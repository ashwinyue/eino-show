// Package agent 提供 Agent 业务逻辑.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	agentpkg "github.com/ashwinyue/eino-show/internal/pkg/agent"
	chatagent "github.com/ashwinyue/eino-show/internal/pkg/agent/chat"
	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/cloudwego/eino/schema"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AgentBiz Agent 业务接口.
// 使用 proto 生成的类型作为请求和响应
type AgentBiz interface {
	// Create 创建自定义 Agent
	Create(ctx context.Context, req *v1.CreateAgentRequest) (*v1.CreateAgentResponse, error)
	// Get 获取 Agent 详情
	Get(ctx context.Context, req *v1.GetAgentRequest) (*v1.GetAgentResponse, error)
	// List 获取 Agent 列表
	List(ctx context.Context, req *v1.ListAgentsRequest) (*v1.ListAgentsResponse, error)
	// Update 更新 Agent
	Update(ctx context.Context, req *v1.UpdateAgentRequest) (*v1.UpdateAgentResponse, error)
	// Delete 删除 Agent
	Delete(ctx context.Context, req *v1.DeleteAgentRequest) (*v1.DeleteAgentResponse, error)
	// ListBuiltin 获取内置 Agent 列表
	ListBuiltin(ctx context.Context) []*v1.BuiltinAgent

	// Execute 执行 Agent（流式响应）
	Execute(ctx context.Context, req *v1.ExecuteRequest) (io.ReadCloser, error)
}

type agentBiz struct {
	store store.IStore
}

// New 创建 AgentBiz 实例.
func New(store store.IStore) AgentBiz {
	return &agentBiz{store: store}
}

// 确保 agentBiz 实现了 AgentBiz 接口.
var _ AgentBiz = (*agentBiz)(nil)

// Create 创建自定义 Agent.
func (b *agentBiz) Create(ctx context.Context, req *v1.CreateAgentRequest) (*v1.CreateAgentResponse, error) {
	// TODO: 从上下文获取 tenantID
	agentM := &model.CustomAgentM{
		TenantID:    1, // TODO: 从认证上下文获取
		Name:        req.Name,
		Description: stringPtr(req.Description),
		Avatar:      stringPtr(req.Avatar),
		IsBuiltin:   false,
		Config:      protoConfigToJSON(req.Config),
	}

	if err := b.store.CustomAgent().Create(ctx, agentM); err != nil {
		return nil, err
	}

	return &v1.CreateAgentResponse{
		Id:        agentM.ID,
		Name:      agentM.Name,
		CreatedAt: timePtrToProto(agentM.CreatedAt),
	}, nil
}

// Get 获取 Agent 详情.
func (b *agentBiz) Get(ctx context.Context, req *v1.GetAgentRequest) (*v1.GetAgentResponse, error) {
	agentM, err := b.store.CustomAgent().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetAgentResponse{
		Agent: modelAgentToProto(agentM),
	}, nil
}

// List 获取 Agent 列表.
func (b *agentBiz) List(ctx context.Context, req *v1.ListAgentsRequest) (*v1.ListAgentsResponse, error) {
	_ = req
	// TODO: 从上下文获取 tenantID
	customAgents, err := b.store.CustomAgent().GetByTenantID(ctx, 1)
	if err != nil {
		return nil, err
	}

	pbAgents := make([]*v1.Agent, 0, len(customAgents))
	for _, a := range customAgents {
		pbAgents = append(pbAgents, modelAgentToProto(a))
	}

	return &v1.ListAgentsResponse{
		Agents: pbAgents,
		Total:  int64(len(customAgents)),
	}, nil
}

// Update 更新 Agent.
func (b *agentBiz) Update(ctx context.Context, req *v1.UpdateAgentRequest) (*v1.UpdateAgentResponse, error) {
	agentM, err := b.store.CustomAgent().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		agentM.Name = *req.Name
	}
	if req.Description != nil {
		agentM.Description = req.Description
	}
	if req.Avatar != nil {
		agentM.Avatar = req.Avatar
	}

	// 更新 Config
	if req.Config != nil {
		agentM.Config = protoConfigToJSON(req.Config)
	}

	if err := b.store.CustomAgent().Update(ctx, agentM); err != nil {
		return nil, err
	}

	return &v1.UpdateAgentResponse{
		Id:        agentM.ID,
		Name:      agentM.Name,
		UpdatedAt: timePtrToProto(agentM.UpdatedAt),
	}, nil
}

// Delete 删除 Agent.
func (b *agentBiz) Delete(ctx context.Context, req *v1.DeleteAgentRequest) (*v1.DeleteAgentResponse, error) {
	if err := b.store.CustomAgent().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteAgentResponse{
		Success: true,
		Message: "Agent deleted successfully",
	}, nil
}

// ListBuiltin 获取内置 Agent 列表.
func (b *agentBiz) ListBuiltin(ctx context.Context) []*v1.BuiltinAgent {
	_ = ctx
	result := make([]*v1.BuiltinAgent, 0, len(BuiltinAgents))
	for _, a := range BuiltinAgents {
		result = append(result, &v1.BuiltinAgent{
			Id:          a.ID,
			Type:        a.Type,
			Name:        a.Name,
			Description: a.Description,
		})
	}
	return result
}

// Execute 执行 Agent (默认实现，需要工厂).
func (b *agentBiz) Execute(ctx context.Context, req *v1.ExecuteRequest) (io.ReadCloser, error) {
	return nil, &notImplementedError{message: "Execute requires Agent factory - use ProvideAgentBizWithFactory"}
}

// executeWithFactory 使用工厂执行 Agent.
func executeWithFactory(ctx context.Context, factory *agentpkg.Factory, store store.IStore, req *v1.ExecuteRequest) (io.ReadCloser, error) {
	// 获取会话信息
	sessionM, err := store.Session().Get(ctx, where.F("id", req.SessionId))
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 获取 Agent ID (默认使用 quick-answer)
	agentID := "quick-answer"
	if sessionM.AgentID != nil && *sessionM.AgentID != "" {
		agentID = *sessionM.AgentID
	}

	// 获取内置 Agent
	builtin := GetBuiltinAgent(agentID)
	if builtin == nil {
		return nil, fmt.Errorf("unknown builtin agent: %s", agentID)
	}

	// 创建 Agent 实例
	agentInstance, err := CreateBuiltinAgent(ctx, factory, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// 构建消息列表
	messages := buildMessagesFromSession(sessionM, req.Question)

	// 创建 SSE 流式响应
	stream := newAgentStream(ctx, builtin.Type, agentInstance, messages)

	return stream, nil
}

// buildMessagesFromSession 从会话历史构建消息列表.
func buildMessagesFromSession(session *model.SessionM, userQuestion string) []*schema.Message {
	// TODO: 从数据库加载历史消息
	// 目前只返回当前用户消息
	return []*schema.Message{
		schema.UserMessage(userQuestion),
	}
}

// newAgentStream 创建 Agent 流式响应.
func newAgentStream(ctx context.Context, agentType string, agentInstance interface{}, messages []*schema.Message) io.ReadCloser {
	return &agentStream{
		ctx:         ctx,
		agentType:   agentType,
		agentInstance: agentInstance,
		messages:    messages,
		resultChan:  make(chan []byte, 10),
		doneChan:    make(chan struct{}),
	}
}

// agentStream Agent 流式响应实现.
type agentStream struct {
	ctx         context.Context
	agentType   string
	agentInstance interface{}
	messages    []*schema.Message
	resultChan  chan []byte
	doneChan    chan struct{}
	closed      bool
}

func (s *agentStream) Read(p []byte) (n int, err error) {
	select {
	case <-s.ctx.Done():
		return 0, s.ctx.Err()
	case data, ok := <-s.resultChan:
		if !ok {
			return 0, io.EOF
		}
		copy(p, data)
		return len(data), nil
	}
}

func (s *agentStream) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	close(s.doneChan)
	return nil
}

// run 执行 Agent 并发送结果到 resultChan.
// 这个方法应该在 goroutine 中调用.
func (s *agentStream) run() {
	defer close(s.resultChan)

	switch s.agentType {
	case "chat":
		s.runChatAgent()
	default:
		s.sendError(fmt.Sprintf("unsupported agent type: %s", s.agentType))
	}
}

// runChatAgent 执行 Chat Agent.
func (s *agentStream) runChatAgent() {
	agent, ok := s.agentInstance.(*chatagent.Agent)
	if !ok {
		s.sendError("agent instance is not ChatAgent")
		return
	}

	// 生成回复
	msg, err := agent.Generate(s.ctx, s.messages)
	if err != nil {
		s.sendError(err.Error())
		return
	}

	// 发送结果
	s.sendEvent("agent_complete", map[string]interface{}{
		"answer": msg.Content,
	})
}

// sendEvent 发送 SSE 事件.
func (s *agentStream) sendEvent(event string, data interface{}) {
	eventData := map[string]interface{}{
		"event": event,
		"data":  data,
	}
	jsonBytes, _ := json.Marshal(eventData)
	jsonBytes = append(jsonBytes, '\n')

	select {
	case s.resultChan <- jsonBytes:
	case <-s.ctx.Done():
	}
}

// sendError 发送错误事件.
func (s *agentStream) sendError(message string) {
	s.sendEvent("agent_error", map[string]interface{}{
		"error": message,
	})
}

// NotImplementedReadCloser 用于未实现功能的 ReadCloser.
type NotImplementedReadCloser struct {
	message string
}

func (n *NotImplementedReadCloser) Read(p []byte) (int, error) {
	return 0, &notImplementedError{message: n.message}
}

func (n *NotImplementedReadCloser) Close() error {
	return nil
}

type notImplementedError struct {
	message string
}

func (e *notImplementedError) Error() string {
	return e.message
}

// ===== 类型转换函数 =====

// modelAgentToProto 将 Model CustomAgentM 转换为 Proto Agent
func modelAgentToProto(a *model.CustomAgentM) *v1.Agent {
	description := ""
	if a.Description != nil {
		description = *a.Description
	}
	avatar := ""
	if a.Avatar != nil {
		avatar = *a.Avatar
	}
	return &v1.Agent{
		Id:          a.ID,
		Name:        a.Name,
		Description: description,
		Avatar:      avatar,
		IsBuiltin:   a.IsBuiltin,
		Config:      modelConfigToProto(a.Config),
		CreatedAt:   timePtrToProto(a.CreatedAt),
		UpdatedAt:   timePtrToProto(a.UpdatedAt),
	}
}

// modelConfigToProto 将 Model CustomAgentConfig 转换为 Proto CustomAgentConfig
func modelConfigToProto(cfg string) *v1.CustomAgentConfig {
	if cfg == "" {
		return nil
	}
	var cfgv1 v1.CustomAgentConfig
	if err := json.Unmarshal([]byte(cfg), &cfgv1); err != nil {
		// 返回默认配置
		return &v1.CustomAgentConfig{
			Temperature:   0.7,
			MaxIterations: 5,
		}
	}
	return &cfgv1
}

// protoConfigToJSON 将 Proto CustomAgentConfig 转换为 JSON 字符串
func protoConfigToJSON(cfg *v1.CustomAgentConfig) string {
	if cfg == nil {
		return "{}"
	}
	data, _ := json.Marshal(cfg)
	return string(data)
}

// ===== 辅助函数 =====

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtrToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
