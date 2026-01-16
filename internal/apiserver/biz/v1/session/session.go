// Package session 提供会话业务逻辑.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	chatagent "github.com/ashwinyue/eino-show/internal/pkg/agent/chat"
	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/cloudwego/eino/schema"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SessionBiz 会话业务接口.
// 使用 proto 生成的类型作为请求和响应
type SessionBiz interface {
	// Create 创建会话
	Create(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error)
	// Get 获取会话详情
	Get(ctx context.Context, req *v1.GetSessionRequest) (*v1.GetSessionResponse, error)
	// List 获取会话列表
	List(ctx context.Context, req *v1.ListSessionsRequest) (*v1.ListSessionsResponse, error)
	// Update 更新会话
	Update(ctx context.Context, req *v1.UpdateSessionRequest) (*v1.UpdateSessionResponse, error)
	// Delete 删除会话
	Delete(ctx context.Context, req *v1.DeleteSessionRequest) (*v1.DeleteSessionResponse, error)

	// QA 问答（流式）- 返回流式响应 reader
	QA(ctx context.Context, sessionID string, question string) (io.ReadCloser, error)
}

type sessionBiz struct {
	store store.IStore
}

// New 创建 SessionBiz 实例.
func New(store store.IStore) SessionBiz {
	return &sessionBiz{store: store}
}

// 确保 sessionBiz 实现了 SessionBiz 接口.
var _ SessionBiz = (*sessionBiz)(nil)

// Create 创建会话.
func (b *sessionBiz) Create(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error) {
	// TODO: 从上下文获取 tenantID
	var title *string
	if req.Title != "" {
		title = &req.Title
	}
	var description *string
	if req.Description != "" {
		description = &req.Description
	}
	var agentID *string
	if req.AgentId != "" {
		agentID = &req.AgentId
	}

	agentConfigJSON := protoAgentConfigToJSON(req.AgentConfig)
	contextConfigJSON := protoContextConfigToJSON(req.ContextConfig)

	sessionM := &model.SessionM{
		TenantID:      1, // TODO: 从认证上下文获取
		Title:         title,
		Description:   description,
		AgentID:       agentID,
		AgentConfig:   agentConfigJSON,
		ContextConfig: contextConfigJSON,
	}

	if err := b.store.Session().Create(ctx, sessionM); err != nil {
		return nil, err
	}

	respTitle := ""
	if sessionM.Title != nil {
		respTitle = *sessionM.Title
	}

	return &v1.CreateSessionResponse{
		Id:        sessionM.ID,
		Title:     respTitle,
		CreatedAt: timePtrToProto(sessionM.CreatedAt),
	}, nil
}

// Get 获取会话详情.
func (b *sessionBiz) Get(ctx context.Context, req *v1.GetSessionRequest) (*v1.GetSessionResponse, error) {
	sessionM, err := b.store.Session().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetSessionResponse{
		Session: modelSessionToProto(sessionM),
	}, nil
}

// List 获取会话列表.
func (b *sessionBiz) List(ctx context.Context, req *v1.ListSessionsRequest) (*v1.ListSessionsResponse, error) {
	_ = req
	// TODO: 从上下文获取 tenantID
	sessions, err := b.store.Session().GetByTenantID(ctx, 1)
	if err != nil {
		return nil, err
	}

	pbSessions := make([]*v1.Session, 0, len(sessions))
	for _, s := range sessions {
		pbSessions = append(pbSessions, modelSessionToProto(s))
	}

	return &v1.ListSessionsResponse{
		Sessions: pbSessions,
		Total:    int64(len(sessions)),
	}, nil
}

// Update 更新会话.
func (b *sessionBiz) Update(ctx context.Context, req *v1.UpdateSessionRequest) (*v1.UpdateSessionResponse, error) {
	sessionM, err := b.store.Session().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		sessionM.Title = req.Title
	}
	if req.Description != nil {
		sessionM.Description = req.Description
	}
	if req.AgentConfig != nil {
		sessionM.AgentConfig = protoAgentConfigToJSON(req.AgentConfig)
	}
	if req.ContextConfig != nil {
		sessionM.ContextConfig = protoContextConfigToJSON(req.ContextConfig)
	}

	if err := b.store.Session().Update(ctx, sessionM); err != nil {
		return nil, err
	}

	respTitle := ""
	if sessionM.Title != nil {
		respTitle = *sessionM.Title
	}

	return &v1.UpdateSessionResponse{
		Id:        sessionM.ID,
		Title:     respTitle,
		UpdatedAt: timePtrToProto(sessionM.UpdatedAt),
	}, nil
}

// Delete 删除会话.
func (b *sessionBiz) Delete(ctx context.Context, req *v1.DeleteSessionRequest) (*v1.DeleteSessionResponse, error) {
	if err := b.store.Session().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteSessionResponse{
		Success: true,
		Message: "Session deleted successfully",
	}, nil
}

// QA 问答（流式）.
// 默认实现，需要 Agent 工厂支持.
func (b *sessionBiz) QA(ctx context.Context, sessionID string, question string) (io.ReadCloser, error) {
	// 获取会话信息
	_, err := b.store.Session().Get(ctx, where.F("id", sessionID))
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 创建一个简单的错误响应流
	return &errorReadCloser{
		message: fmt.Sprintf("QA requires Agent factory integration. Session: %s, Question: %s", sessionID, question),
	}, nil
}

// QAWithFactory 使用 Agent 工厂执行问答.
func QAWithFactory(ctx context.Context, factory factoryInterface, store store.IStore, sessionID string, question string) (io.ReadCloser, error) {
	// 获取会话信息
	sessionM, err := store.Session().Get(ctx, where.F("id", sessionID))
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 获取 Agent ID (默认使用 quick-answer)
	agentID := "quick-answer"
	if sessionM.AgentID != nil && *sessionM.AgentID != "" {
		agentID = *sessionM.AgentID
	}

	// 使用 ContextManager 构建消息
	cm := NewContextManager(sessionM)
	messages := cm.BuildMessages("", nil, question)

	// 创建 Agent 实例并执行
	agentInstance, err := factory.CreateAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// 创建流式响应
	stream := newSessionAgentStream(ctx, agentID, agentInstance, messages)
	return stream, nil
}

// factoryInterface Agent 工厂接口 (避免循环依赖).
type factoryInterface interface {
	CreateAgent(ctx context.Context, agentID string) (interface{}, error)
}

// ===== 类型转换函数 =====

// modelSessionToProto 将 Model Session 转换为 Proto
func modelSessionToProto(s *model.SessionM) *v1.Session {
	title := ""
	if s.Title != nil {
		title = *s.Title
	}
	description := ""
	if s.Description != nil {
		description = *s.Description
	}
	agentID := ""
	if s.AgentID != nil {
		agentID = *s.AgentID
	}

	return &v1.Session{
		Id:            s.ID,
		Title:         title,
		Description:   description,
		AgentId:       agentID,
		AgentConfig:   jsonAgentConfigToProto(s.AgentConfig),
		ContextConfig: jsonContextConfigToProto(s.ContextConfig),
		CreatedAt:     timePtrToProto(s.CreatedAt),
		UpdatedAt:     timePtrToProto(s.UpdatedAt),
	}
}

// jsonAgentConfigToProto 将 JSON 字符串转换为 Proto SessionAgentConfig
func jsonAgentConfigToProto(jsonStr *string) *v1.SessionAgentConfig {
	if jsonStr == nil || *jsonStr == "" {
		return nil
	}
	var cfg v1.SessionAgentConfig
	if err := json.Unmarshal([]byte(*jsonStr), &cfg); err != nil {
		return nil
	}
	return &cfg
}

// protoAgentConfigToJSON 将 Proto SessionAgentConfig 转换为 JSON 字符串
func protoAgentConfigToJSON(cfg *v1.SessionAgentConfig) *string {
	if cfg == nil {
		return nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}
	result := string(data)
	return &result
}

// jsonContextConfigToProto 将 JSON 字符串转换为 Proto SessionContextConfig
func jsonContextConfigToProto(jsonStr *string) *v1.SessionContextConfig {
	if jsonStr == nil || *jsonStr == "" {
		return nil
	}
	var cfg v1.SessionContextConfig
	if err := json.Unmarshal([]byte(*jsonStr), &cfg); err != nil {
		return nil
	}
	return &cfg
}

// protoContextConfigToJSON 将 Proto SessionContextConfig 转换为 JSON 字符串
func protoContextConfigToJSON(cfg *v1.SessionContextConfig) *string {
	if cfg == nil {
		return nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}
	result := string(data)
	return &result
}

// timePtrToProto 将 *time.Time 转换为 protobuf Timestamp
func timePtrToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// ===== 辅助类型 =====

// errorReadCloser 用于返回错误信息的 ReadCloser.
type errorReadCloser struct {
	message string
	read    bool
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	if e.read {
		return 0, io.EOF
	}
	e.read = true
	errorData := fmt.Sprintf(`{"event":"agent_error","data":{"error":"%s"}}` + "\n", e.message)
	copy(p, errorData)
	return len(errorData), nil
}

func (e *errorReadCloser) Close() error {
	return nil
}

// newSessionAgentStream 创建会话 Agent 流式响应.
func newSessionAgentStream(ctx context.Context, agentType string, agentInstance interface{}, messages []*schema.Message) io.ReadCloser {
	return &sessionAgentStream{
		ctx:           ctx,
		agentType:     agentType,
		agentInstance: agentInstance,
		messages:      messages,
		resultChan:    make(chan []byte, 10),
		closed:        false,
	}
}

// sessionAgentStream 会话 Agent 流式响应实现.
type sessionAgentStream struct {
	ctx           context.Context
	agentType     string
	agentInstance interface{}
	messages      []*schema.Message
	resultChan    chan []byte
	closed        bool
	started       bool
}

func (s *sessionAgentStream) Read(p []byte) (n int, err error) {
	// 首次调用时启动 Agent
	if !s.started {
		s.started = true
		go s.run()
	}

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

func (s *sessionAgentStream) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	return nil
}

// run 执行 Agent 并发送结果到 resultChan.
func (s *sessionAgentStream) run() {
	defer close(s.resultChan)

	switch s.agentType {
	case "chat":
		s.runChatAgent()
	default:
		s.sendError(fmt.Sprintf("unsupported agent type: %s", s.agentType))
	}
}

// runChatAgent 执行 Chat Agent.
func (s *sessionAgentStream) runChatAgent() {
	agent, ok := s.agentInstance.(*chatagent.Agent)
	if !ok {
		s.sendError("agent instance is not ChatAgent")
		return
	}

	// 发送思考事件
	s.sendEvent("agent_thinking", map[string]interface{}{
		"message": "正在思考...",
	})

	// 生成回复
	msg, err := agent.Generate(s.ctx, s.messages)
	if err != nil {
		s.sendError(err.Error())
		return
	}

	// 发送完成事件
	s.sendEvent("agent_complete", map[string]interface{}{
		"answer": msg.Content,
	})
}

// sendEvent 发送 SSE 事件.
func (s *sessionAgentStream) sendEvent(event string, data interface{}) {
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
func (s *sessionAgentStream) sendError(message string) {
	s.sendEvent("agent_error", map[string]interface{}{
		"error": message,
	})
}
