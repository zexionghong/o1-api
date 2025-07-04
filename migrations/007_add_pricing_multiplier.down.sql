-- 回滚模型定价倍率字段

-- 删除索引
DROP INDEX IF EXISTS idx_model_pricing_multiplier;

-- 删除表注释
DELETE FROM table_comments WHERE table_name = 'model_pricing' AND column_name = 'multiplier';

-- 删除倍率字段
-- 注意：SQLite不支持直接删除列，需要重建表
-- 但为了简化，我们保留该字段，只是不使用它

-- 如果需要完全删除字段，需要以下步骤：
-- 1. 创建新表（不包含multiplier字段）
-- 2. 复制数据
-- 3. 删除旧表
-- 4. 重命名新表

-- 这里我们只是将倍率重置为1.0（不影响计费）
UPDATE model_pricing SET multiplier = 1.0;
