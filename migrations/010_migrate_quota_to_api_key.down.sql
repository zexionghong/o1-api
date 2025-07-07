-- ============================================================================
-- 回滚配额系统从API Key改回基于用户
-- 版本: 009 回滚
-- 描述: 将配额管理从API Key级别改回用户级别
-- ============================================================================

-- 删除新的配额相关索引
DROP INDEX IF EXISTS idx_quotas_api_key_id;
DROP INDEX IF EXISTS idx_quotas_status;
DROP INDEX IF EXISTS idx_quotas_type_period;
DROP INDEX IF EXISTS idx_quota_usage_api_key_id;
DROP INDEX IF EXISTS idx_quota_usage_quota_id;
DROP INDEX IF EXISTS idx_quota_usage_period;

-- 删除新表
DROP TABLE IF EXISTS quotas;
DROP TABLE IF EXISTS quota_usage;

-- 恢复旧表
ALTER TABLE quotas_old RENAME TO quotas;
ALTER TABLE quota_usage_old RENAME TO quota_usage;

-- 重新创建原有索引
CREATE INDEX idx_quotas_user_id ON quotas(user_id);
CREATE INDEX idx_quota_usage_user_id ON quota_usage(user_id);
CREATE INDEX idx_quota_usage_period ON quota_usage(period_start, period_end);
