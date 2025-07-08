#!/usr/bin/env python3
"""
测试 DuckDuckGo 搜索功能
"""

import requests
import json

def test_duckduckgo_search():
    """测试 DuckDuckGo 搜索"""
    
    # DuckDuckGo 搜索 API
    url = "https://ddg.search2ai.online/search"
    
    data = {
        "q": "人工智能",
        "max_results": 3
    }
    
    print(f"🔍 测试 DuckDuckGo 搜索: {data['q']}")
    print(f"API URL: {url}")
    
    try:
        response = requests.post(url, json=data, timeout=15)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print("✅ DuckDuckGo 搜索成功!")
            
            if "results" in result and len(result["results"]) > 0:
                print(f"找到 {len(result['results'])} 个结果:")
                for i, item in enumerate(result["results"], 1):
                    print(f"  {i}. {item.get('title', 'No title')}")
                    print(f"     {item.get('href', item.get('url', 'No URL'))}")
                    print(f"     {item.get('body', 'No description')[:100]}...")
                    print()
            else:
                print("没有找到搜索结果")
                print(f"完整响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
                
        else:
            print(f"❌ DuckDuckGo 搜索失败: {response.status_code}")
            print(f"错误信息: {response.text}")
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")

if __name__ == "__main__":
    test_duckduckgo_search()
