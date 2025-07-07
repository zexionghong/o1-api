package entities

import (
	"encoding/json"
	"time"
)

// Tool 工具模板
type Tool struct {
	ID           string          `json:"id" gorm:"primaryKey;size:50"`
	Name         string          `json:"name" gorm:"not null;size:100"`
	Description  string          `json:"description" gorm:"type:text"`
	Category     string          `json:"category" gorm:"size:50;index"`
	Icon         string          `json:"icon" gorm:"size:100"`
	Color        string          `json:"color" gorm:"size:20"`
	ConfigSchema json.RawMessage `json:"config_schema" gorm:"type:jsonb"`
	IsActive     bool            `json:"is_active" gorm:"default:true;index"`
	CreatedAt    time.Time       `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"not null;autoUpdateTime"`

	// 关联数据
	SupportedModels []Model `json:"supported_models,omitempty" gorm:"-"`
}

// TableName 指定表名
func (Tool) TableName() string {
	return "tools"
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
	ID          string          `json:"id" gorm:"primaryKey;size:36"`
	UserID      int64           `json:"user_id" gorm:"not null;index"`
	ToolID      string          `json:"tool_id" gorm:"not null;size:50;index"`
	Name        string          `json:"name" gorm:"not null;size:100"`
	Description string          `json:"description" gorm:"type:text"`
	ModelID     int64           `json:"-" gorm:"not null;index"` // 隐藏原始字段
	APIKeyID    int64           `json:"-" gorm:"not null;index"` // 隐藏原始字段
	Config      json.RawMessage `json:"config" gorm:"type:jsonb"`
	IsPublic    bool            `json:"is_public" gorm:"default:false;index"`
	ShareToken  *string         `json:"-" gorm:"uniqueIndex;size:32"` // 隐藏原始字段
	UsageCount  int64           `json:"usage_count" gorm:"default:0"`
	CreatedAt   time.Time       `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"not null;autoUpdateTime"`

	// 关联数据
	Tool      *Tool   `json:"tool,omitempty" gorm:"-"`
	Creator   *User   `json:"creator,omitempty" gorm:"-"`
	APIKey    *APIKey `json:"api_key,omitempty" gorm:"-"`
	ModelName string  `json:"model_name,omitempty" gorm:"-"`

	// 前端期望的字段格式
	Type        string  `json:"type,omitempty" gorm:"-"`       // 从 Tool.Category 映射而来
	ModelIDStr  string  `json:"model_id,omitempty" gorm:"-"`   // ModelID 的字符串版本
	APIKeyIDStr string  `json:"api_key_id,omitempty" gorm:"-"` // APIKeyID 的字符串版本
	ShareURL    *string `json:"share_url,omitempty" gorm:"-"`  // 从 ShareToken 生成的完整URL
}

// TableName 指定表名
func (UserToolInstance) TableName() string {
	return "user_tool_instances"
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
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ToolInstanceID string    `json:"tool_instance_id" gorm:"not null;size:36;index"`
	UserID         *int64    `json:"user_id,omitempty" gorm:"index"`
	SessionID      string    `json:"session_id" gorm:"size:64"`
	RequestCount   int       `json:"request_count" gorm:"default:1"`
	TokensUsed     int       `json:"tokens_used" gorm:"default:0"`
	Cost           float64   `json:"cost" gorm:"type:numeric(10,6);default:0"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null;autoCreateTime;index"`
}

// TableName 指定表名
func (ToolUsageLog) TableName() string {
	return "tool_usage_logs"
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
