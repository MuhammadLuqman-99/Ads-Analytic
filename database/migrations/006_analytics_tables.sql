-- Analytics Events Table
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,
    properties JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    country VARCHAR(2),
    city VARCHAR(255),
    device_type VARCHAR(50),
    browser VARCHAR(100),
    os VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for analytics_events
CREATE INDEX IF NOT EXISTS idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_org_id ON analytics_events(organization_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_timestamp ON analytics_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_analytics_events_session ON analytics_events(session_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_user_timestamp ON analytics_events(user_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_analytics_events_type_timestamp ON analytics_events(event_type, timestamp);

-- Analytics Profiles Table (aggregated user data)
CREATE TABLE IF NOT EXISTS analytics_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_events INTEGER DEFAULT 0,
    total_sessions INTEGER DEFAULT 0,
    properties JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for analytics_profiles
CREATE INDEX IF NOT EXISTS idx_analytics_profiles_last_seen ON analytics_profiles(last_seen_at);
CREATE INDEX IF NOT EXISTS idx_analytics_profiles_first_seen ON analytics_profiles(first_seen_at);

-- Daily Metrics Aggregation Table (for faster queries)
CREATE TABLE IF NOT EXISTS analytics_daily_metrics (
    date DATE NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(20, 4) NOT NULL,
    dimensions JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (date, metric_name, dimensions)
);

-- Index for daily metrics
CREATE INDEX IF NOT EXISTS idx_analytics_daily_metrics_date ON analytics_daily_metrics(date);
CREATE INDEX IF NOT EXISTS idx_analytics_daily_metrics_name ON analytics_daily_metrics(metric_name);

-- Admin Users Table (for admin panel access)
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Index for admin users
CREATE INDEX IF NOT EXISTS idx_admin_users_user_id ON admin_users(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);

-- Audit Log for Admin Actions
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for audit log
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_admin ON admin_audit_log(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_created ON admin_audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_action ON admin_audit_log(action);

-- Function to update analytics_profiles on event insert
CREATE OR REPLACE FUNCTION update_analytics_profile()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.user_id IS NOT NULL THEN
        INSERT INTO analytics_profiles (user_id, first_seen_at, last_seen_at, total_events)
        VALUES (NEW.user_id, NEW.timestamp, NEW.timestamp, 1)
        ON CONFLICT (user_id) DO UPDATE SET
            last_seen_at = GREATEST(analytics_profiles.last_seen_at, NEW.timestamp),
            total_events = analytics_profiles.total_events + 1,
            updated_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update analytics_profiles
DROP TRIGGER IF EXISTS trg_update_analytics_profile ON analytics_events;
CREATE TRIGGER trg_update_analytics_profile
    AFTER INSERT ON analytics_events
    FOR EACH ROW
    EXECUTE FUNCTION update_analytics_profile();

-- Function to aggregate daily metrics (run via cron)
CREATE OR REPLACE FUNCTION aggregate_daily_metrics(target_date DATE)
RETURNS void AS $$
BEGIN
    -- DAU
    INSERT INTO analytics_daily_metrics (date, metric_name, metric_value)
    SELECT target_date, 'dau', COUNT(DISTINCT user_id)
    FROM analytics_events
    WHERE timestamp >= target_date AND timestamp < target_date + INTERVAL '1 day'
    AND user_id IS NOT NULL
    ON CONFLICT (date, metric_name, dimensions) DO UPDATE SET metric_value = EXCLUDED.metric_value;

    -- New Users
    INSERT INTO analytics_daily_metrics (date, metric_name, metric_value)
    SELECT target_date, 'new_users', COUNT(*)
    FROM users
    WHERE created_at >= target_date AND created_at < target_date + INTERVAL '1 day'
    ON CONFLICT (date, metric_name, dimensions) DO UPDATE SET metric_value = EXCLUDED.metric_value;

    -- Event counts by type
    INSERT INTO analytics_daily_metrics (date, metric_name, metric_value, dimensions)
    SELECT target_date, 'events', COUNT(*), jsonb_build_object('type', event_type)
    FROM analytics_events
    WHERE timestamp >= target_date AND timestamp < target_date + INTERVAL '1 day'
    GROUP BY event_type
    ON CONFLICT (date, metric_name, dimensions) DO UPDATE SET metric_value = EXCLUDED.metric_value;

    -- Platform connections
    INSERT INTO analytics_daily_metrics (date, metric_name, metric_value, dimensions)
    SELECT target_date, 'platform_connections', COUNT(DISTINCT user_id), jsonb_build_object('platform', platform)
    FROM platform_connections
    WHERE status = 'active'
    GROUP BY platform
    ON CONFLICT (date, metric_name, dimensions) DO UPDATE SET metric_value = EXCLUDED.metric_value;
END;
$$ LANGUAGE plpgsql;

-- Comments
COMMENT ON TABLE analytics_events IS 'Stores all analytics events for product usage tracking';
COMMENT ON TABLE analytics_profiles IS 'Aggregated user profiles for analytics';
COMMENT ON TABLE analytics_daily_metrics IS 'Pre-aggregated daily metrics for fast dashboard queries';
COMMENT ON TABLE admin_users IS 'Admin users with elevated privileges';
COMMENT ON TABLE admin_audit_log IS 'Audit log for admin actions';
