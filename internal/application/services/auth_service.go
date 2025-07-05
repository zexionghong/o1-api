package services

import (
	"context"
	"fmt"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)

	// Register 用户注册
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)

	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordRequest) error

	// GetUserProfile 获取用户资料
	GetUserProfile(ctx context.Context, userID int64) (*dto.GetUserProfileResponse, error)

	// ValidateUser 验证用户凭据
	ValidateUser(ctx context.Context, username, password string) (*entities.User, error)
}

// authServiceImpl 认证服务实现
type authServiceImpl struct {
	userRepo   repositories.UserRepository
	jwtService JWTService
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repositories.UserRepository, jwtService JWTService) AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Login 用户登录
func (s *authServiceImpl) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 验证用户凭据
	user, err := s.ValidateUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, fmt.Errorf("user account is not active")
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := s.jwtService.GenerateTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 构造响应
	response := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    24 * 60 * 60, // 24小时，应该从配置读取
		User: dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: "",
		},
	}

	if user.FullName != nil {
		response.User.FullName = *user.FullName
	}

	return response, nil
}

// Register 用户注册
func (s *authServiceImpl) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}
	// 如果错误不是"用户不存在"，则返回错误
	if err != nil && err != entities.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}

	// 检查邮箱是否已存在
	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}
	// 如果错误不是"用户不存在"，则返回错误
	if err != nil && err != entities.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	// 哈希密码
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户实体
	user := &entities.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: &hashedPassword,
		FullName:     &req.FullName,
		Status:       entities.UserStatusActive,
		Balance:      0.0,
	}

	// 保存用户
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 构造响应
	response := &dto.RegisterResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  req.FullName,
		Message:   "User registered successfully",
		CreatedAt: user.CreatedAt,
	}

	return response, nil
}

// RefreshToken 刷新令牌
func (s *authServiceImpl) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	// 使用JWT服务刷新令牌
	newAccessToken, newRefreshToken, err := s.jwtService.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    24 * 60 * 60, // 24小时，应该从配置读取
	}, nil
}

// ChangePassword 修改密码
func (s *authServiceImpl) ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordRequest) error {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 验证旧密码
	if user.PasswordHash == nil {
		return fmt.Errorf("user has no password set")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	// 哈希新密码
	newHashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新密码
	user.PasswordHash = &newHashedPassword
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// GetUserProfile 获取用户资料
func (s *authServiceImpl) GetUserProfile(ctx context.Context, userID int64) (*dto.GetUserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	response := &dto.GetUserProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Balance:   user.Balance,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.FullName != nil {
		response.FullName = *user.FullName
	}

	return response, nil
}

// ValidateUser 验证用户凭据
func (s *authServiceImpl) ValidateUser(ctx context.Context, username, password string) (*entities.User, error) {
	// 根据用户名获取用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 检查密码哈希是否存在
	if user.PasswordHash == nil {
		return nil, fmt.Errorf("user has no password set")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

// hashPassword 哈希密码
func (s *authServiceImpl) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
