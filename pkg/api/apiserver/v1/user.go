// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

import "time"

// ===== User 请求/响应类型（对齐 WeKnora）=====

// User 用户信息
type User struct {
	ID                  string    `json:"id"`
	Username            string    `json:"username"`
	Email               string    `json:"email"`
	Avatar              string    `json:"avatar"`
	TenantID            uint64    `json:"tenant_id"`
	IsActive            bool      `json:"is_active"`
	CanAccessAllTenants bool      `json:"can_access_all_tenants"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Tenant 租户信息（对齐前端 TenantInfo）
type Tenant struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	APIKey       string `json:"api_key"`
	Status       string `json:"status"`
	Business     string `json:"business"`
	OwnerID      string `json:"owner_id"`
	StorageQuota int64  `json:"storage_quota"`
	StorageUsed  int64  `json:"storage_used"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	User *User `json:"user"`
}

// GetUserRequest 获取用户请求
type GetUserRequest struct {
	UserID string `uri:"userID" binding:"required"`
}

// GetUserResponse 获取用户响应
type GetUserResponse struct {
	User *User `json:"user"`
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ListUserRequest 用户列表请求（兼容）
type ListUserRequest = ListUsersRequest

// ListUserResponse 用户列表响应
type ListUserResponse struct {
	Users []*User `json:"users"`
	Total int64   `json:"total"`
}

// RegisterRequest 注册请求（对齐 WeKnora）
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterResponse 注册响应（对齐 WeKnora）
type RegisterResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	User    *User   `json:"user,omitempty"`
	Tenant  *Tenant `json:"tenant,omitempty"`
}

// LogoutRequest 登出请求
type LogoutRequest struct{}

// LogoutResponse 登出响应
type LogoutResponse struct {
	Success bool `json:"success"`
}

// GetCurrentUserRequest 获取当前用户请求
type GetCurrentUserRequest struct{}

// GetCurrentUserResponse 获取当前用户响应
type GetCurrentUserResponse struct {
	User   *User   `json:"user"`
	Tenant *Tenant `json:"tenant"`
}

// GetCurrentTenantRequest 获取当前租户请求
type GetCurrentTenantRequest struct{}

// GetCurrentTenantResponse 获取当前租户响应
type GetCurrentTenantResponse struct {
	Tenant *Tenant `json:"tenant"`
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	Users []*User `json:"users"`
	Total int64   `json:"total"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	UserID   string  `uri:"userID" binding:"required"`
	Email    *string `json:"email"`
	Nickname *string `json:"nickname"`
	Phone    *string `json:"phone"`
}

// UpdateUserResponse 更新用户响应
type UpdateUserResponse struct {
	User *User `json:"user"`
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	UserID string `uri:"userID" binding:"required"`
}

// DeleteUserResponse 删除用户响应
type DeleteUserResponse struct {
	Success bool `json:"success"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	UserID      string `uri:"userID" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePasswordResponse 修改密码响应
type ChangePasswordResponse struct {
	Success bool `json:"success"`
}

// LoginRequest 登录请求（对齐 WeKnora）
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse 登录响应（对齐 WeKnora）
type LoginResponse struct {
	Success      bool    `json:"success"`
	Message      string  `json:"message,omitempty"`
	User         *User   `json:"user,omitempty"`
	Tenant       *Tenant `json:"tenant,omitempty"`
	Token        string  `json:"token,omitempty"`
	RefreshToken string  `json:"refresh_token,omitempty"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
