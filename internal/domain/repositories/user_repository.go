package repositories

import (
	"ai-api-gateway/internal/domain/entities"
	"context"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *entities.User) error

	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id int64) (*entities.User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*entities.User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*entities.User, error)

	// Update 更新用户
	Update(ctx context.Context, user *entities.User) error

	// UpdateProfile 更新用户资料（不包括密码）
	UpdateProfile(ctx context.Context, user *entities.User) error

	// UpdateBalance 更新用户余额
	UpdateBalance(ctx context.Context, userID int64, balance float64) error

	// Delete 删除用户（软删除）
	Delete(ctx context.Context, id int64) error

	// List 获取用户列表
	List(ctx context.Context, offset, limit int) ([]*entities.User, error)

	// Count 获取用户总数
	Count(ctx context.Context) (int64, error)

	// GetActiveUsers 获取活跃用户列表
	GetActiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error)
}
