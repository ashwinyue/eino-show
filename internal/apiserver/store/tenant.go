package store

import (
	"context"

	genericstore "github.com/ashwinyue/eino-show/pkg/store"
	"github.com/ashwinyue/eino-show/pkg/store/where"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// TenantStore 定义了 tenant 模块在 store 层所实现的方法.
type TenantStore interface {
	Create(ctx context.Context, obj *model.TenantM) error
	Update(ctx context.Context, obj *model.TenantM) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.TenantM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.TenantM, error)

	TenantExpansion
}

// TenantExpansion 定义了租户操作的附加方法.
// nolint: iface
type TenantExpansion interface {
	// GetByID 根据 ID 获取租户
	GetByID(ctx context.Context, id uint64) (*model.TenantM, error)
	// GetByAPIKey 根据 API Key 获取租户
	GetByAPIKey(ctx context.Context, apiKey string) (*model.TenantM, error)
	// ListAll 获取所有租户
	ListAll(ctx context.Context) ([]*model.TenantM, error)
	// Search 搜索租户
	Search(ctx context.Context, query string, page, pageSize int) (int64, []*model.TenantM, error)
}

// tenantStore 是 TenantStore 接口的实现.
type tenantStore struct {
	store *datastore
	*genericstore.Store[model.TenantM]
}

// 确保 tenantStore 实现了 TenantStore 接口.
var _ TenantStore = (*tenantStore)(nil)

// newTenantStore 创建 tenantStore 的实例.
func newTenantStore(store *datastore) *tenantStore {
	return &tenantStore{
		store: store,
		Store: genericstore.NewStore[model.TenantM](store, NewLogger()),
	}
}

// GetByID 根据 ID 获取租户.
func (s *tenantStore) GetByID(ctx context.Context, id uint64) (*model.TenantM, error) {
	var tenant model.TenantM
	err := s.store.DB(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetByAPIKey 根据 API Key 获取租户.
func (s *tenantStore) GetByAPIKey(ctx context.Context, apiKey string) (*model.TenantM, error) {
	var tenant model.TenantM
	err := s.store.DB(ctx).Where("api_key = ?", apiKey).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// ListAll 获取所有租户.
func (s *tenantStore) ListAll(ctx context.Context) ([]*model.TenantM, error) {
	var list []*model.TenantM
	err := s.store.DB(ctx).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// Search 搜索租户.
func (s *tenantStore) Search(ctx context.Context, query string, page, pageSize int) (int64, []*model.TenantM, error) {
	var list []*model.TenantM
	var total int64

	db := s.store.DB(ctx).Model(&model.TenantM{})
	if query != "" {
		db = db.Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	if page > 0 && pageSize > 0 {
		db = db.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	err := db.Order("created_at DESC").Find(&list).Error
	return total, list, err
}
