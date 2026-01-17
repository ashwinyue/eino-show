// Package apiserver 路由注册函数
// 对齐 WeKnora 的路由组织方式，将路由注册拆分为独立函数
package apiserver

import (
	"github.com/gin-gonic/gin"

	handler "github.com/ashwinyue/eino-show/internal/apiserver/handler/http"
)

// RegisterAuthRoutes 注册认证相关路由
func RegisterAuthRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/logout", h.Logout)
		auth.POST("/refresh", h.RefreshToken)
		// 以下接口需要认证
		auth.Use(authMiddleware)
		auth.GET("/me", h.GetCurrentUser)
		auth.GET("/tenant", h.GetCurrentTenant)
		auth.GET("/validate", h.ValidateToken)
		auth.POST("/change-password", h.ChangePassword)
	}
}

// RegisterSessionRoutes 注册会话相关路由
func RegisterSessionRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	sessions := r.Group("/sessions", authMiddleware)
	{
		sessions.GET("", h.ListSessions)
		sessions.POST("", h.CreateSession)
		sessions.GET("/continue-stream/:session_id", h.ContinueStream)
		sessions.POST("/:session_id/stop", h.StopSession)
		sessions.POST("/:session_id/qa", h.QA)
		sessions.POST("/:session_id/generate_title", h.GenerateTitle)
		sessions.GET("/:id", h.GetSession)
		sessions.PUT("/:id", h.UpdateSession)
		sessions.DELETE("/:id", h.DeleteSession)
	}
}

// RegisterChatRoutes 注册聊天相关路由
func RegisterChatRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	r.POST("/knowledge-chat/:session_id", authMiddleware, h.KnowledgeQA)
	r.POST("/agent-chat/:session_id", authMiddleware, h.AgentQA)
	r.POST("/knowledge-search", authMiddleware, h.SearchKnowledge)
}

// RegisterKnowledgeBaseRoutes 注册知识库相关路由
func RegisterKnowledgeBaseRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	kb := r.Group("/knowledge-bases", authMiddleware)
	{
		kb.GET("", h.ListKnowledgeBases)
		kb.POST("", h.CreateKnowledgeBase)
		kb.GET("/:id/knowledge", h.ListKnowledges)
		kb.GET("/:id/hybrid-search", h.HybridSearch)
		kb.POST("/:id/knowledge/file", h.UploadKnowledgeFromFile)
		kb.POST("/:id/knowledge/url", h.CreateKnowledgeFromURL)
		kb.POST("/:id/knowledge/manual", h.CreateManualKnowledge)
		kb.GET("/:id", h.GetKnowledgeBase)
		kb.GET("/:id/stats", h.GetKnowledgeStats)
		kb.PUT("/:id", h.UpdateKnowledgeBase)
		kb.DELETE("/:id", h.DeleteKnowledgeBase)
	}
	// 扩展路由
	r.POST("/knowledge-bases/copy", authMiddleware, h.CopyKnowledgeBase)
	r.GET("/knowledge-bases/copy/progress/:task_id", authMiddleware, h.GetKBCloneProgress)
}

// RegisterKnowledgeTagRoutes 注册知识库标签相关路由
func RegisterKnowledgeTagRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	kbTags := r.Group("/knowledge-bases/:id/tags", authMiddleware)
	{
		kbTags.GET("", h.ListTags)
		kbTags.POST("", h.CreateTag)
		kbTags.PUT("/:tag_id", h.UpdateTag)
		kbTags.DELETE("/:tag_id", h.DeleteTag)
	}
}

// RegisterFAQRoutes 注册 FAQ 相关路由
func RegisterFAQRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	faq := r.Group("/knowledge-bases/:id/faq", authMiddleware)
	{
		faq.GET("/entries", h.ListFAQEntries)
		faq.GET("/entries/export", h.ExportFAQEntries)
		faq.GET("/entries/:entry_id", h.GetFAQEntry)
		faq.POST("/entries", h.UpsertFAQEntries)
		faq.POST("/entry", h.CreateFAQEntry)
		faq.PUT("/entries/:entry_id", h.UpdateFAQEntry)
		faq.POST("/entries/:entry_id/similar-questions", h.AddSimilarQuestions)
		faq.PUT("/entries/fields", h.UpdateFAQFieldsBatch)
		faq.PUT("/entries/tags", h.UpdateFAQTagBatch)
		faq.DELETE("/entries", h.DeleteFAQEntries)
		faq.POST("/search", h.SearchFAQ)
	}
	r.GET("/faq/import/progress/:task_id", authMiddleware, h.GetFAQImportProgress)
}

// RegisterKnowledgeRoutes 注册知识相关路由
func RegisterKnowledgeRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	knowledge := r.Group("/knowledge", authMiddleware)
	{
		knowledge.GET("/batch", h.GetKnowledgeBatch)
		knowledge.GET("/search", h.SearchKnowledgeByKeyword)
		knowledge.GET("/:id", h.GetKnowledge)
		knowledge.PUT("/:id", h.UpdateKnowledge)
		knowledge.DELETE("/:id", h.DeleteKnowledge)
		knowledge.GET("/:id/download", h.DownloadKnowledgeFile)
		knowledge.PUT("/manual/:id", h.UpdateManualKnowledge)
		knowledge.PUT("/image/:id/:chunk_id", h.UpdateImageInfo)
		knowledge.PUT("/tags", h.UpdateKnowledgeTagBatch)
	}
}

// RegisterChunkRoutes 注册分块相关路由
func RegisterChunkRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	chunks := r.Group("/chunks", authMiddleware)
	{
		chunks.GET("/:knowledge_id", h.ListKnowledgeChunks)
		chunks.GET("/by-id/:id", h.GetChunk)
		chunks.DELETE("/:knowledge_id/:id", h.DeleteChunk)
		chunks.DELETE("/:knowledge_id", h.DeleteChunksByKnowledgeID)
		chunks.PUT("/:knowledge_id/:id", h.UpdateChunk)
		chunks.DELETE("/by-id/:id/questions", h.DeleteGeneratedQuestion)
	}
}

// RegisterAgentRoutes 注册 Agent 相关路由
func RegisterAgentRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	// custom-agents 路由
	customAgents := r.Group("/custom-agents", authMiddleware)
	{
		customAgents.GET("", h.ListAgents)
		customAgents.GET("/:id", h.GetAgent)
		customAgents.POST("", h.CreateAgent)
		customAgents.PUT("/:id", h.UpdateAgent)
		customAgents.DELETE("/:id", h.DeleteAgent)
	}

	// agents 路由（对齐 WeKnora）
	agents := r.Group("/agents", authMiddleware)
	{
		agents.GET("/placeholders", h.GetPlaceholders)
		agents.GET("/builtin", h.ListBuiltinAgents)
		agents.POST("", h.CreateAgent)
		agents.GET("", h.ListAllAgents)
		agents.GET("/:id", h.GetAgent)
		agents.PUT("/:id", h.UpdateAgent)
		agents.DELETE("/:id", h.DeleteAgent)
		agents.POST("/:id/copy", h.CopyAgent)
	}
}

// RegisterModelRoutes 注册模型相关路由
func RegisterModelRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	models := r.Group("/models", authMiddleware)
	{
		models.GET("", h.ListModels)
		models.GET("/:id", h.GetModel)
		models.POST("", h.CreateModel)
		models.PUT("/:id", h.UpdateModel)
		models.DELETE("/:id", h.DeleteModel)
		models.PUT("/:id/default", h.SetDefaultModel)
	}
	r.GET("/models/providers", authMiddleware, h.ListModelProviders)
}

// RegisterMCPServiceRoutes 注册 MCP 服务相关路由
func RegisterMCPServiceRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	mcp := r.Group("/mcp-services", authMiddleware)
	{
		mcp.GET("", h.ListMCPServices)
		mcp.GET("/:id", h.GetMCPService)
		mcp.POST("", h.CreateMCPService)
		mcp.PUT("/:id", h.UpdateMCPService)
		mcp.DELETE("/:id", h.DeleteMCPService)
		mcp.POST("/:id/test", h.TestMCPService)
		mcp.GET("/:id/tools", h.GetMCPServiceTools)
		mcp.GET("/:id/resources", h.GetMCPServiceResources)
	}
}

// RegisterMessageRoutes 注册消息相关路由
func RegisterMessageRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	messages := r.Group("/messages", authMiddleware)
	{
		messages.GET("/:session_id/load", h.LoadMessages)
		messages.DELETE("/:session_id/:id", h.DeleteMessage)
	}
}

// RegisterTenantRoutes 注册租户相关路由
func RegisterTenantRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	r.GET("/tenants/all", authMiddleware, h.ListAllTenants)
	r.GET("/tenants/search", authMiddleware, h.SearchTenants)
	tenants := r.Group("/tenants", authMiddleware)
	{
		tenants.POST("", h.CreateTenant)
		tenants.GET("/:id", h.GetTenant)
		tenants.PUT("/:id", h.UpdateTenant)
		tenants.DELETE("/:id", h.DeleteTenant)
		tenants.GET("", h.ListTenants)
		tenants.GET("/kv/:key", h.GetTenantKV)
		tenants.PUT("/kv/:key", h.UpdateTenantKV)
	}
}

// RegisterSystemRoutes 注册系统相关路由
func RegisterSystemRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	system := r.Group("/system", authMiddleware)
	{
		system.GET("/info", h.GetSystemInfo)
		system.GET("/minio/buckets", h.ListMinioBuckets)
	}
}

// RegisterWebSearchRoutes 注册 Web 搜索相关路由
func RegisterWebSearchRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	webSearch := r.Group("/web-search", authMiddleware)
	{
		webSearch.GET("/providers", h.GetWebSearchProviders)
	}
}

// RegisterEvaluationRoutes 注册评估相关路由
func RegisterEvaluationRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	evaluation := r.Group("/evaluation", authMiddleware)
	{
		evaluation.POST("/", h.Evaluation)
		evaluation.GET("/", h.GetEvaluationResult)
	}
}

// RegisterInitializationRoutes 注册初始化相关路由
func RegisterInitializationRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	init := r.Group("/initialization", authMiddleware)
	{
		init.GET("/config/:kbId", h.GetCurrentConfigByKB)
		init.POST("/initialize/:kbId", h.InitializeByKB)
		init.PUT("/config/:kbId", h.UpdateKBConfig)
		// Ollama
		init.GET("/ollama/status", h.CheckOllamaStatus)
		init.GET("/ollama/models", h.ListOllamaModels)
		init.POST("/ollama/models/check", h.CheckOllamaModels)
		init.POST("/ollama/models/download", h.DownloadOllamaModel)
		init.GET("/ollama/download/progress/:taskId", h.GetDownloadProgress)
		init.GET("/ollama/download/tasks", h.ListDownloadTasks)
		// Remote API
		init.POST("/remote/check", h.CheckRemoteModel)
		init.POST("/embedding/test", h.TestEmbeddingModel)
		init.POST("/rerank/check", h.CheckRerankModel)
		init.POST("/multimodal/test", h.TestMultimodalFunction)
		// Extract
		init.POST("/extract/text-relation", h.ExtractTextRelations)
		init.POST("/extract/fabri-tag", h.FabriTag)
		init.POST("/extract/fabri-text", h.FabriText)
	}
}

// RegisterUserRoutes 注册用户相关路由（/v1 前缀）
func RegisterUserRoutes(r *gin.RouterGroup, h *handler.Handler, authMiddleware gin.HandlerFunc) {
	users := r.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.Use(authMiddleware)
		users.PUT(":userID/change-password", h.ChangePassword)
		users.PUT(":userID", h.UpdateUser)
		users.DELETE(":userID", h.DeleteUser)
		users.GET(":userID", h.GetUser)
		users.GET("", h.ListUser)
	}
}
