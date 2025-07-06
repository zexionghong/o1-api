package repositories

import (
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type toolRepositoryImpl struct {
	db *sqlx.DB
}

// NewToolRepository 创建工具仓库实现
func NewToolRepository(db *sqlx.DB) repositories.ToolRepository {
	return &toolRepositoryImpl{db: db}
}

// GetTools 获取所有工具模板
func (r *toolRepositoryImpl) GetTools(ctx context.Context) ([]*entities.Tool, error) {
	query := `
		SELECT id, name, description, category, icon, color, config_schema, is_active, created_at, updated_at
		FROM tools
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tools: %w", err)
	}
	defer rows.Close()

	var tools []*entities.Tool
	for rows.Next() {
		var tool entities.Tool
		var configSchema string

		err := rows.Scan(
			&tool.ID, &tool.Name, &tool.Description, &tool.Category, &tool.Icon, &tool.Color,
			&configSchema, &tool.IsActive, &tool.CreatedAt, &tool.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool: %w", err)
		}

		// 转换JSON字符串为RawMessage
		tool.ConfigSchema = json.RawMessage(configSchema)

		// 获取支持的模型
		supportedModels, err := r.getToolSupportedModels(ctx, tool.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get supported models for tool %s: %w", tool.ID, err)
		}
		tool.SupportedModels = supportedModels

		tools = append(tools, &tool)
	}

	return tools, nil
}

// GetToolByID 根据ID获取工具模板
func (r *toolRepositoryImpl) GetToolByID(ctx context.Context, id string) (*entities.Tool, error) {
	query := `
		SELECT id, name, description, category, icon, color, config_schema, is_active, created_at, updated_at
		FROM tools
		WHERE id = ? AND is_active = true
	`

	row := r.db.QueryRowxContext(ctx, query, id)

	var tool entities.Tool
	var configSchema string

	err := row.Scan(
		&tool.ID, &tool.Name, &tool.Description, &tool.Category, &tool.Icon, &tool.Color,
		&configSchema, &tool.IsActive, &tool.CreatedAt, &tool.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tool: %w", err)
	}

	// 转换JSON字符串为RawMessage
	tool.ConfigSchema = json.RawMessage(configSchema)

	// 获取支持的模型
	supportedModels, err := r.getToolSupportedModels(ctx, tool.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported models for tool %s: %w", tool.ID, err)
	}
	tool.SupportedModels = supportedModels

	return &tool, nil
}

// CreateUserToolInstance 创建用户工具实例
func (r *toolRepositoryImpl) CreateUserToolInstance(ctx context.Context, instance *entities.UserToolInstance) error {
	query := `
		INSERT INTO user_tool_instances (id, user_id, tool_id, name, description, model_id, api_key_id, config, is_public, share_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	configJSON, err := json.Marshal(instance.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		instance.ID, instance.UserID, instance.ToolID, instance.Name, instance.Description,
		instance.ModelID, instance.APIKeyID, configJSON, instance.IsPublic, instance.ShareToken,
		instance.CreatedAt, instance.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user tool instance: %w", err)
	}

	return nil
}

// GetUserToolInstanceByID 根据ID获取用户工具实例
func (r *toolRepositoryImpl) GetUserToolInstanceByID(ctx context.Context, id string) (*entities.UserToolInstance, error) {
	query := `
		SELECT uti.id, uti.user_id, uti.tool_id, uti.name, uti.description, uti.model_id,
		       uti.api_key_id, uti.config, uti.is_public, uti.share_token, uti.usage_count,
		       uti.created_at, uti.updated_at,
		       t.name as tool_name, t.description as tool_description,
		       t.icon as tool_icon, t.color as tool_color, t.category as tool_category,
		       u.username as creator_username,
		       m.display_name as model_name
		FROM user_tool_instances uti
		LEFT JOIN tools t ON uti.tool_id = t.id
		LEFT JOIN users u ON uti.user_id = u.id
		LEFT JOIN models m ON uti.model_id = m.id
		WHERE uti.id = ?
	`

	row := r.db.QueryRowxContext(ctx, query, id)

	var instance entities.UserToolInstance
	var toolName, toolDesc, toolIcon, toolColor, toolCategory sql.NullString
	var creatorUsername sql.NullString
	var modelName sql.NullString

	err := row.Scan(
		&instance.ID, &instance.UserID, &instance.ToolID, &instance.Name, &instance.Description,
		&instance.ModelID, &instance.APIKeyID, &instance.Config, &instance.IsPublic, &instance.ShareToken,
		&instance.UsageCount, &instance.CreatedAt, &instance.UpdatedAt,
		&toolName, &toolDesc, &toolIcon, &toolColor, &toolCategory,
		&creatorUsername, &modelName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user tool instance: %w", err)
	}

	// 填充关联数据
	if toolName.Valid {
		instance.Tool = &entities.Tool{
			ID:          instance.ToolID,
			Name:        toolName.String,
			Description: toolDesc.String,
			Icon:        toolIcon.String,
			Color:       toolColor.String,
			Category:    toolCategory.String,
		}
	}

	if creatorUsername.Valid {
		instance.Creator = &entities.User{
			ID:       instance.UserID,
			Username: creatorUsername.String,
		}
	}

	if modelName.Valid {
		instance.ModelName = modelName.String
	}

	return &instance, nil
}

// GetUserToolInstancesByUserID 获取用户的工具实例列表
func (r *toolRepositoryImpl) GetUserToolInstancesByUserID(ctx context.Context, userID int64, includePrivate bool) ([]*entities.UserToolInstance, error) {
	query := `
		SELECT uti.id, uti.user_id, uti.tool_id, uti.name, uti.description, uti.model_id,
		       uti.api_key_id, uti.config, uti.is_public, uti.share_token, uti.usage_count,
		       uti.created_at, uti.updated_at,
		       t.name as tool_name, t.description as tool_description,
		       t.icon as tool_icon, t.color as tool_color, t.category as tool_category,
		       u.username as creator_username,
		       m.display_name as model_name
		FROM user_tool_instances uti
		LEFT JOIN tools t ON uti.tool_id = t.id
		LEFT JOIN users u ON uti.user_id = u.id
		LEFT JOIN models m ON uti.model_id = m.id
		WHERE uti.user_id = ?
	`

	args := []interface{}{userID}
	if !includePrivate {
		query += " AND uti.is_public = true"
	}

	query += " ORDER BY uti.created_at DESC"

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tool instances: %w", err)
	}
	defer rows.Close()

	var instances []*entities.UserToolInstance
	for rows.Next() {
		var instance entities.UserToolInstance
		var toolName, toolDesc, toolIcon, toolColor, toolCategory sql.NullString
		var creatorUsername sql.NullString
		var modelName sql.NullString

		err := rows.Scan(
			&instance.ID, &instance.UserID, &instance.ToolID, &instance.Name, &instance.Description,
			&instance.ModelID, &instance.APIKeyID, &instance.Config, &instance.IsPublic, &instance.ShareToken,
			&instance.UsageCount, &instance.CreatedAt, &instance.UpdatedAt,
			&toolName, &toolDesc, &toolIcon, &toolColor, &toolCategory,
			&creatorUsername, &modelName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user tool instance: %w", err)
		}

		// 填充关联数据
		if toolName.Valid {
			instance.Tool = &entities.Tool{
				ID:          instance.ToolID,
				Name:        toolName.String,
				Description: toolDesc.String,
				Icon:        toolIcon.String,
				Color:       toolColor.String,
				Category:    toolCategory.String,
			}
		}

		if creatorUsername.Valid {
			instance.Creator = &entities.User{
				ID:       instance.UserID,
				Username: creatorUsername.String,
			}
		}

		if modelName.Valid {
			instance.ModelName = modelName.String
		}

		instances = append(instances, &instance)
	}

	return instances, nil
}

// GetPublicUserToolInstances 获取公开的用户工具实例
func (r *toolRepositoryImpl) GetPublicUserToolInstances(ctx context.Context, limit, offset int) ([]*entities.UserToolInstance, error) {
	query := `
		SELECT uti.id, uti.user_id, uti.tool_id, uti.name, uti.description, uti.model_id,
		       uti.api_key_id, uti.config, uti.is_public, uti.share_token, uti.usage_count,
		       uti.created_at, uti.updated_at,
		       t.name as tool_name, t.description as tool_description,
		       t.icon as tool_icon, t.color as tool_color, t.category as tool_category,
		       u.username as creator_username,
		       m.display_name as model_name
		FROM user_tool_instances uti
		LEFT JOIN tools t ON uti.tool_id = t.id
		LEFT JOIN users u ON uti.user_id = u.id
		LEFT JOIN models m ON uti.model_id = m.id
		WHERE uti.is_public = true
		ORDER BY uti.usage_count DESC, uti.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get public tool instances: %w", err)
	}
	defer rows.Close()

	var instances []*entities.UserToolInstance
	for rows.Next() {
		var instance entities.UserToolInstance
		var toolName, toolDesc, toolIcon, toolColor, toolCategory sql.NullString
		var creatorUsername sql.NullString
		var modelName sql.NullString

		err := rows.Scan(
			&instance.ID, &instance.UserID, &instance.ToolID, &instance.Name, &instance.Description,
			&instance.ModelID, &instance.APIKeyID, &instance.Config, &instance.IsPublic, &instance.ShareToken,
			&instance.UsageCount, &instance.CreatedAt, &instance.UpdatedAt,
			&toolName, &toolDesc, &toolIcon, &toolColor, &toolCategory,
			&creatorUsername, &modelName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan public tool instance: %w", err)
		}

		// 填充关联数据
		if toolName.Valid {
			instance.Tool = &entities.Tool{
				ID:          instance.ToolID,
				Name:        toolName.String,
				Description: toolDesc.String,
				Icon:        toolIcon.String,
				Color:       toolColor.String,
				Category:    toolCategory.String,
			}
		}

		if creatorUsername.Valid {
			instance.Creator = &entities.User{
				ID:       instance.UserID,
				Username: creatorUsername.String,
			}
		}

		if modelName.Valid {
			instance.ModelName = modelName.String
		}

		instances = append(instances, &instance)
	}

	return instances, nil
}

// GetUserToolInstanceByShareToken 根据分享token获取工具实例
func (r *toolRepositoryImpl) GetUserToolInstanceByShareToken(ctx context.Context, shareToken string) (*entities.UserToolInstance, error) {
	query := `
		SELECT uti.id, uti.user_id, uti.tool_id, uti.name, uti.description, uti.model_id, 
		       uti.api_key_id, uti.config, uti.is_public, uti.share_token, uti.usage_count,
		       uti.created_at, uti.updated_at,
		       t.name as tool_name, t.description as tool_description,
		       t.icon as tool_icon, t.color as tool_color, t.category as tool_category,
		       u.username as creator_username
		FROM user_tool_instances uti
		LEFT JOIN tools t ON uti.tool_id = t.id
		LEFT JOIN users u ON uti.user_id = u.id
		WHERE uti.share_token = ?
	`

	row := r.db.QueryRowxContext(ctx, query, shareToken)

	var instance entities.UserToolInstance
	var toolName, toolDesc, toolIcon, toolColor, toolCategory sql.NullString
	var creatorUsername sql.NullString

	err := row.Scan(
		&instance.ID, &instance.UserID, &instance.ToolID, &instance.Name, &instance.Description,
		&instance.ModelID, &instance.APIKeyID, &instance.Config, &instance.IsPublic, &instance.ShareToken,
		&instance.UsageCount, &instance.CreatedAt, &instance.UpdatedAt,
		&toolName, &toolDesc, &toolIcon, &toolColor, &toolCategory,
		&creatorUsername,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shared tool instance: %w", err)
	}

	// 填充关联数据
	if toolName.Valid {
		instance.Tool = &entities.Tool{
			ID:          instance.ToolID,
			Name:        toolName.String,
			Description: toolDesc.String,
			Icon:        toolIcon.String,
			Color:       toolColor.String,
			Category:    toolCategory.String,
		}
	}

	if creatorUsername.Valid {
		instance.Creator = &entities.User{
			ID:       instance.UserID,
			Username: creatorUsername.String,
		}
	}

	return &instance, nil
}

// UpdateUserToolInstance 更新用户工具实例
func (r *toolRepositoryImpl) UpdateUserToolInstance(ctx context.Context, instance *entities.UserToolInstance) error {
	query := `
		UPDATE user_tool_instances
		SET name = ?, description = ?, model_id = ?, api_key_id = ?, config = ?,
		    is_public = ?, share_token = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`

	configJSON, err := json.Marshal(instance.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		instance.Name, instance.Description, instance.ModelID, instance.APIKeyID, configJSON,
		instance.IsPublic, instance.ShareToken, instance.UpdatedAt, instance.ID, instance.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user tool instance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tool instance not found or not owned by user")
	}

	return nil
}

// DeleteUserToolInstance 删除用户工具实例
func (r *toolRepositoryImpl) DeleteUserToolInstance(ctx context.Context, id string, userID int64) error {
	query := `DELETE FROM user_tool_instances WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user tool instance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tool instance not found or not owned by user")
	}

	return nil
}

// IncrementUsageCount 增加使用次数
func (r *toolRepositoryImpl) IncrementUsageCount(ctx context.Context, instanceID string) error {
	query := `UPDATE user_tool_instances SET usage_count = usage_count + 1 WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, instanceID)
	if err != nil {
		return fmt.Errorf("failed to increment usage count: %w", err)
	}

	return nil
}

// CreateUsageLog 创建使用记录
func (r *toolRepositoryImpl) CreateUsageLog(ctx context.Context, log *entities.ToolUsageLog) error {
	query := `
		INSERT INTO tool_usage_logs (tool_instance_id, user_id, session_id, request_count, tokens_used, cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ToolInstanceID, log.UserID, log.SessionID, log.RequestCount, log.TokensUsed, log.Cost, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create usage log: %w", err)
	}

	return nil
}

// GetUsageLogsByInstanceID 获取工具实例使用记录
func (r *toolRepositoryImpl) GetUsageLogsByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*entities.ToolUsageLog, error) {
	query := `
		SELECT id, tool_instance_id, user_id, session_id, request_count, tokens_used, cost, created_at
		FROM tool_usage_logs
		WHERE tool_instance_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	var logs []*entities.ToolUsageLog
	if err := r.db.SelectContext(ctx, &logs, query, instanceID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to get usage logs: %w", err)
	}

	return logs, nil
}

// getToolSupportedModels 获取工具支持的模型列表
func (r *toolRepositoryImpl) getToolSupportedModels(ctx context.Context, toolID string) ([]entities.Model, error) {
	query := `
		SELECT m.id, m.name, m.slug, m.display_name, m.description, m.model_type,
		       m.context_length, m.max_tokens, m.supports_streaming, m.supports_functions,
		       m.status, m.created_at, m.updated_at
		FROM models m
		INNER JOIN tool_model_support tms ON m.id = tms.model_id
		WHERE tms.tool_id = ? AND m.status = 'active'
		ORDER BY m.display_name
	`

	rows, err := r.db.QueryxContext(ctx, query, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to query supported models: %w", err)
	}
	defer rows.Close()

	var models []entities.Model
	for rows.Next() {
		var model entities.Model

		err := rows.Scan(
			&model.ID, &model.Name, &model.Slug, &model.DisplayName, &model.Description,
			&model.ModelType, &model.ContextLength, &model.MaxTokens, &model.SupportsStreaming,
			&model.SupportsFunctions, &model.Status, &model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}

		models = append(models, model)
	}

	return models, nil
}
