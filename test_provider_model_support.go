package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/logger"
	"ai-api-gateway/internal/infrastructure/repositories"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("🧪 测试提供商模型支持多对多关系修复")
	fmt.Println("====================================================")

	// 创建测试数据库
	testDB, err := createTestDatabase()
	if err != nil {
		log.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	// 创建logger
	loggerInstance := logger.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})

	// 创建仓储工厂
	repoFactory := repositories.NewRepositoryFactory(testDB)

	// 创建服务工厂
	serviceFactory := services.NewServiceFactory(repoFactory, nil, loggerInstance)

	// 创建测试数据
	ctx := context.Background()
	if err := setupTestData(ctx, repoFactory); err != nil {
		log.Fatalf("Failed to setup test data: %v", err)
	}

	// 测试场景1：验证多个提供商支持同一个模型
	fmt.Println("\n📋 测试场景1：一个模型被多个提供商支持")
	if err := testMultipleProvidersForOneModel(ctx, repoFactory, serviceFactory, loggerInstance); err != nil {
		log.Fatalf("Test scenario 1 failed: %v", err)
	}

	// 测试场景2：验证RequestRouter的GetAvailableProviders方法
	fmt.Println("\n📋 测试场景2：RequestRouter获取可用提供商")
	if err := testRequestRouterGetAvailableProviders(ctx, repoFactory, loggerInstance); err != nil {
		log.Fatalf("Test scenario 2 failed: %v", err)
	}

	// 测试场景3：验证优先级排序
	fmt.Println("\n📋 测试场景3：验证提供商优先级排序")
	if err := testProviderPriorityOrdering(ctx, repoFactory); err != nil {
		log.Fatalf("Test scenario 3 failed: %v", err)
	}

	// 测试场景4：验证CRUD操作
	fmt.Println("\n📋 测试场景4：验证CRUD操作")
	if err := testProviderModelSupportCRUD(ctx, repoFactory); err != nil {
		log.Fatalf("Test scenario 4 failed: %v", err)
	}

	fmt.Println("\n✅ 所有测试通过！多对多关系修复成功！")
}

func createTestDatabase() (*sql.DB, error) {
	// 创建内存数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// 执行数据库迁移
	migrationSQL := `
	-- 提供商表
	CREATE TABLE providers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(100) NOT NULL UNIQUE,
		base_url VARCHAR(500) NOT NULL,
		api_key_encrypted TEXT,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		priority INTEGER NOT NULL DEFAULT 1,
		timeout_seconds INTEGER NOT NULL DEFAULT 30,
		retry_attempts INTEGER NOT NULL DEFAULT 3,
		health_check_url VARCHAR(500),
		health_check_interval INTEGER NOT NULL DEFAULT 300,
		last_health_check DATETIME,
		health_status VARCHAR(20) NOT NULL DEFAULT 'healthy',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- 模型表
	CREATE TABLE models (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		provider_id INTEGER NOT NULL,
		name VARCHAR(100) NOT NULL,
		slug VARCHAR(100) NOT NULL,
		display_name VARCHAR(200),
		description TEXT,
		model_type VARCHAR(50) NOT NULL,
		context_length INTEGER,
		max_tokens INTEGER,
		supports_streaming BOOLEAN NOT NULL DEFAULT false,
		supports_functions BOOLEAN NOT NULL DEFAULT false,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(provider_id, slug)
	);

	-- 提供商模型支持表
	CREATE TABLE provider_model_support (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		provider_id INTEGER NOT NULL,
		model_slug VARCHAR(100) NOT NULL,
		upstream_model_name VARCHAR(100),
		enabled BOOLEAN NOT NULL DEFAULT true,
		priority INTEGER NOT NULL DEFAULT 1,
		config TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(provider_id, model_slug)
	);

	-- 创建索引
	CREATE INDEX idx_provider_model_support_provider ON provider_model_support(provider_id);
	CREATE INDEX idx_provider_model_support_model ON provider_model_support(model_slug);
	CREATE INDEX idx_provider_model_support_enabled ON provider_model_support(enabled);
	CREATE INDEX idx_provider_model_support_priority ON provider_model_support(model_slug, priority);
	`

	if _, err := db.Exec(migrationSQL); err != nil {
		return nil, fmt.Errorf("failed to execute migration: %w", err)
	}

	return db, nil
}

func setupTestData(ctx context.Context, repoFactory *repositories.RepositoryFactory) error {
	providerRepo := repoFactory.ProviderRepository()
	supportRepo := repoFactory.ProviderModelSupportRepository()

	// 创建测试提供商
	providers := []*entities.Provider{
		{
			Name:         "OpenAI Official",
			Slug:         "openai-official",
			BaseURL:      "https://api.openai.com/v1",
			Status:       entities.ProviderStatusActive,
			Priority:     1,
			HealthStatus: entities.HealthStatusHealthy,
		},
		{
			Name:         "OpenAI Compatible Provider A",
			Slug:         "openai-compat-a",
			BaseURL:      "https://api.provider-a.com/v1",
			Status:       entities.ProviderStatusActive,
			Priority:     2,
			HealthStatus: entities.HealthStatusHealthy,
		},
		{
			Name:         "OpenAI Compatible Provider B",
			Slug:         "openai-compat-b",
			BaseURL:      "https://api.provider-b.com/v1",
			Status:       entities.ProviderStatusActive,
			Priority:     3,
			HealthStatus: entities.HealthStatusHealthy,
		},
	}

	// 创建提供商
	for _, provider := range providers {
		if err := providerRepo.Create(ctx, provider); err != nil {
			return fmt.Errorf("failed to create provider %s: %w", provider.Name, err)
		}
		fmt.Printf("✓ 创建提供商: %s (ID: %d)\n", provider.Name, provider.ID)
	}

	// 创建模型支持关系 - 多个提供商支持同一个模型
	modelSupports := []*entities.ProviderModelSupport{
		// GPT-4 被三个提供商支持，不同优先级
		{ProviderID: providers[0].ID, ModelSlug: "gpt-4", UpstreamModelName: stringPtr("gpt-4"), Enabled: true, Priority: 1},
		{ProviderID: providers[1].ID, ModelSlug: "gpt-4", UpstreamModelName: stringPtr("gpt-4"), Enabled: true, Priority: 2},
		{ProviderID: providers[2].ID, ModelSlug: "gpt-4", UpstreamModelName: stringPtr("gpt-4"), Enabled: true, Priority: 3},

		// GPT-3.5-turbo 被两个提供商支持
		{ProviderID: providers[0].ID, ModelSlug: "gpt-3.5-turbo", UpstreamModelName: stringPtr("gpt-3.5-turbo"), Enabled: true, Priority: 1},
		{ProviderID: providers[1].ID, ModelSlug: "gpt-3.5-turbo", UpstreamModelName: stringPtr("gpt-3.5-turbo"), Enabled: true, Priority: 2},

		// Claude-3 只被一个提供商支持（Provider B 有特殊的 Claude 支持）
		{ProviderID: providers[2].ID, ModelSlug: "claude-3-opus", UpstreamModelName: stringPtr("claude-3-opus-20240229"), Enabled: true, Priority: 1},
	}

	// 创建模型支持关系
	for _, support := range modelSupports {
		if err := supportRepo.Create(ctx, support); err != nil {
			return fmt.Errorf("failed to create model support: %w", err)
		}
		fmt.Printf("✓ 创建模型支持: Provider %d -> %s (优先级: %d)\n", support.ProviderID, support.ModelSlug, support.Priority)
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}

func testMultipleProvidersForOneModel(ctx context.Context, repoFactory *repositories.RepositoryFactory, serviceFactory *services.ServiceFactory, logger logger.Logger) error {
	supportRepo := repoFactory.ProviderModelSupportRepository()

	// 测试获取支持 gpt-4 的提供商
	supportInfos, err := supportRepo.GetSupportingProviders(ctx, "gpt-4")
	if err != nil {
		return fmt.Errorf("failed to get supporting providers: %w", err)
	}

	fmt.Printf("📊 模型 'gpt-4' 被 %d 个提供商支持:\n", len(supportInfos))

	if len(supportInfos) != 3 {
		return fmt.Errorf("expected 3 providers for gpt-4, got %d", len(supportInfos))
	}

	for i, info := range supportInfos {
		fmt.Printf("  %d. %s (优先级: %d, 上游模型: %s)\n",
			i+1, info.Provider.Name, info.Priority, info.UpstreamModelName)

		if !info.IsAvailable() {
			return fmt.Errorf("provider %s should be available", info.Provider.Name)
		}
	}

	// 验证优先级排序
	if supportInfos[0].Priority != 1 || supportInfos[1].Priority != 2 || supportInfos[2].Priority != 3 {
		return fmt.Errorf("providers are not sorted by priority correctly")
	}

	fmt.Println("✅ 多对多关系测试通过")
	return nil
}

func testRequestRouterGetAvailableProviders(ctx context.Context, repoFactory *repositories.RepositoryFactory, logger logger.Logger) error {
	// 创建一个简化的RequestRouter来测试GetAvailableProviders方法
	// 注意：这里我们只测试核心逻辑，不涉及完整的依赖
	supportRepo := repoFactory.ProviderModelSupportRepository()

	// 直接测试 GetSupportingProviders 方法（这是 RequestRouter 内部使用的）
	supportInfos, err := supportRepo.GetSupportingProviders(ctx, "gpt-4")
	if err != nil {
		return fmt.Errorf("failed to get supporting providers: %w", err)
	}

	fmt.Printf("📊 RequestRouter 查询结果: 模型 'gpt-4' 有 %d 个可用提供商\n", len(supportInfos))

	expectedProviders := []string{"OpenAI Official", "OpenAI Compatible Provider A", "OpenAI Compatible Provider B"}
	if len(supportInfos) != len(expectedProviders) {
		return fmt.Errorf("expected %d providers, got %d", len(expectedProviders), len(supportInfos))
	}

	for i, info := range supportInfos {
		fmt.Printf("  %d. %s (优先级: %d)\n", i+1, info.Provider.Name, info.Priority)
		if info.Provider.Name != expectedProviders[i] {
			return fmt.Errorf("expected provider %s at position %d, got %s", expectedProviders[i], i, info.Provider.Name)
		}
	}

	// 测试不存在的模型
	supportInfos, err = supportRepo.GetSupportingProviders(ctx, "non-existent-model")
	if err != nil {
		return fmt.Errorf("failed to query non-existent model: %w", err)
	}

	if len(supportInfos) != 0 {
		return fmt.Errorf("expected 0 providers for non-existent model, got %d", len(supportInfos))
	}

	fmt.Println("✅ RequestRouter 查询逻辑测试通过")
	return nil
}

func testProviderPriorityOrdering(ctx context.Context, repoFactory *repositories.RepositoryFactory) error {
	supportRepo := repoFactory.ProviderModelSupportRepository()

	// 测试 gpt-3.5-turbo 的优先级排序
	supportInfos, err := supportRepo.GetSupportingProviders(ctx, "gpt-3.5-turbo")
	if err != nil {
		return fmt.Errorf("failed to get supporting providers: %w", err)
	}

	fmt.Printf("📊 模型 'gpt-3.5-turbo' 优先级排序测试:\n")

	if len(supportInfos) != 2 {
		return fmt.Errorf("expected 2 providers for gpt-3.5-turbo, got %d", len(supportInfos))
	}

	// 验证按优先级排序
	for i, info := range supportInfos {
		fmt.Printf("  %d. %s (优先级: %d, 提供商优先级: %d)\n",
			i+1, info.Provider.Name, info.Priority, info.Provider.Priority)

		if i > 0 && info.Priority < supportInfos[i-1].Priority {
			return fmt.Errorf("providers are not sorted by priority correctly")
		}
	}

	// 测试只有一个提供商的模型
	supportInfos, err = supportRepo.GetSupportingProviders(ctx, "claude-3-opus")
	if err != nil {
		return fmt.Errorf("failed to get supporting providers for claude-3-opus: %w", err)
	}

	fmt.Printf("📊 模型 'claude-3-opus' 单提供商测试:\n")

	if len(supportInfos) != 1 {
		return fmt.Errorf("expected 1 provider for claude-3-opus, got %d", len(supportInfos))
	}

	info := supportInfos[0]
	fmt.Printf("  1. %s (上游模型: %s)\n", info.Provider.Name, info.UpstreamModelName)

	if info.UpstreamModelName != "claude-3-opus-20240229" {
		return fmt.Errorf("expected upstream model name 'claude-3-opus-20240229', got '%s'", info.UpstreamModelName)
	}

	fmt.Println("✅ 优先级排序测试通过")
	return nil
}

func testProviderModelSupportCRUD(ctx context.Context, repoFactory *repositories.RepositoryFactory) error {
	supportRepo := repoFactory.ProviderModelSupportRepository()
	providerRepo := repoFactory.ProviderRepository()

	fmt.Println("📊 测试提供商模型支持 CRUD 操作")

	// 创建一个新的提供商
	newProvider := &entities.Provider{
		Name:         "Test Provider",
		Slug:         "test-provider",
		BaseURL:      "https://api.test.com/v1",
		Status:       entities.ProviderStatusActive,
		Priority:     10,
		HealthStatus: entities.HealthStatusHealthy,
	}

	if err := providerRepo.Create(ctx, newProvider); err != nil {
		return fmt.Errorf("failed to create test provider: %w", err)
	}

	// 创建模型支持
	support := &entities.ProviderModelSupport{
		ProviderID:        newProvider.ID,
		ModelSlug:         "test-model",
		UpstreamModelName: stringPtr("test-model-v1"),
		Enabled:           true,
		Priority:          1,
	}

	if err := supportRepo.Create(ctx, support); err != nil {
		return fmt.Errorf("failed to create model support: %w", err)
	}

	fmt.Printf("✓ 创建模型支持: ID %d\n", support.ID)

	// 读取模型支持
	retrievedSupport, err := supportRepo.GetByID(ctx, support.ID)
	if err != nil {
		return fmt.Errorf("failed to get model support by ID: %w", err)
	}

	if retrievedSupport.ModelSlug != "test-model" {
		return fmt.Errorf("expected model slug 'test-model', got '%s'", retrievedSupport.ModelSlug)
	}

	fmt.Printf("✓ 读取模型支持: %s\n", retrievedSupport.ModelSlug)

	// 更新模型支持
	retrievedSupport.Priority = 5
	retrievedSupport.Enabled = false

	if err := supportRepo.Update(ctx, retrievedSupport); err != nil {
		return fmt.Errorf("failed to update model support: %w", err)
	}

	fmt.Printf("✓ 更新模型支持: 优先级 %d, 启用状态 %t\n", retrievedSupport.Priority, retrievedSupport.Enabled)

	// 验证更新
	updatedSupport, err := supportRepo.GetByID(ctx, support.ID)
	if err != nil {
		return fmt.Errorf("failed to get updated model support: %w", err)
	}

	if updatedSupport.Priority != 5 || updatedSupport.Enabled != false {
		return fmt.Errorf("model support was not updated correctly")
	}

	// 删除模型支持
	if err := supportRepo.Delete(ctx, support.ID); err != nil {
		return fmt.Errorf("failed to delete model support: %w", err)
	}

	fmt.Printf("✓ 删除模型支持: ID %d\n", support.ID)

	// 验证删除
	_, err = supportRepo.GetByID(ctx, support.ID)
	if err == nil {
		return fmt.Errorf("model support should have been deleted")
	}

	fmt.Println("✅ CRUD 操作测试通过")
	return nil
}
