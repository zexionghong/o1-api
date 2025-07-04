package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// providerRepositoryImpl 提供商仓储实现
type providerRepositoryImpl struct {
	db *sql.DB
}

// NewProviderRepository 创建提供商仓储
func NewProviderRepository(db *sql.DB) repositories.ProviderRepository {
	return &providerRepositoryImpl{
		db: db,
	}
}

// Create 创建服务提供商
func (r *providerRepositoryImpl) Create(ctx context.Context, provider *entities.Provider) error {
	query := `
		INSERT INTO providers (name, slug, base_url, api_key_encrypted, status, priority, timeout_seconds, retry_attempts, health_check_url, health_check_interval, health_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now
	
	result, err := r.db.ExecContext(ctx, query,
		provider.Name,
		provider.Slug,
		provider.BaseURL,
		provider.APIKeyEncrypted,
		provider.Status,
		provider.Priority,
		provider.TimeoutSeconds,
		provider.RetryAttempts,
		provider.HealthCheckURL,
		provider.HealthCheckInterval,
		provider.HealthStatus,
		provider.CreatedAt,
		provider.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	provider.ID = id
	return nil
}

// GetByID 根据ID获取服务提供商
func (r *providerRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.Provider, error) {
	query := `
		SELECT id, name, slug, base_url, api_key_encrypted, status, priority, timeout_seconds, retry_attempts, health_check_url, health_check_interval, last_health_check, health_status, created_at, updated_at
		FROM providers WHERE id = ?
	`
	
	provider := &entities.Provider{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&provider.ID,
		&provider.Name,
		&provider.Slug,
		&provider.BaseURL,
		&provider.APIKeyEncrypted,
		&provider.Status,
		&provider.Priority,
		&provider.TimeoutSeconds,
		&provider.RetryAttempts,
		&provider.HealthCheckURL,
		&provider.HealthCheckInterval,
		&provider.LastHealthCheck,
		&provider.HealthStatus,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get provider by id: %w", err)
	}
	
	return provider, nil
}

// GetBySlug 根据slug获取服务提供商
func (r *providerRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entities.Provider, error) {
	query := `
		SELECT id, name, slug, base_url, api_key_encrypted, status, priority, timeout_seconds, retry_attempts, health_check_url, health_check_interval, last_health_check, health_status, created_at, updated_at
		FROM providers WHERE slug = ?
	`
	
	provider := &entities.Provider{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&provider.ID,
		&provider.Name,
		&provider.Slug,
		&provider.BaseURL,
		&provider.APIKeyEncrypted,
		&provider.Status,
		&provider.Priority,
		&provider.TimeoutSeconds,
		&provider.RetryAttempts,
		&provider.HealthCheckURL,
		&provider.HealthCheckInterval,
		&provider.LastHealthCheck,
		&provider.HealthStatus,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrProviderNotFound
		}
		return nil, fmt.Errorf("failed to get provider by slug: %w", err)
	}
	
	return provider, nil
}

// Update 更新服务提供商
func (r *providerRepositoryImpl) Update(ctx context.Context, provider *entities.Provider) error {
	query := `
		UPDATE providers 
		SET name = ?, base_url = ?, api_key_encrypted = ?, status = ?, priority = ?, timeout_seconds = ?, retry_attempts = ?, health_check_url = ?, health_check_interval = ?, health_status = ?, updated_at = ?
		WHERE id = ?
	`
	
	provider.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		provider.Name,
		provider.BaseURL,
		provider.APIKeyEncrypted,
		provider.Status,
		provider.Priority,
		provider.TimeoutSeconds,
		provider.RetryAttempts,
		provider.HealthCheckURL,
		provider.HealthCheckInterval,
		provider.HealthStatus,
		provider.UpdatedAt,
		provider.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return entities.ErrProviderNotFound
	}
	
	return nil
}

// UpdateHealthStatus 更新健康状态
func (r *providerRepositoryImpl) UpdateHealthStatus(ctx context.Context, id int64, status entities.HealthStatus) error {
	query := `
		UPDATE providers 
		SET health_status = ?, last_health_check = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update health status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return entities.ErrProviderNotFound
	}
	
	return nil
}

// Delete 删除服务提供商
func (r *providerRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM providers WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return entities.ErrProviderNotFound
	}
	
	return nil
}

// List 获取服务提供商列表
func (r *providerRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.Provider, error) {
	query := `
		SELECT id, name, slug, base_url, api_key_encrypted, status, priority, timeout_seconds, retry_attempts, health_check_url, health_check_interval, last_health_check, health_status, created_at, updated_at
		FROM providers 
		ORDER BY priority ASC, created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}
	defer rows.Close()
	
	var providers []*entities.Provider
	for rows.Next() {
		provider := &entities.Provider{}
		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Slug,
			&provider.BaseURL,
			&provider.APIKeyEncrypted,
			&provider.Status,
			&provider.Priority,
			&provider.TimeoutSeconds,
			&provider.RetryAttempts,
			&provider.HealthCheckURL,
			&provider.HealthCheckInterval,
			&provider.LastHealthCheck,
			&provider.HealthStatus,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate providers: %w", err)
	}
	
	return providers, nil
}

// Count 获取服务提供商总数
func (r *providerRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM providers`
	
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count providers: %w", err)
	}
	
	return count, nil
}

// GetActiveProviders 获取活跃的服务提供商列表
func (r *providerRepositoryImpl) GetActiveProviders(ctx context.Context) ([]*entities.Provider, error) {
	query := `
		SELECT id, name, slug, base_url, api_key_encrypted, status, priority, timeout_seconds, retry_attempts, health_check_url, health_check_interval, last_health_check, health_status, created_at, updated_at
		FROM providers 
		WHERE status = ?
		ORDER BY priority ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ProviderStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get active providers: %w", err)
	}
	defer rows.Close()
	
	var providers []*entities.Provider
	for rows.Next() {
		provider := &entities.Provider{}
		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Slug,
			&provider.BaseURL,
			&provider.APIKeyEncrypted,
			&provider.Status,
			&provider.Priority,
			&provider.TimeoutSeconds,
			&provider.RetryAttempts,
			&provider.HealthCheckURL,
			&provider.HealthCheckInterval,
			&provider.LastHealthCheck,
			&provider.HealthStatus,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate providers: %w", err)
	}
	
	return providers, nil
}

// GetAvailableProviders 获取可用的服务提供商列表（活跃且健康）
func (r *providerRepositoryImpl) GetAvailableProviders(ctx context.Context) ([]*entities.Provider, error) {
	query := `
		SELECT id, name, slug, base_url, api_key_encrypted, status, priority, timeout_seconds, retry_attempts, health_check_url, health_check_interval, last_health_check, health_status, created_at, updated_at
		FROM providers 
		WHERE status = ? AND health_status = ?
		ORDER BY priority ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ProviderStatusActive, entities.HealthStatusHealthy)
	if err != nil {
		return nil, fmt.Errorf("failed to get available providers: %w", err)
	}
	defer rows.Close()
	
	var providers []*entities.Provider
	for rows.Next() {
		provider := &entities.Provider{}
		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Slug,
			&provider.BaseURL,
			&provider.APIKeyEncrypted,
			&provider.Status,
			&provider.Priority,
			&provider.TimeoutSeconds,
			&provider.RetryAttempts,
			&provider.HealthCheckURL,
			&provider.HealthCheckInterval,
			&provider.LastHealthCheck,
			&provider.HealthStatus,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate providers: %w", err)
	}
	
	return providers, nil
}

// GetProvidersByPriority 按优先级获取服务提供商列表
func (r *providerRepositoryImpl) GetProvidersByPriority(ctx context.Context) ([]*entities.Provider, error) {
	return r.GetActiveProviders(ctx)
}

// GetProvidersNeedingHealthCheck 获取需要健康检查的服务提供商列表
func (r *providerRepositoryImpl) GetProvidersNeedingHealthCheck(ctx context.Context) ([]*entities.Provider, error) {
	query := `
		SELECT id, name, slug, base_url, api_key_encrypted, status, priority, timeout_seconds, retry_attempts, health_check_url, health_check_interval, last_health_check, health_status, created_at, updated_at
		FROM providers 
		WHERE status = ? AND (
			last_health_check IS NULL OR 
			datetime(last_health_check, '+' || health_check_interval || ' seconds') <= datetime('now')
		)
		ORDER BY priority ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ProviderStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers needing health check: %w", err)
	}
	defer rows.Close()
	
	var providers []*entities.Provider
	for rows.Next() {
		provider := &entities.Provider{}
		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Slug,
			&provider.BaseURL,
			&provider.APIKeyEncrypted,
			&provider.Status,
			&provider.Priority,
			&provider.TimeoutSeconds,
			&provider.RetryAttempts,
			&provider.HealthCheckURL,
			&provider.HealthCheckInterval,
			&provider.LastHealthCheck,
			&provider.HealthStatus,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate providers: %w", err)
	}
	
	return providers, nil
}
