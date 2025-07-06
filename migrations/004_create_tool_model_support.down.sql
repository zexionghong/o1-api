-- 回滚工具-模型支持关联表

-- 重新添加tools表的supported_models字段
ALTER TABLE tools ADD COLUMN supported_models JSON;

-- 恢复原始数据（重新插入JSON格式的支持模型数据）
UPDATE tools SET supported_models = '["gpt-4o", "gpt-4-turbo", "claude-3-5-sonnet"]' WHERE id = 'chatbot';
UPDATE tools SET supported_models = '["dall-e-3", "stable-diffusion-xl"]' WHERE id = 'image_generator';
UPDATE tools SET supported_models = '["gpt-4o", "gpt-4-turbo", "claude-3-5-sonnet"]' WHERE id = 'text_generator';
UPDATE tools SET supported_models = '["gpt-4o", "claude-3-5-sonnet"]' WHERE id = 'code_assistant';

-- 删除关联表
DROP TABLE IF EXISTS tool_model_support;
