package repositories

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"gorm.io/gorm"
)

// modelRepositoryGorm GORM模型仓储实现
type modelRepositoryGorm struct {
	db *gorm.DB
}

// NewModelRepositoryGorm 创建GORM模型仓储
func NewModelRepositoryGorm(db *gorm.DB) repositories.ModelRepository {
	return &modelRepositoryGorm{
		db: db,
	}
}

// Create 创建模型
func (r *modelRepositoryGorm) Create(ctx context.Context, model *entities.Model) error {
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	return nil
}

// GetByID 根据ID获取模型
func (r *modelRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.Model, error) {
	var model entities.Model
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrModelNotFound
		}
		return nil, fmt.Errorf("failed to get model by id: %w", err)
	}
	return &model, nil
}

// GetBySlug 根据slug获取模型
func (r *modelRepositoryGorm) GetBySlug(ctx context.Context, slug string) (*entities.Model, error) {
	var model entities.Model
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrModelNotFound
		}
		return nil, fmt.Errorf("failed to get model by slug: %w", err)
	}
	return &model, nil
}

// Update 更新模型
func (r *modelRepositoryGorm) Update(ctx context.Context, model *entities.Model) error {
	model.UpdatedAt = time.Now()
	
	result := r.db.WithContext(ctx).Save(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update model: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrModelNotFound
	}
	
	return nil
}

// Delete 删除模型
func (r *modelRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entities.Model{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete model: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrModelNotFound
	}
	
	return nil
}

// List 获取模型列表
func (r *modelRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.Model, error) {
	var models []*entities.Model
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	return models, nil
}

// Count 获取模型总数
func (r *modelRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.Model{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count models: %w", err)
	}
	return count, nil
}

// GetActiveModels 获取活跃的模型列表
func (r *modelRepositoryGorm) GetActiveModels(ctx context.Context) ([]*entities.Model, error) {
	var models []*entities.Model
	if err := r.db.WithContext(ctx).
		Where("status = ?", entities.ModelStatusActive).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get active models: %w", err)
	}
	return models, nil
}

// GetModelsByType 根据类型获取模型列表
func (r *modelRepositoryGorm) GetModelsByType(ctx context.Context, modelType entities.ModelType) ([]*entities.Model, error) {
	var models []*entities.Model
	if err := r.db.WithContext(ctx).
		Where("model_type = ? AND status = ?", modelType, entities.ModelStatusActive).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get models by type: %w", err)
	}
	return models, nil
}

// GetAvailableModels 获取可用的模型列表
func (r *modelRepositoryGorm) GetAvailableModels(ctx context.Context) ([]*entities.Model, error) {
	var models []*entities.Model
	if err := r.db.WithContext(ctx).
		Where("status = ?", entities.ModelStatusActive).
		Order("model_type ASC, created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}
	return models, nil
}

// modelPricingRepositoryGorm GORM模型定价仓储实现
type modelPricingRepositoryGorm struct {
	db *gorm.DB
}

// NewModelPricingRepositoryGorm 创建GORM模型定价仓储
func NewModelPricingRepositoryGorm(db *gorm.DB) repositories.ModelPricingRepository {
	return &modelPricingRepositoryGorm{
		db: db,
	}
}

// Create 创建模型定价
func (r *modelPricingRepositoryGorm) Create(ctx context.Context, pricing *entities.ModelPricing) error {
	if err := r.db.WithContext(ctx).Create(pricing).Error; err != nil {
		return fmt.Errorf("failed to create model pricing: %w", err)
	}
	return nil
}

// GetByID 根据ID获取模型定价
func (r *modelPricingRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.ModelPricing, error) {
	var pricing entities.ModelPricing
	if err := r.db.WithContext(ctx).First(&pricing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrModelPricingNotFound
		}
		return nil, fmt.Errorf("failed to get model pricing by id: %w", err)
	}
	return &pricing, nil
}

// GetByModelID 根据模型ID获取定价列表
func (r *modelPricingRepositoryGorm) GetByModelID(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error) {
	var pricings []*entities.ModelPricing
	if err := r.db.WithContext(ctx).
		Where("model_id = ?", modelID).
		Order("effective_from DESC").
		Find(&pricings).Error; err != nil {
		return nil, fmt.Errorf("failed to get model pricing by model id: %w", err)
	}
	return pricings, nil
}

// GetCurrentPricing 获取模型当前有效定价
func (r *modelPricingRepositoryGorm) GetCurrentPricing(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error) {
	var pricings []*entities.ModelPricing
	now := time.Now()
	
	if err := r.db.WithContext(ctx).
		Where("model_id = ? AND effective_from <= ? AND (effective_until IS NULL OR effective_until > ?)", 
			modelID, now, now).
		Order("pricing_type ASC").
		Find(&pricings).Error; err != nil {
		return nil, fmt.Errorf("failed to get current model pricing: %w", err)
	}
	return pricings, nil
}

// Update 更新模型定价
func (r *modelPricingRepositoryGorm) Update(ctx context.Context, pricing *entities.ModelPricing) error {
	result := r.db.WithContext(ctx).Save(pricing)
	if result.Error != nil {
		return fmt.Errorf("failed to update model pricing: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrModelPricingNotFound
	}
	
	return nil
}

// Delete 删除模型定价
func (r *modelPricingRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entities.ModelPricing{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete model pricing: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrModelPricingNotFound
	}
	
	return nil
}

// List 获取模型定价列表
func (r *modelPricingRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.ModelPricing, error) {
	var pricings []*entities.ModelPricing
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&pricings).Error; err != nil {
		return nil, fmt.Errorf("failed to list model pricing: %w", err)
	}
	return pricings, nil
}

// Count 获取模型定价总数
func (r *modelPricingRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.ModelPricing{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count model pricing: %w", err)
	}
	return count, nil
}

// GetPricingByType 根据定价类型获取定价
func (r *modelPricingRepositoryGorm) GetPricingByType(ctx context.Context, modelID int64, pricingType entities.PricingType) (*entities.ModelPricing, error) {
	var pricing entities.ModelPricing
	now := time.Now()
	
	if err := r.db.WithContext(ctx).
		Where("model_id = ? AND pricing_type = ? AND effective_from <= ? AND (effective_until IS NULL OR effective_until > ?)", 
			modelID, pricingType, now, now).
		Order("effective_from DESC").
		First(&pricing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrModelPricingNotFound
		}
		return nil, fmt.Errorf("failed to get model pricing by type: %w", err)
	}
	return &pricing, nil
}
