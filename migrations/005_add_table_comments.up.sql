-- 为所有表和字段添加注释，提高数据库可读性和维护性
-- 注意：SQLite不支持COMMENT语法，这些注释主要用于文档目的
-- 实际的注释信息将通过应用程序代码和文档来维护

-- 由于SQLite不支持COMMENT ON语法，我们改为创建一个注释表来存储表和字段的说明
CREATE TABLE IF NOT EXISTS table_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    table_name VARCHAR(100) NOT NULL,
    column_name VARCHAR(100), -- NULL表示表级注释
    comment_text TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_name, column_name)
);

-- 插入表和字段注释
-- 用户表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('users', NULL, '用户表 - 存储API网关的用户信息'),
('users', 'id', '用户唯一标识符'),
('users', 'username', '用户名，用于登录和标识'),
('users', 'email', '用户邮箱地址，用于通知和找回密码'),
('users', 'password_hash', '密码哈希值，使用bcrypt加密'),
('users', 'full_name', '用户全名或显示名称'),
('users', 'status', '用户状态：active-活跃, inactive-非活跃, suspended-暂停'),
('users', 'balance', '用户账户余额（美元），用于API调用扣费'),
('users', 'created_at', '用户创建时间'),
('users', 'updated_at', '用户信息最后更新时间');

-- API密钥表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('api_keys', NULL, 'API密钥表 - 存储用户的API访问密钥'),
('api_keys', 'id', 'API密钥唯一标识符'),
('api_keys', 'user_id', '关联的用户ID'),
('api_keys', 'key', 'API密钥字符串，用于身份验证'),
('api_keys', 'key_prefix', 'API密钥前缀，用于快速识别'),
('api_keys', 'name', 'API密钥名称，便于用户管理'),
('api_keys', 'status', 'API密钥状态：active-活跃, inactive-非活跃, revoked-已撤销'),
('api_keys', 'last_used_at', 'API密钥最后使用时间'),
('api_keys', 'expires_at', 'API密钥过期时间，NULL表示永不过期'),
('api_keys', 'created_at', 'API密钥创建时间'),
('api_keys', 'updated_at', 'API密钥最后更新时间');

-- 提供商表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('providers', NULL, '上游AI服务提供商表 - 存储OpenAI、Anthropic等AI服务商信息'),
('providers', 'id', '提供商唯一标识符'),
('providers', 'name', '提供商名称，如OpenAI、Anthropic'),
('providers', 'slug', '提供商标识符，用于URL和配置'),
('providers', 'base_url', '提供商API基础URL'),
('providers', 'status', '提供商状态：active-活跃, inactive-非活跃, maintenance-维护中'),
('providers', 'health_status', '健康检查状态：healthy-健康, unhealthy-不健康, unknown-未知'),
('providers', 'priority', '提供商优先级，数字越小优先级越高'),
('providers', 'timeout_seconds', '请求超时时间（秒）'),
('providers', 'retry_attempts', '请求失败重试次数'),
('providers', 'health_check_interval', '健康检查间隔（秒）'),
('providers', 'created_at', '提供商创建时间'),
('providers', 'updated_at', '提供商信息最后更新时间');

-- 模型表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('models', NULL, 'AI模型表 - 存储各提供商支持的AI模型信息'),
('models', 'id', '模型唯一标识符'),
('models', 'provider_id', '关联的提供商ID'),
('models', 'name', '模型名称，如GPT-4、Claude-3'),
('models', 'slug', '模型标识符，用于API请求'),
('models', 'display_name', '模型显示名称，用于前端展示'),
('models', 'description', '模型描述信息'),
('models', 'model_type', '模型类型：chat-对话, completion-补全, embedding-嵌入, image-图像'),
('models', 'context_length', '模型上下文长度（token数）'),
('models', 'max_tokens', '模型最大输出token数'),
('models', 'supports_streaming', '是否支持流式响应'),
('models', 'supports_functions', '是否支持函数调用'),
('models', 'status', '模型状态：active-活跃, inactive-非活跃, deprecated-已弃用'),
('models', 'created_at', '模型创建时间'),
('models', 'updated_at', '模型信息最后更新时间');

-- 模型定价表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('model_pricing', NULL, '模型定价表 - 存储各模型的价格信息'),
('model_pricing', 'id', '定价记录唯一标识符'),
('model_pricing', 'model_id', '关联的模型ID'),
('model_pricing', 'pricing_type', '定价类型：input-输入token, output-输出token, request-请求次数'),
('model_pricing', 'price_per_unit', '单位价格（美元）'),
('model_pricing', 'unit', '计价单位：token-按token计费, request-按请求计费, character-按字符计费'),
('model_pricing', 'currency', '货币类型，默认USD'),
('model_pricing', 'effective_from', '价格生效开始时间'),
('model_pricing', 'effective_until', '价格生效结束时间，NULL表示永久有效'),
('model_pricing', 'created_at', '定价记录创建时间');

-- 提供商模型支持表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('provider_model_support', NULL, '提供商模型支持表 - 定义哪些提供商支持哪些模型（多对多关系）'),
('provider_model_support', 'id', '支持记录唯一标识符'),
('provider_model_support', 'provider_id', '提供商ID'),
('provider_model_support', 'model_slug', '模型标识符，用户请求时使用的模型名'),
('provider_model_support', 'upstream_model_name', '上游实际模型名，可能与model_slug不同'),
('provider_model_support', 'enabled', '是否启用此模型支持'),
('provider_model_support', 'priority', '该提供商对此模型的优先级，数字越小优先级越高'),
('provider_model_support', 'config', 'JSON格式的额外配置，如参数映射、自定义端点等'),
('provider_model_support', 'created_at', '支持记录创建时间'),
('provider_model_support', 'updated_at', '支持记录最后更新时间');



-- 配额表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('quotas', NULL, '用户配额表 - 定义用户的API使用限制'),
('quotas', 'id', '配额记录唯一标识符'),
('quotas', 'user_id', '关联的用户ID'),
('quotas', 'quota_type', '配额类型：daily-每日, monthly-每月, total-总计'),
('quotas', 'limit_value', '配额限制值'),
('quotas', 'limit_unit', '配额单位：requests-请求数, tokens-token数, cost-费用'),
('quotas', 'reset_period', '重置周期：daily-每日, monthly-每月, never-永不重置'),
('quotas', 'created_at', '配额创建时间'),
('quotas', 'updated_at', '配额最后更新时间');

-- 配额使用表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('quota_usage', NULL, '配额使用表 - 记录用户配额的实际使用情况'),
('quota_usage', 'id', '使用记录唯一标识符'),
('quota_usage', 'quota_id', '关联的配额ID'),
('quota_usage', 'user_id', '关联的用户ID'),
('quota_usage', 'used_value', '已使用的配额值'),
('quota_usage', 'period_start', '统计周期开始时间'),
('quota_usage', 'period_end', '统计周期结束时间'),
('quota_usage', 'created_at', '使用记录创建时间'),
('quota_usage', 'updated_at', '使用记录最后更新时间');

-- 使用日志表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('usage_logs', NULL, 'API使用日志表 - 记录每次API调用的详细信息'),
('usage_logs', 'id', '日志记录唯一标识符'),
('usage_logs', 'user_id', '调用用户ID'),
('usage_logs', 'api_key_id', '使用的API密钥ID'),
('usage_logs', 'provider_id', '实际使用的提供商ID'),
('usage_logs', 'model_id', '使用的模型ID'),
('usage_logs', 'request_id', '请求唯一标识符，用于追踪'),
('usage_logs', 'method', 'HTTP请求方法'),
('usage_logs', 'endpoint', '请求的API端点'),
('usage_logs', 'input_tokens', '输入token数量'),
('usage_logs', 'output_tokens', '输出token数量'),
('usage_logs', 'total_tokens', '总token数量'),
('usage_logs', 'request_size', '请求体大小（字节）'),
('usage_logs', 'response_size', '响应体大小（字节）'),
('usage_logs', 'duration_ms', '请求处理时间（毫秒）'),
('usage_logs', 'status_code', 'HTTP响应状态码'),
('usage_logs', 'error_message', '错误信息，成功时为空'),
('usage_logs', 'cost', '本次调用的费用（美元）'),
('usage_logs', 'created_at', '日志创建时间');

-- 计费记录表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('billing_records', NULL, '计费记录表 - 记录用户的扣费和充值记录'),
('billing_records', 'id', '计费记录唯一标识符'),
('billing_records', 'user_id', '关联的用户ID'),
('billing_records', 'usage_log_id', '关联的使用日志ID，充值时为NULL'),
('billing_records', 'amount', '金额，正数表示扣费，负数表示充值'),
('billing_records', 'currency', '货币类型'),
('billing_records', 'billing_type', '计费类型：usage-使用扣费, recharge-充值, refund-退款'),
('billing_records', 'description', '计费描述信息'),
('billing_records', 'status', '计费状态：pending-待处理, processed-已处理, failed-失败'),
('billing_records', 'processed_at', '处理时间'),
('billing_records', 'created_at', '计费记录创建时间');

-- 迁移记录表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('schema_migrations', NULL, '数据库迁移记录表 - 跟踪已执行的数据库迁移'),
('schema_migrations', 'version', '迁移版本号'),
('schema_migrations', 'applied_at', '迁移执行时间');

-- 注释表自身的注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('table_comments', NULL, '表注释表 - 存储数据库表和字段的说明信息'),
('table_comments', 'id', '注释记录唯一标识符'),
('table_comments', 'table_name', '表名'),
('table_comments', 'column_name', '字段名，NULL表示表级注释'),
('table_comments', 'comment_text', '注释内容'),
('table_comments', 'created_at', '注释创建时间');
