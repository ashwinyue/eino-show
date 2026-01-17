package store

//go:generate mockgen -destination mock_store.go -package store github.com/ashwinyue/eino-show/internal/apiserver/store IStore,UserStore,SessionStore,CustomAgentStore,KnowledgeBaseStore,KnowledgeStore,ChunkStore,ChunkExpansion,AuthTokenStore,ModelStore,MCPServiceStore,EmbeddingStore

import (
	"context"
	"sync"

	"github.com/ashwinyue/eino-show/internal/pkg/llmcontext"
	"github.com/ashwinyue/eino-show/pkg/store/where"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 是一个 Wire 的 Provider 集合，用于声明依赖注入的规则.
// 包含 NewStore 构造函数，用于生成 datastore 实例.
// wire.Bind 用于将接口 IStore 与具体实现 *datastore 绑定，
// 从而在依赖 IStore 的地方，能够自动注入 *datastore 实例.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*datastore)))

var (
	once sync.Once
	// S 全局变量，方便其它包直接调用已初始化好的 datastore 实例.
	S *datastore
)

// IStore 定义了 Store 层需要实现的方法.
type IStore interface {
	// DB 返回 Store 层的 *gorm.DB 实例，在少数场景下会被用到.
	DB(ctx context.Context, wheres ...where.Where) *gorm.DB
	TX(ctx context.Context, fn func(ctx context.Context) error) error

	User() UserStore
	Tenant() TenantStore
	Session() SessionStore
	Message() MessageStore
	CustomAgent() CustomAgentStore
	KnowledgeBase() KnowledgeBaseStore
	Knowledge() KnowledgeStore
	Chunk() ChunkStore
	AuthToken() AuthTokenStore
	Model() ModelStore
	MCPService() MCPServiceStore
	Embedding() EmbeddingStore
	KnowledgeTag() KnowledgeTagStore
	WorkflowCheckpoint() WorkflowCheckpointStore
	Summary() SummaryStore
	FAQ() FAQStore
	ContextStorage() llmcontext.ContextStorage
}

// transactionKey 用于在 context.Context 中存储事务上下文的键.
type transactionKey struct{}

// datastore 是 IStore 的具体实现.
type datastore struct {
	core *gorm.DB

	// 可以根据需要添加其他数据库实例
	// fake *gorm.DB
}

// 确保 datastore 实现了 IStore 接口.
var _ IStore = (*datastore)(nil)

// NewStore 创建一个 IStore 类型的实例.
func NewStore(db *gorm.DB) *datastore {
	// 确保 S 只被初始化一次
	once.Do(func() {
		S = &datastore{
			core: db,
		}
	})

	return S
}

// DB 根据传入的条件（wheres）对数据库实例进行筛选.
// 如果未传入任何条件，则返回上下文中的数据库实例（事务实例或核心数据库实例）.
func (store *datastore) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	db := store.core
	// 从上下文中提取事务实例
	if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
		db = tx
	}

	// 遍历所有传入的条件并逐一叠加到数据库查询对象上
	for _, whr := range wheres {
		db = whr.Where(db)
	}
	return db
}

// TX 返回一个新的事务实例.
// nolint: fatcontext
func (store *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return store.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

// User 返回一个实现了 UserStore 接口的实例.
func (store *datastore) User() UserStore {
	return newUserStore(store)
}

// Tenant 返回一个实现了 TenantStore 接口的实例.
func (store *datastore) Tenant() TenantStore {
	return newTenantStore(store)
}

// Session 返回一个实现了 SessionStore 接口的实例.
func (store *datastore) Session() SessionStore {
	return newSessionStore(store)
}

// Message 返回一个实现了 MessageStore 接口的实例.
func (store *datastore) Message() MessageStore {
	return newMessageStore(store)
}

// CustomAgent 返回一个实现了 CustomAgentStore 接口的实例.
func (store *datastore) CustomAgent() CustomAgentStore {
	return newCustomAgentStore(store)
}

// KnowledgeBase 返回一个实现了 KnowledgeBaseStore 接口的实例.
func (store *datastore) KnowledgeBase() KnowledgeBaseStore {
	return newKnowledgeBaseStore(store)
}

// Knowledge 返回一个实现了 KnowledgeStore 接口的实例.
func (store *datastore) Knowledge() KnowledgeStore {
	return newKnowledgeStore(store)
}

// Chunk 返回一个实现了 ChunkStore 接口的实例.
func (store *datastore) Chunk() ChunkStore {
	return newChunkStore(store)
}

// AuthToken 返回一个实现了 AuthTokenStore 接口.
func (store *datastore) AuthToken() AuthTokenStore {
	return newAuthTokenStore(store)
}

// Model 返回一个实现了 ModelStore 接口的实例.
func (store *datastore) Model() ModelStore {
	return newModelStore(store)
}

// MCPService 返回一个实现了 MCPServiceStore 接口的实例.
func (store *datastore) MCPService() MCPServiceStore {
	return newMCPServiceStore(store)
}

// Embedding 返回一个实现了 EmbeddingStore 接口的实例.
func (store *datastore) Embedding() EmbeddingStore {
	return newEmbeddingStore(store)
}

// KnowledgeTag 返回一个实现了 KnowledgeTagStore 接口的实例.
func (store *datastore) KnowledgeTag() KnowledgeTagStore {
	return newKnowledgeTagStore(store)
}

// WorkflowCheckpoint 返回一个实现了 WorkflowCheckpointStore 接口的实例.
func (store *datastore) WorkflowCheckpoint() WorkflowCheckpointStore {
	return newWorkflowCheckpointStore(store)
}

// Summary 返回一个实现了 SummaryStore 接口的实例.
func (store *datastore) Summary() SummaryStore {
	return newSummaryStore(store)
}

// FAQ 返回一个实现了 FAQStore 接口的实例.
func (store *datastore) FAQ() FAQStore {
	return newFAQStore(store)
}

// ContextStorage 返回一个实现了 ContextStorage 接口的实例.
func (store *datastore) ContextStorage() llmcontext.ContextStorage {
	return newDBContextStorage(store)
}
