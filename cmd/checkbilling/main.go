package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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

	fmt.Println("=== 检查用户余额变化 ===")
	if err := checkUserBalance(ctx, db); err != nil {
		log.Printf("Failed to check user balance: %v", err)
	}

	fmt.Println("\n=== 检查使用日志记录 ===")
	if err := checkUsageLogs(ctx, db); err != nil {
		log.Printf("Failed to check usage logs: %v", err)
	}

	fmt.Println("\n=== 检查计费记录 ===")
	if err := checkBillingRecords(ctx, db); err != nil {
		log.Printf("Failed to check billing records: %v", err)
	}
}

func checkUserBalance(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT id, username, balance, updated_at 
		FROM users 
		WHERE username = 'e2e-test-user'
	`

	var id int64
	var username string
	var balance float64
	var updatedAt string

	err := db.QueryRowContext(ctx, query).Scan(&id, &username, &balance, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ 测试用户不存在")
			return nil
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	fmt.Printf("👤 用户: %s (ID: %d)\n", username, id)
	fmt.Printf("💰 当前余额: %.8f USD\n", balance)
	fmt.Printf("🕒 最后更新: %s\n", updatedAt)

	// 检查余额是否从初始的100美元减少了
	if balance < 100.0 {
		fmt.Printf("✅ 余额已扣减: %.8f USD\n", 100.0-balance)
	} else {
		fmt.Printf("⚠️  余额未变化，可能没有执行扣费\n")
	}

	return nil
}

func checkUsageLogs(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT 
			ul.id, ul.request_id, ul.method, ul.endpoint,
			ul.input_tokens, ul.output_tokens, ul.total_tokens,
			ul.duration_ms, ul.status_code, ul.cost,
			ul.created_at,
			m.name as model_name,
			p.name as provider_name
		FROM usage_logs ul
		LEFT JOIN models m ON ul.model_id = m.id
		LEFT JOIN providers p ON ul.provider_id = p.id
		WHERE ul.user_id = (SELECT id FROM users WHERE username = 'e2e-test-user')
		ORDER BY ul.created_at DESC
		LIMIT 10
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query usage logs: %w", err)
	}
	defer rows.Close()

	logCount := 0
	totalCost := 0.0

	fmt.Printf("%-5s %-25s %-10s %-15s %-8s %-8s %-10s %-12s %-20s\n", 
		"ID", "Request ID", "Method", "Model", "Input", "Output", "Cost", "Status", "Created")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────────────────────")

	for rows.Next() {
		var id int64
		var requestID, method, endpoint string
		var inputTokens, outputTokens, totalTokens int
		var durationMs, statusCode int
		var cost float64
		var createdAt string
		var modelName, providerName *string

		err := rows.Scan(&id, &requestID, &method, &endpoint, 
			&inputTokens, &outputTokens, &totalTokens,
			&durationMs, &statusCode, &cost, &createdAt,
			&modelName, &providerName)
		if err != nil {
			return fmt.Errorf("failed to scan usage log: %w", err)
		}

		model := "Unknown"
		if modelName != nil {
			model = *modelName
		}

		fmt.Printf("%-5d %-25s %-10s %-15s %-8d %-8d %-10.8f %-12d %-20s\n",
			id, requestID, method, model, inputTokens, outputTokens, cost, statusCode, createdAt)

		logCount++
		totalCost += cost
	}

	if logCount == 0 {
		fmt.Println("❌ 没有找到使用日志记录")
	} else {
		fmt.Printf("\n📊 统计:\n")
		fmt.Printf("   📝 日志记录数: %d\n", logCount)
		fmt.Printf("   💰 总成本: %.8f USD\n", totalCost)
	}

	return rows.Err()
}

func checkBillingRecords(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT 
			br.id, br.usage_log_id, br.amount, br.currency,
			br.billing_type, br.description, br.status,
			br.processed_at, br.created_at
		FROM billing_records br
		WHERE br.user_id = (SELECT id FROM users WHERE username = 'e2e-test-user')
		ORDER BY br.created_at DESC
		LIMIT 10
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query billing records: %w", err)
	}
	defer rows.Close()

	recordCount := 0
	totalAmount := 0.0

	fmt.Printf("%-5s %-12s %-12s %-8s %-15s %-10s %-20s\n", 
		"ID", "Usage Log", "Amount", "Currency", "Type", "Status", "Created")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────")

	for rows.Next() {
		var id, usageLogID int64
		var amount float64
		var currency, billingType, status string
		var description *string
		var processedAt, createdAt *string

		err := rows.Scan(&id, &usageLogID, &amount, &currency,
			&billingType, &description, &status,
			&processedAt, &createdAt)
		if err != nil {
			return fmt.Errorf("failed to scan billing record: %w", err)
		}

		created := "NULL"
		if createdAt != nil {
			created = *createdAt
		}

		fmt.Printf("%-5d %-12d %-12.8f %-8s %-15s %-10s %-20s\n",
			id, usageLogID, amount, currency, billingType, status, created)

		recordCount++
		totalAmount += amount
	}

	if recordCount == 0 {
		fmt.Println("❌ 没有找到计费记录")
	} else {
		fmt.Printf("\n📊 统计:\n")
		fmt.Printf("   📝 计费记录数: %d\n", recordCount)
		fmt.Printf("   💰 总扣费金额: %.8f USD\n", totalAmount)
	}

	return rows.Err()
}
