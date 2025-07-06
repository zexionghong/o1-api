package utils

import (
	"ai-api-gateway/internal/application/dto"
)

// PaginationHelper 分页助手
type PaginationHelper struct{}

// NewPaginationHelper 创建分页助手
func NewPaginationHelper() *PaginationHelper {
	return &PaginationHelper{}
}

// BuildListResponse 构建列表响应
func (h *PaginationHelper) BuildListResponse(
	data interface{},
	total int64,
	pagination *dto.PaginationRequest,
) *dto.ListResponseBase {
	// 计算总页数
	paginationResp := &dto.PaginationResponse{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Total:    total,
	}
	paginationResp.CalculateTotalPages()

	return &dto.ListResponseBase{
		Data:       data,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: paginationResp.TotalPages,
	}
}

// ValidateAndSetDefaults 验证并设置分页默认值
func (h *PaginationHelper) ValidateAndSetDefaults(pagination *dto.PaginationRequest) {
	pagination.SetDefaults()
}

// GetOffsetAndLimit 获取偏移量和限制数
func (h *PaginationHelper) GetOffsetAndLimit(pagination *dto.PaginationRequest) (int, int) {
	return pagination.GetOffset(), pagination.GetLimit()
}
