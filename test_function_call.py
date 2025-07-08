#!/usr/bin/env python3
"""
Function Call 功能测试脚本

测试 AI API Gateway 的 Function Call 功能，包括搜索、新闻和网页爬取。
"""

import json
import requests
import time
from typing import Dict, Any

# 配置
API_BASE_URL = "http://localhost:8080"
API_KEY = "test-api-key"  # 需要替换为实际的 API 密钥

def test_function_call_search():
    """测试搜索功能"""
    print("🔍 测试搜索功能...")
    
    url = f"{API_BASE_URL}/v1/chat/completions"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "user",
                "content": "请搜索一下最新的人工智能发展趋势"
            }
        ],
        "max_tokens": 1000,
        "temperature": 0.7,
        "tools": [
            {
                "type": "function",
                "function": {
                    "name": "search",
                    "description": "Search for information on the internet",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "query": {
                                "type": "string",
                                "description": "The search query to execute"
                            }
                        },
                        "required": ["query"]
                    }
                }
            }
        ],
        "tool_choice": "auto"
    }
    
    try:
        response = requests.post(url, headers=headers, json=data, timeout=60)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print("✅ 搜索功能测试成功")
            print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
        else:
            print(f"❌ 搜索功能测试失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 搜索功能测试异常: {e}")

def test_function_call_news():
    """测试新闻搜索功能"""
    print("\n📰 测试新闻搜索功能...")
    
    url = f"{API_BASE_URL}/v1/chat/completions"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "user",
                "content": "请搜索今天的科技新闻"
            }
        ],
        "max_tokens": 1000,
        "temperature": 0.7,
        "tools": [
            {
                "type": "function",
                "function": {
                    "name": "news",
                    "description": "Search for news articles",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "query": {
                                "type": "string",
                                "description": "The news search query to execute"
                            }
                        },
                        "required": ["query"]
                    }
                }
            }
        ],
        "tool_choice": "auto"
    }
    
    try:
        response = requests.post(url, headers=headers, json=data, timeout=60)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print("✅ 新闻搜索功能测试成功")
            print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
        else:
            print(f"❌ 新闻搜索功能测试失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 新闻搜索功能测试异常: {e}")

def test_function_call_crawler():
    """测试网页爬取功能"""
    print("\n🕷️ 测试网页爬取功能...")
    
    url = f"{API_BASE_URL}/v1/chat/completions"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "user",
                "content": "请帮我获取这个网页的内容：https://www.example.com"
            }
        ],
        "max_tokens": 1000,
        "temperature": 0.7,
        "tools": [
            {
                "type": "function",
                "function": {
                    "name": "crawler",
                    "description": "Get the content of a specified URL",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "url": {
                                "type": "string",
                                "description": "The URL of the webpage to crawl"
                            }
                        },
                        "required": ["url"]
                    }
                }
            }
        ],
        "tool_choice": "auto"
    }
    
    try:
        response = requests.post(url, headers=headers, json=data, timeout=60)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print("✅ 网页爬取功能测试成功")
            print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
        else:
            print(f"❌ 网页爬取功能测试失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 网页爬取功能测试异常: {e}")

def test_auto_function_call():
    """测试自动 Function Call 功能"""
    print("\n🤖 测试自动 Function Call 功能...")
    
    url = f"{API_BASE_URL}/v1/chat/completions"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # 不提供 tools，让系统自动判断是否需要使用 Function Call
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {
                "role": "user",
                "content": "今天的天气怎么样？请搜索一下北京的天气情况。"
            }
        ],
        "max_tokens": 1000,
        "temperature": 0.7
    }
    
    try:
        response = requests.post(url, headers=headers, json=data, timeout=60)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print("✅ 自动 Function Call 功能测试成功")
            print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
        else:
            print(f"❌ 自动 Function Call 功能测试失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 自动 Function Call 功能测试异常: {e}")

def test_health_check():
    """测试健康检查"""
    print("\n❤️ 测试健康检查...")
    
    url = f"{API_BASE_URL}/health"
    
    try:
        response = requests.get(url, timeout=10)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print("✅ 健康检查成功")
            print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
        else:
            print(f"❌ 健康检查失败: {response.text}")
            
    except Exception as e:
        print(f"❌ 健康检查异常: {e}")

def main():
    """主函数"""
    print("🚀 开始测试 Function Call 功能")
    print("=" * 50)
    
    # 首先测试健康检查
    test_health_check()
    
    # 测试各种 Function Call 功能
    test_function_call_search()
    test_function_call_news()
    test_function_call_crawler()
    test_auto_function_call()
    
    print("\n" + "=" * 50)
    print("🎉 Function Call 功能测试完成")

if __name__ == "__main__":
    main()
