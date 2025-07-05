package repositories

import (
	"context"
	"database/sql"

	"github.com/go-redis/redis/v8"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	redisInfra "ai-api-gateway/internal/infrastructure/redis"
)

// CachedModelRepositoryImpl 带缓存的模型仓储实现
type CachedModelRepositoryImpl struct {
	db                  *sql.DB
	cache               *redisInfra.CacheService
	invalidationService *redisInfra.CacheInvalidationService
	baseRepo            repositories.ModelRepository
}

// NewCachedModelRepository 创建带缓存的模型仓储
func NewCachedModelRepository(
	db *sql.DB,
	cache *redisInfra.CacheService,
	invalidationService *redisInfra.CacheInvalidationService,
) repositories.ModelRepository {
	baseRepo := NewModelRepository(db)
	return &CachedModelRepositoryImpl{
		db:                  db,
		cache:               cache,
		invalidationService: invalidationService,
		baseRepo:            baseRepo,
	}
}

// GetByID 根据ID获取模型（带缓存）
func (r *CachedModelRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.Model, error) {
	// 尝试从缓存获取
	if r.cache != nil && r.cache.IsEnabled() {
		if model, err := r.cache.GetModel(ctx, id); err == nil {
			return model, nil
		} else if err != redis.Nil {
			// 记录缓存错误但继续查询数据库
			// 这里可以添加日志记录
		}
	}

	// 从数据库查询
	model, err := r.baseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if r.cache != nil && r.cache.IsEnabled() {
		if err := r.cache.SetModel(ctx, model); err != nil {
			// 记录缓存错误但不影响主流程
		}
	}

	return model, nil
}

// GetActiveModels 获取活跃模型列表（带缓存）
func (r *CachedModelRepositoryImpl) GetActiveModels(ctx context.Context) ([]*entities.Model, error) {
	// 尝试从缓存获取
	if r.cache != nil && r.cache.IsEnabled() {
		if models, err := r.cache.GetActiveModels(ctx); err == nil {
			return models, nil
		} else if err != redis.Nil {
			// 记录缓存错误但继续查询数据库
		}
	}

	// 从数据库查询
	models, err := r.baseRepo.GetActiveModels(ctx)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if r.cache != nil && r.cache.IsEnabled() {
		if err := r.cache.SetActiveModels(ctx, models); err != nil {
			// 记录缓存错误但不影响主流程
		}
	}

	return models, nil
}

// GetModelsByType 根据类型获取模型列表（带缓存）
func (r *CachedModelRepositoryImpl) GetModelsByType(ctx context.Context, modelType entities.ModelType) ([]*entities.Model, error) {
	// 尝试从缓存获取
	if r.cache != nil && r.cache.IsEnabled() {
		if models, err := r.cache.GetModelsByType(ctx, string(modelType)); err == nil {
			return models, nil
		} else if err != redis.Nil {
			// 记录缓存错误但继续查询数据库
		}
	}

	// 从数据库查询
	models, err := r.baseRepo.GetModelsByType(ctx, modelType)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if r.cache != nil && r.cache.IsEnabled() {
		if err := r.cache.SetModelsByType(ctx, string(modelType), models); err != nil {
			// 记录缓存错误但不影响主流程
		}
	}

	return models, nil
}

// GetAvailableModels 获取可用模型列表（带缓存）
func (r *CachedModelRepositoryImpl) GetAvailableModels(ctx context.Context) ([]*entities.Model, error) {
	// 尝试从缓存获取
	if r.cache != nil && r.cache.IsEnabled() {
		if models, err := r.cache.GetAvailableModels(ctx); err == nil {
			return models, nil
		} else if err != redis.Nil {
			// 记录缓存错误但继续查询数据库
		}
	}

	// 从数据库查询
	models, err := r.baseRepo.GetAvailableModels(ctx)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if r.cache != nil && r.cache.IsEnabled() {
		if err := r.cache.SetAvailableModels(ctx, models); err != nil {
			// 记录缓存错误但不影响主流程
		}
	}

	return models, nil
}

// Create 创建模型（带缓存失效）
func (r *CachedModelRepositoryImpl) Create(ctx context.Context, model *entities.Model) error {
	// 执行数据库操作
	err := r.baseRepo.Create(ctx, model)
	if err != nil {
		return err
	}

	// 失效相关缓存
	if r.invalidationService != nil {
		if err := r.invalidationService.InvalidateModelCache(ctx, model.ID, string(model.ModelType)); err != nil {
			// 记录缓存失效错误但不影响主流程
		}
	}

	return nil
}

// Update 更新模型（带缓存失效）
func (r *CachedModelRepositoryImpl) Update(ctx context.Context, model *entities.Model) error {
	// 执行数据库操作
	err := r.baseRepo.Update(ctx, model)
	if err != nil {
		return err
	}

	// 失效相关缓存
	if r.invalidationService != nil {
		if err := r.invalidationService.InvalidateModelCache(ctx, model.ID, string(model.ModelType)); err != nil {
			// 记录缓存失效错误但不影响主流程
		}
	}

	return nil
}

// Delete 删除模型（带缓存失效）
func (r *CachedModelRepositoryImpl) Delete(ctx context.Context, id int64) error {
	// 先获取模型信息以便失效缓存
	model, err := r.GetByID(ctx, id)
	if err != nil && err != entities.ErrModelNotFound {
		return err
	}

	// 执行数据库操作
	err = r.baseRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 失效相关缓存
	if r.invalidationService != nil && model != nil {
		if err := r.invalidationService.InvalidateModelCache(ctx, id, string(model.ModelType)); err != nil {
			// 记录缓存失效错误但不影响主流程
		}
	}

	return nil
}

// 以下方法直接委托给基础仓储，因为它们不需要特殊的缓存处理

func (r *CachedModelRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entities.Model, error) {
	return r.baseRepo.GetBySlug(ctx, slug)
}

func (r *CachedModelRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.Model, error) {
	return r.baseRepo.List(ctx, offset, limit)
}

func (r *CachedModelRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.baseRepo.Count(ctx)
}
