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
	ID                int64       `json:"id" db:"id"`
	ProviderID        int64       `json:"provider_id" db:"provider_id"`
	Name              string      `json:"name" db:"name"`
	Slug              string      `json:"slug" db:"slug"`
	DisplayName       *string     `json:"display_name,omitempty" db:"display_name"`
	Description       *string     `json:"description,omitempty" db:"description"`
	ModelType         ModelType   `json:"model_type" db:"model_type"`
	ContextLength     *int        `json:"context_length,omitempty" db:"context_length"`
	MaxTokens         *int        `json:"max_tokens,omitempty" db:"max_tokens"`
	SupportsStreaming bool        `json:"supports_streaming" db:"supports_streaming"`
	SupportsFunctions bool        `json:"supports_functions" db:"supports_functions"`
	Status            ModelStatus `json:"status" db:"status"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
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
	ID             int64       `json:"id" db:"id"`
	ModelID        int64       `json:"model_id" db:"model_id"`
	PricingType    PricingType `json:"pricing_type" db:"pricing_type"`
	PricePerUnit   float64     `json:"price_per_unit" db:"price_per_unit"`
	Multiplier     float64     `json:"multiplier" db:"multiplier"` // 价格倍率，默认1.5
	Unit           PricingUnit `json:"unit" db:"unit"`
	Currency       string      `json:"currency" db:"currency"`
	EffectiveFrom  time.Time   `json:"effective_from" db:"effective_from"`
	EffectiveUntil *time.Time  `json:"effective_until,omitempty" db:"effective_until"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
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
