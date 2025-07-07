package repositories

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"

	"gorm.io/gorm"
)

// providerModelSupportRepositoryGorm GORM提供商模型支持仓储实现
type providerModelSupportRepositoryGorm struct {
	db *gorm.DB
}

// NewProviderModelSupportRepositoryGorm 创建GORM提供商模型支持仓储
func NewProviderModelSupportRepositoryGorm(db *gorm.DB) repositories.ProviderModelSupportRepository {
	return &providerModelSupportRepositoryGorm{
		db: db,
	}
}

// Create 创建提供商模型支持
func (r *providerModelSupportRepositoryGorm) Create(ctx context.Context, support *entities.ProviderModelSupport) error {
	if err := r.db.WithContext(ctx).Create(support).Error; err != nil {
		return fmt.Errorf("failed to create provider model support: %w", err)
	}
	return nil
}

// GetByID 根据ID获取提供商模型支持
func (r *providerModelSupportRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.ProviderModelSupport, error) {
	var support entities.ProviderModelSupport
	if err := r.db.WithContext(ctx).First(&support, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrProviderModelSupportNotFound
		}
		return nil, fmt.Errorf("failed to get provider model support by id: %w", err)
	}
	return &support, nil
}

// GetByProviderAndModel 根据提供商和模型获取支持信息
func (r *providerModelSupportRepositoryGorm) GetByProviderAndModel(ctx context.Context, providerID int64, modelSlug string) (*entities.ProviderModelSupport, error) {
	var support entities.ProviderModelSupport
	if err := r.db.WithContext(ctx).
		Where("provider_id = ? AND model_slug = ?", providerID, modelSlug).
		First(&support).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrProviderModelSupportNotFound
		}
		return nil, fmt.Errorf("failed to get provider model support: %w", err)
	}
	return &support, nil
}

// GetSupportingProviders 获取支持指定模型的提供商列表
func (r *providerModelSupportRepositoryGorm) GetSupportingProviders(ctx context.Context, modelSlug string) ([]*entities.ModelSupportInfo, error) {
	var results []*entities.ModelSupportInfo

	// 使用原生SQL查询来获取关联数据
	query := `
		SELECT
			pms.id, pms.provider_id, pms.model_slug, pms.upstream_model_name,
			pms.enabled, pms.priority, pms.config,
			p.id as p_id, p.name as provider_name, p.slug as provider_slug,
			p.base_url as provider_base_url, p.api_key_encrypted as provider_api_key,
			p.timeout_seconds as provider_timeout_seconds, p.retry_attempts as provider_retry_attempts,
			p.status as provider_status, p.health_status as provider_health_status
		FROM provider_model_support pms
		JOIN providers p ON pms.provider_id = p.id
		WHERE pms.model_slug = ? AND pms.enabled = true AND p.status = 'active'
		ORDER BY pms.priority ASC, p.priority ASC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, modelSlug).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get supporting providers: %w", err)
	}
	defer rows.Close()

	// 调试：检查是否有行返回
	fmt.Printf("DEBUG: Query executed for model_slug=%s\n", modelSlug)

	rowCount := 0
	for rows.Next() {
		rowCount++
		fmt.Printf("DEBUG: Processing row %d\n", rowCount)

		var info entities.ModelSupportInfo
		var support entities.ProviderModelSupport
		var provider entities.Provider

		err := rows.Scan(
			&support.ID, &support.ProviderID, &support.ModelSlug, &support.UpstreamModelName,
			&support.Enabled, &support.Priority, &support.Config,
			&provider.ID, &provider.Name, &provider.Slug,
			&provider.BaseURL, &provider.APIKeyEncrypted, &provider.TimeoutSeconds, &provider.RetryAttempts,
			&provider.Status, &provider.HealthStatus,
		)
		if err != nil {
			fmt.Printf("DEBUG: Scan error: %v\n", err)
			return nil, fmt.Errorf("failed to scan provider model support: %w", err)
		}

		fmt.Printf("DEBUG: Scanned provider: ID=%d, Name=%s, BaseURL=%s, APIKey=%v\n",
			provider.ID, provider.Name, provider.BaseURL, provider.APIKeyEncrypted != nil)

		info.Provider = &provider
		info.ModelSlug = support.ModelSlug
		info.UpstreamModelName = support.GetUpstreamModelName()
		info.Priority = support.Priority
		info.Enabled = support.Enabled
		info.Support = &support

		// 解析配置
		if config, err := support.GetConfig(); err == nil {
			info.Config = config
		}

		results = append(results, &info)
	}

	fmt.Printf("DEBUG: Total rows processed: %d, results count: %d\n", rowCount, len(results))
	return results, nil
}

// GetProviderSupportedModels 获取提供商支持的模型列表
func (r *providerModelSupportRepositoryGorm) GetProviderSupportedModels(ctx context.Context, providerID int64) ([]*entities.ProviderModelSupport, error) {
	var supports []*entities.ProviderModelSupport
	if err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Order("priority ASC, model_slug ASC").
		Find(&supports).Error; err != nil {
		return nil, fmt.Errorf("failed to get provider supported models: %w", err)
	}
	return supports, nil
}

// Update 更新提供商模型支持
func (r *providerModelSupportRepositoryGorm) Update(ctx context.Context, support *entities.ProviderModelSupport) error {
	support.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Save(support)
	if result.Error != nil {
		return fmt.Errorf("failed to update provider model support: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// Delete 删除提供商模型支持
func (r *providerModelSupportRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entities.ProviderModelSupport{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete provider model support: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// EnableSupport 启用模型支持
func (r *providerModelSupportRepositoryGorm) EnableSupport(ctx context.Context, providerID int64, modelSlug string) error {
	result := r.db.WithContext(ctx).Model(&entities.ProviderModelSupport{}).
		Where("provider_id = ? AND model_slug = ?", providerID, modelSlug).
		Updates(map[string]interface{}{
			"enabled":    true,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to enable provider model support: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// DisableSupport 禁用模型支持
func (r *providerModelSupportRepositoryGorm) DisableSupport(ctx context.Context, providerID int64, modelSlug string) error {
	result := r.db.WithContext(ctx).Model(&entities.ProviderModelSupport{}).
		Where("provider_id = ? AND model_slug = ?", providerID, modelSlug).
		Updates(map[string]interface{}{
			"enabled":    false,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to disable provider model support: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// List 获取提供商模型支持列表
func (r *providerModelSupportRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.ProviderModelSupport, error) {
	var supports []*entities.ProviderModelSupport
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&supports).Error; err != nil {
		return nil, fmt.Errorf("failed to list provider model support: %w", err)
	}
	return supports, nil
}

// Count 获取提供商模型支持总数
func (r *providerModelSupportRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.ProviderModelSupport{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count provider model support: %w", err)
	}
	return count, nil
}
