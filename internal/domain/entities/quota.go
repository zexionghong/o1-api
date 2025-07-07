package entities

import (
	"fmt"
	"time"
)

// QuotaType 配额类型枚举
type QuotaType string

const (
	QuotaTypeRequests QuotaType = "requests"
	QuotaTypeTokens   QuotaType = "tokens"
	QuotaTypeCost     QuotaType = "cost"
)

// QuotaPeriod 配额周期枚举
type QuotaPeriod string

const (
	QuotaPeriodMinute QuotaPeriod = "minute"
	QuotaPeriodHour   QuotaPeriod = "hour"
	QuotaPeriodDay    QuotaPeriod = "day"
	QuotaPeriodMonth  QuotaPeriod = "month"
)

// QuotaStatus 配额状态枚举
type QuotaStatus string

const (
	QuotaStatusActive   QuotaStatus = "active"
	QuotaStatusInactive QuotaStatus = "inactive"
)

// Quota 配额设置实体
type Quota struct {
	ID         int64        `json:"id" db:"id"`
	APIKeyID   int64        `json:"api_key_id" db:"api_key_id"`
	QuotaType  QuotaType    `json:"quota_type" db:"quota_type"`
	Period     *QuotaPeriod `json:"period,omitempty" db:"period"` // NULL表示总限额
	LimitValue float64      `json:"limit_value" db:"limit_value"`
	ResetTime  *string      `json:"reset_time,omitempty" db:"reset_time"` // HH:MM格式
	Status     QuotaStatus  `json:"status" db:"status"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
}

// IsActive 检查配额是否处于活跃状态
func (q *Quota) IsActive() bool {
	return q.Status == QuotaStatusActive
}

// IsTotalQuota 检查是否为总限额（不分周期）
func (q *Quota) IsTotalQuota() bool {
	return q.Period == nil
}

// IsPeriodQuota 检查是否为周期限额
func (q *Quota) IsPeriodQuota() bool {
	return q.Period != nil
}

// GetPeriodDuration 获取周期时长
func (q *Quota) GetPeriodDuration() time.Duration {
	if q.Period == nil {
		return 0 // 总限额没有周期
	}
	switch *q.Period {
	case QuotaPeriodMinute:
		return time.Minute
	case QuotaPeriodHour:
		return time.Hour
	case QuotaPeriodDay:
		return 24 * time.Hour
	case QuotaPeriodMonth:
		return 30 * 24 * time.Hour // 简化为30天
	default:
		return time.Hour
	}
}

// GetPeriodStart 获取指定时间的周期开始时间
func (q *Quota) GetPeriodStart(at time.Time) time.Time {
	if q.Period == nil {
		return time.Time{} // 总限额没有周期开始时间
	}
	switch *q.Period {
	case QuotaPeriodMinute:
		return time.Date(at.Year(), at.Month(), at.Day(), at.Hour(), at.Minute(), 0, 0, at.Location())
	case QuotaPeriodHour:
		return time.Date(at.Year(), at.Month(), at.Day(), at.Hour(), 0, 0, 0, at.Location())
	case QuotaPeriodDay:
		resetHour, resetMinute := q.getResetTime()
		start := time.Date(at.Year(), at.Month(), at.Day(), resetHour, resetMinute, 0, 0, at.Location())
		if at.Before(start) {
			start = start.AddDate(0, 0, -1)
		}
		return start
	case QuotaPeriodMonth:
		resetHour, resetMinute := q.getResetTime()
		start := time.Date(at.Year(), at.Month(), 1, resetHour, resetMinute, 0, 0, at.Location())
		if at.Before(start) {
			start = start.AddDate(0, -1, 0)
		}
		return start
	default:
		return time.Date(at.Year(), at.Month(), at.Day(), at.Hour(), 0, 0, 0, at.Location())
	}
}

// GetPeriodEnd 获取指定时间的周期结束时间
func (q *Quota) GetPeriodEnd(at time.Time) time.Time {
	if q.Period == nil {
		return time.Time{} // 总限额没有周期结束时间
	}
	start := q.GetPeriodStart(at)
	switch *q.Period {
	case QuotaPeriodMinute:
		return start.Add(time.Minute)
	case QuotaPeriodHour:
		return start.Add(time.Hour)
	case QuotaPeriodDay:
		return start.AddDate(0, 0, 1)
	case QuotaPeriodMonth:
		return start.AddDate(0, 1, 0)
	default:
		return start.Add(time.Hour)
	}
}

// getResetTime 解析重置时间
func (q *Quota) getResetTime() (hour, minute int) {
	if q.ResetTime == nil {
		return 0, 0 // 默认午夜重置
	}

	// 简单解析HH:MM格式
	var h, m int
	if n, _ := fmt.Sscanf(*q.ResetTime, "%d:%d", &h, &m); n == 2 {
		if h >= 0 && h <= 23 && m >= 0 && m <= 59 {
			return h, m
		}
	}
	return 0, 0
}

// QuotaUsage 配额使用情况实体
type QuotaUsage struct {
	ID          int64      `json:"id" db:"id"`
	APIKeyID    int64      `json:"api_key_id" db:"api_key_id"`
	QuotaID     int64      `json:"quota_id" db:"quota_id"`
	PeriodStart *time.Time `json:"period_start,omitempty" db:"period_start"` // 总限额时为NULL
	PeriodEnd   *time.Time `json:"period_end,omitempty" db:"period_end"`     // 总限额时为NULL
	UsedValue   float64    `json:"used_value" db:"used_value"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// IsWithinPeriod 检查指定时间是否在周期内
func (qu *QuotaUsage) IsWithinPeriod(at time.Time) bool {
	if qu.PeriodStart == nil || qu.PeriodEnd == nil {
		return true // 总限额没有周期限制
	}
	return !at.Before(*qu.PeriodStart) && at.Before(*qu.PeriodEnd)
}

// AddUsage 增加使用量
func (qu *QuotaUsage) AddUsage(value float64) {
	qu.UsedValue += value
	qu.UpdatedAt = time.Now()
}

// GetRemainingQuota 获取剩余配额
func (qu *QuotaUsage) GetRemainingQuota(limit float64) float64 {
	remaining := limit - qu.UsedValue
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsExceeded 检查是否超出配额
func (qu *QuotaUsage) IsExceeded(limit float64) bool {
	return qu.UsedValue >= limit
}

// GetUsagePercentage 获取使用百分比
func (qu *QuotaUsage) GetUsagePercentage(limit float64) float64 {
	if limit <= 0 {
		return 0
	}
	percentage := (qu.UsedValue / limit) * 100
	if percentage > 100 {
		return 100
	}
	return percentage
}
