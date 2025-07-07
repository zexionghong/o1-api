-- 用户工具表 - 存储用户创建的AI工具
CREATE TABLE user_tools (
    id VARCHAR(36) PRIMARY KEY DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    user_id INTEGER NOT NULL,                    -- 关联的用户ID
    name VARCHAR(100) NOT NULL,                  -- 工具名称
    description TEXT,                            -- 工具描述
    type VARCHAR(50) NOT NULL,                   -- 工具类型：chatbot, image_generator, text_generator, code_assistant
    model_id VARCHAR(50) NOT NULL,               -- 使用的模型ID
    api_key_id VARCHAR(36) NOT NULL,             -- 关联的API密钥ID
    config JSON,                                 -- 工具配置（JSON格式）
    is_public BOOLEAN DEFAULT FALSE,             -- 是否公开
    share_token VARCHAR(32) UNIQUE,              -- 分享令牌
    usage_count INTEGER DEFAULT 0,               -- 使用次数
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_user_tools_user_id ON user_tools(user_id);
CREATE INDEX idx_user_tools_type ON user_tools(type);
CREATE INDEX idx_user_tools_is_public ON user_tools(is_public);
CREATE INDEX idx_user_tools_share_token ON user_tools(share_token);
CREATE INDEX idx_user_tools_created_at ON user_tools(created_at);

-- 工具使用记录表 - 记录工具的使用情况
CREATE TABLE tool_usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_id VARCHAR(36) NOT NULL,               -- 工具ID
    user_id INTEGER,                             -- 使用者ID（可能为空，如果是匿名访问分享链接）
    session_id VARCHAR(64),                      -- 会话ID
    input_tokens INTEGER DEFAULT 0,             -- 输入令牌数
    output_tokens INTEGER DEFAULT 0,            -- 输出令牌数
    cost DECIMAL(10, 6) DEFAULT 0,              -- 费用
    status VARCHAR(20) DEFAULT 'success',       -- 状态：success, error, timeout
    error_message TEXT,                          -- 错误信息
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (tool_id) REFERENCES user_tools(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_tool_usage_logs_tool_id ON tool_usage_logs(tool_id);
CREATE INDEX idx_tool_usage_logs_user_id ON tool_usage_logs(user_id);
CREATE INDEX idx_tool_usage_logs_created_at ON tool_usage_logs(created_at);

-- 工具收藏表 - 用户收藏的工具
CREATE TABLE tool_favorites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,                   -- 用户ID
    tool_id VARCHAR(36) NOT NULL,               -- 工具ID
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (tool_id) REFERENCES user_tools(id) ON DELETE CASCADE,
    UNIQUE(user_id, tool_id)
);

-- 创建索引
CREATE INDEX idx_tool_favorites_user_id ON tool_favorites(user_id);
CREATE INDEX idx_tool_favorites_tool_id ON tool_favorites(tool_id);
