-- AI API Gateway Initial Schema
-- 注意：不使用外键约束，关系在应用层维护

-- 用户表
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, deleted
    balance DECIMAL(15,6) NOT NULL DEFAULT 0.000000,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- API密钥表
CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, expired, revoked
    permissions TEXT, -- JSON格式存储权限
    expires_at DATETIME,
    last_used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 服务提供商表
CREATE TABLE providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    base_url VARCHAR(500) NOT NULL,
    api_key_encrypted TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, inactive, maintenance
    priority INTEGER NOT NULL DEFAULT 1,
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    retry_attempts INTEGER NOT NULL DEFAULT 3,
    health_check_url VARCHAR(500),
    health_check_interval INTEGER NOT NULL DEFAULT 60,
    last_health_check DATETIME,
    health_status VARCHAR(20) DEFAULT 'unknown', -- healthy, unhealthy, unknown
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- AI模型表
CREATE TABLE models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    display_name VARCHAR(200),
    description TEXT,
    model_type VARCHAR(50) NOT NULL, -- chat, completion, embedding, etc.
    context_length INTEGER,
    max_tokens INTEGER,
    supports_streaming BOOLEAN NOT NULL DEFAULT false,
    supports_functions BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, deprecated, disabled
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, slug)
);

-- 模型定价表
CREATE TABLE model_pricing (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id INTEGER NOT NULL,
    pricing_type VARCHAR(20) NOT NULL, -- input, output, request
    price_per_unit DECIMAL(15,8) NOT NULL,
    unit VARCHAR(20) NOT NULL, -- token, request, character
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    effective_from DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    effective_until DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 配额设置表
CREATE TABLE quotas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    quota_type VARCHAR(20) NOT NULL, -- requests, tokens, cost
    period VARCHAR(20) NOT NULL, -- minute, hour, day, month
    limit_value DECIMAL(15,6) NOT NULL,
    reset_time VARCHAR(10), -- HH:MM for daily/monthly resets
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, quota_type, period)
);

-- 配额使用情况表
CREATE TABLE quota_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    quota_id INTEGER NOT NULL,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    used_value DECIMAL(15,6) NOT NULL DEFAULT 0.000000,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, quota_id, period_start)
);

-- 使用日志表
CREATE TABLE usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    api_key_id INTEGER NOT NULL,
    provider_id INTEGER NOT NULL,
    model_id INTEGER NOT NULL,
    request_id VARCHAR(100) NOT NULL UNIQUE,
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(200) NOT NULL,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    request_size INTEGER DEFAULT 0,
    response_size INTEGER DEFAULT 0,
    duration_ms INTEGER NOT NULL,
    status_code INTEGER NOT NULL,
    error_message TEXT,
    cost DECIMAL(15,8) DEFAULT 0.00000000,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 计费记录表
CREATE TABLE billing_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    usage_log_id INTEGER NOT NULL,
    amount DECIMAL(15,8) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_type VARCHAR(20) NOT NULL, -- usage, adjustment, refund
    description TEXT,
    processed_at DATETIME,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processed, failed
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
