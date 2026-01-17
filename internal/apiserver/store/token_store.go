package store

import (
	"context"

	"github.com/ashwinyue/eino-show/pkg/store/where"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
)

// AuthTokenStore 定义了 Token 存储的接口.
type AuthTokenStore interface {
	// Create 创建一个新 token 记录
	Create(ctx context.Context, token *model.AuthTokenM) error
	// Get 根据条件获取 token
	Get(ctx context.Context, whs ...where.Where) (*model.AuthTokenM, error)
	// GetByTokenValue 根据 token 值获取记录
	GetByTokenValue(ctx context.Context, tokenValue string) (*model.AuthTokenM, error)
	// Revoke 撤销指定的 token
	Revoke(ctx context.Context, tokenID string) error
	// RevokeByUser 撤销用户的所有 token
	RevokeByUser(ctx context.Context, userID string) error
	// Delete 删除 token 记录
	Delete(ctx context.Context, whs ...where.Where) error
}

// authTokenStore 是 AuthTokenStore 接口的具体实现.
type authTokenStore struct {
	store *datastore
}

// 确保 authTokenStore 实现了 AuthTokenStore 接口.
var _ AuthTokenStore = (*authTokenStore)(nil)

// newAuthTokenStore 创建一个新的 authTokenStore 实例.
func newAuthTokenStore(store *datastore) *authTokenStore {
	return &authTokenStore{store: store}
}

// Create 创建一个新 token 记录.
func (s *authTokenStore) Create(ctx context.Context, token *model.AuthTokenM) error {
	return s.store.DB(ctx).Create(token).Error
}

// Get 根据条件获取 token.
func (s *authTokenStore) Get(ctx context.Context, whs ...where.Where) (*model.AuthTokenM, error) {
	var token model.AuthTokenM
	db := s.store.DB(ctx)
	for _, wh := range whs {
		db = wh.Where(db)
	}
	err := db.First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByTokenValue 根据 token 值获取记录.
func (s *authTokenStore) GetByTokenValue(ctx context.Context, tokenValue string) (*model.AuthTokenM, error) {
	var token model.AuthTokenM
	err := s.store.DB(ctx).Where("token = ?", tokenValue).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Revoke 撤销指定的 token.
func (s *authTokenStore) Revoke(ctx context.Context, tokenID string) error {
	return s.store.DB(ctx).Model(&model.AuthTokenM{}).
		Where("id = ?", tokenID).
		Update("is_revoked", true).Error
}

// RevokeByUser 撤销用户的所有 token.
func (s *authTokenStore) RevokeByUser(ctx context.Context, userID string) error {
	return s.store.DB(ctx).Model(&model.AuthTokenM{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Update("is_revoked", true).Error
}

// Delete 删除 token 记录.
func (s *authTokenStore) Delete(ctx context.Context, whs ...where.Where) error {
	db := s.store.DB(ctx)
	for _, wh := range whs {
		db = wh.Where(db)
	}
	return db.Delete(&model.AuthTokenM{}).Error
}

// Token 检查并删除过期的 token.
func (s *authTokenStore) DeleteExpired(ctx context.Context) error {
	return s.store.DB(ctx).Where("expires_at < NOW()").Delete(&model.AuthTokenM{}).Error
}
