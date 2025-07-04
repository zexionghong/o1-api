package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// modelPricingRepositoryImpl 模型定价仓储实现
type modelPricingRepositoryImpl struct {
	db *sql.DB
}

// NewModelPricingRepository 创建模型定价仓储实例
func NewModelPricingRepository(db *sql.DB) repositories.ModelPricingRepository {
	return &modelPricingRepositoryImpl{db: db}
}

// Create 创建模型定价
func (r *modelPricingRepositoryImpl) Create(ctx context.Context, pricing *entities.ModelPricing) error {
	query := `
		INSERT INTO model_pricing (model_id, pricing_type, price_per_unit, multiplier, unit, currency, effective_from, effective_until, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		pricing.ModelID,
		pricing.PricingType,
		pricing.PricePerUnit,
		pricing.Multiplier,
		pricing.Unit,
		pricing.Currency,
		pricing.EffectiveFrom,
		pricing.EffectiveUntil,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create model pricing: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	pricing.ID = id
	pricing.CreatedAt = time.Now()

	return nil
}

// GetByID 根据ID获取模型定价
func (r *modelPricingRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.ModelPricing, error) {
	query := `
		SELECT id, model_id, pricing_type, price_per_unit, multiplier, unit, currency, effective_from, effective_until, created_at
		FROM model_pricing WHERE id = ?
	`

	pricing := &entities.ModelPricing{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pricing.ID,
		&pricing.ModelID,
		&pricing.PricingType,
		&pricing.PricePerUnit,
		&pricing.Multiplier,
		&pricing.Unit,
		&pricing.Currency,
		&pricing.EffectiveFrom,
		&pricing.EffectiveUntil,
		&pricing.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrModelNotFound
		}
		return nil, fmt.Errorf("failed to get model pricing by id: %w", err)
	}

	return pricing, nil
}

// GetByModelID 根据模型ID获取定价列表
func (r *modelPricingRepositoryImpl) GetByModelID(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error) {
	query := `
		SELECT id, model_id, pricing_type, price_per_unit, multiplier, unit, currency, effective_from, effective_until, created_at
		FROM model_pricing
		WHERE model_id = ?
		ORDER BY effective_from DESC
	`

	rows, err := r.db.QueryContext(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model pricing by model id: %w", err)
	}
	defer rows.Close()

	var pricings []*entities.ModelPricing
	for rows.Next() {
		pricing := &entities.ModelPricing{}
		err := rows.Scan(
			&pricing.ID,
			&pricing.ModelID,
			&pricing.PricingType,
			&pricing.PricePerUnit,
			&pricing.Multiplier,
			&pricing.Unit,
			&pricing.Currency,
			&pricing.EffectiveFrom,
			&pricing.EffectiveUntil,
			&pricing.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model pricing: %w", err)
		}
		pricings = append(pricings, pricing)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate model pricing rows: %w", err)
	}

	return pricings, nil
}

// GetCurrentPricing 获取模型当前有效定价
func (r *modelPricingRepositoryImpl) GetCurrentPricing(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error) {
	now := time.Now()
	query := `
		SELECT id, model_id, pricing_type, price_per_unit, multiplier, unit, currency, effective_from, effective_until, created_at
		FROM model_pricing
		WHERE model_id = ?
		  AND effective_from <= ?
		  AND (effective_until IS NULL OR effective_until > ?)
		ORDER BY pricing_type, effective_from DESC
	`

	rows, err := r.db.QueryContext(ctx, query, modelID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get current model pricing: %w", err)
	}
	defer rows.Close()

	var pricings []*entities.ModelPricing
	for rows.Next() {
		pricing := &entities.ModelPricing{}
		err := rows.Scan(
			&pricing.ID,
			&pricing.ModelID,
			&pricing.PricingType,
			&pricing.PricePerUnit,
			&pricing.Multiplier,
			&pricing.Unit,
			&pricing.Currency,
			&pricing.EffectiveFrom,
			&pricing.EffectiveUntil,
			&pricing.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan current model pricing: %w", err)
		}
		pricings = append(pricings, pricing)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate current model pricing rows: %w", err)
	}

	return pricings, nil
}

// Update 更新模型定价
func (r *modelPricingRepositoryImpl) Update(ctx context.Context, pricing *entities.ModelPricing) error {
	query := `
		UPDATE model_pricing
		SET model_id = ?, pricing_type = ?, price_per_unit = ?, multiplier = ?, unit = ?, currency = ?,
		    effective_from = ?, effective_until = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		pricing.ModelID,
		pricing.PricingType,
		pricing.PricePerUnit,
		pricing.Multiplier,
		pricing.Unit,
		pricing.Currency,
		pricing.EffectiveFrom,
		pricing.EffectiveUntil,
		pricing.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update model pricing: %w", err)
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

// Delete 删除模型定价
func (r *modelPricingRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM model_pricing WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete model pricing: %w", err)
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

// List 获取模型定价列表
func (r *modelPricingRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.ModelPricing, error) {
	query := `
		SELECT id, model_id, pricing_type, price_per_unit, unit, currency, effective_from, effective_until, created_at
		FROM model_pricing 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list model pricing: %w", err)
	}
	defer rows.Close()

	var pricings []*entities.ModelPricing
	for rows.Next() {
		pricing := &entities.ModelPricing{}
		err := rows.Scan(
			&pricing.ID,
			&pricing.ModelID,
			&pricing.PricingType,
			&pricing.PricePerUnit,
			&pricing.Unit,
			&pricing.Currency,
			&pricing.EffectiveFrom,
			&pricing.EffectiveUntil,
			&pricing.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model pricing: %w", err)
		}
		pricings = append(pricings, pricing)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate model pricing rows: %w", err)
	}

	return pricings, nil
}

// Count 获取模型定价总数
func (r *modelPricingRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM model_pricing`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count model pricing: %w", err)
	}

	return count, nil
}

// GetPricingByType 根据定价类型获取定价
func (r *modelPricingRepositoryImpl) GetPricingByType(ctx context.Context, modelID int64, pricingType entities.PricingType) (*entities.ModelPricing, error) {
	now := time.Now()
	query := `
		SELECT id, model_id, pricing_type, price_per_unit, multiplier, unit, currency, effective_from, effective_until, created_at
		FROM model_pricing
		WHERE model_id = ?
		  AND pricing_type = ?
		  AND effective_from <= ?
		  AND (effective_until IS NULL OR effective_until > ?)
		ORDER BY effective_from DESC
		LIMIT 1
	`

	pricing := &entities.ModelPricing{}
	err := r.db.QueryRowContext(ctx, query, modelID, pricingType, now, now).Scan(
		&pricing.ID,
		&pricing.ModelID,
		&pricing.PricingType,
		&pricing.PricePerUnit,
		&pricing.Multiplier,
		&pricing.Unit,
		&pricing.Currency,
		&pricing.EffectiveFrom,
		&pricing.EffectiveUntil,
		&pricing.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrModelNotFound
		}
		return nil, fmt.Errorf("failed to get model pricing by type: %w", err)
	}

	return pricing, nil
}
