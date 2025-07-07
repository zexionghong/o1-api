-- ============================================================================
-- AI API Gateway PostgreSQL 初始化脚本
-- 版本: 完整版
-- 描述: 创建完整的数据库架构，包括所有表、索引、注释和初始数据
-- 注意: 不使用外键约束，关系在应用层维护（符合用户偏好）
-- ============================================================================

-- 设置客户端编码
SET client_encoding = 'UTF8';

-- ============================================================================
-- 用户管理表
-- ============================================================================

-- 用户表 - 存储API网关的用户信息
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255),
    password_hash VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    balance NUMERIC(15,6) NOT NULL DEFAULT 0.000000,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 认证授权表
-- ============================================================================

-- API密钥表 - 存储用户的API访问密钥
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    key VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    permissions TEXT,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 上游服务提供商表
-- ============================================================================

-- 服务提供商表 - 存储上游AI服务提供商信息
CREATE TABLE providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    base_url VARCHAR(500) NOT NULL,
    api_key_encrypted TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    priority INTEGER NOT NULL DEFAULT 1,
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    retry_attempts INTEGER NOT NULL DEFAULT 3,
    health_check_url VARCHAR(500),
    health_check_interval INTEGER NOT NULL DEFAULT 60,
    last_health_check TIMESTAMP,
    health_status VARCHAR(20) DEFAULT 'unknown',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- AI模型管理表
-- ============================================================================

-- AI模型表 - 存储各提供商支持的AI模型信息（不包含provider_id）
CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200),
    description TEXT,
    model_type VARCHAR(50) NOT NULL,
    context_length INTEGER,
    max_tokens INTEGER,
    supports_streaming BOOLEAN NOT NULL DEFAULT false,
    supports_functions BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 提供商模型支持表
-- ============================================================================

-- 提供商模型支持表 - 多对多关系，支持一个模型被多个提供商支持
CREATE TABLE provider_model_support (
    id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL,
    model_slug VARCHAR(100) NOT NULL,
    upstream_model_name VARCHAR(100),
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 1,
    config TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_provider_model_support UNIQUE(provider_id, model_slug)
);

-- ============================================================================
-- 计费定价表
-- ============================================================================

-- 模型定价表 - 存储各模型的计费价格信息
CREATE TABLE model_pricing (
    id SERIAL PRIMARY KEY,
    model_id INTEGER NOT NULL,
    pricing_type VARCHAR(20) NOT NULL,
    price_per_unit NUMERIC(15,8) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    multiplier NUMERIC(5,2) NOT NULL DEFAULT 1.5,
    effective_from TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    effective_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 配额管理表（基于API Key）
-- ============================================================================

-- 配额设置表 - 定义API Key的使用限制
CREATE TABLE quotas (
    id SERIAL PRIMARY KEY,
    api_key_id INTEGER NOT NULL,
    quota_type VARCHAR(20) NOT NULL,
    period VARCHAR(20),
    limit_value NUMERIC(15,6) NOT NULL,
    reset_time VARCHAR(10),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_quotas_api_key_type_period UNIQUE(api_key_id, quota_type, period)
);

-- 配额使用情况表 - 记录API Key在特定时间段内的配额使用情况
CREATE TABLE quota_usage (
    id SERIAL PRIMARY KEY,
    api_key_id INTEGER NOT NULL,
    quota_id INTEGER NOT NULL,
    period_start TIMESTAMP,
    period_end TIMESTAMP,
    used_value NUMERIC(15,6) NOT NULL DEFAULT 0.000000,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_quota_usage_api_key_quota_period UNIQUE(api_key_id, quota_id, period_start)
);

-- ============================================================================
-- 使用统计表
-- ============================================================================

-- 使用日志表 - 记录每次API调用的详细信息
CREATE TABLE usage_logs (
    id SERIAL PRIMARY KEY,
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
    cost NUMERIC(15,8) DEFAULT 0.00000000,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 计费管理表
-- ============================================================================

-- 计费记录表 - 记录用户的计费和扣费信息
CREATE TABLE billing_records (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    usage_log_id INTEGER NOT NULL,
    amount NUMERIC(15,8) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_type VARCHAR(20) NOT NULL,
    description TEXT,
    processed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 工具管理表
-- ============================================================================

-- 工具模板表 - 存储系统预定义的工具模板
CREATE TABLE tools (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    icon VARCHAR(100),
    color VARCHAR(20),
    config_schema JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 工具模型支持关联表
CREATE TABLE tool_model_support (
    id SERIAL PRIMARY KEY,
    tool_id VARCHAR(50) NOT NULL,
    model_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_tool_model_support UNIQUE(tool_id, model_id)
);

-- 用户工具实例表 - 存储用户创建的工具实例
CREATE TABLE user_tool_instances (
    id VARCHAR(36) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    tool_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    model_id INTEGER NOT NULL,
    api_key_id INTEGER NOT NULL,
    config JSONB,
    is_public BOOLEAN DEFAULT FALSE,
    share_token VARCHAR(32) UNIQUE,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 工具使用记录表 - 记录工具实例的使用情况
CREATE TABLE tool_usage_logs (
    id SERIAL PRIMARY KEY,
    tool_instance_id VARCHAR(36) NOT NULL,
    user_id INTEGER,
    session_id VARCHAR(64),
    request_count INTEGER DEFAULT 1,
    tokens_used INTEGER DEFAULT 0,
    cost NUMERIC(10, 6) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 创建索引以提高查询性能
-- ============================================================================

-- 用户表索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- API密钥表索引
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key ON api_keys(key);
CREATE INDEX idx_api_keys_status ON api_keys(status);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);

-- 提供商表索引
CREATE INDEX idx_providers_slug ON providers(slug);
CREATE INDEX idx_providers_status ON providers(status);
CREATE INDEX idx_providers_health_status ON providers(health_status);

-- 模型表索引
CREATE INDEX idx_models_slug ON models(slug);
CREATE INDEX idx_models_type ON models(model_type);
CREATE INDEX idx_models_status ON models(status);

-- 提供商模型支持表索引
CREATE INDEX idx_provider_model_support_provider_id ON provider_model_support(provider_id);
CREATE INDEX idx_provider_model_support_model_slug ON provider_model_support(model_slug);
CREATE INDEX idx_provider_model_support_enabled ON provider_model_support(enabled);

-- 模型定价表索引
CREATE INDEX idx_model_pricing_model_id ON model_pricing(model_id);
CREATE INDEX idx_model_pricing_type ON model_pricing(pricing_type);
CREATE INDEX idx_model_pricing_effective ON model_pricing(effective_from, effective_until);
CREATE INDEX idx_model_pricing_model_type ON model_pricing(model_id, pricing_type);
CREATE INDEX idx_model_pricing_multiplier ON model_pricing(model_id, pricing_type, multiplier);

-- 配额表索引
CREATE INDEX idx_quotas_api_key_id ON quotas(api_key_id);
CREATE INDEX idx_quotas_status ON quotas(status);
CREATE INDEX idx_quotas_type_period ON quotas(quota_type, period);

-- 配额使用表索引
CREATE INDEX idx_quota_usage_api_key_id ON quota_usage(api_key_id);
CREATE INDEX idx_quota_usage_quota_id ON quota_usage(quota_id);
CREATE INDEX idx_quota_usage_period ON quota_usage(period_start, period_end);

-- 使用日志表索引
CREATE INDEX idx_usage_logs_user_id ON usage_logs(user_id);
CREATE INDEX idx_usage_logs_api_key_id ON usage_logs(api_key_id);
CREATE INDEX idx_usage_logs_provider_id ON usage_logs(provider_id);
CREATE INDEX idx_usage_logs_model_id ON usage_logs(model_id);
CREATE INDEX idx_usage_logs_created_at ON usage_logs(created_at);
CREATE INDEX idx_usage_logs_request_id ON usage_logs(request_id);
CREATE INDEX idx_usage_logs_status_code ON usage_logs(status_code);

-- 计费记录表索引
CREATE INDEX idx_billing_records_user_id ON billing_records(user_id);
CREATE INDEX idx_billing_records_usage_log_id ON billing_records(usage_log_id);
CREATE INDEX idx_billing_records_status ON billing_records(status);
CREATE INDEX idx_billing_records_created_at ON billing_records(created_at);
CREATE INDEX idx_billing_records_billing_type ON billing_records(billing_type);

-- 工具表索引
CREATE INDEX idx_tools_category ON tools(category);
CREATE INDEX idx_tools_is_active ON tools(is_active);

-- 工具模型支持表索引
CREATE INDEX idx_tool_model_support_tool_id ON tool_model_support(tool_id);
CREATE INDEX idx_tool_model_support_model_id ON tool_model_support(model_id);

-- 用户工具实例表索引
CREATE INDEX idx_user_tool_instances_user_id ON user_tool_instances(user_id);
CREATE INDEX idx_user_tool_instances_tool_id ON user_tool_instances(tool_id);
CREATE INDEX idx_user_tool_instances_model_id ON user_tool_instances(model_id);
CREATE INDEX idx_user_tool_instances_api_key_id ON user_tool_instances(api_key_id);
CREATE INDEX idx_user_tool_instances_is_public ON user_tool_instances(is_public);
CREATE INDEX idx_user_tool_instances_share_token ON user_tool_instances(share_token);

-- 工具使用记录表索引
CREATE INDEX idx_tool_usage_logs_tool_instance_id ON tool_usage_logs(tool_instance_id);
CREATE INDEX idx_tool_usage_logs_user_id ON tool_usage_logs(user_id);
CREATE INDEX idx_tool_usage_logs_created_at ON tool_usage_logs(created_at);

-- ============================================================================
-- 表和列注释
-- ============================================================================

-- 用户表注释
COMMENT ON TABLE users IS 'API网关用户信息表';
COMMENT ON COLUMN users.id IS '用户唯一标识';
COMMENT ON COLUMN users.username IS '用户名，全局唯一';
COMMENT ON COLUMN users.email IS '邮箱地址，全局唯一';
COMMENT ON COLUMN users.full_name IS '用户全名（可选）';
COMMENT ON COLUMN users.password_hash IS '密码哈希值，用于用户名密码登录';
COMMENT ON COLUMN users.status IS '用户状态：active(活跃), suspended(暂停), deleted(已删除)';
COMMENT ON COLUMN users.balance IS '用户余额，支持6位小数精度，单位为美元';
COMMENT ON COLUMN users.created_at IS '账户创建时间';
COMMENT ON COLUMN users.updated_at IS '最后更新时间';

-- API密钥表注释
COMMENT ON TABLE api_keys IS 'API访问密钥表';
COMMENT ON COLUMN api_keys.id IS 'API密钥唯一标识';
COMMENT ON COLUMN api_keys.user_id IS '关联的用户ID（应用层维护关系）';
COMMENT ON COLUMN api_keys.key IS 'API密钥完整值';
COMMENT ON COLUMN api_keys.key_prefix IS '密钥前缀，用于快速识别（如：sk-xxx）';
COMMENT ON COLUMN api_keys.name IS '密钥名称，便于用户管理（可选）';
COMMENT ON COLUMN api_keys.status IS '密钥状态：active(活跃), suspended(暂停), expired(过期), revoked(撤销)';
COMMENT ON COLUMN api_keys.permissions IS '权限配置，JSON格式存储';
COMMENT ON COLUMN api_keys.expires_at IS '密钥过期时间（可选，NULL表示永不过期）';
COMMENT ON COLUMN api_keys.last_used_at IS '最后使用时间，用于统计和清理';
COMMENT ON COLUMN api_keys.created_at IS '密钥创建时间';
COMMENT ON COLUMN api_keys.updated_at IS '最后更新时间';

-- 提供商表注释
COMMENT ON TABLE providers IS '上游AI服务提供商信息表';
COMMENT ON COLUMN providers.id IS '提供商唯一标识';
COMMENT ON COLUMN providers.name IS '提供商名称，如"OpenAI"、"Anthropic"';
COMMENT ON COLUMN providers.slug IS '提供商标识符，如"openai"、"anthropic"';
COMMENT ON COLUMN providers.base_url IS '提供商API基础URL';
COMMENT ON COLUMN providers.api_key_encrypted IS '加密存储的API密钥（用于调用上游服务）';
COMMENT ON COLUMN providers.status IS '提供商状态：active(活跃), inactive(停用), maintenance(维护中)';
COMMENT ON COLUMN providers.priority IS '负载均衡优先级，数字越小优先级越高';
COMMENT ON COLUMN providers.timeout_seconds IS '请求超时时间（秒）';
COMMENT ON COLUMN providers.retry_attempts IS '失败重试次数';
COMMENT ON COLUMN providers.health_check_url IS '健康检查URL（可选）';
COMMENT ON COLUMN providers.health_check_interval IS '健康检查间隔（秒）';
COMMENT ON COLUMN providers.last_health_check IS '最后一次健康检查时间';
COMMENT ON COLUMN providers.health_status IS '健康状态：healthy(健康), unhealthy(不健康), unknown(未知)';
COMMENT ON COLUMN providers.created_at IS '提供商添加时间';
COMMENT ON COLUMN providers.updated_at IS '最后更新时间';

-- 模型表注释
COMMENT ON TABLE models IS 'AI模型信息表';
COMMENT ON COLUMN models.id IS '模型唯一标识';
COMMENT ON COLUMN models.name IS '模型完整名称，如"gpt-4-turbo-preview"';
COMMENT ON COLUMN models.slug IS '模型标识符，如"gpt-4-turbo"，用于API请求';
COMMENT ON COLUMN models.display_name IS '模型显示名称，如"GPT-4 Turbo"';
COMMENT ON COLUMN models.description IS '模型描述信息';
COMMENT ON COLUMN models.model_type IS '模型类型：chat(对话), completion(补全), embedding(嵌入), image(图像)等';
COMMENT ON COLUMN models.context_length IS '上下文长度限制（token数）';
COMMENT ON COLUMN models.max_tokens IS '单次响应最大token数';
COMMENT ON COLUMN models.supports_streaming IS '是否支持流式响应';
COMMENT ON COLUMN models.supports_functions IS '是否支持函数调用';
COMMENT ON COLUMN models.status IS '模型状态：active(活跃), deprecated(已弃用), disabled(已禁用)';
COMMENT ON COLUMN models.created_at IS '模型添加时间';
COMMENT ON COLUMN models.updated_at IS '最后更新时间';

-- 提供商模型支持表注释
COMMENT ON TABLE provider_model_support IS '提供商模型支持关联表，支持多对多关系';
COMMENT ON COLUMN provider_model_support.id IS '关联记录唯一标识';
COMMENT ON COLUMN provider_model_support.provider_id IS '提供商ID';
COMMENT ON COLUMN provider_model_support.model_slug IS '模型标识符';
COMMENT ON COLUMN provider_model_support.upstream_model_name IS '上游实际使用的模型名（可能与model_slug不同）';
COMMENT ON COLUMN provider_model_support.enabled IS '是否启用该提供商对此模型的支持';
COMMENT ON COLUMN provider_model_support.priority IS '该提供商对此模型的优先级（数字越小优先级越高）';
COMMENT ON COLUMN provider_model_support.config IS 'JSON格式的额外配置（如特殊参数映射等）';
COMMENT ON COLUMN provider_model_support.created_at IS '关联创建时间';
COMMENT ON COLUMN provider_model_support.updated_at IS '最后更新时间';

-- 模型定价表注释
COMMENT ON TABLE model_pricing IS '模型计费价格信息表';
COMMENT ON COLUMN model_pricing.id IS '定价记录唯一标识';
COMMENT ON COLUMN model_pricing.model_id IS '关联的模型ID（应用层维护关系）';
COMMENT ON COLUMN model_pricing.pricing_type IS '定价类型：input(输入token), output(输出token), request(请求次数)';
COMMENT ON COLUMN model_pricing.price_per_unit IS '单位价格，支持8位小数精度';
COMMENT ON COLUMN model_pricing.unit IS '计费单位：token(按token), request(按请求), character(按字符)';
COMMENT ON COLUMN model_pricing.currency IS '货币代码，如USD、CNY等';
COMMENT ON COLUMN model_pricing.multiplier IS '价格倍率，在原价基础上的倍数，默认1.5倍';
COMMENT ON COLUMN model_pricing.effective_from IS '价格生效开始时间';
COMMENT ON COLUMN model_pricing.effective_until IS '价格生效结束时间（NULL表示永久有效）';
COMMENT ON COLUMN model_pricing.created_at IS '定价记录创建时间';

-- 配额表注释
COMMENT ON TABLE quotas IS 'API Key配额设置表';
COMMENT ON COLUMN quotas.id IS '配额设置唯一标识';
COMMENT ON COLUMN quotas.api_key_id IS '关联的API密钥ID（应用层维护关系）';
COMMENT ON COLUMN quotas.quota_type IS '配额类型：requests(请求次数), tokens(token数量), cost(费用金额)';
COMMENT ON COLUMN quotas.period IS '配额周期：minute(分钟), hour(小时), day(天), month(月), NULL表示总限额';
COMMENT ON COLUMN quotas.limit_value IS '配额限制值';
COMMENT ON COLUMN quotas.reset_time IS '重置时间，格式HH:MM，用于日/月配额的具体重置时间点';
COMMENT ON COLUMN quotas.status IS '配额状态：active(生效), inactive(停用)';
COMMENT ON COLUMN quotas.created_at IS '配额创建时间';
COMMENT ON COLUMN quotas.updated_at IS '最后更新时间';

-- 配额使用表注释
COMMENT ON TABLE quota_usage IS 'API Key配额使用情况表';
COMMENT ON COLUMN quota_usage.id IS '使用记录唯一标识';
COMMENT ON COLUMN quota_usage.api_key_id IS '关联的API密钥ID（应用层维护关系）';
COMMENT ON COLUMN quota_usage.quota_id IS '关联的配额设置ID（应用层维护关系）';
COMMENT ON COLUMN quota_usage.period_start IS '统计周期开始时间（总限额时为NULL）';
COMMENT ON COLUMN quota_usage.period_end IS '统计周期结束时间（总限额时为NULL）';
COMMENT ON COLUMN quota_usage.used_value IS '已使用的配额值';
COMMENT ON COLUMN quota_usage.created_at IS '记录创建时间';
COMMENT ON COLUMN quota_usage.updated_at IS '最后更新时间';

-- 使用日志表注释
COMMENT ON TABLE usage_logs IS 'API调用使用日志表';
COMMENT ON COLUMN usage_logs.id IS '日志记录唯一标识';
COMMENT ON COLUMN usage_logs.user_id IS '调用用户ID（应用层维护关系）';
COMMENT ON COLUMN usage_logs.api_key_id IS '使用的API密钥ID（应用层维护关系）';
COMMENT ON COLUMN usage_logs.provider_id IS '实际调用的提供商ID（应用层维护关系）';
COMMENT ON COLUMN usage_logs.model_id IS '使用的模型ID（应用层维护关系）';
COMMENT ON COLUMN usage_logs.request_id IS '请求唯一标识符，用于追踪和去重';
COMMENT ON COLUMN usage_logs.method IS 'HTTP方法：GET, POST等';
COMMENT ON COLUMN usage_logs.endpoint IS '调用的API端点路径';
COMMENT ON COLUMN usage_logs.input_tokens IS '输入token数量';
COMMENT ON COLUMN usage_logs.output_tokens IS '输出token数量';
COMMENT ON COLUMN usage_logs.total_tokens IS '总token数量（input + output）';
COMMENT ON COLUMN usage_logs.request_size IS '请求体大小（字节）';
COMMENT ON COLUMN usage_logs.response_size IS '响应体大小（字节）';
COMMENT ON COLUMN usage_logs.duration_ms IS '请求处理时长（毫秒）';
COMMENT ON COLUMN usage_logs.status_code IS 'HTTP状态码';
COMMENT ON COLUMN usage_logs.error_message IS '错误信息（如果有）';
COMMENT ON COLUMN usage_logs.cost IS '本次调用的费用';
COMMENT ON COLUMN usage_logs.created_at IS '调用时间';

-- 计费记录表注释
COMMENT ON TABLE billing_records IS '用户计费记录表';
COMMENT ON COLUMN billing_records.id IS '计费记录唯一标识';
COMMENT ON COLUMN billing_records.user_id IS '关联的用户ID（应用层维护关系）';
COMMENT ON COLUMN billing_records.usage_log_id IS '关联的使用日志ID（应用层维护关系）';
COMMENT ON COLUMN billing_records.amount IS '计费金额，支持8位小数精度';
COMMENT ON COLUMN billing_records.currency IS '货币代码';
COMMENT ON COLUMN billing_records.billing_type IS '计费类型：usage(使用费), adjustment(调整), refund(退款)';
COMMENT ON COLUMN billing_records.description IS '计费描述信息';
COMMENT ON COLUMN billing_records.processed_at IS '计费处理时间';
COMMENT ON COLUMN billing_records.status IS '计费状态：pending(待处理), processed(已处理), failed(失败)';
COMMENT ON COLUMN billing_records.created_at IS '计费记录创建时间';

-- 工具表注释
COMMENT ON TABLE tools IS '系统预定义工具模板表';
COMMENT ON COLUMN tools.id IS '工具唯一标识';
COMMENT ON COLUMN tools.name IS '工具名称';
COMMENT ON COLUMN tools.description IS '工具描述';
COMMENT ON COLUMN tools.category IS '工具分类';
COMMENT ON COLUMN tools.icon IS '工具图标';
COMMENT ON COLUMN tools.color IS '工具颜色';
COMMENT ON COLUMN tools.config_schema IS '工具配置的JSON Schema';
COMMENT ON COLUMN tools.is_active IS '是否激活';
COMMENT ON COLUMN tools.created_at IS '创建时间';
COMMENT ON COLUMN tools.updated_at IS '更新时间';

-- 工具模型支持表注释
COMMENT ON TABLE tool_model_support IS '工具模型支持关联表';
COMMENT ON COLUMN tool_model_support.id IS '关联记录唯一标识';
COMMENT ON COLUMN tool_model_support.tool_id IS '工具ID';
COMMENT ON COLUMN tool_model_support.model_id IS '模型ID';
COMMENT ON COLUMN tool_model_support.created_at IS '关联创建时间';

-- 用户工具实例表注释
COMMENT ON TABLE user_tool_instances IS '用户创建的工具实例表';
COMMENT ON COLUMN user_tool_instances.id IS '工具实例唯一标识';
COMMENT ON COLUMN user_tool_instances.user_id IS '用户ID';
COMMENT ON COLUMN user_tool_instances.tool_id IS '关联工具模板ID';
COMMENT ON COLUMN user_tool_instances.name IS '用户自定义名称';
COMMENT ON COLUMN user_tool_instances.description IS '用户自定义描述';
COMMENT ON COLUMN user_tool_instances.model_id IS '用户选择的模型ID';
COMMENT ON COLUMN user_tool_instances.api_key_id IS '用户选择的API Key ID';
COMMENT ON COLUMN user_tool_instances.config IS '用户自定义配置';
COMMENT ON COLUMN user_tool_instances.is_public IS '是否公开分享';
COMMENT ON COLUMN user_tool_instances.share_token IS '分享token';
COMMENT ON COLUMN user_tool_instances.usage_count IS '使用次数';
COMMENT ON COLUMN user_tool_instances.created_at IS '创建时间';
COMMENT ON COLUMN user_tool_instances.updated_at IS '更新时间';

-- 工具使用记录表注释
COMMENT ON TABLE tool_usage_logs IS '工具实例使用记录表';
COMMENT ON COLUMN tool_usage_logs.id IS '使用记录唯一标识';
COMMENT ON COLUMN tool_usage_logs.tool_instance_id IS '关联用户工具实例ID';
COMMENT ON COLUMN tool_usage_logs.user_id IS '使用者ID（可能不是工具创建者）';
COMMENT ON COLUMN tool_usage_logs.session_id IS '会话ID';
COMMENT ON COLUMN tool_usage_logs.request_count IS '请求次数';
COMMENT ON COLUMN tool_usage_logs.tokens_used IS '使用的token数';
COMMENT ON COLUMN tool_usage_logs.cost IS '产生的费用';
COMMENT ON COLUMN tool_usage_logs.created_at IS '使用时间';

-- ============================================================================
-- 初始数据插入
-- ============================================================================

-- 插入提供商数据
INSERT INTO providers (id, name, slug, base_url, status, health_status, priority, timeout_seconds, retry_attempts, health_check_interval, created_at, updated_at)
VALUES
    (1, 'OpenAI', 'openai', 'https://api.openai.com/v1', 'active', 'healthy', 1, 30, 3, 60, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (2, 'Anthropic', 'anthropic', 'https://api.anthropic.com/v1', 'active', 'healthy', 2, 30, 3, 60, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 重置序列（PostgreSQL需要手动重置序列）
SELECT setval('providers_id_seq', (SELECT MAX(id) FROM providers));

-- 插入模型数据
INSERT INTO models (id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at)
VALUES
    -- OpenAI 模型
    (1, 'gpt-4', 'gpt-4', 'GPT-4', 'OpenAI GPT-4 模型', 'chat', 8192, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (2, 'gpt-4-32k', 'gpt-4-32k', 'GPT-4 32K', 'OpenAI GPT-4 32K 上下文模型', 'chat', 32768, 16384, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (3, 'gpt-4-turbo', 'gpt-4-turbo', 'GPT-4 Turbo', 'OpenAI GPT-4 Turbo 模型', 'chat', 128000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (4, 'gpt-4o', 'gpt-4o', 'GPT-4o', 'OpenAI GPT-4o 多模态模型', 'chat', 128000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (5, 'gpt-4o-mini', 'gpt-4o-mini', 'GPT-4o Mini', 'OpenAI GPT-4o Mini 轻量版', 'chat', 128000, 16384, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (6, 'gpt-3.5-turbo', 'gpt-3.5-turbo', 'GPT-3.5 Turbo', 'OpenAI GPT-3.5 Turbo 模型', 'chat', 16385, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (7, 'gpt-3.5-turbo-16k', 'gpt-3.5-turbo-16k', 'GPT-3.5 Turbo 16K', 'OpenAI GPT-3.5 Turbo 16K 上下文模型', 'chat', 16385, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (8, 'text-embedding-3-small', 'text-embedding-3-small', 'Text Embedding 3 Small', 'OpenAI 文本嵌入模型 Small', 'embedding', 8191, null, false, false, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (9, 'text-embedding-3-large', 'text-embedding-3-large', 'Text Embedding 3 Large', 'OpenAI 文本嵌入模型 Large', 'embedding', 8191, null, false, false, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (10, 'text-embedding-ada-002', 'text-embedding-ada-002', 'Text Embedding Ada 002', 'OpenAI 文本嵌入模型 Ada', 'embedding', 8191, null, false, false, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

    -- Anthropic 模型
    (11, 'claude-3-5-sonnet-20240620', 'claude-3-5-sonnet', 'Claude 3.5 Sonnet', 'Anthropic Claude 3.5 Sonnet 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (12, 'claude-3-opus-20240229', 'claude-3-opus', 'Claude 3 Opus', 'Anthropic Claude 3 Opus 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (13, 'claude-3-sonnet-20240229', 'claude-3-sonnet', 'Claude 3 Sonnet', 'Anthropic Claude 3 Sonnet 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (14, 'claude-3-haiku-20240307', 'claude-3-haiku', 'Claude 3 Haiku', 'Anthropic Claude 3 Haiku 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 重置序列
SELECT setval('models_id_seq', (SELECT MAX(id) FROM models));

-- 插入提供商模型支持关系
INSERT INTO provider_model_support (provider_id, model_slug, upstream_model_name, enabled, priority, created_at, updated_at)
VALUES
    -- OpenAI 提供商支持的模型
    (1, 'gpt-4', 'gpt-4', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'gpt-4-32k', 'gpt-4-32k', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'gpt-4-turbo', 'gpt-4-turbo-preview', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'gpt-4o', 'gpt-4o', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'gpt-4o-mini', 'gpt-4o-mini', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'gpt-3.5-turbo', 'gpt-3.5-turbo', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'gpt-3.5-turbo-16k', 'gpt-3.5-turbo-16k', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'text-embedding-3-small', 'text-embedding-3-small', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'text-embedding-3-large', 'text-embedding-3-large', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (1, 'text-embedding-ada-002', 'text-embedding-ada-002', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

    -- Anthropic 提供商支持的模型
    (2, 'claude-3-5-sonnet', 'claude-3-5-sonnet-20240620', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (2, 'claude-3-opus', 'claude-3-opus-20240229', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (2, 'claude-3-sonnet', 'claude-3-sonnet-20240229', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (2, 'claude-3-haiku', 'claude-3-haiku-20240307', true, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 插入模型定价数据（基于2024年7月的真实价格数据，包含1.5倍倍率）
INSERT INTO model_pricing (model_id, pricing_type, price_per_unit, unit, currency, multiplier, effective_from, effective_until, created_at)
VALUES
    -- GPT-4 定价
    (1, 'input', 0.03, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (1, 'output', 0.06, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- GPT-4 32K 定价
    (2, 'input', 0.06, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (2, 'output', 0.12, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- GPT-4 Turbo 定价
    (3, 'input', 0.01, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (3, 'output', 0.03, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- GPT-4o 定价
    (4, 'input', 0.005, 'token', 'USD', 1.5, '2024-05-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (4, 'output', 0.015, 'token', 'USD', 1.5, '2024-05-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- GPT-4o Mini 定价
    (5, 'input', 0.00015, 'token', 'USD', 1.5, '2024-07-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (5, 'output', 0.0006, 'token', 'USD', 1.5, '2024-07-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- GPT-3.5 Turbo 定价
    (6, 'input', 0.0005, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (6, 'output', 0.0015, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- GPT-3.5 Turbo 16K 定价
    (7, 'input', 0.001, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (7, 'output', 0.002, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- Text Embedding 3 Small 定价
    (8, 'input', 0.00002, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- Text Embedding 3 Large 定价
    (9, 'input', 0.00013, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- Text Embedding Ada 002 定价
    (10, 'input', 0.0001, 'token', 'USD', 1.5, '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- Claude 3.5 Sonnet 定价
    (11, 'input', 0.003, 'token', 'USD', 1.5, '2024-06-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (11, 'output', 0.015, 'token', 'USD', 1.5, '2024-06-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- Claude 3 Opus 定价
    (12, 'input', 0.015, 'token', 'USD', 1.5, '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (12, 'output', 0.075, 'token', 'USD', 1.5, '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- Claude 3 Sonnet 定价
    (13, 'input', 0.003, 'token', 'USD', 1.5, '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (13, 'output', 0.015, 'token', 'USD', 1.5, '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),

    -- Claude 3 Haiku 定价
    (14, 'input', 0.00025, 'token', 'USD', 1.5, '2024-03-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (14, 'output', 0.00125, 'token', 'USD', 1.5, '2024-03-01 00:00:00', NULL, CURRENT_TIMESTAMP);

-- 插入默认的工具模板
INSERT INTO tools (id, name, description, category, icon, color, config_schema, is_active, created_at, updated_at) VALUES
('chatbot', 'AI Chatbot', 'Create intelligent conversational AI', 'Communication', 'solar:chat-round-bold-duotone', '#45B7D1',
 '{"type": "object", "properties": {"system_prompt": {"type": "string", "default": "You are a helpful assistant."}, "temperature": {"type": "number", "minimum": 0, "maximum": 2, "default": 0.7}, "max_tokens": {"type": "integer", "minimum": 1, "maximum": 4000, "default": 2000}}}'::jsonb,
 true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

('image_generator', 'Image Generator', 'Generate images from text descriptions', 'Creative', 'solar:gallery-bold-duotone', '#4ECDC4',
 '{"type": "object", "properties": {"style": {"type": "string", "enum": ["natural", "vivid"], "default": "natural"}, "size": {"type": "string", "enum": ["1024x1024", "1792x1024", "1024x1792"], "default": "1024x1024"}, "quality": {"type": "string", "enum": ["standard", "hd"], "default": "standard"}}}'::jsonb,
 true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

('text_generator', 'Text Generator', 'Generate high-quality text content', 'Content', 'solar:document-text-bold-duotone', '#96CEB4',
 '{"type": "object", "properties": {"tone": {"type": "string", "enum": ["professional", "casual", "creative", "academic"], "default": "professional"}, "length": {"type": "string", "enum": ["short", "medium", "long"], "default": "medium"}, "format": {"type": "string", "enum": ["paragraph", "bullet_points", "numbered_list"], "default": "paragraph"}}}'::jsonb,
 true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),

('code_assistant', 'Code Assistant', 'AI-powered coding help and code generation', 'Development', 'solar:code-bold-duotone', '#FECA57',
 '{"type": "object", "properties": {"language": {"type": "string", "enum": ["javascript", "python", "java", "go", "rust", "typescript"], "default": "javascript"}, "style": {"type": "string", "enum": ["clean", "commented", "optimized"], "default": "clean"}, "include_tests": {"type": "boolean", "default": false}}}'::jsonb,
 true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 插入工具模型支持关系
INSERT INTO tool_model_support (tool_id, model_id, created_at)
SELECT 'chatbot', m.id, CURRENT_TIMESTAMP
FROM models m
WHERE m.slug IN ('gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet');

INSERT INTO tool_model_support (tool_id, model_id, created_at)
SELECT 'text_generator', m.id, CURRENT_TIMESTAMP
FROM models m
WHERE m.slug IN ('gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet');

INSERT INTO tool_model_support (tool_id, model_id, created_at)
SELECT 'code_assistant', m.id, CURRENT_TIMESTAMP
FROM models m
WHERE m.slug IN ('gpt-4o', 'claude-3-5-sonnet');

-- ============================================================================
-- 完成初始化
-- ============================================================================

-- 显示初始化完成信息
DO $$
BEGIN
    RAISE NOTICE '============================================================================';
    RAISE NOTICE 'AI API Gateway PostgreSQL 数据库初始化完成！';
    RAISE NOTICE '============================================================================';
    RAISE NOTICE '已创建的表：';
    RAISE NOTICE '- users: 用户管理';
    RAISE NOTICE '- api_keys: API密钥管理';
    RAISE NOTICE '- providers: 服务提供商';
    RAISE NOTICE '- models: AI模型';
    RAISE NOTICE '- provider_model_support: 提供商模型支持';
    RAISE NOTICE '- model_pricing: 模型定价';
    RAISE NOTICE '- quotas: 配额管理';
    RAISE NOTICE '- quota_usage: 配额使用';
    RAISE NOTICE '- usage_logs: 使用日志';
    RAISE NOTICE '- billing_records: 计费记录';
    RAISE NOTICE '- tools: 工具模板';
    RAISE NOTICE '- tool_model_support: 工具模型支持';
    RAISE NOTICE '- user_tool_instances: 用户工具实例';
    RAISE NOTICE '- tool_usage_logs: 工具使用记录';
    RAISE NOTICE '============================================================================';
    RAISE NOTICE '初始数据：';
    RAISE NOTICE '- 2个服务提供商 (OpenAI, Anthropic)';
    RAISE NOTICE '- 14个AI模型';
    RAISE NOTICE '- 完整的定价数据 (包含1.5倍倍率)';
    RAISE NOTICE '- 4个工具模板';
    RAISE NOTICE '============================================================================';
    RAISE NOTICE '数据库已准备就绪，可以启动AI API Gateway服务！';
    RAISE NOTICE '============================================================================';
END $$;
