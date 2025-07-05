-- 添加用户密码字段
-- 为用户表添加密码哈希字段以支持用户名密码登录

ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);

-- 为现有用户设置默认密码哈希（临时密码：password123）
-- 这是bcrypt哈希值，对应密码 "password123"
UPDATE users SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPeOXANBVA9cyv6YF6oy/1Ry9BtjO' WHERE password_hash IS NULL;

-- 添加索引以提高查询性能
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
