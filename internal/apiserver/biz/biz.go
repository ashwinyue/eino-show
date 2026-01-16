package biz

//go:generate mockgen -destination mock_biz.go -package biz github.com/ashwinyue/eino-show/internal/apiserver/biz IBiz

import (
	"github.com/google/wire"
	"github.com/onexstack/onexstack/pkg/authz"

	agentv1 "github.com/ashwinyue/eino-show/internal/apiserver/biz/v1/agent"
	knowledgev1 "github.com/ashwinyue/eino-show/internal/apiserver/biz/v1/knowledge"
	sessionv1 "github.com/ashwinyue/eino-show/internal/apiserver/biz/v1/session"
	userv1 "github.com/ashwinyue/eino-show/internal/apiserver/biz/v1/user"

	"github.com/ashwinyue/eino-show/internal/apiserver/store"
)

// ProviderSet 是一个 Wire 的 Provider 集合，用于声明依赖注入的规则.
// 包含 NewBiz 构造函数，用于生成 biz 实例.
// wire.Bind 用于将接口 IBiz 与具体实现 *biz 绑定，
// 这样依赖 IBiz 的地方会自动注入 *biz 实例.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz 定义了业务层需要实现的方法.
type IBiz interface {
	// UserV1 获取用户业务接口.
	UserV1() userv1.UserBiz

	// SessionV1 获取会话业务接口.
	SessionV1() sessionv1.SessionBiz
	// AgentV1 获取 Agent 业务接口.
	AgentV1() agentv1.AgentBiz
	// KnowledgeV1 获取知识库业务接口.
	KnowledgeV1() knowledgev1.KnowledgeBiz
}

// biz 是 IBiz 的一个具体实现.
type biz struct {
	store store.IStore
	authz *authz.Authz
}

// 确保 biz 实现了 IBiz 接口.
var _ IBiz = (*biz)(nil)

// NewBiz 创建一个 IBiz 类型的实例.
func NewBiz(store store.IStore, authz *authz.Authz) *biz {
	return &biz{store: store, authz: authz}
}

// UserV1 返回一个实现了 UserBiz 接口的实例.
func (b *biz) UserV1() userv1.UserBiz {
	return userv1.New(b.store, b.authz)
}

// SessionV1 返回一个实现了 SessionBiz 接口的实例.
func (b *biz) SessionV1() sessionv1.SessionBiz {
	return sessionv1.New(b.store)
}

// AgentV1 返回一个实现了 AgentBiz 接口的实例.
func (b *biz) AgentV1() agentv1.AgentBiz {
	return agentv1.New(b.store)
}

// KnowledgeV1 返回一个实现了 KnowledgeBiz 接口的实例.
func (b *biz) KnowledgeV1() knowledgev1.KnowledgeBiz {
	return knowledgev1.New(b.store)
}
