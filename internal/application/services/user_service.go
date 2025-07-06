package services

import (
	"context"
	"fmt"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	redisInfra "ai-api-gateway/internal/infrastructure/redis"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
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
	userRepo    repositories.UserRepository
	cache       *redisInfra.CacheService
	lockService *redisInfra.DistributedLockService
}

// NewUserService 创建用户服务
func NewUserService(userRepo repositories.UserRepository, cache *redisInfra.CacheService, lockService *redisInfra.DistributedLockService) UserService {
	return &userServiceImpl{
		userRepo:    userRepo,
		cache:       cache,
		lockService: lockService,
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

	// 如果提供了密码，进行哈希处理
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := s.hashPassword(*req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = &hashedPassword
	}

	// 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 返回响应
	return (&dto.UserResponse{}).FromEntity(user), nil
}

// hashPassword 哈希密码
func (s *userServiceImpl) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// GetUser 获取用户
func (s *userServiceImpl) GetUser(ctx context.Context, id int64) (*dto.UserResponse, error) {
	// 尝试从缓存获取用户
	if s.cache != nil {
		var cachedUser entities.User
		cacheKey := fmt.Sprintf("user:%d", id)
		err := s.cache.Get(ctx, cacheKey, &cachedUser)
		if err == nil {
			// 缓存命中
			return (&dto.UserResponse{}).FromEntity(&cachedUser), nil
		}
		// 缓存未命中，继续从数据库查询
	}

	// 从数据库获取用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 将用户信息缓存5分钟
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		// 使用5分钟的缓存时间
		s.cache.Set(ctx, cacheKey, user, 5*60*1000000000) // 5分钟的纳秒数
	}

	return (&dto.UserResponse{}).FromEntity(user), nil
}

// GetUserByUsername 根据用户名获取用户
func (s *userServiceImpl) GetUserByUsername(ctx context.Context, username string) (*dto.UserResponse, error) {
	// 尝试从缓存获取用户
	if s.cache != nil {
		var cachedUser entities.User
		cacheKey := fmt.Sprintf("user:username:%s", username)
		err := s.cache.Get(ctx, cacheKey, &cachedUser)
		if err == nil {
			// 缓存命中
			return (&dto.UserResponse{}).FromEntity(&cachedUser), nil
		}
		// 缓存未命中，继续从数据库查询
	}

	// 从数据库获取用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 将用户信息缓存5分钟
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:username:%s", username)
		s.cache.Set(ctx, cacheKey, user, 5*60*1000000000) // 5分钟

		// 同时缓存用户ID索引
		userIDCacheKey := fmt.Sprintf("user:%d", user.ID)
		s.cache.Set(ctx, userIDCacheKey, user, 5*60*1000000000) // 5分钟
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

	// 使用UpdateProfile方法，只更新用户资料，不影响密码
	if err := s.userRepo.UpdateProfile(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// 清除相关缓存
	if s.cache != nil {
		// 清除用户ID缓存
		userIDCacheKey := fmt.Sprintf("user:%d", user.ID)
		s.cache.Delete(ctx, userIDCacheKey)

		// 清除用户名缓存
		usernameCacheKey := fmt.Sprintf("user:username:%s", user.Username)
		s.cache.Delete(ctx, usernameCacheKey)

		// 如果邮箱有缓存，也清除
		emailCacheKey := fmt.Sprintf("user:email:%s", user.Email)
		s.cache.Delete(ctx, emailCacheKey)
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
	// 如果有分布式锁服务，使用锁保护余额更新
	if s.lockService != nil {
		lockKey := redisInfra.GetUserLockKey(id)
		return s.updateBalanceWithLock(ctx, id, req, lockKey)
	}

	// 没有锁服务时的普通更新
	return s.updateBalanceInternal(ctx, id, req)
}

// updateBalanceWithLock 使用分布式锁更新余额
func (s *userServiceImpl) updateBalanceWithLock(ctx context.Context, id int64, req *dto.BalanceUpdateRequest, lockKey string) (*dto.UserResponse, error) {
	var result *dto.UserResponse
	var updateErr error

	// 使用分布式锁执行余额更新
	lockErr := s.lockService.WithLock(ctx, lockKey, nil, func() error {
		result, updateErr = s.updateBalanceInternal(ctx, id, req)
		return updateErr
	})

	if lockErr != nil {
		if lockErr == redisInfra.ErrLockNotObtained {
			return nil, fmt.Errorf("failed to obtain lock for user balance update, please try again")
		}
		return nil, fmt.Errorf("lock error during balance update: %w", lockErr)
	}

	return result, updateErr
}

// updateBalanceInternal 内部余额更新逻辑
func (s *userServiceImpl) updateBalanceInternal(ctx context.Context, id int64, req *dto.BalanceUpdateRequest) (*dto.UserResponse, error) {
	// 先尝试从缓存获取用户
	var user *entities.User
	var err error

	if s.cache != nil {
		user, err = s.cache.GetUser(ctx, id)
		if err != nil && err != redis.Nil {
			// 缓存错误，记录但继续从数据库获取
			// 这里可以添加日志记录
		}
	}

	// 如果缓存中没有，从数据库获取
	if user == nil {
		user, err = s.userRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
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

	// 使用专门的UpdateBalance方法，只更新余额
	if err := s.userRepo.UpdateBalance(ctx, user.ID, user.Balance); err != nil {
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	// 更新缓存
	if s.cache != nil {
		if err := s.cache.SetUser(ctx, user); err != nil {
			// 缓存更新失败，记录但不影响主流程
			// 这里可以添加日志记录
		}
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
