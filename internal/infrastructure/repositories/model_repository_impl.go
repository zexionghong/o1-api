package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// modelRepositoryImpl 模型仓储实现
type modelRepositoryImpl struct {
	db *sql.DB
}

// NewModelRepository 创建模型仓储
func NewModelRepository(db *sql.DB) repositories.ModelRepository {
	return &modelRepositoryImpl{
		db: db,
	}
}

// Create 创建模型
func (r *modelRepositoryImpl) Create(ctx context.Context, model *entities.Model) error {
	query := `
		INSERT INTO models (provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
	
	result, err := r.db.ExecContext(ctx, query,
		model.ProviderID,
		model.Name,
		model.Slug,
		model.DisplayName,
		model.Description,
		model.ModelType,
		model.ContextLength,
		model.MaxTokens,
		model.SupportsStreaming,
		model.SupportsFunctions,
		model.Status,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	model.ID = id
	return nil
}

// GetByID 根据ID获取模型
func (r *modelRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.Model, error) {
	query := `
		SELECT id, provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at
		FROM models WHERE id = ?
	`
	
	model := &entities.Model{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID,
		&model.ProviderID,
		&model.Name,
		&model.Slug,
		&model.DisplayName,
		&model.Description,
		&model.ModelType,
		&model.ContextLength,
		&model.MaxTokens,
		&model.SupportsStreaming,
		&model.SupportsFunctions,
		&model.Status,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrModelNotFound
		}
		return nil, fmt.Errorf("failed to get model by id: %w", err)
	}
	
	return model, nil
}

// GetBySlug 根据slug获取模型
func (r *modelRepositoryImpl) GetBySlug(ctx context.Context, providerID int64, slug string) (*entities.Model, error) {
	query := `
		SELECT id, provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at
		FROM models WHERE provider_id = ? AND slug = ?
	`
	
	model := &entities.Model{}
	err := r.db.QueryRowContext(ctx, query, providerID, slug).Scan(
		&model.ID,
		&model.ProviderID,
		&model.Name,
		&model.Slug,
		&model.DisplayName,
		&model.Description,
		&model.ModelType,
		&model.ContextLength,
		&model.MaxTokens,
		&model.SupportsStreaming,
		&model.SupportsFunctions,
		&model.Status,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrModelNotFound
		}
		return nil, fmt.Errorf("failed to get model by slug: %w", err)
	}
	
	return model, nil
}

// GetByProviderID 根据提供商ID获取模型列表
func (r *modelRepositoryImpl) GetByProviderID(ctx context.Context, providerID int64) ([]*entities.Model, error) {
	query := `
		SELECT id, provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at
		FROM models 
		WHERE provider_id = ?
		ORDER BY name ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get models by provider id: %w", err)
	}
	defer rows.Close()
	
	var models []*entities.Model
	for rows.Next() {
		model := &entities.Model{}
		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.Name,
			&model.Slug,
			&model.DisplayName,
			&model.Description,
			&model.ModelType,
			&model.ContextLength,
			&model.MaxTokens,
			&model.SupportsStreaming,
			&model.SupportsFunctions,
			&model.Status,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, model)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate models: %w", err)
	}
	
	return models, nil
}

// Update 更新模型
func (r *modelRepositoryImpl) Update(ctx context.Context, model *entities.Model) error {
	query := `
		UPDATE models 
		SET name = ?, display_name = ?, description = ?, model_type = ?, context_length = ?, max_tokens = ?, supports_streaming = ?, supports_functions = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	
	model.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		model.Name,
		model.DisplayName,
		model.Description,
		model.ModelType,
		model.ContextLength,
		model.MaxTokens,
		model.SupportsStreaming,
		model.SupportsFunctions,
		model.Status,
		model.UpdatedAt,
		model.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return entities.ErrModelNotFound
	}
	
	return nil
}

// Delete 删除模型
func (r *modelRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM models WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return entities.ErrModelNotFound
	}
	
	return nil
}

// List 获取模型列表
func (r *modelRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.Model, error) {
	query := `
		SELECT id, provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at
		FROM models 
		ORDER BY provider_id ASC, name ASC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()
	
	var models []*entities.Model
	for rows.Next() {
		model := &entities.Model{}
		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.Name,
			&model.Slug,
			&model.DisplayName,
			&model.Description,
			&model.ModelType,
			&model.ContextLength,
			&model.MaxTokens,
			&model.SupportsStreaming,
			&model.SupportsFunctions,
			&model.Status,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, model)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate models: %w", err)
	}
	
	return models, nil
}

// Count 获取模型总数
func (r *modelRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM models`
	
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count models: %w", err)
	}
	
	return count, nil
}

// GetActiveModels 获取活跃的模型列表
func (r *modelRepositoryImpl) GetActiveModels(ctx context.Context) ([]*entities.Model, error) {
	query := `
		SELECT id, provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at
		FROM models 
		WHERE status = ?
		ORDER BY provider_id ASC, name ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, entities.ModelStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get active models: %w", err)
	}
	defer rows.Close()
	
	var models []*entities.Model
	for rows.Next() {
		model := &entities.Model{}
		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.Name,
			&model.Slug,
			&model.DisplayName,
			&model.Description,
			&model.ModelType,
			&model.ContextLength,
			&model.MaxTokens,
			&model.SupportsStreaming,
			&model.SupportsFunctions,
			&model.Status,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, model)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate models: %w", err)
	}
	
	return models, nil
}

// GetModelsByType 根据类型获取模型列表
func (r *modelRepositoryImpl) GetModelsByType(ctx context.Context, modelType entities.ModelType) ([]*entities.Model, error) {
	query := `
		SELECT id, provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at
		FROM models 
		WHERE model_type = ? AND status = ?
		ORDER BY provider_id ASC, name ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, modelType, entities.ModelStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get models by type: %w", err)
	}
	defer rows.Close()
	
	var models []*entities.Model
	for rows.Next() {
		model := &entities.Model{}
		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.Name,
			&model.Slug,
			&model.DisplayName,
			&model.Description,
			&model.ModelType,
			&model.ContextLength,
			&model.MaxTokens,
			&model.SupportsStreaming,
			&model.SupportsFunctions,
			&model.Status,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, model)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate models: %w", err)
	}
	
	return models, nil
}

// GetAvailableModels 获取可用的模型列表（活跃且提供商可用）
func (r *modelRepositoryImpl) GetAvailableModels(ctx context.Context, providerID int64) ([]*entities.Model, error) {
	query := `
		SELECT m.id, m.provider_id, m.name, m.slug, m.display_name, m.description, m.model_type, m.context_length, m.max_tokens, m.supports_streaming, m.supports_functions, m.status, m.created_at, m.updated_at
		FROM models m
		INNER JOIN providers p ON m.provider_id = p.id
		WHERE m.provider_id = ? AND m.status = ? AND p.status = ? AND p.health_status = ?
		ORDER BY m.name ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, providerID, entities.ModelStatusActive, entities.ProviderStatusActive, entities.HealthStatusHealthy)
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}
	defer rows.Close()
	
	var models []*entities.Model
	for rows.Next() {
		model := &entities.Model{}
		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.Name,
			&model.Slug,
			&model.DisplayName,
			&model.Description,
			&model.ModelType,
			&model.ContextLength,
			&model.MaxTokens,
			&model.SupportsStreaming,
			&model.SupportsFunctions,
			&model.Status,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, model)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate models: %w", err)
	}
	
	return models, nil
}
