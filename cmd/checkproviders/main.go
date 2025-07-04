package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath = flag.String("db", "./data/gateway.db", "Database file path")
	)
	flag.Parse()

	// 打开数据库连接
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	fmt.Println("=== 当前提供商配置检查 ===")

	if err := checkProviders(ctx, db); err != nil {
		log.Fatalf("Failed to check providers: %v", err)
	}

	fmt.Println("\n=== 模型路由配置检查 ===")

	if err := checkModelRouting(ctx, db); err != nil {
		log.Fatalf("Failed to check model routing: %v", err)
	}

	fmt.Println("\n=== 系统就绪状态 ===")

	if err := checkSystemReadiness(ctx, db); err != nil {
		log.Fatalf("System readiness check failed: %v", err)
	}
}

func checkProviders(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT id, name, slug, base_url, status, health_status, priority
		FROM providers 
		ORDER BY priority ASC, name ASC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query providers: %w", err)
	}
	defer rows.Close()

	fmt.Printf("%-5s %-20s %-15s %-40s %-10s %-10s %-8s\n", 
		"ID", "Name", "Slug", "Base URL", "Status", "Health", "Priority")
	fmt.Println(strings.Repeat("-", 110))

	providerCount := 0
	activeCount := 0

	for rows.Next() {
		var id int64
		var name, slug, baseURL, status, healthStatus string
		var priority int

		err := rows.Scan(&id, &name, &slug, &baseURL, &status, &healthStatus, &priority)
		if err != nil {
			return fmt.Errorf("failed to scan provider: %w", err)
		}

		statusIcon := "❌"
		if status == "active" {
			statusIcon = "✅"
			activeCount++
		}

		healthIcon := "❓"
		switch healthStatus {
		case "healthy":
			healthIcon = "💚"
		case "unhealthy":
			healthIcon = "❤️"
		}

		fmt.Printf("%-5d %-20s %-15s %-40s %s%-8s %s%-8s %-8d\n", 
			id, name, slug, baseURL, statusIcon, status, healthIcon, healthStatus, priority)

		providerCount++
	}

	fmt.Printf("\n📊 统计: 总计 %d 个提供商, %d 个活跃\n", providerCount, activeCount)

	if activeCount == 0 {
		fmt.Println("⚠️  警告: 没有活跃的提供商!")
	}

	return rows.Err()
}

func checkModelRouting(ctx context.Context, db *sql.DB) error {
	// 检查热门模型的路由配置
	popularModels := []string{"gpt-4", "gpt-3.5-turbo", "claude-3-haiku", "claude-3-sonnet", "claude-3-opus"}

	for _, modelSlug := range popularModels {
		fmt.Printf("\n🔍 检查模型: %s\n", modelSlug)

		query := `
			SELECT 
				p.name as provider_name,
				p.status as provider_status,
				p.health_status,
				pms.upstream_model_name,
				pms.enabled,
				pms.priority
			FROM provider_model_support pms
			JOIN providers p ON pms.provider_id = p.id
			WHERE pms.model_slug = ?
			ORDER BY pms.priority ASC, p.priority ASC
		`

		rows, err := db.QueryContext(ctx, query, modelSlug)
		if err != nil {
			return fmt.Errorf("failed to query model routing for %s: %w", modelSlug, err)
		}

		supportCount := 0
		availableCount := 0

		for rows.Next() {
			var providerName, providerStatus, healthStatus string
			var upstreamModel *string
			var enabled bool
			var priority int

			err := rows.Scan(&providerName, &providerStatus, &healthStatus, &upstreamModel, &enabled, &priority)
			if err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan model routing: %w", err)
			}

			supportCount++

			statusIcon := "❌"
			if enabled && providerStatus == "active" {
				statusIcon = "✅"
				availableCount++
			}

			upstream := modelSlug
			if upstreamModel != nil && *upstreamModel != "" {
				upstream = *upstreamModel
			}

			fmt.Printf("   %s %s -> %s (Priority: %d, Health: %s)\n", 
				statusIcon, providerName, upstream, priority, healthStatus)
		}
		rows.Close()

		if supportCount == 0 {
			fmt.Printf("   ❌ 没有提供商支持此模型\n")
		} else if availableCount == 0 {
			fmt.Printf("   ⚠️  有 %d 个提供商支持，但都不可用\n", supportCount)
		} else {
			fmt.Printf("   ✅ %d/%d 个提供商可用\n", availableCount, supportCount)
		}
	}

	return nil
}

func checkSystemReadiness(ctx context.Context, db *sql.DB) error {
	checks := []struct {
		name  string
		query string
		min   int
	}{
		{"活跃提供商", "SELECT COUNT(*) FROM providers WHERE status = 'active'", 1},
		{"活跃模型", "SELECT COUNT(*) FROM models WHERE status = 'active'", 1},
		{"模型支持映射", "SELECT COUNT(*) FROM provider_model_support WHERE enabled = true", 1},
		{"定价记录", "SELECT COUNT(*) FROM model_pricing", 1},
	}

	allPassed := true

	for _, check := range checks {
		var count int
		err := db.QueryRowContext(ctx, check.query).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to execute check '%s': %w", check.name, err)
		}

		status := "✅"
		if count < check.min {
			status = "❌"
			allPassed = false
		}

		fmt.Printf("%s %s: %d (最少需要: %d)\n", status, check.name, count, check.min)
	}

	fmt.Println()
	if allPassed {
		fmt.Println("🎉 系统就绪! 可以开始处理API请求")
		fmt.Println()
		fmt.Println("💡 下一步:")
		fmt.Println("   1. 启动网关服务: go run cmd/server/main.go")
		fmt.Println("   2. 运行E2E测试: go run cmd/e2etest/main.go -action=setup")
		fmt.Println("   3. 测试API调用: go run cmd/e2etest/main.go -action=test -apikey=YOUR_API_KEY")
	} else {
		fmt.Println("❌ 系统未就绪! 请检查上述失败项")
		fmt.Println()
		fmt.Println("💡 可能的解决方案:")
		fmt.Println("   1. 运行迁移: go run cmd/migrate/main.go -direction=up")
		fmt.Println("   2. 添加提供商: 手动插入providers表或使用管理工具")
		fmt.Println("   3. 配置模型支持: go run cmd/modelsupport/main.go")
	}

	return nil
}
