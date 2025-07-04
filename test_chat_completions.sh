#!/bin/bash

# 测试 /v1/chat/completions 接口的正确参数格式

BASE_URL="http://localhost:8080"

echo "🚀 测试 /v1/chat/completions 接口"
echo "=================================="

# 1. 创建用户和 API Key
echo "1. 创建测试用户和 API Key..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/users" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "chat_test_user",
        "email": "chat@example.com",
        "balance": 100.0
    }')

API_KEY_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/api-keys" \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": 1,
        "name": "Chat Test Key"
    }')

API_KEY=$(echo "$API_KEY_RESPONSE" | grep -o '"key":"[^"]*"' | cut -d'"' -f4)
echo "API Key: $API_KEY"

# 2. 测试正确的聊天补全请求格式
echo ""
echo "2. 测试正确的聊天补全请求格式..."

echo ""
echo "✅ 基本聊天请求:"
curl -X POST "$BASE_URL/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "user",
                "content": "Hello, how are you?"
            }
        ],
        "max_tokens": 100,
        "temperature": 0.7
    }' | jq .

echo ""
echo "✅ 带系统提示的聊天请求:"
curl -X POST "$BASE_URL/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "system",
                "content": "You are a helpful assistant that responds in Chinese."
            },
            {
                "role": "user",
                "content": "介绍一下你自己"
            }
        ],
        "max_tokens": 150,
        "temperature": 0.5
    }' | jq .

echo ""
echo "✅ 多轮对话请求:"
curl -X POST "$BASE_URL/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "user",
                "content": "What is the capital of France?"
            },
            {
                "role": "assistant",
                "content": "The capital of France is Paris."
            },
            {
                "role": "user",
                "content": "What is the population of that city?"
            }
        ],
        "max_tokens": 100,
        "temperature": 0.3
    }' | jq .

# 3. 测试错误的请求格式
echo ""
echo "3. 测试错误的请求格式..."

echo ""
echo "❌ 使用 prompt 参数 (应该失败):"
curl -X POST "$BASE_URL/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{
        "model": "gpt-3.5-turbo",
        "prompt": "Hello, how are you?",
        "max_tokens": 100
    }' | jq .

echo ""
echo "❌ 缺少 messages 参数 (应该失败):"
curl -X POST "$BASE_URL/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{
        "model": "gpt-3.5-turbo",
        "max_tokens": 100
    }' | jq .

echo ""
echo "❌ 空的 messages 数组 (应该失败):"
curl -X POST "$BASE_URL/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{
        "model": "gpt-3.5-turbo",
        "messages": [],
        "max_tokens": 100
    }' | jq .

echo ""
echo "🎉 测试完成！"
echo ""
echo "📋 总结："
echo "✅ /v1/chat/completions 接口使用 messages 数组"
echo "✅ 每个消息包含 role 和 content 字段"
echo "✅ role 可以是: system, user, assistant"
echo "❌ 不要使用 prompt 参数 (那是 /v1/completions 接口的)"
echo ""
echo "🔗 在 Swagger UI 中测试: $BASE_URL/swagger/index.html"
