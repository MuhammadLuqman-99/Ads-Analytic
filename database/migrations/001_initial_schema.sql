-- ============================================================================
-- SaaS Ads Analytics Platform - PostgreSQL Schema
-- Supports: Meta Ads, TikTok Ads, Shopee Ads
-- Version: 1.0.0
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE platform_type AS ENUM ('meta', 'tiktok', 'shopee');
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'expired', 'revoked');
CREATE TYPE campaign_status AS ENUM ('active', 'paused', 'deleted', 'archived', 'draft');
CREATE TYPE campaign_objective AS ENUM (
    'awareness', 
    'traffic', 
    'engagement', 
    'leads', 
    'app_promotion', 
    'sales', 
    'conversions',
    'video_views',
    'messages',
    'store_traffic'
);
CREATE TYPE order_status AS ENUM (
    'pending', 
    'confirmed', 
    'shipped', 
    'delivered', 
    'cancelled', 
    'refunded',
    'return_requested',
    'returned'
);
CREATE TYPE user_role AS ENUM ('owner', 'admin', 'analyst', 'viewer');
CREATE TYPE subscription_plan AS ENUM ('free', 'starter', 'professional', 'enterprise');

-- ============================================================================
-- 1. USERS & ORGANIZATIONS (Multi-Tenant)
-- ============================================================================

-- Organizations (Tenants)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    logo_url VARCHAR(500),
    subscription_plan subscription_plan NOT NULL DEFAULT 'free',
    subscription_expires_at TIMESTAMP WITH TIME ZONE,
    settings JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url VARCHAR(500),
    phone VARCHAR(20),
    email_verified_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Organization Members (Many-to-Many with roles)
CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role user_role NOT NULL DEFAULT 'viewer',
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP WITH TIME ZONE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, user_id)
);

-- ============================================================================
-- 2. CONNECTED ACCOUNTS (OAuth Tokens for each platform)
-- ============================================================================

CREATE TABLE connected_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    -- Platform-specific identifiers
    platform_account_id VARCHAR(255) NOT NULL,
    platform_account_name VARCHAR(255),
    platform_user_id VARCHAR(255),
    
    -- OAuth tokens (encrypted at application level)
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    token_expires_at TIMESTAMP WITH TIME ZONE,
    token_scopes TEXT[], -- Array of granted scopes
    
    -- Account status
    status account_status NOT NULL DEFAULT 'active',
    last_synced_at TIMESTAMP WITH TIME ZONE,
    sync_error TEXT,
    
    -- Additional metadata
    account_timezone VARCHAR(50),
    account_currency VARCHAR(3) DEFAULT 'MYR',
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint per organization and platform account
    UNIQUE(organization_id, platform, platform_account_id)
);

-- Token refresh history for audit
CREATE TABLE token_refresh_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    connected_account_id UUID NOT NULL REFERENCES connected_accounts(id) ON DELETE CASCADE,
    refresh_status VARCHAR(20) NOT NULL, -- 'success', 'failed'
    error_message TEXT,
    old_expires_at TIMESTAMP WITH TIME ZONE,
    new_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 3. AD CAMPAIGNS (Normalized structure for all platforms)
-- ============================================================================

-- Ad Accounts (Business/Ad accounts within connected accounts)
CREATE TABLE ad_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    connected_account_id UUID NOT NULL REFERENCES connected_accounts(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    -- Platform-specific identifiers
    platform_ad_account_id VARCHAR(255) NOT NULL,
    platform_ad_account_name VARCHAR(255),
    
    -- Account settings
    currency VARCHAR(3) DEFAULT 'MYR',
    timezone VARCHAR(50),
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(connected_account_id, platform_ad_account_id)
);

-- Campaigns
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ad_account_id UUID NOT NULL REFERENCES ad_accounts(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    -- Platform-specific identifiers
    platform_campaign_id VARCHAR(255) NOT NULL,
    platform_campaign_name VARCHAR(500),
    
    -- Campaign details
    objective campaign_objective,
    status campaign_status NOT NULL DEFAULT 'active',
    
    -- Budget
    daily_budget DECIMAL(15, 2),
    lifetime_budget DECIMAL(15, 2),
    budget_currency VARCHAR(3) DEFAULT 'MYR',
    
    -- Schedule
    start_date DATE,
    end_date DATE,
    
    -- Platform-specific data
    platform_data JSONB DEFAULT '{}',
    
    -- Timestamps
    platform_created_at TIMESTAMP WITH TIME ZONE,
    platform_updated_at TIMESTAMP WITH TIME ZONE,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(ad_account_id, platform_campaign_id)
);

-- Ad Sets / Ad Groups
CREATE TABLE ad_sets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    -- Platform-specific identifiers
    platform_ad_set_id VARCHAR(255) NOT NULL,
    platform_ad_set_name VARCHAR(500),
    
    -- Status
    status campaign_status NOT NULL DEFAULT 'active',
    
    -- Budget & Bidding
    daily_budget DECIMAL(15, 2),
    lifetime_budget DECIMAL(15, 2),
    bid_amount DECIMAL(15, 4),
    bid_strategy VARCHAR(100),
    
    -- Targeting (stored as JSONB for flexibility)
    targeting JSONB DEFAULT '{}',
    
    -- Schedule
    start_date DATE,
    end_date DATE,
    
    -- Platform-specific data
    platform_data JSONB DEFAULT '{}',
    
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(campaign_id, platform_ad_set_id)
);

-- Ads (Creative level)
CREATE TABLE ads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ad_set_id UUID NOT NULL REFERENCES ad_sets(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    -- Platform-specific identifiers
    platform_ad_id VARCHAR(255) NOT NULL,
    platform_ad_name VARCHAR(500),
    
    -- Status
    status campaign_status NOT NULL DEFAULT 'active',
    
    -- Creative details
    headline VARCHAR(500),
    description TEXT,
    call_to_action VARCHAR(100),
    destination_url TEXT,
    display_url VARCHAR(255),
    
    -- Media
    image_url TEXT,
    video_url TEXT,
    thumbnail_url TEXT,
    
    -- Platform-specific creative data
    creative_data JSONB DEFAULT '{}',
    platform_data JSONB DEFAULT '{}',
    
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(ad_set_id, platform_ad_id)
);

-- ============================================================================
-- 4. AD METRICS (Daily Snapshots)
-- ============================================================================

-- Campaign-level daily metrics
CREATE TABLE campaign_metrics_daily (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    -- Date of the metrics
    metric_date DATE NOT NULL,
    
    -- Core metrics
    impressions BIGINT DEFAULT 0,
    reach BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    unique_clicks BIGINT DEFAULT 0,
    
    -- Cost metrics
    spend DECIMAL(15, 4) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'MYR',
    
    -- Engagement metrics
    likes BIGINT DEFAULT 0,
    comments BIGINT DEFAULT 0,
    shares BIGINT DEFAULT 0,
    saves BIGINT DEFAULT 0,
    video_views BIGINT DEFAULT 0,
    video_views_p25 BIGINT DEFAULT 0,
    video_views_p50 BIGINT DEFAULT 0,
    video_views_p75 BIGINT DEFAULT 0,
    video_views_p100 BIGINT DEFAULT 0,
    
    -- Conversion metrics
    conversions BIGINT DEFAULT 0,
    conversion_value DECIMAL(15, 4) DEFAULT 0,
    add_to_cart BIGINT DEFAULT 0,
    checkout_initiated BIGINT DEFAULT 0,
    purchases BIGINT DEFAULT 0,
    purchase_value DECIMAL(15, 4) DEFAULT 0,
    
    -- Calculated metrics (can be computed, but stored for quick access)
    ctr DECIMAL(10, 6), -- Click-through rate
    cpc DECIMAL(15, 4), -- Cost per click
    cpm DECIMAL(15, 4), -- Cost per mille
    cpa DECIMAL(15, 4), -- Cost per acquisition
    roas DECIMAL(10, 4), -- Return on ad spend
    
    -- Platform-specific metrics
    platform_metrics JSONB DEFAULT '{}',
    
    -- Sync metadata
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(campaign_id, metric_date)
);

-- Ad Set-level daily metrics
CREATE TABLE ad_set_metrics_daily (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ad_set_id UUID NOT NULL REFERENCES ad_sets(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    metric_date DATE NOT NULL,
    
    -- Core metrics
    impressions BIGINT DEFAULT 0,
    reach BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    unique_clicks BIGINT DEFAULT 0,
    
    -- Cost metrics
    spend DECIMAL(15, 4) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'MYR',
    
    -- Engagement metrics
    likes BIGINT DEFAULT 0,
    comments BIGINT DEFAULT 0,
    shares BIGINT DEFAULT 0,
    saves BIGINT DEFAULT 0,
    video_views BIGINT DEFAULT 0,
    
    -- Conversion metrics
    conversions BIGINT DEFAULT 0,
    conversion_value DECIMAL(15, 4) DEFAULT 0,
    purchases BIGINT DEFAULT 0,
    purchase_value DECIMAL(15, 4) DEFAULT 0,
    
    -- Calculated metrics
    ctr DECIMAL(10, 6),
    cpc DECIMAL(15, 4),
    cpm DECIMAL(15, 4),
    cpa DECIMAL(15, 4),
    roas DECIMAL(10, 4),
    
    platform_metrics JSONB DEFAULT '{}',
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(ad_set_id, metric_date)
);

-- Ad-level daily metrics
CREATE TABLE ad_metrics_daily (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ad_id UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
    ad_set_id UUID NOT NULL REFERENCES ad_sets(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform platform_type NOT NULL,
    
    metric_date DATE NOT NULL,
    
    -- Core metrics
    impressions BIGINT DEFAULT 0,
    reach BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    unique_clicks BIGINT DEFAULT 0,
    
    -- Cost metrics
    spend DECIMAL(15, 4) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'MYR',
    
    -- Engagement metrics
    likes BIGINT DEFAULT 0,
    comments BIGINT DEFAULT 0,
    shares BIGINT DEFAULT 0,
    saves BIGINT DEFAULT 0,
    video_views BIGINT DEFAULT 0,
    
    -- Conversion metrics
    conversions BIGINT DEFAULT 0,
    conversion_value DECIMAL(15, 4) DEFAULT 0,
    purchases BIGINT DEFAULT 0,
    purchase_value DECIMAL(15, 4) DEFAULT 0,
    
    -- Calculated metrics
    ctr DECIMAL(10, 6),
    cpc DECIMAL(15, 4),
    cpm DECIMAL(15, 4),
    cpa DECIMAL(15, 4),
    roas DECIMAL(10, 4),
    
    platform_metrics JSONB DEFAULT '{}',
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(ad_id, metric_date)
);

-- ============================================================================
-- 5. ORDERS/SALES (For ROAS calculation from Shopee)
-- ============================================================================

-- Shopee Shops (linked to connected accounts)
CREATE TABLE shopee_shops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    connected_account_id UUID NOT NULL REFERENCES connected_accounts(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    shop_id VARCHAR(100) NOT NULL,
    shop_name VARCHAR(255),
    shop_region VARCHAR(10), -- MY, SG, TH, etc.
    
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(connected_account_id, shop_id)
);

-- Orders
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shopee_shop_id UUID REFERENCES shopee_shops(id) ON DELETE SET NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Platform identifiers
    platform platform_type NOT NULL DEFAULT 'shopee',
    platform_order_id VARCHAR(100) NOT NULL,
    platform_order_sn VARCHAR(100), -- Shopee order serial number
    
    -- Order details
    status order_status NOT NULL DEFAULT 'pending',
    
    -- Customer info (anonymized/hashed for privacy)
    customer_id VARCHAR(255),
    customer_name VARCHAR(255),
    
    -- Pricing
    currency VARCHAR(3) DEFAULT 'MYR',
    subtotal DECIMAL(15, 4) NOT NULL DEFAULT 0,
    shipping_fee DECIMAL(15, 4) DEFAULT 0,
    discount_amount DECIMAL(15, 4) DEFAULT 0,
    voucher_amount DECIMAL(15, 4) DEFAULT 0,
    seller_discount DECIMAL(15, 4) DEFAULT 0,
    platform_discount DECIMAL(15, 4) DEFAULT 0,
    total_amount DECIMAL(15, 4) NOT NULL DEFAULT 0,
    
    -- Commission & Fees
    commission_fee DECIMAL(15, 4) DEFAULT 0,
    service_fee DECIMAL(15, 4) DEFAULT 0,
    transaction_fee DECIMAL(15, 4) DEFAULT 0,
    
    -- Profit calculation
    estimated_profit DECIMAL(15, 4),
    
    -- Tracking
    tracking_number VARCHAR(100),
    shipping_carrier VARCHAR(100),
    
    -- Attribution (link to ad campaign if trackable)
    attributed_campaign_id UUID REFERENCES campaigns(id) ON DELETE SET NULL,
    attributed_ad_id UUID REFERENCES ads(id) ON DELETE SET NULL,
    utm_source VARCHAR(100),
    utm_medium VARCHAR(100),
    utm_campaign VARCHAR(255),
    utm_content VARCHAR(255),
    
    -- Timestamps
    order_date DATE NOT NULL,
    order_created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    paid_at TIMESTAMP WITH TIME ZONE,
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    -- Platform-specific data
    platform_data JSONB DEFAULT '{}',
    
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(organization_id, platform, platform_order_id)
);

-- Order Items
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Product info
    platform_item_id VARCHAR(100),
    platform_product_id VARCHAR(100),
    sku VARCHAR(100),
    product_name VARCHAR(500),
    variation_name VARCHAR(255),
    
    -- Pricing
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(15, 4) NOT NULL,
    discount_amount DECIMAL(15, 4) DEFAULT 0,
    total_price DECIMAL(15, 4) NOT NULL,
    
    -- Cost (for profit calculation)
    unit_cost DECIMAL(15, 4),
    
    -- Status
    status order_status,
    
    platform_data JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Daily Sales Summary (for quick ROAS calculation)
CREATE TABLE sales_summary_daily (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shopee_shop_id UUID REFERENCES shopee_shops(id) ON DELETE SET NULL,
    platform platform_type NOT NULL DEFAULT 'shopee',
    
    summary_date DATE NOT NULL,
    
    -- Order counts
    total_orders INTEGER DEFAULT 0,
    confirmed_orders INTEGER DEFAULT 0,
    cancelled_orders INTEGER DEFAULT 0,
    
    -- Revenue
    gross_revenue DECIMAL(15, 4) DEFAULT 0,
    net_revenue DECIMAL(15, 4) DEFAULT 0, -- After discounts
    total_discounts DECIMAL(15, 4) DEFAULT 0,
    shipping_collected DECIMAL(15, 4) DEFAULT 0,
    
    -- Costs
    total_cogs DECIMAL(15, 4) DEFAULT 0, -- Cost of goods sold
    commission_fees DECIMAL(15, 4) DEFAULT 0,
    service_fees DECIMAL(15, 4) DEFAULT 0,
    
    -- Profit
    estimated_profit DECIMAL(15, 4) DEFAULT 0,
    
    -- Items
    total_items_sold INTEGER DEFAULT 0,
    
    currency VARCHAR(3) DEFAULT 'MYR',
    
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(organization_id, shopee_shop_id, platform, summary_date)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Organizations
CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_is_active ON organizations(is_active);

-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Organization Members
CREATE INDEX idx_org_members_org ON organization_members(organization_id);
CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_org_members_role ON organization_members(role);

-- Connected Accounts
CREATE INDEX idx_connected_accounts_org ON connected_accounts(organization_id);
CREATE INDEX idx_connected_accounts_platform ON connected_accounts(platform);
CREATE INDEX idx_connected_accounts_status ON connected_accounts(status);
CREATE INDEX idx_connected_accounts_org_platform ON connected_accounts(organization_id, platform);

-- Ad Accounts
CREATE INDEX idx_ad_accounts_org ON ad_accounts(organization_id);
CREATE INDEX idx_ad_accounts_connected ON ad_accounts(connected_account_id);
CREATE INDEX idx_ad_accounts_platform ON ad_accounts(platform);

-- Campaigns
CREATE INDEX idx_campaigns_org ON campaigns(organization_id);
CREATE INDEX idx_campaigns_ad_account ON campaigns(ad_account_id);
CREATE INDEX idx_campaigns_platform ON campaigns(platform);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_org_platform ON campaigns(organization_id, platform);
CREATE INDEX idx_campaigns_date_range ON campaigns(start_date, end_date);

-- Ad Sets
CREATE INDEX idx_ad_sets_org ON ad_sets(organization_id);
CREATE INDEX idx_ad_sets_campaign ON ad_sets(campaign_id);
CREATE INDEX idx_ad_sets_platform ON ad_sets(platform);
CREATE INDEX idx_ad_sets_status ON ad_sets(status);

-- Ads
CREATE INDEX idx_ads_org ON ads(organization_id);
CREATE INDEX idx_ads_campaign ON ads(campaign_id);
CREATE INDEX idx_ads_ad_set ON ads(ad_set_id);
CREATE INDEX idx_ads_platform ON ads(platform);
CREATE INDEX idx_ads_status ON ads(status);

-- Campaign Metrics Daily (Optimized for date range queries)
CREATE INDEX idx_campaign_metrics_org ON campaign_metrics_daily(organization_id);
CREATE INDEX idx_campaign_metrics_campaign ON campaign_metrics_daily(campaign_id);
CREATE INDEX idx_campaign_metrics_date ON campaign_metrics_daily(metric_date);
CREATE INDEX idx_campaign_metrics_platform ON campaign_metrics_daily(platform);
CREATE INDEX idx_campaign_metrics_org_date ON campaign_metrics_daily(organization_id, metric_date);
CREATE INDEX idx_campaign_metrics_org_platform_date ON campaign_metrics_daily(organization_id, platform, metric_date);
CREATE INDEX idx_campaign_metrics_campaign_date_range ON campaign_metrics_daily(campaign_id, metric_date DESC);

-- Ad Set Metrics Daily
CREATE INDEX idx_ad_set_metrics_org ON ad_set_metrics_daily(organization_id);
CREATE INDEX idx_ad_set_metrics_ad_set ON ad_set_metrics_daily(ad_set_id);
CREATE INDEX idx_ad_set_metrics_date ON ad_set_metrics_daily(metric_date);
CREATE INDEX idx_ad_set_metrics_platform ON ad_set_metrics_daily(platform);
CREATE INDEX idx_ad_set_metrics_org_date ON ad_set_metrics_daily(organization_id, metric_date);
CREATE INDEX idx_ad_set_metrics_org_platform_date ON ad_set_metrics_daily(organization_id, platform, metric_date);

-- Ad Metrics Daily
CREATE INDEX idx_ad_metrics_org ON ad_metrics_daily(organization_id);
CREATE INDEX idx_ad_metrics_ad ON ad_metrics_daily(ad_id);
CREATE INDEX idx_ad_metrics_date ON ad_metrics_daily(metric_date);
CREATE INDEX idx_ad_metrics_platform ON ad_metrics_daily(platform);
CREATE INDEX idx_ad_metrics_org_date ON ad_metrics_daily(organization_id, metric_date);
CREATE INDEX idx_ad_metrics_org_platform_date ON ad_metrics_daily(organization_id, platform, metric_date);

-- Shopee Shops
CREATE INDEX idx_shopee_shops_org ON shopee_shops(organization_id);
CREATE INDEX idx_shopee_shops_connected ON shopee_shops(connected_account_id);

-- Orders (Optimized for date range and attribution queries)
CREATE INDEX idx_orders_org ON orders(organization_id);
CREATE INDEX idx_orders_shop ON orders(shopee_shop_id);
CREATE INDEX idx_orders_platform ON orders(platform);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_date ON orders(order_date);
CREATE INDEX idx_orders_org_date ON orders(organization_id, order_date);
CREATE INDEX idx_orders_org_platform_date ON orders(organization_id, platform, order_date);
CREATE INDEX idx_orders_attributed_campaign ON orders(attributed_campaign_id);
CREATE INDEX idx_orders_attributed_ad ON orders(attributed_ad_id);
CREATE INDEX idx_orders_platform_order_id ON orders(platform_order_id);

-- Order Items
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_org ON order_items(organization_id);
CREATE INDEX idx_order_items_sku ON order_items(sku);
CREATE INDEX idx_order_items_product ON order_items(platform_product_id);

-- Sales Summary Daily
CREATE INDEX idx_sales_summary_org ON sales_summary_daily(organization_id);
CREATE INDEX idx_sales_summary_shop ON sales_summary_daily(shopee_shop_id);
CREATE INDEX idx_sales_summary_date ON sales_summary_daily(summary_date);
CREATE INDEX idx_sales_summary_org_date ON sales_summary_daily(organization_id, summary_date);
CREATE INDEX idx_sales_summary_platform ON sales_summary_daily(platform);
CREATE INDEX idx_sales_summary_org_platform_date ON sales_summary_daily(organization_id, platform, summary_date);

-- Token Refresh Logs
CREATE INDEX idx_token_refresh_account ON token_refresh_logs(connected_account_id);
CREATE INDEX idx_token_refresh_created ON token_refresh_logs(created_at);

-- ============================================================================
-- TRIGGERS FOR updated_at
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables with updated_at
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_org_members_updated_at BEFORE UPDATE ON organization_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_connected_accounts_updated_at BEFORE UPDATE ON connected_accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ad_accounts_updated_at BEFORE UPDATE ON ad_accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ad_sets_updated_at BEFORE UPDATE ON ad_sets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ads_updated_at BEFORE UPDATE ON ads FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_campaign_metrics_updated_at BEFORE UPDATE ON campaign_metrics_daily FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ad_set_metrics_updated_at BEFORE UPDATE ON ad_set_metrics_daily FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ad_metrics_updated_at BEFORE UPDATE ON ad_metrics_daily FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_shopee_shops_updated_at BEFORE UPDATE ON shopee_shops FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_order_items_updated_at BEFORE UPDATE ON order_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sales_summary_updated_at BEFORE UPDATE ON sales_summary_daily FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE organizations IS 'Multi-tenant organizations (SaaS customers)';
COMMENT ON TABLE users IS 'User accounts that can belong to multiple organizations';
COMMENT ON TABLE organization_members IS 'Many-to-many relationship between users and organizations with roles';
COMMENT ON TABLE connected_accounts IS 'OAuth connections to ad platforms (Meta, TikTok, Shopee)';
COMMENT ON TABLE ad_accounts IS 'Ad/Business accounts within connected platform accounts';
COMMENT ON TABLE campaigns IS 'Normalized ad campaigns from all platforms';
COMMENT ON TABLE ad_sets IS 'Ad sets/Ad groups within campaigns';
COMMENT ON TABLE ads IS 'Individual ads with creative details';
COMMENT ON TABLE campaign_metrics_daily IS 'Daily performance snapshots at campaign level';
COMMENT ON TABLE ad_set_metrics_daily IS 'Daily performance snapshots at ad set level';
COMMENT ON TABLE ad_metrics_daily IS 'Daily performance snapshots at ad level';
COMMENT ON TABLE shopee_shops IS 'Shopee shop accounts linked for order syncing';
COMMENT ON TABLE orders IS 'E-commerce orders from Shopee for ROAS calculation';
COMMENT ON TABLE order_items IS 'Individual items within orders';
COMMENT ON TABLE sales_summary_daily IS 'Pre-aggregated daily sales for quick ROAS queries';
