-- 添加模型定价倍率字段
-- 用于在原有价格基础上按倍率计费

-- 为 model_pricing 表添加倍率字段
ALTER TABLE model_pricing ADD COLUMN multiplier DECIMAL(5,2) NOT NULL DEFAULT 1.5;

-- 更新表注释
INSERT OR REPLACE INTO table_comments (table_name, column_name, comment_text) VALUES
('model_pricing', 'multiplier', '价格倍率，在原价基础上的倍数，默认1.5倍');

-- 更新现有记录的倍率为1.5
UPDATE model_pricing SET multiplier = 1.5 WHERE multiplier = 1.5;

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_model_pricing_multiplier ON model_pricing(model_id, pricing_type, multiplier);
