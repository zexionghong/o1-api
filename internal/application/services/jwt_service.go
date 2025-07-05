package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/config"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService JWT服务接口
type JWTService interface {
	// GenerateTokens 生成访问令牌和刷新令牌
	GenerateTokens(ctx context.Context, user *entities.User) (accessToken, refreshToken string, err error)
	
	// ValidateAccessToken 验证访问令牌
	ValidateAccessToken(ctx context.Context, tokenString string) (*JWTClaims, error)
	
	// ValidateRefreshToken 验证刷新令牌
	ValidateRefreshToken(ctx context.Context, tokenString string) (*JWTClaims, error)
	
	// RefreshTokens 刷新令牌
	RefreshTokens(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	
	// ExtractUserFromToken 从令牌中提取用户信息
	ExtractUserFromToken(ctx context.Context, tokenString string) (*entities.User, error)
}

// jwtServiceImpl JWT服务实现
type jwtServiceImpl struct {
	config *config.JWTConfig
}

// NewJWTService 创建JWT服务
func NewJWTService(config *config.JWTConfig) JWTService {
	return &jwtServiceImpl{
		config: config,
	}
}

// GenerateTokens 生成访问令牌和刷新令牌
func (s *jwtServiceImpl) GenerateTokens(ctx context.Context, user *entities.User) (accessToken, refreshToken string, err error) {
	now := time.Now()
	
	// 生成访问令牌
	accessClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Audience:  []string{s.config.Audience},
			Subject:   fmt.Sprintf("%d", user.ID),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}
	
	// 生成刷新令牌
	refreshClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Audience:  []string{s.config.Audience},
			Subject:   fmt.Sprintf("%d", user.ID),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	
	return accessToken, refreshToken, nil
}

// ValidateAccessToken 验证访问令牌
func (s *jwtServiceImpl) ValidateAccessToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	return s.validateToken(tokenString)
}

// ValidateRefreshToken 验证刷新令牌
func (s *jwtServiceImpl) ValidateRefreshToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	return s.validateToken(tokenString)
}

// validateToken 验证令牌的通用方法
func (s *jwtServiceImpl) validateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	
	// 验证发行者和受众
	if claims.Issuer != s.config.Issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}
	
	if len(claims.Audience) == 0 || claims.Audience[0] != s.config.Audience {
		return nil, fmt.Errorf("invalid token audience")
	}
	
	return claims, nil
}

// RefreshTokens 刷新令牌
func (s *jwtServiceImpl) RefreshTokens(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	// 验证刷新令牌
	claims, err := s.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}
	
	// 创建用户对象用于生成新令牌
	user := &entities.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
	}
	
	// 生成新的令牌对
	return s.GenerateTokens(ctx, user)
}

// ExtractUserFromToken 从令牌中提取用户信息
func (s *jwtServiceImpl) ExtractUserFromToken(ctx context.Context, tokenString string) (*entities.User, error) {
	claims, err := s.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	
	return &entities.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
	}, nil
}
