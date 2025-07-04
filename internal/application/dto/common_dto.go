package dto

import (
	"time"
)

// Response 通用响应结构
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `json:"page" validate:"min=1" form:"page"`
	PageSize int `json:"page_size" validate:"min=1,max=100" form:"page_size"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse 创建成功响应
func SuccessResponse(data interface{}, message string) *Response {
	return &Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ErrorResponse 创建错误响应
func ErrorResponse(code, message string, details map[string]interface{}) *Response {
	return &Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

// GetOffset 计算偏移量
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *PaginationRequest) GetLimit() int {
	return p.PageSize
}

// CalculateTotalPages 计算总页数
func (p *PaginationResponse) CalculateTotalPages() {
	if p.PageSize > 0 {
		p.TotalPages = int((p.Total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
}

// SetDefaults 设置默认值
func (p *PaginationRequest) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// StatsResponse 统计响应
type StatsResponse struct {
	TotalUsers     int64 `json:"total_users"`
	TotalAPIKeys   int64 `json:"total_api_keys"`
	TotalProviders int64 `json:"total_providers"`
	TotalModels    int64 `json:"total_models"`
	TotalRequests  int64 `json:"total_requests"`
}

// UsageResponse 使用情况响应
type UsageResponse struct {
	TotalRequests int     `json:"total_requests" example:"100"`
	TotalTokens   int     `json:"total_tokens" example:"5000"`
	TotalCost     float64 `json:"total_cost" example:"1.25"`
}
