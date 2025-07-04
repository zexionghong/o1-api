package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// apiKeyRepositoryImpl API密钥仓储实现
type apiKeyRepositoryImpl struct {
	db *sql.DB
}

// NewAPIKeyRepository 创建API密钥仓储
func NewAPIKeyRepository(db *sql.DB) repositories.APIKeyRepository {
	return &apiKeyRepositoryImpl{
		db: db,
	}
}

// Create 创建API密钥
func (r *apiKeyRepositoryImpl) Create(ctx context.Context, apiKey *entities.APIKey) error {
	query := `
		INSERT INTO api_keys (user_id, key, key_prefix, name, status, permissions, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	apiKey.CreatedAt = now
	apiKey.UpdatedAt = now

	// 序列化权限
	permissionsJSON, err := apiKey.MarshalPermissions()
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		apiKey.UserID,
		apiKey.Key,
		apiKey.KeyPrefix,
		apiKey.Name,
		apiKey.Status,
		permissionsJSON,
		apiKey.ExpiresAt,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create api key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	apiKey.ID = id
	return nil
}

// GetByID 根据ID获取API密钥
func (r *apiKeyRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.APIKey, error) {
	query := `
		SELECT id, user_id, key, key_prefix, name, status, permissions, expires_at, last_used_at, created_at, updated_at
		FROM api_keys WHERE id = ?
	`

	apiKey := &entities.APIKey{}
	var permissionsJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Key,
		&apiKey.KeyPrefix,
		&apiKey.Name,
		&apiKey.Status,
		&permissionsJSON,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("failed to get api key by id: %w", err)
	}

	// 反序列化权限
	if permissionsJSON.Valid {
		if err := apiKey.UnmarshalPermissions(permissionsJSON.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	return apiKey, nil
}

// GetByKey 根据密钥获取API密钥
func (r *apiKeyRepositoryImpl) GetByKey(ctx context.Context, key string) (*entities.APIKey, error) {
	query := `
		SELECT id, user_id, key, key_prefix, name, status, permissions, expires_at, last_used_at, created_at, updated_at
		FROM api_keys WHERE key = ?
	`

	apiKey := &entities.APIKey{}
	var permissionsJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Key,
		&apiKey.KeyPrefix,
		&apiKey.Name,
		&apiKey.Status,
		&permissionsJSON,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("failed to get api key by key: %w", err)
	}

	// 反序列化权限
	if permissionsJSON.Valid {
		if err := apiKey.UnmarshalPermissions(permissionsJSON.String); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	return apiKey, nil
}

// GetByUserID 根据用户ID获取API密钥列表
func (r *apiKeyRepositoryImpl) GetByUserID(ctx context.Context, userID int64) ([]*entities.APIKey, error) {
	query := `
		SELECT id, user_id, key, key_prefix, name, status, permissions, expires_at, last_used_at, created_at, updated_at
		FROM api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get api keys by user id: %w", err)
	}
	defer rows.Close()

	var apiKeys []*entities.APIKey
	for rows.Next() {
		apiKey := &entities.APIKey{}
		var permissionsJSON sql.NullString

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Key,
			&apiKey.KeyPrefix,
			&apiKey.Name,
			&apiKey.Status,
			&permissionsJSON,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}

		// 反序列化权限
		if permissionsJSON.Valid {
			if err := apiKey.UnmarshalPermissions(permissionsJSON.String); err != nil {
				return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
			}
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate api keys: %w", err)
	}

	return apiKeys, nil
}

// Update 更新API密钥
func (r *apiKeyRepositoryImpl) Update(ctx context.Context, apiKey *entities.APIKey) error {
	query := `
		UPDATE api_keys 
		SET name = ?, status = ?, permissions = ?, expires_at = ?, last_used_at = ?, updated_at = ?
		WHERE id = ?
	`

	apiKey.UpdatedAt = time.Now()

	// 序列化权限
	permissionsJSON, err := apiKey.MarshalPermissions()
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		apiKey.Name,
		apiKey.Status,
		permissionsJSON,
		apiKey.ExpiresAt,
		apiKey.LastUsedAt,
		apiKey.UpdatedAt,
		apiKey.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update api key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAPIKeyNotFound
	}

	return nil
}

// UpdateLastUsed 更新最后使用时间
func (r *apiKeyRepositoryImpl) UpdateLastUsed(ctx context.Context, id int64) error {
	query := `
		UPDATE api_keys 
		SET last_used_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAPIKeyNotFound
	}

	return nil
}

// UpdateStatus 更新状态
func (r *apiKeyRepositoryImpl) UpdateStatus(ctx context.Context, id int64, status entities.APIKeyStatus) error {
	query := `
		UPDATE api_keys 
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAPIKeyNotFound
	}

	return nil
}

// Delete 删除API密钥
func (r *apiKeyRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM api_keys WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAPIKeyNotFound
	}

	return nil
}

// List 获取API密钥列表
func (r *apiKeyRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.APIKey, error) {
	query := `
		SELECT id, user_id, key, key_prefix, name, status, permissions, expires_at, last_used_at, created_at, updated_at
		FROM api_keys
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*entities.APIKey
	for rows.Next() {
		apiKey := &entities.APIKey{}
		var permissionsJSON sql.NullString

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Key,
			&apiKey.KeyPrefix,
			&apiKey.Name,
			&apiKey.Status,
			&permissionsJSON,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}

		// 反序列化权限
		if permissionsJSON.Valid {
			if err := apiKey.UnmarshalPermissions(permissionsJSON.String); err != nil {
				return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
			}
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate api keys: %w", err)
	}

	return apiKeys, nil
}

// Count 获取API密钥总数
func (r *apiKeyRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM api_keys`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count api keys: %w", err)
	}

	return count, nil
}

// GetActiveKeys 获取活跃的API密钥列表
func (r *apiKeyRepositoryImpl) GetActiveKeys(ctx context.Context, userID int64) ([]*entities.APIKey, error) {
	query := `
		SELECT id, user_id, key, key_prefix, name, status, permissions, expires_at, last_used_at, created_at, updated_at
		FROM api_keys
		WHERE user_id = ? AND status = ? AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, entities.APIKeyStatusActive, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get active keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*entities.APIKey
	for rows.Next() {
		apiKey := &entities.APIKey{}
		var permissionsJSON sql.NullString

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Key,
			&apiKey.KeyPrefix,
			&apiKey.Name,
			&apiKey.Status,
			&permissionsJSON,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}

		// 反序列化权限
		if permissionsJSON.Valid {
			if err := apiKey.UnmarshalPermissions(permissionsJSON.String); err != nil {
				return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
			}
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate api keys: %w", err)
	}

	return apiKeys, nil
}

// GetExpiredKeys 获取过期的API密钥列表
func (r *apiKeyRepositoryImpl) GetExpiredKeys(ctx context.Context, limit int) ([]*entities.APIKey, error) {
	query := `
		SELECT id, user_id, key, key_prefix, name, status, permissions, expires_at, last_used_at, created_at, updated_at
		FROM api_keys
		WHERE expires_at IS NOT NULL AND expires_at <= ? AND status = ?
		ORDER BY expires_at ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now(), entities.APIKeyStatusActive, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*entities.APIKey
	for rows.Next() {
		apiKey := &entities.APIKey{}
		var permissionsJSON sql.NullString

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Key,
			&apiKey.KeyPrefix,
			&apiKey.Name,
			&apiKey.Status,
			&permissionsJSON,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}

		// 反序列化权限
		if permissionsJSON.Valid {
			if err := apiKey.UnmarshalPermissions(permissionsJSON.String); err != nil {
				return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
			}
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate api keys: %w", err)
	}

	return apiKeys, nil
}
