package repositories

import (
	"ai-api-gateway/internal/domain/entities"
	"context"
)

// ToolRepository 工具仓库接口
type ToolRepository interface {
	// 工具模板相关
	GetTools(ctx context.Context) ([]*entities.Tool, error)
	GetToolByID(ctx context.Context, id string) (*entities.Tool, error)

	// 用户工具实例相关
	CreateUserToolInstance(ctx context.Context, instance *entities.UserToolInstance) error
	GetUserToolInstanceByID(ctx context.Context, id string) (*entities.UserToolInstance, error)
	GetUserToolInstancesByUserID(ctx context.Context, userID int64, includePrivate bool) ([]*entities.UserToolInstance, error)
	GetPublicUserToolInstances(ctx context.Context, limit, offset int) ([]*entities.UserToolInstance, error)
	GetUserToolInstanceByShareToken(ctx context.Context, shareToken string) (*entities.UserToolInstance, error)
	UpdateUserToolInstance(ctx context.Context, instance *entities.UserToolInstance) error
	DeleteUserToolInstance(ctx context.Context, id string, userID int64) error
	IncrementUsageCount(ctx context.Context, instanceID string) error

	// 工具使用记录相关
	CreateUsageLog(ctx context.Context, log *entities.ToolUsageLog) error
	GetUsageLogsByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*entities.ToolUsageLog, error)
}
