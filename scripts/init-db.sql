-- =============================================================================
-- Ads Analytics Platform - Database Initialization
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create basic schema if migrations haven't run yet
-- Note: This is a safety net - migrations should handle the actual schema

-- Grant permissions (if using a non-superuser)
-- GRANT ALL PRIVILEGES ON DATABASE ads_aggregator TO ads_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ads_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ads_user;

-- Log initialization
DO $$
BEGIN
    RAISE NOTICE 'Database initialized at %', NOW();
END $$;
