-- ============================================================================
-- AI API Gateway 初始数据库架构
-- 版本: 001
-- 描述: 创建核心业务表，包括用户、提供商、模型、API密钥、配额、使用日志、计费记录等
-- 注意: 不使用外键约束，关系在应用层维护（符合用户偏好）
-- ============================================================================

-- ============================================================================
-- 用户管理表
-- ============================================================================

-- 用户表 - 存储API网关的用户信息
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 用户唯一标识
    username VARCHAR(100) NOT NULL UNIQUE,                   -- 用户名，全局唯一
    email VARCHAR(255) NOT NULL UNIQUE,                      -- 邮箱地址，全局唯一
    full_name VARCHAR(255),                                  -- 用户全名（可选）
    status VARCHAR(20) NOT NULL DEFAULT 'active',           -- 用户状态：active(活跃), suspended(暂停), deleted(已删除)
    balance DECIMAL(15,6) NOT NULL DEFAULT 0.000000,        -- 用户余额，支持6位小数精度，单位为美元
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 账户创建时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 最后更新时间
);

-- ============================================================================
-- 认证授权表
-- ============================================================================

-- API密钥表 - 存储用户的API访问密钥
CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- API密钥唯一标识
    user_id INTEGER NOT NULL,                                -- 关联的用户ID（应用层维护关系）
    key_hash VARCHAR(255) NOT NULL UNIQUE,                   -- API密钥的哈希值，用于验证
    key_prefix VARCHAR(20) NOT NULL,                         -- 密钥前缀，用于快速识别（如：sk-xxx）
    name VARCHAR(100),                                       -- 密钥名称，便于用户管理（可选）
    status VARCHAR(20) NOT NULL DEFAULT 'active',           -- 密钥状态：active(活跃), suspended(暂停), expired(过期), revoked(撤销)
    permissions TEXT,                                        -- 权限配置，JSON格式存储（如：模型访问权限、速率限制等）
    expires_at DATETIME,                                     -- 密钥过期时间（可选，NULL表示永不过期）
    last_used_at DATETIME,                                   -- 最后使用时间，用于统计和清理
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 密钥创建时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 最后更新时间
);

-- ============================================================================
-- 上游服务提供商表
-- ============================================================================

-- 服务提供商表 - 存储上游AI服务提供商信息（如OpenAI、Anthropic等）
CREATE TABLE providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 提供商唯一标识
    name VARCHAR(100) NOT NULL UNIQUE,                       -- 提供商名称，如"OpenAI"、"Anthropic"
    slug VARCHAR(50) NOT NULL UNIQUE,                        -- 提供商标识符，如"openai"、"anthropic"
    base_url VARCHAR(500) NOT NULL,                          -- 提供商API基础URL
    api_key_encrypted TEXT,                                  -- 加密存储的API密钥（用于调用上游服务）
    status VARCHAR(20) NOT NULL DEFAULT 'active',           -- 提供商状态：active(活跃), inactive(停用), maintenance(维护中)
    priority INTEGER NOT NULL DEFAULT 1,                    -- 负载均衡优先级，数字越小优先级越高
    timeout_seconds INTEGER NOT NULL DEFAULT 30,            -- 请求超时时间（秒）
    retry_attempts INTEGER NOT NULL DEFAULT 3,              -- 失败重试次数
    health_check_url VARCHAR(500),                          -- 健康检查URL（可选）
    health_check_interval INTEGER NOT NULL DEFAULT 60,      -- 健康检查间隔（秒）
    last_health_check DATETIME,                             -- 最后一次健康检查时间
    health_status VARCHAR(20) DEFAULT 'unknown',            -- 健康状态：healthy(健康), unhealthy(不健康), unknown(未知)
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 提供商添加时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 最后更新时间
);

-- ============================================================================
-- AI模型管理表
-- ============================================================================

-- AI模型表 - 存储各提供商支持的AI模型信息
CREATE TABLE models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 模型唯一标识
    provider_id INTEGER NOT NULL,                            -- 所属提供商ID（应用层维护关系）
    name VARCHAR(100) NOT NULL,                              -- 模型完整名称，如"gpt-4-turbo-preview"
    slug VARCHAR(100) NOT NULL,                              -- 模型标识符，如"gpt-4-turbo"，用于API请求
    display_name VARCHAR(200),                               -- 模型显示名称，如"GPT-4 Turbo"
    description TEXT,                                        -- 模型描述信息
    model_type VARCHAR(50) NOT NULL,                        -- 模型类型：chat(对话), completion(补全), embedding(嵌入), image(图像)等
    context_length INTEGER,                                  -- 上下文长度限制（token数）
    max_tokens INTEGER,                                      -- 单次响应最大token数
    supports_streaming BOOLEAN NOT NULL DEFAULT false,      -- 是否支持流式响应
    supports_functions BOOLEAN NOT NULL DEFAULT false,      -- 是否支持函数调用
    status VARCHAR(20) NOT NULL DEFAULT 'active',           -- 模型状态：active(活跃), deprecated(已弃用), disabled(已禁用)
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 模型添加时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间
    UNIQUE(provider_id, slug)                               -- 同一提供商内模型标识符唯一
);

-- ============================================================================
-- 计费定价表
-- ============================================================================

-- 模型定价表 - 存储各模型的计费价格信息
CREATE TABLE model_pricing (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 定价记录唯一标识
    model_id INTEGER NOT NULL,                               -- 关联的模型ID（应用层维护关系）
    pricing_type VARCHAR(20) NOT NULL,                      -- 定价类型：input(输入token), output(输出token), request(请求次数)
    price_per_unit DECIMAL(15,8) NOT NULL,                  -- 单位价格，支持8位小数精度
    unit VARCHAR(20) NOT NULL,                              -- 计费单位：token(按token), request(按请求), character(按字符)
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',             -- 货币代码，如USD、CNY等
    effective_from DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 价格生效开始时间
    effective_until DATETIME,                               -- 价格生效结束时间（NULL表示永久有效）
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 定价记录创建时间
);

-- ============================================================================
-- 配额管理表
-- ============================================================================

-- 配额设置表 - 定义用户的使用限制
CREATE TABLE quotas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 配额设置唯一标识
    user_id INTEGER NOT NULL,                                -- 关联的用户ID（应用层维护关系）
    quota_type VARCHAR(20) NOT NULL,                        -- 配额类型：requests(请求次数), tokens(token数量), cost(费用金额)
    period VARCHAR(20) NOT NULL,                            -- 配额周期：minute(分钟), hour(小时), day(天), month(月)
    limit_value DECIMAL(15,6) NOT NULL,                     -- 配额限制值
    reset_time VARCHAR(10),                                 -- 重置时间，格式HH:MM，用于日/月配额的具体重置时间点
    status VARCHAR(20) NOT NULL DEFAULT 'active',           -- 配额状态：active(生效), inactive(停用)
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 配额创建时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间
    UNIQUE(user_id, quota_type, period)                     -- 同一用户的同类型同周期配额唯一
);

-- 配额使用情况表 - 记录用户在特定时间段内的配额使用情况
CREATE TABLE quota_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 使用记录唯一标识
    user_id INTEGER NOT NULL,                                -- 关联的用户ID（应用层维护关系）
    quota_id INTEGER NOT NULL,                               -- 关联的配额设置ID（应用层维护关系）
    period_start DATETIME NOT NULL,                         -- 统计周期开始时间
    period_end DATETIME NOT NULL,                           -- 统计周期结束时间
    used_value DECIMAL(15,6) NOT NULL DEFAULT 0.000000,     -- 已使用的配额值
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 记录创建时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间
    UNIQUE(user_id, quota_id, period_start)                 -- 同一用户同一配额同一周期的记录唯一
);

-- ============================================================================
-- 使用统计表
-- ============================================================================

-- 使用日志表 - 记录每次API调用的详细信息
CREATE TABLE usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 日志记录唯一标识
    user_id INTEGER NOT NULL,                                -- 调用用户ID（应用层维护关系）
    api_key_id INTEGER NOT NULL,                             -- 使用的API密钥ID（应用层维护关系）
    provider_id INTEGER NOT NULL,                            -- 实际调用的提供商ID（应用层维护关系）
    model_id INTEGER NOT NULL,                               -- 使用的模型ID（应用层维护关系）
    request_id VARCHAR(100) NOT NULL UNIQUE,                 -- 请求唯一标识符，用于追踪和去重
    method VARCHAR(10) NOT NULL,                             -- HTTP方法：GET, POST等
    endpoint VARCHAR(200) NOT NULL,                          -- 调用的API端点路径
    input_tokens INTEGER DEFAULT 0,                          -- 输入token数量
    output_tokens INTEGER DEFAULT 0,                         -- 输出token数量
    total_tokens INTEGER DEFAULT 0,                          -- 总token数量（input + output）
    request_size INTEGER DEFAULT 0,                          -- 请求体大小（字节）
    response_size INTEGER DEFAULT 0,                         -- 响应体大小（字节）
    duration_ms INTEGER NOT NULL,                            -- 请求处理时长（毫秒）
    status_code INTEGER NOT NULL,                            -- HTTP状态码
    error_message TEXT,                                      -- 错误信息（如果有）
    cost DECIMAL(15,8) DEFAULT 0.00000000,                  -- 本次调用的费用
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 调用时间
);

-- ============================================================================
-- 计费管理表
-- ============================================================================

-- 计费记录表 - 记录用户的计费和扣费信息
CREATE TABLE billing_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 计费记录唯一标识
    user_id INTEGER NOT NULL,                                -- 关联的用户ID（应用层维护关系）
    usage_log_id INTEGER NOT NULL,                           -- 关联的使用日志ID（应用层维护关系）
    amount DECIMAL(15,8) NOT NULL,                           -- 计费金额，支持8位小数精度
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',             -- 货币代码
    billing_type VARCHAR(20) NOT NULL,                      -- 计费类型：usage(使用费), adjustment(调整), refund(退款)
    description TEXT,                                        -- 计费描述信息
    processed_at DATETIME,                                   -- 计费处理时间
    status VARCHAR(20) NOT NULL DEFAULT 'pending',          -- 计费状态：pending(待处理), processed(已处理), failed(失败)
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 计费记录创建时间
);

-- 创建索引以提高查询性能
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_status ON api_keys(status);

CREATE INDEX idx_models_provider_id ON models(provider_id);
CREATE INDEX idx_models_status ON models(status);

CREATE INDEX idx_model_pricing_model_id ON model_pricing(model_id);
CREATE INDEX idx_model_pricing_effective ON model_pricing(effective_from, effective_until);

CREATE INDEX idx_quotas_user_id ON quotas(user_id);
CREATE INDEX idx_quota_usage_user_id ON quota_usage(user_id);
CREATE INDEX idx_quota_usage_period ON quota_usage(period_start, period_end);

CREATE INDEX idx_usage_logs_user_id ON usage_logs(user_id);
CREATE INDEX idx_usage_logs_api_key_id ON usage_logs(api_key_id);
CREATE INDEX idx_usage_logs_created_at ON usage_logs(created_at);
CREATE INDEX idx_usage_logs_request_id ON usage_logs(request_id);

CREATE INDEX idx_billing_records_user_id ON billing_records(user_id);
CREATE INDEX idx_billing_records_status ON billing_records(status);
CREATE INDEX idx_billing_records_created_at ON billing_records(created_at);
