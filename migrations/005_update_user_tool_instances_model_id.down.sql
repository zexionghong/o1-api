-- 回滚用户工具实例表的model_id字段类型

-- 创建原始的user_tool_instances表结构
CREATE TABLE user_tool_instances_old (
    id VARCHAR(36) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    tool_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    model_id VARCHAR(50) NOT NULL,      -- 恢复为VARCHAR类型
    api_key_id INTEGER NOT NULL,
    config JSON,
    is_public BOOLEAN DEFAULT FALSE,
    share_token VARCHAR(32) UNIQUE,
    usage_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tool_id) REFERENCES tools(id),
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

-- 删除当前表
DROP TABLE user_tool_instances;

-- 重命名回原始表
ALTER TABLE user_tool_instances_old RENAME TO user_tool_instances;
