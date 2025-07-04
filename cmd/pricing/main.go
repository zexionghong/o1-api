package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	repoImpl "ai-api-gateway/internal/infrastructure/repositories"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath      = flag.String("db", "./data/gateway.db", "Database file path")
		action      = flag.String("action", "list", "Action: list, add, update, delete")
		modelID     = flag.Int64("model", 0, "Model ID")
		pricingType = flag.String("type", "", "Pricing type: input, output, request")
		price       = flag.Float64("price", 0, "Price per unit")
		unit        = flag.String("unit", "token", "Unit: token, request, character")
		currency    = flag.String("currency", "USD", "Currency")
		effective   = flag.String("effective", "", "Effective date (YYYY-MM-DD)")
		until       = flag.String("until", "", "Effective until date (YYYY-MM-DD)")
	)
	flag.Parse()

	// 打开数据库连接
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 创建仓储
	repoFactory := repoImpl.NewRepositoryFactory(db)
	modelPricingRepo := repoFactory.ModelPricingRepository()
	modelRepo := repoFactory.ModelRepository()

	ctx := context.Background()

	switch *action {
	case "list":
		if err := listPricing(ctx, modelPricingRepo, modelRepo, *modelID); err != nil {
			log.Fatalf("Failed to list pricing: %v", err)
		}
	case "add":
		if err := addPricing(ctx, modelPricingRepo, *modelID, *pricingType, *price, *unit, *currency, *effective, *until); err != nil {
			log.Fatalf("Failed to add pricing: %v", err)
		}
	case "update":
		if err := updatePricing(ctx, modelPricingRepo, *modelID, *pricingType, *price, *unit, *currency, *effective, *until); err != nil {
			log.Fatalf("Failed to update pricing: %v", err)
		}
	case "delete":
		if err := deletePricing(ctx, modelPricingRepo, *modelID, *pricingType); err != nil {
			log.Fatalf("Failed to delete pricing: %v", err)
		}
	case "models":
		if err := listModels(ctx, modelRepo); err != nil {
			log.Fatalf("Failed to list models: %v", err)
		}
	default:
		fmt.Printf("Invalid action: %s\n", *action)
		fmt.Println("Available actions: list, add, update, delete, models")
		os.Exit(1)
	}
}

func listPricing(ctx context.Context, repo repositories.ModelPricingRepository, modelRepo repositories.ModelRepository, modelID int64) error {
	fmt.Println("=== Model Pricing ===")

	var pricings []*entities.ModelPricing
	var err error

	if modelID > 0 {
		pricings, err = repo.GetByModelID(ctx, modelID)
		if err != nil {
			return err
		}
		fmt.Printf("Pricing for Model ID: %d\n", modelID)
	} else {
		pricings, err = repo.List(ctx, 0, 100)
		if err != nil {
			return err
		}
		fmt.Println("All Model Pricing (first 100)")
	}

	if len(pricings) == 0 {
		fmt.Println("No pricing data found")
		return nil
	}

	// 获取模型信息用于显示
	modelMap := make(map[int64]*entities.Model)
	for _, pricing := range pricings {
		if _, exists := modelMap[pricing.ModelID]; !exists {
			model, err := modelRepo.GetByID(ctx, pricing.ModelID)
			if err == nil {
				modelMap[pricing.ModelID] = model
			}
		}
	}

	fmt.Printf("%-5s %-20s %-10s %-15s %-10s %-8s %-12s %-12s\n",
		"ID", "Model", "Type", "Price/Unit", "Unit", "Currency", "From", "Until")
	fmt.Println(strings.Repeat("-", 100))

	for _, pricing := range pricings {
		modelName := "Unknown"
		if model, exists := modelMap[pricing.ModelID]; exists {
			modelName = model.Name
		}

		until := "Forever"
		if pricing.EffectiveUntil != nil {
			until = pricing.EffectiveUntil.Format("2006-01-02")
		}

		fmt.Printf("%-5d %-20s %-10s %-15.8f %-10s %-8s %-12s %-12s\n",
			pricing.ID,
			modelName,
			pricing.PricingType,
			pricing.PricePerUnit,
			pricing.Unit,
			pricing.Currency,
			pricing.EffectiveFrom.Format("2006-01-02"),
			until,
		)
	}

	return nil
}

func addPricing(ctx context.Context, repo repositories.ModelPricingRepository, modelID int64, pricingType string, price float64, unit, currency, effective, until string) error {
	if modelID == 0 {
		return fmt.Errorf("model ID is required")
	}
	if pricingType == "" {
		return fmt.Errorf("pricing type is required")
	}
	if price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	// 解析日期
	effectiveFrom := time.Now()
	if effective != "" {
		var err error
		effectiveFrom, err = time.Parse("2006-01-02", effective)
		if err != nil {
			return fmt.Errorf("invalid effective date format: %v", err)
		}
	}

	var effectiveUntil *time.Time
	if until != "" {
		untilDate, err := time.Parse("2006-01-02", until)
		if err != nil {
			return fmt.Errorf("invalid until date format: %v", err)
		}
		effectiveUntil = &untilDate
	}

	pricing := &entities.ModelPricing{
		ModelID:        modelID,
		PricingType:    entities.PricingType(pricingType),
		PricePerUnit:   price,
		Unit:           entities.PricingUnit(unit),
		Currency:       currency,
		EffectiveFrom:  effectiveFrom,
		EffectiveUntil: effectiveUntil,
	}

	if err := repo.Create(ctx, pricing); err != nil {
		return err
	}

	fmt.Printf("Successfully added pricing: ID=%d, Model=%d, Type=%s, Price=%.8f\n",
		pricing.ID, modelID, pricingType, price)
	return nil
}

func updatePricing(ctx context.Context, repo repositories.ModelPricingRepository, modelID int64, pricingType string, price float64, unit, currency, effective, until string) error {
	if modelID == 0 {
		return fmt.Errorf("model ID is required")
	}
	if pricingType == "" {
		return fmt.Errorf("pricing type is required")
	}

	// 查找现有定价
	pricing, err := repo.GetPricingByType(ctx, modelID, entities.PricingType(pricingType))
	if err != nil {
		return fmt.Errorf("pricing not found: %v", err)
	}

	// 更新字段
	if price > 0 {
		pricing.PricePerUnit = price
	}
	if unit != "" {
		pricing.Unit = entities.PricingUnit(unit)
	}
	if currency != "" {
		pricing.Currency = currency
	}
	if effective != "" {
		effectiveFrom, err := time.Parse("2006-01-02", effective)
		if err != nil {
			return fmt.Errorf("invalid effective date format: %v", err)
		}
		pricing.EffectiveFrom = effectiveFrom
	}
	if until != "" {
		untilDate, err := time.Parse("2006-01-02", until)
		if err != nil {
			return fmt.Errorf("invalid until date format: %v", err)
		}
		pricing.EffectiveUntil = &untilDate
	}

	if err := repo.Update(ctx, pricing); err != nil {
		return err
	}

	fmt.Printf("Successfully updated pricing: ID=%d, Model=%d, Type=%s, Price=%.8f\n",
		pricing.ID, modelID, pricingType, pricing.PricePerUnit)
	return nil
}

func deletePricing(ctx context.Context, repo repositories.ModelPricingRepository, modelID int64, pricingType string) error {
	if modelID == 0 {
		return fmt.Errorf("model ID is required")
	}
	if pricingType == "" {
		return fmt.Errorf("pricing type is required")
	}

	// 查找现有定价
	pricing, err := repo.GetPricingByType(ctx, modelID, entities.PricingType(pricingType))
	if err != nil {
		return fmt.Errorf("pricing not found: %v", err)
	}

	if err := repo.Delete(ctx, pricing.ID); err != nil {
		return err
	}

	fmt.Printf("Successfully deleted pricing: ID=%d, Model=%d, Type=%s\n",
		pricing.ID, modelID, pricingType)
	return nil
}

func listModels(ctx context.Context, repo repositories.ModelRepository) error {
	fmt.Println("=== Available Models ===")

	models, err := repo.List(ctx, 0, 100)
	if err != nil {
		return err
	}

	if len(models) == 0 {
		fmt.Println("No models found")
		return nil
	}

	fmt.Printf("%-5s %-30s %-20s %-15s %-10s\n", "ID", "Name", "Slug", "Type", "Status")
	fmt.Println(strings.Repeat("-", 85))

	for _, model := range models {
		fmt.Printf("%-5d %-30s %-20s %-15s %-10s\n",
			model.ID,
			model.Name,
			model.Slug,
			model.ModelType,
			model.Status,
		)
	}

	return nil
}
