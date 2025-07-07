package entities

import (
	"time"
)

// ModelType 模型类型枚举
type ModelType string

const (
	ModelTypeChat       ModelType = "chat"
	ModelTypeCompletion ModelType = "completion"
	ModelTypeEmbedding  ModelType = "embedding"
	ModelTypeImage      ModelType = "image"
	ModelTypeAudio      ModelType = "audio"
)

// ModelStatus 模型状态枚举
type ModelStatus string

const (
	ModelStatusActive     ModelStatus = "active"
	ModelStatusDeprecated ModelStatus = "deprecated"
	ModelStatusDisabled   ModelStatus = "disabled"
)

// Model AI模型实体
type Model struct {
	ID                int64       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name              string      `json:"name" gorm:"not null;size:100"`
	Slug              string      `json:"slug" gorm:"uniqueIndex;not null;size:100"`
	DisplayName       *string     `json:"display_name,omitempty" gorm:"size:200"`
	Description       *string     `json:"description,omitempty" gorm:"type:text"`
	ModelType         ModelType   `json:"model_type" gorm:"not null;size:50;index"`
	ContextLength     *int        `json:"context_length,omitempty"`
	MaxTokens         *int        `json:"max_tokens,omitempty"`
	SupportsStreaming bool        `json:"supports_streaming" gorm:"not null;default:false"`
	SupportsFunctions bool        `json:"supports_functions" gorm:"not null;default:false"`
	Status            ModelStatus `json:"status" gorm:"not null;default:active;size:20;index"`
	CreatedAt         time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt         time.Time   `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

// TableName 指定表名
func (Model) TableName() string {
	return "models"
}

// IsAvailable 检查模型是否可用
func (m *Model) IsAvailable() bool {
	return m.Status == ModelStatusActive
}

// IsActive 检查模型是否处于活跃状态
func (m *Model) IsActive() bool {
	return m.Status == ModelStatusActive
}

// GetDisplayName 获取显示名称
func (m *Model) GetDisplayName() string {
	if m.DisplayName != nil && *m.DisplayName != "" {
		return *m.DisplayName
	}
	return m.Name
}

// GetContextLength 获取上下文长度
func (m *Model) GetContextLength() int {
	if m.ContextLength != nil {
		return *m.ContextLength
	}
	return 4096 // 默认值
}

// GetMaxTokens 获取最大token数
func (m *Model) GetMaxTokens() int {
	if m.MaxTokens != nil {
		return *m.MaxTokens
	}
	return m.GetContextLength() / 2 // 默认为上下文长度的一半
}

// CanStream 检查是否支持流式输出
func (m *Model) CanStream() bool {
	return m.SupportsStreaming
}

// CanUseFunctions 检查是否支持函数调用
func (m *Model) CanUseFunctions() bool {
	return m.SupportsFunctions
}

// PricingType 定价类型枚举
type PricingType string

const (
	PricingTypeInput   PricingType = "input"
	PricingTypeOutput  PricingType = "output"
	PricingTypeRequest PricingType = "request"
)

// PricingUnit 定价单位枚举
type PricingUnit string

const (
	PricingUnitToken     PricingUnit = "token"
	PricingUnitRequest   PricingUnit = "request"
	PricingUnitCharacter PricingUnit = "character"
)

// ModelPricing 模型定价实体
type ModelPricing struct {
	ID             int64       `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelID        int64       `json:"model_id" gorm:"not null;index"`
	PricingType    PricingType `json:"pricing_type" gorm:"not null;size:20;index"`
	PricePerUnit   float64     `json:"price_per_unit" gorm:"type:numeric(15,8);not null"`
	Multiplier     float64     `json:"multiplier" gorm:"type:numeric(5,2);not null;default:1.5"` // 价格倍率，默认1.5
	Unit           PricingUnit `json:"unit" gorm:"not null;size:20"`
	Currency       string      `json:"currency" gorm:"not null;default:USD;size:3"`
	EffectiveFrom  time.Time   `json:"effective_from" gorm:"not null;default:CURRENT_TIMESTAMP"`
	EffectiveUntil *time.Time  `json:"effective_until,omitempty"`
	CreatedAt      time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
}

// TableName 指定表名
func (ModelPricing) TableName() string {
	return "model_pricing"
}

// IsEffective 检查定价是否在有效期内
func (mp *ModelPricing) IsEffective(at time.Time) bool {
	if at.Before(mp.EffectiveFrom) {
		return false
	}

	if mp.EffectiveUntil != nil && at.After(*mp.EffectiveUntil) {
		return false
	}

	return true
}

// CalculateCost 计算成本（应用倍率）
func (mp *ModelPricing) CalculateCost(units int) float64 {
	baseCost := float64(units) * mp.PricePerUnit
	return baseCost * mp.Multiplier
}

// CalculateBaseCost 计算基础成本（不应用倍率）
func (mp *ModelPricing) CalculateBaseCost(units int) float64 {
	return float64(units) * mp.PricePerUnit
}

// GetFinalPrice 获取应用倍率后的最终单价
func (mp *ModelPricing) GetFinalPrice() float64 {
	return mp.PricePerUnit * mp.Multiplier
}
