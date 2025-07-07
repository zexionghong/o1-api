-- 插入模型价格数据
-- 基于2024年7月的真实价格数据

-- 首先插入提供商数据（如果不存在）
INSERT OR IGNORE INTO providers (id, name, slug, base_url, status, health_status, priority, timeout_seconds, retry_attempts, health_check_interval, created_at, updated_at)
VALUES
    (1, 'OpenAI', 'openai', 'https://api.openai.com/v1', 'active', 'healthy', 1, 30, 3, 60, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (2, 'Anthropic', 'anthropic', 'https://api.anthropic.com/v1', 'active', 'healthy', 2, 30, 3, 60, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 插入模型数据（如果不存在）
INSERT OR IGNORE INTO models (id, provider_id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at)
VALUES 
    -- OpenAI 模型
    (1, 1, 'gpt-4', 'gpt-4', 'GPT-4', 'OpenAI GPT-4 模型', 'chat', 8192, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (2, 1, 'gpt-4-32k', 'gpt-4-32k', 'GPT-4 32K', 'OpenAI GPT-4 32K 上下文模型', 'chat', 32768, 16384, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (3, 1, 'gpt-4-turbo', 'gpt-4-turbo', 'GPT-4 Turbo', 'OpenAI GPT-4 Turbo 模型', 'chat', 128000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (4, 1, 'gpt-4o', 'gpt-4o', 'GPT-4o', 'OpenAI GPT-4o 多模态模型', 'chat', 128000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (5, 1, 'gpt-4o-mini', 'gpt-4o-mini', 'GPT-4o Mini', 'OpenAI GPT-4o Mini 轻量版', 'chat', 128000, 16384, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (6, 1, 'gpt-3.5-turbo', 'gpt-3.5-turbo', 'GPT-3.5 Turbo', 'OpenAI GPT-3.5 Turbo 模型', 'chat', 16385, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (7, 1, 'gpt-3.5-turbo-16k', 'gpt-3.5-turbo-16k', 'GPT-3.5 Turbo 16K', 'OpenAI GPT-3.5 Turbo 16K 上下文模型', 'chat', 16385, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (8, 1, 'text-embedding-3-small', 'text-embedding-3-small', 'Text Embedding 3 Small', 'OpenAI 文本嵌入模型 Small', 'embedding', 8191, null, false, false, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (9, 1, 'text-embedding-3-large', 'text-embedding-3-large', 'Text Embedding 3 Large', 'OpenAI 文本嵌入模型 Large', 'embedding', 8191, null, false, false, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (10, 1, 'text-embedding-ada-002', 'text-embedding-ada-002', 'Text Embedding Ada 002', 'OpenAI 文本嵌入模型 Ada', 'embedding', 8191, null, false, false, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    -- Anthropic 模型
    (11, 2, 'claude-3-5-sonnet-20240620', 'claude-3-5-sonnet', 'Claude 3.5 Sonnet', 'Anthropic Claude 3.5 Sonnet 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (12, 2, 'claude-3-opus-20240229', 'claude-3-opus', 'Claude 3 Opus', 'Anthropic Claude 3 Opus 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (13, 2, 'claude-3-sonnet-20240229', 'claude-3-sonnet', 'Claude 3 Sonnet', 'Anthropic Claude 3 Sonnet 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (14, 2, 'claude-3-haiku-20240307', 'claude-3-haiku', 'Claude 3 Haiku', 'Anthropic Claude 3 Haiku 模型', 'chat', 200000, 4096, true, true, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 插入模型定价数据
-- OpenAI GPT-4 定价 (每1K tokens)
INSERT INTO model_pricing (model_id, pricing_type, price_per_unit, unit, currency, effective_from, effective_until, created_at)
VALUES 
    -- GPT-4 定价
    (1, 'input', 0.03, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (1, 'output', 0.06, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- GPT-4 32K 定价
    (2, 'input', 0.06, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (2, 'output', 0.12, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- GPT-4 Turbo 定价
    (3, 'input', 0.01, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (3, 'output', 0.03, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- GPT-4o 定价
    (4, 'input', 0.005, 'token', 'USD', '2024-05-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (4, 'output', 0.015, 'token', 'USD', '2024-05-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- GPT-4o Mini 定价
    (5, 'input', 0.00015, 'token', 'USD', '2024-07-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (5, 'output', 0.0006, 'token', 'USD', '2024-07-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- GPT-3.5 Turbo 定价
    (6, 'input', 0.0005, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (6, 'output', 0.0015, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- GPT-3.5 Turbo 16K 定价
    (7, 'input', 0.001, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (7, 'output', 0.002, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- Text Embedding 3 Small 定价
    (8, 'input', 0.00002, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- Text Embedding 3 Large 定价
    (9, 'input', 0.00013, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- Text Embedding Ada 002 定价
    (10, 'input', 0.0001, 'token', 'USD', '2024-01-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- Claude 3.5 Sonnet 定价
    (11, 'input', 0.003, 'token', 'USD', '2024-06-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (11, 'output', 0.015, 'token', 'USD', '2024-06-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- Claude 3 Opus 定价
    (12, 'input', 0.015, 'token', 'USD', '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (12, 'output', 0.075, 'token', 'USD', '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- Claude 3 Sonnet 定价
    (13, 'input', 0.003, 'token', 'USD', '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (13, 'output', 0.015, 'token', 'USD', '2024-02-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    
    -- Claude 3 Haiku 定价
    (14, 'input', 0.00025, 'token', 'USD', '2024-03-01 00:00:00', NULL, CURRENT_TIMESTAMP),
    (14, 'output', 0.00125, 'token', 'USD', '2024-03-01 00:00:00', NULL, CURRENT_TIMESTAMP);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_model_pricing_model_id ON model_pricing(model_id);
CREATE INDEX IF NOT EXISTS idx_model_pricing_type ON model_pricing(pricing_type);
CREATE INDEX IF NOT EXISTS idx_model_pricing_effective ON model_pricing(effective_from, effective_until);
CREATE INDEX IF NOT EXISTS idx_model_pricing_model_type ON model_pricing(model_id, pricing_type);
