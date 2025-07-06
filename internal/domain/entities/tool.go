package entities

import (
	"encoding/json"
	"time"
)

// Tool 工具模板
type Tool struct {
	ID           string          `json:"id" db:"id"`
	Name         string          `json:"name" db:"name"`
	Description  string          `json:"description" db:"description"`
	Category     string          `json:"category" db:"category"`
	Icon         string          `json:"icon" db:"icon"`
	Color        string          `json:"color" db:"color"`
	ConfigSchema json.RawMessage `json:"config_schema" db:"config_schema"`
	IsActive     bool            `json:"is_active" db:"is_active"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`

	// 关联数据
	SupportedModels []Model `json:"supported_models,omitempty"`
}

// GetSupportedModelNames 获取支持的模型名称列表
func (t *Tool) GetSupportedModelNames() []string {
	var names []string
	for _, model := range t.SupportedModels {
		names = append(names, model.Name)
	}
	return names
}

// UserToolInstance 用户工具实例
type UserToolInstance struct {
	ID          string          `json:"id" db:"id"`
	UserID      int64           `json:"user_id" db:"user_id"`
	ToolID      string          `json:"tool_id" db:"tool_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	ModelID     int64           `json:"model_id" db:"model_id"`
	APIKeyID    int64           `json:"api_key_id" db:"api_key_id"`
	Config      json.RawMessage `json:"config" db:"config"`
	IsPublic    bool            `json:"is_public" db:"is_public"`
	ShareToken  *string         `json:"share_token,omitempty" db:"share_token"`
	UsageCount  int64           `json:"usage_count" db:"usage_count"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`

	// 关联数据
	Tool      *Tool   `json:"tool,omitempty"`
	Creator   *User   `json:"creator,omitempty"`
	APIKey    *APIKey `json:"api_key,omitempty"`
	ModelName string  `json:"model_name,omitempty"`
}

// GetConfig 获取工具配置
func (t *UserToolInstance) GetConfig() (map[string]interface{}, error) {
	var config map[string]interface{}
	if err := json.Unmarshal(t.Config, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// GetShareURL 获取分享链接
func (t *UserToolInstance) GetShareURL(baseURL string) string {
	if t.ShareToken == nil {
		return ""
	}
	return baseURL + "/tools/share/" + *t.ShareToken
}

// ToolUsageLog 工具使用记录
type ToolUsageLog struct {
	ID             int64     `json:"id" db:"id"`
	ToolInstanceID string    `json:"tool_instance_id" db:"tool_instance_id"`
	UserID         *int64    `json:"user_id,omitempty" db:"user_id"`
	SessionID      string    `json:"session_id" db:"session_id"`
	RequestCount   int       `json:"request_count" db:"request_count"`
	TokensUsed     int       `json:"tokens_used" db:"tokens_used"`
	Cost           float64   `json:"cost" db:"cost"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// CreateUserToolInstanceRequest 创建用户工具实例请求
type CreateUserToolInstanceRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=100"`
	Description string                 `json:"description" validate:"max=500"`
	ToolID      string                 `json:"tool_id" validate:"required"`
	ModelID     int64                  `json:"model_id" validate:"required"`
	APIKeyID    int64                  `json:"api_key_id" validate:"required"`
	Config      map[string]interface{} `json:"config"`
	IsPublic    bool                   `json:"is_public"`
}

// UpdateUserToolInstanceRequest 更新用户工具实例请求
type UpdateUserToolInstanceRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=500"`
	ModelID     *int64                 `json:"model_id,omitempty"`
	APIKeyID    *int64                 `json:"api_key_id,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	IsPublic    *bool                  `json:"is_public,omitempty"`
}
