package main

import (
	"context"
	"fmt"

	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/functioncall"
	"ai-api-gateway/internal/infrastructure/logger"
)

func main() {
	fmt.Println("🚀 测试 Google 搜索功能")

	// 创建日志器
	logConfig := &config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	loggerImpl := logger.NewLogger(logConfig)

	// 创建搜索配置
	searchConfig := &functioncall.SearchConfig{
		Service:     "google",
		MaxResults:  3,
		GoogleCX:    "05afc7eed6abd4a3c",
		GoogleKey:   "AIzaSyDkYBKg1k2U8zTm0LPQlVIIGNRNrLmCvK4",
	}

	// 创建搜索服务
	searchService := functioncall.NewSearchService(searchConfig, loggerImpl)

	// 测试搜索
	ctx := context.Background()
	result, err := searchService.Search(ctx, "人工智能")
	if err != nil {
		fmt.Printf("❌ 搜索失败: %v\n", err)
	} else {
		fmt.Printf("✅ 搜索成功！\n")
		fmt.Printf("结果长度: %d 字符\n", len(result))
		fmt.Printf("结果预览: %.300s...\n", result)
	}

	// 测试关键词检测
	fmt.Println("\n🤖 测试关键词检测:")
	testCases := []string{
		"请搜索一下最新的AI发展",
		"今天的天气怎么样",
		"你好，我是小明",
	}

	for _, content := range testCases {
		messages := []clients.AIMessage{
			{Role: "user", Content: content},
		}
		shouldUse := functioncall.ShouldUseFunctionCall(messages)
		fmt.Printf("  '%s' -> %v\n", content, shouldUse)
	}

	fmt.Println("\n🎉 测试完成")
}
