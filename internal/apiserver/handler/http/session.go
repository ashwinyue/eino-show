// Package http 提供 HTTP 处理器.
package http

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/gin-gonic/gin"

	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ===== Session 请求/响应类型（对齐 WeKnora）=====

// SessionAgentConfig 会话级别的 Agent 配置（对齐 WeKnora 前端）
type SessionAgentConfig struct {
	Enabled        bool     `json:"enabled"`
	MaxIterations  int      `json:"max_iterations"`
	Temperature    float64  `json:"temperature"`
	KnowledgeBases []string `json:"knowledge_bases"`
	KnowledgeIDs   []string `json:"knowledge_ids"`
	AllowedTools   []string `json:"allowed_tools"`
}

// CreateSessionRequest 创建会话请求（对齐 WeKnora 前端）
type CreateSessionRequest struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	AgentID     string              `json:"agent_id"`
	AgentConfig *SessionAgentConfig `json:"agent_config"` // 前端发送的 agent 配置
}

// SessionResponse 会话响应
type SessionResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TenantID    uint64    `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

// StopSessionRequest 停止会话请求
type StopSessionRequest struct {
	MessageID string `json:"message_id" binding:"required"`
}

// MentionedItemRequest 对齐 WeKnora 的 @提及项
type MentionedItemRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`    // "kb" or "file"
	KBType string `json:"kb_type"` // "document" or "faq"
}

// CreateKnowledgeQARequest 对齐 WeKnora 的问答请求
type CreateKnowledgeQARequest struct {
	Query            string                 `json:"query" binding:"required"`
	KnowledgeBaseIDs []string               `json:"knowledge_base_ids"`
	KnowledgeIDs     []string               `json:"knowledge_ids"`
	AgentEnabled     bool                   `json:"agent_enabled"`
	AgentID          string                 `json:"agent_id"`
	WebSearchEnabled bool                   `json:"web_search_enabled"`
	SummaryModelID   string                 `json:"summary_model_id"`
	MentionedItems   []MentionedItemRequest `json:"mentioned_items"`
	DisableTitle     bool                   `json:"disable_title"`
}

// GenerateTitleRequest 对齐 WeKnora 的标题生成请求
type GenerateTitleRequest struct {
	Messages []MessageRequest `json:"messages" binding:"required"`
}

// MessageRequest 消息请求
type MessageRequest struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SearchKnowledgeRequest 知识搜索请求
type SearchKnowledgeRequest struct {
	Query            string   `json:"query" binding:"required"`
	KnowledgeBaseID  string   `json:"knowledge_base_id"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids"`
	KnowledgeIDs     []string `json:"knowledge_ids"`
}

// ExecuteRequest 问答执行请求（简化版）
type ExecuteRequest struct {
	Question string `json:"question" binding:"required"`
}

// ===== Session Handler =====

// CreateSession 创建会话（对齐 WeKnora POST /sessions）
func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := contextx.TenantID(c.Request.Context())

	_ = tenantID
	resp, err := h.biz.Session().Create(c.Request.Context(), &v1.CreateSessionRequest{
		Title:       req.Title,
		Description: req.Description,
		AgentID:     req.AgentID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetSession 获取会话详情（对齐 WeKnora GET /sessions/:id）
func (h *Handler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id is required"})
		return
	}

	resp, err := h.biz.Session().Get(c.Request.Context(), &v1.GetSessionRequest{Id: sessionID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListSessions 获取会话列表（对齐 WeKnora GET /sessions）
func (h *Handler) ListSessions(c *gin.Context) {
	resp, err := h.biz.Session().List(c.Request.Context(), &v1.ListSessionsRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateSession 更新会话（对齐 WeKnora PUT /sessions/:id）
func (h *Handler) UpdateSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id is required"})
		return
	}

	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.biz.Session().Update(c.Request.Context(), sessionID, &v1.UpdateSessionRequest{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteSession 删除会话（对齐 WeKnora DELETE /sessions/:id）
func (h *Handler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id is required"})
		return
	}

	if _, err := h.biz.Session().Delete(c.Request.Context(), &v1.DeleteSessionRequest{Id: sessionID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Session deleted successfully",
	})
}

// StopSession 停止会话对话（对齐 WeKnora POST /sessions/:session_id/stop）
func (h *Handler) StopSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req StopSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: implement Stop
	_ = sessionID
	_ = req
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Generation stopped",
	})
}

// QA 问答（流式）- 对齐 WeKnora POST /sessions/:session_id/qa
func (h *Handler) QA(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.streamQA(c, sessionID, req.Question)
}

// KnowledgeQA 知识问答（对齐 WeKnora POST /knowledge-chat/:session_id）
func (h *Handler) KnowledgeQA(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req CreateKnowledgeQARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.streamQAWithOptions(c, sessionID, req)
}

// AgentQA Agent问答（对齐 WeKnora POST /agent-chat/:session_id）
func (h *Handler) AgentQA(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req CreateKnowledgeQARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Agent 模式默认启用
	req.AgentEnabled = true
	h.streamQAWithOptions(c, sessionID, req)
}

// GenerateTitle 生成会话标题（对齐 WeKnora POST /sessions/:session_id/generate_title）
func (h *Handler) GenerateTitle(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	var req GenerateTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用 biz 层生成标题
	v1Messages := make([]v1.MessageRequest, len(req.Messages))
	for i, m := range req.Messages {
		v1Messages[i] = v1.MessageRequest{ID: m.ID, Role: m.Role, Content: m.Content}
	}
	title, err := h.biz.Session().GenerateTitle(c.Request.Context(), sessionID, &v1.GenerateTitleRequest{
		Messages: v1Messages,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    title,
	})
}

// ContinueStream 继续流式响应（对齐 WeKnora GET /sessions/continue-stream/:session_id）
func (h *Handler) ContinueStream(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	messageID := c.Query("message_id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 发送完成事件 (当前简化实现，后续可接入 StreamManager)
	c.SSEvent("done", map[string]interface{}{
		"session_id": sessionID,
		"message_id": messageID,
		"done":       true,
	})
}

// SearchKnowledge 知识搜索（对齐 WeKnora POST /knowledge-search）
func (h *Handler) SearchKnowledge(c *gin.Context) {
	var req SearchKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 合并 knowledge_base_id 和 knowledge_base_ids
	kbIDs := req.KnowledgeBaseIDs
	if req.KnowledgeBaseID != "" && len(kbIDs) == 0 {
		kbIDs = []string{req.KnowledgeBaseID}
	}

	if len(kbIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "knowledge_base_id or knowledge_base_ids is required"})
		return
	}

	// 调用 biz 层搜索知识
	results, err := h.biz.Session().SearchKnowledge(c.Request.Context(), "", &v1.CreateKnowledgeQARequest{
		Query:            req.Query,
		KnowledgeBaseIDs: kbIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// ===== 流式响应辅助方法（统一使用 ADK）=====

// streamQA 流式问答内部实现
func (h *Handler) streamQA(c *gin.Context, sessionID, question string) {
	h.streamQAWithADK(c, sessionID, CreateKnowledgeQARequest{
		Query:        question,
		AgentEnabled: true,
	})
}

// streamQAWithOptions 带选项的流式问答
func (h *Handler) streamQAWithOptions(c *gin.Context, sessionID string, req CreateKnowledgeQARequest) {
	h.streamQAWithADK(c, sessionID, req)
}

// streamQAWithADK 使用 ADK Agent 的流式问答（对齐 WeKnora 实现）
func (h *Handler) streamQAWithADK(c *gin.Context, sessionID string, req CreateKnowledgeQARequest) {
	ctx := c.Request.Context()
	requestID := c.GetHeader("X-Request-ID")

	// 1. 保存用户消息（对齐 WeKnora helpers.go createUserMessage）
	if _, err := h.biz.Session().SaveMessage(ctx, sessionID, "user", req.Query, requestID); err != nil {
		c.Header("Content-Type", "text/event-stream")
		c.SSEvent("error", map[string]string{"error": "failed to save user message"})
		return
	}

	// 2. 获取 ADK Runner 和消息（Eino ADK 标准方式）
	runnerInterface, messageID, messages, err := h.biz.Session().GetADKRunner(ctx, sessionID, &v1.CreateKnowledgeQARequest{
		Query:            req.Query,
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
		KnowledgeIDs:     req.KnowledgeIDs,
		AgentEnabled:     true, // 统一使用 Agent 模式
		SummaryModelID:   req.SummaryModelID,
		WebSearchEnabled: req.WebSearchEnabled,
	})
	if err != nil {
		c.Header("Content-Type", "text/event-stream")
		c.SSEvent("error", map[string]string{"error": err.Error()})
		return
	}

	// 3. 创建助手消息（对齐 WeKnora helpers.go createAssistantMessage，初始 is_completed=false）
	assistantMsg, err := h.biz.Session().SaveMessage(ctx, sessionID, "assistant", "", messageID)
	if err != nil {
		c.Header("Content-Type", "text/event-stream")
		c.SSEvent("error", map[string]string{"error": "failed to create assistant message"})
		return
	}

	// 类型断言
	runner, ok := runnerInterface.(*adk.Runner)
	if !ok {
		c.Header("Content-Type", "text/event-stream")
		c.SSEvent("error", map[string]string{"error": "invalid runner type"})
		return
	}

	// 转换消息格式
	var adkMessages []adk.Message
	for _, msg := range messages {
		adkMessages = append(adkMessages, msg)
	}

	// 4. 异步生成会话标题（对齐 WeKnora：流开始时并行执行）
	// 使用 channel 接收生成的标题，以便在 stop 事件前发送
	var titleChan chan string
	if !req.DisableTitle {
		titleChan = make(chan string, 1)
		go func() {
			title := h.biz.Session().GenerateTitleSync(context.Background(), sessionID, req.Query)
			titleChan <- title
		}()
	}

	// 使用 ADK SSE Handler 处理，传入 assistantMsg.ID 用于更新
	// 创建更新消息的回调函数
	updateFn := func(ctx context.Context, messageID, content string, agentSteps []v1.AgentStep, isCompleted bool) error {
		return h.biz.Session().UpdateMessageContent(ctx, messageID, content, agentSteps, isCompleted)
	}
	sseHandler := NewADKSSEHandlerWithSave(runner, sessionID, assistantMsg.ID, adkMessages, updateFn, ctx)
	sseHandler.SetTitleChannel(titleChan) // 设置标题 channel
	sseHandler.HandleStream(c)
}
