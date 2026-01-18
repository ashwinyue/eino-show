package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	"github.com/ashwinyue/eino-show/internal/pkg/llmcontext"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

// SessionBiz 会话业务接口.
type SessionBiz interface {
	Create(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error)
	Get(ctx context.Context, req *v1.GetSessionRequest) (*v1.GetSessionResponse, error)
	List(ctx context.Context, req *v1.ListSessionsRequest) (*v1.ListSessionsResponse, error)
	Update(ctx context.Context, id string, req *v1.UpdateSessionRequest) (*v1.UpdateSessionResponse, error)
	Delete(ctx context.Context, req *v1.DeleteSessionRequest) (*v1.DeleteSessionResponse, error)
	// GetADKRunner 获取 ADK Runner 用于流式处理（Eino ADK 标准方式）
	GetADKRunner(ctx context.Context, sessionID string, req *v1.CreateKnowledgeQARequest) (interface{}, string, []*schema.Message, error)
	GenerateTitle(ctx context.Context, sessionID string, req *v1.GenerateTitleRequest) (string, error)
	// GenerateTitleAsync 异步生成会话标题（对齐 WeKnora）
	GenerateTitleAsync(ctx context.Context, sessionID, userMessage string)
	// GenerateTitleSync 同步生成会话标题并返回（用于 SSE 事件）
	GenerateTitleSync(ctx context.Context, sessionID, userMessage string) string
	SearchKnowledge(ctx context.Context, sessionID string, req *v1.CreateKnowledgeQARequest) (interface{}, error)
	ClearContext(ctx context.Context, sessionID string) error
	GetMessages(ctx context.Context, sessionID string) ([]*v1.MessageResponse, error)
	// GetMessagesWithPagination 获取会话消息（支持分页，对齐 WeKnora）
	GetMessagesWithPagination(ctx context.Context, sessionID, limit, beforeTime string) ([]*v1.MessageResponse, error)
	DeleteMessage(ctx context.Context, sessionID, messageID string) error
	// SaveMessage 保存消息到数据库（对齐 WeKnora）
	SaveMessage(ctx context.Context, sessionID, role, content, requestID string) (*model.MessageM, error)
	// UpdateMessageContent 更新消息内容和 agent_steps（用于流式累积）
	UpdateMessageContent(ctx context.Context, messageID, content string, agentSteps []v1.AgentStep, isCompleted bool) error
}

type sessionBiz struct {
	store      store.IStore
	qaExecutor *qaExecutor
	ctxManager llmcontext.Manager
}

// NewWithQA 创建带 QA 配置的 Session Biz.
func NewWithQA(ctx context.Context, s store.IStore, qaCfg *QAConfig) (SessionBiz, error) {
	executor, err := newQAExecutor(ctx, qaCfg)
	if err != nil {
		return nil, err
	}
	// 创建上下文管理器 (使用数据库存储)
	ctxManager := llmcontext.NewManager(&llmcontext.Config{
		MaxTokens: 4096, // 默认 4k tokens
		Storage:   s.ContextStorage(),
	})
	return &sessionBiz{store: s, qaExecutor: executor, ctxManager: ctxManager}, nil
}

// New 创建 Session Biz（不带 QA 功能）.
func New(s store.IStore) SessionBiz {
	return &sessionBiz{store: s, ctxManager: llmcontext.NewDefaultManager(), qaExecutor: nil}
}

// NewWithRedis 创建带 Redis 上下文存储的 Session Biz.
// 如果 redisClient 为 nil，回退到数据库存储.
func NewWithRedis(s store.IStore, redisClient *redis.Client) SessionBiz {
	var storage llmcontext.ContextStorage

	if redisClient != nil {
		// 使用 Redis 存储 (和 WeKnora 一致)
		redisStorage, err := llmcontext.NewRedisStorage(&llmcontext.RedisStorageConfig{
			Client:    redisClient,
			KeyPrefix: "llmcontext:",
			TTL:       24 * time.Hour,
		})
		if err == nil {
			storage = redisStorage
		}
	}

	// Redis 不可用时回退到数据库存储
	if storage == nil {
		storage = s.ContextStorage()
	}

	ctxManager := llmcontext.NewManager(&llmcontext.Config{
		MaxTokens: 4096,
		Storage:   storage,
	})

	return &sessionBiz{store: s, ctxManager: ctxManager}
}

func (b *sessionBiz) Create(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error) {
	tenantID := contextx.TenantID(ctx)
	now := time.Now()

	sessionM := &model.SessionM{
		ID:          uuid.New().String(),
		TenantID:    int32(tenantID),
		Title:       &req.Title,
		Description: &req.Description,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// 设置关联的 Agent
	if req.AgentID != "" {
		sessionM.AgentID = &req.AgentID
	}

	if err := b.store.Session().Create(ctx, sessionM); err != nil {
		return nil, err
	}

	return &v1.CreateSessionResponse{
		Success: true,
		Data:    toSessionResponse(sessionM),
	}, nil
}

func (b *sessionBiz) Get(ctx context.Context, req *v1.GetSessionRequest) (*v1.GetSessionResponse, error) {
	sessionM, err := b.store.Session().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetSessionResponse{
		Success: true,
		Data:    toSessionResponse(sessionM),
	}, nil
}

func (b *sessionBiz) List(ctx context.Context, req *v1.ListSessionsRequest) (*v1.ListSessionsResponse, error) {
	tenantID := contextx.TenantID(ctx)
	opts := where.NewWhere().F("tenant_id", tenantID)
	if req.PageSize > 0 {
		opts.P(req.Page, req.PageSize)
	}

	total, list, err := b.store.Session().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	sessions := make([]*v1.SessionResponse, len(list))
	for i, s := range list {
		sessions[i] = toSessionResponse(s)
	}

	return &v1.ListSessionsResponse{
		Success:  true,
		Data:     sessions,
		Total:    total,
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
	}, nil
}

func (b *sessionBiz) Update(ctx context.Context, id string, req *v1.UpdateSessionRequest) (*v1.UpdateSessionResponse, error) {
	sessionM, err := b.store.Session().Get(ctx, where.F("id", id))
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		sessionM.Title = req.Title
	}
	if req.Description != nil {
		sessionM.Description = req.Description
	}
	now := time.Now()
	sessionM.UpdatedAt = &now

	if err := b.store.Session().Update(ctx, sessionM); err != nil {
		return nil, err
	}

	return &v1.UpdateSessionResponse{
		Success: true,
		Data:    toSessionResponse(sessionM),
	}, nil
}

func (b *sessionBiz) Delete(ctx context.Context, req *v1.DeleteSessionRequest) (*v1.DeleteSessionResponse, error) {
	if err := b.store.Session().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteSessionResponse{
		Success: true,
	}, nil
}

func (b *sessionBiz) GenerateTitle(ctx context.Context, sessionID string, req *v1.GenerateTitleRequest) (string, error) {
	// 获取会话
	sessionM, err := b.store.Session().Get(ctx, where.F("id", sessionID))
	if err != nil {
		return "", err
	}

	// 如果已有标题，直接返回
	if sessionM.Title != nil && *sessionM.Title != "" {
		return *sessionM.Title, nil
	}

	// 获取第一条用户消息
	var userContent string
	if len(req.Messages) > 0 {
		for _, m := range req.Messages {
			if m.Role == "user" {
				userContent = m.Content
				break
			}
		}
	} else {
		// 从数据库获取第一条用户消息
		messages, err := b.store.Message().GetBySessionID(ctx, sessionID)
		if err == nil && len(messages) > 0 {
			for _, m := range messages {
				if m.Role == "user" {
					userContent = m.Content
					break
				}
			}
		}
	}

	if userContent == "" {
		return "", nil // 没有用户消息，无法生成标题
	}

	// 使用 LLM 生成标题
	title, err := b.generateTitleWithLLM(ctx, userContent)
	if err != nil {
		// LLM 失败时回退到简单截取
		title = truncateTitle(userContent, 50)
	}

	// 更新会话标题
	sessionM.Title = &title
	now := time.Now()
	sessionM.UpdatedAt = &now
	if err := b.store.Session().Update(ctx, sessionM); err != nil {
		return "", err
	}

	return title, nil
}

func (b *sessionBiz) SearchKnowledge(ctx context.Context, sessionID string, req *v1.CreateKnowledgeQARequest) (interface{}, error) {
	// 从数据库搜索知识分块
	kbIDs := req.KnowledgeBaseIDs

	if len(kbIDs) == 0 {
		return map[string]interface{}{"results": []interface{}{}, "total": 0}, nil
	}

	// 搜索分块 (简单关键词匹配)
	var allResults []map[string]interface{}
	for _, kbID := range kbIDs {
		_, chunks, err := b.store.Chunk().List(ctx, where.F("knowledge_base_id", kbID))
		if err != nil {
			continue
		}
		for _, chunk := range chunks {
			// 简单包含匹配
			if req.Query != "" && !containsIgnoreCase(chunk.Content, req.Query) {
				continue
			}
			allResults = append(allResults, map[string]interface{}{
				"id":           chunk.ID,
				"knowledge_id": chunk.KnowledgeID,
				"content":      chunk.Content,
				"score":        1.0,
			})
			if len(allResults) >= 10 {
				break
			}
		}
		if len(allResults) >= 10 {
			break
		}
	}

	return map[string]interface{}{
		"results": allResults,
		"total":   len(allResults),
	}, nil
}

// containsIgnoreCase 忽略大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1, c2 := s[i+j], substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// ClearContext 清除会话的 LLM 上下文.
func (b *sessionBiz) ClearContext(ctx context.Context, sessionID string) error {
	if b.ctxManager == nil {
		return nil
	}
	return b.ctxManager.ClearContext(ctx, sessionID)
}

// SaveMessage 保存消息到数据库（对齐 WeKnora 实现）.
// 在事务中同时创建 Message 和 SessionItem 记录，确保数据一致性.
func (b *sessionBiz) SaveMessage(ctx context.Context, sessionID, role, content, requestID string) (*model.MessageM, error) {
	now := time.Now()
	msg := &model.MessageM{
		ID:                  uuid.New().String(),
		SessionID:           sessionID,
		Role:                role,
		Content:             content,
		RequestID:           requestID,
		KnowledgeReferences: "[]",
		IsCompleted:         role == "user", // 用户消息立即完成，助手消息需要等流式完成
		CreatedAt:           &now,
		UpdatedAt:           &now,
	}

	// 使用 CreateWithSessionItem 在事务中同时创建 Message 和 SessionItem
	if err := b.store.Message().CreateWithSessionItem(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// UpdateMessageContent 更新消息内容和 agent_steps（用于流式累积）.
func (b *sessionBiz) UpdateMessageContent(ctx context.Context, messageID, content string, agentSteps []v1.AgentStep, isCompleted bool) error {
	msg, err := b.store.Message().Get(ctx, where.F("id", messageID))
	if err != nil {
		return err
	}

	msg.Content = content
	msg.IsCompleted = isCompleted
	now := time.Now()
	msg.UpdatedAt = &now

	// 保存 agent_steps 到数据库
	if len(agentSteps) > 0 {
		stepsJSON, err := json.Marshal(agentSteps)
		if err == nil {
			stepsStr := string(stepsJSON)
			msg.AgentSteps = &stepsStr
		}
	}

	return b.store.Message().Update(ctx, msg)
}

// loadSessionHistory 从数据库加载会话历史消息并转换为 Eino 格式.
func (b *sessionBiz) loadSessionHistory(ctx context.Context, sessionID string, limit int) ([]*schema.Message, error) {
	messages, err := b.store.Message().GetRecentBySessionID(ctx, sessionID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]*schema.Message, 0, len(messages))
	for _, m := range messages {
		var msg *schema.Message
		switch m.Role {
		case "user":
			msg = schema.UserMessage(m.Content)
		case "assistant":
			msg = schema.AssistantMessage(m.Content, nil)
		case "system":
			msg = schema.SystemMessage(m.Content)
		default:
			continue
		}
		result = append(result, msg)
	}
	return result, nil
}

// GetMessages 获取会话消息列表.
func (b *sessionBiz) GetMessages(ctx context.Context, sessionID string) ([]*v1.MessageResponse, error) {
	messages, err := b.store.Message().GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	result := make([]*v1.MessageResponse, 0, len(messages))
	for _, m := range messages {
		resp := &v1.MessageResponse{
			ID:          m.ID,
			SessionID:   m.SessionID,
			Role:        m.Role,
			Content:     m.Content,
			IsCompleted: m.IsCompleted,
		}
		if m.CreatedAt != nil {
			resp.CreatedAt = *m.CreatedAt
		}
		// 解析 agent_steps
		if m.AgentSteps != nil && *m.AgentSteps != "" {
			var steps []map[string]interface{}
			if json.Unmarshal([]byte(*m.AgentSteps), &steps) == nil {
				resp.AgentSteps = steps
			}
		}
		result = append(result, resp)
	}
	return result, nil
}

// GetMessagesWithPagination 获取会话消息（支持分页，对齐 WeKnora GetMessagesBySessionBeforeTime）.
// 参数:
//   - sessionID: 会话 ID
//   - limit: 每次加载的消息数量
//   - beforeTime: 时间游标，返回此时间之前的消息
func (b *sessionBiz) GetMessagesWithPagination(ctx context.Context, sessionID, limit, beforeTime string) ([]*v1.MessageResponse, error) {
	// 如果没有指定分页参数，返回所有消息
	if limit == "" && beforeTime == "" {
		return b.GetMessages(ctx, sessionID)
	}

	// 解析 limit
	limitNum := 20 // 默认 20 条
	if limit != "" {
		_, err := fmt.Sscanf(limit, "%d", &limitNum)
		if err != nil {
			limitNum = 20
		}
	}

	// 解析 beforeTime
	var beforeTimeParsed time.Time
	var parseErr error
	if beforeTime != "" {
		beforeTimeParsed, parseErr = time.Parse(time.RFC3339Nano, beforeTime)
		if parseErr != nil {
			// 如果解析失败，使用当前时间
			beforeTimeParsed = time.Now()
		}
	} else {
		beforeTimeParsed = time.Now()
	}

	// 从数据库获取指定时间之前的消息
	var messages []*model.MessageM
	var err error

	// 尝试使用分页查询
	if beforeTime != "" {
		messages, err = b.store.Message().GetBySessionIDBeforeTime(ctx, sessionID, beforeTimeParsed, limitNum)
	} else {
		// 如果没有 beforeTime，使用 GetBySessionID 限制数量
		allMessages, err := b.store.Message().GetBySessionID(ctx, sessionID)
		if err == nil {
			// 取最早的 limitNum 条
			if len(allMessages) > limitNum {
				messages = allMessages[:limitNum]
			} else {
				messages = allMessages
			}
		}
	}

	if err != nil {
		return nil, err
	}

	result := make([]*v1.MessageResponse, 0, len(messages))
	for _, m := range messages {
		resp := &v1.MessageResponse{
			ID:          m.ID,
			SessionID:   m.SessionID,
			Role:        m.Role,
			Content:     m.Content,
			IsCompleted: m.IsCompleted,
		}
		if m.CreatedAt != nil {
			resp.CreatedAt = *m.CreatedAt
		}
		// 解析 agent_steps
		if m.AgentSteps != nil && *m.AgentSteps != "" {
			var steps []map[string]interface{}
			if json.Unmarshal([]byte(*m.AgentSteps), &steps) == nil {
				resp.AgentSteps = steps
			}
		}
		result = append(result, resp)
	}
	return result, nil
}

// DeleteMessage 删除指定会话中的消息.
func (b *sessionBiz) DeleteMessage(ctx context.Context, sessionID, messageID string) error {
	return b.store.Message().Delete(ctx, where.NewWhere().F("id", messageID).F("session_id", sessionID))
}

// GetADKRunner 获取 ADK Runner 用于流式处理（Eino ADK 标准方式）.
// 返回: Runner, MessageID, Messages, Error
func (b *sessionBiz) GetADKRunner(ctx context.Context, sessionID string, req *v1.CreateKnowledgeQARequest) (interface{}, string, []*schema.Message, error) {
	if b.qaExecutor == nil {
		return nil, "", nil, fmt.Errorf("QA not configured")
	}
	result, err := b.qaExecutor.getADKAgent(ctx, &AgentQARequest{
		SessionID:        sessionID,
		Query:            req.Query,
		AgentType:        "react",
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
		KnowledgeIDs:     req.KnowledgeIDs,
		ModelID:          req.SummaryModelID,
		WebSearchEnabled: req.WebSearchEnabled,
	})
	if err != nil {
		return nil, "", nil, err
	}
	return result.Runner, result.MessageID, result.Messages, nil
}

// toSessionResponse 将 model.SessionM 转换为 v1.SessionResponse
func toSessionResponse(s *model.SessionM) *v1.SessionResponse {
	resp := &v1.SessionResponse{
		ID:       s.ID,
		TenantID: uint64(s.TenantID),
	}
	if s.Title != nil {
		resp.Title = *s.Title
	}
	if s.Description != nil {
		resp.Description = *s.Description
	}
	if s.AgentID != nil {
		resp.AgentID = *s.AgentID
	}
	if s.CreatedAt != nil {
		resp.CreatedAt = *s.CreatedAt
	}
	if s.UpdatedAt != nil {
		resp.UpdatedAt = *s.UpdatedAt
	}
	return resp
}
