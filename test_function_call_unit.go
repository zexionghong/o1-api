package main

import (
	"context"
	"fmt"
	"log"

	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/functioncall"
	"ai-api-gateway/internal/infrastructure/logger"
)

func main() {
	fmt.Println("🚀 测试 Function Call 核心功能")

	// 创建日志器
	logConfig := &config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	loggerImpl := logger.NewLogger(logConfig)

	// 创建搜索配置
	searchConfig := &functioncall.SearchConfig{
		Service:    "google",
		MaxResults: 5,
		GoogleCX:   "05afc7eed6abd4a3c",
		GoogleKey:  "AIzaSyDkYBKg1k2U8zTm0LPQlVIIGNRNrLmCvK4",
	}

	// 创建搜索服务
	searchService := functioncall.NewSearchService(searchConfig, loggerImpl)

	// 创建 Function Call 处理器
	functionCallHandler := functioncall.NewFunctionCallHandler(searchService, loggerImpl)

	// 测试搜索功能
	fmt.Println("\n🔍 测试搜索功能...")
	ctx := context.Background()

	searchResult, err := searchService.Search(ctx, "人工智能最新发展")
	if err != nil {
		log.Printf("❌ 搜索失败: %v", err)
	} else {
		fmt.Printf("✅ 搜索成功，结果长度: %d 字符\n", len(searchResult))
		fmt.Printf("搜索结果预览: %.200s...\n", searchResult)
	}

	// 测试获取可用工具
	fmt.Println("\n🛠️ 测试获取可用工具...")
	tools := functionCallHandler.GetAvailableTools()
	fmt.Printf("✅ 可用工具数量: %d\n", len(tools))
	for i, tool := range tools {
		fmt.Printf("  %d. %s: %s\n", i+1, tool.Function.Name, tool.Function.Description)
	}

	// 测试关键词检测
	fmt.Println("\n🤖 测试关键词检测...")
	testMessages := []struct {
		content  string
		expected bool
	}{
		{"请搜索一下最新的AI发展", true},
		{"今天的天气怎么样", true},
		{"你好，我是小明", false},
		{"什么是机器学习", true},
		{"帮我查找相关资料", true},
	}

	for _, test := range testMessages {
		messages := []clients.AIMessage{
			{Role: "user", Content: test.content},
		}
		result := functioncall.ShouldUseFunctionCall(messages)
		status := "❌"
		if result == test.expected {
			status = "✅"
		}
		fmt.Printf("  %s '%s' -> %v (期望: %v)\n", status, test.content, result, test.expected)
	}

	fmt.Println("\n🎉 Function Call 核心功能测试完成")
}
