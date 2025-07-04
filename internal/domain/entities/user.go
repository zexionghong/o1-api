package entities

import (
	"time"
)

// UserStatus 用户状态枚举
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User 用户实体
type User struct {
	ID        int64      `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Email     string     `json:"email" db:"email"`
	FullName  *string    `json:"full_name,omitempty" db:"full_name"`
	Status    UserStatus `json:"status" db:"status"`
	Balance   float64    `json:"balance" db:"balance"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// IsActive 检查用户是否处于活跃状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanMakeRequest 检查用户是否可以发起请求
func (u *User) CanMakeRequest() bool {
	return u.IsActive() && u.Balance >= 0
}

// DeductBalance 扣减用户余额
func (u *User) DeductBalance(amount float64) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	if u.Balance < amount {
		return ErrInsufficientBalance
	}
	u.Balance -= amount
	u.UpdatedAt = time.Now()
	return nil
}

// AddBalance 增加用户余额
func (u *User) AddBalance(amount float64) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	u.Balance += amount
	u.UpdatedAt = time.Now()
	return nil
}
