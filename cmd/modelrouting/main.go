package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/repositories"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath = flag.String("db", "./data/gateway.db", "Database file path")
		action = flag.String("action", "mapping", "Action: mapping, test, available")
		model  = flag.String("model", "", "Model slug to test routing")
	)
	flag.Parse()

	// 打开数据库连接
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	switch *action {
	case "mapping":
		if err := showModelProviderMapping(ctx, db); err != nil {
			log.Fatalf("Failed to show mapping: %v", err)
		}
	case "test":
		if *model == "" {
			log.Fatalf("Model parameter is required for test action")
		}
		if err := testModelRouting(ctx, db, *model); err != nil {
			log.Fatalf("Failed to test routing: %v", err)
		}
	case "available":
		if err := showAvailableModels(ctx, db); err != nil {
			log.Fatalf("Failed to show available models: %v", err)
		}
	default:
		fmt.Printf("Invalid action: %s\n", *action)
		fmt.Println("Available actions: mapping, test, available")
	}
}

func showModelProviderMapping(ctx context.Context, db *sql.DB) error {
	fmt.Println("=== Model to Provider Mapping ===")

	// 创建仓储
	repoFactory := repositories.NewRepositoryFactory(db)
	providerRepo := repoFactory.ProviderRepository()
	modelRepo := repoFactory.ModelRepository()

	// 获取所有提供商
	providers, err := providerRepo.List(ctx, 0, 100)
	if err != nil {
		return fmt.Errorf("failed to get providers: %w", err)
	}

	for _, provider := range providers {
		fmt.Printf("\n🏢 Provider: %s (%s)\n", provider.Name, provider.Slug)
		fmt.Printf("   Status: %s, Health: %s, Priority: %d\n", 
			provider.Status, provider.HealthStatus, provider.Priority)

		// 获取该提供商的模型
		models, err := modelRepo.GetByProviderID(ctx, provider.ID)
		if err != nil {
			fmt.Printf("   ❌ Failed to get models: %v\n", err)
			continue
		}

		if len(models) == 0 {
			fmt.Printf("   📭 No models configured\n")
			continue
		}

		fmt.Printf("   📋 Models (%d):\n", len(models))
		for _, model := range models {
			status := "✅"
			if model.Status != entities.ModelStatusActive {
				status = "❌"
			}
			fmt.Printf("      %s %s (%s) - %s\n", 
				status, model.Name, model.Slug, model.ModelType)
		}
	}

	return nil
}

func testModelRouting(ctx context.Context, db *sql.DB) error {
	fmt.Printf("=== Testing Model Routing for '%s' ===\n", *model)

	// 创建服务
	repoFactory := repositories.NewRepositoryFactory(db)
	serviceFactory := services.NewServiceFactory(repoFactory)
	providerService := serviceFactory.ProviderService()
	modelService := serviceFactory.ModelService()

	// 模拟请求路由器的逻辑
	fmt.Println("\n1️⃣ Getting all available providers...")
	allProviders, err := providerService.GetAvailableProviders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get providers: %w", err)
	}
	fmt.Printf("   Found %d available providers\n", len(allProviders))

	// 查找支持指定模型的提供商
	fmt.Printf("\n2️⃣ Finding providers that support model '%s'...\n", *model)
	var supportingProviders []*entities.Provider
	
	for _, provider := range allProviders {
		fmt.Printf("   🔍 Checking provider: %s\n", provider.Name)
		
		// 获取该提供商的可用模型
		models, err := modelService.GetAvailableModels(ctx, provider.ID)
		if err != nil {
			fmt.Printf("      ❌ Failed to get models: %v\n", err)
			continue
		}

		// 检查是否有匹配的模型
		found := false
		for _, model := range models {
			if model.Slug == *model {
				supportingProviders = append(supportingProviders, provider)
				fmt.Printf("      ✅ Supports model '%s' (ID: %d)\n", model.Slug, model.ID)
				found = true
				break
			}
		}
		
		if !found {
			fmt.Printf("      ❌ Does not support model '%s'\n", *model)
		}
	}

	if len(supportingProviders) == 0 {
		fmt.Printf("\n❌ No providers support model '%s'\n", *model)
		fmt.Println("\n💡 Available models:")
		return showAvailableModels(ctx, db)
	}

	fmt.Printf("\n3️⃣ Found %d provider(s) supporting model '%s':\n", len(supportingProviders), *model)
	for i, provider := range supportingProviders {
		fmt.Printf("   %d. %s (Priority: %d, Status: %s)\n", 
			i+1, provider.Name, provider.Priority, provider.Status)
	}

	// 模拟负载均衡选择
	fmt.Printf("\n4️⃣ Load balancer would select: %s\n", supportingProviders[0].Name)
	
	// 获取具体的模型信息
	selectedProvider := supportingProviders[0]
	selectedModel, err := modelService.GetModelBySlug(ctx, selectedProvider.ID, *model)
	if err != nil {
		return fmt.Errorf("failed to get model details: %w", err)
	}

	fmt.Printf("\n5️⃣ Model details:\n")
	fmt.Printf("   ID: %d\n", selectedModel.ID)
	fmt.Printf("   Name: %s\n", selectedModel.Name)
	fmt.Printf("   Type: %s\n", selectedModel.ModelType)
	fmt.Printf("   Context Length: %d\n", selectedModel.GetContextLength())
	fmt.Printf("   Max Tokens: %d\n", selectedModel.GetMaxTokens())
	fmt.Printf("   Supports Streaming: %t\n", selectedModel.SupportsStreaming)
	fmt.Printf("   Supports Functions: %t\n", selectedModel.SupportsFunctions)

	fmt.Printf("\n✅ Request would be routed to: %s -> %s\n", 
		selectedProvider.Name, selectedProvider.BaseURL)

	return nil
}

func showAvailableModels(ctx context.Context, db *sql.DB) error {
	fmt.Println("=== Available Models ===")

	// 创建仓储
	repoFactory := repositories.NewRepositoryFactory(db)
	modelRepo := repoFactory.ModelRepository()
	providerRepo := repoFactory.ProviderRepository()

	// 获取所有活跃模型
	models, err := modelRepo.GetActiveModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active models: %w", err)
	}

	// 按提供商分组
	providerModels := make(map[int64][]*entities.Model)
	for _, model := range models {
		providerModels[model.ProviderID] = append(providerModels[model.ProviderID], model)
	}

	// 获取提供商信息
	providers, err := providerRepo.List(ctx, 0, 100)
	if err != nil {
		return fmt.Errorf("failed to get providers: %w", err)
	}

	providerMap := make(map[int64]*entities.Provider)
	for _, provider := range providers {
		providerMap[provider.ID] = provider
	}

	fmt.Printf("\n📋 Available models for API requests:\n")
	fmt.Printf("%-30s %-20s %-15s %-10s\n", "Model Slug", "Provider", "Type", "Streaming")
	fmt.Println(strings.Repeat("-", 80))

	for providerID, models := range providerModels {
		provider := providerMap[providerID]
		if provider == nil || provider.Status != entities.ProviderStatusActive {
			continue
		}

		for _, model := range models {
			streaming := "No"
			if model.SupportsStreaming {
				streaming = "Yes"
			}
			
			fmt.Printf("%-30s %-20s %-15s %-10s\n",
				model.Slug,
				provider.Name,
				model.ModelType,
				streaming,
			)
		}
	}

	fmt.Printf("\n💡 Usage example:\n")
	fmt.Printf("   curl -X POST http://localhost:8080/v1/chat/completions \\\n")
	fmt.Printf("     -H \"Authorization: Bearer YOUR_API_KEY\" \\\n")
	fmt.Printf("     -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("     -d '{\n")
	fmt.Printf("       \"model\": \"gpt-4\",\n")
	fmt.Printf("       \"messages\": [{\"role\": \"user\", \"content\": \"Hello!\"}]\n")
	fmt.Printf("     }'\n")

	return nil
}

func testModelRouting(ctx context.Context, db *sql.DB, modelSlug string) error {
	fmt.Printf("=== Testing Model Routing for '%s' ===\n", modelSlug)

	// 创建服务
	repoFactory := repositories.NewRepositoryFactory(db)
	serviceFactory := services.NewServiceFactory(repoFactory)
	providerService := serviceFactory.ProviderService()
	modelService := serviceFactory.ModelService()

	// 模拟请求路由器的逻辑
	fmt.Println("\n1️⃣ Getting all available providers...")
	allProviders, err := providerService.GetAvailableProviders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get providers: %w", err)
	}
	fmt.Printf("   Found %d available providers\n", len(allProviders))

	// 查找支持指定模型的提供商
	fmt.Printf("\n2️⃣ Finding providers that support model '%s'...\n", modelSlug)
	var supportingProviders []*entities.Provider
	
	for _, provider := range allProviders {
		fmt.Printf("   🔍 Checking provider: %s\n", provider.Name)
		
		// 获取该提供商的可用模型
		models, err := modelService.GetAvailableModels(ctx, provider.ID)
		if err != nil {
			fmt.Printf("      ❌ Failed to get models: %v\n", err)
			continue
		}

		// 检查是否有匹配的模型
		found := false
		for _, model := range models {
			if model.Slug == modelSlug {
				supportingProviders = append(supportingProviders, provider)
				fmt.Printf("      ✅ Supports model '%s' (ID: %d)\n", model.Slug, model.ID)
				found = true
				break
			}
		}
		
		if !found {
			fmt.Printf("      ❌ Does not support model '%s'\n", modelSlug)
		}
	}

	if len(supportingProviders) == 0 {
		fmt.Printf("\n❌ No providers support model '%s'\n", modelSlug)
		fmt.Println("\n💡 Available models:")
		return showAvailableModels(ctx, db)
	}

	fmt.Printf("\n3️⃣ Found %d provider(s) supporting model '%s':\n", len(supportingProviders), modelSlug)
	for i, provider := range supportingProviders {
		fmt.Printf("   %d. %s (Priority: %d, Status: %s)\n", 
			i+1, provider.Name, provider.Priority, provider.Status)
	}

	// 模拟负载均衡选择
	fmt.Printf("\n4️⃣ Load balancer would select: %s\n", supportingProviders[0].Name)
	
	// 获取具体的模型信息
	selectedProvider := supportingProviders[0]
	selectedModel, err := modelService.GetModelBySlug(ctx, selectedProvider.ID, modelSlug)
	if err != nil {
		return fmt.Errorf("failed to get model details: %w", err)
	}

	fmt.Printf("\n5️⃣ Model details:\n")
	fmt.Printf("   ID: %d\n", selectedModel.ID)
	fmt.Printf("   Name: %s\n", selectedModel.Name)
	fmt.Printf("   Type: %s\n", selectedModel.ModelType)
	fmt.Printf("   Context Length: %d\n", selectedModel.GetContextLength())
	fmt.Printf("   Max Tokens: %d\n", selectedModel.GetMaxTokens())
	fmt.Printf("   Supports Streaming: %t\n", selectedModel.SupportsStreaming)
	fmt.Printf("   Supports Functions: %t\n", selectedModel.SupportsFunctions)

	fmt.Printf("\n✅ Request would be routed to: %s -> %s\n", 
		selectedProvider.Name, selectedProvider.BaseURL)

	return nil
}
