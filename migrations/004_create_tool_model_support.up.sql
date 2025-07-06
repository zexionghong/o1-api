-- 创建工具-模型支持关联表

-- 工具-模型支持关联表
CREATE TABLE tool_model_support (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_id VARCHAR(50) NOT NULL,
    model_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE,
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    UNIQUE(tool_id, model_id)
);

-- 创建索引
CREATE INDEX idx_tool_model_support_tool_id ON tool_model_support(tool_id);
CREATE INDEX idx_tool_model_support_model_id ON tool_model_support(model_id);

-- 迁移现有数据：将tools表中的supported_models JSON数据迁移到关联表
-- 注意：这里需要根据实际的模型名称和ID进行映射

-- Chatbot 支持的模型
INSERT INTO tool_model_support (tool_id, model_id) 
SELECT 'chatbot', m.id 
FROM models m 
WHERE m.name IN ('gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet-20240620');

-- Image Generator 支持的模型  
INSERT INTO tool_model_support (tool_id, model_id)
SELECT 'image_generator', m.id
FROM models m
WHERE m.name IN ('dall-e-3', 'stable-diffusion-xl');

-- Text Generator 支持的模型
INSERT INTO tool_model_support (tool_id, model_id)
SELECT 'text_generator', m.id
FROM models m
WHERE m.name IN ('gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet-20240620');

-- Code Assistant 支持的模型
INSERT INTO tool_model_support (tool_id, model_id)
SELECT 'code_assistant', m.id
FROM models m
WHERE m.name IN ('gpt-4o', 'claude-3-5-sonnet-20240620');

-- 删除tools表中的supported_models字段
ALTER TABLE tools DROP COLUMN supported_models;
