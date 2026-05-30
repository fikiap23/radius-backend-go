CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name              VARCHAR(255) NOT NULL,
  email             CITEXT NOT NULL,
  password_hash     TEXT,
  email_verified_at TIMESTAMPTZ,
  avatar_url        TEXT,
  last_login_at     TIMESTAMPTZ,
  timezone          VARCHAR(64),
  locale            VARCHAR(10) NOT NULL DEFAULT 'en',
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at        TIMESTAMPTZ,
  CONSTRAINT users_email_unique UNIQUE (email),
  CONSTRAINT users_email_format CHECK (email ~* '^[^@]+@[^@]+\.[^@]+$')
);

CREATE INDEX idx_users_deleted_at ON users (deleted_at);
