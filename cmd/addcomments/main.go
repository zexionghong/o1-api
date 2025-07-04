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

	// 创建注释表
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS table_comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			table_name VARCHAR(100) NOT NULL,
			column_name VARCHAR(100), -- NULL表示表级注释
			comment_text TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(table_name, column_name)
		);
	`

	_, err = db.ExecContext(ctx, createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table_comments: %v", err)
	}

	fmt.Println("✅ Created table_comments table")

	// 插入所有表注释
	comments := []struct {
		table  string
		column *string
		text   string
	}{
		// 用户表注释
		{"users", nil, "用户表 - 存储API网关的用户信息"},
		{"users", stringPtr("id"), "用户唯一标识符"},
		{"users", stringPtr("username"), "用户名，用于登录和标识"},
		{"users", stringPtr("email"), "用户邮箱地址，用于通知和找回密码"},
		{"users", stringPtr("password_hash"), "密码哈希值，使用bcrypt加密"},
		{"users", stringPtr("full_name"), "用户全名或显示名称"},
		{"users", stringPtr("status"), "用户状态：active-活跃, inactive-非活跃, suspended-暂停"},
		{"users", stringPtr("balance"), "用户账户余额（美元），用于API调用扣费"},
		{"users", stringPtr("created_at"), "用户创建时间"},
		{"users", stringPtr("updated_at"), "用户信息最后更新时间"},

		// API密钥表注释
		{"api_keys", nil, "API密钥表 - 存储用户的API访问密钥"},
		{"api_keys", stringPtr("id"), "API密钥唯一标识符"},
		{"api_keys", stringPtr("user_id"), "关联的用户ID"},
		{"api_keys", stringPtr("key_hash"), "API密钥哈希值，用于身份验证"},
		{"api_keys", stringPtr("key_prefix"), "API密钥前缀，用于快速识别"},
		{"api_keys", stringPtr("name"), "API密钥名称，便于用户管理"},
		{"api_keys", stringPtr("status"), "API密钥状态：active-活跃, inactive-非活跃, revoked-已撤销"},
		{"api_keys", stringPtr("last_used_at"), "API密钥最后使用时间"},
		{"api_keys", stringPtr("expires_at"), "API密钥过期时间，NULL表示永不过期"},
		{"api_keys", stringPtr("created_at"), "API密钥创建时间"},
		{"api_keys", stringPtr("updated_at"), "API密钥最后更新时间"},

		// 提供商表注释
		{"providers", nil, "上游AI服务提供商表 - 存储OpenAI、Anthropic等AI服务商信息"},
		{"providers", stringPtr("id"), "提供商唯一标识符"},
		{"providers", stringPtr("name"), "提供商名称，如OpenAI、Anthropic"},
		{"providers", stringPtr("slug"), "提供商标识符，用于URL和配置"},
		{"providers", stringPtr("base_url"), "提供商API基础URL"},
		{"providers", stringPtr("status"), "提供商状态：active-活跃, inactive-非活跃, maintenance-维护中"},
		{"providers", stringPtr("health_status"), "健康检查状态：healthy-健康, unhealthy-不健康, unknown-未知"},
		{"providers", stringPtr("priority"), "提供商优先级，数字越小优先级越高"},
		{"providers", stringPtr("timeout_seconds"), "请求超时时间（秒）"},
		{"providers", stringPtr("retry_attempts"), "请求失败重试次数"},
		{"providers", stringPtr("health_check_interval"), "健康检查间隔（秒）"},
		{"providers", stringPtr("created_at"), "提供商创建时间"},
		{"providers", stringPtr("updated_at"), "提供商信息最后更新时间"},

		// 模型表注释
		{"models", nil, "AI模型表 - 存储各提供商支持的AI模型信息"},
		{"models", stringPtr("id"), "模型唯一标识符"},
		{"models", stringPtr("provider_id"), "关联的提供商ID"},
		{"models", stringPtr("name"), "模型名称，如GPT-4、Claude-3"},
		{"models", stringPtr("slug"), "模型标识符，用于API请求"},
		{"models", stringPtr("display_name"), "模型显示名称，用于前端展示"},
		{"models", stringPtr("description"), "模型描述信息"},
		{"models", stringPtr("model_type"), "模型类型：chat-对话, completion-补全, embedding-嵌入, image-图像"},
		{"models", stringPtr("context_length"), "模型上下文长度（token数）"},
		{"models", stringPtr("max_tokens"), "模型最大输出token数"},
		{"models", stringPtr("supports_streaming"), "是否支持流式响应"},
		{"models", stringPtr("supports_functions"), "是否支持函数调用"},
		{"models", stringPtr("status"), "模型状态：active-活跃, inactive-非活跃, deprecated-已弃用"},
		{"models", stringPtr("created_at"), "模型创建时间"},
		{"models", stringPtr("updated_at"), "模型信息最后更新时间"},

		// 模型定价表注释
		{"model_pricing", nil, "模型定价表 - 存储各模型的价格信息"},
		{"model_pricing", stringPtr("id"), "定价记录唯一标识符"},
		{"model_pricing", stringPtr("model_id"), "关联的模型ID"},
		{"model_pricing", stringPtr("pricing_type"), "定价类型：input-输入token, output-输出token, request-请求次数"},
		{"model_pricing", stringPtr("price_per_unit"), "单位价格（美元）"},
		{"model_pricing", stringPtr("unit"), "计价单位：token-按token计费, request-按请求计费, character-按字符计费"},
		{"model_pricing", stringPtr("currency"), "货币类型，默认USD"},
		{"model_pricing", stringPtr("effective_from"), "价格生效开始时间"},
		{"model_pricing", stringPtr("effective_until"), "价格生效结束时间，NULL表示永久有效"},
		{"model_pricing", stringPtr("created_at"), "定价记录创建时间"},

		// 提供商模型支持表注释
		{"provider_model_support", nil, "提供商模型支持表 - 定义哪些提供商支持哪些模型（多对多关系）"},
		{"provider_model_support", stringPtr("id"), "支持记录唯一标识符"},
		{"provider_model_support", stringPtr("provider_id"), "提供商ID"},
		{"provider_model_support", stringPtr("model_slug"), "模型标识符，用户请求时使用的模型名"},
		{"provider_model_support", stringPtr("upstream_model_name"), "上游实际模型名，可能与model_slug不同"},
		{"provider_model_support", stringPtr("enabled"), "是否启用此模型支持"},
		{"provider_model_support", stringPtr("priority"), "该提供商对此模型的优先级，数字越小优先级越高"},
		{"provider_model_support", stringPtr("config"), "JSON格式的额外配置，如参数映射、自定义端点等"},
		{"provider_model_support", stringPtr("created_at"), "支持记录创建时间"},
		{"provider_model_support", stringPtr("updated_at"), "支持记录最后更新时间"},

		// 配额表注释
		{"quotas", nil, "用户配额表 - 定义用户的API使用限制"},
		{"quotas", stringPtr("id"), "配额记录唯一标识符"},
		{"quotas", stringPtr("user_id"), "关联的用户ID"},
		{"quotas", stringPtr("quota_type"), "配额类型：daily-每日, monthly-每月, total-总计"},
		{"quotas", stringPtr("limit_value"), "配额限制值"},
		{"quotas", stringPtr("period"), "配额周期：minute-分钟, hour-小时, day-天, month-月"},
		{"quotas", stringPtr("reset_time"), "重置时间点"},
		{"quotas", stringPtr("status"), "配额状态：active-生效, inactive-停用"},
		{"quotas", stringPtr("created_at"), "配额创建时间"},
		{"quotas", stringPtr("updated_at"), "配额最后更新时间"},

		// 配额使用表注释
		{"quota_usage", nil, "配额使用表 - 记录用户配额的实际使用情况"},
		{"quota_usage", stringPtr("id"), "使用记录唯一标识符"},
		{"quota_usage", stringPtr("quota_id"), "关联的配额ID"},
		{"quota_usage", stringPtr("user_id"), "关联的用户ID"},
		{"quota_usage", stringPtr("used_value"), "已使用的配额值"},
		{"quota_usage", stringPtr("period_start"), "统计周期开始时间"},
		{"quota_usage", stringPtr("period_end"), "统计周期结束时间"},
		{"quota_usage", stringPtr("created_at"), "使用记录创建时间"},
		{"quota_usage", stringPtr("updated_at"), "使用记录最后更新时间"},

		// 使用日志表注释
		{"usage_logs", nil, "API使用日志表 - 记录每次API调用的详细信息"},
		{"usage_logs", stringPtr("id"), "日志记录唯一标识符"},
		{"usage_logs", stringPtr("user_id"), "调用用户ID"},
		{"usage_logs", stringPtr("api_key_id"), "使用的API密钥ID"},
		{"usage_logs", stringPtr("provider_id"), "实际使用的提供商ID"},
		{"usage_logs", stringPtr("model_id"), "使用的模型ID"},
		{"usage_logs", stringPtr("request_id"), "请求唯一标识符，用于追踪"},
		{"usage_logs", stringPtr("method"), "HTTP请求方法"},
		{"usage_logs", stringPtr("endpoint"), "请求的API端点"},
		{"usage_logs", stringPtr("input_tokens"), "输入token数量"},
		{"usage_logs", stringPtr("output_tokens"), "输出token数量"},
		{"usage_logs", stringPtr("total_tokens"), "总token数量"},
		{"usage_logs", stringPtr("request_size"), "请求体大小（字节）"},
		{"usage_logs", stringPtr("response_size"), "响应体大小（字节）"},
		{"usage_logs", stringPtr("duration_ms"), "请求处理时间（毫秒）"},
		{"usage_logs", stringPtr("status_code"), "HTTP响应状态码"},
		{"usage_logs", stringPtr("error_message"), "错误信息，成功时为空"},
		{"usage_logs", stringPtr("cost"), "本次调用的费用（美元）"},
		{"usage_logs", stringPtr("created_at"), "日志创建时间"},

		// 计费记录表注释
		{"billing_records", nil, "计费记录表 - 记录用户的扣费和充值记录"},
		{"billing_records", stringPtr("id"), "计费记录唯一标识符"},
		{"billing_records", stringPtr("user_id"), "关联的用户ID"},
		{"billing_records", stringPtr("usage_log_id"), "关联的使用日志ID，充值时为NULL"},
		{"billing_records", stringPtr("amount"), "金额，正数表示扣费，负数表示充值"},
		{"billing_records", stringPtr("currency"), "货币类型"},
		{"billing_records", stringPtr("billing_type"), "计费类型：usage-使用扣费, recharge-充值, refund-退款"},
		{"billing_records", stringPtr("description"), "计费描述信息"},
		{"billing_records", stringPtr("status"), "计费状态：pending-待处理, processed-已处理, failed-失败"},
		{"billing_records", stringPtr("processed_at"), "处理时间"},
		{"billing_records", stringPtr("created_at"), "计费记录创建时间"},

		// 注释表自身的注释
		{"table_comments", nil, "表注释表 - 存储数据库表和字段的说明信息"},
		{"table_comments", stringPtr("id"), "注释记录唯一标识符"},
		{"table_comments", stringPtr("table_name"), "表名"},
		{"table_comments", stringPtr("column_name"), "字段名，NULL表示表级注释"},
		{"table_comments", stringPtr("comment_text"), "注释内容"},
		{"table_comments", stringPtr("created_at"), "注释创建时间"},
	}

	// 插入注释
	insertSQL := "INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES (?, ?, ?)"
	for _, comment := range comments {
		_, err := db.ExecContext(ctx, insertSQL, comment.table, comment.column, comment.text)
		if err != nil {
			log.Printf("Failed to insert comment for %s.%v: %v", comment.table, comment.column, err)
		}
	}

	fmt.Printf("✅ Inserted %d table and column comments\n", len(comments))
	fmt.Println("🎉 Database comments added successfully!")
}

func stringPtr(s string) *string {
	return &s
}
