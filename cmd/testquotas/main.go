package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	// 测试API密钥
	apiKey := "ak_ede198ed25b71c95cb9b38ac970e4f248ed2c6d1d658a19475b2afeab5cf9822"
	gatewayURL := "http://localhost:8080"

	fmt.Println("🧪 Testing quota limits...")
	fmt.Println("📋 Current quota settings:")
	fmt.Println("   • 每分钟最多10次请求")
	fmt.Println("   • 每分钟最多1000个token")
	fmt.Println("   • 每分钟最多花费0.1美元")
	fmt.Println()

	// 测试1: 正常请求（应该成功）
	fmt.Println("🔍 Test 1: Normal request (should succeed)")
	err := sendTestRequest(gatewayURL, apiKey, 1)
	if err != nil {
		fmt.Printf("❌ Test 1 failed: %v\n", err)
	} else {
		fmt.Println("✅ Test 1 passed: Normal request succeeded")
	}

	fmt.Println()

	// 测试2: 快速连续请求（测试请求数量限制）
	fmt.Println("🔍 Test 2: Rapid requests (testing request quota)")
	fmt.Println("   Sending 12 requests rapidly (quota: 10/minute)...")

	successCount := 0
	quotaExceededCount := 0

	for i := 1; i <= 12; i++ {
		fmt.Printf("   Request %d: ", i)
		err := sendTestRequest(gatewayURL, apiKey, i)
		if err != nil {
			if isQuotaExceededError(err) {
				fmt.Printf("❌ Quota exceeded (expected after 10 requests)\n")
				quotaExceededCount++
			} else {
				fmt.Printf("❌ Error: %v\n", err)
			}
		} else {
			fmt.Printf("✅ Success\n")
			successCount++
		}

		// 短暂延迟避免网络问题
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n📊 Results: %d successful, %d quota exceeded\n", successCount, quotaExceededCount)

	if successCount <= 10 && quotaExceededCount >= 2 {
		fmt.Println("✅ Test 2 passed: Request quota working correctly")
	} else {
		fmt.Println("❌ Test 2 failed: Request quota not working as expected")
	}

	fmt.Println()

	// 测试3: 等待一分钟后再次测试（配额应该重置）
	fmt.Println("🔍 Test 3: Quota reset test")
	fmt.Println("   Waiting 65 seconds for quota reset...")

	// 显示倒计时
	for i := 65; i > 0; i-- {
		fmt.Printf("\r   Countdown: %d seconds remaining...", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("\r   Countdown: Complete!                    \n")

	fmt.Println("   Testing request after quota reset...")
	err = sendTestRequest(gatewayURL, apiKey, 999)
	if err != nil {
		fmt.Printf("❌ Test 3 failed: %v\n", err)
	} else {
		fmt.Println("✅ Test 3 passed: Request succeeded after quota reset")
	}

	fmt.Println()
	fmt.Println("🎉 Quota testing completed!")
}

func sendTestRequest(gatewayURL, apiKey string, requestNum int) error {
	// 构建请求
	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": fmt.Sprintf("Test request #%d. Say 'OK' only.", requestNum)},
		},
		"max_tokens":  5,
		"temperature": 0.1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s/v1/chat/completions", gatewayURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func isQuotaExceededError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "quota") ||
		strings.Contains(errStr, "QUOTA_EXCEEDED") ||
		strings.Contains(errStr, "Too Many Requests")
}
