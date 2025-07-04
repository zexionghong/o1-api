package services

import (
	"context"
	"fmt"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/domain/repositories"
)

// UserService 用户服务接口
type UserService interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)

	// GetUser 获取用户
	GetUser(ctx context.Context, id int64) (*dto.UserResponse, error)

	// GetUserByUsername 根据用户名获取用户
	GetUserByUsername(ctx context.Context, username string) (*dto.UserResponse, error)

	// GetUserByEmail 根据邮箱获取用户
	GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error)

	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserResponse, error)

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, id int64) error

	// ListUsers 获取用户列表
	ListUsers(ctx context.Context, pagination *dto.PaginationRequest) (*dto.UserListResponse, error)

	// UpdateBalance 更新用户余额
	UpdateBalance(ctx context.Context, id int64, req *dto.BalanceUpdateRequest) (*dto.UserResponse, error)

	// GetActiveUsers 获取活跃用户列表
	GetActiveUsers(ctx context.Context, pagination *dto.PaginationRequest) (*dto.UserListResponse, error)
}

// userServiceImpl 用户服务实现
type userServiceImpl struct {
	userRepo repositories.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户
func (s *userServiceImpl) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, fmt.Errorf("email already exists")
	}

	// 创建用户实体
	user := req.ToEntity()

	// 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 返回响应
	return (&dto.UserResponse{}).FromEntity(user), nil
}

// GetUser 获取用户
func (s *userServiceImpl) GetUser(ctx context.Context, id int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return (&dto.UserResponse{}).FromEntity(user), nil
}

// GetUserByUsername 根据用户名获取用户
func (s *userServiceImpl) GetUserByUsername(ctx context.Context, username string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return (&dto.UserResponse{}).FromEntity(user), nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *userServiceImpl) GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return (&dto.UserResponse{}).FromEntity(user), nil
}

// UpdateUser 更新用户
func (s *userServiceImpl) UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 获取现有用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Username != nil {
		// 检查用户名是否已被其他用户使用
		if existingUser, err := s.userRepo.GetByUsername(ctx, *req.Username); err == nil && existingUser.ID != id {
			return nil, fmt.Errorf("username already exists")
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// 检查邮箱是否已被其他用户使用
		if existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email); err == nil && existingUser.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = *req.Email
	}

	if req.FullName != nil {
		user.FullName = req.FullName
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return (&dto.UserResponse{}).FromEntity(user), nil
}

// DeleteUser 删除用户
func (s *userServiceImpl) DeleteUser(ctx context.Context, id int64) error {
	return s.userRepo.Delete(ctx, id)
}

// ListUsers 获取用户列表
func (s *userServiceImpl) ListUsers(ctx context.Context, pagination *dto.PaginationRequest) (*dto.UserListResponse, error) {
	pagination.SetDefaults()

	// 获取用户列表
	users, err := s.userRepo.List(ctx, pagination.GetOffset(), pagination.GetLimit())
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 获取总数
	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 构造响应
	response := &dto.UserListResponse{
		Users:    dto.FromUserEntities(users),
		Total:    total,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}

	// 计算总页数
	paginationResp := &dto.PaginationResponse{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    total,
	}
	paginationResp.CalculateTotalPages()
	response.TotalPages = paginationResp.TotalPages

	return response, nil
}

// UpdateBalance 更新用户余额
func (s *userServiceImpl) UpdateBalance(ctx context.Context, id int64, req *dto.BalanceUpdateRequest) (*dto.UserResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新余额
	switch req.Operation {
	case "add":
		if err := user.AddBalance(req.Amount); err != nil {
			return nil, err
		}
	case "deduct":
		if err := user.DeductBalance(req.Amount); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid operation: %s", req.Operation)
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	return (&dto.UserResponse{}).FromEntity(user), nil
}

// GetActiveUsers 获取活跃用户列表
func (s *userServiceImpl) GetActiveUsers(ctx context.Context, pagination *dto.PaginationRequest) (*dto.UserListResponse, error) {
	pagination.SetDefaults()

	// 获取活跃用户列表
	users, err := s.userRepo.GetActiveUsers(ctx, pagination.GetOffset(), pagination.GetLimit())
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	// 获取总数（这里简化处理，实际应该有专门的计数方法）
	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 构造响应
	response := &dto.UserListResponse{
		Users:    dto.FromUserEntities(users),
		Total:    total,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}

	// 计算总页数
	paginationResp := &dto.PaginationResponse{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    total,
	}
	paginationResp.CalculateTotalPages()
	response.TotalPages = paginationResp.TotalPages

	return response, nil
}
