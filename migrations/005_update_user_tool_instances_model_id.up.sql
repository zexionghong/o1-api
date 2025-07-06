-- 更新用户工具实例表的model_id字段类型

-- 创建新的user_tool_instances表结构
CREATE TABLE user_tool_instances_new (
    id VARCHAR(36) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    tool_id VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    model_id INTEGER NOT NULL,          -- 改为INTEGER类型
    api_key_id INTEGER NOT NULL,
    config JSON,
    is_public BOOLEAN DEFAULT FALSE,
    share_token VARCHAR(32) UNIQUE,
    usage_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tool_id) REFERENCES tools(id),
    FOREIGN KEY (model_id) REFERENCES models(id),
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

-- 如果有数据，需要迁移（这里假设没有数据，因为是新功能）
-- 如果有数据，需要根据model名称查找对应的ID进行迁移

-- 删除旧表
DROP TABLE user_tool_instances;

-- 重命名新表
ALTER TABLE user_tool_instances_new RENAME TO user_tool_instances;
