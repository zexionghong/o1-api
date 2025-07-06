package services

import (
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ToolService 工具服务
type ToolService struct {
	toolRepo   repositories.ToolRepository
	apiKeyRepo repositories.APIKeyRepository
	db         *sqlx.DB
}

// NewToolService 创建工具服务
func NewToolService(toolRepo repositories.ToolRepository, apiKeyRepo repositories.APIKeyRepository, db *sqlx.DB) *ToolService {
	return &ToolService{
		toolRepo:   toolRepo,
		apiKeyRepo: apiKeyRepo,
		db:         db,
	}
}

// GetTools 获取所有工具模板
func (s *ToolService) GetTools(ctx context.Context) ([]*entities.Tool, error) {
	return s.toolRepo.GetTools(ctx)
}

// GetToolByID 根据ID获取工具模板
func (s *ToolService) GetToolByID(ctx context.Context, id string) (*entities.Tool, error) {
	return s.toolRepo.GetToolByID(ctx, id)
}

// CreateUserToolInstance 创建用户工具实例
func (s *ToolService) CreateUserToolInstance(ctx context.Context, userID int64, req *entities.CreateUserToolInstanceRequest) (*entities.UserToolInstance, error) {
	// 验证工具模板是否存在
	tool, err := s.toolRepo.GetToolByID(ctx, req.ToolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool: %w", err)
	}
	if tool == nil {
		return nil, fmt.Errorf("tool not found")
	}

	// 验证模型是否被工具支持
	modelSupported := false
	for _, model := range tool.SupportedModels {
		if model.ID == req.ModelID {
			modelSupported = true
			break
		}
	}
	if !modelSupported {
		return nil, fmt.Errorf("model %d is not supported by tool %s", req.ModelID, req.ToolID)
	}

	// 验证API Key是否属于用户且有效
	apiKey, err := s.apiKeyRepo.GetByID(ctx, req.APIKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	if apiKey == nil || apiKey.UserID != userID {
		return nil, fmt.Errorf("API key not found or not owned by user")
	}
	if apiKey.Status != "active" {
		return nil, fmt.Errorf("API key is not active")
	}

	// 创建工具实例
	instance := &entities.UserToolInstance{
		ID:          uuid.New().String(),
		UserID:      userID,
		ToolID:      req.ToolID,
		Name:        req.Name,
		Description: req.Description,
		ModelID:     req.ModelID,
		APIKeyID:    req.APIKeyID,
		IsPublic:    req.IsPublic,
		UsageCount:  0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 处理配置
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		instance.Config = configJSON
	} else {
		instance.Config = json.RawMessage("{}")
	}

	// 如果是公开工具，生成分享token
	if req.IsPublic {
		shareToken, err := s.generateShareToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate share token: %w", err)
		}
		instance.ShareToken = &shareToken
	}

	// 保存到数据库
	if err := s.toolRepo.CreateUserToolInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to create user tool instance: %w", err)
	}

	// 获取完整的工具实例信息（包含关联数据）
	return s.toolRepo.GetUserToolInstanceByID(ctx, instance.ID)
}

// GetUserToolInstances 获取用户的工具实例列表
func (s *ToolService) GetUserToolInstances(ctx context.Context, userID int64, category string) ([]*entities.UserToolInstance, error) {
	switch category {
	case "my_tools":
		return s.toolRepo.GetUserToolInstancesByUserID(ctx, userID, true)
	case "public":
		return s.toolRepo.GetPublicUserToolInstances(ctx, 100, 0)
	case "shared":
		// 获取用户的公开工具（有分享链接的）
		instances, err := s.toolRepo.GetUserToolInstancesByUserID(ctx, userID, true)
		if err != nil {
			return nil, err
		}
		var sharedInstances []*entities.UserToolInstance
		for _, instance := range instances {
			if instance.ShareToken != nil {
				sharedInstances = append(sharedInstances, instance)
			}
		}
		return sharedInstances, nil
	default:
		// 默认返回用户的所有工具实例
		return s.toolRepo.GetUserToolInstancesByUserID(ctx, userID, true)
	}
}

// GetUserToolInstanceByID 获取用户工具实例详情
func (s *ToolService) GetUserToolInstanceByID(ctx context.Context, id string, userID int64) (*entities.UserToolInstance, error) {
	instance, err := s.toolRepo.GetUserToolInstanceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tool instance: %w", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("tool instance not found")
	}

	// 检查权限：工具所有者或公开工具
	if instance.UserID != userID && !instance.IsPublic {
		return nil, fmt.Errorf("access denied")
	}

	return instance, nil
}

// UpdateUserToolInstance 更新用户工具实例
func (s *ToolService) UpdateUserToolInstance(ctx context.Context, id string, userID int64, req *entities.UpdateUserToolInstanceRequest) (*entities.UserToolInstance, error) {
	// 获取现有工具实例
	instance, err := s.toolRepo.GetUserToolInstanceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tool instance: %w", err)
	}
	if instance == nil {
		return nil, fmt.Errorf("tool instance not found")
	}
	if instance.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// 更新字段
	if req.Name != nil {
		instance.Name = *req.Name
	}
	if req.Description != nil {
		instance.Description = *req.Description
	}
	if req.ModelID != nil {
		// 验证新模型是否被工具支持
		tool, err := s.toolRepo.GetToolByID(ctx, instance.ToolID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tool: %w", err)
		}

		modelSupported := false
		for _, model := range tool.SupportedModels {
			if model.ID == *req.ModelID {
				modelSupported = true
				break
			}
		}
		if !modelSupported {
			return nil, fmt.Errorf("model %d is not supported by tool %s", *req.ModelID, instance.ToolID)
		}

		instance.ModelID = *req.ModelID
	}
	if req.APIKeyID != nil {
		// 验证新API Key是否属于用户且有效
		apiKey, err := s.apiKeyRepo.GetByID(ctx, *req.APIKeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get API key: %w", err)
		}
		if apiKey == nil || apiKey.UserID != userID {
			return nil, fmt.Errorf("API key not found or not owned by user")
		}
		if apiKey.Status != "active" {
			return nil, fmt.Errorf("API key is not active")
		}

		instance.APIKeyID = *req.APIKeyID
	}
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		instance.Config = configJSON
	}
	if req.IsPublic != nil {
		instance.IsPublic = *req.IsPublic

		// 如果变为公开且没有分享token，生成一个
		if *req.IsPublic && instance.ShareToken == nil {
			shareToken, err := s.generateShareToken()
			if err != nil {
				return nil, fmt.Errorf("failed to generate share token: %w", err)
			}
			instance.ShareToken = &shareToken
		}
		// 如果变为私有，清除分享token
		if !*req.IsPublic {
			instance.ShareToken = nil
		}
	}

	instance.UpdatedAt = time.Now()

	// 保存更新
	if err := s.toolRepo.UpdateUserToolInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to update user tool instance: %w", err)
	}

	// 返回更新后的工具实例信息
	return s.toolRepo.GetUserToolInstanceByID(ctx, id)
}

// DeleteUserToolInstance 删除用户工具实例
func (s *ToolService) DeleteUserToolInstance(ctx context.Context, id string, userID int64) error {
	return s.toolRepo.DeleteUserToolInstance(ctx, id, userID)
}

// GetSharedToolInstance 获取分享的工具实例
func (s *ToolService) GetSharedToolInstance(ctx context.Context, shareToken string) (*entities.UserToolInstance, error) {
	return s.toolRepo.GetUserToolInstanceByShareToken(ctx, shareToken)
}

// IncrementUsageCount 增加工具实例使用次数
func (s *ToolService) IncrementUsageCount(ctx context.Context, instanceID string) error {
	return s.toolRepo.IncrementUsageCount(ctx, instanceID)
}

// GetAvailableModels 获取可用模型列表
func (s *ToolService) GetAvailableModels(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, display_name, model_type, status
		FROM models
		WHERE status = 'active' AND model_type = 'chat'
		ORDER BY display_name
	`

	rows, err := s.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query models: %w", err)
	}
	defer rows.Close()

	var models []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, displayName, modelType, status string

		err := rows.Scan(&id, &name, &displayName, &modelType, &status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}

		models = append(models, map[string]interface{}{
			"id":           id,
			"name":         name,
			"display_name": displayName,
			"model_type":   modelType,
			"status":       status,
		})
	}

	return models, nil
}

// GetUserAPIKeys 获取用户API密钥列表
func (s *ToolService) GetUserAPIKeys(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, key_prefix, status
		FROM api_keys
		WHERE user_id = ? AND status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, keyPrefix, status string

		err := rows.Scan(&id, &name, &keyPrefix, &status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		apiKeys = append(apiKeys, map[string]interface{}{
			"id":         id,
			"name":       name,
			"key_prefix": keyPrefix,
			"status":     status,
		})
	}

	return apiKeys, nil
}

// generateShareToken 生成分享token
func (s *ToolService) generateShareToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
