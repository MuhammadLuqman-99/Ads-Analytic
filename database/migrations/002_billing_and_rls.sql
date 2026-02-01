-- ============================================================================
-- SaaS Ads Analytics Platform - Billing & Row-Level Security
-- Version: 2.0.0
-- ============================================================================

-- ============================================================================
-- NEW ENUM TYPES FOR BILLING
-- ============================================================================

-- Drop old subscription_plan if exists and create new plan_tier
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'plan_tier') THEN
        CREATE TYPE plan_tier AS ENUM ('free', 'pro', 'business');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'subscription_status') THEN
        CREATE TYPE subscription_status AS ENUM (
            'active',
            'past_due',
            'canceled',
            'trialing',
            'paused',
            'unpaid',
            'incomplete'
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'billing_cycle') THEN
        CREATE TYPE billing_cycle AS ENUM ('monthly', 'yearly');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_status') THEN
        CREATE TYPE payment_status AS ENUM (
            'pending',
            'succeeded',
            'failed',
            'refunded',
            'disputed'
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'usage_type') THEN
        CREATE TYPE usage_type AS ENUM (
            'api_call',
            'data_sync',
            'report_generate',
            'webhook_receive',
            'export'
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'quota_alert_type') THEN
        CREATE TYPE quota_alert_type AS ENUM ('warning', 'critical', 'exceeded');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'plan_change_type') THEN
        CREATE TYPE plan_change_type AS ENUM ('upgrade', 'downgrade', 'cycle_change');
    END IF;
END $$;

-- ============================================================================
-- SUBSCRIPTIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Plan info
    plan_tier plan_tier NOT NULL DEFAULT 'free',
    status subscription_status NOT NULL DEFAULT 'active',
    billing_cycle billing_cycle,

    -- Stripe integration
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    stripe_price_id VARCHAR(255),

    -- Billing period
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    trial_ends_at TIMESTAMP WITH TIME ZONE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN DEFAULT false,

    -- Payment info
    last_payment_at TIMESTAMP WITH TIME ZONE,
    last_payment_amount DECIMAL(10, 2),
    payment_fail_count INTEGER DEFAULT 0,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_org ON subscriptions(organization_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_customer ON subscriptions(stripe_customer_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_subscription ON subscriptions(stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_tier ON subscriptions(plan_tier);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);

-- ============================================================================
-- PAYMENT HISTORY TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS payment_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,

    -- Payment details
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'MYR',
    status payment_status NOT NULL,
    description VARCHAR(500),

    -- Stripe references
    stripe_payment_intent_id VARCHAR(255),
    stripe_invoice_id VARCHAR(255),
    stripe_charge_id VARCHAR(255),

    -- Invoice
    invoice_number VARCHAR(50),
    invoice_url VARCHAR(500),
    invoice_pdf VARCHAR(500),

    -- Payment method
    payment_method VARCHAR(50),
    payment_method_last4 VARCHAR(4),
    payment_method_brand VARCHAR(20),

    -- Timestamps
    paid_at TIMESTAMP WITH TIME ZONE,
    refunded_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    fail_reason VARCHAR(500),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_history_org ON payment_history(organization_id);
CREATE INDEX IF NOT EXISTS idx_payment_history_subscription ON payment_history(subscription_id);
CREATE INDEX IF NOT EXISTS idx_payment_history_stripe_payment ON payment_history(stripe_payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_payment_history_status ON payment_history(status);
CREATE INDEX IF NOT EXISTS idx_payment_history_created ON payment_history(created_at);

-- ============================================================================
-- ORGANIZATION USAGE TABLE (Daily tracking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS organization_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    date DATE NOT NULL,

    -- API Usage
    api_calls_count BIGINT DEFAULT 0,
    api_calls_limit BIGINT DEFAULT 100,

    -- Data Sync
    data_sync_count BIGINT DEFAULT 0,
    records_synced BIGINT DEFAULT 0,

    -- Reports
    reports_generated BIGINT DEFAULT 0,
    exports_count BIGINT DEFAULT 0,

    -- Webhooks
    webhooks_received BIGINT DEFAULT 0,

    -- Storage (in bytes)
    storage_used_bytes BIGINT DEFAULT 0,
    storage_limit_bytes BIGINT DEFAULT 104857600, -- 100MB default

    -- Connected accounts
    connected_accounts INTEGER DEFAULT 0,
    accounts_limit INTEGER DEFAULT 1,

    -- Active users
    active_users_count INTEGER DEFAULT 0,
    users_limit INTEGER DEFAULT 1,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id, date)
);

CREATE INDEX IF NOT EXISTS idx_org_usage_org ON organization_usage(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_usage_date ON organization_usage(date);
CREATE INDEX IF NOT EXISTS idx_org_usage_org_date ON organization_usage(organization_id, date);

-- ============================================================================
-- USAGE EVENTS TABLE (Detailed logging)
-- ============================================================================

CREATE TABLE IF NOT EXISTS usage_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Event details
    event_type usage_type NOT NULL,
    event_action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(255),

    -- Request info
    request_method VARCHAR(10),
    request_path VARCHAR(500),
    response_status INTEGER,
    response_time_ms INTEGER,

    -- Size metrics
    request_size_bytes BIGINT,
    response_size_bytes BIGINT,

    -- IP and user agent
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_usage_events_org ON usage_events(organization_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_user ON usage_events(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_type ON usage_events(event_type);
CREATE INDEX IF NOT EXISTS idx_usage_events_created ON usage_events(created_at);
CREATE INDEX IF NOT EXISTS idx_usage_events_org_created ON usage_events(organization_id, created_at);

-- ============================================================================
-- FEATURE USAGE TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS feature_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    feature_key VARCHAR(50) NOT NULL,

    -- Usage
    usage_count BIGINT DEFAULT 0,
    usage_limit BIGINT DEFAULT -1, -- -1 = unlimited
    last_used_at TIMESTAMP WITH TIME ZONE,

    -- Period
    period_type VARCHAR(20), -- daily, monthly, lifetime
    period_start TIMESTAMP WITH TIME ZONE,

    -- Status
    is_enabled BOOLEAN DEFAULT true,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id, feature_key)
);

CREATE INDEX IF NOT EXISTS idx_feature_usage_org ON feature_usage(organization_id);
CREATE INDEX IF NOT EXISTS idx_feature_usage_key ON feature_usage(feature_key);

-- ============================================================================
-- QUOTA ALERTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS quota_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Alert info
    alert_type quota_alert_type NOT NULL,
    quota_type VARCHAR(50) NOT NULL, -- api_calls, storage, accounts
    current_usage BIGINT,
    usage_limit BIGINT,
    usage_percent DECIMAL(5, 2),

    -- Notification
    notified_at TIMESTAMP WITH TIME ZONE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID REFERENCES users(id),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_quota_alerts_org ON quota_alerts(organization_id);
CREATE INDEX IF NOT EXISTS idx_quota_alerts_type ON quota_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_quota_alerts_created ON quota_alerts(created_at);

-- ============================================================================
-- PLAN CHANGE REQUESTS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS plan_change_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,

    -- Change details
    change_type plan_change_type NOT NULL,
    from_plan plan_tier NOT NULL,
    to_plan plan_tier NOT NULL,
    from_cycle billing_cycle,
    to_cycle billing_cycle,

    -- Scheduling
    effective_at TIMESTAMP WITH TIME ZONE NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, processed, canceled

    -- Proration
    prorated_amount DECIMAL(10, 2),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plan_change_org ON plan_change_requests(organization_id);
CREATE INDEX IF NOT EXISTS idx_plan_change_subscription ON plan_change_requests(subscription_id);
CREATE INDEX IF NOT EXISTS idx_plan_change_status ON plan_change_requests(status);
CREATE INDEX IF NOT EXISTS idx_plan_change_effective ON plan_change_requests(effective_at);

-- ============================================================================
-- COUPONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS coupons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),

    -- Discount
    discount_type VARCHAR(20) NOT NULL, -- percentage, fixed
    discount_value DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3),

    -- Limits
    max_redemptions INTEGER,
    current_redemptions INTEGER DEFAULT 0,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_until TIMESTAMP WITH TIME ZONE,
    applicable_plans TEXT[], -- Array of plan tiers
    duration_months INTEGER, -- 0 = forever

    -- Stripe
    stripe_coupon_id VARCHAR(255),

    -- Status
    is_active BOOLEAN DEFAULT true,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_active ON coupons(is_active);
CREATE INDEX IF NOT EXISTS idx_coupons_valid ON coupons(valid_from, valid_until);

-- ============================================================================
-- CREDIT TRANSACTIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS credit_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Transaction
    type VARCHAR(20) NOT NULL, -- purchase, usage, refund, bonus
    amount DECIMAL(10, 2) NOT NULL, -- positive or negative
    balance_after DECIMAL(10, 2) NOT NULL,
    description VARCHAR(500),

    -- Reference
    reference_type VARCHAR(50), -- payment, api_call, etc.
    reference_id VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_credit_tx_org ON credit_transactions(organization_id);
CREATE INDEX IF NOT EXISTS idx_credit_tx_type ON credit_transactions(type);
CREATE INDEX IF NOT EXISTS idx_credit_tx_created ON credit_transactions(created_at);

-- ============================================================================
-- TRIGGERS FOR updated_at
-- ============================================================================

CREATE TRIGGER update_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_history_updated_at
    BEFORE UPDATE ON payment_history
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_org_usage_updated_at
    BEFORE UPDATE ON organization_usage
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_feature_usage_updated_at
    BEFORE UPDATE ON feature_usage
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plan_change_updated_at
    BEFORE UPDATE ON plan_change_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_coupons_updated_at
    BEFORE UPDATE ON coupons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ROW-LEVEL SECURITY POLICIES
-- ============================================================================

-- Enable RLS on all tenant-scoped tables
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE connected_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE ad_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE ad_sets ENABLE ROW LEVEL SECURITY;
ALTER TABLE ads ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaign_metrics_daily ENABLE ROW LEVEL SECURITY;
ALTER TABLE ad_set_metrics_daily ENABLE ROW LEVEL SECURITY;
ALTER TABLE ad_metrics_daily ENABLE ROW LEVEL SECURITY;
ALTER TABLE shopee_shops ENABLE ROW LEVEL SECURITY;
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_summary_daily ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE payment_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE feature_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE quota_alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE plan_change_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_transactions ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- RLS Helper Function
-- Get current user's organization IDs from JWT claims
-- ============================================================================

CREATE OR REPLACE FUNCTION current_user_org_ids() RETURNS UUID[] AS $$
BEGIN
    -- This function returns the organization IDs that the current user has access to
    -- The value is set via: SET LOCAL app.current_org_ids = 'uuid1,uuid2,...';
    RETURN string_to_array(current_setting('app.current_org_ids', true), ',')::UUID[];
EXCEPTION
    WHEN OTHERS THEN
        RETURN ARRAY[]::UUID[];
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION current_user_org_id() RETURNS UUID AS $$
BEGIN
    -- Returns the primary organization ID for the current request
    -- Set via: SET LOCAL app.current_org_id = 'uuid';
    RETURN current_setting('app.current_org_id', true)::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION is_org_member(org_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    RETURN org_id = ANY(current_user_org_ids());
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================================
-- RLS POLICIES FOR ORGANIZATIONS
-- ============================================================================

DROP POLICY IF EXISTS org_select_policy ON organizations;
CREATE POLICY org_select_policy ON organizations
    FOR SELECT
    USING (id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_insert_policy ON organizations;
CREATE POLICY org_insert_policy ON organizations
    FOR INSERT
    WITH CHECK (true); -- Anyone can create an organization

DROP POLICY IF EXISTS org_update_policy ON organizations;
CREATE POLICY org_update_policy ON organizations
    FOR UPDATE
    USING (id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_delete_policy ON organizations;
CREATE POLICY org_delete_policy ON organizations
    FOR DELETE
    USING (id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR ORGANIZATION MEMBERS
-- ============================================================================

DROP POLICY IF EXISTS org_members_select_policy ON organization_members;
CREATE POLICY org_members_select_policy ON organization_members
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_members_insert_policy ON organization_members;
CREATE POLICY org_members_insert_policy ON organization_members
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_members_update_policy ON organization_members;
CREATE POLICY org_members_update_policy ON organization_members
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_members_delete_policy ON organization_members;
CREATE POLICY org_members_delete_policy ON organization_members
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR CONNECTED ACCOUNTS
-- ============================================================================

DROP POLICY IF EXISTS connected_accounts_select_policy ON connected_accounts;
CREATE POLICY connected_accounts_select_policy ON connected_accounts
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS connected_accounts_insert_policy ON connected_accounts;
CREATE POLICY connected_accounts_insert_policy ON connected_accounts
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS connected_accounts_update_policy ON connected_accounts;
CREATE POLICY connected_accounts_update_policy ON connected_accounts
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS connected_accounts_delete_policy ON connected_accounts;
CREATE POLICY connected_accounts_delete_policy ON connected_accounts
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR AD ACCOUNTS
-- ============================================================================

DROP POLICY IF EXISTS ad_accounts_select_policy ON ad_accounts;
CREATE POLICY ad_accounts_select_policy ON ad_accounts
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_accounts_insert_policy ON ad_accounts;
CREATE POLICY ad_accounts_insert_policy ON ad_accounts
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_accounts_update_policy ON ad_accounts;
CREATE POLICY ad_accounts_update_policy ON ad_accounts
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_accounts_delete_policy ON ad_accounts;
CREATE POLICY ad_accounts_delete_policy ON ad_accounts
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR CAMPAIGNS
-- ============================================================================

DROP POLICY IF EXISTS campaigns_select_policy ON campaigns;
CREATE POLICY campaigns_select_policy ON campaigns
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS campaigns_insert_policy ON campaigns;
CREATE POLICY campaigns_insert_policy ON campaigns
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS campaigns_update_policy ON campaigns;
CREATE POLICY campaigns_update_policy ON campaigns
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS campaigns_delete_policy ON campaigns;
CREATE POLICY campaigns_delete_policy ON campaigns
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR AD SETS
-- ============================================================================

DROP POLICY IF EXISTS ad_sets_select_policy ON ad_sets;
CREATE POLICY ad_sets_select_policy ON ad_sets
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_sets_insert_policy ON ad_sets;
CREATE POLICY ad_sets_insert_policy ON ad_sets
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_sets_update_policy ON ad_sets;
CREATE POLICY ad_sets_update_policy ON ad_sets
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_sets_delete_policy ON ad_sets;
CREATE POLICY ad_sets_delete_policy ON ad_sets
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR ADS
-- ============================================================================

DROP POLICY IF EXISTS ads_select_policy ON ads;
CREATE POLICY ads_select_policy ON ads
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ads_insert_policy ON ads;
CREATE POLICY ads_insert_policy ON ads
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ads_update_policy ON ads;
CREATE POLICY ads_update_policy ON ads
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ads_delete_policy ON ads;
CREATE POLICY ads_delete_policy ON ads
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR METRICS TABLES
-- ============================================================================

-- Campaign Metrics
DROP POLICY IF EXISTS campaign_metrics_select_policy ON campaign_metrics_daily;
CREATE POLICY campaign_metrics_select_policy ON campaign_metrics_daily
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS campaign_metrics_insert_policy ON campaign_metrics_daily;
CREATE POLICY campaign_metrics_insert_policy ON campaign_metrics_daily
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS campaign_metrics_update_policy ON campaign_metrics_daily;
CREATE POLICY campaign_metrics_update_policy ON campaign_metrics_daily
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS campaign_metrics_delete_policy ON campaign_metrics_daily;
CREATE POLICY campaign_metrics_delete_policy ON campaign_metrics_daily
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Ad Set Metrics
DROP POLICY IF EXISTS ad_set_metrics_select_policy ON ad_set_metrics_daily;
CREATE POLICY ad_set_metrics_select_policy ON ad_set_metrics_daily
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_set_metrics_insert_policy ON ad_set_metrics_daily;
CREATE POLICY ad_set_metrics_insert_policy ON ad_set_metrics_daily
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_set_metrics_update_policy ON ad_set_metrics_daily;
CREATE POLICY ad_set_metrics_update_policy ON ad_set_metrics_daily
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_set_metrics_delete_policy ON ad_set_metrics_daily;
CREATE POLICY ad_set_metrics_delete_policy ON ad_set_metrics_daily
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Ad Metrics
DROP POLICY IF EXISTS ad_metrics_select_policy ON ad_metrics_daily;
CREATE POLICY ad_metrics_select_policy ON ad_metrics_daily
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_metrics_insert_policy ON ad_metrics_daily;
CREATE POLICY ad_metrics_insert_policy ON ad_metrics_daily
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_metrics_update_policy ON ad_metrics_daily;
CREATE POLICY ad_metrics_update_policy ON ad_metrics_daily
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS ad_metrics_delete_policy ON ad_metrics_daily;
CREATE POLICY ad_metrics_delete_policy ON ad_metrics_daily
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR SHOPEE/ORDERS
-- ============================================================================

-- Shopee Shops
DROP POLICY IF EXISTS shopee_shops_select_policy ON shopee_shops;
CREATE POLICY shopee_shops_select_policy ON shopee_shops
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS shopee_shops_insert_policy ON shopee_shops;
CREATE POLICY shopee_shops_insert_policy ON shopee_shops
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS shopee_shops_update_policy ON shopee_shops;
CREATE POLICY shopee_shops_update_policy ON shopee_shops
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS shopee_shops_delete_policy ON shopee_shops;
CREATE POLICY shopee_shops_delete_policy ON shopee_shops
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Orders
DROP POLICY IF EXISTS orders_select_policy ON orders;
CREATE POLICY orders_select_policy ON orders
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS orders_insert_policy ON orders;
CREATE POLICY orders_insert_policy ON orders
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS orders_update_policy ON orders;
CREATE POLICY orders_update_policy ON orders
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS orders_delete_policy ON orders;
CREATE POLICY orders_delete_policy ON orders
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Order Items
DROP POLICY IF EXISTS order_items_select_policy ON order_items;
CREATE POLICY order_items_select_policy ON order_items
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS order_items_insert_policy ON order_items;
CREATE POLICY order_items_insert_policy ON order_items
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS order_items_update_policy ON order_items;
CREATE POLICY order_items_update_policy ON order_items
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS order_items_delete_policy ON order_items;
CREATE POLICY order_items_delete_policy ON order_items
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Sales Summary
DROP POLICY IF EXISTS sales_summary_select_policy ON sales_summary_daily;
CREATE POLICY sales_summary_select_policy ON sales_summary_daily
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS sales_summary_insert_policy ON sales_summary_daily;
CREATE POLICY sales_summary_insert_policy ON sales_summary_daily
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS sales_summary_update_policy ON sales_summary_daily;
CREATE POLICY sales_summary_update_policy ON sales_summary_daily
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS sales_summary_delete_policy ON sales_summary_daily;
CREATE POLICY sales_summary_delete_policy ON sales_summary_daily
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- RLS POLICIES FOR BILLING TABLES
-- ============================================================================

-- Subscriptions
DROP POLICY IF EXISTS subscriptions_select_policy ON subscriptions;
CREATE POLICY subscriptions_select_policy ON subscriptions
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS subscriptions_insert_policy ON subscriptions;
CREATE POLICY subscriptions_insert_policy ON subscriptions
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS subscriptions_update_policy ON subscriptions;
CREATE POLICY subscriptions_update_policy ON subscriptions
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS subscriptions_delete_policy ON subscriptions;
CREATE POLICY subscriptions_delete_policy ON subscriptions
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Payment History
DROP POLICY IF EXISTS payment_history_select_policy ON payment_history;
CREATE POLICY payment_history_select_policy ON payment_history
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS payment_history_insert_policy ON payment_history;
CREATE POLICY payment_history_insert_policy ON payment_history
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS payment_history_update_policy ON payment_history;
CREATE POLICY payment_history_update_policy ON payment_history
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS payment_history_delete_policy ON payment_history;
CREATE POLICY payment_history_delete_policy ON payment_history
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Organization Usage
DROP POLICY IF EXISTS org_usage_select_policy ON organization_usage;
CREATE POLICY org_usage_select_policy ON organization_usage
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_usage_insert_policy ON organization_usage;
CREATE POLICY org_usage_insert_policy ON organization_usage
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_usage_update_policy ON organization_usage;
CREATE POLICY org_usage_update_policy ON organization_usage
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS org_usage_delete_policy ON organization_usage;
CREATE POLICY org_usage_delete_policy ON organization_usage
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Usage Events
DROP POLICY IF EXISTS usage_events_select_policy ON usage_events;
CREATE POLICY usage_events_select_policy ON usage_events
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS usage_events_insert_policy ON usage_events;
CREATE POLICY usage_events_insert_policy ON usage_events
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

-- Feature Usage
DROP POLICY IF EXISTS feature_usage_select_policy ON feature_usage;
CREATE POLICY feature_usage_select_policy ON feature_usage
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS feature_usage_insert_policy ON feature_usage;
CREATE POLICY feature_usage_insert_policy ON feature_usage
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS feature_usage_update_policy ON feature_usage;
CREATE POLICY feature_usage_update_policy ON feature_usage
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS feature_usage_delete_policy ON feature_usage;
CREATE POLICY feature_usage_delete_policy ON feature_usage
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Quota Alerts
DROP POLICY IF EXISTS quota_alerts_select_policy ON quota_alerts;
CREATE POLICY quota_alerts_select_policy ON quota_alerts
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS quota_alerts_insert_policy ON quota_alerts;
CREATE POLICY quota_alerts_insert_policy ON quota_alerts
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS quota_alerts_update_policy ON quota_alerts;
CREATE POLICY quota_alerts_update_policy ON quota_alerts
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS quota_alerts_delete_policy ON quota_alerts;
CREATE POLICY quota_alerts_delete_policy ON quota_alerts
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Plan Change Requests
DROP POLICY IF EXISTS plan_change_select_policy ON plan_change_requests;
CREATE POLICY plan_change_select_policy ON plan_change_requests
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS plan_change_insert_policy ON plan_change_requests;
CREATE POLICY plan_change_insert_policy ON plan_change_requests
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS plan_change_update_policy ON plan_change_requests;
CREATE POLICY plan_change_update_policy ON plan_change_requests
    FOR UPDATE
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS plan_change_delete_policy ON plan_change_requests;
CREATE POLICY plan_change_delete_policy ON plan_change_requests
    FOR DELETE
    USING (organization_id = ANY(current_user_org_ids()));

-- Credit Transactions
DROP POLICY IF EXISTS credit_tx_select_policy ON credit_transactions;
CREATE POLICY credit_tx_select_policy ON credit_transactions
    FOR SELECT
    USING (organization_id = ANY(current_user_org_ids()));

DROP POLICY IF EXISTS credit_tx_insert_policy ON credit_transactions;
CREATE POLICY credit_tx_insert_policy ON credit_transactions
    FOR INSERT
    WITH CHECK (organization_id = ANY(current_user_org_ids()));

-- ============================================================================
-- BYPASS RLS FOR SERVICE ROLE
-- Create a service role that bypasses RLS for background jobs
-- ============================================================================

-- Policies for service role (bypasses RLS)
-- This should be run with superuser privileges

-- Grant all on tables to service role (run as superuser)
-- GRANT ALL ON ALL TABLES IN SCHEMA public TO ads_service;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ads_service;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE subscriptions IS 'Organization subscription plans and billing info';
COMMENT ON TABLE payment_history IS 'Record of all payment transactions';
COMMENT ON TABLE organization_usage IS 'Daily usage tracking per organization';
COMMENT ON TABLE usage_events IS 'Detailed log of all API and feature usage';
COMMENT ON TABLE feature_usage IS 'Track usage of specific features per organization';
COMMENT ON TABLE quota_alerts IS 'Quota warning and limit exceeded alerts';
COMMENT ON TABLE plan_change_requests IS 'Pending and processed plan change requests';
COMMENT ON TABLE coupons IS 'Discount coupons for subscriptions';
COMMENT ON TABLE credit_transactions IS 'Credit balance transactions for prepaid model';

COMMENT ON FUNCTION current_user_org_ids() IS 'Returns organization IDs accessible by current user from session setting';
COMMENT ON FUNCTION current_user_org_id() IS 'Returns primary organization ID for current request';
COMMENT ON FUNCTION is_org_member(UUID) IS 'Checks if given org ID is in current user accessible orgs';
