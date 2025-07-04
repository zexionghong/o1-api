-- 回滚提供商模型支持表

-- 删除索引
DROP INDEX IF EXISTS idx_provider_model_support_priority;
DROP INDEX IF EXISTS idx_provider_model_support_enabled;
DROP INDEX IF EXISTS idx_provider_model_support_model;
DROP INDEX IF EXISTS idx_provider_model_support_provider;

-- 删除表
DROP TABLE IF EXISTS provider_model_support;
