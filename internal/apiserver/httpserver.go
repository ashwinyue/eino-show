// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package apiserver

import (
	"context"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

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

	// 注册 REST API 路由
	c.InstallRESTAPI(engine)

	httpsrv := server.NewHTTPServer(c.cfg.HTTPOptions, c.cfg.TLSOptions, engine)

	return &ginServer{srv: httpsrv}
}

// InstallRESTAPI 注册 API 路由。路由的路径和 HTTP 方法，严格遵循 REST 规范.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	// 注册业务无关的 API 接口
	InstallGenericAPI(engine)

	// 创建核心业务处理器
	handler := handler.NewHandler(c.biz, c.val)

	// 注册健康检查接口
	engine.GET("/healthz", handler.Healthz)

	// 注册用户登录和令牌刷新接口。这2个接口比较简单，所以没有 API 版本
	engine.POST("/login", handler.Login)
	// 刷新令牌不需要认证中间件，因为它从请求体中读取 refreshToken 进行验证
	engine.PUT("/refresh-token", handler.RefreshToken)

	// 认证和授权中间件
	authMiddlewares := []gin.HandlerFunc{mw.AuthnMiddleware(c.retriever), mw.AuthzMiddleware(c.authz)}

	// 注册 v1 版本 API 路由分组
	v1 := engine.Group("/v1")
	{
		// 用户相关路由
		userv1 := v1.Group("/users")
		{
			// 创建用户。这里要注意：创建用户是不用进行认证和授权的
			userv1.POST("", handler.CreateUser)
			userv1.Use(authMiddlewares...)                                // 应用中间件。之后的接口需要认证和授权
			userv1.PUT(":userID/change-password", handler.ChangePassword) // 修改用户密码
			userv1.PUT(":userID", handler.UpdateUser)                     // 更新用户信息
			userv1.DELETE(":userID", handler.DeleteUser)                  // 删除用户
			userv1.GET(":userID", handler.GetUser)                        // 查询用户详情
			userv1.GET("", handler.ListUser)                              // 查询用户列表
		}

		// Agent 相关路由
		agentv1 := v1.Group("/custom-agents", authMiddlewares...)
		{
			agentv1.GET("", handler.ListAgents)
			agentv1.GET("/:id", handler.GetAgent)
			agentv1.POST("", handler.CreateAgent)
			agentv1.PUT("/:id", handler.UpdateAgent)
			agentv1.DELETE("/:id", handler.DeleteAgent)
		}

		// 内置 Agent 路由
		builtinv1 := v1.Group("/agents/builtin", authMiddlewares...)
		{
			builtinv1.GET("", handler.ListBuiltinAgents)
		}

		// Session 相关路由
		sessionv1 := v1.Group("/sessions", authMiddlewares...)
		{
			sessionv1.GET("", handler.ListSessions)
			sessionv1.GET("/:id", handler.GetSession)
			sessionv1.POST("", handler.CreateSession)
			sessionv1.PUT("/:id", handler.UpdateSession)
			sessionv1.DELETE("/:id", handler.DeleteSession)
		}

		// Knowledge Base 相关路由
		kbv1 := v1.Group("/knowledge-bases", authMiddlewares...)
		{
			kbv1.GET("", handler.ListKnowledgeBases)
			kbv1.POST("", handler.CreateKnowledgeBase)

			// 更具体的路由（带 /knowledge 后缀）必须放在通用的 /:id 之前
			kbv1.GET("/:id/knowledge", handler.ListKnowledges)

			kbv1.GET("/:id", handler.GetKnowledgeBase)
			kbv1.GET("/:id/stats", handler.GetKnowledgeStats)
			kbv1.PUT("/:id", handler.UpdateKnowledgeBase)
			kbv1.DELETE("/:id", handler.DeleteKnowledgeBase)
		}

		// Knowledge 路由（独立于 Knowledge Base）
		knowledgev1 := v1.Group("/knowledge", authMiddlewares...)
		{
			knowledgev1.DELETE("/:id", handler.DeleteKnowledge)
		}
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
