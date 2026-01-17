// Package user 提供用户业务逻辑（对齐 WeKnora）.
package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/ashwinyue/eino-show/pkg/store/where"
	"github.com/ashwinyue/eino-show/pkg/token"
)

// UserBiz 用户业务接口.
type UserBiz interface {
	// Login 用户登录
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)
	// Register 用户注册
	Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error)
	// Logout 用户登出
	Logout(ctx context.Context, tokenString string) error
	// GetCurrentUser 获取当前用户
	GetCurrentUser(ctx context.Context) (*v1.GetCurrentUserResponse, error)
	// GetCurrentTenant 获取当前租户
	GetCurrentTenant(ctx context.Context) (*v1.GetCurrentTenantResponse, error)
	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error
	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, refreshToken string) (*v1.RefreshTokenResponse, error)
}

// userBiz 用户业务实现.
type userBiz struct {
	store store.IStore
}

var _ UserBiz = (*userBiz)(nil)

// New 创建用户业务实例.
func New(store store.IStore) UserBiz {
	return &userBiz{store: store}
}

// getTenantID 安全获取 TenantID.
func getTenantID(tenantID *int32) uint64 {
	if tenantID == nil {
		return 0
	}
	return uint64(*tenantID)
}

// Login 用户登录（对齐 WeKnora）.
func (b *userBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// 通过邮箱查找用户
	userM, err := b.store.User().Get(ctx, where.F("email", req.Email))
	if err != nil {
		return &v1.LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// 检查用户是否激活
	if !userM.IsActive {
		return &v1.LoginResponse{
			Success: false,
			Message: "Account is disabled",
		}, nil
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(userM.PasswordHash), []byte(req.Password)); err != nil {
		return &v1.LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// 使用 pkg/token 生成 JWT Token
	accessToken, _, err := token.Sign(userM.ID)
	if err != nil {
		return &v1.LoginResponse{
			Success: false,
			Message: "Login failed",
		}, nil
	}

	// 生成 refresh token（使用不同的过期时间）
	refreshToken, _, err := token.Sign(userM.ID)
	if err != nil {
		return &v1.LoginResponse{
			Success: false,
			Message: "Login failed",
		}, nil
	}

	// 获取租户信息
	var tenant *v1.Tenant
	if userM.TenantID != nil {
		tenantM, err := b.store.Tenant().GetByID(ctx, uint64(*userM.TenantID))
		if err == nil && tenantM != nil {
			tenant = &v1.Tenant{
				ID:           uint64(tenantM.ID),
				Name:         tenantM.Name,
				Description:  stringValue(tenantM.Description),
				APIKey:       tenantM.APIKey,
				Status:       stringValue(tenantM.Status),
				Business:     tenantM.Business,
				OwnerID:      userM.ID,
				StorageQuota: tenantM.StorageQuota,
				StorageUsed:  tenantM.StorageUsed,
				CreatedAt:    timeString(tenantM.CreatedAt),
				UpdatedAt:    timeString(tenantM.UpdatedAt),
			}
		}
	}

	return &v1.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: &v1.User{
			ID:                  userM.ID,
			Username:            userM.Username,
			Email:               userM.Email,
			TenantID:            getTenantID(userM.TenantID),
			IsActive:            userM.IsActive,
			CanAccessAllTenants: userM.CanAccessAllTenants,
		},
		Tenant: tenant,
	}, nil
}

// stringValue 安全获取 *string 的值.
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// timeString 将 *time.Time 转换为 ISO8601 字符串.
func timeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// Register 用户注册（对齐 WeKnora）.
func (b *userBiz) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	// 检查用户名是否已存在
	existingUser, _ := b.store.User().Get(ctx, where.F("username", req.Username))
	if existingUser != nil {
		return &v1.RegisterResponse{
			Success: false,
			Message: "用户名已存在",
		}, nil
	}

	// 检查邮箱是否已存在
	existingEmail, _ := b.store.User().Get(ctx, where.F("email", req.Email))
	if existingEmail != nil {
		return &v1.RegisterResponse{
			Success: false,
			Message: "邮箱已存在",
		}, nil
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &v1.RegisterResponse{
			Success: false,
			Message: "密码加密失败",
		}, nil
	}

	// 创建用户
	now := time.Now()
	userM := &model.UserM{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Email:        req.Email,
		IsActive:     true,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if err := b.store.User().Create(ctx, userM); err != nil {
		return &v1.RegisterResponse{
			Success: false,
			Message: "创建用户失败",
		}, nil
	}

	return &v1.RegisterResponse{
		Success: true,
		Message: "注册成功",
		User: &v1.User{
			ID:       userM.ID,
			Username: userM.Username,
			Email:    userM.Email,
			IsActive: userM.IsActive,
		},
	}, nil
}

// GetCurrentUser 获取当前用户.
func (b *userBiz) GetCurrentUser(ctx context.Context) (*v1.GetCurrentUserResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == "" {
		return nil, errors.New("用户未登录")
	}

	userM, err := b.store.User().Get(ctx, where.F("id", userID))
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	tenantID := contextx.TenantID(ctx)
	tenantM, err := b.store.Tenant().Get(ctx, where.F("id", tenantID))
	if err != nil {
		return nil, errors.New("租户不存在")
	}

	return &v1.GetCurrentUserResponse{
		User:   toUserResponse(userM),
		Tenant: toTenantResponse(tenantM),
	}, nil
}

// GetCurrentTenant 获取当前租户.
func (b *userBiz) GetCurrentTenant(ctx context.Context) (*v1.GetCurrentTenantResponse, error) {
	tenantID := contextx.TenantID(ctx)
	if tenantID == 0 {
		return nil, errors.New("租户未设置")
	}

	tenantM, err := b.store.Tenant().Get(ctx, where.F("id", tenantID))
	if err != nil {
		return nil, errors.New("租户不存在")
	}

	return &v1.GetCurrentTenantResponse{
		Tenant: toTenantResponse(tenantM),
	}, nil
}

// ChangePassword 修改密码.
func (b *userBiz) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	userM, err := b.store.User().Get(ctx, where.F("id", userID))
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(userM.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	userM.PasswordHash = string(hashedPassword)
	now := time.Now()
	userM.UpdatedAt = &now

	return b.store.User().Update(ctx, userM)
}

// Logout 用户登出（对齐 WeKnora）.
func (b *userBiz) Logout(ctx context.Context, tokenString string) error {
	// pkg/token 不支持 token 撤销，客户端需要删除本地 token
	_ = tokenString
	return nil
}

// RefreshToken 刷新令牌（对齐 WeKnora）.
func (b *userBiz) RefreshToken(ctx context.Context, refreshToken string) (*v1.RefreshTokenResponse, error) {
	// 验证旧的 refresh token
	userID, err := token.ParseIdentity(refreshToken, "")
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 生成新的 token
	newAccessToken, expireAt, err := token.Sign(userID)
	if err != nil {
		return nil, errors.New("failed to generate new token")
	}

	newRefreshToken, _, err := token.Sign(userID)
	if err != nil {
		return nil, errors.New("failed to generate new refresh token")
	}

	return &v1.RefreshTokenResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(time.Until(expireAt).Seconds()),
	}, nil
}

// toUserResponse 将 model.UserM 转换为 v1.User
func toUserResponse(u *model.UserM) *v1.User {
	if u == nil {
		return nil
	}
	resp := &v1.User{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		IsActive: u.IsActive,
	}
	if u.TenantID != nil {
		resp.TenantID = uint64(*u.TenantID)
	}
	if u.CreatedAt != nil {
		resp.CreatedAt = *u.CreatedAt
	}
	if u.UpdatedAt != nil {
		resp.UpdatedAt = *u.UpdatedAt
	}
	return resp
}

// toTenantResponse 将 model.TenantM 转换为 v1.Tenant
func toTenantResponse(t *model.TenantM) *v1.Tenant {
	if t == nil {
		return nil
	}
	resp := &v1.Tenant{
		ID:           uint64(t.ID),
		Name:         t.Name,
		APIKey:       t.APIKey,
		Business:     t.Business,
		StorageQuota: t.StorageQuota,
		StorageUsed:  t.StorageUsed,
	}
	if t.Description != nil {
		resp.Description = *t.Description
	}
	if t.Status != nil {
		resp.Status = *t.Status
	}
	if t.CreatedAt != nil {
		resp.CreatedAt = t.CreatedAt.Format(time.RFC3339)
	}
	if t.UpdatedAt != nil {
		resp.UpdatedAt = t.UpdatedAt.Format(time.RFC3339)
	}
	return resp
}
