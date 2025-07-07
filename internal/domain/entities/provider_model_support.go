package entities

import (
	"encoding/json"
	"time"
)

// ProviderModelSupport 提供商模型支持实体
type ProviderModelSupport struct {
	ID                int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProviderID        int64     `json:"provider_id" gorm:"not null;index"`
	ModelSlug         string    `json:"model_slug" gorm:"not null;size:100;index"`
	UpstreamModelName *string   `json:"upstream_model_name,omitempty" gorm:"size:100"`
	Enabled           bool      `json:"enabled" gorm:"not null;default:true;index"`
	Priority          int       `json:"priority" gorm:"not null;default:1"`
	Config            *string   `json:"config,omitempty" gorm:"type:text"`
	CreatedAt         time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

// TableName 指定表名
func (ProviderModelSupport) TableName() string {
	return "provider_model_support"
}

// ProviderModelConfig 提供商模型配置
type ProviderModelConfig struct {
	ParameterMapping map[string]string `json:"parameter_mapping,omitempty"` // 参数映射
	MaxTokens        *int              `json:"max_tokens,omitempty"`        // 最大token限制
	Temperature      *float64          `json:"temperature,omitempty"`       // 默认温度
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`    // 自定义请求头
	Endpoint         *string           `json:"endpoint,omitempty"`          // 自定义端点
}

// IsEnabled 检查是否启用
func (pms *ProviderModelSupport) IsEnabled() bool {
	return pms.Enabled
}

// GetUpstreamModelName 获取上游模型名称
func (pms *ProviderModelSupport) GetUpstreamModelName() string {
	if pms.UpstreamModelName != nil && *pms.UpstreamModelName != "" {
		return *pms.UpstreamModelName
	}
	return pms.ModelSlug // 默认使用model_slug
}

// GetConfig 获取配置
func (pms *ProviderModelSupport) GetConfig() (*ProviderModelConfig, error) {
	if pms.Config == nil || *pms.Config == "" {
		return &ProviderModelConfig{}, nil
	}

	var config ProviderModelConfig
	if err := json.Unmarshal([]byte(*pms.Config), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SetConfig 设置配置
func (pms *ProviderModelSupport) SetConfig(config *ProviderModelConfig) error {
	if config == nil {
		pms.Config = nil
		return nil
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configStr := string(data)
	pms.Config = &configStr
	return nil
}

// ModelSupportInfo 模型支持信息（用于查询结果）
type ModelSupportInfo struct {
	Provider          *Provider             `json:"provider"`
	ModelSlug         string                `json:"model_slug"`
	UpstreamModelName string                `json:"upstream_model_name"`
	Priority          int                   `json:"priority"`
	Enabled           bool                  `json:"enabled"`
	Config            *ProviderModelConfig  `json:"config,omitempty"`
	Support           *ProviderModelSupport `json:"support"`
}

// IsAvailable 检查模型支持是否可用
func (msi *ModelSupportInfo) IsAvailable() bool {
	return msi.Enabled &&
		msi.Provider != nil &&
		msi.Provider.IsAvailable()
}
