-- 回滚：将key字段改回key_hash

-- 1. 删除索引
DROP INDEX IF EXISTS idx_api_keys_key;
DROP INDEX IF EXISTS idx_api_keys_status;
DROP INDEX IF EXISTS idx_api_keys_user_id;

-- 2. 创建旧的表结构
CREATE TABLE api_keys_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, expired, revoked
    permissions TEXT, -- JSON格式存储权限
    expires_at DATETIME,
    last_used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. 删除新表
DROP TABLE api_keys;

-- 4. 重命名回原来的表名
ALTER TABLE api_keys_old RENAME TO api_keys;

-- 5. 重新创建原来的索引
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_status ON api_keys(status);
