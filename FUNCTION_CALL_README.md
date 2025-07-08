# Function Call 功能说明

本项目已成功集成了类似 search2ai 的 Function Call 功能，支持搜索、新闻和网页爬取等功能。

## 功能特性

### 支持的工具函数

1. **search** - 网络搜索
   - 描述：在互联网上搜索信息
   - 参数：query (string) - 搜索查询

2. **news** - 新闻搜索
   - 描述：搜索新闻文章
   - 参数：query (string) - 新闻搜索查询

3. **crawler** - 网页爬取
   - 描述：获取指定 URL 的网页内容
   - 参数：url (string) - 要爬取的网页 URL

### 支持的搜索服务

- **search1api** - Search1API 服务（需要 API 密钥）
- **google** - Google Custom Search API（需要 API 密钥和搜索引擎 ID）
- **bing** - Bing Search API（需要 API 密钥）
- **serpapi** - SerpAPI（需要 API 密钥）
- **serper** - Serper API（需要 API 密钥）
- **duckduckgo** - DuckDuckGo 搜索（免费，默认选项）
- **searxng** - SearXNG 自建实例（需要配置服务器地址）

## 配置说明

### 1. 启用 Function Call

在 `configs/config.yaml` 中启用 Function Call 功能：

```yaml
function_call:
  enabled: true  # 设置为 true 启用功能
  
  search_service:
    service: "duckduckgo"  # 选择搜索服务
    max_results: 10        # 最大搜索结果数
    crawl_results: 0       # 深度搜索数量
```

### 2. 配置搜索服务

根据选择的搜索服务，配置相应的 API 密钥：

```yaml
function_call:
  search_service:
    # 使用 Google 搜索
    service: "google"
    google_cx: "your_google_custom_search_engine_id"
    google_key: "your_google_api_key"
    
    # 或使用 Bing 搜索
    service: "bing"
    bing_key: "your_bing_search_api_key"
    
    # 或使用免费的 DuckDuckGo（推荐用于测试）
    service: "duckduckgo"
```

## 使用方法

### 1. 手动指定工具

在请求中明确指定要使用的工具：

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "请搜索最新的人工智能发展趋势"
    }
  ],
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
```

### 2. 自动工具调用

系统会自动检测用户消息中的关键词，判断是否需要使用搜索功能：

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "今天的天气怎么样？请搜索一下北京的天气情况。"
    }
  ]
}
```

触发自动搜索的关键词包括：
- 搜索、查找、search、find、lookup
- 新闻、news、最新、latest、recent
- 网页、网站、url、webpage、website
- 什么是、what is、how to、怎么
- 今天、today、现在、now、当前、current

## 测试方法

### 1. 运行测试脚本

```bash
python test_function_call.py
```

### 2. 手动测试

使用 curl 命令测试：

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "请搜索一下最新的科技新闻"
      }
    ],
    "max_tokens": 1000
  }'
```

## 工作流程

1. **请求接收**：用户发送聊天请求
2. **工具检测**：系统检测是否需要使用工具或用户明确指定了工具
3. **第一次 AI 调用**：发送请求到 AI 模型，模型决定是否调用工具
4. **工具执行**：如果模型决定调用工具，系统执行相应的搜索/爬取操作
5. **第二次 AI 调用**：将工具执行结果添加到对话历史，再次调用 AI 模型生成最终回复
6. **返回结果**：返回包含搜索结果的最终回复

## 注意事项

1. **API 密钥**：使用付费搜索服务需要配置相应的 API 密钥
2. **速率限制**：注意各搜索服务的速率限制
3. **成本控制**：Function Call 会产生额外的 AI 模型调用成本
4. **超时设置**：搜索和爬取操作可能需要较长时间，建议设置合适的超时时间
5. **错误处理**：搜索失败时系统会返回错误信息，不会中断整个对话流程

## 扩展开发

要添加新的工具函数：

1. 在 `internal/infrastructure/functioncall/function_call_handler.go` 中添加新的工具定义
2. 在 `executeFunction` 方法中添加新的函数执行逻辑
3. 实现具体的功能逻辑
4. 更新配置文件和文档

## 故障排查

1. **Function Call 不工作**：检查配置文件中 `function_call.enabled` 是否为 `true`
2. **搜索失败**：检查搜索服务配置和 API 密钥
3. **超时错误**：增加请求超时时间或选择更快的搜索服务
4. **权限错误**：确保 API 密钥有效且有足够的配额
