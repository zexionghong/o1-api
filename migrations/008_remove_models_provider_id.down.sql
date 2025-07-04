-- 008_remove_models_provider_id.down.sql
-- 回滚：恢复 models 表中的 provider_id 字段

-- 检查是否存在备份表
-- 如果不存在备份表，则无法回滚

-- 删除当前的 models 表
DROP TABLE IF EXISTS models;

-- 从备份表恢复原始的 models 表结构
CREATE TABLE models AS SELECT * FROM models_backup;

-- 重新创建索引
CREATE INDEX idx_models_provider ON models(provider_id);
CREATE INDEX idx_models_slug ON models(provider_id, slug);
CREATE INDEX idx_models_type ON models(model_type);
CREATE INDEX idx_models_status ON models(status);

-- 删除在前向迁移中创建的 provider_model_support 数据
-- 注意：这只删除由迁移脚本自动创建的数据，不删除手动添加的数据
DELETE FROM provider_model_support 
WHERE created_at >= (
    SELECT MIN(created_at) FROM models_backup
) AND upstream_model_name = model_slug;

-- 删除备份表
DROP TABLE models_backup;

-- 注意：此回滚脚本假设：
-- 1. models_backup 表仍然存在
-- 2. 没有在迁移后手动修改过 models 或 provider_model_support 表
-- 3. 如果有手动添加的 provider_model_support 数据，需要手动处理
