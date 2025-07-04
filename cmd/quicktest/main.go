package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 测试API密钥
	apiKey := "ak_ede198ed25b71c95cb9b38ac970e4f248ed2c6d1d658a19475b2afeab5cf9822"
	gatewayURL := "http://localhost:8080"

	// 构建请求
	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": "Say 'Quick test successful' and nothing else."},
		},
		"max_tokens":  20,
		"temperature": 0.1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		return
	}

	// 发送请求
	url := fmt.Sprintf("%s/v1/chat/completions", gatewayURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	fmt.Println("🚀 Sending test request...")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📊 Status: %d\n", resp.StatusCode)
	fmt.Printf("📄 Response: %s\n", string(body))

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ Request successful!")
		
		// 解析响应
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err == nil {
			if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if message, ok := choice["message"].(map[string]interface{}); ok {
						if content, ok := message["content"].(string); ok {
							fmt.Printf("💬 AI Response: %s\n", content)
						}
					}
				}
			}
			
			if usage, ok := response["usage"].(map[string]interface{}); ok {
				if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
					if completionTokens, ok := usage["completion_tokens"].(float64); ok {
						if totalTokens, ok := usage["total_tokens"].(float64); ok {
							fmt.Printf("📈 Token Usage: %d input + %d output = %d total\n", 
								int(promptTokens), int(completionTokens), int(totalTokens))
						}
					}
				}
			}
		}
	} else {
		fmt.Printf("❌ Request failed with status %d\n", resp.StatusCode)
	}
}
