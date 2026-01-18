package apiserver

import (
	"context"

	"github.com/ashwinyue/eino-show/pkg/core"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	handler "github.com/ashwinyue/eino-show/internal/apiserver/handler/http"
	"github.com/ashwinyue/eino-show/internal/pkg/errno"
	mw "github.com/ashwinyue/eino-show/internal/pkg/middleware/gin"
	"github.com/ashwinyue/eino-show/internal/pkg/server"
)

// ginServer 定义一个使用 Gin 框架开发的 HTTP 服务器.
type ginServer struct {
	srv server.Server
}

// 确保 *ginServer 实现了 server.Server 接口.
var _ server.Server = (*ginServer)(nil)

// NewGinServer 初始化一个新的 Gin 服务器实例.
func (c *ServerConfig) NewGinServer() server.Server {
	// 创建 Gin 引擎
	engine := gin.New()

	// 注册全局中间件，用于恢复 panic、设置 HTTP 头、添加请求 ID 等
	engine.Use(gin.Recovery(), mw.NoCache, mw.Cors, mw.Secure, mw.RequestIDMiddleware())

	// 注册限流中间件
	c.installRateLimitMiddleware(engine)

	// 注册 REST API 路由
	c.InstallRESTAPI(engine)

	httpsrv := server.NewHTTPServer(c.cfg.HTTPOptions, c.cfg.TLSOptions, engine)

	return &ginServer{srv: httpsrv}
}

// installRateLimitMiddleware 安装限流中间件.
func (c *ServerConfig) installRateLimitMiddleware(engine *gin.Engine) {
	// 创建组合限流器配置
	rateLimitCfg := &mw.CombinedConfig{
		GlobalRate: 1000, // 全局每秒 1000 请求
		UserRate:   100,  // 每用户每秒 100 请求
		ExcludePaths: []string{
			"/healthz",
			"/metrics",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
		},
		// 对话接口单独限流（更严格）
		EndpointLimits: []mw.EndpointLimit{
			{Path: "/api/v1/sessions/*/qa", Method: "POST", Rate: 50},                    // 问答接口：50 QPS
			{Path: "/api/v1/knowledge-chat/*", Method: "POST", Rate: 50},                 // 知识问答：50 QPS
			{Path: "/api/v1/agent-chat/*", Method: "POST", Rate: 30},                     // Agent 问答：30 QPS
			{Path: "/api/v1/knowledge-bases/*/knowledge/file", Method: "POST", Rate: 10}, // 文件上传：10 QPS
		},
	}

	// 如果有 Redis 客户端，启用用户级限流
	if c.biz != nil {
		// 尝试从 biz 获取 Redis 客户端（如果可用）
		// 注意：这里简化处理，实际可通过依赖注入获取 Redis 客户端
	}

	// 创建并安装限流中间件
	limiter := mw.NewCombinedRateLimiter(rateLimitCfg)
	engine.Use(limiter.Middleware())
}

// InstallRESTAPI 注册 API 路由。路由的路径和 HTTP 方法，严格遵循 REST 规范.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	// 注册业务无关的 API 接口
	InstallGenericAPI(engine)

	// 创建核心业务处理器
	h := handler.NewHandler(c.biz, c.val, c.cfg.WebSearchOptions)

	// 注册健康检查接口
	engine.GET("/healthz", h.Healthz)

	// 认证中间件
	authMiddleware := mw.AuthnMiddleware(c.retriever)

	// 注册 /api/v1 前缀路由（对齐前端和 WeKnora）
	apiV1 := engine.Group("/api/v1")
	{
		RegisterAuthRoutes(apiV1, h, authMiddleware)
		RegisterSessionRoutes(apiV1, h, authMiddleware)
		RegisterChatRoutes(apiV1, h, authMiddleware)
		RegisterKnowledgeBaseRoutes(apiV1, h, authMiddleware)
		RegisterKnowledgeTagRoutes(apiV1, h, authMiddleware)
		RegisterFAQRoutes(apiV1, h, authMiddleware)
		RegisterKnowledgeRoutes(apiV1, h, authMiddleware)
		RegisterChunkRoutes(apiV1, h, authMiddleware)
		RegisterAgentRoutes(apiV1, h, authMiddleware)
		RegisterModelRoutes(apiV1, h, authMiddleware)
		RegisterMCPServiceRoutes(apiV1, h, authMiddleware)
		RegisterMessageRoutes(apiV1, h, authMiddleware)
		RegisterTenantRoutes(apiV1, h, authMiddleware)
		RegisterSystemRoutes(apiV1, h, authMiddleware)
		RegisterWebSearchRoutes(apiV1, h, authMiddleware)
		RegisterEvaluationRoutes(apiV1, h, authMiddleware)
		RegisterInitializationRoutes(apiV1, h, authMiddleware)
	}

	// 注册 /v1 前缀路由（兼容旧版本）
	v1 := engine.Group("/v1")
	{
		RegisterUserRoutes(v1, h, authMiddleware)
		registerV1CompatRoutes(v1, h, authMiddleware)
	}
}

// registerV1CompatRoutes 注册 /v1 兼容路由
func registerV1CompatRoutes(v1 *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	// Agent 相关路由
	agentv1 := v1.Group("/custom-agents", authMiddleware)
	{
		agentv1.GET("", h.ListAgents)
		agentv1.GET("/:id", h.GetAgent)
		agentv1.POST("", h.CreateAgent)
		agentv1.PUT("/:id", h.UpdateAgent)
		agentv1.DELETE("/:id", h.DeleteAgent)
	}

	// 内置 Agent 路由
	builtinv1 := v1.Group("/agents/builtin", authMiddleware)
	{
		builtinv1.GET("", h.ListBuiltinAgents)
	}

	// Session 相关路由
	sessionv1 := v1.Group("/sessions", authMiddleware)
	{
		sessionv1.GET("", h.ListSessions)
		sessionv1.GET("/:id", h.GetSession)
		sessionv1.POST("", h.CreateSession)
		sessionv1.PUT("/:id", h.UpdateSession)
		sessionv1.DELETE("/:id", h.DeleteSession)
		sessionv1.POST("/:id/stop", h.StopSession)
		sessionv1.POST("/:id/qa", h.QA)
	}

	// Knowledge Base 相关路由
	kbv1 := v1.Group("/knowledge-bases", authMiddleware)
	{
		kbv1.GET("", h.ListKnowledgeBases)
		kbv1.POST("", h.CreateKnowledgeBase)
		kbv1.GET("/:id/knowledge", h.ListKnowledges)
		kbv1.GET("/:id/hybrid-search", h.HybridSearch)
		kbv1.POST("/:id/knowledge/file", h.UploadKnowledgeFromFile)
		kbv1.POST("/:id/knowledge/url", h.CreateKnowledgeFromURL)
		kbv1.POST("/:id/knowledge/manual", h.CreateManualKnowledge)
		kbv1.GET("/:id", h.GetKnowledgeBase)
		kbv1.GET("/:id/stats", h.GetKnowledgeStats)
		kbv1.PUT("/:id", h.UpdateKnowledgeBase)
		kbv1.DELETE("/:id", h.DeleteKnowledgeBase)
	}

	// Knowledge 路由
	knowledgev1 := v1.Group("/knowledge", authMiddleware)
	{
		knowledgev1.DELETE("/:id", h.DeleteKnowledge)
	}

	// Chunk 相关路由
	chunkv1 := v1.Group("/chunks", authMiddleware)
	{
		chunkv1.GET("", h.ListChunks)
		chunkv1.GET("/by-id/:id", h.GetChunk)
		chunkv1.PUT("/:id", h.UpdateChunk)
		chunkv1.DELETE("/:id", h.DeleteChunk)
	}

	// MCP 服务相关路由
	mcpv1 := v1.Group("/mcp-services", authMiddleware)
	{
		mcpv1.GET("", h.ListMCPServices)
		mcpv1.GET("/:id", h.GetMCPService)
		mcpv1.POST("", h.CreateMCPService)
		mcpv1.PUT("/:id", h.UpdateMCPService)
		mcpv1.DELETE("/:id", h.DeleteMCPService)
		mcpv1.POST("/:id/test", h.TestMCPService)
		mcpv1.GET("/:id/tools", h.GetMCPServiceTools)
	}

	// Model 相关路由
	modelv1 := v1.Group("/models", authMiddleware)
	{
		modelv1.GET("", h.ListModels)
		modelv1.GET("/:id", h.GetModel)
		modelv1.POST("", h.CreateModel)
		modelv1.PUT("/:id", h.UpdateModel)
		modelv1.DELETE("/:id", h.DeleteModel)
		modelv1.PUT("/:id/default", h.SetDefaultModel)
	}
}

// InstallGenericAPI 注册业务无关的路由，例如 pprof、404 处理等.
func InstallGenericAPI(engine *gin.Engine) {
	// 注册 pprof 路由
	pprof.Register(engine)

	// 注册 404 路由处理
	engine.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrPageNotFound, nil)
	})
}

// RunOrDie 启动 Gin 服务器，出错则程序崩溃退出.
func (s *ginServer) RunOrDie() {
	s.srv.RunOrDie()
}

// GracefulStop 优雅停止服务器.
func (s *ginServer) GracefulStop(ctx context.Context) {
	s.srv.GracefulStop(ctx)
}
