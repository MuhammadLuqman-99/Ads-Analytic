-- Migration: Email Verification & Password Reset Tokens
-- Created: 2024-01-01

-- Email verification and password reset tokens table
CREATE TABLE IF NOT EXISTS verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(50) NOT NULL, -- 'email_verification', 'password_reset'
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    is_used BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_verification_tokens_user ON verification_tokens(user_id);
CREATE INDEX idx_verification_tokens_token ON verification_tokens(token);
CREATE INDEX idx_verification_tokens_type ON verification_tokens(token_type);
CREATE INDEX idx_verification_tokens_expires ON verification_tokens(expires_at);

-- Function to auto-cleanup expired tokens (optional, can run via cron)
CREATE OR REPLACE FUNCTION cleanup_expired_verification_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM verification_tokens
    WHERE expires_at < NOW() - INTERVAL '1 day'
    OR (is_used = true AND used_at < NOW() - INTERVAL '7 days');
END;
$$ LANGUAGE plpgsql;

-- Comment
COMMENT ON TABLE verification_tokens IS 'Stores email verification and password reset tokens';
COMMENT ON COLUMN verification_tokens.token_type IS 'Type of token: email_verification or password_reset';
