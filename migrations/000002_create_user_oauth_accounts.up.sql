CREATE TABLE user_oauth_accounts (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  provider         VARCHAR(32) NOT NULL,
  provider_user_id VARCHAR(255) NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT user_oauth_accounts_provider_user_unique UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_user_oauth_accounts_user_id ON user_oauth_accounts (user_id);
