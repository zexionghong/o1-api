package repositories

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"

	"gorm.io/gorm"
)

// userRepositoryGorm GORM用户仓储实现
type userRepositoryGorm struct {
	db *gorm.DB
}

// NewUserRepositoryGorm 创建GORM用户仓储
func NewUserRepositoryGorm(db *gorm.DB) repositories.UserRepository {
	return &userRepositoryGorm{
		db: db,
	}
}

// Create 创建用户
func (r *userRepositoryGorm) Create(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepositoryGorm) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepositoryGorm) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepositoryGorm) Update(ctx context.Context, user *entities.User) error {
	user.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// Delete 删除用户（软删除）
func (r *userRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Model(&entities.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     entities.UserStatusDeleted,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// List 获取用户列表
func (r *userRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User
	if err := r.db.WithContext(ctx).
		Select("id, username, email, full_name, status, balance, created_at, updated_at").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// Count 获取用户总数
func (r *userRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.User{}).
		Where("status != ?", entities.UserStatusDeleted).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// GetActiveUsers 获取活跃用户列表
func (r *userRepositoryGorm) GetActiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User
	if err := r.db.WithContext(ctx).
		Select("id, username, email, full_name, status, balance, created_at, updated_at").
		Where("status = ?", entities.UserStatusActive).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	return users, nil
}

// UpdateBalance 更新用户余额
func (r *userRepositoryGorm) UpdateBalance(ctx context.Context, userID int64, newBalance float64) error {
	result := r.db.WithContext(ctx).Model(&entities.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"balance":    newBalance,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update user balance: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// IncrementBalance 增加用户余额
func (r *userRepositoryGorm) IncrementBalance(ctx context.Context, userID int64, amount float64) error {
	result := r.db.WithContext(ctx).Model(&entities.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"balance":    gorm.Expr("balance + ?", amount),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to increment user balance: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// DecrementBalance 减少用户余额
func (r *userRepositoryGorm) DecrementBalance(ctx context.Context, userID int64, amount float64) error {
	result := r.db.WithContext(ctx).Model(&entities.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"balance":    gorm.Expr("balance - ?", amount),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to decrement user balance: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// GetUsersByStatus 根据状态获取用户列表
func (r *userRepositoryGorm) GetUsersByStatus(ctx context.Context, status entities.UserStatus, offset, limit int) ([]*entities.User, error) {
	var users []*entities.User
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by status: %w", err)
	}
	return users, nil
}

// UpdateStatus 更新用户状态
func (r *userRepositoryGorm) UpdateStatus(ctx context.Context, userID int64, status entities.UserStatus) error {
	result := r.db.WithContext(ctx).Model(&entities.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update user status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// UpdateProfile 更新用户资料（不包括密码）
func (r *userRepositoryGorm) UpdateProfile(ctx context.Context, user *entities.User) error {
	user.UpdatedAt = time.Now()

	// 只更新指定字段，不包括密码
	result := r.db.WithContext(ctx).Model(user).
		Select("username", "email", "full_name", "status", "updated_at").
		Updates(user)

	if result.Error != nil {
		return fmt.Errorf("failed to update user profile: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}
