-- 创建提供商模型支持表，支持多对多关系
-- 这样一个提供商可以支持多个模型，一个模型也可以被多个提供商支持

CREATE TABLE provider_model_support (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id INTEGER NOT NULL,
    model_slug VARCHAR(100) NOT NULL,
    upstream_model_name VARCHAR(100), -- 上游实际使用的模型名（可能与model_slug不同）
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 1, -- 该提供商对此模型的优先级（数字越小优先级越高）
    config TEXT, -- JSON格式的额外配置（如特殊参数映射等）
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, model_slug)
);

-- 插入现有的模型支持关系（基于当前的models表）
INSERT INTO provider_model_support (provider_id, model_slug, upstream_model_name, priority)
SELECT 
    provider_id,
    slug as model_slug,
    slug as upstream_model_name, -- 默认情况下上游模型名与slug相同
    1 as priority
FROM models 
WHERE status = 'active';

-- 示例：添加一个兼容所有OpenAI模型的提供商支持
-- 假设我们有一个provider_id=3的兼容提供商
-- INSERT INTO provider_model_support (provider_id, model_slug, upstream_model_name, priority) VALUES
-- (3, 'gpt-4', 'gpt-4', 2),
-- (3, 'gpt-4-turbo', 'gpt-4-turbo', 2),
-- (3, 'gpt-3.5-turbo', 'gpt-3.5-turbo', 2),
-- (3, 'claude-3-opus', 'claude-3-opus-20240229', 2); -- 上游使用不同的模型名

-- 创建索引以提高查询性能
CREATE INDEX idx_provider_model_support_provider ON provider_model_support(provider_id);
CREATE INDEX idx_provider_model_support_model ON provider_model_support(model_slug);
CREATE INDEX idx_provider_model_support_enabled ON provider_model_support(enabled);
CREATE INDEX idx_provider_model_support_priority ON provider_model_support(model_slug, priority);
