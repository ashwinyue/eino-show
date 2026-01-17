package biz

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/ashwinyue/eino-show/internal/apiserver/biz/agent"
	"github.com/ashwinyue/eino-show/internal/apiserver/biz/faq"
	"github.com/ashwinyue/eino-show/internal/apiserver/biz/knowledge"
	"github.com/ashwinyue/eino-show/internal/apiserver/biz/mcp"
	"github.com/ashwinyue/eino-show/internal/apiserver/biz/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/biz/session"
	"github.com/ashwinyue/eino-show/internal/apiserver/biz/tenant"
	"github.com/ashwinyue/eino-show/internal/apiserver/biz/user"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	agentmodel "github.com/ashwinyue/eino-show/internal/pkg/agent/model"
)

// ProviderSet 是一个 Wire 的 Provider 集合.
var ProviderSet = wire.NewSet(NewBiz)

// IBiz 定义了业务层需要实现的方法.
type IBiz interface {
	User() user.UserBiz
	Tenant() tenant.TenantBiz
	Session() session.SessionBiz
	Agent() agent.AgentBiz
	Knowledge() knowledge.KnowledgeBiz
	MCP() mcp.MCPBiz
	Model() model.ModelBiz
	FAQ() faq.FAQBiz
}

type biz struct {
	store       store.IStore
	redisClient *redis.Client

	// 懒加载的 Session Biz (增强模式)
	sessionBiz     session.SessionBiz
	sessionBizOnce sync.Once
}

var _ IBiz = (*biz)(nil)

// NewBiz 创建业务层实例 (redisClient 可为 nil).
func NewBiz(store store.IStore, redisClient *redis.Client) IBiz {
	return &biz{store: store, redisClient: redisClient}
}

func (b *biz) User() user.UserBiz       { return user.New(b.store) }
func (b *biz) Tenant() tenant.TenantBiz { return tenant.New(b.store) }

func (b *biz) Session() session.SessionBiz {
	b.sessionBizOnce.Do(func() {
		ctx := context.Background()

		// 优先从数据库获取默认 chat 模型配置
		chatModelConfig := b.getDefaultChatModelConfig(ctx)

		// 尝试使用增强配置创建 SessionBiz
		cfg := &session.SessionConfig{
			Store:           b.store,
			RedisClient:     b.redisClient,
			ChatModelConfig: chatModelConfig,
			EnableEnhanced:  true, // 启用增强模式
		}
		sessionBiz, err := session.NewSessionBizWithConfig(ctx, cfg)
		if err != nil {
			// 回退到基础模式
			b.sessionBiz = session.NewWithRedis(b.store, b.redisClient)
		} else {
			b.sessionBiz = sessionBiz
		}
	})
	return b.sessionBiz
}

// getDefaultChatModelConfig 获取默认 ChatModel 配置
// 优先从数据库获取，失败则回退到环境变量
func (b *biz) getDefaultChatModelConfig(ctx context.Context) *agentmodel.Config {
	// 尝试从数据库获取默认 chat 模型
	dbModel, err := b.store.Model().GetDefault(ctx, "chat")
	if err == nil && dbModel != nil {
		// 解析 Parameters JSON
		var params agentmodel.ModelParameters
		if dbModel.Parameters != "" && dbModel.Parameters != "{}" {
			_ = json.Unmarshal([]byte(dbModel.Parameters), &params)
		}

		// 从数据库模型构建配置
		return &agentmodel.Config{
			Provider: dbModel.Source,
			Model:    dbModel.Name,
			APIKey:   params.APIKey,
			BaseURL:  params.BaseURL,
		}
	}

	// 回退到环境变量配置
	return agentmodel.DefaultConfig()
}

func (b *biz) Agent() agent.AgentBiz             { return agent.New(b.store) }
func (b *biz) Knowledge() knowledge.KnowledgeBiz { return knowledge.New(b.store) }
func (b *biz) MCP() mcp.MCPBiz                   { return mcp.New(b.store) }
func (b *biz) Model() model.ModelBiz             { return model.New(b.store) }
func (b *biz) FAQ() faq.FAQBiz                   { return faq.New(b.store) }
