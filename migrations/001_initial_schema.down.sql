-- Drop all indexes first
DROP INDEX IF EXISTS idx_billing_records_created_at;
DROP INDEX IF EXISTS idx_billing_records_status;
DROP INDEX IF EXISTS idx_billing_records_user_id;

DROP INDEX IF EXISTS idx_usage_logs_request_id;
DROP INDEX IF EXISTS idx_usage_logs_created_at;
DROP INDEX IF EXISTS idx_usage_logs_api_key_id;
DROP INDEX IF EXISTS idx_usage_logs_user_id;

DROP INDEX IF EXISTS idx_quota_usage_period;
DROP INDEX IF EXISTS idx_quota_usage_user_id;
DROP INDEX IF EXISTS idx_quotas_user_id;

DROP INDEX IF EXISTS idx_model_pricing_effective;
DROP INDEX IF EXISTS idx_model_pricing_model_id;

DROP INDEX IF EXISTS idx_models_status;
DROP INDEX IF EXISTS idx_models_provider_id;

DROP INDEX IF EXISTS idx_api_keys_status;
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_api_keys_user_id;

-- Drop all tables in reverse order
DROP TABLE IF EXISTS billing_records;
DROP TABLE IF EXISTS usage_logs;
DROP TABLE IF EXISTS quota_usage;
DROP TABLE IF EXISTS quotas;
DROP TABLE IF EXISTS model_pricing;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;
