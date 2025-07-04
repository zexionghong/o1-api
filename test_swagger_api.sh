#!/bin/bash

# AI API Gateway Swagger API 测试脚本
# 演示如何使用 API key 进行认证和调试

set -e

BASE_URL="http://localhost:8080"
echo "🚀 AI API Gateway Swagger API 测试"
echo "=================================="
echo "Base URL: $BASE_URL"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}💡 $1${NC}"
}

# 检查服务器状态
print_step "检查服务器状态"
if curl -s "$BASE_URL/health/ready" > /dev/null; then
    print_success "服务器运行正常"
else
    print_error "服务器未运行，请先启动服务器: go run cmd/server/main.go"
    exit 1
fi

# 1. 创建用户
print_step "1. 创建测试用户"
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/users" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "swagger_test_user",
        "email": "swagger@example.com",
        "balance": 100.0
    }')

if echo "$USER_RESPONSE" | grep -q '"success":true'; then
    USER_ID=$(echo "$USER_RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2)
    print_success "用户创建成功，ID: $USER_ID"
else
    print_info "用户可能已存在，继续测试..."
    USER_ID=1
fi

# 2. 创建 API Key
print_step "2. 创建 API Key"
API_KEY_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/api-keys" \
    -H "Content-Type: application/json" \
    -d "{
        \"user_id\": $USER_ID,
        \"name\": \"Swagger Test Key\"
    }")

if echo "$API_KEY_RESPONSE" | grep -q '"success":true'; then
    API_KEY=$(echo "$API_KEY_RESPONSE" | grep -o '"key":"[^"]*"' | cut -d'"' -f4)
    print_success "API Key 创建成功"
    print_info "API Key: $API_KEY"
else
    print_error "API Key 创建失败"
    echo "响应: $API_KEY_RESPONSE"
    exit 1
fi

echo ""
print_step "3. 测试 Swagger 文档访问"
print_info "Swagger UI: $BASE_URL/swagger/index.html"
print_info "Swagger JSON: $BASE_URL/swagger/doc.json"

# 测试无需认证的接口
echo ""
print_step "4. 测试无需认证的接口"

echo "   健康检查:"
curl -s "$BASE_URL/health" | head -c 100
echo "..."

echo ""
echo "   模型列表 (无认证，应该失败):"
MODELS_RESPONSE=$(curl -s -w "%{http_code}" "$BASE_URL/v1/models")
HTTP_CODE="${MODELS_RESPONSE: -3}"
if [ "$HTTP_CODE" = "401" ]; then
    print_success "正确返回 401 未认证错误"
else
    print_error "预期返回 401，实际返回 $HTTP_CODE"
fi

# 测试需要认证的接口
echo ""
print_step "5. 测试需要认证的接口"

echo "   使用 API Key 获取模型列表:"
MODELS_AUTH_RESPONSE=$(curl -s -w "%{http_code}" "$BASE_URL/v1/models" \
    -H "Authorization: Bearer $API_KEY")
HTTP_CODE="${MODELS_AUTH_RESPONSE: -3}"
if [ "$HTTP_CODE" = "200" ]; then
    print_success "认证成功，获取模型列表"
    echo "${MODELS_AUTH_RESPONSE%???}" | head -c 100
    echo "..."
else
    print_error "认证失败，HTTP 状态码: $HTTP_CODE"
fi

echo ""
echo "   获取使用统计:"
USAGE_RESPONSE=$(curl -s -w "%{http_code}" "$BASE_URL/v1/usage" \
    -H "Authorization: Bearer $API_KEY")
HTTP_CODE="${USAGE_RESPONSE: -3}"
if [ "$HTTP_CODE" = "200" ]; then
    print_success "获取使用统计成功"
    echo "${USAGE_RESPONSE%???}"
else
    print_error "获取使用统计失败，HTTP 状态码: $HTTP_CODE"
fi

# Swagger 使用说明
echo ""
print_step "6. Swagger UI 使用说明"
echo ""
print_info "现在您可以在 Swagger UI 中测试 API："
echo ""
echo "1. 打开浏览器访问: $BASE_URL/swagger/index.html"
echo "2. 点击右上角的 'Authorize' 按钮"
echo "3. 输入: Bearer $API_KEY"
echo "4. 点击 'Authorize' 确认"
echo "5. 现在可以测试需要认证的 API 接口了！"
echo ""
print_info "推荐测试的接口："
echo "   • POST /v1/chat/completions - 聊天补全"
echo "   • GET /v1/models - 模型列表"
echo "   • GET /v1/usage - 使用统计"
echo "   • GET /health/* - 健康检查接口"
echo ""
print_success "测试完成！享受使用 Swagger 文档调试 API 吧！"
