-- 回滚模型价格数据插入

-- 删除索引
DROP INDEX IF EXISTS idx_model_pricing_model_type;
DROP INDEX IF EXISTS idx_model_pricing_effective;
DROP INDEX IF EXISTS idx_model_pricing_type;
DROP INDEX IF EXISTS idx_model_pricing_model_id;

-- 删除模型定价数据
DELETE FROM model_pricing WHERE model_id IN (1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14);

-- 删除模型数据
DELETE FROM models WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14);

-- 删除提供商数据
DELETE FROM providers WHERE id IN (1, 2);
