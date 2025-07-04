-- 008_remove_models_provider_id.up.sql
-- 移除 models 表中的 provider_id 字段，因为现在模型和提供商的关系完全由 provider_model_support 表维护

-- 首先检查是否存在依赖于 provider_id 的数据
-- 如果有数据，需要先迁移到 provider_model_support 表

-- 创建临时表来备份现有的模型数据
CREATE TABLE models_backup AS SELECT * FROM models;

-- 删除 models 表中的 provider_id 字段和相关约束
-- 由于 SQLite 不支持直接删除列，我们需要重建表

-- 创建新的 models 表结构（不包含 provider_id）
CREATE TABLE models_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 将数据从旧表迁移到新表（去除重复的模型）
-- 如果同一个 slug 有多个 provider_id，只保留一个
INSERT INTO models_new (id, name, slug, display_name, description, model_type, context_length, max_tokens, supports_streaming, supports_functions, status, created_at, updated_at)
SELECT 
    MIN(id) as id,
    name,
    slug,
    display_name,
    description,
    model_type,
    context_length,
    max_tokens,
    supports_streaming,
    supports_functions,
    status,
    MIN(created_at) as created_at,
    MAX(updated_at) as updated_at
FROM models
GROUP BY slug;

-- 确保 provider_model_support 表中有对应的关系数据
-- 为每个原来的 (provider_id, model_slug) 组合创建支持关系
INSERT OR IGNORE INTO provider_model_support (provider_id, model_slug, upstream_model_name, enabled, priority, created_at, updated_at)
SELECT 
    m.provider_id,
    m.slug,
    m.slug as upstream_model_name,
    true as enabled,
    1 as priority,
    m.created_at,
    m.updated_at
FROM models m
WHERE NOT EXISTS (
    SELECT 1 FROM provider_model_support pms 
    WHERE pms.provider_id = m.provider_id AND pms.model_slug = m.slug
);

-- 删除旧表并重命名新表
DROP TABLE models;
ALTER TABLE models_new RENAME TO models;

-- 重新创建索引
CREATE INDEX idx_models_slug ON models(slug);
CREATE INDEX idx_models_type ON models(model_type);
CREATE INDEX idx_models_status ON models(status);

-- 删除备份表（如果迁移成功）
-- DROP TABLE models_backup;

-- 添加注释
-- 注意：models_backup 表被保留以防需要回滚
-- 如果确认迁移成功，可以手动删除 models_backup 表
