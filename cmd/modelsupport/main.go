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
		dbPath   = flag.String("db", "./data/gateway.db", "Database file path")
		action   = flag.String("action", "show", "Action: show, add, remove, test")
		provider = flag.Int64("provider", 0, "Provider ID")
		model    = flag.String("model", "", "Model slug")
		upstream = flag.String("upstream", "", "Upstream model name")
		priority = flag.Int("priority", 1, "Priority (lower number = higher priority)")
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
	case "show":
		if err := showModelSupport(ctx, db); err != nil {
			log.Fatalf("Failed to show model support: %v", err)
		}
	case "add":
		if *provider == 0 || *model == "" {
			log.Fatalf("Provider ID and model slug are required for add action")
		}
		if err := addModelSupport(ctx, db, *provider, *model, *upstream, *priority); err != nil {
			log.Fatalf("Failed to add model support: %v", err)
		}
	case "remove":
		if *provider == 0 || *model == "" {
			log.Fatalf("Provider ID and model slug are required for remove action")
		}
		if err := removeModelSupport(ctx, db, *provider, *model); err != nil {
			log.Fatalf("Failed to remove model support: %v", err)
		}
	case "test":
		if *model == "" {
			log.Fatalf("Model slug is required for test action")
		}
		if err := testModelRouting(ctx, db, *model); err != nil {
			log.Fatalf("Failed to test model routing: %v", err)
		}
	default:
		fmt.Printf("Invalid action: %s\n", *action)
		fmt.Println("Available actions: show, add, remove, test")
	}
}

func showModelSupport(ctx context.Context, db *sql.DB) error {
	fmt.Println("=== Provider Model Support Mapping ===")

	// 查询所有提供商
	providerQuery := `
		SELECT id, name, slug, status, health_status, priority 
		FROM providers 
		ORDER BY priority ASC, name ASC
	`

	providerRows, err := db.QueryContext(ctx, providerQuery)
	if err != nil {
		return fmt.Errorf("failed to query providers: %w", err)
	}
	defer providerRows.Close()

	for providerRows.Next() {
		var provider struct {
			ID           int64
			Name         string
			Slug         string
			Status       string
			HealthStatus string
			Priority     int
		}

		err := providerRows.Scan(&provider.ID, &provider.Name, &provider.Slug,
			&provider.Status, &provider.HealthStatus, &provider.Priority)
		if err != nil {
			return fmt.Errorf("failed to scan provider: %w", err)
		}

		fmt.Printf("\n🏢 Provider: %s (%s)\n", provider.Name, provider.Slug)
		fmt.Printf("   Status: %s, Health: %s, Priority: %d\n",
			provider.Status, provider.HealthStatus, provider.Priority)

		// 查询该提供商支持的模型
		supportQuery := `
			SELECT model_slug, upstream_model_name, enabled, priority, config
			FROM provider_model_support 
			WHERE provider_id = ?
			ORDER BY priority ASC, model_slug ASC
		`

		supportRows, err := db.QueryContext(ctx, supportQuery, provider.ID)
		if err != nil {
			fmt.Printf("   ❌ Failed to query model support: %v\n", err)
			continue
		}

		supports := []struct {
			ModelSlug         string
			UpstreamModelName *string
			Enabled           bool
			Priority          int
			Config            *string
		}{}

		for supportRows.Next() {
			var support struct {
				ModelSlug         string
				UpstreamModelName *string
				Enabled           bool
				Priority          int
				Config            *string
			}

			err := supportRows.Scan(&support.ModelSlug, &support.UpstreamModelName,
				&support.Enabled, &support.Priority, &support.Config)
			if err != nil {
				fmt.Printf("   ❌ Failed to scan support: %v\n", err)
				continue
			}

			supports = append(supports, support)
		}
		supportRows.Close()

		if len(supports) == 0 {
			fmt.Printf("   📭 No model support configured\n")
			continue
		}

		fmt.Printf("   📋 Supported Models (%d):\n", len(supports))
		for _, support := range supports {
			status := "✅"
			if !support.Enabled {
				status = "❌"
			}

			upstream := support.ModelSlug
			if support.UpstreamModelName != nil && *support.UpstreamModelName != "" {
				upstream = *support.UpstreamModelName
			}

			fmt.Printf("      %s %s -> %s (Priority: %d)\n",
				status, support.ModelSlug, upstream, support.Priority)
		}
	}

	return nil
}

func addModelSupport(ctx context.Context, db *sql.DB, providerID int64, modelSlug, upstreamModel string, priority int) error {
	// 检查提供商是否存在
	var providerName string
	err := db.QueryRowContext(ctx, "SELECT name FROM providers WHERE id = ?", providerID).Scan(&providerName)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("provider with ID %d not found", providerID)
		}
		return fmt.Errorf("failed to check provider: %w", err)
	}

	// 设置上游模型名
	var upstreamModelPtr *string
	if upstreamModel != "" && upstreamModel != modelSlug {
		upstreamModelPtr = &upstreamModel
	}

	// 插入或更新模型支持
	query := `
		INSERT INTO provider_model_support (provider_id, model_slug, upstream_model_name, priority, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(provider_id, model_slug) DO UPDATE SET
			upstream_model_name = excluded.upstream_model_name,
			priority = excluded.priority,
			enabled = true,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = db.ExecContext(ctx, query, providerID, modelSlug, upstreamModelPtr, priority)
	if err != nil {
		return fmt.Errorf("failed to add model support: %w", err)
	}

	upstreamDisplay := modelSlug
	if upstreamModelPtr != nil {
		upstreamDisplay = *upstreamModelPtr
	}

	fmt.Printf("✅ Added model support: %s supports '%s' -> '%s' (Priority: %d)\n",
		providerName, modelSlug, upstreamDisplay, priority)

	return nil
}

func removeModelSupport(ctx context.Context, db *sql.DB, providerID int64, modelSlug string) error {
	// 检查支持是否存在
	var exists bool
	err := db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM provider_model_support WHERE provider_id = ? AND model_slug = ?)",
		providerID, modelSlug).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check model support: %w", err)
	}

	if !exists {
		return fmt.Errorf("model support not found: provider %d, model %s", providerID, modelSlug)
	}

	// 删除模型支持
	_, err = db.ExecContext(ctx,
		"DELETE FROM provider_model_support WHERE provider_id = ? AND model_slug = ?",
		providerID, modelSlug)
	if err != nil {
		return fmt.Errorf("failed to remove model support: %w", err)
	}

	fmt.Printf("✅ Removed model support: provider %d no longer supports '%s'\n",
		providerID, modelSlug)

	return nil
}

func testModelRouting(ctx context.Context, db *sql.DB, modelSlug string) error {
	fmt.Printf("=== Testing Model Routing for '%s' ===\n", modelSlug)

	// 查询支持该模型的提供商
	query := `
		SELECT 
			p.id, p.name, p.slug, p.status, p.health_status, p.priority as provider_priority,
			pms.upstream_model_name, pms.enabled, pms.priority as model_priority
		FROM provider_model_support pms
		JOIN providers p ON pms.provider_id = p.id
		WHERE pms.model_slug = ? AND pms.enabled = true AND p.status = 'active'
		ORDER BY pms.priority ASC, p.priority ASC
	`

	rows, err := db.QueryContext(ctx, query, modelSlug)
	if err != nil {
		return fmt.Errorf("failed to query supporting providers: %w", err)
	}
	defer rows.Close()

	var providers []struct {
		ID               int64
		Name             string
		Slug             string
		Status           string
		HealthStatus     string
		ProviderPriority int
		UpstreamModel    *string
		Enabled          bool
		ModelPriority    int
	}

	for rows.Next() {
		var provider struct {
			ID               int64
			Name             string
			Slug             string
			Status           string
			HealthStatus     string
			ProviderPriority int
			UpstreamModel    *string
			Enabled          bool
			ModelPriority    int
		}

		err := rows.Scan(&provider.ID, &provider.Name, &provider.Slug,
			&provider.Status, &provider.HealthStatus, &provider.ProviderPriority,
			&provider.UpstreamModel, &provider.Enabled, &provider.ModelPriority)
		if err != nil {
			return fmt.Errorf("failed to scan provider: %w", err)
		}

		providers = append(providers, provider)
	}

	if len(providers) == 0 {
		fmt.Printf("❌ No providers support model '%s'\n", modelSlug)
		return nil
	}

	fmt.Printf("✅ Found %d provider(s) supporting model '%s':\n", len(providers), modelSlug)
	fmt.Printf("%-20s %-15s %-10s %-15s %-15s\n", "Provider", "Health", "Priority", "Upstream Model", "Model Priority")
	fmt.Println(strings.Repeat("-", 80))

	for _, provider := range providers {
		upstreamModel := modelSlug
		if provider.UpstreamModel != nil && *provider.UpstreamModel != "" {
			upstreamModel = *provider.UpstreamModel
		}

		fmt.Printf("%-20s %-15s %-10d %-15s %-15d\n",
			provider.Name,
			provider.HealthStatus,
			provider.ProviderPriority,
			upstreamModel,
			provider.ModelPriority,
		)
	}

	// 选择最高优先级的提供商
	selectedProvider := providers[0]
	upstreamModel := modelSlug
	if selectedProvider.UpstreamModel != nil && *selectedProvider.UpstreamModel != "" {
		upstreamModel = *selectedProvider.UpstreamModel
	}

	fmt.Printf("\n🎯 Selected Provider: %s\n", selectedProvider.Name)
	fmt.Printf("   Request would be routed to: %s\n", selectedProvider.Name)
	fmt.Printf("   Upstream model name: %s\n", upstreamModel)

	return nil
}
