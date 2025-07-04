"""
curl 'https://njhtjp-chat.302.ai/api/openai/v1/chat/completions' \
  -H 'accept: application/json, text/event-stream' \
  -H 'accept-language: zh-CN,zh;q=0.9' \
  -H 'authorization: Bearer sk-peuTZ2MkU5qZTmcKkT2MkUiHa9ZhWvwHPeRctcLkUVsG2Dv1' \
  -H 'content-type: application/json' \
  -b '_ga=GA1.1.364614762.1750662449; _pk_id.3.f081=d9cb09a539337ee6.1751271036.; _ss_s_uid=1e77a2c52f2ec876690b423abff5a90f; _pk_ref.3.f081=%5B%22%22%2C%22%22%2C1751523919%2C%22https%3A%2F%2Fdash.302.ai%2F%22%5D; _ga_V62NZ36GMW=GS2.1.s1751528657$o20$g0$t1751528657$j60$l0$h0; sl-session=VM8/RMWRZ2gbOQ0mg64ipA==' \
  -H 'lang: cn' \
  -H 'mcp-user-id: njhtjp-chat' \
  -H 'origin: https://njhtjp-chat.302.ai' \
  -H 'priority: u=1, i' \
  -H 'referer: https://njhtjp-chat.302.ai/' \
  -H 'sec-ch-ua: "Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36' \
  --data-raw $'{"messages":[{"role":"system","content":"Current model: qwen3-1.7b. Current time: Thu Jul 03 2025 17:47:01 GMT+0800 (中国标准时间). Latex inline: (\\\\x^2\\\\) Latex block: $$e=mc^2$$\\n"},{"role":"system","content":"这是历史聊天总结作为前情提要：看起来您只发送了一个下划线 `_`，这可能是：\\n\\n1. **误触发送** - 如果是意外发送，没关系！\\n2. **占位符** - 如果您在思考问题，我可以等您继续\\n3. **测试** - 如果在测试对话是否正常，一切正常！\\n\\n如果您有任何关于Go项目、编程或其他问题，随时可以提问。我刚才介绍了Go项目中`model`目录的作用，如果您还有相关疑问或想了解其他内容，请告诉我！"},{"role":"user","content":"价格呢，每个模型价格都不同"},{"role":"assistant","content":"你说得对，需要单独管理模型价格。我补充价格相关的表设计：\\n\\n## 价格管理表结构\\n\\n### 1. 模型价格表 (model_pricing)\\n```sql\\nCREATE TABLE model_pricing (\\n    id BIGINT PRIMARY KEY AUTO_INCREMENT,\\n    provider VARCHAR(50) NOT NULL, -- openai, anthropic, azure等\\n    model_name VARCHAR(100) NOT NULL, -- gpt-4, gpt-3.5-turbo等\\n    model_version VARCHAR(50) DEFAULT \'default\', -- 模型版本\\n    input_price DECIMAL(10,8) NOT NULL, -- 输入token价格(每1k tokens)\\n    output_price DECIMAL(10,8) NOT NULL, -- 输出token价格(每1k tokens)\\n    currency VARCHAR(10) DEFAULT \'USD\',\\n    effective_date DATE NOT NULL, -- 价格生效日期\\n    status TINYINT DEFAULT 1, -- 1:启用 0:禁用\\n    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\\n    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,\\n    UNIQUE KEY uk_provider_model_date (provider, model_name, model_version, effective_date),\\n    INDEX idx_model_effective (model_name, effective_date)\\n);\\n```\\n\\n### 2. 价格历史表 (pricing_history)\\n```sql\\nCREATE TABLE pricing_history (\\n    id BIGINT PRIMARY KEY AUTO_INCREMENT,\\n    model_pricing_id BIGINT NOT NULL,\\n    old_input_price DECIMAL(10,8),\\n    old_output_price DECIMAL(10,8),\\n    new_input_price DECIMAL(10,8),\\n    new_output_price DECIMAL(10,8),\\n    change_reason VARCHAR(255),\\n    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\\n    FOREIGN KEY (model_pricing_id) REFERENCES model_pricing(id)\\n);\\n```\\n\\n### 3. 用户自定义价格表 (user_custom_pricing)\\n```sql\\nCREATE TABLE user_custom_pricing (\\n    id BIGINT PRIMARY KEY AUTO_INCREMENT,\\n    user_id BIGINT NOT NULL,\\n    model_name VARCHAR(100) NOT NULL,\\n    markup_rate DECIMAL(5,4) DEFAULT 1.0000, -- 加价倍率，1.2表示加价20%\\n    fixed_markup DECIMAL(10,8) DEFAULT 0, -- 固定加价(每1k tokens)\\n    status TINYINT DEFAULT 1,\\n    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,\\n    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,\\n    UNIQUE KEY uk_user_model (user_id, model_name)\\n);\\n```\\n\\n## 更新使用记录表\\n```sql\\n-- 在usage_logs表中添加价格相关字段\\nALTER TABLE usage_logs ADD COLUMN input_price DECIMAL(10,8) DEFAULT 0; -- 输入token单价\\nALTER TABLE usage_logs ADD COLUMN output_price DECIMAL(10,8) DEFAULT 0; -- 输出token单价\\nALTER TABLE usage_logs ADD COLUMN input_cost DECIMAL(10,6) DEFAULT 0; -- 输入费用\\nALTER TABLE usage_logs ADD COLUMN output_cost DECIMAL(10,6) DEFAULT 0; -- 输出费用\\n-- total_cost = input_cost + output_cost\\n```\\n\\n## 价格管理逻辑\\n\\n### 1. 价格获取函数\\n```sql\\n-- 获取指定日期的模型价格\\nDELIMITER //\\nCREATE FUNCTION get_model_price(\\n    p_provider VARCHAR(50),\\n    p_model_name VARCHAR(100),\\n    p_date DATE,\\n    p_price_type ENUM(\'input\', \'output\')\\n) RETURNS DECIMAL(10,8)\\nREADS SQL DATA\\nBEGIN\\n    DECLARE price DECIMAL(10,8) DEFAULT 0;\\n    \\n    SELECT \\n        CASE \\n            WHEN p_price_type = \'input\' THEN input_price\\n            WHEN p_price_type = \'output\' THEN output_price\\n        END INTO price\\n    FROM model_pricing \\n    WHERE provider = p_provider \\n        AND model_name = p_model_name\\n        AND effective_date <= p_date\\n        AND status = 1\\n    ORDER BY effective_date DESC \\n    LIMIT 1;\\n    \\n    RETURN IFNULL(price, 0);\\nEND //\\nDELIMITER ;\\n```\\n\\n### 2. 费用计算示例数据\\n```sql\\n-- 插入常见模型价格 (2024年价格)\\nINSERT INTO model_pricing (provider, model_name, input_price, output_price, effective_date) VALUES\\n(\'openai\', \'gpt-4\', 0.03, 0.06, \'2024-01-01\'),\\n(\'openai\', \'gpt-4-turbo\', 0.01, 0.03, \'2024-01-01\'),\\n(\'openai\', \'gpt-3.5-turbo\', 0.0015, 0.002, \'2024-01-01\'),\\n(\'anthropic\', \'claude-3-opus\', 0.015, 0.075, \'2024-01-01\'),\\n(\'anthropic\', \'claude-3-sonnet\', 0.003, 0.015, \'2024-01-01\'),\\n(\'anthropic\', \'claude-3-haiku\', 0.00025, 0.00125, \'2024-01-01\');\\n```\\n\\n## 业务逻辑要点\\n\\n### 1. 费用计算流程\\n```python\\ndef calculate_cost(provider, model_name, input_tokens, output_tokens, user_id=None):\\n    # 获取基础价格\\n    base_input_price = get_model_price(provider, model_name, today(), \'input\')\\n    base_output_price = get_model_price(provider, model_name, today(), \'output\')\\n    \\n    # 获取用户自定义加价\\n    markup_rate = get_user_markup_rate(user_id, model_name)\\n    fixed_markup = get_user_fixed_markup(user_id, model_name)\\n    \\n    # 计算实际价格\\n    actual_input_price = base_input_price * markup_rate + fixed_markup\\n    actual_output_price = base_output_price * markup_rate + fixed_markup\\n    \\n    # 计算费用\\n    input_cost = (input_tokens / 1000) * actual_input_price\\n    output_cost = (output_tokens / 1000) * actual_output_price\\n    \\n    return input_cost + output_cost\\n```\\n\\n### 2. 价格管理功能\\n- **价格更新**：支持批量更新模型价格\\n- **价格历史**：记录价格变更历史\\n- **用户定价**：允许管理员为特定用户设置特殊价格\\n- **价格预警**：上游价格变动时自动通知\\n\\n### 3. 成本控制\\n- 实时计算请求成本\\n- 支持预付费和后付费模式\\n- 成本超限自动停止服务\\n- 详细的成本分析报表\\n\\n这样的设计可以灵活应对不同模型的价格差异，同时支持个性化定价策略。"},{"role":"user","content":"go项目的model目录是干嘛的"},{"role":"assistant","content":"在Go项目中，`model` 目录通常用来存放**数据模型**相关的代码，主要包括：\\n\\n## 1. 数据结构定义\\n```go\\n// model/user.go\\ntype User struct {\\n    ID        uint      `json:\\"id\\" gorm:\\"primaryKey\\"`\\n    Username  string    `json:\\"username\\" gorm:\\"uniqueIndex\\"`\\n    Email     string    `json:\\"email\\" gorm:\\"uniqueIndex\\"`\\n    Password  string    `json:\\"-\\" gorm:\\"column:password_hash\\"`\\n    Status    int       `json:\\"status\\" gorm:\\"default:1\\"`\\n    CreatedAt time.Time `json:\\"created_at\\"`\\n    UpdatedAt time.Time `json:\\"updated_at\\"`\\n}\\n```\\n\\n## 2. 数据库操作方法\\n```go\\n// model/user.go\\nfunc (u *User) Create(db *gorm.DB) error {\\n    return db.Create(u).Error\\n}\\n\\nfunc (u *User) GetByID(db *gorm.DB, id uint) error {\\n    return db.First(u, id).Error\\n}\\n\\nfunc (u *User) Update(db *gorm.DB) error {\\n    return db.Save(u).Error\\n}\\n```\\n\\n## 3. 业务逻辑方法\\n```go\\n// model/user.go\\nfunc (u *User) CheckPassword(password string) bool {\\n    return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil\\n}\\n\\nfunc (u *User) HashPassword(password string) error {\\n    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)\\n    if err \u0021= nil {\\n        return err\\n    }\\n    u.Password = string(hash)\\n    return nil\\n}\\n```\\n\\n## 4. 常见的目录结构\\n```\\nmodel/\\n├── user.go          # 用户模型\\n├── api_key.go       # API密钥模型\\n├── usage_log.go     # 使用记录模型\\n├── quota_config.go  # 额度配置模型\\n├── model.go         # 公共模型定义\\n└── init.go          # 数据库初始化\\n```\\n\\n## 5. 与其他层的关系\\n\\n### MVC架构中的位置\\n- **Model层**：处理数据逻辑\\n- **View层**：处理展示逻辑（模板/JSON响应）\\n- **Controller层**：处理业务逻辑和请求路由\\n\\n### 典型的调用流程\\n```\\nHTTP请求 → Controller → Model → Database\\n                ↓\\nHTTP响应 ← View ← Controller ← Model\\n```\\n\\n## 6. 实际示例\\n```go\\n// model/api_key.go\\ntype ApiKey struct {\\n    ID        uint      `json:\\"id\\" gorm:\\"primaryKey\\"`\\n    UserID    uint      `json:\\"user_id\\"`\\n    KeyName   string    `json:\\"key_name\\"`\\n    ApiKey    string    `json:\\"api_key\\" gorm:\\"uniqueIndex\\"`\\n    Status    int       `json:\\"status\\" gorm:\\"default:1\\"`\\n    ExpiresAt *time.Time `json:\\"expires_at\\"`\\n    CreatedAt time.Time `json:\\"created_at\\"`\\n    \\n    // 关联关系\\n    User User `json:\\"user\\" gorm:\\"foreignKey:UserID\\"`\\n}\\n\\n// 生成API密钥\\nfunc (ak *ApiKey) GenerateKey() {\\n    ak.ApiKey = \\"sk-\\" + generateRandomString(48)\\n}\\n\\n// 检查是否过期\\nfunc (ak *ApiKey) IsExpired() bool {\\n    if ak.ExpiresAt == nil {\\n        return false\\n    }\\n    return time.Now().After(*ak.ExpiresAt)\\n}\\n```\\n\\n## 主要作用\\n1. **数据抽象**：将数据库表结构抽象为Go结构体\\n2. **业务封装**：封装与数据相关的业务逻辑\\n3. **数据验证**：定义数据验证规则\\n4. **关系映射**：定义表之间的关联关系\\n5. **复用性**：提供可复用的数据操作方法\\n\\n简单来说，`model` 目录就是你的数据层，负责定义数据结构和数据相关的操作方法。"},{"role":"user","content":"你好"}],"model":"qwen3-1.7b","stream":true,"presence_penalty":0,"frequency_penalty":0,"top_p":1,"textract":true,"attachments":[],"temperature":0.5,"enable_thinking":true}'
"""

import requests

url = 'https://njhtjp-chat.302.ai/api/openai/v1/chat/completions'
headers = {
    'accept': 'application/json, text/event-stream',
    'accept-language': 'zh-CN,zh;q=0.9',
    'authorization': 'Bearer sk-4278fc65b1d32cc99fe69fc25bf352261fab3aa0b08488d919dce0097b0f3915',
    'content-type': 'application/json',
    'lang': 'cn',
    'mcp-user-id': 'njhtjp-chat',
    'origin': 'https://njhtjp-chat.302.ai',
    'priority': 'u=1, i',
    'referer': 'https://njhtjp-chat.302.ai/',
    'sec-ch-ua': '"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"',
    'sec-ch-ua-mobile': '?0',
    'sec-ch-ua-platform': '"Windows"',
    'sec-fetch-dest': 'empty',
    'sec-fetch-mode': 'cors',
    'sec-fetch-site': 'same-origin',
    'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36',
}
data = {
    "messages": [
        {"role": "user", "content": "你好"}
    ],
    "model": "qwen3-1.7b",
    "stream": True,
    "presence_penalty": 0,
    "frequency_penalty": 0,
    "top_p": 1,
    "textract": True,
    "attachments": [],
    "temperature": 0.5,
    "enable_thinking": True,
}
response = requests.post(url, headers=headers, json=data)
print(response.text)




