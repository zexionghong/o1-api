package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/repositories"

	_ "modernc.org/sqlite"
)

func main() {
	// 打开数据库连接
	db, err := sql.Open("sqlite", "./data/gateway.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 创建仓储
	repoFactory := repositories.NewRepositoryFactory(db)
	quotaRepo := repoFactory.QuotaRepository()
	userRepo := repoFactory.UserRepository()

	// 查找测试用户
	user, err := userRepo.GetByUsername(ctx, "e2e-test-user")
	if err != nil {
		log.Fatalf("Failed to find test user: %v", err)
	}

	fmt.Printf("🔍 Found test user: %s (ID: %d)\n", user.Username, user.ID)

	// 定义配额设置
	quotaConfigs := []struct {
		quotaType   entities.QuotaType
		period      entities.QuotaPeriod
		limitValue  float64
		description string
	}{
		{entities.QuotaTypeRequests, entities.QuotaPeriodMinute, 10, "每分钟最多10次请求"},
		{entities.QuotaTypeRequests, entities.QuotaPeriodHour, 100, "每小时最多100次请求"},
		{entities.QuotaTypeRequests, entities.QuotaPeriodDay, 1000, "每天最多1000次请求"},
		{entities.QuotaTypeTokens, entities.QuotaPeriodMinute, 1000, "每分钟最多1000个token"},
		{entities.QuotaTypeTokens, entities.QuotaPeriodHour, 10000, "每小时最多10000个token"},
		{entities.QuotaTypeTokens, entities.QuotaPeriodDay, 100000, "每天最多100000个token"},
		{entities.QuotaTypeCost, entities.QuotaPeriodMinute, 0.1, "每分钟最多花费0.1美元"},
		{entities.QuotaTypeCost, entities.QuotaPeriodHour, 1.0, "每小时最多花费1美元"},
		{entities.QuotaTypeCost, entities.QuotaPeriodDay, 10.0, "每天最多花费10美元"},
	}

	fmt.Println("\n📋 Creating quota settings...")

	// 创建配额设置
	for _, config := range quotaConfigs {
		quota := &entities.Quota{
			UserID:     user.ID,
			QuotaType:  config.quotaType,
			Period:     config.period,
			LimitValue: config.limitValue,
			Status:     entities.QuotaStatusActive,
		}

		err := quotaRepo.Create(ctx, quota)
		if err != nil {
			log.Printf("❌ Failed to create quota %s/%s: %v", config.quotaType, config.period, err)
			continue
		}

		fmt.Printf("✅ Created quota: %s - %s (Limit: %.6f)\n",
			config.description, config.period, config.limitValue)
	}

	fmt.Println("\n🎉 Quota setup completed!")

	// 显示当前配额状态
	fmt.Println("\n📊 Current quota settings:")
	quotas, err := quotaRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to get quotas: %v", err)
		return
	}

	fmt.Printf("%-15s %-10s %-15s %-10s\n", "Type", "Period", "Limit", "Status")
	fmt.Println("─────────────────────────────────────────────────────────")

	for _, quota := range quotas {
		fmt.Printf("%-15s %-10s %-15.6f %-10s\n",
			quota.QuotaType, quota.Period, quota.LimitValue, quota.Status)
	}

	fmt.Println("\n💡 Tips:")
	fmt.Println("   • 配额检查将在API请求时自动执行")
	fmt.Println("   • 超出配额的请求将被拒绝")
	fmt.Println("   • 配额使用情况会自动重置")
	fmt.Println("   • 可以通过API查看配额状态")
}
