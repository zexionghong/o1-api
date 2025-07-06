-- 创建工具相关表

-- 工具模板表 - 存储系统预定义的工具模板
CREATE TABLE tools (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    icon VARCHAR(100),
    color VARCHAR(20),
    supported_models JSON, -- 支持的模型ID列表
    config_schema JSON,    -- 工具配置的JSON Schema
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 用户工具实例表 - 存储用户创建的工具实例
CREATE TABLE user_tool_instances (
    id VARCHAR(36) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    tool_id VARCHAR(50) NOT NULL,      -- 关联工具模板
    name VARCHAR(100) NOT NULL,        -- 用户自定义名称
    description TEXT,                  -- 用户自定义描述
    model_id VARCHAR(50) NOT NULL,     -- 用户选择的模型
    api_key_id INTEGER NOT NULL,       -- 用户选择的API Key
    config JSON,                       -- 用户自定义配置
    is_public BOOLEAN DEFAULT FALSE,   -- 是否公开分享
    share_token VARCHAR(32) UNIQUE,    -- 分享token
    usage_count INTEGER DEFAULT 0,     -- 使用次数
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tool_id) REFERENCES tools(id),
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

-- 工具使用记录表 - 记录工具实例的使用情况
CREATE TABLE tool_usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_instance_id VARCHAR(36) NOT NULL,  -- 关联用户工具实例
    user_id INTEGER,                        -- 使用者ID（可能不是工具创建者）
    session_id VARCHAR(64),
    request_count INTEGER DEFAULT 1,
    tokens_used INTEGER DEFAULT 0,
    cost DECIMAL(10, 6) DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (tool_instance_id) REFERENCES user_tool_instances(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- 插入默认的工具模板
INSERT INTO tools (id, name, description, category, icon, color, supported_models, config_schema) VALUES
('chatbot', 'AI Chatbot', 'Create intelligent conversational AI', 'Communication', 'solar:chat-round-bold-duotone', '#45B7D1',
 '["gpt-4o", "gpt-4-turbo", "claude-3-5-sonnet"]',
 '{"type": "object", "properties": {"system_prompt": {"type": "string", "default": "You are a helpful assistant."}, "temperature": {"type": "number", "minimum": 0, "maximum": 2, "default": 0.7}, "max_tokens": {"type": "integer", "minimum": 1, "maximum": 4000, "default": 2000}}}'),

('image_generator', 'Image Generator', 'Generate images from text descriptions', 'Creative', 'solar:gallery-bold-duotone', '#4ECDC4',
 '["dall-e-3", "stable-diffusion-xl"]',
 '{"type": "object", "properties": {"style": {"type": "string", "enum": ["natural", "vivid"], "default": "natural"}, "size": {"type": "string", "enum": ["1024x1024", "1792x1024", "1024x1792"], "default": "1024x1024"}, "quality": {"type": "string", "enum": ["standard", "hd"], "default": "standard"}}}'),

('text_generator', 'Text Generator', 'Generate and edit text content', 'Creative', 'solar:text-bold-duotone', '#FF6B6B',
 '["gpt-4o", "gpt-4-turbo", "claude-3-5-sonnet"]',
 '{"type": "object", "properties": {"writing_style": {"type": "string", "enum": ["formal", "casual", "creative", "technical"], "default": "casual"}, "temperature": {"type": "number", "minimum": 0, "maximum": 2, "default": 0.7}, "max_tokens": {"type": "integer", "minimum": 1, "maximum": 4000, "default": 1000}}}'),

('code_assistant', 'Code Assistant', 'AI-powered coding helper', 'Development', 'solar:code-bold-duotone', '#FFEAA7',
 '["gpt-4o", "claude-3-5-sonnet"]',
 '{"type": "object", "properties": {"language": {"type": "string", "enum": ["javascript", "python", "go", "java", "typescript", "rust"], "default": "javascript"}, "code_style": {"type": "string", "enum": ["clean", "commented", "optimized"], "default": "clean"}, "max_tokens": {"type": "integer", "minimum": 1, "maximum": 4000, "default": 2000}}}');
