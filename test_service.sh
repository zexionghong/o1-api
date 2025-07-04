#!/bin/bash

# AI API Gateway 服务测试脚本
# 使用curl测试各种API端点

# 配置
BASE_URL="http://localhost:8080"
API_KEY="ak_4278fc65b1d32cc99fe69fc25bf352261fab3aa0b08488d919dce0097b0f3915"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0

# 打印标题
print_header() {
    echo -e "\n${BLUE}============================================================${NC}"
    echo -e "${BLUE}🧪 测试: $1${NC}"
    echo -e "${BLUE}============================================================${NC}"
}

# 打印结果
print_result() {
    local success=$1
    local message=$2
    local details=$3
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$success" = "true" ]; then
        echo -e "${GREEN}✅ 成功: $message${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ 失败: $message${NC}"
    fi
    
    if [ -n "$details" ]; then
        echo -e "   详情: $details"
    fi
}

# 测试HTTP请求
test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    local data=$5
    local headers=$6
    
    local url="${BASE_URL}${endpoint}"
    local curl_cmd="curl -s -w '%{http_code}' -o /tmp/response.json"
    
    # 添加headers
    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd $headers"
    fi
    
    # 添加数据
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    # 添加方法
    if [ "$method" != "GET" ]; then
        curl_cmd="$curl_cmd -X $method"
    fi
    
    curl_cmd="$curl_cmd $url"
    
    # 执行请求
    local status_code
    status_code=$(eval $curl_cmd)
    local curl_exit_code=$?
    
    # 检查curl是否成功执行
    if [ $curl_exit_code -ne 0 ]; then
        print_result "false" "$description" "连接失败 - 服务器可能未启动"
        return 1
    fi
    
    # 检查状态码
    if [ "$status_code" = "$expected_status" ]; then
        print_result "true" "$description (状态码: $status_code)"
        return 0
    else
        local response_content=""
        if [ -f "/tmp/response.json" ]; then
            response_content=$(cat /tmp/response.json | head -c 100)
        fi
        print_result "false" "$description (状态码: $status_code)" "$response_content"
        return 1
    fi
}

# 测试健康检查
test_health_check() {
    print_header "健康检查"
    test_endpoint "GET" "/health" "200" "健康检查"
    
    # 显示响应内容
    if [ -f "/tmp/response.json" ]; then
        local status=$(cat /tmp/response.json | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$status" ]; then
            echo "   服务状态: $status"
        fi
    fi
}

# 测试就绪检查
test_readiness_check() {
    print_header "就绪检查"
    test_endpoint "GET" "/health/ready" "200" "就绪检查"
}

# 测试存活检查
test_liveness_check() {
    print_header "存活检查"
    test_endpoint "GET" "/health/live" "200" "存活检查"
}

# 测试统计信息
test_stats() {
    print_header "统计信息"
    test_endpoint "GET" "/health/stats" "200" "获取统计信息"
}

# 测试监控指标
test_metrics() {
    print_header "监控指标"
    test_endpoint "GET" "/metrics" "200" "获取监控指标"
    
    # 检查指标内容
    if [ -f "/tmp/response.json" ]; then
        local line_count=$(wc -l < /tmp/response.json)
        echo "   指标行数: $line_count"
    fi
}

# 测试模型列表API
test_models_api() {
    print_header "模型列表API"
    local auth_header="-H 'Authorization: Bearer $API_KEY'"
    test_endpoint "GET" "/v1/models" "200" "获取模型列表" "" "$auth_header"
    
    # 显示模型信息
    if [ -f "/tmp/response.json" ]; then
        local model_count=$(cat /tmp/response.json | grep -o '"id"' | wc -l)
        echo "   模型数量: $model_count"
        
        # 显示前几个模型名称
        local models=$(cat /tmp/response.json | grep -o '"id":"[^"]*"' | head -3 | cut -d'"' -f4)
        if [ -n "$models" ]; then
            echo "   模型列表:"
            echo "$models" | while read -r model; do
                echo "     - $model"
            done
        fi
    fi
}

# 测试聊天完成API
test_chat_completions() {
    print_header "聊天完成API"
    local auth_header="-H 'Authorization: Bearer $API_KEY'"
    local chat_data='{
        "model": "gpt-3.5-turbo",
        "messages": [
            {"role": "user", "content": "Hello! This is a test."}
        ],
        "max_tokens": 50
    }'
    
    test_endpoint "POST" "/v1/chat/completions" "200" "聊天完成请求" "$chat_data" "$auth_header"
    
    # 注意：如果没有配置真实的AI提供商API密钥，这个测试可能会失败
    # 但我们仍然可以检查请求是否被正确处理
}

# 测试管理API
test_admin_apis() {
    print_header "管理API"
    
    # 测试获取用户列表
    test_endpoint "GET" "/admin/users/" "200" "获取用户列表"
    
    # 显示用户信息
    if [ -f "/tmp/response.json" ]; then
        local user_count=$(cat /tmp/response.json | grep -o '"username"' | wc -l)
        echo "   用户数量: $user_count"
    fi
}

# 测试错误处理
test_error_handling() {
    print_header "错误处理"
    
    # 测试404
    test_endpoint "GET" "/nonexistent" "404" "404错误处理"
    
    # 测试无效API密钥
    local invalid_auth="-H 'Authorization: Bearer invalid_key'"
    test_endpoint "GET" "/v1/models" "401" "无效API密钥处理" "" "$invalid_auth"
}

# 性能测试
test_performance() {
    print_header "性能测试"
    
    echo "测试响应时间..."
    local start_time=$(date +%s%N)
    
    curl -s -o /dev/null "$BASE_URL/health"
    local curl_exit_code=$?
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # 转换为毫秒
    
    if [ $curl_exit_code -eq 0 ]; then
        print_result "true" "响应时间测试" "${duration}ms"
        
        if [ $duration -lt 100 ]; then
            echo "   性能: 优秀 (<100ms)"
        elif [ $duration -lt 500 ]; then
            echo "   性能: 良好 (<500ms)"
        else
            echo "   性能: 需要优化 (>500ms)"
        fi
    else
        print_result "false" "响应时间测试" "连接失败"
    fi
}

# 主函数
main() {
    echo -e "${BLUE}🚀 开始测试AI API Gateway服务${NC}"
    echo -e "${BLUE}📍 服务地址: $BASE_URL${NC}"
    echo -e "${BLUE}🔑 API密钥: ${API_KEY:0:10}...${NC}"
    echo ""
    
    # 运行所有测试
    test_health_check
    test_readiness_check
    test_liveness_check
    test_stats
    test_metrics
    test_models_api
    test_chat_completions
    test_admin_apis
    test_error_handling
    test_performance
    
    # 打印总结
    print_header "测试总结"
    echo -e "${BLUE}📊 测试结果: $PASSED_TESTS/$TOTAL_TESTS 通过${NC}"
    
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    
    if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
        echo -e "${GREEN}🎉 所有测试通过！服务运行正常。${NC}"
        exit 0
    elif [ $success_rate -ge 80 ]; then
        echo -e "${YELLOW}⚠️  大部分测试通过，但有些功能可能需要配置。${NC}"
        exit 0
    else
        echo -e "${RED}❌ 多个测试失败，请检查服务配置。${NC}"
        exit 1
    fi
}

# 清理函数
cleanup() {
    rm -f /tmp/response.json
}

# 设置清理
trap cleanup EXIT

# 检查curl是否可用
if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ curl命令未找到，请安装curl${NC}"
    exit 1
fi

# 运行主函数
main "$@"
