package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// providerModelSupportRepositoryImpl 提供商模型支持仓储实现
type providerModelSupportRepositoryImpl struct {
	db *sql.DB
}

// NewProviderModelSupportRepository 创建提供商模型支持仓储
func NewProviderModelSupportRepository(db *sql.DB) repositories.ProviderModelSupportRepository {
	return &providerModelSupportRepositoryImpl{
		db: db,
	}
}

// Create 创建提供商模型支持
func (r *providerModelSupportRepositoryImpl) Create(ctx context.Context, support *entities.ProviderModelSupport) error {
	query := `
		INSERT INTO provider_model_support (provider_id, model_slug, upstream_model_name, enabled, priority, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	support.CreatedAt = now
	support.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		support.ProviderID,
		support.ModelSlug,
		support.UpstreamModelName,
		support.Enabled,
		support.Priority,
		support.Config,
		support.CreatedAt,
		support.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create provider model support: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	support.ID = id
	return nil
}

// GetByID 根据ID获取提供商模型支持
func (r *providerModelSupportRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.ProviderModelSupport, error) {
	query := `
		SELECT id, provider_id, model_slug, upstream_model_name, enabled, priority, config, created_at, updated_at
		FROM provider_model_support WHERE id = ?
	`

	support := &entities.ProviderModelSupport{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&support.ID,
		&support.ProviderID,
		&support.ModelSlug,
		&support.UpstreamModelName,
		&support.Enabled,
		&support.Priority,
		&support.Config,
		&support.CreatedAt,
		&support.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrProviderModelSupportNotFound
		}
		return nil, fmt.Errorf("failed to get provider model support by id: %w", err)
	}

	return support, nil
}

// GetByProviderAndModel 根据提供商和模型获取支持信息
func (r *providerModelSupportRepositoryImpl) GetByProviderAndModel(ctx context.Context, providerID int64, modelSlug string) (*entities.ProviderModelSupport, error) {
	query := `
		SELECT id, provider_id, model_slug, upstream_model_name, enabled, priority, config, created_at, updated_at
		FROM provider_model_support WHERE provider_id = ? AND model_slug = ?
	`

	support := &entities.ProviderModelSupport{}
	err := r.db.QueryRowContext(ctx, query, providerID, modelSlug).Scan(
		&support.ID,
		&support.ProviderID,
		&support.ModelSlug,
		&support.UpstreamModelName,
		&support.Enabled,
		&support.Priority,
		&support.Config,
		&support.CreatedAt,
		&support.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrProviderModelSupportNotFound
		}
		return nil, fmt.Errorf("failed to get provider model support: %w", err)
	}

	return support, nil
}

// GetSupportingProviders 获取支持指定模型的提供商列表
func (r *providerModelSupportRepositoryImpl) GetSupportingProviders(ctx context.Context, modelSlug string) ([]*entities.ModelSupportInfo, error) {
	query := `
		SELECT 
			pms.id, pms.provider_id, pms.model_slug, pms.upstream_model_name, pms.enabled, pms.priority, pms.config, pms.created_at, pms.updated_at,
			p.id, p.name, p.slug, p.base_url, p.api_key_encrypted, p.status, p.priority, p.timeout_seconds, p.retry_attempts, 
			p.health_check_url, p.health_check_interval, p.last_health_check, p.health_status, p.created_at, p.updated_at
		FROM provider_model_support pms
		INNER JOIN providers p ON pms.provider_id = p.id
		WHERE pms.model_slug = ? AND pms.enabled = true AND p.status = ? AND p.health_status = ?
		ORDER BY pms.priority ASC, p.priority ASC
	`

	rows, err := r.db.QueryContext(ctx, query, modelSlug, entities.ProviderStatusActive, entities.HealthStatusHealthy)
	if err != nil {
		return nil, fmt.Errorf("failed to get supporting providers: %w", err)
	}
	defer rows.Close()

	var supportInfos []*entities.ModelSupportInfo
	for rows.Next() {
		support := &entities.ProviderModelSupport{}
		provider := &entities.Provider{}

		err := rows.Scan(
			&support.ID, &support.ProviderID, &support.ModelSlug, &support.UpstreamModelName,
			&support.Enabled, &support.Priority, &support.Config, &support.CreatedAt, &support.UpdatedAt,
			&provider.ID, &provider.Name, &provider.Slug, &provider.BaseURL, &provider.APIKeyEncrypted,
			&provider.Status, &provider.Priority, &provider.TimeoutSeconds, &provider.RetryAttempts,
			&provider.HealthCheckURL, &provider.HealthCheckInterval, &provider.LastHealthCheck,
			&provider.HealthStatus, &provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supporting provider: %w", err)
		}

		upstreamModelName := support.ModelSlug
		if support.UpstreamModelName != nil {
			upstreamModelName = *support.UpstreamModelName
		}

		supportInfo := &entities.ModelSupportInfo{
			Provider:          provider,
			ModelSlug:         support.ModelSlug,
			UpstreamModelName: upstreamModelName,
			Priority:          support.Priority,
			Enabled:           support.Enabled,
			Support:           support,
		}

		// 解析配置
		if support.Config != nil {
			config, err := support.GetConfig()
			if err == nil {
				supportInfo.Config = config
			}
		}

		supportInfos = append(supportInfos, supportInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate supporting providers: %w", err)
	}

	return supportInfos, nil
}

// GetProviderSupportedModels 获取提供商支持的模型列表
func (r *providerModelSupportRepositoryImpl) GetProviderSupportedModels(ctx context.Context, providerID int64) ([]*entities.ProviderModelSupport, error) {
	query := `
		SELECT id, provider_id, model_slug, upstream_model_name, enabled, priority, config, created_at, updated_at
		FROM provider_model_support 
		WHERE provider_id = ? AND enabled = true
		ORDER BY priority ASC, model_slug ASC
	`

	rows, err := r.db.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider supported models: %w", err)
	}
	defer rows.Close()

	var supports []*entities.ProviderModelSupport
	for rows.Next() {
		support := &entities.ProviderModelSupport{}
		err := rows.Scan(
			&support.ID,
			&support.ProviderID,
			&support.ModelSlug,
			&support.UpstreamModelName,
			&support.Enabled,
			&support.Priority,
			&support.Config,
			&support.CreatedAt,
			&support.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider model support: %w", err)
		}

		supports = append(supports, support)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate provider supported models: %w", err)
	}

	return supports, nil
}

// Update 更新提供商模型支持
func (r *providerModelSupportRepositoryImpl) Update(ctx context.Context, support *entities.ProviderModelSupport) error {
	query := `
		UPDATE provider_model_support
		SET upstream_model_name = ?, enabled = ?, priority = ?, config = ?, updated_at = ?
		WHERE id = ?
	`

	support.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		support.UpstreamModelName,
		support.Enabled,
		support.Priority,
		support.Config,
		support.UpdatedAt,
		support.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update provider model support: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// Delete 删除提供商模型支持
func (r *providerModelSupportRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM provider_model_support WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete provider model support: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// EnableSupport 启用模型支持
func (r *providerModelSupportRepositoryImpl) EnableSupport(ctx context.Context, providerID int64, modelSlug string) error {
	query := `
		UPDATE provider_model_support
		SET enabled = true, updated_at = ?
		WHERE provider_id = ? AND model_slug = ?
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), providerID, modelSlug)
	if err != nil {
		return fmt.Errorf("failed to enable model support: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// DisableSupport 禁用模型支持
func (r *providerModelSupportRepositoryImpl) DisableSupport(ctx context.Context, providerID int64, modelSlug string) error {
	query := `
		UPDATE provider_model_support
		SET enabled = false, updated_at = ?
		WHERE provider_id = ? AND model_slug = ?
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), providerID, modelSlug)
	if err != nil {
		return fmt.Errorf("failed to disable model support: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrProviderModelSupportNotFound
	}

	return nil
}

// List 获取提供商模型支持列表
func (r *providerModelSupportRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.ProviderModelSupport, error) {
	query := `
		SELECT id, provider_id, model_slug, upstream_model_name, enabled, priority, config, created_at, updated_at
		FROM provider_model_support
		ORDER BY provider_id ASC, priority ASC, model_slug ASC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list provider model supports: %w", err)
	}
	defer rows.Close()

	var supports []*entities.ProviderModelSupport
	for rows.Next() {
		support := &entities.ProviderModelSupport{}
		err := rows.Scan(
			&support.ID,
			&support.ProviderID,
			&support.ModelSlug,
			&support.UpstreamModelName,
			&support.Enabled,
			&support.Priority,
			&support.Config,
			&support.CreatedAt,
			&support.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider model support: %w", err)
		}

		supports = append(supports, support)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate provider model supports: %w", err)
	}

	return supports, nil
}

// Count 获取提供商模型支持总数
func (r *providerModelSupportRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM provider_model_support`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count provider model supports: %w", err)
	}

	return count, nil
}
