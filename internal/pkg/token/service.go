// Package token 提供 Token 管理服务.
package token

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"

	"github.com/ashwinyue/eino-show/internal/pkg/log"
)

const (
	// TokenTypeAccess 访问令牌类型
	TokenTypeAccess = "access_token"
	// TokenTypeRefresh 刷新令牌类型
	TokenTypeRefresh = "refresh_token"

	// AccessTokenExpiration 访问令牌有效期 (24小时)
	AccessTokenExpiration = 24 * time.Hour
	// RefreshTokenExpiration 刷新令牌有效期 (7天)
	RefreshTokenExpiration = 7 * 24 * time.Hour
)

// Service Token 管理服务.
type Service struct {
	store store.IStore
	secret string
}

// New 创建 Token 管理服务.
func New(store store.IStore, secret string) *Service {
	return &Service{
		store:  store,
		secret: secret,
	}
}

// GenerateTokens 生成访问令牌和刷新令牌.
func (s *Service) GenerateTokens(ctx context.Context, userID string) (accessToken, refreshToken string, accessExpireAt, refreshExpireAt time.Time, err error) {
	now := time.Now()
	accessExpireAt = now.Add(AccessTokenExpiration)
	refreshExpireAt = now.Add(RefreshTokenExpiration)

	// 生成访问令牌
	accessToken, err = s.generateJWT(userID, accessExpireAt, TokenTypeAccess)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err = s.generateJWT(userID, refreshExpireAt, TokenTypeRefresh)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 存储 token 到数据库
	if err := s.storeTokens(ctx, userID, accessToken, refreshToken, accessExpireAt, refreshExpireAt); err != nil {
		log.Errorw("Failed to store tokens", "err", err)
		return "", "", time.Time{}, time.Time{}, err
	}

	return accessToken, refreshToken, accessExpireAt, refreshExpireAt, nil
}

// generateJWT 生成 JWT token.
func (s *Service) generateJWT(userID string, expireAt time.Time, tokenType string) (string, error) {
	// 使用与 middleware 一致的 claim key名
	claims := jwt.MapClaims{
		"x-user-id": userID, // 与 known.XUserID 一致
		"type":      tokenType,
		"nbf":       time.Now().Unix(),
		"iat":       time.Now().Unix(),
		"exp":       expireAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// storeTokens 存储 token 到数据库.
func (s *Service) storeTokens(ctx context.Context, userID, accessToken, refreshToken string, accessExpireAt, refreshExpireAt time.Time) error {
	// 存储访问令牌
	accessRecord := &model.AuthTokenM{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     accessToken,
		TokenType: TokenTypeAccess,
		ExpiresAt: accessExpireAt,
		IsRevoked: false,
	}
	if err := s.store.AuthToken().Create(ctx, accessRecord); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	// 存储刷新令牌
	refreshRecord := &model.AuthTokenM{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     refreshToken,
		TokenType: TokenTypeRefresh,
		ExpiresAt: refreshExpireAt,
		IsRevoked: false,
	}
	if err := s.store.AuthToken().Create(ctx, refreshRecord); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// ValidateToken 验证 token 并返回用户 ID.
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	// 解析 JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["x-user-id"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("invalid user ID in token")
	}

	// 检查 token 是否被撤销
	tokenRecord, err := s.store.AuthToken().GetByTokenValue(ctx, tokenString)
	if err != nil {
		return "", fmt.Errorf("token not found in database")
	}
	if tokenRecord.IsRevoked {
		return "", fmt.Errorf("token is revoked")
	}

	return userID, nil
}

// RefreshToken 使用刷新令牌获取新的访问令牌.
func (s *Service) RefreshToken(ctx context.Context, oldRefreshToken string) (newAccessToken, newRefreshToken string, accessExpireAt, refreshExpireAt time.Time, err error) {
	// 验证旧的刷新令牌
	userID, err := s.ValidateToken(ctx, oldRefreshToken)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 撤销旧的刷新令牌
	if err := s.revokeToken(ctx, oldRefreshToken); err != nil {
		log.Errorw("Failed to revoke old refresh token", "err", err)
	}

	// 生成新的 token 对
	return s.GenerateTokens(ctx, userID)
}

// RevokeToken 撤销指定的 token.
func (s *Service) RevokeToken(ctx context.Context, tokenString string) error {
	return s.revokeToken(ctx, tokenString)
}

// RevokeAllUserTokens 撤销用户的所有 token.
func (s *Service) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.store.AuthToken().RevokeByUser(ctx, userID)
}

// revokeToken 撤销指定的 token.
func (s *Service) revokeToken(ctx context.Context, tokenString string) error {
	tokenRecord, err := s.store.AuthToken().GetByTokenValue(ctx, tokenString)
	if err != nil {
		return err
	}
	return s.store.AuthToken().Revoke(ctx, tokenRecord.ID)
}

// DeleteExpiredTokens 删除过期的 token.
func (s *Service) DeleteExpiredTokens(ctx context.Context) error {
	// 直接通过 DB 删除过期的 token
	db := s.store.DB(ctx)
	return db.Where("expires_at < NOW()").Delete(&model.AuthTokenM{}).Error
}
