-- 简化API Key存储：将key_hash改为key，直接存储完整的API Key
-- 这样可以直接查询和返回完整的API Key

-- 1. 创建新的临时表
CREATE TABLE api_keys_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    key VARCHAR(255) NOT NULL UNIQUE,  -- 直接存储完整的API Key
    key_prefix VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, expired, revoked
    permissions TEXT, -- JSON格式存储权限
    expires_at DATETIME,
    last_used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. 如果有现有数据，需要手动处理（因为无法从hash反推原始key）
-- 这里我们只是创建新表结构，现有数据将丢失
-- 在生产环境中，需要提前通知用户重新生成API Key

-- 3. 删除旧表
DROP TABLE api_keys;

-- 4. 重命名新表
ALTER TABLE api_keys_new RENAME TO api_keys;

-- 5. 重新创建索引
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key ON api_keys(key);  -- 将key_hash索引改为key索引
CREATE INDEX idx_api_keys_status ON api_keys(status);
