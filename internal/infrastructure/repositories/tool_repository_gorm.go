package repositories

import (
	"context"
	"fmt"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"

	"gorm.io/gorm"
)

// toolRepositoryGorm GORM工具仓储实现
type toolRepositoryGorm struct {
	db *gorm.DB
}

// NewToolRepositoryGorm 创建GORM工具仓储
func NewToolRepositoryGorm(db *gorm.DB) repositories.ToolRepository {
	return &toolRepositoryGorm{
		db: db,
	}
}

// GetTools 获取工具模板列表
func (r *toolRepositoryGorm) GetTools(ctx context.Context) ([]*entities.Tool, error) {
	var tools []*entities.Tool
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("category ASC, name ASC").
		Find(&tools).Error; err != nil {
		return nil, fmt.Errorf("failed to get tools: %w", err)
	}

	// 为每个工具加载支持的模型
	for _, tool := range tools {
		if err := r.loadSupportedModels(ctx, tool); err != nil {
			return nil, fmt.Errorf("failed to load supported models for tool %s: %w", tool.ID, err)
		}
	}

	return tools, nil
}

// GetToolByID 根据ID获取工具模板
func (r *toolRepositoryGorm) GetToolByID(ctx context.Context, id string) (*entities.Tool, error) {
	var tool entities.Tool
	if err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&tool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrToolNotFound
		}
		return nil, fmt.Errorf("failed to get tool by id: %w", err)
	}

	// 加载支持的模型
	if err := r.loadSupportedModels(ctx, &tool); err != nil {
		return nil, fmt.Errorf("failed to load supported models for tool %s: %w", tool.ID, err)
	}

	return &tool, nil
}

// CreateUserToolInstance 创建用户工具实例
func (r *toolRepositoryGorm) CreateUserToolInstance(ctx context.Context, instance *entities.UserToolInstance) error {
	if err := r.db.WithContext(ctx).Create(instance).Error; err != nil {
		return fmt.Errorf("failed to create user tool instance: %w", err)
	}
	return nil
}

// GetUserToolInstanceByID 获取用户工具实例
func (r *toolRepositoryGorm) GetUserToolInstanceByID(ctx context.Context, id string) (*entities.UserToolInstance, error) {
	var instance entities.UserToolInstance
	if err := r.db.WithContext(ctx).First(&instance, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrToolInstanceNotFound
		}
		return nil, fmt.Errorf("failed to get user tool instance: %w", err)
	}

	// 填充关联数据
	if err := r.populateUserToolInstanceRelations(ctx, &instance); err != nil {
		return nil, fmt.Errorf("failed to populate relations: %w", err)
	}

	return &instance, nil
}

// GetUserToolInstancesByUserID 获取用户工具实例列表
func (r *toolRepositoryGorm) GetUserToolInstancesByUserID(ctx context.Context, userID int64, includePrivate bool) ([]*entities.UserToolInstance, error) {
	var instances []*entities.UserToolInstance
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if !includePrivate {
		query = query.Where("is_public = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&instances).Error; err != nil {
		return nil, fmt.Errorf("failed to get user tool instances: %w", err)
	}

	// 填充关联数据
	for _, instance := range instances {
		if err := r.populateUserToolInstanceRelations(ctx, instance); err != nil {
			return nil, fmt.Errorf("failed to populate relations: %w", err)
		}
	}

	return instances, nil
}

// UpdateUserToolInstance 更新用户工具实例
func (r *toolRepositoryGorm) UpdateUserToolInstance(ctx context.Context, instance *entities.UserToolInstance) error {
	result := r.db.WithContext(ctx).Save(instance)
	if result.Error != nil {
		return fmt.Errorf("failed to update user tool instance: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrToolInstanceNotFound
	}

	return nil
}

// DeleteUserToolInstance 删除用户工具实例
func (r *toolRepositoryGorm) DeleteUserToolInstance(ctx context.Context, id string, userID int64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&entities.UserToolInstance{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user tool instance: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrToolInstanceNotFound
	}

	return nil
}

// GetPublicUserToolInstances 获取公开的工具实例列表
func (r *toolRepositoryGorm) GetPublicUserToolInstances(ctx context.Context, limit, offset int) ([]*entities.UserToolInstance, error) {
	var instances []*entities.UserToolInstance
	if err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Order("usage_count DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&instances).Error; err != nil {
		return nil, fmt.Errorf("failed to get public tool instances: %w", err)
	}

	// 填充关联数据
	for _, instance := range instances {
		if err := r.populateUserToolInstanceRelations(ctx, instance); err != nil {
			return nil, fmt.Errorf("failed to populate relations: %w", err)
		}
	}

	return instances, nil
}

// GetUserToolInstanceByShareToken 根据分享令牌获取工具实例
func (r *toolRepositoryGorm) GetUserToolInstanceByShareToken(ctx context.Context, shareToken string) (*entities.UserToolInstance, error) {
	var instance entities.UserToolInstance
	if err := r.db.WithContext(ctx).Where("share_token = ? AND is_public = ?", shareToken, true).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrToolInstanceNotFound
		}
		return nil, fmt.Errorf("failed to get tool instance by share token: %w", err)
	}

	// 填充关联数据
	if err := r.populateUserToolInstanceRelations(ctx, &instance); err != nil {
		return nil, fmt.Errorf("failed to populate relations: %w", err)
	}

	return &instance, nil
}

// IncrementUsageCount 增加使用次数
func (r *toolRepositoryGorm) IncrementUsageCount(ctx context.Context, instanceID string) error {
	result := r.db.WithContext(ctx).Model(&entities.UserToolInstance{}).
		Where("id = ?", instanceID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("failed to increment usage count: %w", result.Error)
	}

	return nil
}

// CreateUsageLog 创建工具使用记录
func (r *toolRepositoryGorm) CreateUsageLog(ctx context.Context, log *entities.ToolUsageLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create tool usage log: %w", err)
	}
	return nil
}

// GetUsageLogsByInstanceID 获取工具使用记录列表
func (r *toolRepositoryGorm) GetUsageLogsByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*entities.ToolUsageLog, error) {
	var logs []*entities.ToolUsageLog
	if err := r.db.WithContext(ctx).
		Where("tool_instance_id = ?", instanceID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get tool usage logs: %w", err)
	}
	return logs, nil
}

// loadSupportedModels 为工具加载支持的模型列表
func (r *toolRepositoryGorm) loadSupportedModels(ctx context.Context, tool *entities.Tool) error {
	var models []entities.Model
	if err := r.db.WithContext(ctx).
		Table("models").
		Joins("INNER JOIN tool_model_support ON models.id = tool_model_support.model_id").
		Where("tool_model_support.tool_id = ? AND models.status = ?", tool.ID, entities.ModelStatusActive).
		Order("models.name ASC").
		Find(&models).Error; err != nil {
		return fmt.Errorf("failed to load supported models: %w", err)
	}

	tool.SupportedModels = models
	return nil
}

// populateUserToolInstanceRelations 填充用户工具实例的关联数据
func (r *toolRepositoryGorm) populateUserToolInstanceRelations(ctx context.Context, instance *entities.UserToolInstance) error {
	// 填充工具模板信息
	var tool entities.Tool
	if err := r.db.WithContext(ctx).Where("id = ?", instance.ToolID).First(&tool).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to get tool: %w", err)
		}
	} else {
		instance.Tool = &tool
		// 映射 category 到前端期望的 type
		instance.Type = r.mapCategoryToType(tool.Category)
	}

	// 填充创建者信息
	var creator entities.User
	if err := r.db.WithContext(ctx).Where("id = ?", instance.UserID).First(&creator).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to get creator: %w", err)
		}
	} else {
		instance.Creator = &creator
	}

	// 填充API Key信息
	var apiKey entities.APIKey
	if err := r.db.WithContext(ctx).Where("id = ?", instance.APIKeyID).First(&apiKey).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to get api key: %w", err)
		}
	} else {
		instance.APIKey = &apiKey
	}

	// 填充模型名称
	var model entities.Model
	if err := r.db.WithContext(ctx).Where("id = ?", instance.ModelID).First(&model).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to get model: %w", err)
		}
	} else {
		instance.ModelName = model.Name
	}

	// 填充前端期望的字段格式
	instance.ModelIDStr = fmt.Sprintf("%d", instance.ModelID)
	instance.APIKeyIDStr = fmt.Sprintf("%d", instance.APIKeyID)

	// 生成分享URL
	if instance.ShareToken != nil {
		shareURL := fmt.Sprintf("/tools/share/%s", *instance.ShareToken)
		instance.ShareURL = &shareURL
	}

	return nil
}

// mapCategoryToType 将工具类别映射到前端期望的类型
func (r *toolRepositoryGorm) mapCategoryToType(category string) string {
	switch category {
	case "Communication":
		return "chatbot"
	case "Image":
		return "image_generator"
	case "Text":
		return "text_generator"
	case "Code":
		return "code_assistant"
	case "Data":
		return "data_analyzer"
	default:
		return "chatbot" // 默认类型
	}
}
