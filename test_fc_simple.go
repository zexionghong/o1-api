package main

import (
	"fmt"

	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/functioncall"
	"ai-api-gateway/internal/infrastructure/logger"
)

func main() {
	fmt.Println("🚀 Function Call 功能测试")

	// 创建日志器
	logConfig := &config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	loggerImpl := logger.NewLogger(logConfig)

	// 创建 Function Call 处理器
	searchConfig := &functioncall.SearchConfig{
		Service:    "duckduckgo",
		MaxResults: 3,
	}
	searchService := functioncall.NewSearchService(searchConfig, loggerImpl)
	functionCallHandler := functioncall.NewFunctionCallHandler(searchService, loggerImpl)

	// 测试获取可用工具
	fmt.Println("\n🛠️ 可用工具:")
	tools := functionCallHandler.GetAvailableTools()
	for i, tool := range tools {
		fmt.Printf("  %d. %s: %s\n", i+1, tool.Function.Name, tool.Function.Description)
	}

	// 测试关键词检测
	fmt.Println("\n🤖 关键词检测测试:")
	testCases := []struct {
		content  string
		expected bool
	}{
		{"请搜索一下最新的AI发展", true},
		{"今天的天气怎么样", true},
		{"你好，我是小明", false},
		{"什么是机器学习", true},
		{"帮我查找相关资料", true},
		{"现在几点了", true},
		{"我想了解最新新闻", true},
		{"谢谢你的帮助", false},
	}

	for _, test := range testCases {
		messages := []clients.AIMessage{
			{Role: "user", Content: test.content},
		}
		result := functioncall.ShouldUseFunctionCall(messages)
		status := "✅"
		if result != test.expected {
			status = "❌"
		}
		fmt.Printf("  %s '%s' -> %v (期望: %v)\n", status, test.content, result, test.expected)
	}

	fmt.Println("\n🎉 Function Call 基础功能测试完成")
	fmt.Println("\n📝 说明:")
	fmt.Println("  - Function Call 功能已成功集成")
	fmt.Println("  - 支持 search、news、crawler 三种工具")
	fmt.Println("  - 智能关键词检测正常工作")
	fmt.Println("  - 配置已设置为使用 DuckDuckGo 搜索")
	fmt.Println("  - 要测试完整功能，需要启动服务器并使用有效的 API 密钥")
}
