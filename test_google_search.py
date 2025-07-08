#!/usr/bin/env python3
"""
直接测试 Google Custom Search API
"""

import requests
import json

def test_google_search():
    """测试 Google Custom Search API"""
    
    # Google Custom Search API 配置
    api_key = "AIzaSyAJ-0mmqqaR610601edOxYw4MsS6GoavcY"
    search_engine_id = "05afc7eed6abd4a3c"
    query = "人工智能"
    # https://www.googleapis.com/customsearch/v1?key=INSERT_YOUR_API_KEY&cx=017576662512468239146:omuauf_lfve&q=lectures
    # 构建 API URL
    url = f"https://www.googleapis.com/customsearch/v1"
    params = {
        "key": api_key,
        "cx": search_engine_id,
        "q": query,
        "num": 3  # 返回3个结果
    }
    
    print(f"🔍 测试 Google 搜索: {query}")
    print(f"API URL: {url}")
    print(f"参数: {params}")
    
    try:
        response = requests.get(url, params=params, timeout=10)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print("✅ Google 搜索成功!")
            
            if "items" in data:
                print(f"找到 {len(data['items'])} 个结果:")
                for i, item in enumerate(data["items"], 1):
                    print(f"  {i}. {item['title']}")
                    print(f"     {item['link']}")
                    print(f"     {item.get('snippet', 'No snippet')[:100]}...")
                    print()
            else:
                print("没有找到搜索结果")
                
        else:
            print(f"❌ Google 搜索失败: {response.status_code}")
            print(f"错误信息: {response.text}")
            
    except Exception as e:
        print(f"❌ 请求异常: {e}")

if __name__ == "__main__":
    test_google_search()
