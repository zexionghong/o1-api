-- ============================================================================
-- 迁移配额系统从基于用户改为基于API Key
-- 版本: 009
-- 描述: 将配额管理从用户级别改为API Key级别，支持总限额和周期限额
-- ============================================================================

-- 备份现有数据（如果需要的话）
-- CREATE TABLE quotas_backup AS SELECT * FROM quotas;
-- CREATE TABLE quota_usage_backup AS SELECT * FROM quota_usage;

-- 删除现有的配额相关索引
DROP INDEX IF EXISTS idx_quotas_user_id;
DROP INDEX IF EXISTS idx_quota_usage_user_id;
DROP INDEX IF EXISTS idx_quota_usage_period;

-- 重命名现有表以备份
ALTER TABLE quotas RENAME TO quotas_old;
ALTER TABLE quota_usage RENAME TO quota_usage_old;

-- 创建新的配额设置表 - 基于API Key的配额管理
CREATE TABLE quotas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 配额设置唯一标识
    api_key_id INTEGER NOT NULL,                             -- 关联的API密钥ID（应用层维护关系）
    quota_type VARCHAR(20) NOT NULL,                        -- 配额类型：requests(请求次数), tokens(token数量), cost(费用金额)
    period VARCHAR(20),                                      -- 配额周期：minute(分钟), hour(小时), day(天), month(月), NULL表示总限额
    limit_value DECIMAL(15,6) NOT NULL,                     -- 配额限制值
    reset_time VARCHAR(10),                                 -- 重置时间，格式HH:MM，用于日/月配额的具体重置时间点
    status VARCHAR(20) NOT NULL DEFAULT 'active',           -- 配额状态：active(生效), inactive(停用)
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 配额创建时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间
    UNIQUE(api_key_id, quota_type, period)                  -- 同一API Key的同类型同周期配额唯一
);

-- 创建新的配额使用情况表 - 基于API Key的使用记录
CREATE TABLE quota_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,                    -- 使用记录唯一标识
    api_key_id INTEGER NOT NULL,                             -- 关联的API密钥ID（应用层维护关系）
    quota_id INTEGER NOT NULL,                               -- 关联的配额设置ID（应用层维护关系）
    period_start DATETIME,                                   -- 统计周期开始时间（总限额时为NULL）
    period_end DATETIME,                                     -- 统计周期结束时间（总限额时为NULL）
    used_value DECIMAL(15,6) NOT NULL DEFAULT 0.000000,     -- 已使用的配额值
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 记录创建时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 最后更新时间
    UNIQUE(api_key_id, quota_id, period_start)              -- 同一API Key同一配额同一周期的记录唯一
);

-- 创建新的索引
CREATE INDEX idx_quotas_api_key_id ON quotas(api_key_id);
CREATE INDEX idx_quotas_status ON quotas(status);
CREATE INDEX idx_quotas_type_period ON quotas(quota_type, period);

CREATE INDEX idx_quota_usage_api_key_id ON quota_usage(api_key_id);
CREATE INDEX idx_quota_usage_quota_id ON quota_usage(quota_id);
CREATE INDEX idx_quota_usage_period ON quota_usage(period_start, period_end);

-- 注释：由于这是一个重大的架构变更，我们不进行数据迁移
-- 如果需要保留现有数据，可以手动执行以下类型的迁移脚本：
-- INSERT INTO quotas (api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at)
-- SELECT ak.id, q.quota_type, q.period, q.limit_value, q.reset_time, q.status, q.created_at, q.updated_at
-- FROM quotas_old q
-- JOIN api_keys ak ON ak.user_id = q.user_id
-- WHERE ak.status = 'active';

-- 清理旧表（可选，建议先保留一段时间）
-- DROP TABLE quotas_old;
-- DROP TABLE quota_usage_old;
