-- 回滚用户密码字段添加
-- 删除密码相关的索引和字段

DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;

-- 删除密码字段（注意：SQLite不支持DROP COLUMN，需要重建表）
-- 创建临时表
CREATE TABLE users_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    balance DECIMAL(15,6) NOT NULL DEFAULT 0.000000,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 复制数据
INSERT INTO users_temp (id, username, email, full_name, status, balance, created_at, updated_at)
SELECT id, username, email, full_name, status, balance, created_at, updated_at FROM users;

-- 删除原表
DROP TABLE users;

-- 重命名临时表
ALTER TABLE users_temp RENAME TO users;
